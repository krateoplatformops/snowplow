package api

import (
	"context"
	"fmt"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"k8s.io/client-go/rest"
)

type endpointReferenceMapper struct {
	authnNS  string
	username string
	rc       *rest.Config
}

func (m *endpointReferenceMapper) resolveAll(ctx context.Context, items []*templates.API) (map[string]endpoints.Endpoint, error) {
	endpointsMap := make(map[string]endpoints.Endpoint, len(items))
	for _, el := range items {
		res, err := m.resolveOne(ctx, el.EndpointRef)
		if err != nil {
			return endpointsMap, err
		}

		endpointsMap[el.Name] = res
	}

	return endpointsMap, nil
}

func (m *endpointReferenceMapper) resolveOne(ctx context.Context, ref *templates.Reference) (endpoints.Endpoint, error) {
	ep := endpoints.Endpoint{}

	isInternal := false
	if ref == nil {
		ref = &templates.Reference{
			Namespace: m.authnNS,
			Name:      fmt.Sprintf("%s-clientconfig", kubeutil.MakeDNS1123Compatible(m.username)),
		}
		isInternal = true
	}

	ep, err := endpoints.FromSecret(ctx, m.rc, ref.Name, ref.Namespace)
	if err != nil {
		return ep, err
	}
	if isInternal && !env.TestMode() {
		ep.ServerURL = "https://kubernetes.default.svc"
	}

	return ep, nil
}

type resolveEndpointReferenceOptions struct {
	Reference *templates.Reference
	AuthnNS   string
	Username  string
	RC        *rest.Config
}

func resolveEndpointReference(ctx context.Context, opts resolveEndpointReferenceOptions) (endpoints.Endpoint, error) {
	ep := endpoints.Endpoint{}

	isInternal := false
	if opts.Reference == nil {
		opts.Reference = &templates.Reference{
			Namespace: opts.AuthnNS,
			Name:      fmt.Sprintf("%s-clientconfig", kubeutil.MakeDNS1123Compatible(opts.Username)),
		}
		isInternal = true
	}

	ep, err := endpoints.FromSecret(ctx, opts.RC, opts.Reference.Name, opts.Reference.Namespace)
	if err != nil {
		return ep, err
	}
	if isInternal && !env.TestMode() {
		ep.ServerURL = "https://kubernetes.default.svc"
	}

	return ep, nil
}
