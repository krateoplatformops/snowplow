// Package apis contains Kubernetes API for eventrouter service.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		templatesv1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
