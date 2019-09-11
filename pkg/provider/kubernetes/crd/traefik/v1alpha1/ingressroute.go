package v1alpha1

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/types"
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
// Let's Encrypt, use an empty TLS struct, e.g. in YAML:
//
// tls: {} # inline format
//
// tls:
//   secretName: # block format
type TLS struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the
	// certificate details.
	SecretName string `json:"secretName"`
	// Options is a reference to a TLSOption, that specifies the parameters of the TLS connection.
	Options      *TLSOptionRef  `json:"options,omitempty"`
	CertResolver string         `json:"certResolver,omitempty"`
	Domains      []types.Domain `json:"domains,omitempty"`
}

// TLSOptionRef is a ref to the TLSOption resources.
type TLSOptionRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Service defines an upstream to proxy traffic.
type Service struct {
	Name           string       `json:"name"`
	Port           int32        `json:"port"`
	Scheme         string       `json:"scheme,omitempty"`
	HealthCheck    *HealthCheck `json:"healthCheck,omitempty"`
	Strategy       string       `json:"strategy,omitempty"`
	ServersOptions *Options     `json:"options,omitempty"`
	Weight         *int         `json:"weight,omitempty"`
}

// Options configures options for the servers
type Options struct {
	PassHostHeader     *bool                       `json:"passHostHeader,omitempty"`
	ResponseForwarding *dynamic.ResponseForwarding `json:"responseForwarding,omitempty"`
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
