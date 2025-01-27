package templaterefs

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
		xcontext.WithJQ(),
	)

	all := Expand(ctx, ExpandOptions{
		TemplateIterators: []*v1.TemplateIterator{
			{
				Iterator: ptr.To(".items"),
				Template: &v1.ObjectReference{
					Reference: v1.Reference{
						Name:      `${ .metadata.name + "-card" }`,
						Namespace: `${ .metadata.namespace }`,
					},
					APIVersion: "templates.krateo.io/v1alpha1",
					Resource:   "widgets",
				},
			},
		},
		Dict: dict,
	})

	for _, el := range all {
		fmt.Println("Name:", el.Name, "Namespace:", el.Namespace)
	}

	// Output:
	// Name: aaaaaa-card Namespace: kube-system
	// Name: bbbbbb-card Namespace: demo-system
	// Name: cccccc-card Namespace: test-system
}
