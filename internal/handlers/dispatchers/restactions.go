package dispatchers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/krateoplatformops/snowplow/apis"
	v1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	"github.com/krateoplatformops/snowplow/internal/objects"
	"github.com/krateoplatformops/snowplow/internal/resolvers/restactions"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func RESTAction() http.Handler {
	return &customformHandler{
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

	gvr, err := util.ParseGVR(req)
	if err != nil {
		log.Error("unable to parse group version resource", slog.Any("err", err))
		response.BadRequest(wri, err)
		return
	}

	nsn, err := util.ParseNamespacedName(req)
	if err != nil {
		log.Error("unable to parse namespaced name", slog.Any("err", err))
		response.BadRequest(wri, err)
		return
	}

	got := objects.Get(req.Context(), objects.Reference{
		Name: nsn.Name, Namespace: nsn.Namespace,
		APIVersion: gvr.GroupVersion().String(),
		Resource:   gvr.Resource,
	})
	if got.Err != nil {
		response.Encode(wri, got.Err)
		return
	}

	res, err := ResolveRESTAction(req.Context(), got.Unstructured, ResolveRESTActionOptions{
		Username:   req.Header.Get(xcontext.LabelKrateoUser),
		UserGroups: strings.Split(req.Header.Get(xcontext.LabelKrateoGroups), ","),
		AuthnNS:    r.authnNS,
	})
	if err != nil {
		log.Error("unable to resolve custom form",
			slog.String("name", nsn.String()), slog.String("gvr", gvr.String()), slog.Any("err", err))
		response.InternalError(wri, err)
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	enc.Encode(res)
}

type ResolveRESTActionOptions struct {
	AuthnNS    string
	Username   string
	UserGroups []string
}

func ResolveRESTAction(ctx context.Context, in *unstructured.Unstructured, opts ResolveRESTActionOptions) (runtime.Object, error) {
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return nil, err
	}

	var cr v1.RESTAction
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(in.Object, &cr)
	if err != nil {
		return nil, err
	}

	ctx = xcontext.BuildContext(ctx, xcontext.WithJQTemplate())
	return restactions.Resolve(ctx, restactions.ResolveOptions{
		In:         &cr,
		Username:   opts.Username,
		UserGroups: opts.UserGroups,
		AuthnNS:    opts.AuthnNS,
	})
}
