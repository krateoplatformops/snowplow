package resourcesrefs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/kubeconfig"
	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"github.com/krateoplatformops/snowplow/internal/rbac"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func Resolve(ctx context.Context, items []templatesv1.ResourceRef) ([]templatesv1.ResourceRefResult, error) {

	ep, err := xcontext.UserConfig(ctx)
	if err != nil {
		return nil, err
	}

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		return nil, err
	}

	results := []templatesv1.ResourceRefResult{}
	for _, el := range items {
		res, err2 := resolveOne(ctx, rc, &el)
		if err2 != nil {
			err = errors.Join(err, err2)
			continue
		}

		results = append(results, res...)
	}

	return results, nil
}

func resolveOne(ctx context.Context, rc *rest.Config, in *templatesv1.ResourceRef) ([]templatesv1.ResourceRefResult, error) {
	all := []templatesv1.ResourceRefResult{}
	if in == nil {
		return all, nil
	}

	gv, err := schema.ParseGroupVersion(in.APIVersion)
	if err != nil {
		return all, err
	}
	gvr := gv.WithResource(in.Resource)

	gvk, err := dynamic.KindFor(rc, gvr)
	if err != nil {
		return all, err
	}

	verbs := mapVerbs(in.Verb)
	for _, verb := range verbs {
		ok := rbac.UserCan(ctx, rbac.UserCanOptions{
			Verb:          verb,
			GroupResource: gvr.GroupResource(),
			Namespace:     in.Namespace,
		})
		if !ok {
			xcontext.Logger(ctx).Warn("action not allowed",
				slog.String("verb", verb),
				slog.String("group", gvr.Group),
				slog.String("resource", gvr.Resource),
				slog.String("namespace", in.Namespace))
			continue
		}

		el := templatesv1.ResourceRefResult{
			ID:   in.ID,
			Verb: kubeToREST[verb],
		}
		if in.Name == "" {
			el.Path = fmt.Sprintf("/call?resource=%s&apiVersion=%s&namespace=%s",
				gvr.Resource, gvr.GroupVersion().String(), in.Namespace)
		} else {
			el.Path = fmt.Sprintf("/call?resource=%s&apiVersion=%s&name=%s&namespace=%s",
				gvr.Resource, gvr.GroupVersion().String(), in.Name, in.Namespace)
		}

		if el.Verb == http.MethodPost || el.Verb == http.MethodPut || el.Verb == http.MethodPatch {
			el.Payload = &templatesv1.ResourceRefPayload{
				Kind:       gvk.Kind,
				APIVersion: in.APIVersion,
				MetaData: &templatesv1.Reference{
					Name:      in.Name,
					Namespace: in.Namespace,
				},
			}
		}

		all = append(all, el)
	}

	return all, nil
}
