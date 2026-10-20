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

// ExternalAction transforms matching requests into external actions.
type ExternalAction struct {
	// Enrichment references same-namespace QueryEnrichment middlewares in execution order.
	Enrichment     []string                               `json:"enrichment,omitempty"`
	RuntimeValues  map[string]RuntimeValueSource          `json:"runtimeValues,omitempty"`
	SelectionMode  string                                 `json:"selectionMode,omitempty"`
	CredentialSets map[string]ExternalActionCredentialSet `json:"credentialSets,omitempty"`
	Pollers        []ExternalActionPoller                 `json:"pollers,omitempty"`
	Rules          []ExternalActionRule                   `json:"rules,omitempty"`
	Queue          ExternalActionQueue                    `json:"queue,omitempty"`
}

// QueryEnrichment defines event metadata lookups with namespaced runtime credentials.
// +k8s:deepcopy-gen=true
type QueryEnrichment struct {
	RuntimeValues  map[string]RuntimeValueSource          `json:"runtimeValues,omitempty"`
	CredentialSets map[string]ExternalActionCredentialSet `json:"credentialSets,omitempty"`
	InputTemplate  string                                 `json:"inputTemplate,omitempty"`
	Queries        []EnrichmentQuery                      `json:"queries,omitempty"`
	OutputTemplate string                                 `json:"outputTemplate,omitempty"`
}

// EnrichmentQuery accepts the same Kubernetes duration format as other middlewares.
// +k8s:deepcopy-gen=true
type EnrichmentQuery struct {
	Name           string                                  `json:"name,omitempty"`
	When           string                                  `json:"when,omitempty"`
	HTTP           dynamic.ExternalActionHTTP              `json:"http,omitempty"`
	Body           dynamic.ExternalActionBody              `json:"body,omitempty"`
	AllowedOrigins []string                                `json:"allowedOrigins,omitempty"`
	ResultTemplate string                                  `json:"resultTemplate,omitempty"`
	MaxConcurrent  int                                     `json:"maxConcurrent,omitempty"`
	CacheEntries   int                                     `json:"cacheEntries,omitempty"`
	CacheBytes     int64                                   `json:"cacheBytes,omitempty"`
	Pagination     *dynamic.ExternalActionPollerPagination `json:"pagination,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	Timeout *intstr.IntOrString `json:"timeout,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	CacheTTL *intstr.IntOrString `json:"cacheTTL,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionPoller periodically observes a configured HTTP source.
type ExternalActionPoller struct {
	Name         string                                  `json:"name,omitempty"`
	Request      dynamic.ExternalActionPollerRequest     `json:"request,omitempty"`
	Observation  dynamic.ExternalActionPollerObservation `json:"observation,omitempty"`
	Pagination   *dynamic.ExternalActionPollerPagination `json:"pagination,omitempty"`
	Coordination *ExternalActionDeduplication            `json:"coordination,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	Interval *intstr.IntOrString `json:"interval,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	InitialDelay *intstr.IntOrString `json:"initialDelay,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	Timeout *intstr.IntOrString `json:"timeout,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionCredentialSet defines one reloadable credential destination.
type ExternalActionCredentialSet struct {
	Auth                dynamic.ExternalActionAuth `json:"auth,omitempty"`
	ImmutableParameters []string                   `json:"immutableParameters,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	ReloadInterval *intstr.IntOrString `json:"reloadInterval,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionRule defines one external action.
type ExternalActionRule struct {
	Name          string                             `json:"name,omitempty"`
	Match         dynamic.ExternalActionMatch        `json:"match,omitempty"`
	Phase         string                             `json:"phase,omitempty"`
	Mode          string                             `json:"mode,omitempty"`
	FailurePolicy string                             `json:"failurePolicy,omitempty"`
	Body          dynamic.ExternalActionBody         `json:"body,omitempty"`
	ResponseBody  dynamic.ExternalActionResponseBody `json:"responseBody,omitempty"`
	ForEach       *dynamic.ExternalActionForEach     `json:"forEach,omitempty"`
	Deduplication *ExternalActionDeduplication       `json:"deduplication,omitempty"`
	Binding       *ExternalActionBinding             `json:"binding,omitempty"`
	HTTP          *dynamic.ExternalActionHTTP        `json:"http,omitempty"`
	Exec          *dynamic.ExternalActionExec        `json:"exec,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	Timeout *intstr.IntOrString `json:"timeout,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionDeduplication suppresses repeated actions produced by polling endpoints.
type ExternalActionDeduplication struct {
	KeyTemplate string                             `json:"keyTemplate,omitempty"`
	Memory      *dynamic.ExternalActionDedupMemory `json:"memory,omitempty"`
	MySQL       *dynamic.ExternalActionDedupMySQL  `json:"mysql,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	TTL *intstr.IntOrString `json:"ttl,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	Lease *intstr.IntOrString `json:"lease,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionBinding pins a lifecycle key to one immutable credential set.
type ExternalActionBinding struct {
	Mode string `json:"mode,omitempty"`
	// +kubebuilder:validation:Enum=fail;skip
	OnMissing             string                               `json:"onMissing,omitempty"`
	KeyTemplate           string                               `json:"keyTemplate,omitempty"`
	CredentialSet         string                               `json:"credentialSet,omitempty"`
	AllowedCredentialSets []string                             `json:"allowedCredentialSets,omitempty"`
	Match                 dynamic.ExternalActionBindingMatch   `json:"match,omitempty"`
	Memory                *dynamic.ExternalActionBindingMemory `json:"memory,omitempty"`
	MySQL                 *dynamic.ExternalActionBindingMySQL  `json:"mysql,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	TTL *intstr.IntOrString `json:"ttl,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionQueue controls bounded asynchronous dispatch.
type ExternalActionQueue struct {
	Capacity       int   `json:"capacity,omitempty"`
	Workers        int   `json:"workers,omitempty"`
	MaxQueuedBytes int64 `json:"maxQueuedBytes,omitempty"`
	// +kubebuilder:validation:Pattern="^([0-9]+(ns|us|µs|ms|s|m|h)?)+$"
	// +kubebuilder:validation:XIntOrString
	EnqueueWait *intstr.IntOrString `json:"enqueueWait,omitempty"`
}
