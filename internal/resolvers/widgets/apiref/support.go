package apiref

import (
	"encoding/json"

	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func convertToRESTAction(api map[string]any) (templatesv1.RESTAction, error) {
	dat, err := json.Marshal(api)
	if err != nil {
		return templatesv1.RESTAction{}, err
	}

	var ra templatesv1.RESTAction
	err = json.Unmarshal(dat, &ra)

	return ra, err
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
