package apis

import (
	templates "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type TemplateKind int

const (
	Unknown TemplateKind = iota
	CustomFormTemplate
	FormTemplate
	CollectionTemplate
	WidgetTemplate
)

func GetTemplateKind(gr schema.GroupResource) TemplateKind {
	if gr.Group != templates.SchemeGroupVersion.Group {
		return Unknown
	}

	switch gr.Resource {
	case "forms":
		return FormTemplate
	case "customforms":
		return CustomFormTemplate
	case "collections":
		return CollectionTemplate
	case "widgets":
		return WidgetTemplate
	}

	return Unknown
}
