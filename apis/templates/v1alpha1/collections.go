package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemplateIterator struct {
	// Iterator defines a field on which iterate.
	Iterator *string `json:"iterator,omitempty"`
	// Template to use during iteration.
	Template *ObjectReference `json:"template,omitempty"`
}

// CollectionSpec defines specifications for a Collection.
type CollectionSpec struct {
	// Type of this object.
	Type string `json:"type"`
	//+listType=atomic
	// TemplateIterators array of template iterator references.
	TemplateIterators []*TemplateIterator `json:"widgetsRefs,omitempty"`
	// PropsRef reference to a config map of extra properties for this Collection.
	PropsRef *Reference `json:"propsRef,omitempty"`
	//+listType=atomic
	// API array of api calls.
	API []*API `json:"api,omitempty"`
}

// CollectionStatus defines a status for this CollectionIterator.
type CollectionStatus struct {
	// UID is the uinique identifier of this object.
	UID string `json:"uid,omitempty"`
	// Name user defined name of this collection.
	Name string `json:"name,omitempty"`
	// Type of this collection.
	Type string `json:"type,omitempty"`
	// Props are user defined extra attributes.
	Props map[string]string `json:"props,omitempty"`
	//+listType=atomic
	// Items array of expanded widgets.
	Items []*runtime.RawExtension `json:"items,omitempty"`
}

// +structType=atomic
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,shortName=coll,categories={krateo,collections}

type Collection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CollectionSpec   `json:"spec,omitempty"`
	Status CollectionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CollectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Collection `json:"items"`
}
