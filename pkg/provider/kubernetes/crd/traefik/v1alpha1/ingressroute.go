package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressRouteSpec is a specification for a IngressRouteSpec resource.
type IngressRouteSpec struct {
	Routes      []Route  `json:"routes"`
	EntryPoints []string `json:"entryPoints"`
	TLS         *TLS     `json:"tls,omitempty"`
}

// Route contains the set of routes.
type Route struct {
	Match       string          `json:"match"`
	Kind        string          `json:"kind"`
	Priority    int             `json:"priority"`
	Services    []Service       `json:"services,omitempty"`
	Middlewares []MiddlewareRef `json:"middlewares"`
}

// TLS contains the TLS certificates configuration of the routes. To enable
// Let's Encrypt, set a SecretName with an empty value.
type TLS struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the
	// certificate details.
	SecretName string `json:"secretName"`
	// TODO MinimumProtocolVersion string `json:"minimumProtocolVersion,omitempty"`
}

// Service defines an upstream to proxy traffic.
type Service struct {
	Name string `json:"name"`
	Port int32  `json:"port"`
	// TODO Weight      int          `json:"weight,omitempty"`
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
	Strategy    string       `json:"strategy,omitempty"`
}

// MiddlewareRef is a ref to the Middleware resources.
type MiddlewareRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// HealthCheck is the HealthCheck definition.
type HealthCheck struct {
	Path            string            `json:"path"`
	Host            string            `json:"host,omitempty"`
	Scheme          string            `json:"scheme"`
	IntervalSeconds int64             `json:"intervalSeconds"`
	TimeoutSeconds  int64             `json:"timeoutSeconds"`
	Headers         map[string]string `json:"headers"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRoute is an Ingress CRD specification.
type IngressRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteList is a list of IngressRoutes.
type IngressRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []IngressRoute `json:"items"`
}
