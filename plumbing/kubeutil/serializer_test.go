//go:build unit
// +build unit

package kubeutil

import (
	"bytes"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FakeObject simula un oggetto Kubernetes serializzabile
type FakeObject struct{}

func (f *FakeObject) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }
func (f *FakeObject) DeepCopyObject() runtime.Object   { return &FakeObject{} }

func TestToYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       runtime.Object
		expectError bool
	}{
		{
			name:        "Valid object should serialize correctly",
			input:       &FakeObject{},
			expectError: false,
		},
		{
			name:        "Nil object should not return error",
			input:       nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ToYAML(&buf, tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if err == nil && buf.Len() == 0 {
				t.Errorf("expected YAML output, but got empty buffer")
			}
		})
	}
}
