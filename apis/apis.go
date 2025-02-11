package apis

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	templatesv1 "github.com/krateoplatformops/snowplow/apis/templates/v1"
)

func init() {
	AddToSchemes = append(AddToSchemes,
		templatesv1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	if s == nil {
		return fmt.Errorf("runtime.Scheme cannot be nil")
	}
	return AddToSchemes.AddToScheme(s)
}
