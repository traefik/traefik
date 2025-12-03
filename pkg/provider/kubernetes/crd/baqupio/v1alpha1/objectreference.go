package v1alpha1

// ObjectReference is a generic reference to a Baqup resource.
type ObjectReference struct {
	// Name defines the name of the referenced Baqup resource.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Baqup resource.
	Namespace string `json:"namespace,omitempty"`
}
