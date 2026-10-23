package dynamic

import ptypes "github.com/traefik/paerser/types"

// +k8s:deepcopy-gen=true

// AccessGovernance holds the protocol-neutral access governance configuration.
// It records the identity, authorization, audit, and metering dimensions of an access event.
type AccessGovernance struct {
	RuntimeValues   map[string]RuntimeValue             `json:"runtimeValues,omitempty" toml:"runtimeValues,omitempty" yaml:"runtimeValues,omitempty" label:"-" export:"true"`
	Databases       map[string]AccessGovernanceDatabase `json:"databases,omitempty" toml:"databases,omitempty" yaml:"databases,omitempty" label:"-" export:"true"`
	DefaultDatabase string                              `json:"defaultDatabase,omitempty" toml:"defaultDatabase,omitempty" yaml:"defaultDatabase,omitempty" export:"true"`
	Routes          []AccessGovernanceRoute             `json:"routes,omitempty" toml:"routes,omitempty" yaml:"routes,omitempty" label:"-" export:"true"`
	Identity        []AccessGovernanceIdentitySource    `json:"identity,omitempty" toml:"identity,omitempty" yaml:"identity,omitempty" label:"-" export:"true"`
	Queue           AccessGovernanceQueue               `json:"queue,omitempty" toml:"queue,omitempty" yaml:"queue,omitempty" export:"true"`
	Capture         AccessGovernanceCapture             `json:"capture,omitempty" toml:"capture,omitempty" yaml:"capture,omitempty" export:"true"`
	Dimensions      AccessGovernanceDimensions          `json:"dimensions,omitempty" toml:"dimensions,omitempty" yaml:"dimensions,omitempty" export:"true"`
	Schema          AccessGovernanceSchema              `json:"schema,omitempty" toml:"schema,omitempty" yaml:"schema,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceDatabase defines one MySQL destination.
type AccessGovernanceDatabase struct {
	DSN                string          `json:"dsn,omitempty" toml:"dsn,omitempty" yaml:"dsn,omitempty" loggable:"false"`
	DSNFile            string          `json:"dsnFile,omitempty" toml:"dsnFile,omitempty" yaml:"dsnFile,omitempty"`
	HTTPTable          string          `json:"httpTable,omitempty" toml:"httpTable,omitempty" yaml:"httpTable,omitempty"`
	NetworkTable       string          `json:"networkTable,omitempty" toml:"networkTable,omitempty" yaml:"networkTable,omitempty"`
	MaxOpenConnections int             `json:"maxOpenConnections,omitempty" toml:"maxOpenConnections,omitempty" yaml:"maxOpenConnections,omitempty" export:"true"`
	MaxIdleConnections int             `json:"maxIdleConnections,omitempty" toml:"maxIdleConnections,omitempty" yaml:"maxIdleConnections,omitempty" export:"true"`
	ConnectionMaxAge   ptypes.Duration `json:"connectionMaxAge,omitempty" toml:"connectionMaxAge,omitempty" yaml:"connectionMaxAge,omitempty" export:"true"`
	// ConnectionID is populated internally from the unresolved runtime source.
	// It is intentionally not configurable or serializable.
	ConnectionID string `json:"-" toml:"-" yaml:"-"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceRoute sends matching events to a named database.
type AccessGovernanceRoute struct {
	Name          string            `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Database      string            `json:"database,omitempty" toml:"database,omitempty" yaml:"database,omitempty"`
	Protocols     []string          `json:"protocols,omitempty" toml:"protocols,omitempty" yaml:"protocols,omitempty"`
	EntryPoints   []string          `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty"`
	Routers       []string          `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty"`
	Services      []string          `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty"`
	Methods       []string          `json:"methods,omitempty" toml:"methods,omitempty" yaml:"methods,omitempty"`
	Hosts         []string          `json:"hosts,omitempty" toml:"hosts,omitempty" yaml:"hosts,omitempty"`
	PathRegex     string            `json:"pathRegex,omitempty" toml:"pathRegex,omitempty" yaml:"pathRegex,omitempty"`
	Headers       map[string]string `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty"`
	BodyJSONPaths map[string]string `json:"bodyJSONPaths,omitempty" toml:"bodyJSONPaths,omitempty" yaml:"bodyJSONPaths,omitempty"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceIdentitySource defines an ordered principal lookup source.
type AccessGovernanceIdentitySource struct {
	Kind     string `json:"kind,omitempty" toml:"kind,omitempty" yaml:"kind,omitempty"`
	Name     string `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	JSONPath string `json:"jsonPath,omitempty" toml:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`
	Regex    string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	Group    int    `json:"group,omitempty" toml:"group,omitempty" yaml:"group,omitempty" export:"true"`
	Decode   string `json:"decode,omitempty" toml:"decode,omitempty" yaml:"decode,omitempty"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceQueue controls bounded asynchronous database writes.
type AccessGovernanceQueue struct {
	Capacity      int             `json:"capacity,omitempty" toml:"capacity,omitempty" yaml:"capacity,omitempty" export:"true"`
	Workers       int             `json:"workers,omitempty" toml:"workers,omitempty" yaml:"workers,omitempty" export:"true"`
	BatchSize     int             `json:"batchSize,omitempty" toml:"batchSize,omitempty" yaml:"batchSize,omitempty" export:"true"`
	FlushInterval ptypes.Duration `json:"flushInterval,omitempty" toml:"flushInterval,omitempty" yaml:"flushInterval,omitempty" export:"true"`
	EnqueueWait   ptypes.Duration `json:"enqueueWait,omitempty" toml:"enqueueWait,omitempty" yaml:"enqueueWait,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceCapture controls optional request metadata capture.
type AccessGovernanceCapture struct {
	BodyReadMode   string `json:"bodyReadMode,omitempty" toml:"bodyReadMode,omitempty" yaml:"bodyReadMode,omitempty"`
	MaxBodyBytes   int64  `json:"maxBodyBytes,omitempty" toml:"maxBodyBytes,omitempty" yaml:"maxBodyBytes,omitempty" export:"true"`
	StoreQuery     bool   `json:"storeQuery,omitempty" toml:"storeQuery,omitempty" yaml:"storeQuery,omitempty" export:"true"`
	StoreUserAgent bool   `json:"storeUserAgent,omitempty" toml:"storeUserAgent,omitempty" yaml:"storeUserAgent,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceDimensions supplies future-facing authentication, authorization, and billing dimensions.
type AccessGovernanceDimensions struct {
	TenantID       string            `json:"tenantID,omitempty" toml:"tenantID,omitempty" yaml:"tenantID,omitempty"`
	AuthnMethod    string            `json:"authnMethod,omitempty" toml:"authnMethod,omitempty" yaml:"authnMethod,omitempty"`
	PolicyID       string            `json:"policyID,omitempty" toml:"policyID,omitempty" yaml:"policyID,omitempty"`
	BillingAccount string            `json:"billingAccount,omitempty" toml:"billingAccount,omitempty" yaml:"billingAccount,omitempty"`
	Meter          string            `json:"meter,omitempty" toml:"meter,omitempty" yaml:"meter,omitempty"`
	Labels         map[string]string `json:"labels,omitempty" toml:"labels,omitempty" yaml:"labels,omitempty"`
}

// +k8s:deepcopy-gen=true

// AccessGovernanceSchema controls schema bootstrapping and table names.
type AccessGovernanceSchema struct {
	AutoCreate bool `json:"autoCreate,omitempty" toml:"autoCreate,omitempty" yaml:"autoCreate,omitempty" export:"true"`
}
