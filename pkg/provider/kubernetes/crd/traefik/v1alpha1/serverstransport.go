package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// ServersTransport is a specification for a ServersTransport resource.
type ServersTransport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ServersTransportSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// ServersTransportSpec options to configure communication between Traefik and the servers.
type ServersTransportSpec struct {
	// ServerName used to contact the server.
	ServerName string `json:"serverName,omitempty"`
	// Disable SSL certificate verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// Add cert file for self-signed certificate.
	RootCAsSecrets []string `json:"rootCAsSecrets,omitempty"`
	// Certificates for mTLS.
	CertificatesSecrets []string `json:"certificatesSecrets,omitempty"`
	// If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost,omitempty"`
	// Timeouts for requests forwarded to the backend servers.
	ForwardingTimeouts *ForwardingTimeouts `json:"forwardingTimeouts,omitempty"`
	// Disable HTTP/2 for connections with backend servers.
	DisableHTTP2 bool `json:"disableHTTP2,omitempty"`
	// URI used to match against SAN URI during the peer certificate verification.
	PeerCertURI string `json:"peerCertURI,omitempty"`
}

// +k8s:deepcopy-gen=true

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	// DialTimeout is the amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists.
	DialTimeout *intstr.IntOrString `json:"dialTimeout,omitempty"`
	// ResponseHeaderTimeout is the amount of time to wait for a server's response headers after fully writing the request (including its body, if any).
	// If zero, no timeout exists.
	ResponseHeaderTimeout *intstr.IntOrString `json:"responseHeaderTimeout,omitempty"`
	// IdleConnTimeout is the maximum period for which an idle HTTP keep-alive connection will remain open before closing itself.
	IdleConnTimeout *intstr.IntOrString `json:"idleConnTimeout,omitempty"`
	// ReadIdleTimeout is the timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection. If zero, no health check is performed.
	ReadIdleTimeout *intstr.IntOrString `json:"readIdleTimeout,omitempty"`
	// PingTimeout is the timeout after which the HTTP/2 connection will be closed if a response to ping is not received.
	PingTimeout *intstr.IntOrString `json:"pingTimeout,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServersTransportList is a list of ServersTransport resources.
type ServersTransportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServersTransport `json:"items"`
}
