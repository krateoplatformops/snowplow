package schema

import (
	"fmt"

	"github.com/krateoplatformops/plumbing/maps"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func extractOpenAPISchemaFromCRD(crd map[string]any, version string) (*apiextensions.CustomResourceValidation, error) {
	versions, found, err := unstructured.NestedSlice(crd, "spec", "versions")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no versions found in CRD")
	}

	for _, v := range versions {
		versionMap, ok := v.(map[string]any)
		if !ok {
			continue
		}

		if name, found := versionMap["name"].(string); !found || name != version {
			continue
		}

		schemaData, exists, err := maps.NestedMap(versionMap,
			"schema", "openAPIV3Schema",
			"properties", "spec", "properties", widgetDataKey)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("schema OpenAPI v3 not found for version: %s", version)
		}

		return buildValidationFromSchemaData(schemaData)
	}

	return nil, fmt.Errorf("version [%s] not found in CRD schema", version)
}

/*
func dumpSchemaRecursively(s *apiextensions.JSONSchemaProps, prefix string) {
	if s == nil {
		fmt.Printf("%s<nil>\n", prefix)
		return
	}
	fmt.Printf("%sType=%v Properties=[", prefix, s.Type)
	for k := range s.Properties {
		fmt.Printf("%s%s ", k, "")
	}
	fmt.Printf("]\n")

	for name, prop := range s.Properties {
		dumpSchemaRecursively(&prop, prefix+"  "+name+".")
	}
	if s.Items != nil && s.Items.Schema != nil {
		dumpSchemaRecursively(s.Items.Schema, prefix+"Items.")
	}
}
*/

func buildValidationFromSchemaData(data map[string]interface{}) (*apiextensions.CustomResourceValidation, error) {
	// 1. From map to v1.JSONSchemaProps via YAML
	yml, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal to YAML: %w", err)
	}
	var schemaV1 apiextv1.JSONSchemaProps
	if err := yaml.Unmarshal(yml, &schemaV1); err != nil {
		return nil, fmt.Errorf("unmarshal to v1 JSONSchemaProps: %w", err)
	}

	// 2. From v1.JSONSchemaProps to internal JSONSchemaProps
	var schemaInternal apiextensions.JSONSchemaProps
	if err := apiextv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(
		&schemaV1, &schemaInternal, nil); err != nil {
		return nil, fmt.Errorf("convert v1→internal: %w", err)
	}

	// 3. Dump for debug (to be removed ASAP)
	//dumpSchemaRecursively(&schemaInternal, "")

	return &apiextensions.CustomResourceValidation{
		OpenAPIV3Schema: &schemaInternal,
	}, nil
}
