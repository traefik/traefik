package rackspace

// APIKeyCredentials API credential
type APIKeyCredentials struct {
	Username string `json:"username"`
	APIKey   string `json:"apiKey"`
}

// Auth auth credentials
type Auth struct {
	APIKeyCredentials `json:"RAX-KSKEY:apiKeyCredentials"`
}

// AuthData Auth data
type AuthData struct {
	Auth `json:"auth"`
}

// Identity Identity
type Identity struct {
	Access Access `json:"access"`
}

// Access Access
type Access struct {
	ServiceCatalog []ServiceCatalog `json:"serviceCatalog"`
	Token          Token            `json:"token"`
}

// Token Token
type Token struct {
	ID string `json:"id"`
}

// ServiceCatalog ServiceCatalog
type ServiceCatalog struct {
	Endpoints []Endpoint `json:"endpoints"`
	Name      string     `json:"name"`
}

// Endpoint Endpoint
type Endpoint struct {
	PublicURL string `json:"publicURL"`
	TenantID  string `json:"tenantId"`
}

// ZoneSearchResponse represents the response when querying Rackspace DNS zones
type ZoneSearchResponse struct {
	TotalEntries int          `json:"totalEntries"`
	HostedZones  []HostedZone `json:"domains"`
}

// HostedZone HostedZone
type HostedZone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Records is the list of records sent/received from the DNS API
type Records struct {
	Record []Record `json:"records"`
}

// Record represents a Rackspace DNS record
type Record struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`
	ID   string `json:"id,omitempty"`
}
