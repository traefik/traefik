package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// ServersTransportTCP is the CRD implementation of a TCPServersTransport.
// If no tcpServersTransport is specified, a default one named default will be used.
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
	DialTimeout *intstr.IntOrString `json:"dialTimeout,omitempty"`
	// DialKeepAlive is the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled.
	DialKeepAlive *intstr.IntOrString `json:"dialKeepAlive,omitempty"`
	// TerminationDelay defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability.
	TerminationDelay *intstr.IntOrString `json:"terminationDelay,omitempty"`
	// TLS defines the TLS configuration
	TLS *TLSClientConfig `description:"Defines the TLS configuration." json:"tls,omitempty"`
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
