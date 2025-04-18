package data

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestResolveWidgetDataTemplates(t *testing.T) {
	tests := []struct {
		name        string
		obj         map[string]any
		dict        map[string]any
		want        []EvalResult
		expectedErr bool
	}{
		{
			name: "No template",
			obj: map[string]any{
				"spec": map[string]any{},
			},
			dict: map[string]any{},
			want: []EvalResult{},
		},
		{
			name: "Valid template with full path",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath":    "spec.widgetData.value",
							"expression": "${ .data.value }",
						},
					},
				},
			},
			dict: map[string]any{
				"data": map[string]any{
					"value": "test-value",
				},
			},
			want: []EvalResult{
				{Path: "spec.widgetData.value", Value: "test-value"},
			},
		},
		{
			name: "Valid template with relative path",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath":    "value",
							"expression": "${ .data.value }",
						},
					},
				},
			},
			dict: map[string]any{
				"data": map[string]any{
					"value": "test-value-xxx",
				},
			},
			want: []EvalResult{
				{Path: "value", Value: "test-value-xxx"},
			},
		},
		{
			name: "Valid template with starting dot in the path",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath":    ".value",
							"expression": "${ .data.value }",
						},
					},
				},
			},
			dict: map[string]any{
				"data": map[string]any{
					"value": "test-value-yyy",
				},
			},
			want: []EvalResult{
				{Path: "value", Value: "test-value-yyy"},
			},
		},
		{
			name: "Error during expression evaluation",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath":    "value",
							"expression": "${ .invalid }",
						},
					},
				},
			},
			dict:        map[string]any{},
			expectedErr: false,
			want: []EvalResult{
				{Path: "value", Value: "null"},
			},
		},
		{
			name: "Template with missing path",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"expression": ".data.value",
						},
					},
				},
			},
			dict: map[string]any{},
			want: []EvalResult{},
		},
		{
			name: "Template with missing expression",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath": "value",
						},
					},
				},
			},
			dict: map[string]any{},
			want: []EvalResult{},
		},
		{
			name: "Invalid template (not a map[string]string)",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						123,
					},
				},
			},
			dict: map[string]any{},
			want: []EvalResult{},
		},
		{
			name: "Invalid template (empty forPath)",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath":    "",
							"expression": ".data.value",
						},
					},
				},
			},
			dict: map[string]any{},
			want: []EvalResult{},
		},
		{
			name: "Invalid template (empty expression)",
			obj: map[string]any{
				"spec": map[string]any{
					"widgetDataTemplate": []any{
						map[string]any{
							"forPath":    "value",
							"expression": "",
						},
					},
				},
			},
			dict: map[string]any{},
			want: []EvalResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveTemplates(context.Background(), ResolveOptions{
				Widget: &unstructured.Unstructured{Object: tt.obj},
				Dict:   tt.dict,
			})

			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
