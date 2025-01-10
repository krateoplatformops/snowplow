package objects

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/krateoplatformops/snowplow/internal/dynamic"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/http/response"
	"github.com/krateoplatformops/snowplow/plumbing/kubeconfig"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	lastAppliedConfigAnnotation = "kubectl.kubernetes.io/last-applied-configuration"
)

type Result struct {
	GVR          schema.GroupVersionResource
	Unstructured *unstructured.Unstructured
	Err          *response.Status
}

type Reference struct {
	Name       string
	Namespace  string
	Resource   string
	APIVersion string
}

func Get(ctx context.Context, ref Reference) (res Result) {
	log := xcontext.Logger(ctx)

	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		log.Error("unable to parse group version", slog.Any("reference", ref), slog.Any("err", err))
		res.Err = response.New(http.StatusBadRequest, err)
		return
	}
	res.GVR = gv.WithResource(ref.Resource)

	ep, err := xcontext.UserConfig(ctx)
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		res.Err = response.New(http.StatusUnauthorized, err)
		return
	}

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		log.Error("unable to create kubernetes client config", slog.Any("err", err))
		res.Err = response.New(http.StatusInternalServerError, err)
		return
	}

	cli, err := dynamic.NewClient(rc)
	if err != nil {
		log.Error("unable to create kubernetes dynamic client", slog.Any("err", err))
		res.Err = response.New(http.StatusInternalServerError, err)
		return
	}

	uns, err := cli.Get(context.Background(), ref.Name, dynamic.Options{
		Namespace: ref.Namespace,
		GVR:       res.GVR,
	})
	if err != nil {
		log.Error("unable to get resource",
			slog.String("name", ref.Name), slog.String("namespace", ref.Namespace),
			slog.String("gvr", res.GVR.String()), slog.Any("err", err))

		res.Err = response.New(http.StatusInternalServerError, err)
		if apierrors.IsForbidden(err) {
			res.Err = response.New(http.StatusForbidden, err)
		} else if apierrors.IsNotFound(err) {
			res.Err = response.New(http.StatusNotFound, err)
		}

		return
	}

	annotations := uns.GetAnnotations()
	if annotations != nil {
		delete(annotations, lastAppliedConfigAnnotation)
		uns.SetAnnotations(annotations)
	}
	uns.SetManagedFields(nil)

	res.Unstructured = uns
	res.Err = nil
	return
}
