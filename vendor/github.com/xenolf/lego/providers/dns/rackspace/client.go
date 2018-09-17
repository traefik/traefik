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

// Identity  Identity
type Identity struct {
	Access struct {
		ServiceCatalog []struct {
			Endpoints []struct {
				PublicURL string `json:"publicURL"`
				TenantID  string `json:"tenantId"`
			} `json:"endpoints"`
			Name string `json:"name"`
		} `json:"serviceCatalog"`
		Token struct {
			ID string `json:"id"`
		} `json:"token"`
	} `json:"access"`
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
