package widgets

import (
	"fmt"
	"net/http"

	"github.com/krateoplatformops/plumbing/maps"
	"github.com/krateoplatformops/snowplow/internal/dynamic"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeschema "k8s.io/apimachinery/pkg/runtime/schema"
)

func getWidgetData(obj map[string]any, key string) (map[string]any, error) {
	data, ok, err := maps.NestedMap(obj, "spec", key)
	if err != nil {
		return map[string]any{}, err
	}
	if !ok {
		name := dynamic.GetName(obj)
		namespace := dynamic.GetNamespace(obj)
		gv, _ := runtimeschema.ParseGroupVersion(dynamic.GetAPIVersion(obj))
		err := &apierrors.StatusError{
			ErrStatus: metav1.Status{
				Status: metav1.StatusFailure,
				Code:   http.StatusNotFound,
				Reason: metav1.StatusReasonNotFound,
				Details: &metav1.StatusDetails{
					Group: gv.Group,
					Kind:  dynamic.GetKind(obj),
					Name:  name,
				},
				Message: fmt.Sprintf("spec %q not found in %s @ %s", key, name, namespace),
			}}
		return map[string]any{}, err
	}

	return data, nil
}

/*
func mapToRawExtension(m map[string]any) (*runtime.RawExtension, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &runtime.RawExtension{
		Raw: data,
	}, nil
}
*/
