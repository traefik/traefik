package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Filter holds the structure of the filter type.
type Filter struct {
	ID          string `json:"id,omitempty"`
	Expression  string `json:"expression"`
	Paused      bool   `json:"paused"`
	Description string `json:"description"`

	// Property is mentioned in documentation however isn't populated in
	// any of the API requests. For now, let's just omit it unless it's
	// provided.
	Ref string `json:"ref,omitempty"`
}

// FiltersDetailResponse is the API response that is returned
// for requesting all filters on a zone.
type FiltersDetailResponse struct {
	Result     []Filter `json:"result"`
	ResultInfo `json:"result_info"`
	Response
}

// FilterDetailResponse is the API response that is returned
// for requesting a single filter on a zone.
type FilterDetailResponse struct {
	Result     Filter `json:"result"`
	ResultInfo `json:"result_info"`
	Response
}

// FilterValidateExpression represents the JSON payload for checking
// an expression.
type FilterValidateExpression struct {
	Expression string `json:"expression"`
}

// FilterValidateExpressionResponse represents the API response for
// checking the expression. It conforms to the JSON API approach however
// we don't need all of the fields exposed.
type FilterValidateExpressionResponse struct {
	Success bool                                `json:"success"`
	Errors  []FilterValidationExpressionMessage `json:"errors"`
}

// FilterValidationExpressionMessage represents the API error message.
type FilterValidationExpressionMessage struct {
	Message string `json:"message"`
}

// Filter returns a single filter in a zone based on the filter ID.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/get/#get-by-filter-id
func (api *API) Filter(zoneID, filterID string) (Filter, error) {
	uri := fmt.Sprintf("/zones/%s/filters/%s", zoneID, filterID)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return Filter{}, errors.Wrap(err, errMakeRequestError)
	}

	var filterResponse FilterDetailResponse
	err = json.Unmarshal(res, &filterResponse)
	if err != nil {
		return Filter{}, errors.Wrap(err, errUnmarshalError)
	}

	return filterResponse.Result, nil
}

// Filters returns all filters for a zone.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/get/#get-all-filters
func (api *API) Filters(zoneID string, pageOpts PaginationOptions) ([]Filter, error) {
	uri := "/zones/" + zoneID + "/filters"
	v := url.Values{}

	if pageOpts.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(pageOpts.PerPage))
	}

	if pageOpts.Page > 0 {
		v.Set("page", strconv.Itoa(pageOpts.Page))
	}

	if len(v) > 0 {
		uri = uri + "?" + v.Encode()
	}

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []Filter{}, errors.Wrap(err, errMakeRequestError)
	}

	var filtersResponse FiltersDetailResponse
	err = json.Unmarshal(res, &filtersResponse)
	if err != nil {
		return []Filter{}, errors.Wrap(err, errUnmarshalError)
	}

	return filtersResponse.Result, nil
}

// CreateFilters creates new filters.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/post/
func (api *API) CreateFilters(zoneID string, filters []Filter) ([]Filter, error) {
	uri := "/zones/" + zoneID + "/filters"

	res, err := api.makeRequest("POST", uri, filters)
	if err != nil {
		return []Filter{}, errors.Wrap(err, errMakeRequestError)
	}

	var filtersResponse FiltersDetailResponse
	err = json.Unmarshal(res, &filtersResponse)
	if err != nil {
		return []Filter{}, errors.Wrap(err, errUnmarshalError)
	}

	return filtersResponse.Result, nil
}

// UpdateFilter updates a single filter.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/put/#update-a-single-filter
func (api *API) UpdateFilter(zoneID string, filter Filter) (Filter, error) {
	if filter.ID == "" {
		return Filter{}, errors.Errorf("filter ID cannot be empty")
	}

	uri := fmt.Sprintf("/zones/%s/filters/%s", zoneID, filter.ID)

	res, err := api.makeRequest("PUT", uri, filter)
	if err != nil {
		return Filter{}, errors.Wrap(err, errMakeRequestError)
	}

	var filterResponse FilterDetailResponse
	err = json.Unmarshal(res, &filterResponse)
	if err != nil {
		return Filter{}, errors.Wrap(err, errUnmarshalError)
	}

	return filterResponse.Result, nil
}

// UpdateFilters updates many filters at once.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/put/#update-multiple-filters
func (api *API) UpdateFilters(zoneID string, filters []Filter) ([]Filter, error) {
	for _, filter := range filters {
		if filter.ID == "" {
			return []Filter{}, errors.Errorf("filter ID cannot be empty")
		}
	}

	uri := "/zones/" + zoneID + "/filters"

	res, err := api.makeRequest("PUT", uri, filters)
	if err != nil {
		return []Filter{}, errors.Wrap(err, errMakeRequestError)
	}

	var filtersResponse FiltersDetailResponse
	err = json.Unmarshal(res, &filtersResponse)
	if err != nil {
		return []Filter{}, errors.Wrap(err, errUnmarshalError)
	}

	return filtersResponse.Result, nil
}

// DeleteFilter deletes a single filter.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/delete/#delete-a-single-filter
func (api *API) DeleteFilter(zoneID, filterID string) error {
	if filterID == "" {
		return errors.Errorf("filter ID cannot be empty")
	}

	uri := fmt.Sprintf("/zones/%s/filters/%s", zoneID, filterID)

	_, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	return nil
}

// DeleteFilters deletes multiple filters.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/delete/#delete-multiple-filters
func (api *API) DeleteFilters(zoneID string, filterIDs []string) error {
	ids := strings.Join(filterIDs, ",")
	uri := fmt.Sprintf("/zones/%s/filters?id=%s", zoneID, ids)

	_, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	return nil
}

// ValidateFilterExpression checks correctness of a filter expression.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-filters/validation/
func (api *API) ValidateFilterExpression(expression string) error {
	uri := fmt.Sprintf("/filters/validate-expr")
	expressionPayload := FilterValidateExpression{Expression: expression}

	_, err := api.makeRequest("POST", uri, expressionPayload)
	if err != nil {
		var filterValidationResponse FilterValidateExpressionResponse

		jsonErr := json.Unmarshal([]byte(err.Error()), &filterValidationResponse)
		if jsonErr != nil {
			return errors.Wrap(jsonErr, errUnmarshalError)
		}

		if filterValidationResponse.Success != true {
			// Unsure why but the API returns `errors` as an array but it only
			// ever shows the issue with one problem at a time ¯\_(ツ)_/¯
			return errors.Errorf(filterValidationResponse.Errors[0].Message)
		}
	}

	return nil
}
