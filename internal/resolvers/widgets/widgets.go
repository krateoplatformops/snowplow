package widgets

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	crdschema "github.com/krateoplatformops/snowplow/internal/resolvers/crds/schema"
	xenv "github.com/krateoplatformops/snowplow/plumbing/env"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

const (
	widgetDataKey = "widgetData"
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
	widgetData, err := getWidgetData(opts.In.Object, widgetDataKey)
	if err != nil {
		return opts.In, err
	}

	dict, err := resolveRESTActionRef(ctx, opts)
	if err != nil {
		return opts.In, err
	}
	spew.Dump(dict)

	evalJQ(widgetData, dict)

	err = unstructured.SetNestedMap(opts.In.Object, widgetData, "status")
	if err != nil {
		return opts.In, err
	}

	if xenv.TestMode() {
		err = crdschema.ValidateObjectStatus(ctx, opts.RC, opts.In.Object)
	} else {
		err = crdschema.ValidateObjectStatus(ctx, nil, opts.In.Object)
	}
	if err != nil {
		return opts.In, err
	}

	return opts.In, nil
}
