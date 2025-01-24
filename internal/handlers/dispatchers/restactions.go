package dispatchers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/krateoplatformops/snowplow/apis"
	v1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/resolvers/restactions"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"k8s.io/apimachinery/pkg/runtime"
)

func RESTAction() http.Handler {
	return &restActionHandler{
		authnNS: env.String("AUTHN_NAMESPACE", ""),
		verbose: env.True("DEBUG"),
	}
}

type restActionHandler struct {
	authnNS string
	verbose bool
}

var _ http.Handler = (*restActionHandler)(nil)

func (r *restActionHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := xcontext.Logger(req.Context())

	got := fetchObject(req)
	if got.Err != nil {
		response.Encode(wri, got.Err)
		return
	}

	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		log.Error("unable to add apis to scheme",
			slog.Any("err", err))
		response.InternalError(wri, err)
		return
	}

	var cr v1.RESTAction
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(got.Unstructured.Object, &cr)
	if err != nil {
		log.Error("unable to convert unstructured to typed rest action",
			slog.String("name", got.Unstructured.GetName()),
			slog.String("namespace", got.Unstructured.GetNamespace()),
			slog.Any("err", err))
		response.InternalError(wri, err)
		return
	}

	ctx := xcontext.BuildContext(req.Context(), xcontext.WithJQ())
	res, err := restactions.Resolve(ctx, restactions.ResolveOptions{
		In:         &cr,
		Username:   req.Header.Get(xcontext.LabelKrateoUser),
		UserGroups: strings.Split(req.Header.Get(xcontext.LabelKrateoGroups), ","),
		AuthnNS:    r.authnNS,
	})
	if err != nil {
		log.Error("unable to resolve rest action",
			slog.String("name", cr.GetName()),
			slog.String("namespace", cr.GetNamespace()),
			slog.Any("err", err))
		response.InternalError(wri, err)
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	enc.Encode(res)
}
