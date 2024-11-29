package kubeconfig

import (
	"context"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClientConfig(ctx context.Context, ep *endpoints.Endpoint) (*rest.Config, error) {
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

	res.Wrap(Transport)

	return res, nil
}
