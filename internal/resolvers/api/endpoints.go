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

func (m *endpointReferenceMapper) resolveOne(ctx context.Context, ref *templates.Reference) (endpoints.Endpoint, error) {
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
