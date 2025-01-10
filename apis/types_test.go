package apis

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetTemplateKind(t *testing.T) {
	cases := []struct {
		gr   schema.GroupResource
		want TemplateKind
	}{
		{
			schema.GroupResource{Group: "templates.krateo.io", Resource: "forms"},
			FormTemplate,
		},
		{
			schema.GroupResource{Group: "templates.krateo.io", Resource: "customforms"},
			CustomFormTemplate,
		},
		{
			schema.GroupResource{Group: "templates.krateo.io", Resource: "widgets"},
			WidgetTemplate,
		},
		{
			schema.GroupResource{Group: "templates.krateo.io", Resource: "collections"},
			CollectionTemplate,
		},
		{
			schema.GroupResource{Group: "core.krateo.io", Resource: "collections"},
			Unknown,
		},
	}

	for _, tc := range cases {
		got := GetTemplateKind(tc.gr)
		if got != tc.want {
			t.Fatalf("expected template kind [%v], got: %v", tc.want, got)
		}
	}
}
