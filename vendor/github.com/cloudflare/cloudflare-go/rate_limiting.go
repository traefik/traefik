package cloudflare

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// RateLimit is a policy than can be applied to limit traffic within a customer domain
type RateLimit struct {
	ID          string                  `json:"id,omitempty"`
	Disabled    bool                    `json:"disabled,omitempty"`
	Description string                  `json:"description,omitempty"`
	Match       RateLimitTrafficMatcher `json:"match"`
	Bypass      []RateLimitKeyValue     `json:"bypass,omitempty"`
	Threshold   int                     `json:"threshold"`
	Period      int                     `json:"period"`
	Action      RateLimitAction         `json:"action"`
	Correlate   *RateLimitCorrelate     `json:"correlate,omitempty"`
}

// RateLimitTrafficMatcher contains the rules that will be used to apply a rate limit to traffic
type RateLimitTrafficMatcher struct {
	Request  RateLimitRequestMatcher  `json:"request"`
	Response RateLimitResponseMatcher `json:"response"`
}

// RateLimitRequestMatcher contains the matching rules pertaining to requests
type RateLimitRequestMatcher struct {
	Methods    []string `json:"methods,omitempty"`
	Schemes    []string `json:"schemes,omitempty"`
	URLPattern string   `json:"url,omitempty"`
}

// RateLimitResponseMatcher contains the matching rules pertaining to responses
type RateLimitResponseMatcher struct {
	Statuses      []int                            `json:"status,omitempty"`
	OriginTraffic *bool                            `json:"origin_traffic,omitempty"` // api defaults to true so we need an explicit empty value
	Headers       []RateLimitResponseMatcherHeader `json:"headers,omitempty"`
}

// RateLimitResponseMatcherHeader contains the structure of the origin
// HTTP headers used in request matcher checks.
type RateLimitResponseMatcherHeader struct {
	Name  string `json:"name"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

// RateLimitKeyValue is k-v formatted as expected in the rate limit description
type RateLimitKeyValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RateLimitAction is the action that will be taken when the rate limit threshold is reached
type RateLimitAction struct {
	Mode     string                   `json:"mode"`
	Timeout  int                      `json:"timeout"`
	Response *RateLimitActionResponse `json:"response"`
}

// RateLimitActionResponse is the response that will be returned when rate limit action is triggered
type RateLimitActionResponse struct {
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

// RateLimitCorrelate pertainings to NAT support
type RateLimitCorrelate struct {
	By string `json:"by"`
}

type rateLimitResponse struct {
	Response
	Result RateLimit `json:"result"`
}

type rateLimitListResponse struct {
	Response
	Result     []RateLimit `json:"result"`
	ResultInfo ResultInfo  `json:"result_info"`
}

// CreateRateLimit creates a new rate limit for a zone.
//
// API reference: https://api.cloudflare.com/#rate-limits-for-a-zone-create-a-ratelimit
func (api *API) CreateRateLimit(zoneID string, limit RateLimit) (RateLimit, error) {
	uri := "/zones/" + zoneID + "/rate_limits"
	res, err := api.makeRequest("POST", uri, limit)
	if err != nil {
		return RateLimit{}, errors.Wrap(err, errMakeRequestError)
	}
	var r rateLimitResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return RateLimit{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ListRateLimits returns Rate Limits for a zone, paginated according to the provided options
//
// API reference: https://api.cloudflare.com/#rate-limits-for-a-zone-list-rate-limits
func (api *API) ListRateLimits(zoneID string, pageOpts PaginationOptions) ([]RateLimit, ResultInfo, error) {
	v := url.Values{}
	if pageOpts.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(pageOpts.PerPage))
	}
	if pageOpts.Page > 0 {
		v.Set("page", strconv.Itoa(pageOpts.Page))
	}

	uri := "/zones/" + zoneID + "/rate_limits"
	if len(v) > 0 {
		uri = uri + "?" + v.Encode()
	}

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []RateLimit{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	var r rateLimitListResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []RateLimit{}, ResultInfo{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, r.ResultInfo, nil
}

// ListAllRateLimits returns all Rate Limits for a zone.
//
// API reference: https://api.cloudflare.com/#rate-limits-for-a-zone-list-rate-limits
func (api *API) ListAllRateLimits(zoneID string) ([]RateLimit, error) {
	pageOpts := PaginationOptions{
		PerPage: 100, // this is the max page size allowed
		Page:    1,
	}

	allRateLimits := make([]RateLimit, 0)
	for {
		rateLimits, resultInfo, err := api.ListRateLimits(zoneID, pageOpts)
		if err != nil {
			return []RateLimit{}, err
		}
		allRateLimits = append(allRateLimits, rateLimits...)
		// total pages is not returned on this call
		// if number of records is less than the max, this must be the last page
		// in case TotalCount % PerPage = 0, the last request will return an empty list
		if resultInfo.Count < resultInfo.PerPage {
			break
		}
		// continue with the next page
		pageOpts.Page = pageOpts.Page + 1
	}

	return allRateLimits, nil
}

// RateLimit fetches detail about one Rate Limit for a zone.
//
// API reference: https://api.cloudflare.com/#rate-limits-for-a-zone-rate-limit-details
func (api *API) RateLimit(zoneID, limitID string) (RateLimit, error) {
	uri := "/zones/" + zoneID + "/rate_limits/" + limitID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return RateLimit{}, errors.Wrap(err, errMakeRequestError)
	}
	var r rateLimitResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return RateLimit{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// UpdateRateLimit lets you replace a Rate Limit for a zone.
//
// API reference: https://api.cloudflare.com/#rate-limits-for-a-zone-update-rate-limit
func (api *API) UpdateRateLimit(zoneID, limitID string, limit RateLimit) (RateLimit, error) {
	uri := "/zones/" + zoneID + "/rate_limits/" + limitID
	res, err := api.makeRequest("PUT", uri, limit)
	if err != nil {
		return RateLimit{}, errors.Wrap(err, errMakeRequestError)
	}
	var r rateLimitResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return RateLimit{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// DeleteRateLimit deletes a Rate Limit for a zone.
//
// API reference: https://api.cloudflare.com/#rate-limits-for-a-zone-delete-rate-limit
func (api *API) DeleteRateLimit(zoneID, limitID string) error {
	uri := "/zones/" + zoneID + "/rate_limits/" + limitID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	var r rateLimitResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}
	return nil
}
