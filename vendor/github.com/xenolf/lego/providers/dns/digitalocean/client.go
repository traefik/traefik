package digitalocean

const defaultBaseURL = "https://api.digitalocean.com"

// txtRecordRequest represents the request body to DO's API to make a TXT record
type txtRecordRequest struct {
	RecordType string `json:"type"`
	Name       string `json:"name"`
	Data       string `json:"data"`
	TTL        int    `json:"ttl"`
}

// txtRecordResponse represents a response from DO's API after making a TXT record
type txtRecordResponse struct {
	DomainRecord struct {
		ID   int    `json:"id"`
		Type string `json:"type"`
		Name string `json:"name"`
		Data string `json:"data"`
	} `json:"domain_record"`
}

type digitalOceanAPIError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
