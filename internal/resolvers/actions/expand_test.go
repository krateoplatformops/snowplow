package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	xcontext "github.com/krateoplatformops/snowplow/plumbing/context"
	"github.com/krateoplatformops/snowplow/plumbing/ptr"
)

func ExampleExpand() {
	const (
		dummy = `
{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "labels": {
                    "k8s-app": "kube-dns",
                    "pod-template-hash": "668d6bf9bc"
                },
                "name": "aaaaaa",
                "namespace": "kube-system"
			}
		},
		{
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "name": "bbbbbb",
                "namespace": "demo-system"
			}
		},
		{
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "name": "cccccc",
                "namespace": "test-system"
			}
		}
	]
}`
	)

	dict := map[string]any{}
	dec := json.NewDecoder(strings.NewReader(dummy))
	dec.Decode(&dict)

	ctx := xcontext.BuildContext(context.Background(),
		xcontext.WithLogger(nil),
		xcontext.WithJQTemplate(),
	)

	all := Expand(ctx, dict, &v1.ActionTemplateIterator{
		Iterator: ptr.To(".items"),
		Template: &v1.ActionTemplate{
			ID:         "dummy",
			Name:       `${ .metadata.name + "-card" }`,
			Namespace:  `${ .metadata.namespace }`,
			APIVersion: "composition.krateo.io/v1-1-3",
			Resource:   "fireworksapps",
			Verb:       "PUT",
		},
	})

	for _, el := range all {
		fmt.Println("Name:", el.Name, "Namespace:", el.Namespace, "Verb:", el.Verb)
	}

	// Output:
	// Name: aaaaaa-card Namespace: kube-system Verb: PUT
	// Name: bbbbbb-card Namespace: demo-system Verb: PUT
	// Name: cccccc-card Namespace: test-system Verb: PUT
}
