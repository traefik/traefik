package cloudflare

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// LoadBalancerPool represents a load balancer pool's properties.
type LoadBalancerPool struct {
	ID                string               `json:"id,omitempty"`
	CreatedOn         *time.Time           `json:"created_on,omitempty"`
	ModifiedOn        *time.Time           `json:"modified_on,omitempty"`
	Description       string               `json:"description"`
	Name              string               `json:"name"`
	Enabled           bool                 `json:"enabled"`
	MinimumOrigins    int                  `json:"minimum_origins,omitempty"`
	Monitor           string               `json:"monitor,omitempty"`
	Origins           []LoadBalancerOrigin `json:"origins"`
	NotificationEmail string               `json:"notification_email,omitempty"`

	// CheckRegions defines the geographic region(s) from where to run health-checks from - e.g. "WNAM", "WEU", "SAF", "SAM".
	// Providing a null/empty value means "all regions", which may not be available to all plan types.
	CheckRegions []string `json:"check_regions"`
}

type LoadBalancerOrigin struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Enabled bool    `json:"enabled"`
	Weight  float64 `json:"weight"`
}

// LoadBalancerMonitor represents a load balancer monitor's properties.
type LoadBalancerMonitor struct {
	ID            string              `json:"id,omitempty"`
	CreatedOn     *time.Time          `json:"created_on,omitempty"`
	ModifiedOn    *time.Time          `json:"modified_on,omitempty"`
	Type          string              `json:"type"`
	Description   string              `json:"description"`
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Header        map[string][]string `json:"header"`
	Timeout       int                 `json:"timeout"`
	Retries       int                 `json:"retries"`
	Interval      int                 `json:"interval"`
	ExpectedBody  string              `json:"expected_body"`
	ExpectedCodes string              `json:"expected_codes"`
}

// LoadBalancer represents a load balancer's properties.
type LoadBalancer struct {
	ID           string              `json:"id,omitempty"`
	CreatedOn    *time.Time          `json:"created_on,omitempty"`
	ModifiedOn   *time.Time          `json:"modified_on,omitempty"`
	Description  string              `json:"description"`
	Name         string              `json:"name"`
	TTL          int                 `json:"ttl,omitempty"`
	FallbackPool string              `json:"fallback_pool"`
	DefaultPools []string            `json:"default_pools"`
	RegionPools  map[string][]string `json:"region_pools"`
	PopPools     map[string][]string `json:"pop_pools"`
	Proxied      bool                `json:"proxied"`
	Persistence  string              `json:"session_affinity,omitempty"`
}

// loadBalancerPoolResponse represents the response from the load balancer pool endpoints.
type loadBalancerPoolResponse struct {
	Response
	Result LoadBalancerPool `json:"result"`
}

// loadBalancerPoolListResponse represents the response from the List Pools endpoint.
type loadBalancerPoolListResponse struct {
	Response
	Result     []LoadBalancerPool `json:"result"`
	ResultInfo ResultInfo         `json:"result_info"`
}

// loadBalancerMonitorResponse represents the response from the load balancer monitor endpoints.
type loadBalancerMonitorResponse struct {
	Response
	Result LoadBalancerMonitor `json:"result"`
}

// loadBalancerMonitorListResponse represents the response from the List Monitors endpoint.
type loadBalancerMonitorListResponse struct {
	Response
	Result     []LoadBalancerMonitor `json:"result"`
	ResultInfo ResultInfo            `json:"result_info"`
}

// loadBalancerResponse represents the response from the load balancer endpoints.
type loadBalancerResponse struct {
	Response
	Result LoadBalancer `json:"result"`
}

// loadBalancerListResponse represents the response from the List Load Balancers endpoint.
type loadBalancerListResponse struct {
	Response
	Result     []LoadBalancer `json:"result"`
	ResultInfo ResultInfo     `json:"result_info"`
}

// CreateLoadBalancerPool creates a new load balancer pool.
//
// API reference: https://api.cloudflare.com/#load-balancer-pools-create-a-pool
func (api *API) CreateLoadBalancerPool(pool LoadBalancerPool) (LoadBalancerPool, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/pools"
	res, err := api.makeRequest("POST", uri, pool)
	if err != nil {
		return LoadBalancerPool{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerPoolResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancerPool{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ListLoadBalancerPools lists load balancer pools connected to an account.
//
// API reference: https://api.cloudflare.com/#load-balancer-pools-list-pools
func (api *API) ListLoadBalancerPools() ([]LoadBalancerPool, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/pools"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerPoolListResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// LoadBalancerPoolDetails returns the details for a load balancer pool.
//
// API reference: https://api.cloudflare.com/#load-balancer-pools-pool-details
func (api *API) LoadBalancerPoolDetails(poolID string) (LoadBalancerPool, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/pools/" + poolID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return LoadBalancerPool{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerPoolResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancerPool{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// DeleteLoadBalancerPool disables and deletes a load balancer pool.
//
// API reference: https://api.cloudflare.com/#load-balancer-pools-delete-a-pool
func (api *API) DeleteLoadBalancerPool(poolID string) error {
	uri := api.userBaseURL("/user") + "/load_balancers/pools/" + poolID
	if _, err := api.makeRequest("DELETE", uri, nil); err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	return nil
}

// ModifyLoadBalancerPool modifies a configured load balancer pool.
//
// API reference: https://api.cloudflare.com/#load-balancer-pools-modify-a-pool
func (api *API) ModifyLoadBalancerPool(pool LoadBalancerPool) (LoadBalancerPool, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/pools/" + pool.ID
	res, err := api.makeRequest("PUT", uri, pool)
	if err != nil {
		return LoadBalancerPool{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerPoolResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancerPool{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// CreateLoadBalancerMonitor creates a new load balancer monitor.
//
// API reference: https://api.cloudflare.com/#load-balancer-monitors-create-a-monitor
func (api *API) CreateLoadBalancerMonitor(monitor LoadBalancerMonitor) (LoadBalancerMonitor, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/monitors"
	res, err := api.makeRequest("POST", uri, monitor)
	if err != nil {
		return LoadBalancerMonitor{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerMonitorResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancerMonitor{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ListLoadBalancerMonitors lists load balancer monitors connected to an account.
//
// API reference: https://api.cloudflare.com/#load-balancer-monitors-list-monitors
func (api *API) ListLoadBalancerMonitors() ([]LoadBalancerMonitor, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/monitors"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerMonitorListResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// LoadBalancerMonitorDetails returns the details for a load balancer monitor.
//
// API reference: https://api.cloudflare.com/#load-balancer-monitors-monitor-details
func (api *API) LoadBalancerMonitorDetails(monitorID string) (LoadBalancerMonitor, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/monitors/" + monitorID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return LoadBalancerMonitor{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerMonitorResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancerMonitor{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// DeleteLoadBalancerMonitor disables and deletes a load balancer monitor.
//
// API reference: https://api.cloudflare.com/#load-balancer-monitors-delete-a-monitor
func (api *API) DeleteLoadBalancerMonitor(monitorID string) error {
	uri := api.userBaseURL("/user") + "/load_balancers/monitors/" + monitorID
	if _, err := api.makeRequest("DELETE", uri, nil); err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	return nil
}

// ModifyLoadBalancerMonitor modifies a configured load balancer monitor.
//
// API reference: https://api.cloudflare.com/#load-balancer-monitors-modify-a-monitor
func (api *API) ModifyLoadBalancerMonitor(monitor LoadBalancerMonitor) (LoadBalancerMonitor, error) {
	uri := api.userBaseURL("/user") + "/load_balancers/monitors/" + monitor.ID
	res, err := api.makeRequest("PUT", uri, monitor)
	if err != nil {
		return LoadBalancerMonitor{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerMonitorResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancerMonitor{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// CreateLoadBalancer creates a new load balancer.
//
// API reference: https://api.cloudflare.com/#load-balancers-create-a-load-balancer
func (api *API) CreateLoadBalancer(zoneID string, lb LoadBalancer) (LoadBalancer, error) {
	uri := "/zones/" + zoneID + "/load_balancers"
	res, err := api.makeRequest("POST", uri, lb)
	if err != nil {
		return LoadBalancer{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancer{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ListLoadBalancers lists load balancers configured on a zone.
//
// API reference: https://api.cloudflare.com/#load-balancers-list-load-balancers
func (api *API) ListLoadBalancers(zoneID string) ([]LoadBalancer, error) {
	uri := "/zones/" + zoneID + "/load_balancers"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerListResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// LoadBalancerDetails returns the details for a load balancer.
//
// API reference: https://api.cloudflare.com/#load-balancers-load-balancer-details
func (api *API) LoadBalancerDetails(zoneID, lbID string) (LoadBalancer, error) {
	uri := "/zones/" + zoneID + "/load_balancers/" + lbID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return LoadBalancer{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancer{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// DeleteLoadBalancer disables and deletes a load balancer.
//
// API reference: https://api.cloudflare.com/#load-balancers-delete-a-load-balancer
func (api *API) DeleteLoadBalancer(zoneID, lbID string) error {
	uri := "/zones/" + zoneID + "/load_balancers/" + lbID
	if _, err := api.makeRequest("DELETE", uri, nil); err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	return nil
}

// ModifyLoadBalancer modifies a configured load balancer.
//
// API reference: https://api.cloudflare.com/#load-balancers-modify-a-load-balancer
func (api *API) ModifyLoadBalancer(zoneID string, lb LoadBalancer) (LoadBalancer, error) {
	uri := "/zones/" + zoneID + "/load_balancers/" + lb.ID
	res, err := api.makeRequest("PUT", uri, lb)
	if err != nil {
		return LoadBalancer{}, errors.Wrap(err, errMakeRequestError)
	}
	var r loadBalancerResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return LoadBalancer{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}
