package widgets

import (
	"context"
	"log/slog"

	xcontext "github.com/krateoplatformops/plumbing/context"
	xenv "github.com/krateoplatformops/plumbing/env"
	"github.com/krateoplatformops/plumbing/maps"
	crdschema "github.com/krateoplatformops/snowplow/internal/resolvers/crds/schema"
	"github.com/krateoplatformops/snowplow/internal/resolvers/resourcesrefs"
	"github.com/krateoplatformops/snowplow/internal/resolvers/widgets/apiref"
	"github.com/krateoplatformops/snowplow/internal/resolvers/widgets/data"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

const (
	widgetDataKey         = "widgetData"
	widgetDataTemplateKey = "widgetDataTemplate"
	resourcesRefsKey      = "resourcesRefs"
)

type Widget = unstructured.Unstructured

type ResolveOptions struct {
	In         *Widget
	RC         *rest.Config
	AuthnNS    string
	Username   string
	UserGroups []string
}

func Resolve(ctx context.Context, opts ResolveOptions) (*Widget, error) {
	log := xcontext.Logger(ctx)

	src, err := getWidgetData(opts.In.Object, widgetDataKey)
	if err != nil {
		return opts.In, err
	}
	log.Debug("WidgetData before resolving API Ref", slog.Any(widgetDataKey, src))

	dict, err := apiref.Resolve(ctx, apiref.ResolveOptions{
		RC:         opts.RC,
		Widget:     opts.In,
		AuthnNS:    opts.AuthnNS,
		Username:   opts.Username,
		UserGroups: opts.UserGroups,
	})
	if err != nil {
		log.Error("unable to resolve api reference", slog.Any("err", err))
		maps.SetNestedField(opts.In.Object, err.Error(), "status", "error")
		return opts.In, err
	}
	log.Debug("resolved API reference", slog.Any("apiRef", dict))

	evals, err := data.ResolveTemplates(ctx, data.ResolveOptions{
		Widget: opts.In,
		Dict:   dict,
	})
	if err != nil {
		log.Error("unable to resolve widget data templates", slog.Any("err", err))
		maps.SetNestedField(opts.In.Object, err.Error(), "status", "error")
		return opts.In, err
	}
	log.Debug("widgetDataTemplate array after evaluation", slog.Any("widgetDataTemplate", evals))

	for _, el := range evals {
		fields := maps.ParsePath(el.Path)
		if len(fields) == 0 {
			continue
		}

		err = maps.SetNestedValue(src, fields, el.Value)
		if err != nil {
			log.Error("unable to set nested field value", slog.Any("err", err))
			return opts.In, err
		}
	}

	resrefs, err := resourcesrefs.Resolve(ctx, resourcesrefs.ResolveOptions{
		RC: opts.RC, Widget: opts.In,
		AuthnNS:  opts.AuthnNS,
		Username: opts.Username, UserGroups: opts.UserGroups,
	})
	if err != nil {
		maps.SetNestedField(opts.In.Object, err.Error(), "status", "error")
		return opts.In, err
	}

	err = unstructured.SetNestedMap(opts.In.Object, src, "status", widgetDataKey)
	if err != nil {
		return opts.In, err
	}

	if len(resrefs) > 0 {
		err = unstructured.SetNestedSlice(opts.In.Object, resrefs, "status", resourcesRefsKey)
		if err != nil {
			return opts.In, err
		}
	}

	if xenv.TestMode() {
		err = crdschema.ValidateObjectStatus(ctx, opts.RC, opts.In.Object)
	} else {
		err = crdschema.ValidateObjectStatus(ctx, nil, opts.In.Object)
	}
	if err != nil {
		maps.SetNestedField(opts.In.Object, err.Error(), "status", "error")
		return opts.In, err
	}

	return opts.In, nil
}
