package resourcesrefs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/kubeconfig"
	"github.com/krateoplatformops/plumbing/maps"
	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"github.com/krateoplatformops/snowplow/internal/rbac"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

const (
	resourcesRefsKey = "resourcesRefs"
)

type ResolveOptions struct {
	RC      *rest.Config
	Widget  *unstructured.Unstructured
	AuthnNS string
}

func Resolve(ctx context.Context, opts ResolveOptions) ([]any, error) {
	log := xcontext.Logger(ctx)

	arr, ok, err := maps.NestedSlice(opts.Widget.Object, "spec", resourcesRefsKey)
	if err != nil {
		log.Error("unable to look for backendEndpoints existence", slog.Any("err", err))
		return nil, err
	}
	if !ok {
		log.Warn("no backendEndpoints found")
		return []any{}, nil
	}

	acts, err := fromUnstructuredSlice(arr)
	if err != nil {
		log.Error("unable to convert []any to []BackendEndpoint", slog.Any("err", err))
		return nil, err
	}

	ep, err := xcontext.UserConfig(ctx)
	if err != nil {
		log.Error("unable to get user endpoint", slog.Any("err", err))
		return nil, err
	}

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		log.Error("unable to create user client config", slog.Any("err", err))
		return nil, err
	}

	results := []templatesv1.ResourceRefResult{}
	for _, el := range acts {
		res, err2 := resolveOne(ctx, rc, &el)
		if err2 != nil {
			err = errors.Join(err, err2)
			continue
		}

		results = append(results, res...)
	}

	return toUnstructuredSlice(results)
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
			xcontext.Logger(ctx).Error("action not allowed",
				slog.String("verb", verb),
				slog.String("group", gvr.Group),
				slog.String("resource", gvr.Resource),
				slog.String("namespace", in.Namespace))
			continue
		}

		el := templatesv1.ResourceRefResult{
			ID:   in.ID,
			Verb: kubeToREST[verb],
			Path: fmt.Sprintf("/call?resource=%s&apiVersion=%s&name=%s&namespace=%s",
				gvr.Resource, gvr.GroupVersion().String(), in.Name, in.Namespace),
		}

		el.Payload = &templatesv1.ResourceRefPayload{
			Kind:       gvk.Kind,
			APIVersion: in.APIVersion,
			MetaData: &templatesv1.Reference{
				Name:      in.Name,
				Namespace: in.Namespace,
			},
		}

		all = append(all, el)
	}

	return all, nil
}

/*
func convertToBackendEndpoints(arr []any) ([]templatesv1.BackendEndpoint, error) {
	dat, err := json.Marshal(arr)
	if err != nil {
		return []templatesv1.BackendEndpoint{}, err
	}

	obj := []templatesv1.BackendEndpoint{}
	err = json.Unmarshal(dat, &obj)

	return obj, err
}
*/

func fromUnstructuredSlice(in []any) ([]templatesv1.ResourceRef, error) {
	var out []templatesv1.ResourceRef

	for _, item := range in {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unexpected item type %T, expected map[string]any", item)
		}

		bytes, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map: %w", err)
		}

		var ep templatesv1.ResourceRef
		if err := json.Unmarshal(bytes, &ep); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to BackendEndpoint: %w", err)
		}

		out = append(out, ep)
	}

	return out, nil
}

func toUnstructuredSlice(results []templatesv1.ResourceRefResult) ([]any, error) {
	var list []any

	for _, r := range results {
		bytes, err := json.Marshal(&r)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal BackendEndpointResult: %w", err)
		}

		var m map[string]any
		if err := json.Unmarshal(bytes, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal BackendEndpointResult into map: %w", err)
		}

		list = append(list, m)
	}

	return list, nil
}
