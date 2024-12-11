package dispatchers

import (
	"net/http"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func All(skip bool) map[schema.GroupVersionResource]http.Handler {
	if skip {
		return map[schema.GroupVersionResource]http.Handler{}
	}

	return map[schema.GroupVersionResource]http.Handler{
		{
			Group:    "templates.krateo.io",
			Version:  "v1alpha1",
			Resource: "customforms",
		}: CustomForm(),
	}
}
