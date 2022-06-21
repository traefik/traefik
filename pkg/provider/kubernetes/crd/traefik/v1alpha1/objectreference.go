package v1alpha1

// ObjectReference is a generic reference to a Traefik resource.
type ObjectReference struct {
	// Name defines the name of the referenced Traefik resource.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Traefik resource.
	Namespace string `json:"namespace,omitempty"`
}
