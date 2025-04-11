package widgets

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krateoplatformops/snowplow/internal/dynamic"
	"github.com/krateoplatformops/snowplow/plumbing/maps"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func rawExtensionToMap(raw *runtime.RawExtension) (map[string]any, error) {
	if raw == nil {
		return map[string]any{}, nil
	}

	var data []byte
	if raw.Raw != nil {
		data = raw.Raw
	} else if raw.Object != nil {
		var err error
		data, err = json.Marshal(raw.Object)
		if err != nil {
			return map[string]any{}, err
		}
	} else {
		return map[string]any{}, nil
	}

	var result map[string]any
	err := json.Unmarshal(data, &result)

	return result, err
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
