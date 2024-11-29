package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/server/middlewares"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// @Summary List resources by category in a specified namespace.
// @Description Resources List
// @ID list
// @Param  X-Krateo-User    header  string  true  "Krateo User"
// @Param  X-Krateo-Groups  header  string  true  "Krateo User Groups"
// @Param  category         query   string  true  "Resource category"
// @Param  ns               query   string  false  "Namespace"
// @Produce  json
// @Success 200 {object} map[string]any
// @Failure 400 {object} status.Status
// @Failure 401 {object} status.Status
// @Failure 404 {object} status.Status
// @Failure 500 {object} status.Status
// @Router /list [get]
func List(authnNS string, verbose bool) http.HandlerFunc {
	return func(wri http.ResponseWriter, req *http.Request) {
		cat := req.URL.Query().Get("category")
		ns := req.URL.Query().Get("ns")

		if len(cat) == 0 {
			status.BadRequest(wri, fmt.Errorf("missing 'category' params"))
			return
		}

		log := middlewares.LogFromContext(req.Context())

		rc, err := middlewares.RESTConfigFromContext(req.Context())
		if err != nil {
			log.Error("unable to get user client config", slog.Any("err", err))
			status.Unauthorized(wri, err)
			return
		}

		cli, err := dynamic.NewClient(rc)
		if err != nil {
			log.Error("cannot create dynamic client", slog.Any("err", err))
			status.InternalError(wri, err)
			return
		}

		log.Debug("performing discovery", slog.String("category", cat))
		res, err := cli.Discover(context.Background(), cat)
		if err != nil {
			log.Error("discovery failed", slog.Any("err", err))
			status.InternalError(wri, err)
			return
		}
		log.Debug(fmt.Sprintf("discovery terminated (found: %d)", len(res)))

		rt := []unstructured.Unstructured{}

		for _, gvr := range res {
			opts := dynamic.Options{
				Namespace: ns,
				GVR:       gvr,
			}

			obj, err := cli.List(context.Background(), opts)
			if err != nil {
				log.Error("cannot list resources",
					slog.String("gvr", gvr.String()), slog.Any("err", err))
				if apierrors.IsForbidden(err) {
					status.Forbidden(wri, err)
					return
				}
				continue
			}

			for _, x := range obj.Items {
				unstructured.RemoveNestedField(
					x.UnstructuredContent(), "metadata", "managedFields")
				rt = append(rt, x)
			}
		}

		wri.Header().Set("Content-Type", "application/json")
		wri.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(wri)
		enc.SetIndent("", "  ")
		enc.Encode(rt)
	}
}
