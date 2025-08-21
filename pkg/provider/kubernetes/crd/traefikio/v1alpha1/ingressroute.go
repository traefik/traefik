package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressRouteSpec defines the desired state of IngressRoute.
type IngressRouteSpec struct {
	// Routes defines the list of routes.
	Routes []Route `json:"routes"`
	// EntryPoints defines the list of entry point names to bind to.
	// Entry points have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/entrypoints/
	// Default: all.
	EntryPoints []string `json:"entryPoints,omitempty"`
	// TLS defines the TLS configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#tls
	TLS *TLS `json:"tls,omitempty"`
}

// Route holds the HTTP route configuration.
type Route struct {
	// Match defines the router's rule.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#rule
	Match string `json:"match"`
	// Kind defines the kind of the route.
	// Rule is the only supported kind.
	// If not defined, defaults to Rule.
	// +kubebuilder:validation:Enum=Rule
	Kind string `json:"kind,omitempty"`
	// Priority defines the router's priority.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#priority
	// +kubebuilder:validation:Maximum=9223372036854774807
	Priority int `json:"priority,omitempty"`
	// Syntax defines the router's rule syntax.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#rulesyntax
	// Deprecated: Please do not use this field and rewrite the router rules to use the v3 syntax.
	Syntax string `json:"syntax,omitempty"`
	// Services defines the list of Service.
	// It can contain any combination of TraefikService and/or reference to a Kubernetes Service.
	Services []Service `json:"services,omitempty"`
	// Middlewares defines the list of references to Middleware resources.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#kind-middleware
	Middlewares []MiddlewareRef `json:"middlewares,omitempty"`
	// Observability defines the observability configuration for a router.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#observability
	Observability *dynamic.RouterObservabilityConfig `json:"observability,omitempty"`
}

// TLS holds the TLS configuration.
// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#tls
type TLS struct {
	// SecretName is the name of the referenced Kubernetes Secret to specify the certificate details.
	SecretName string `json:"secretName,omitempty"`
	// Options defines the reference to a TLSOption, that specifies the parameters of the TLS connection.
	// If not defined, the `default` TLSOption is used.
	// More info: https://doc.traefik.io/traefik/v3.5/https/tls/#tls-options
	Options *TLSOptionRef `json:"options,omitempty"`
	// Store defines the reference to the TLSStore, that will be used to store certificates.
	// Please note that only `default` TLSStore can be used.
	Store *TLSStoreRef `json:"store,omitempty"`
	// CertResolver defines the name of the certificate resolver to use.
	// Cert resolvers have to be configured in the static configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/https/acme/#certificate-resolvers
	CertResolver string `json:"certResolver,omitempty"`
	// Domains defines the list of domains that will be used to issue certificates.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/routers/#domains
	Domains []types.Domain `json:"domains,omitempty"`
}

// TLSOptionRef is a reference to a TLSOption resource.
type TLSOptionRef struct {
	// Name defines the name of the referenced TLSOption.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#kind-tlsoption
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced TLSOption.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#kind-tlsoption
	Namespace string `json:"namespace,omitempty"`
}

// TLSStoreRef is a reference to a TLSStore resource.
type TLSStoreRef struct {
	// Name defines the name of the referenced TLSStore.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#kind-tlsstore
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced TLSStore.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/providers/kubernetes-crd/#kind-tlsstore
	Namespace string `json:"namespace,omitempty"`
}

// LoadBalancerSpec defines the desired state of LoadBalancer.
// It can reference either a Kubernetes Service object (a load-balancer of servers),
// or a TraefikService object (a load-balancer of Traefik services).
type LoadBalancerSpec struct {
	// Name defines the name of the referenced Kubernetes Service or TraefikService.
	// The differentiation between the two is specified in the Kind field.
	Name string `json:"name"`
	// Kind defines the kind of the Service.
	// +kubebuilder:validation:Enum=Service;TraefikService
	Kind string `json:"kind,omitempty"`
	// Namespace defines the namespace of the referenced Kubernetes Service or TraefikService.
	Namespace string `json:"namespace,omitempty"`
	// Sticky defines the sticky sessions configuration.
	// More info: https://doc.traefik.io/traefik/v3.5/routing/services/#sticky-sessions
	Sticky *dynamic.Sticky `json:"sticky,omitempty"`
	// Port defines the port of a Kubernetes Service.
	// This can be a reference to a named port.
	// +kubebuilder:validation:XIntOrString
	Port intstr.IntOrString `json:"port,omitempty"`
	// Scheme defines the scheme to use for the request to the upstream Kubernetes Service.
	// It defaults to https when Kubernetes Service port is 443, http otherwise.
	Scheme string `json:"scheme,omitempty"`
	// Strategy defines the load balancing strategy between the servers.
	// Supported values are: wrr (Weighed round-robin) and p2c (Power of two choices).
	// RoundRobin value is deprecated and supported for backward compatibility.
	// TODO: when the deprecated RoundRobin value will be removed, set the default value to wrr.
	// +kubebuilder:validation:Enum=wrr;p2c;RoundRobin
	Strategy dynamic.BalancerStrategy `json:"strategy,omitempty"`
	// PassHostHeader defines whether the client Host header is forwarded to the upstream Kubernetes Service.
	// By default, passHostHeader is true.
	PassHostHeader *bool `json:"passHostHeader,omitempty"`
	// ResponseForwarding defines how Traefik forwards the response from the upstream Kubernetes Service to the client.
	ResponseForwarding *ResponseForwarding `json:"responseForwarding,omitempty"`
	// ServersTransport defines the name of ServersTransport resource to use.
	// It allows to configure the transport between Traefik and your servers.
	// Can only be used on a Kubernetes Service.
	ServersTransport string `json:"serversTransport,omitempty"`
	// Weight defines the weight and should only be specified when Name references a TraefikService object
	// (and to be precise, one that embeds a Weighted Round Robin).
	// +kubebuilder:validation:Minimum=0
	Weight *int `json:"weight,omitempty"`
	// NativeLB controls, when creating the load-balancer,
	// whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.
	// The Kubernetes Service itself does load-balance to the pods.
	// By default, NativeLB is false.
	NativeLB *bool `json:"nativeLB,omitempty"`
	// NodePortLB controls, when creating the load-balancer,
	// whether the LB's children are directly the nodes internal IPs using the nodePort when the service type is NodePort.
	// It allows services to be reachable when Traefik runs externally from the Kubernetes cluster but within the same network of the nodes.
	// By default, NodePortLB is false.
	NodePortLB bool `json:"nodePortLB,omitempty"`
	// Healthcheck defines health checks for ExternalName services.
	HealthCheck *ServerHealthCheck `json:"healthCheck,omitempty"`
	// PassiveHealthCheck defines passive health checks for ExternalName services.
	PassiveHealthCheck *PassiveServerHealthCheck `json:"passiveHealthCheck,omitempty"`
}

type ResponseForwarding struct {
	// FlushInterval defines the interval, in milliseconds, in between flushes to the client while copying the response body.
	// A negative value means to flush immediately after each write to the client.
	// This configuration is ignored when ReverseProxy recognizes a response as a streaming response;
	// for such responses, writes are flushed to the client immediately.
	// Default: 100ms
	FlushInterval string `json:"flushInterval,omitempty"`
}

type ServerHealthCheck struct {
	// Scheme replaces the server URL scheme for the health check endpoint.
	Scheme string `json:"scheme,omitempty"`
	// Mode defines the health check mode.
	// If defined to grpc, will use the gRPC health check protocol to probe the server.
	// Default: http
	Mode string `json:"mode,omitempty"`
	// Path defines the server URL path for the health check endpoint.
	Path string `json:"path,omitempty"`
	// Method defines the healthcheck method.
	Method string `json:"method,omitempty"`
	// Status defines the expected HTTP status code of the response to the health check request.
	Status int `json:"status,omitempty"`
	// Port defines the server URL port for the health check endpoint.
	Port int `json:"port,omitempty"`
	// Interval defines the frequency of the health check calls for healthy targets.
	// Default: 30s
	Interval *intstr.IntOrString `json:"interval,omitempty"`
	// UnhealthyInterval defines the frequency of the health check calls for unhealthy targets.
	// When UnhealthyInterval is not defined, it defaults to the Interval value.
	// Default: 30s
	UnhealthyInterval *intstr.IntOrString `json:"unhealthyInterval,omitempty"`
	// Timeout defines the maximum duration Traefik will wait for a health check request before considering the server unhealthy.
	// Default: 5s
	Timeout *intstr.IntOrString `json:"timeout,omitempty"`
	// Hostname defines the value of hostname in the Host header of the health check request.
	Hostname string `json:"hostname,omitempty"`
	// FollowRedirects defines whether redirects should be followed during the health check calls.
	// Default: true
	FollowRedirects *bool `json:"followRedirects,omitempty"`
	// Headers defines custom headers to be sent to the health check endpoint.
	Headers map[string]string `json:"headers,omitempty"`
}

type PassiveServerHealthCheck struct {
	// FailureWindow defines the time window during which the failed attempts must occur for the server to be marked as unhealthy. It also defines for how long the server will be considered unhealthy.
	FailureWindow *intstr.IntOrString `json:"failureWindow,omitempty"`
	// MaxFailedAttempts is the number of consecutive failed attempts allowed within the failure window before marking the server as unhealthy.
	MaxFailedAttempts *int `json:"maxFailedAttempts,omitempty"`
}

// Service defines an upstream HTTP service to proxy traffic to.
type Service struct {
	LoadBalancerSpec `json:",inline"`
}

// MiddlewareRef is a reference to a Middleware resource.
type MiddlewareRef struct {
	// Name defines the name of the referenced Middleware resource.
	Name string `json:"name"`
	// Namespace defines the namespace of the referenced Middleware resource.
	Namespace string `json:"namespace,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// IngressRoute is the CRD implementation of a Traefik HTTP Router.
type IngressRoute struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec IngressRouteSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressRouteList is a collection of IngressRoute.
type IngressRouteList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of IngressRoute.
	Items []IngressRoute `json:"items"`
}
