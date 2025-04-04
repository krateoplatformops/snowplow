package widgets

import (
	"context"
	"fmt"
	"sort"

	v1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"github.com/krateoplatformops/snowplow/internal/resolvers/restactions"
	"github.com/krateoplatformops/snowplow/plumbing/maps"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

const (
	widgetDataKey = "widgetData"
	apiKey        = "api"
)

type Widget = unstructured.Unstructured

type ResolveOptions struct {
	In *Widget
	RC *rest.Config
}

func Resolve(ctx context.Context, opts ResolveOptions) (*Widget, error) {

	widgetData, ok, err := unstructured.NestedMap(opts.In.Object, "spec", widgetDataKey)
	if err != nil {
		return opts.In, err
	}
	if !ok {
		return opts.In, fmt.Errorf("missing %q in spec (%s @ %s)",
			widgetDataKey, opts.In.GetName(), opts.In.GetNamespace())
	}

	status := runtime.DeepCopyJSON(widgetData)

	paths := maps.LeafPaths(status, "")
	sort.Strings(paths)

	for _, path := range paths {
		fields := maps.ParsePath(path)

		value, found := maps.NestedValue(opts.In.Object, fields)
		if !found {
			continue
		}

		fmt.Printf("Path: %s, Value: %v\n", path, value)
		// if the value is string, we can try to evaluate a JQ expression
		if strValue, ok := value.(string); ok {
			fmt.Printf("  ==> evaluate JQ: %s\n", strValue)
		}
	}

	unstructured.SetNestedMap(opts.In.Object, status, "status")

	return opts.In, nil
}

func resolveRESTActionRef(ctx context.Context, opts ResolveOptions) error {
	api, ok, err := unstructured.NestedMap(opts.In.Object, "spec", apiKey)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var ra v1.RESTAction
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(api, &ra)
	if err != nil {
		return err
	}

	restactions.Resolve(ctx, restactions.ResolveOptions{
		In: &ra,
		// TODO other params
	})
}
