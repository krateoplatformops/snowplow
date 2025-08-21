package apiref

import (
	"context"
	"fmt"

	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/objects"
	"github.com/krateoplatformops/snowplow/internal/resolvers/restactions"
	"k8s.io/client-go/rest"
)

type ResolveOptions struct {
	RC      *rest.Config
	ApiRef  templatesv1.ObjectReference
	AuthnNS string
	PerPage int
	Page    int
	Extras  map[string]any
}

func Resolve(ctx context.Context, opts ResolveOptions) (map[string]any, error) {
	if opts.ApiRef.Name == "" || opts.ApiRef.Namespace == "" {
		return map[string]any{}, nil
	}

	res := objects.Get(ctx, opts.ApiRef)
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
		PerPage: opts.PerPage,
		Page:    opts.Page,
		Extras:  opts.Extras,
	}

	if _, err = restactions.Resolve(ctx, raopts); err != nil {
		return map[string]any{}, err
	}

	return rawExtensionToMap(ra.Status)
}
