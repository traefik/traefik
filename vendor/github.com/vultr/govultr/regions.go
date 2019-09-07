package govultr

import (
	"context"
	"net/http"
	"strconv"
)

// RegionService is the interface to interact with Region endpoints on the Vultr API
// Link: https://www.vultr.com/api/#regions
type RegionService interface {
	Availability(ctx context.Context, regionID int, planType string) ([]int, error)
	BareMetalAvailability(ctx context.Context, regionID int) ([]int, error)
	Vc2Availability(ctx context.Context, regionID int) ([]int, error)
	Vdc2Availability(ctx context.Context, regionID int) ([]int, error)
	List(ctx context.Context) ([]Region, error)
}

// RegionServiceHandler handles interaction with the region methods for the Vultr API
type RegionServiceHandler struct {
	Client *Client
}

// Region represents a Vultr region
type Region struct {
	RegionID     string `json:"DCID"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	Continent    string `json:"continent"`
	State        string `json:"state"`
	Ddos         bool   `json:"ddos_protection"`
	BlockStorage bool   `json:"block_storage"`
	RegionCode   string `json:"regioncode"`
}

// Availability retrieves a list of the VPSPLANIDs currently available for a given location.
func (r *RegionServiceHandler) Availability(ctx context.Context, regionID int, planType string) ([]int, error) {

	uri := "/v1/regions/availability"

	req, err := r.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("DCID", strconv.Itoa(regionID))

	// Optional planType filter
	if planType != "" {
		q.Add("type", planType)
	}
	req.URL.RawQuery = q.Encode()

	var regions []int
	err = r.Client.DoWithContext(ctx, req, &regions)

	if err != nil {
		return nil, err
	}

	return regions, nil
}

// BareMetalAvailability retrieve a list of the METALPLANIDs currently available for a given location.
func (r *RegionServiceHandler) BareMetalAvailability(ctx context.Context, regionID int) ([]int, error) {

	uri := "/v1/regions/availability_baremetal"

	regions, err := r.instanceAvailability(ctx, uri, regionID)

	if err != nil {
		return nil, err
	}

	return regions, nil
}

// Vc2Availability retrieve a list of the vc2 VPSPLANIDs currently available for a given location.
func (r *RegionServiceHandler) Vc2Availability(ctx context.Context, regionID int) ([]int, error) {

	uri := "/v1/regions/availability_vc2"

	regions, err := r.instanceAvailability(ctx, uri, regionID)

	if err != nil {
		return nil, err
	}

	return regions, nil
}

// Vdc2Availability retrieves a list of the vdc2 VPSPLANIDs currently available for a given location.
func (r *RegionServiceHandler) Vdc2Availability(ctx context.Context, regionID int) ([]int, error) {

	uri := "/v1/regions/availability_vdc2"

	regions, err := r.instanceAvailability(ctx, uri, regionID)

	if err != nil {
		return nil, err
	}

	return regions, nil
}

// List retrieves a list of all active regions
func (r *RegionServiceHandler) List(ctx context.Context) ([]Region, error) {

	uri := "/v1/regions/list"

	req, err := r.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var regionsMap map[string]Region
	err = r.Client.DoWithContext(ctx, req, &regionsMap)

	if err != nil {
		return nil, err
	}

	var region []Region
	for _, r := range regionsMap {
		region = append(region, r)
	}

	return region, nil
}

// instanceAvailability keeps the similar calls dry
func (r *RegionServiceHandler) instanceAvailability(ctx context.Context, uri string, regionID int) ([]int, error) {
	req, err := r.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("DCID", strconv.Itoa(regionID))
	req.URL.RawQuery = q.Encode()

	var regions []int
	err = r.Client.DoWithContext(ctx, req, &regions)

	if err != nil {
		return nil, err
	}

	return regions, nil
}
