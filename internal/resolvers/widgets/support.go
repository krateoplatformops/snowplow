package widgets

import (
	"fmt"
	"strings"

	"github.com/krateoplatformops/plumbing/maps"
)

func toAnySlice[T any](in []T) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func nestedString(obj map[string]any, fields ...string) (string, error) {
	val, found := maps.NestedValue(obj, fields)
	if !found {
		return "", nil
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%v access error: %v is of the type %T, expected string",
			strings.Join(fields, "."), val, val)
	}
	return s, nil
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
