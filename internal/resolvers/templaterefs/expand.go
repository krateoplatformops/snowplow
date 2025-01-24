package templaterefs

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

type ExpandOptions struct {
	TemplateIterators []*templates.TemplateIterator
	Dict              map[string]any
}

func Expand(ctx context.Context, opts ExpandOptions) (all []*templates.ObjectReference) {
	if len(opts.TemplateIterators) == 0 {
		return all
	}

	for _, el := range opts.TemplateIterators {
		all = append(all, renderOne(ctx, renderOneOptions{
			in:   el,
			dict: opts.Dict,
		})...)
	}

	return all
}

type renderOneOptions struct {
	in   *templates.TemplateIterator
	dict map[string]any
}

func renderOne(ctx context.Context, opts renderOneOptions) (all []*templates.ObjectReference) {
	if opts.in == nil {
		return all
	}

	log := xcontext.Logger(ctx)

	tpl := xcontext.JQ(ctx)
	if tpl == nil {
		log.Error("missing jq template engine")
		return all
	}

	it := ptr.Deref(opts.in.Iterator, "")

	tot := 1
	hasIter := (len(it) > 0)
	if hasIter {
		len, err := tpl.Execute(fmt.Sprintf("${ %s | length }", it), opts.dict)
		if err != nil {
			log.Error("unable to execute jq template: %s", slog.Any("err", err))
		}

		tot, err = strconv.Atoi(len)
		if err != nil {
			log.Error("atoi failure", slog.Any("err", err))
			tot = 1
		}
	}

	hackQueryFn := func(i int, q string) string {
		if !hasIter {
			return q
		}

		el := fmt.Sprintf("%s[%d]", it, i)
		q = strings.Replace(q, "${", fmt.Sprintf("${ %s | ", el), 1)
		return q
	}

	render := func(i int, s string, ds map[string]any) string {
		out, err := tpl.Execute(hackQueryFn(i, s), ds)
		if err != nil {
			out = err.Error()
		}
		return out
	}

	for i := 0; i < tot; i++ {
		all = append(all, &templates.ObjectReference{
			Reference: templates.Reference{
				Name:      render(i, opts.in.Template.Name, opts.dict),
				Namespace: render(i, opts.in.Template.Namespace, opts.dict),
			},
			Resource:   render(i, opts.in.Template.Resource, opts.dict),
			APIVersion: render(i, opts.in.Template.APIVersion, opts.dict),
		})
	}

	return
}
