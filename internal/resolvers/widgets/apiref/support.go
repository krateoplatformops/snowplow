package apiref

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
)

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
