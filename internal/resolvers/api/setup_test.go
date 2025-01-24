package api

import (
	"context"
	"fmt"
	"net/http"

	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func Example_expandIterator2() {
	dict := map[string]any{
		"namespaces": []any{"demo-system", "krateo-system", "example-system"},
	}

	ctx := xcontext.BuildContext(context.Background(), xcontext.WithJQ())
	all := createRequestOptions(ctx, &templates.API{
		Name: "example",
		Path: `${ "/api/v1/namespaces/" + (.) + "/pods" }`,
		DependsOn: &templates.Dependency{
			Name:     "namespaces",
			Iterator: ptr.To(".[]"),
		},
	}, dict)

	for _, el := range all {
		fmt.Println(ptr.Deref(el.Verb, http.MethodGet), el.Path)
		for _, y := range el.Headers {
			fmt.Printf("  %s\n", y)
		}
	}

	// Output:
	// GET /api/v1/namespaces/demo-system/pods
	// GET /api/v1/namespaces/krateo-system/pods
	// GET /api/v1/namespaces/example-system/pods
}

func Example_expandIterator2_no_iter() {
	dict := map[string]any{
		"namespaces": []any{"demo-system", "krateo-system", "example-system"},
	}

	ctx := xcontext.BuildContext(context.Background(), xcontext.WithJQ())
	all := createRequestOptions(ctx, &templates.API{
		Name: "example",
		Path: `${ "/api/v1/namespaces/" + (.namespaces[2]) + "/pods" }`,
		Verb: ptr.To(string(http.MethodPost)),
		DependsOn: &templates.Dependency{
			Name: "namespaces",
		},
	}, dict)

	for _, el := range all {
		fmt.Println(ptr.Deref(el.Verb, http.MethodGet), el.Path)
		for _, y := range el.Headers {
			fmt.Printf("  %s\n", y)
		}
	}

	// Output:
	// POST /api/v1/namespaces/example-system/pods
}
