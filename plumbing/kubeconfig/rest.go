package kubeconfig

import (
	"context"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"github.com/krateoplatformops/snowplow/plumbing/server/traceid"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClientConfig(ctx context.Context, ep *endpoints.Endpoint) (*rest.Config, error) {
	ep.Debug = env.Bool("DEBUG", false)
	ep.ServerURL = "https://kubernetes.default.svc"

	dat, err := Marshal(ep)
	if err != nil {
		return nil, err
	}

	ccf, err := clientcmd.NewClientConfigFromBytes(dat)
	if err != nil {
		return nil, err
	}

	res, err := ccf.ClientConfig()
	if err != nil {
		return nil, err
	}
	res.Wrap(traceid.Transport)

	return res, nil
}
