package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/krateoplatformops/snowplow/plumbing/cache"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func Plurals() http.Handler {
	return &pluralsHandler{
		store: cache.NewTTL[string, names](),
	}
}

var _ http.Handler = (*pluralsHandler)(nil)

type pluralsHandler struct {
	store *cache.TTLCache[string, names]
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

	tmp, ok := r.store.Get(gvk.String())
	if !ok {
		log.Debug("cache miss", slog.String("gvk", gvk.String()))
		tmp, err = r.resolveAPINames(gvk)
		if err != nil {
			log.Error("unable to discover API names",
				slog.String("gvk", gvk.String()), slog.Any("err", err))
			if apierrors.IsNotFound(err) {
				response.NotFound(wri, err)
			} else {
				response.InternalError(wri, err)
			}
			return
		}

		r.store.Set(gvk.String(), tmp, time.Hour*48)
	} else {
		log.Debug("cache hit", slog.String("gvk", gvk.String()))
	}

	if len(tmp.Plural) == 0 {
		msg := fmt.Sprintf("no names found for %q", gvk.GroupVersion().String())
		log.Warn(msg)
		response.NotFound(wri, fmt.Errorf("%s", msg))
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&tmp); err != nil {
		log.Error("unable to serve api call response", slog.Any("err", err))
	}
}

func (r *pluralsHandler) resolveAPINames(gvk schema.GroupVersionKind) (names, error) {
	rc, err := rest.InClusterConfig()
	if err != nil {
		return names{}, err
	}

	dc, err := discovery.NewDiscoveryClientForConfig(rc)
	if err != nil {
		return names{}, err
	}

	list, err := dc.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return names{}, err
	}

	if list == nil || len(list.APIResources) == 0 {
		return names{}, nil
	}

	var tmp names
	for _, el := range list.APIResources {
		if el.Kind != gvk.Kind {
			continue
		}

		tmp = names{
			Plural:   el.Name,
			Singular: el.SingularName,
			Shorts:   el.ShortNames,
		}
		break
	}

	return tmp, nil
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
