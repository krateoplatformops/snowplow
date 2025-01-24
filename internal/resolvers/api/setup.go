package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	httpcall "github.com/krateoplatformops/snowplow/plumbing/http/request"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func createRequestOptions(ctx context.Context, in *templates.API, dict map[string]any) (all []httpcall.RequestOptions) {
	log := xcontext.Logger(ctx)

	tpl := xcontext.JQ(ctx)
	if tpl == nil {
		log.Error("missing jq engine")
		return []httpcall.RequestOptions{}
	}

	var it string
	if in.DependsOn != nil {
		it = ptr.Deref(in.DependsOn.Iterator, "")
	}

	tot := 1
	if len(it) > 0 {
		q := fmt.Sprintf("${ %s | length }", it)
		count, err := tpl.Execute(q, dict)
		if err != nil {
			log.Error("unable to execute jq query", slog.String("query", q), slog.Any("err", err))
			count = "1"
		}

		tot, err = strconv.Atoi(count)
		if err != nil {
			log.Warn("atoi failure, assuming count=1", slog.Any("err", err))
			tot = 1
		}
	}
	log.Debug("resolved iterator", slog.String("name", in.Name), slog.Int("count", tot))

	render := func(i int, s string, ds map[string]any) string {
		exp := hackQueryFn(it, i, s)
		out, err := tpl.Execute(exp, ds)
		if err != nil {
			out = err.Error()
		}
		return out
	}

	all = make([]httpcall.RequestOptions, 0, tot)
	for i := 0; i < tot; i++ {
		el := httpcall.RequestOptions{
			Path: render(i, in.Path, dict),
			Verb: ptr.To(ptr.Deref(in.Verb, http.MethodGet)),
		}

		if in.Payload != nil {
			el.Payload = ptr.To(ptr.Deref(in.Payload, ""))
		}

		if in.Headers != nil {
			el.Headers = make([]string, len(in.Headers))
			copy(el.Headers, in.Headers)
		}

		all = append(all, el)
	}

	return all
}

func hackQueryFn(it string, i int, q string) string {
	if len(it) == 0 {
		return q
	}

	el := fmt.Sprintf("%s[%d]", it, i)
	q = strings.Replace(q, "${", fmt.Sprintf("${ %s | ", el), 1)
	return q
}
