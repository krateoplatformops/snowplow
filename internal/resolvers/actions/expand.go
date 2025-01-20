package actions

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

func Expand(ctx context.Context, dict map[string]any, in ...*templates.ActionTemplateIterator) (all []*templates.ActionTemplate) {
	if len(in) == 0 {
		return all
	}

	for _, el := range in {
		all = append(all, renderOne(ctx, renderOneOptions{
			in:   el,
			dict: dict,
		})...)
	}

	return all
}

type renderOneOptions struct {
	in   *templates.ActionTemplateIterator
	dict map[string]any
}

func renderOne(ctx context.Context, opts renderOneOptions) (all []*templates.ActionTemplate) {
	if opts.in == nil {
		return all
	}

	log := xcontext.Logger(ctx)

	tpl := xcontext.JQTemplate(ctx)
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
		all = append(all, &templates.ActionTemplate{
			Name:       render(i, opts.in.Template.Name, opts.dict),
			Namespace:  render(i, opts.in.Template.Namespace, opts.dict),
			Resource:   render(i, opts.in.Template.Resource, opts.dict),
			APIVersion: render(i, opts.in.Template.APIVersion, opts.dict),
			Verb:       opts.in.Template.Verb,
		})
	}

	return
}
