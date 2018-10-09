package cloudflare

import "time"

// CustomPage represents a custom page configuration.
type CustomPage struct {
	CreatedOn      string    `json:"created_on"`
	ModifiedOn     time.Time `json:"modified_on"`
	URL            string    `json:"url"`
	State          string    `json:"state"`
	RequiredTokens []string  `json:"required_tokens"`
	PreviewTarget  string    `json:"preview_target"`
	Description    string    `json:"description"`
}

// CustomPageResponse represents the response from the custom pages endpoint.
type CustomPageResponse struct {
	Response
	Result []CustomPage `json:"result"`
}

// https://api.cloudflare.com/#custom-pages-for-a-zone-available-custom-pages
// GET /zones/:zone_identifier/custom_pages

// https://api.cloudflare.com/#custom-pages-for-a-zone-custom-page-details
// GET /zones/:zone_identifier/custom_pages/:identifier

// https://api.cloudflare.com/#custom-pages-for-a-zone-update-custom-page-url
// PUT /zones/:zone_identifier/custom_pages/:identifier
