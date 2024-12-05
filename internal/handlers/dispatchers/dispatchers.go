package dispatchers

import (
	"net/http"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Empty() map[schema.GroupVersionResource]http.Handler {
	return map[schema.GroupVersionResource]http.Handler{}
}
