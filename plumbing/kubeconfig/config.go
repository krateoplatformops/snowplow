package kubeconfig

import (
	"context"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClientConfig(ctx context.Context, ep endpoints.Endpoint) (*rest.Config, error) {
	dat, err := Marshal(&ep)
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

	log := xcontext.Logger(ctx)
	traceId := xcontext.TraceId(ctx, true)

	if env.True("DEBUG") {
		res.Wrap(newDebuggingRoundTripper(log, traceId, env.True("TRACE")))
	}

	return res, nil
}
