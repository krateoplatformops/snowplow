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
