package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "templates.krateo.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// CustomForm type metadata.
var (
	CustomFormKind             = reflect.TypeOf(CustomForm{}).Name()
	CustomFormGroupKind        = schema.GroupKind{Group: Group, Kind: CustomFormKind}.String()
	CustomFormKindAPIVersion   = CustomFormKind + "." + SchemeGroupVersion.String()
	CustomFormGroupVersionKind = SchemeGroupVersion.WithKind(CustomFormKind)
)

func init() {
	SchemeBuilder.Register(&CustomForm{}, &CustomFormList{})
}
