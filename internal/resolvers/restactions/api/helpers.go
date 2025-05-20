package api

import (
	"fmt"
	"strings"
)

func nestedSliceNoCopy(obj map[string]any, fields ...string) ([]any, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}

	items, ok := val.([]any)
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected []any",
			jsonPath(fields), val, val)
	}

	return items, true, nil
}

// nestedMapNoCopy returns a map[string]interface{} value of a nested field.
// Returns false if value is not found and an error if not a map[string]interface{}.
func nestedMapNoCopy(obj map[string]any, fields ...string) (map[string]any, bool, error) {
	val, found, err := NestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}

	m, ok := val.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]any",
			jsonPath(fields), val, val)
	}

	return m, true, nil
}

// NestedFieldNoCopy returns a reference to a nested field.
// Returns false if value is not found and an error if unable
// to traverse obj.
//
// Note: fields passed to this function are treated as keys within the passed
// object; no array/slice syntax is supported.
func NestedFieldNoCopy(obj map[string]any, fields ...string) (any, bool, error) {
	var val interface{} = obj

	for i, field := range fields {
		if val == nil {
			return nil, false, nil
		}

		if m, ok := val.(map[string]any); ok {
			val, ok = m[field]
			if !ok {
				return nil, false, nil
			}
		} else {
			return nil, false,
				fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]any",
					jsonPath(fields[:i+1]), val, val)
		}
	}

	return val, true, nil
}

func jsonPath(fields []string) string {
	return "." + strings.Join(fields, ".")
}
