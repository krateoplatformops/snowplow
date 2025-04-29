package v1

// BackendEndpoint defines a template for an action.
type BackendEndpoint struct {
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
}

// BackendEndpointResult defines the action result after evaluating a template.
type BackendEndpointResult struct {
	// ID of this action.
	ID string `json:"id,omitempty"`
	// Path is the HTTP request path.
	Path string `json:"path,omitempty"`
	// Verb is the HTTP request verb.
	Verb string `json:"verb,omitempty"`
	// Payload the payload for the action result
	Payload *BackendEndpointResultPayload `json:"payload,omitempty"`
}

// BackendEndpointResultPayload is the template action result payload.
type BackendEndpointResultPayload struct {
	Kind       string     `json:"kind,omitempty"`
	APIVersion string     `json:"apiVersion,omitempty"`
	MetaData   *Reference `json:"metadata,omitempty"`
}
