package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

func TestValidate(t *testing.T) {

	validSchema := &apiextensions.CustomResourceValidation{
		OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensions.JSONSchemaProps{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
		},
	}

	validDocument := []byte(`{"name": "John", "age": 30}`)
	invalidDocument := []byte(`{"name": "John", "age": "thirty"}`)

	t.Run("valid document", func(t *testing.T) {
		var jsonObj map[string]any
		err := json.Unmarshal(validDocument, &jsonObj)
		assert.NoError(t, err)
		err = validateCustomResource(validSchema, jsonObj)
		assert.NoError(t, err)
	})

	t.Run("invalid document type mismatch", func(t *testing.T) {
		var jsonObj map[string]any
		err := json.Unmarshal(invalidDocument, &jsonObj)
		assert.NoError(t, err)
		err = validateCustomResource(validSchema, jsonObj)
		assert.Error(t, err)
	})
}
