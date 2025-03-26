package api

import (
	"context"
	"log/slog"
	"net/http"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	httpcall "github.com/krateoplatformops/snowplow/plumbing/http/request"
	"github.com/krateoplatformops/snowplow/plumbing/jqutil"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func createRequestOptions(ctx context.Context, in *templates.API, dict map[string]any) (all []httpcall.RequestOptions) {
	it := ""
	if in.DependsOn != nil {
		it = ptr.Deref(in.DependsOn.Iterator, "")
	}

	log := xcontext.Logger(ctx)

	if len(it) == 0 {
		all = make([]httpcall.RequestOptions, 0, 1)
		el := createRequestOption(in, dict)
		all = append(all, el)
		return
	}

	all = []httpcall.RequestOptions{}

	action := func(sa any) error {
		el := createRequestOption(in, sa)
		all = append(all, el)
		return nil
	}

	err := jqutil.ForEach(context.TODO(), jqutil.EvalOptions{Query: it, Unquote: true, Data: dict}, action)
	if err != nil {
		log.Error("unable to execute iterator", slog.String("query", it), slog.Any("err", err))
	}

	return all
}

func createRequestOption(in *templates.API, ds any) (out httpcall.RequestOptions) {
	out.ContinueOnError = ptr.Deref(in.ContinueOnError, false)
	out.ErrorKey = ptr.Deref(in.ErrorKey, "error")

	out.Path = evalJQ(in.Path, ds)
	out.Verb = ptr.To(ptr.Deref(in.Verb, http.MethodGet))

	if in.Payload != nil {
		out.Payload = ptr.To(evalJQ(*in.Payload, ds))
	}

	if in.Headers != nil {
		out.Headers = make([]string, 0, len(in.Headers))
		//copy(el.Headers, in.Headers)
		for _, h := range in.Headers {
			out.Headers = append(out.Headers, evalJQ(h, ds))
		}
	}

	return
}

func evalJQ(q string, ds any) string {
	q, ok := jqutil.MaybeQuery(q)
	if !ok {
		return q
	}

	out, err := jqutil.Eval(context.TODO(), jqutil.EvalOptions{Query: q, Unquote: true, Data: ds})
	if err != nil {
		out = err.Error()
	}

	return out
}
