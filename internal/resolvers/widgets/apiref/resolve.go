package apiref

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krateoplatformops/plumbing/maps"
	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/objects"
	"github.com/krateoplatformops/snowplow/internal/resolvers/restactions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

const (
	apiRefKey = "apiRef"
)

type ResolveOptions struct {
	RC      *rest.Config
	Widget  *unstructured.Unstructured
	AuthnNS string
}

func Resolve(ctx context.Context, opts ResolveOptions) (map[string]any, error) {
	ref, err := getRESTActionRef(opts.Widget.Object)
	if err != nil {
		return map[string]any{}, err
	}
	if ref.Name == "" || ref.Namespace == "" {
		return map[string]any{}, nil
	}

	res := objects.Get(ctx, ref)
	if res.Err != nil {
		return map[string]any{}, fmt.Errorf("%s", res.Err.Message)
	}

	ra, err := convertToRESTAction(res.Unstructured.Object)
	if res.Err != nil {
		return map[string]any{}, err
	}

	raopts := restactions.ResolveOptions{
		In:      &ra,
		SArc:    opts.RC,
		AuthnNS: opts.AuthnNS,
	}

	if _, err = restactions.Resolve(ctx, raopts); err != nil {
		return map[string]any{}, err
	}

	return rawExtensionToMap(ra.Status)
}

func getRESTActionRef(in map[string]any) (objects.Reference, error) {
	api, ok, err := maps.NestedMapNoCopy(in, "spec", apiRefKey)
	if err != nil {
		return objects.Reference{}, err
	}
	if !ok {
		return objects.Reference{}, nil
	}

	dat, err := json.Marshal(api)
	if err != nil {
		return objects.Reference{}, err
	}

	ref := objects.Reference{
		Resource:   "restactions",
		APIVersion: fmt.Sprintf("%s/%s", templatesv1.Group, templatesv1.Version),
	}
	err = json.Unmarshal(dat, &ref)

	return ref, err
}

func convertToRESTAction(api map[string]any) (templatesv1.RESTAction, error) {
	dat, err := json.Marshal(api)
	if err != nil {
		return templatesv1.RESTAction{}, err
	}

	var ra templatesv1.RESTAction
	err = json.Unmarshal(dat, &ra)

	return ra, err
}
