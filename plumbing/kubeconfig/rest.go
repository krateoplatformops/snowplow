package kubeconfig

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"

	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/env"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
)

func NewClientConfig(ctx context.Context, ep endpoints.Endpoint) (*rest.Config, error) {
	//ep.ServerURL = "https://kubernetes.default.svc"

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

func newDebuggingRoundTripper(log *slog.Logger, traceId string, verbose bool) transport.WrapperFunc {
	return func(rt http.RoundTripper) http.RoundTripper {
		return &debuggingRoundTripper{
			delegatedRoundTripper: rt,
			traceId:               traceId,
			log:                   log,
			verbose:               verbose,
		}
	}
}

type debuggingRoundTripper struct {
	delegatedRoundTripper http.RoundTripper
	traceId               string
	log                   *slog.Logger
	verbose               bool
}

func (rt *debuggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	traceId := req.Header.Get(xcontext.LabelKrateoTraceId)
	if len(traceId) == 0 {
		traceId = rt.traceId // xcontext.TraceId(req.Context(), true)
	}
	req.Header.Set(xcontext.LabelKrateoTraceId, traceId)

	b, err := httputil.DumpRequestOut(req, rt.verbose)
	if err != nil {
		return nil, err
	}
	rt.log.Debug("request details", "wire", string(b))

	resp, err := rt.delegatedRoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	withBody := req.URL.Query().Get("watch") != "true"
	withBody = withBody && rt.verbose

	b, err = httputil.DumpResponse(resp, withBody)
	if err != nil {
		return nil, err
	}
	rt.log.Debug("response details", "wire", string(b))

	return resp, err
}
