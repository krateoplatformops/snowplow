package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type FormSpec struct {
	SchemaDefinitionRef      *Reference `json:"schemaDefinitionRef,omitempty"`
	CompositionDefinitionRef *Reference `json:"compositionDefinitionRef,omitempty"`
}

type FormStatusContent struct {
	Kind       string                `json:"kind,omitempty"`
	APIVersion string                `json:"apiVersion,omitempty"`
	Schema     *runtime.RawExtension `json:"schema,omitempty"`
	Instance   *runtime.RawExtension `json:"instance,omitempty"`
}

type FormStatus struct {
	Content *FormStatusContent `json:"content,omitempty"`
	//+listType=atomic
	Actions []*Action `json:"actions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Form struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FormSpec   `json:"spec,omitempty"`
	Status FormStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FormList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Form `json:"items"`
}
