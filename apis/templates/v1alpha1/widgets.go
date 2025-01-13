package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppTemplate is a template for a widget app.
type AppTemplate struct {
	// Iterator defines a field on which iterate.
	Iterator *string `json:"iterator,omitempty"`
	// Template to use during iteration.
	Template map[string]string `json:"template,omitempty"`
}

// WidgetSpec defines widget specifications.
type WidgetSpec struct {
	// Type of this object.
	Type string `json:"type"`
	//+listType=atomic
	// Actions is an array of actions.
	Actions []*ActionTemplateIterator `json:"actions,omitempty"`
	// PropsRef reference to a config map of extra properties.
	PropsRef *Reference `json:"propsRef,omitempty"`
	// App is the app template
	App *AppTemplate `json:"app,omitempty"`
	//+listType=atomic
	// API array of api calls.
	API []*API `json:"api,omitempty"`
}

// WidgetResult contains all widget details
type WidgetResult struct {
	// Traits is the expandend map of the widget properties.
	Traits map[string]string `json:"app,omitempty"`
	//+listType=atomic
	// Actions is the array of all available actions.
	Actions []*ActionResult `json:"actions,omitempty"`
}

// WidgetStatus contains the Widget results.
type WidgetStatus struct {
	// UID is the uinique identifier of this object.
	UID string `json:"uid,omitempty"`
	// Name user defined name of this object.
	Name string `json:"name,omitempty"`
	// Type of this object.
	Type string `json:"type,omitempty"`
	// Props are user defined extra attributes.
	Props map[string]string `json:"props,omitempty"`
	//+listType=atomic
	// Items array of expanded widget details.
	Items []*WidgetResult `json:"items,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,shortName=wdg,categories={krateo,widgets}

// Widget is ui widgets configuration.
type Widget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WidgetSpec   `json:"spec,omitempty"`
	Status WidgetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WidgetList contains a list of Widget
type WidgetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Widget `json:"items"`
}
