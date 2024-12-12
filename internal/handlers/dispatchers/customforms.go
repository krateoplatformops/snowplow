package dispatchers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/krateoplatformops/snowplow/apis"
	"github.com/krateoplatformops/snowplow/apis/templates/v1alpha1"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"github.com/krateoplatformops/snowplow/internal/handlers/util"
	"github.com/krateoplatformops/snowplow/internal/resolvers/customforms"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/http/response/status"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func CustomForm() http.Handler {
	return &customformHandler{
		authnNS: env.String("AUTHN_NAMESPACE", ""),
		verbose: env.True("DEBUG"),
	}
}

const (
	lastAppliedConfigAnnotation = "kubectl.kubernetes.io/last-applied-configuration"
)

type customformHandler struct {
	authnNS string
	verbose bool
}

var _ http.Handler = (*customformHandler)(nil)

func (r *customformHandler) ServeHTTP(wri http.ResponseWriter, req *http.Request) {
	log := xcontext.Logger(req.Context())

	gvr, err := util.ParseGVR(req)
	if err != nil {
		log.Error("unable to parse group version resource", slog.Any("err", err))
		status.BadRequest(wri, err)
		return
	}

	nsn, err := util.ParseNamespacedName(req)
	if err != nil {
		log.Error("unable to parse namespaced name", slog.Any("err", err))
		status.BadRequest(wri, err)
		return
	}

	ep, err := xcontext.UserConfig(req.Context())
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		status.Unauthorized(wri, err)
		return
	}

	rc, err := kubeconfig.NewClientConfig(req.Context(), ep)
	if err != nil {
		log.Error("unable to create kubernetes client config", slog.Any("err", err))
		status.InternalError(wri, err)
		return
	}

	cli, err := dynamic.NewClient(rc)
	if err != nil {
		log.Error("unable to create kubernetes dynamic client", slog.Any("err", err))
		status.InternalError(wri, err)
		return
	}

	uns, err := cli.Get(context.Background(), nsn.Name, dynamic.Options{
		Namespace: nsn.Namespace,
		GVR:       gvr,
	})
	if err != nil {
		log.Error("unable to get resource", slog.String("name", nsn.String()),
			slog.String("gvr", gvr.String()), slog.Any("err", err))
		if apierrors.IsForbidden(err) {
			status.Forbidden(wri, err)
			return
		}

		if apierrors.IsNotFound(err) {
			status.NotFound(wri, err)
			return
		}

		status.InternalError(wri, err)
		return
	}

	annotations := uns.GetAnnotations()
	if annotations != nil {
		delete(annotations, lastAppliedConfigAnnotation)
		uns.SetAnnotations(annotations)
	}
	uns.SetManagedFields(nil)

	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		log.Error("unable to register templates scheme",
			slog.String("name", nsn.String()), slog.String("gvr", gvr.String()), slog.Any("err", err))
		status.InternalError(wri, err)
		return
	}

	var cr v1alpha1.CustomForm
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(uns.Object, &cr)
	if err != nil {
		log.Error("unable to convert unstructured to typed",
			slog.String("name", nsn.String()), slog.String("gvr", gvr.String()), slog.Any("err", err))
		status.InternalError(wri, err)
		return
	}

	ctx := xcontext.BuildContext(req.Context(), xcontext.WithJQTemplate())
	res, err := customforms.Resolve(ctx, customforms.ResolveOptions{
		In:         &cr,
		Username:   req.Header.Get(xcontext.LabelKrateoUser),
		UserGroups: strings.Split(req.Header.Get(xcontext.LabelKrateoGroups), ","),
		AuthnNS:    r.authnNS,
		Verbose:    r.verbose,
	})
	if err != nil {
		log.Error("unable to resolve custom form",
			slog.String("name", nsn.String()), slog.String("gvr", gvr.String()), slog.Any("err", err))
		status.InternalError(wri, err)
		return
	}

	wri.Header().Set("Content-Type", "application/json")
	wri.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(wri)
	enc.SetIndent("", "  ")
	enc.Encode(res)
}
