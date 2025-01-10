package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CustomFormAppTemplate is the custom form app template.
type CustomFormAppTemplate struct {
	// Schema holds the JQ expression to retrieve the schema of this form.
	Schema string `json:"schema"`
	//+listType=atomic
	// PropertiesToHide a list of schema values to hide.
	PropertiesToHide []string `json:"propertiesToHide,omitempty"`
	//+listType=atomic
	// PropertiesToOverride a list of schema values to override.
	PropertiesToOverride []Data `json:"propertiesToOverride,omitempty"`
}

// CustomFormApp is the custom form app template.
type CustomFormApp struct {
	Template *CustomFormAppTemplate `json:"template,omitempty"`
}

// CustomFormSpec is the custom form app specification.
type CustomFormSpec struct {
	// Type of this object.
	Type string `json:"type"`
	// PropsRef reference to a config map of extra properties.
	PropsRef *Reference `json:"propsRef,omitempty"`
	//+listType=atomic
	// Actions is an array of actions.
	Actions []*Action `json:"actions,omitempty"`
	// App defines app properties.
	App *CustomFormApp `json:"app,omitempty"`
	//+listType=atomic
	// API array of api calls.
	API []*API `json:"api,omitempty"`
}

// CustomFormStatusContent wraps the custom form content
type CustomFormStatusContent struct {
	// Schema holds this custom form schema.
	Schema *runtime.RawExtension `json:"schema,omitempty"`
}

// CustomFormStatus wraps the custom form response.
type CustomFormStatus struct {
	// Type of this object.
	Type string `json:"type"`
	// Name user defined name of this object.
	Name string `json:"name"`
	// UID is the uinique identifier of this object.
	UID *string `json:"uid,omitempty"`
	// Props are user defined extra attributes.
	Props   map[string]string        `json:"props,omitempty"`
	Content *CustomFormStatusContent `json:"content,omitempty"`
	//+listType=atomic
	// Actions is the array of all available actions.
	Actions []*ActionResultTemplate `json:"actions,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,shortName=cform,categories={krateo,customforms}

// CustomForm design a custom form.
type CustomForm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomFormSpec   `json:"spec,omitempty"`
	Status CustomFormStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomFormList is a list of custom forms
type CustomFormList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CustomForm `json:"items"`
}
