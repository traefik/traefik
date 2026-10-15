package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// +k8s:deepcopy-gen=true

// RuntimeValueSource declares one arbitrarily named value owned by this
// Middleware configuration. Exactly one source must be configured.
type RuntimeValueSource struct {
	Value           string                    `json:"value,omitempty"`
	File            string                    `json:"file,omitempty"`
	SecretKeyRef    *RuntimeObjectKeySelector `json:"secretKeyRef,omitempty"`
	ConfigMapKeyRef *RuntimeObjectKeySelector `json:"configMapKeyRef,omitempty"`
}

// +k8s:deepcopy-gen=true

// RuntimeObjectKeySelector selects one key from an object in the Middleware
// namespace. Cross-namespace secret access is intentionally not supported.
type RuntimeObjectKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// +k8s:deepcopy-gen=true

// AccessGovernance records access governance events.
type AccessGovernance struct {
	RuntimeValues   map[string]RuntimeValueSource            `json:"runtimeValues,omitempty"`
	Databases       map[string]AccessGovernanceDatabase      `json:"databases,omitempty"`
	DefaultDatabase string                                   `json:"defaultDatabase,omitempty"`
	Routes          []dynamic.AccessGovernanceRoute          `json:"routes,omitempty"`
	Identity        []dynamic.AccessGovernanceIdentitySource `json:"identity,omitempty"`
	Queue           AccessGovernanceQueue                    `json:"queue,omitempty"`
	Capture         dynamic.AccessGovernanceCapture          `json:"capture,omitempty"`
	Dimensions      dynamic.AccessGovernanceDimensions       `json:"dimensions,omitempty"`
	Schema          dynamic.AccessGovernanceSchema           `json:"schema,omitempty"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceDatabase defines a MySQL destination.
type AccessGovernanceDatabase struct {
	DSN                string `json:"dsn,omitempty"`
	DSNFile            string `json:"dsnFile,omitempty"`
	HTTPTable          string `json:"httpTable,omitempty"`
	NetworkTable       string `json:"networkTable,omitempty"`
	MaxOpenConnections int    `json:"maxOpenConnections,omitempty"`
	MaxIdleConnections int    `json:"maxIdleConnections,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	ConnectionMaxAge *intstr.IntOrString `json:"connectionMaxAge,omitempty"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceQueue controls bounded asynchronous database writes.
type AccessGovernanceQueue struct {
	Capacity  int `json:"capacity,omitempty"`
	Workers   int `json:"workers,omitempty"`
	BatchSize int `json:"batchSize,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	FlushInterval *intstr.IntOrString `json:"flushInterval,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	EnqueueWait *intstr.IntOrString `json:"enqueueWait,omitempty"`
}
