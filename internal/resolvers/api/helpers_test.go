//go:build unit
// +build unit

package api

import (
	"testing"
)

func TestNestedSliceNoCopy(t *testing.T) {
	tests := []struct {
		name   string
		obj    map[string]any
		fields []string
		expect []any
		found  bool
		err    bool
	}{
		{
			name: "valid nested slice",
			obj: map[string]any{
				"data": map[string]any{
					"items": []any{1, 2, 3},
				},
			},
			fields: []string{"data", "items"},
			expect: []any{1, 2, 3},
			found:  true,
			err:    false,
		},
		{
			name: "field not found",
			obj: map[string]any{
				"data": map[string]any{},
			},
			fields: []string{"data", "items"},
			expect: nil,
			found:  false,
			err:    false,
		},
		{
			name: "not a slice",
			obj: map[string]any{
				"data": map[string]any{
					"items": "not a slice",
				},
			},
			fields: []string{"data", "items"},
			expect: nil,
			found:  false,
			err:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, found, err := nestedSliceNoCopy(tc.obj, tc.fields...)

			if found != tc.found {
				t.Errorf("expected found %v, got %v", tc.found, found)
			}
			if (err != nil) != tc.err {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			if !deepEqual(result, tc.expect) {
				t.Errorf("expected %v, got %v", tc.expect, result)
			}
		})
	}
}

func TestNestedMapNoCopy(t *testing.T) {
	tests := []struct {
		name   string
		obj    map[string]any
		fields []string
		expect map[string]any
		found  bool
		err    bool
	}{
		{
			name: "valid nested map",
			obj: map[string]any{
				"data": map[string]any{
					"config": map[string]any{"key": "value"},
				},
			},
			fields: []string{"data", "config"},
			expect: map[string]any{"key": "value"},
			found:  true,
			err:    false,
		},
		{
			name: "field not found",
			obj: map[string]any{
				"data": map[string]any{},
			},
			fields: []string{"data", "config"},
			expect: nil,
			found:  false,
			err:    false,
		},
		{
			name: "not a map",
			obj: map[string]any{
				"data": map[string]any{
					"config": "not a map",
				},
			},
			fields: []string{"data", "config"},
			expect: nil,
			found:  false,
			err:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, found, err := nestedMapNoCopy(tc.obj, tc.fields...)

			if found != tc.found {
				t.Errorf("expected found %v, got %v", tc.found, found)
			}
			if (err != nil) != tc.err {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			if !deepEqual(result, tc.expect) {
				t.Errorf("expected %v, got %v", tc.expect, result)
			}
		})
	}
}

func TestNestedFieldNoCopy(t *testing.T) {
	tests := []struct {
		name   string
		obj    map[string]any
		fields []string
		expect any
		found  bool
		err    bool
	}{
		{
			name: "valid nested field",
			obj: map[string]any{
				"data": map[string]any{
					"config": "value",
				},
			},
			fields: []string{"data", "config"},
			expect: "value",
			found:  true,
			err:    false,
		},
		{
			name: "field not found",
			obj: map[string]any{
				"data": map[string]any{},
			},
			fields: []string{"data", "config"},
			expect: nil,
			found:  false,
			err:    false,
		},
		{
			name: "not a map",
			obj: map[string]any{
				"data": "not a map",
			},
			fields: []string{"data", "config"},
			expect: nil,
			found:  false,
			err:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, found, err := NestedFieldNoCopy(tc.obj, tc.fields...)

			if found != tc.found {
				t.Errorf("expected found %v, got %v", tc.found, found)
			}
			if (err != nil) != tc.err {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			if !deepEqual(result, tc.expect) {
				t.Errorf("expected %v, got %v", tc.expect, result)
			}
		})
	}
}
