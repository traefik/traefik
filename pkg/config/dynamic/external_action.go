package dynamic

import ptypes "github.com/traefik/paerser/types"

// +k8s:deepcopy-gen=true

// ExternalAction holds rules that transform a request into an authenticated external action.
type ExternalAction struct {
	RuntimeValues  map[string]RuntimeValue                `json:"runtimeValues,omitempty" toml:"runtimeValues,omitempty" yaml:"runtimeValues,omitempty" label:"-" export:"true"`
	SelectionMode  string                                 `json:"selectionMode,omitempty" toml:"selectionMode,omitempty" yaml:"selectionMode,omitempty"`
	CredentialSets map[string]ExternalActionCredentialSet `json:"credentialSets,omitempty" toml:"credentialSets,omitempty" yaml:"credentialSets,omitempty" label:"-" export:"true"`
	Pollers        []ExternalActionPoller                 `json:"pollers,omitempty" toml:"pollers,omitempty" yaml:"pollers,omitempty" label:"-" export:"true"`
	Rules          []ExternalActionRule                   `json:"rules,omitempty" toml:"rules,omitempty" yaml:"rules,omitempty" label:"-" export:"true"`
	Queue          ExternalActionQueue                    `json:"queue,omitempty" toml:"queue,omitempty" yaml:"queue,omitempty" export:"true"`
	// ConfigurationID is populated internally from the stable Middleware
	// instance name. It prevents unrelated runtime-value changes from changing
	// a credential set's process-level identity.
	ConfigurationID string `json:"-" toml:"-" yaml:"-"`
}

// +k8s:deepcopy-gen=true

// ExternalActionPoller periodically observes an HTTP source and feeds its
// response through the same rule, template, binding, and deduplication engine
// as an ordinary proxied response.
type ExternalActionPoller struct {
	Name         string                          `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Interval     ptypes.Duration                 `json:"interval,omitempty" toml:"interval,omitempty" yaml:"interval,omitempty" export:"true"`
	InitialDelay ptypes.Duration                 `json:"initialDelay,omitempty" toml:"initialDelay,omitempty" yaml:"initialDelay,omitempty" export:"true"`
	Timeout      ptypes.Duration                 `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	Request      ExternalActionPollerRequest     `json:"request,omitempty" toml:"request,omitempty" yaml:"request,omitempty" export:"true"`
	Observation  ExternalActionPollerObservation `json:"observation,omitempty" toml:"observation,omitempty" yaml:"observation,omitempty" export:"true"`
	Pagination   *ExternalActionPollerPagination `json:"pagination,omitempty" toml:"pagination,omitempty" yaml:"pagination,omitempty" export:"true"`
	Coordination *ExternalActionDeduplication    `json:"coordination,omitempty" toml:"coordination,omitempty" yaml:"coordination,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionPollerRequest defines the independently authenticated source request.
type ExternalActionPollerRequest struct {
	URL                string                      `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty"`
	Method             string                      `json:"method,omitempty" toml:"method,omitempty" yaml:"method,omitempty"`
	Headers            map[string]string           `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	Body               string                      `json:"body,omitempty" toml:"body,omitempty" yaml:"body,omitempty"`
	ContentType        string                      `json:"contentType,omitempty" toml:"contentType,omitempty" yaml:"contentType,omitempty"`
	Auth               *ExternalActionAuth         `json:"auth,omitempty" toml:"auth,omitempty" yaml:"auth,omitempty" export:"true"`
	TLS                *ExternalActionTLS          `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	ResponseValidation *ExternalActionHTTPResponse `json:"responseValidation,omitempty" toml:"responseValidation,omitempty" yaml:"responseValidation,omitempty" export:"true"`
	MaxResponseBytes   int64                       `json:"maxResponseBytes,omitempty" toml:"maxResponseBytes,omitempty" yaml:"maxResponseBytes,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionPollerObservation supplies the stable request identity used to
// select existing ExternalAction rules for a polled response.
type ExternalActionPollerObservation struct {
	Method  string                     `json:"method,omitempty" toml:"method,omitempty" yaml:"method,omitempty"`
	Scheme  string                     `json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty"`
	Host    string                     `json:"host,omitempty" toml:"host,omitempty" yaml:"host,omitempty"`
	Path    string                     `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty"`
	Query   map[string]string          `json:"query,omitempty" toml:"query,omitempty" yaml:"query,omitempty"`
	Headers map[string]string          `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	Event   *ExternalActionPollerEvent `json:"event,omitempty" toml:"event,omitempty" yaml:"event,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionPollerEvent wraps a polled response in the same neutral HTTP
// observation envelope accepted from external event producers.
type ExternalActionPollerEvent struct {
	Source    string `json:"source,omitempty" toml:"source,omitempty" yaml:"source,omitempty"`
	Type      string `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty"`
	Principal string `json:"principal,omitempty" toml:"principal,omitempty" yaml:"principal,omitempty"`
	Host      string `json:"host,omitempty" toml:"host,omitempty" yaml:"host,omitempty"`
	Path      string `json:"path,omitempty" toml:"path,omitempty" yaml:"path,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionPollerPagination bounds page-number based source discovery.
type ExternalActionPollerPagination struct {
	PageParameter     string `json:"pageParameter,omitempty" toml:"pageParameter,omitempty" yaml:"pageParameter,omitempty"`
	PageSizeParameter string `json:"pageSizeParameter,omitempty" toml:"pageSizeParameter,omitempty" yaml:"pageSizeParameter,omitempty"`
	StartPage         int    `json:"startPage,omitempty" toml:"startPage,omitempty" yaml:"startPage,omitempty" export:"true"`
	PageSize          int    `json:"pageSize,omitempty" toml:"pageSize,omitempty" yaml:"pageSize,omitempty" export:"true"`
	MaxPages          int    `json:"maxPages,omitempty" toml:"maxPages,omitempty" yaml:"maxPages,omitempty" export:"true"`
	ItemsJSONPath     string `json:"itemsJSONPath,omitempty" toml:"itemsJSONPath,omitempty" yaml:"itemsJSONPath,omitempty"`
	TotalJSONPath     string `json:"totalJSONPath,omitempty" toml:"totalJSONPath,omitempty" yaml:"totalJSONPath,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionCredentialSet defines reloadable credentials with an immutable destination identity.
type ExternalActionCredentialSet struct {
	Auth                ExternalActionAuth `json:"auth,omitempty" toml:"auth,omitempty" yaml:"auth,omitempty" export:"true"`
	ImmutableParameters []string           `json:"immutableParameters,omitempty" toml:"immutableParameters,omitempty" yaml:"immutableParameters,omitempty"`
	ReloadInterval      ptypes.Duration    `json:"reloadInterval,omitempty" toml:"reloadInterval,omitempty" yaml:"reloadInterval,omitempty" export:"true"`
	// CacheIdentity is derived internally from unresolved runtime references.
	CacheIdentity string `json:"-" toml:"-" yaml:"-"`
}

// +k8s:deepcopy-gen=true

// ExternalActionRule defines when and how an action is dispatched.
type ExternalActionRule struct {
	Name          string                       `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Match         ExternalActionMatch          `json:"match,omitempty" toml:"match,omitempty" yaml:"match,omitempty"`
	Phase         string                       `json:"phase,omitempty" toml:"phase,omitempty" yaml:"phase,omitempty"`
	Mode          string                       `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty"`
	FailurePolicy string                       `json:"failurePolicy,omitempty" toml:"failurePolicy,omitempty" yaml:"failurePolicy,omitempty"`
	Body          ExternalActionBody           `json:"body,omitempty" toml:"body,omitempty" yaml:"body,omitempty"`
	ResponseBody  ExternalActionResponseBody   `json:"responseBody,omitempty" toml:"responseBody,omitempty" yaml:"responseBody,omitempty"`
	ForEach       *ExternalActionForEach       `json:"forEach,omitempty" toml:"forEach,omitempty" yaml:"forEach,omitempty" export:"true"`
	Deduplication *ExternalActionDeduplication `json:"deduplication,omitempty" toml:"deduplication,omitempty" yaml:"deduplication,omitempty" export:"true"`
	Binding       *ExternalActionBinding       `json:"binding,omitempty" toml:"binding,omitempty" yaml:"binding,omitempty" export:"true"`
	HTTP          *ExternalActionHTTP          `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
	Exec          *ExternalActionExec          `json:"exec,omitempty" toml:"exec,omitempty" yaml:"exec,omitempty" export:"true"`
	Timeout       ptypes.Duration              `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionMatch selects source requests and optional response statuses.
type ExternalActionMatch struct {
	Methods               []string          `json:"methods,omitempty" toml:"methods,omitempty" yaml:"methods,omitempty"`
	Hosts                 []string          `json:"hosts,omitempty" toml:"hosts,omitempty" yaml:"hosts,omitempty"`
	PathRegex             string            `json:"pathRegex,omitempty" toml:"pathRegex,omitempty" yaml:"pathRegex,omitempty"`
	Headers               map[string]string `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	Query                 map[string]string `json:"query,omitempty" toml:"query,omitempty" yaml:"query,omitempty"`
	BodyJSONPaths         map[string]string `json:"bodyJSONPaths,omitempty" toml:"bodyJSONPaths,omitempty" yaml:"bodyJSONPaths,omitempty"`
	ResponseStatuses      []string          `json:"responseStatuses,omitempty" toml:"responseStatuses,omitempty" yaml:"responseStatuses,omitempty"`
	ResponseBodyJSONPaths map[string]string `json:"responseBodyJSONPaths,omitempty" toml:"responseBodyJSONPaths,omitempty" yaml:"responseBodyJSONPaths,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionBody defines source-body handling and the target payload template.
type ExternalActionBody struct {
	ReadMode    string `json:"readMode,omitempty" toml:"readMode,omitempty" yaml:"readMode,omitempty"`
	MaxBytes    int64  `json:"maxBytes,omitempty" toml:"maxBytes,omitempty" yaml:"maxBytes,omitempty" export:"true"`
	Overflow    string `json:"overflow,omitempty" toml:"overflow,omitempty" yaml:"overflow,omitempty"`
	Format      string `json:"format,omitempty" toml:"format,omitempty" yaml:"format,omitempty"`
	Template    string `json:"template,omitempty" toml:"template,omitempty" yaml:"template,omitempty"`
	ContentType string `json:"contentType,omitempty" toml:"contentType,omitempty" yaml:"contentType,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionResponseBody controls bounded capture and parsing of the source service response.
type ExternalActionResponseBody struct {
	ReadMode string `json:"readMode,omitempty" toml:"readMode,omitempty" yaml:"readMode,omitempty"`
	MaxBytes int64  `json:"maxBytes,omitempty" toml:"maxBytes,omitempty" yaml:"maxBytes,omitempty" export:"true"`
	Overflow string `json:"overflow,omitempty" toml:"overflow,omitempty" yaml:"overflow,omitempty"`
	Format   string `json:"format,omitempty" toml:"format,omitempty" yaml:"format,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionForEach expands one source request or response into independently templated actions.
type ExternalActionForEach struct {
	Source         string            `json:"source,omitempty" toml:"source,omitempty" yaml:"source,omitempty"`
	JSONPath       string            `json:"jsonPath,omitempty" toml:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`
	MatchJSONPaths map[string]string `json:"matchJSONPaths,omitempty" toml:"matchJSONPaths,omitempty" yaml:"matchJSONPaths,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionDeduplication suppresses repeated actions produced by polling endpoints.
type ExternalActionDeduplication struct {
	KeyTemplate string                     `json:"keyTemplate,omitempty" toml:"keyTemplate,omitempty" yaml:"keyTemplate,omitempty"`
	TTL         ptypes.Duration            `json:"ttl,omitempty" toml:"ttl,omitempty" yaml:"ttl,omitempty" export:"true"`
	Lease       ptypes.Duration            `json:"lease,omitempty" toml:"lease,omitempty" yaml:"lease,omitempty" export:"true"`
	Memory      *ExternalActionDedupMemory `json:"memory,omitempty" toml:"memory,omitempty" yaml:"memory,omitempty" export:"true"`
	MySQL       *ExternalActionDedupMySQL  `json:"mysql,omitempty" toml:"mysql,omitempty" yaml:"mysql,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionDedupMemory configures process-local deduplication.
type ExternalActionDedupMemory struct {
	MaxEntries int `json:"maxEntries,omitempty" toml:"maxEntries,omitempty" yaml:"maxEntries,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionDedupMySQL configures shared deduplication across Traefik replicas.
type ExternalActionDedupMySQL struct {
	DSN        string `json:"dsn,omitempty" toml:"dsn,omitempty" yaml:"dsn,omitempty" loggable:"false"`
	DSNFile    string `json:"dsnFile,omitempty" toml:"dsnFile,omitempty" yaml:"dsnFile,omitempty"`
	Table      string `json:"table,omitempty" toml:"table,omitempty" yaml:"table,omitempty"`
	AutoCreate bool   `json:"autoCreate,omitempty" toml:"autoCreate,omitempty" yaml:"autoCreate,omitempty"`
	// ConnectionID is populated internally from the unresolved runtime source.
	ConnectionID string `json:"-" toml:"-" yaml:"-"`
}

// +k8s:deepcopy-gen=true

// ExternalActionBinding pins a lifecycle key to one immutable credential set.
type ExternalActionBinding struct {
	Mode string `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty"`
	// OnMissing controls a resolve binding that is absent or expired. The
	// default "fail" preserves retry semantics; "skip" silently skips this
	// observation without completing its deduplication key, so a later poll can
	// retry after the lifecycle start creates the binding. The
	// "resolveOrCreate" mode is intended for reconciliation observations: it
	// resolves an existing lifecycle binding, or selects one allowed credential
	// set from Match and creates the missing binding without hard-coding a
	// destination in middleware code.
	OnMissing             string                       `json:"onMissing,omitempty" toml:"onMissing,omitempty" yaml:"onMissing,omitempty" export:"true"`
	KeyTemplate           string                       `json:"keyTemplate,omitempty" toml:"keyTemplate,omitempty" yaml:"keyTemplate,omitempty"`
	CredentialSet         string                       `json:"credentialSet,omitempty" toml:"credentialSet,omitempty" yaml:"credentialSet,omitempty"`
	AllowedCredentialSets []string                     `json:"allowedCredentialSets,omitempty" toml:"allowedCredentialSets,omitempty" yaml:"allowedCredentialSets,omitempty"`
	Match                 ExternalActionBindingMatch   `json:"match,omitempty" toml:"match,omitempty" yaml:"match,omitempty" export:"true"`
	TTL                   ptypes.Duration              `json:"ttl,omitempty" toml:"ttl,omitempty" yaml:"ttl,omitempty" export:"true"`
	Memory                *ExternalActionBindingMemory `json:"memory,omitempty" toml:"memory,omitempty" yaml:"memory,omitempty" export:"true"`
	MySQL                 *ExternalActionBindingMySQL  `json:"mysql,omitempty" toml:"mysql,omitempty" yaml:"mysql,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionBindingMatch compares source identity fields with immutable credential parameters.
type ExternalActionBindingMatch struct {
	Headers               map[string]string `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	BodyJSONPaths         map[string]string `json:"bodyJSONPaths,omitempty" toml:"bodyJSONPaths,omitempty" yaml:"bodyJSONPaths,omitempty"`
	ResponseBodyJSONPaths map[string]string `json:"responseBodyJSONPaths,omitempty" toml:"responseBodyJSONPaths,omitempty" yaml:"responseBodyJSONPaths,omitempty"`
	Templates             map[string]string `json:"templates,omitempty" toml:"templates,omitempty" yaml:"templates,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionBindingMemory configures process-local lifecycle bindings.
type ExternalActionBindingMemory struct {
	MaxEntries int `json:"maxEntries,omitempty" toml:"maxEntries,omitempty" yaml:"maxEntries,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionBindingMySQL configures shared lifecycle bindings across Traefik replicas and restarts.
type ExternalActionBindingMySQL struct {
	DSN        string `json:"dsn,omitempty" toml:"dsn,omitempty" yaml:"dsn,omitempty" loggable:"false"`
	DSNFile    string `json:"dsnFile,omitempty" toml:"dsnFile,omitempty" yaml:"dsnFile,omitempty"`
	Table      string `json:"table,omitempty" toml:"table,omitempty" yaml:"table,omitempty"`
	AutoCreate bool   `json:"autoCreate,omitempty" toml:"autoCreate,omitempty" yaml:"autoCreate,omitempty"`
	// ConnectionID is populated internally from the unresolved runtime source.
	ConnectionID string `json:"-" toml:"-" yaml:"-"`
}

// +k8s:deepcopy-gen=true

// ExternalActionHTTP defines an HTTP action target.
type ExternalActionHTTP struct {
	URL                string                      `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty"`
	Method             string                      `json:"method,omitempty" toml:"method,omitempty" yaml:"method,omitempty"`
	Headers            map[string]string           `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	Auth               *ExternalActionAuth         `json:"auth,omitempty" toml:"auth,omitempty" yaml:"auth,omitempty" export:"true"`
	TLS                *ExternalActionTLS          `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	ResponseValidation *ExternalActionHTTPResponse `json:"responseValidation,omitempty" toml:"responseValidation,omitempty" yaml:"responseValidation,omitempty" export:"true"`
	MaxResponseBytes   int64                       `json:"maxResponseBytes,omitempty" toml:"maxResponseBytes,omitempty" yaml:"maxResponseBytes,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionHTTPResponse validates HTTP and application-level target responses.
type ExternalActionHTTPResponse struct {
	Statuses        []string `json:"statuses,omitempty" toml:"statuses,omitempty" yaml:"statuses,omitempty"`
	SuccessJSONPath string   `json:"successJSONPath,omitempty" toml:"successJSONPath,omitempty" yaml:"successJSONPath,omitempty"`
	SuccessValues   []string `json:"successValues,omitempty" toml:"successValues,omitempty" yaml:"successValues,omitempty"`
	BodyRegex       string   `json:"bodyRegex,omitempty" toml:"bodyRegex,omitempty" yaml:"bodyRegex,omitempty"`
	MessageJSONPath string   `json:"messageJSONPath,omitempty" toml:"messageJSONPath,omitempty" yaml:"messageJSONPath,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionExec defines a direct, no-shell process action.
type ExternalActionExec struct {
	Command string            `json:"command,omitempty" toml:"command,omitempty" yaml:"command,omitempty"`
	Args    []string          `json:"args,omitempty" toml:"args,omitempty" yaml:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty" toml:"env,omitempty" yaml:"env,omitempty"`
	Stdin   string            `json:"stdin,omitempty" toml:"stdin,omitempty" yaml:"stdin,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionAuth defines target authentication.
type ExternalActionAuth struct {
	CredentialSet  string            `json:"credentialSet,omitempty" toml:"credentialSet,omitempty" yaml:"credentialSet,omitempty"`
	Type           string            `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty"`
	Username       string            `json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty"`
	Password       string            `json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty" loggable:"false"`
	PasswordFile   string            `json:"passwordFile,omitempty" toml:"passwordFile,omitempty" yaml:"passwordFile,omitempty"`
	Token          string            `json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	TokenFile      string            `json:"tokenFile,omitempty" toml:"tokenFile,omitempty" yaml:"tokenFile,omitempty"`
	Header         string            `json:"header,omitempty" toml:"header,omitempty" yaml:"header,omitempty"`
	Scheme         string            `json:"scheme,omitempty" toml:"scheme,omitempty" yaml:"scheme,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty" toml:"parameters,omitempty" yaml:"parameters,omitempty" loggable:"false"`
	ParameterFiles map[string]string `json:"parameterFiles,omitempty" toml:"parameterFiles,omitempty" yaml:"parameterFiles,omitempty"`
}

// +k8s:deepcopy-gen=true

// ExternalActionTLS defines optional target TLS trust and client authentication.
type ExternalActionTLS struct {
	CAFile             string `json:"caFile,omitempty" toml:"caFile,omitempty" yaml:"caFile,omitempty"`
	CertFile           string `json:"certFile,omitempty" toml:"certFile,omitempty" yaml:"certFile,omitempty"`
	KeyFile            string `json:"keyFile,omitempty" toml:"keyFile,omitempty" yaml:"keyFile,omitempty"`
	ServerName         string `json:"serverName,omitempty" toml:"serverName,omitempty" yaml:"serverName,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// ExternalActionQueue controls bounded asynchronous dispatch.
type ExternalActionQueue struct {
	Capacity       int             `json:"capacity,omitempty" toml:"capacity,omitempty" yaml:"capacity,omitempty" export:"true"`
	Workers        int             `json:"workers,omitempty" toml:"workers,omitempty" yaml:"workers,omitempty" export:"true"`
	MaxQueuedBytes int64           `json:"maxQueuedBytes,omitempty" toml:"maxQueuedBytes,omitempty" yaml:"maxQueuedBytes,omitempty" export:"true"`
	EnqueueWait    ptypes.Duration `json:"enqueueWait,omitempty" toml:"enqueueWait,omitempty" yaml:"enqueueWait,omitempty" export:"true"`
}
