package widgets

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
	"github.com/krateoplatformops/snowplow/plumbing/tmpl"
)

func Resolve(ctx context.Context, app *templates.AppTemplate, dict map[string]any) (all []map[string]string) {
	if app == nil {
		return all
	}

	tpl := xcontext.JQTemplate(ctx)
	if tpl == nil {
		log := xcontext.Logger(ctx)
		log.Error("unable to resolve customform app template",
			slog.String("err", "missing jq template engine"))
		return
	}

	return render(ctx, renderOptions{
		tpl:  tpl,
		in:   app,
		dict: dict,
	})
}

type renderOptions struct {
	tpl  tmpl.JQTemplate
	in   *templates.AppTemplate
	dict map[string]any
}

func render(ctx context.Context, opts renderOptions) (all []map[string]string) {
	if opts.in == nil {
		return all
	}

	it := ptr.Deref(opts.in.Iterator, "")

	tot := 1
	hasIter := (len(it) > 0)
	if hasIter {
		len, err := opts.tpl.Execute(fmt.Sprintf("${ %s | length }", it), opts.dict)
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable to execute jq template", slog.Any("err", err))
		}

		tot, err = strconv.Atoi(len)
		if err != nil {
			log := xcontext.Logger(ctx)
			log.Error("unable to convert len to int", slog.Any("err", err))
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
		out, err := opts.tpl.Execute(hackQueryFn(i, s), ds)
		if err != nil {
			out = err.Error()
		}
		return out
	}

	for i := 0; i < tot; i++ {
		traits := map[string]string{}
		for k, v := range opts.in.Template {
			traits[k] = render(i, v, opts.dict)
		}

		all = append(all, traits)
	}

	return
}
