package data

import (
	"context"
	"strings"

	"github.com/krateoplatformops/plumbing/jqutil"
	"github.com/krateoplatformops/plumbing/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	widgetDataTemplateKey = "widgetDataTemplate"
)

type EvalResult struct {
	Path  string
	Value any
}

type ResolveOptions struct {
	Widget *unstructured.Unstructured
	Dict   map[string]any
}

func ResolveTemplates(ctx context.Context, opts ResolveOptions) ([]EvalResult, error) {
	widgetDataTemplate, _, err := maps.NestedSliceNoCopy(opts.Widget.Object, "spec", widgetDataTemplateKey)
	if err != nil {
		return []EvalResult{}, err
	}

	res := make([]EvalResult, 0, len(widgetDataTemplate))

	for _, el := range widgetDataTemplate {
		item, ok := el.(map[string]any)
		if !ok {
			continue
		}

		path, ok := item["forPath"].(string)
		if !ok || path == "" {
			continue
		}
		path = strings.TrimSpace(path)

		expression, ok := item["expression"].(string)
		if !ok || expression == "" {
			continue
		}

		if path[0] == '.' {
			path = path[1:]
		}

		s := expression
		if exp, ok := jqutil.MaybeQuery(expression); ok {

			s, err = jqutil.Eval(ctx, jqutil.EvalOptions{
				Query: exp, Data: opts.Dict, Unquote: true,
			})
			if err != nil {
				return res, err
			}
		}

		res = append(res, EvalResult{
			Path:  path,
			Value: jqutil.InferType(s),
		})
	}

	return res, nil
}
