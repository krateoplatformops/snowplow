package dispatchers

import (
	"net/http"
)

func All(authnNS string) map[string]http.Handler {
	return map[string]http.Handler{
		"restactions.templates.krateo.io": RESTAction(authnNS),
		"widgets.templates.krateo.io":     Widgets(authnNS),
	}
}
