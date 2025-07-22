package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// ServersTransportTCP is the CRD implementation of a TCPServersTransport.
// If no tcpServersTransport is specified, a default one named default@internal will be used.
// The default@internal tcpServersTransport can be configured in the static configuration.
// More info: https://doc.traefik.io/traefik/v3.5/routing/services/#serverstransport_3
type ServersTransportTCP struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec ServersTransportTCPSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// ServersTransportTCPSpec defines the desired state of a ServersTransportTCP.
type ServersTransportTCPSpec struct {
	// DialTimeout is the amount of time to wait until a connection to a backend server can be established.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	DialTimeout *intstr.IntOrString `json:"dialTimeout,omitempty"`
	// DialKeepAlive is the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	DialKeepAlive *intstr.IntOrString `json:"dialKeepAlive,omitempty"`
	// TerminationDelay defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability.
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	TerminationDelay *intstr.IntOrString `json:"terminationDelay,omitempty"`
	// TLS defines the TLS configuration
	TLS *TLSClientConfig `description:"Defines the TLS configuration." json:"tls,omitempty"`
}

// TLSClientConfig defines the desired state of a TLSClientConfig.
type TLSClientConfig struct {
	// ServerName defines the server name used to contact the server.
	ServerName string `json:"serverName,omitempty"`
	// InsecureSkipVerify disables TLS certificate verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// RootCAs defines a list of CA certificate Secrets or ConfigMaps used to validate server certificates.
	RootCAs []RootCA `json:"rootCAs,omitempty"`
	// RootCAsSecrets defines a list of CA secret used to validate self-signed certificate.
	// Deprecated: RootCAsSecrets is deprecated, please use the RootCAs option instead.
	RootCAsSecrets []string `json:"rootCAsSecrets,omitempty"`
	// CertificatesSecrets defines a list of secret storing client certificates for mTLS.
	CertificatesSecrets []string `json:"certificatesSecrets,omitempty"`
	// MaxIdleConnsPerHost controls the maximum idle (keep-alive) to keep per-host.
	// PeerCertURI defines the peer cert URI used to match against SAN URI during the peer certificate verification.
	PeerCertURI string `json:"peerCertURI,omitempty"`
	// Spiffe defines the SPIFFE configuration.
	Spiffe *dynamic.Spiffe `json:"spiffe,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServersTransportTCPList is a collection of ServersTransportTCP resources.
type ServersTransportTCPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of ServersTransportTCP.
	Items []ServersTransportTCP `json:"items"`
}
