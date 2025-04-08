package dispatchers

import (
	"net/http"
)

func All(skip bool) map[string]http.Handler {
	if skip {
		return map[string]http.Handler{}
	}

	return map[string]http.Handler{
		"restactions.templates.krateo.io": RESTAction(),
		"widgets.templates.krateo.io":     Widgets(),
	}
}
