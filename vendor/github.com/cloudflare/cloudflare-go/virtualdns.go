package cloudflare

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// VirtualDNS represents a Virtual DNS configuration.
type VirtualDNS struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	OriginIPs            []string `json:"origin_ips"`
	VirtualDNSIPs        []string `json:"virtual_dns_ips"`
	MinimumCacheTTL      uint     `json:"minimum_cache_ttl"`
	MaximumCacheTTL      uint     `json:"maximum_cache_ttl"`
	DeprecateAnyRequests bool     `json:"deprecate_any_requests"`
	ModifiedOn           string   `json:"modified_on"`
}

// VirtualDNSAnalyticsMetrics respresents a group of aggregated Virtual DNS metrics.
type VirtualDNSAnalyticsMetrics struct {
	QueryCount         *int64   `json:"queryCount"`
	UncachedCount      *int64   `json:"uncachedCount"`
	StaleCount         *int64   `json:"staleCount"`
	ResponseTimeAvg    *float64 `json:"responseTimeAvg"`
	ResponseTimeMedian *float64 `json:"responseTimeMedian"`
	ResponseTime90th   *float64 `json:"responseTime90th"`
	ResponseTime99th   *float64 `json:"responseTime99th"`
}

// VirtualDNSAnalytics represents a set of aggregated Virtual DNS metrics.
// TODO: Add the queried data and not only the aggregated values.
type VirtualDNSAnalytics struct {
	Totals VirtualDNSAnalyticsMetrics `json:"totals"`
	Min    VirtualDNSAnalyticsMetrics `json:"min"`
	Max    VirtualDNSAnalyticsMetrics `json:"max"`
}

// VirtualDNSUserAnalyticsOptions represents range and dimension selection on analytics endpoint
type VirtualDNSUserAnalyticsOptions struct {
	Metrics []string
	Since   *time.Time
	Until   *time.Time
}

// VirtualDNSResponse represents a Virtual DNS response.
type VirtualDNSResponse struct {
	Response
	Result *VirtualDNS `json:"result"`
}

// VirtualDNSListResponse represents an array of Virtual DNS responses.
type VirtualDNSListResponse struct {
	Response
	Result []*VirtualDNS `json:"result"`
}

// VirtualDNSAnalyticsResponse represents a Virtual DNS analytics response.
type VirtualDNSAnalyticsResponse struct {
	Response
	Result VirtualDNSAnalytics `json:"result"`
}

// CreateVirtualDNS creates a new Virtual DNS cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--create-a-virtual-dns-cluster
func (api *API) CreateVirtualDNS(v *VirtualDNS) (*VirtualDNS, error) {
	res, err := api.makeRequest("POST", "/user/virtual_dns", v)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// VirtualDNS fetches a single virtual DNS cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--get-a-virtual-dns-cluster
func (api *API) VirtualDNS(virtualDNSID string) (*VirtualDNS, error) {
	uri := "/user/virtual_dns/" + virtualDNSID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// ListVirtualDNS lists the virtual DNS clusters associated with an account.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--get-virtual-dns-clusters
func (api *API) ListVirtualDNS() ([]*VirtualDNS, error) {
	res, err := api.makeRequest("GET", "/user/virtual_dns", nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSListResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// UpdateVirtualDNS updates a Virtual DNS cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--modify-a-virtual-dns-cluster
func (api *API) UpdateVirtualDNS(virtualDNSID string, vv VirtualDNS) error {
	uri := "/user/virtual_dns/" + virtualDNSID
	res, err := api.makeRequest("PUT", uri, vv)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}

	return nil
}

// DeleteVirtualDNS deletes a Virtual DNS cluster. Note that this cannot be
// undone, and will stop all traffic to that cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--delete-a-virtual-dns-cluster
func (api *API) DeleteVirtualDNS(virtualDNSID string) error {
	uri := "/user/virtual_dns/" + virtualDNSID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}

	return nil
}

// encode encodes non-nil fields into URL encoded form.
func (o VirtualDNSUserAnalyticsOptions) encode() string {
	v := url.Values{}
	if o.Since != nil {
		v.Set("since", (*o.Since).UTC().Format(time.RFC3339))
	}
	if o.Until != nil {
		v.Set("until", (*o.Until).UTC().Format(time.RFC3339))
	}
	if o.Metrics != nil {
		v.Set("metrics", strings.Join(o.Metrics, ","))
	}
	return v.Encode()
}

// VirtualDNSUserAnalytics retrieves analytics report for a specified dimension and time range
func (api *API) VirtualDNSUserAnalytics(virtualDNSID string, o VirtualDNSUserAnalyticsOptions) (VirtualDNSAnalytics, error) {
	uri := "/user/virtual_dns/" + virtualDNSID + "/dns_analytics/report?" + o.encode()
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return VirtualDNSAnalytics{}, errors.Wrap(err, errMakeRequestError)
	}

	response := VirtualDNSAnalyticsResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return VirtualDNSAnalytics{}, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}
