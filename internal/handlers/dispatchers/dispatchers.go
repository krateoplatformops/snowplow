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
			Version:  "v1",
			Resource: "customforms",
		}: CustomForm(),
		{
			Group:    "templates.krateo.io",
			Version:  "v1",
			Resource: "collections",
		}: Collection(),
		{
			Group:    "templates.krateo.io",
			Version:  "v1",
			Resource: "widgets",
		}: Widget(),
		{
			Group:    "templates.krateo.io",
			Version:  "v1",
			Resource: "restactions",
		}: RESTAction(),
	}
}
