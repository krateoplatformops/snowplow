package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/krateoplatformops/plumbing/cache"
	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/http/response"
	"github.com/krateoplatformops/plumbing/kubeutil/plurals"
	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Plurals() http.Handler {
	return &pluralsHandler{
		store: cache.NewTTL[string, plurals.Info](),
	}
}

var _ http.Handler = (*pluralsHandler)(nil)

type pluralsHandler struct {
	store *cache.TTLCache[string, plurals.Info]
}

// @Summary Names Endpoint
// @Description Returns information about Kubernetes API names
// @ID names
// @Param  apiVersion       query   string  true  "API Group and Version"
// @Param  kind             query   string  true  "API Kind"
// @Produce  json
// @Success 200 {object} names
// @Failure 400 {object} response.Status
// @Failure 401 {object} response.Status
// @Failure 404 {object} response.Status
// @Failure 500 {object} response.Status
// @Router /api-info/names [get]
func (r *pluralsHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	gvk, err := r.validateRequest(req)
	if err != nil {
		response.BadRequest(wri, err)
		return
	}

	log := xcontext.Logger(req.Context())

	start := time.Now()

	tmp, err := plurals.Get(gvk, plurals.GetOptions{
		Logger:       log,
		Cache:        r.store,
		ResolverFunc: plurals.ResolveAPINames,
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			response.NotFound(wri, err)
		} else {
			response.InternalError(wri, err)
		}
		return
	}

	log.Info("plurals successfully resolved",
		slog.String("gvk", gvk.String()),
		slog.String("duration", util.ETA(start)),
	)

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&tmp); err != nil {
		log.Error("unable to serve api call response", slog.Any("err", err))
	}
}

func (r *pluralsHandler) validateRequest(req *http.Request) (gvk schema.GroupVersionKind, err error) {
	apiVersion := req.URL.Query().Get("apiVersion")
	if len(apiVersion) == 0 {
		err = fmt.Errorf("missing 'apiVersion' query parameter")
		return
	}

	kind := req.URL.Query().Get("kind")
	if len(apiVersion) == 0 {
		err = fmt.Errorf("missing 'kind' query parameter")
		return
	}

	var gv schema.GroupVersion
	gv, err = schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return
	}
	gvk = gv.WithKind(kind)

	return
}

type names struct {
	Plural   string   `json:"plural"`
	Singular string   `json:"singular"`
	Shorts   []string `json:"shorts"`
}
