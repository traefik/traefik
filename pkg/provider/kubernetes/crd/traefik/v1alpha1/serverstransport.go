package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServersTransport is a specification for a ServersTransport resource.
type ServersTransport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ServersTransportSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// ServersTransportSpec options to configure communication between Traefik and the servers.
type ServersTransportSpec struct {
	ServerName          string              `description:"ServerName used to contact the server" json:"serverName,omitempty" toml:"serverName,omitempty" yaml:"serverName,omitempty" export:"true"`
	InsecureSkipVerify  bool                `description:"Disable SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAsSecrets      []string            `description:"Add cert file for self-signed certificate." json:"rootCAsSecrets,omitempty" toml:"rootCAsSecrets,omitempty" yaml:"rootCAsSecrets,omitempty"`
	CertificatesSecrets []string            `description:"Certificates for mTLS." json:"certificatesSecrets,omitempty" toml:"certificatesSecrets,omitempty" yaml:"certificatesSecrets,omitempty"`
	MaxIdleConnsPerHost int                 `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used" json:"maxIdleConnsPerHost,omitempty" toml:"maxIdleConnsPerHost,omitempty" yaml:"maxIdleConnsPerHost,omitempty" export:"true"`
	ForwardingTimeouts  *ForwardingTimeouts `description:"Timeouts for requests forwarded to the backend servers." json:"forwardingTimeouts,omitempty" toml:"forwardingTimeouts,omitempty" yaml:"forwardingTimeouts,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           *intstr.IntOrString `description:"The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	ResponseHeaderTimeout *intstr.IntOrString `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists." json:"responseHeaderTimeout,omitempty" toml:"responseHeaderTimeout,omitempty" yaml:"responseHeaderTimeout,omitempty" export:"true"`
	IdleConnTimeout       *intstr.IntOrString `description:"The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself" json:"idleConnTimeout,omitempty" toml:"idleConnTimeout,omitempty" yaml:"idleConnTimeout,omitempty" export:"true"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServersTransportList is a list of ServersTransport resources.
type ServersTransportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServersTransport `json:"items"`
}
