package v1alpha1

// Reference to a named object.
type Reference struct {
	// Name of the referenced object.
	Name string `json:"name"`

	// Namespace of the referenced object.
	Namespace string `json:"namespace"`
}

// API represents a request to an HTTP service
type API struct {
	// Name is a (unique) identifier
	Name string `json:"name"`
	// Path is the request URI path
	Path *string `json:"path,omitempty"`
	// Verb is the request method (GET if omitted)
	Verb *string `json:"verb,omitempty"`
	//+listType=atomic
	// Headers is an array of custom request headers
	Headers []string `json:"headers,omitempty"`
	// Payload is the request body
	Payload *string `json:"payload,omitempty"`
	// EndpointRef a reference to an Endpoint
	EndpointRef *Reference `json:"endpointRef,omitempty"`
	// DependOn reference to the identifier (name) of another API on which this depends
	DependOn *string `json:"dependOn,omitempty"`
}

// ObjectReference is a reference to a named object in a specified namespace.
type ObjectReference struct {
	Reference  `json:",inline"`
	Resource   string `json:"resource,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

// Data is a key value pair.
type Data struct {
	// Name of the data
	Name string `json:"name"`
	// Value of the data. Can be also a JQ expression.
	Value string `json:"value,omitempty"`
	// AsString if true the value will be considered verbatim as string.
	AsString *bool `json:"asString,omitempty"`
}
