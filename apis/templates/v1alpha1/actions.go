package v1alpha1

type ActionTemplateIterator struct {
	Iterator *string         `json:"iterator,omitempty"`
	Template *ActionTemplate `json:"template,omitempty"`
}

// ActionTemplate defines a template for an action.
type ActionTemplate struct {
	// ID for the action.
	ID string `json:"id,omitempty"`
	// Name of the related resource.
	Name string `json:"name,omitempty"`
	// Namespace of the related resource.
	Namespace string `json:"namespace,omitempty"`
	// Resource on which the action will act.
	Resource string `json:"resource,omitempty"`
	// APIVersion for the related resource
	APIVersion string `json:"apiVersion,omitempty"`
	// Verb is the HTTP request verb.
	Verb string `json:"verb,omitempty"`
	//+listType=atomic
	// PayloadToOverride a list of values to override.
	PayloadToOverride []Data `json:"payloadToOverride,omitempty"`
}

// Action wraps results of an action template.
type ActionResultTemplate struct {
	Template *ActionResult `json:"template,omitempty"`
}

// ActionResult defines the action result after evaluating a template.
type ActionResult struct {
	// ID of this action.
	ID string `json:"id,omitempty"`
	// Path is the HTTP request path.
	Path string `json:"path,omitempty"`
	// Verb is the HTTP request verb.
	Verb string `json:"verb,omitempty"`
	//+listType=atomic
	// PayloadToOverride a list of values to override.
	PayloadToOverride []Data `json:"payloadToOverride,omitempty"`
	// Payload the payload for the action result
	Payload *ActionResultPayload `json:"payload,omitempty"`
}

// ActionResultPayload is the template action result payload.
type ActionResultPayload struct {
	Kind       string     `json:"kind,omitempty"`
	APIVersion string     `json:"apiVersion,omitempty"`
	MetaData   *Reference `json:"metadata,omitempty"`
}
