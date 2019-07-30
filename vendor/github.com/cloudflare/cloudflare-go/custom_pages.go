package cloudflare

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// CustomPage represents a custom page configuration.
type CustomPage struct {
	CreatedOn      time.Time   `json:"created_on"`
	ModifiedOn     time.Time   `json:"modified_on"`
	URL            interface{} `json:"url"`
	State          string      `json:"state"`
	RequiredTokens []string    `json:"required_tokens"`
	PreviewTarget  string      `json:"preview_target"`
	Description    string      `json:"description"`
	ID             string      `json:"id"`
}

// CustomPageResponse represents the response from the custom pages endpoint.
type CustomPageResponse struct {
	Response
	Result []CustomPage `json:"result"`
}

// CustomPageDetailResponse represents the response from the custom page endpoint.
type CustomPageDetailResponse struct {
	Response
	Result CustomPage `json:"result"`
}

// CustomPageOptions is used to determine whether or not the operation
// should take place on an account or zone level based on which is
// provided to the function.
//
// A non-empty value denotes desired use.
type CustomPageOptions struct {
	AccountID string
	ZoneID    string
}

// CustomPageParameters is used to update a particular custom page with
// the values provided.
type CustomPageParameters struct {
	URL   interface{} `json:"url"`
	State string      `json:"state"`
}

// CustomPages lists custom pages for a zone or account.
//
// Zone API reference: https://api.cloudflare.com/#custom-pages-for-a-zone-list-available-custom-pages
// Account API reference: https://api.cloudflare.com/#custom-pages-account--list-custom-pages
func (api *API) CustomPages(options *CustomPageOptions) ([]CustomPage, error) {
	var (
		pageType, identifier string
	)

	if options.AccountID == "" && options.ZoneID == "" {
		return nil, errors.New("either account ID or zone ID must be provided")
	}

	if options.AccountID != "" && options.ZoneID != "" {
		return nil, errors.New("account ID and zone ID are mutually exclusive")
	}

	// Should the account ID be defined, treat this as an account level operation.
	if options.AccountID != "" {
		pageType = "accounts"
		identifier = options.AccountID
	} else {
		pageType = "zones"
		identifier = options.ZoneID
	}

	uri := fmt.Sprintf("/%s/%s/custom_pages", pageType, identifier)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	var customPageResponse CustomPageResponse
	err = json.Unmarshal(res, &customPageResponse)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return customPageResponse.Result, nil
}

// CustomPage lists a single custom page based on the ID.
//
// Zone API reference: https://api.cloudflare.com/#custom-pages-for-a-zone-custom-page-details
// Account API reference: https://api.cloudflare.com/#custom-pages-account--custom-page-details
func (api *API) CustomPage(options *CustomPageOptions, customPageID string) (CustomPage, error) {
	var (
		pageType, identifier string
	)

	if options.AccountID == "" && options.ZoneID == "" {
		return CustomPage{}, errors.New("either account ID or zone ID must be provided")
	}

	if options.AccountID != "" && options.ZoneID != "" {
		return CustomPage{}, errors.New("account ID and zone ID are mutually exclusive")
	}

	// Should the account ID be defined, treat this as an account level operation.
	if options.AccountID != "" {
		pageType = "accounts"
		identifier = options.AccountID
	} else {
		pageType = "zones"
		identifier = options.ZoneID
	}

	uri := fmt.Sprintf("/%s/%s/custom_pages/%s", pageType, identifier, customPageID)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return CustomPage{}, errors.Wrap(err, errMakeRequestError)
	}

	var customPageResponse CustomPageDetailResponse
	err = json.Unmarshal(res, &customPageResponse)
	if err != nil {
		return CustomPage{}, errors.Wrap(err, errUnmarshalError)
	}

	return customPageResponse.Result, nil
}

// UpdateCustomPage updates a single custom page setting.
//
// Zone API reference: https://api.cloudflare.com/#custom-pages-for-a-zone-update-custom-page-url
// Account API reference: https://api.cloudflare.com/#custom-pages-account--update-custom-page
func (api *API) UpdateCustomPage(options *CustomPageOptions, customPageID string, pageParameters CustomPageParameters) (CustomPage, error) {
	var (
		pageType, identifier string
	)

	if options.AccountID == "" && options.ZoneID == "" {
		return CustomPage{}, errors.New("either account ID or zone ID must be provided")
	}

	if options.AccountID != "" && options.ZoneID != "" {
		return CustomPage{}, errors.New("account ID and zone ID are mutually exclusive")
	}

	// Should the account ID be defined, treat this as an account level operation.
	if options.AccountID != "" {
		pageType = "accounts"
		identifier = options.AccountID
	} else {
		pageType = "zones"
		identifier = options.ZoneID
	}

	uri := fmt.Sprintf("/%s/%s/custom_pages/%s", pageType, identifier, customPageID)

	res, err := api.makeRequest("PUT", uri, pageParameters)
	if err != nil {
		return CustomPage{}, errors.Wrap(err, errMakeRequestError)
	}

	var customPageResponse CustomPageDetailResponse
	err = json.Unmarshal(res, &customPageResponse)
	if err != nil {
		return CustomPage{}, errors.Wrap(err, errUnmarshalError)
	}

	return customPageResponse.Result, nil
}
