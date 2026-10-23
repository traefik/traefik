package dynamic

import ptypes "github.com/traefik/paerser/types"

// QueryEnrichment prepares event-scoped metadata without delaying the business response.
// +k8s:deepcopy-gen=true
type QueryEnrichment struct {
	RuntimeValues  map[string]RuntimeValue                `json:"runtimeValues,omitempty" toml:"runtimeValues,omitempty" yaml:"runtimeValues,omitempty" label:"-" export:"true"`
	CredentialSets map[string]ExternalActionCredentialSet `json:"credentialSets,omitempty" toml:"credentialSets,omitempty" yaml:"credentialSets,omitempty" label:"-" export:"true"`
	InputTemplate  string                                 `json:"inputTemplate,omitempty" toml:"inputTemplate,omitempty" yaml:"inputTemplate,omitempty"`
	Queries        []EnrichmentQuery                      `json:"queries,omitempty" toml:"queries,omitempty" yaml:"queries,omitempty" label:"-" export:"true"`
	OutputTemplate string                                 `json:"outputTemplate,omitempty" toml:"outputTemplate,omitempty" yaml:"outputTemplate,omitempty"`
}

// EnrichmentQuery performs a bounded lookup and maps its response into event metadata.
// +k8s:deepcopy-gen=true
type EnrichmentQuery struct {
	Name           string             `json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	When           string             `json:"when,omitempty" toml:"when,omitempty" yaml:"when,omitempty"`
	HTTP           ExternalActionHTTP `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" export:"true"`
	Body           ExternalActionBody `json:"body,omitempty" toml:"body,omitempty" yaml:"body,omitempty" export:"true"`
	AllowedOrigins []string           `json:"allowedOrigins,omitempty" toml:"allowedOrigins,omitempty" yaml:"allowedOrigins,omitempty"`
	ResultTemplate string             `json:"resultTemplate,omitempty" toml:"resultTemplate,omitempty" yaml:"resultTemplate,omitempty"`
	Timeout        ptypes.Duration    `json:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" export:"true"`
	MaxConcurrent  int                `json:"maxConcurrent,omitempty" toml:"maxConcurrent,omitempty" yaml:"maxConcurrent,omitempty" export:"true"`
	CacheTTL       ptypes.Duration    `json:"cacheTTL,omitempty" toml:"cacheTTL,omitempty" yaml:"cacheTTL,omitempty" export:"true"`
	CacheEntries   int                `json:"cacheEntries,omitempty" toml:"cacheEntries,omitempty" yaml:"cacheEntries,omitempty" export:"true"`
	CacheBytes     int64              `json:"cacheBytes,omitempty" toml:"cacheBytes,omitempty" yaml:"cacheBytes,omitempty" export:"true"`
	// Pagination maps each page with ResultTemplate and returns the mapped pages as an array.
	Pagination *ExternalActionPollerPagination `json:"pagination,omitempty" toml:"pagination,omitempty" yaml:"pagination,omitempty" export:"true"`
}
