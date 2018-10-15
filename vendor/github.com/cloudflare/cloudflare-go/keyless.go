package cloudflare

import "time"

// KeylessSSL represents Keyless SSL configuration.
type KeylessSSL struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Status      string    `json:"success"`
	Enabled     bool      `json:"enabled"`
	Permissions []string  `json:"permissions"`
	CreatedOn   time.Time `json:"created_on"`
	ModifiedOn  time.Time `json:"modifed_on"`
}

// KeylessSSLResponse represents the response from the Keyless SSL endpoint.
type KeylessSSLResponse struct {
	Response
	Result []KeylessSSL `json:"result"`
}

// CreateKeyless creates a new Keyless SSL configuration for the zone.
//
// API reference: https://api.cloudflare.com/#keyless-ssl-for-a-zone-create-a-keyless-ssl-configuration
func (api *API) CreateKeyless() {
}

// ListKeyless lists Keyless SSL configurations for a zone.
//
// API reference: https://api.cloudflare.com/#keyless-ssl-for-a-zone-list-keyless-ssls
func (api *API) ListKeyless() {
}

// Keyless provides the configuration for a given Keyless SSL identifier.
//
// API reference: https://api.cloudflare.com/#keyless-ssl-for-a-zone-keyless-ssl-details
func (api *API) Keyless() {
}

// UpdateKeyless updates an existing Keyless SSL configuration.
//
// API reference: https://api.cloudflare.com/#keyless-ssl-for-a-zone-update-keyless-configuration
func (api *API) UpdateKeyless() {
}

// DeleteKeyless deletes an existing Keyless SSL configuration.
//
// API reference: https://api.cloudflare.com/#keyless-ssl-for-a-zone-delete-keyless-configuration
func (api *API) DeleteKeyless() {
}
