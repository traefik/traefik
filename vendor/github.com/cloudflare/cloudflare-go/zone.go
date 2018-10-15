package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// Owner describes the resource owner.
type Owner struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	OwnerType string `json:"owner_type"`
}

// Zone describes a Cloudflare zone.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// DevMode contains the time in seconds until development expires (if
	// positive) or since it expired (if negative). It will be 0 if never used.
	DevMode           int          `json:"development_mode"`
	OriginalNS        []string     `json:"original_name_servers"`
	OriginalRegistrar string       `json:"original_registrar"`
	OriginalDNSHost   string       `json:"original_dnshost"`
	CreatedOn         time.Time    `json:"created_on"`
	ModifiedOn        time.Time    `json:"modified_on"`
	NameServers       []string     `json:"name_servers"`
	Owner             Owner        `json:"owner"`
	Permissions       []string     `json:"permissions"`
	Plan              ZoneRatePlan `json:"plan"`
	PlanPending       ZoneRatePlan `json:"plan_pending,omitempty"`
	Status            string       `json:"status"`
	Paused            bool         `json:"paused"`
	Type              string       `json:"type"`
	Host              struct {
		Name    string
		Website string
	} `json:"host"`
	VanityNS    []string `json:"vanity_name_servers"`
	Betas       []string `json:"betas"`
	DeactReason string   `json:"deactivation_reason"`
	Meta        ZoneMeta `json:"meta"`
}

// ZoneMeta describes metadata about a zone.
type ZoneMeta struct {
	// custom_certificate_quota is broken - sometimes it's a string, sometimes a number!
	// CustCertQuota     int    `json:"custom_certificate_quota"`
	PageRuleQuota     int  `json:"page_rule_quota"`
	WildcardProxiable bool `json:"wildcard_proxiable"`
	PhishingDetected  bool `json:"phishing_detected"`
}

// ZoneRatePlan contains the plan information for a zone.
type ZoneRatePlan struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name,omitempty"`
	Price      int                      `json:"price,omitempty"`
	Currency   string                   `json:"currency,omitempty"`
	Duration   int                      `json:"duration,omitempty"`
	Frequency  string                   `json:"frequency,omitempty"`
	Components []zoneRatePlanComponents `json:"components,omitempty"`
}

type zoneRatePlanComponents struct {
	Name      string `json:"name"`
	Default   int    `json:"Default"`
	UnitPrice int    `json:"unit_price"`
}

// ZoneID contains only the zone ID.
type ZoneID struct {
	ID string `json:"id"`
}

// ZoneResponse represents the response from the Zone endpoint containing a single zone.
type ZoneResponse struct {
	Response
	Result Zone `json:"result"`
}

// ZonesResponse represents the response from the Zone endpoint containing an array of zones.
type ZonesResponse struct {
	Response
	Result []Zone `json:"result"`
}

// ZoneIDResponse represents the response from the Zone endpoint, containing only a zone ID.
type ZoneIDResponse struct {
	Response
	Result ZoneID `json:"result"`
}

// AvailableZoneRatePlansResponse represents the response from the Available Rate Plans endpoint.
type AvailableZoneRatePlansResponse struct {
	Response
	Result []ZoneRatePlan `json:"result"`
	ResultInfo
}

// ZoneRatePlanResponse represents the response from the Plan Details endpoint.
type ZoneRatePlanResponse struct {
	Response
	Result ZoneRatePlan `json:"result"`
}

// ZoneSetting contains settings for a zone.
type ZoneSetting struct {
	ID            string      `json:"id"`
	Editable      bool        `json:"editable"`
	ModifiedOn    string      `json:"modified_on"`
	Value         interface{} `json:"value"`
	TimeRemaining int         `json:"time_remaining"`
}

// ZoneSettingResponse represents the response from the Zone Setting endpoint.
type ZoneSettingResponse struct {
	Response
	Result []ZoneSetting `json:"result"`
}

// ZoneSSLSetting contains ssl setting for a zone.
type ZoneSSLSetting struct {
	ID                string `json:"id"`
	Editable          bool   `json:"editable"`
	ModifiedOn        string `json:"modified_on"`
	Value             string `json:"value"`
	CertificateStatus string `json:"certificate_status"`
}

// ZoneSettingResponse represents the response from the Zone SSL Setting endpoint.
type ZoneSSLSettingResponse struct {
	Response
	Result ZoneSSLSetting `json:"result"`
}

// ZoneAnalyticsData contains totals and timeseries analytics data for a zone.
type ZoneAnalyticsData struct {
	Totals     ZoneAnalytics   `json:"totals"`
	Timeseries []ZoneAnalytics `json:"timeseries"`
}

// zoneAnalyticsDataResponse represents the response from the Zone Analytics Dashboard endpoint.
type zoneAnalyticsDataResponse struct {
	Response
	Result ZoneAnalyticsData `json:"result"`
}

// ZoneAnalyticsColocation contains analytics data by datacenter.
type ZoneAnalyticsColocation struct {
	ColocationID string          `json:"colo_id"`
	Timeseries   []ZoneAnalytics `json:"timeseries"`
}

// zoneAnalyticsColocationResponse represents the response from the Zone Analytics By Co-location endpoint.
type zoneAnalyticsColocationResponse struct {
	Response
	Result []ZoneAnalyticsColocation `json:"result"`
}

// ZoneAnalytics contains analytics data for a zone.
type ZoneAnalytics struct {
	Since    time.Time `json:"since"`
	Until    time.Time `json:"until"`
	Requests struct {
		All         int            `json:"all"`
		Cached      int            `json:"cached"`
		Uncached    int            `json:"uncached"`
		ContentType map[string]int `json:"content_type"`
		Country     map[string]int `json:"country"`
		SSL         struct {
			Encrypted   int `json:"encrypted"`
			Unencrypted int `json:"unencrypted"`
		} `json:"ssl"`
		HTTPStatus map[string]int `json:"http_status"`
	} `json:"requests"`
	Bandwidth struct {
		All         int            `json:"all"`
		Cached      int            `json:"cached"`
		Uncached    int            `json:"uncached"`
		ContentType map[string]int `json:"content_type"`
		Country     map[string]int `json:"country"`
		SSL         struct {
			Encrypted   int `json:"encrypted"`
			Unencrypted int `json:"unencrypted"`
		} `json:"ssl"`
	} `json:"bandwidth"`
	Threats struct {
		All     int            `json:"all"`
		Country map[string]int `json:"country"`
		Type    map[string]int `json:"type"`
	} `json:"threats"`
	Pageviews struct {
		All           int            `json:"all"`
		SearchEngines map[string]int `json:"search_engines"`
	} `json:"pageviews"`
	Uniques struct {
		All int `json:"all"`
	}
}

// ZoneAnalyticsOptions represents the optional parameters in Zone Analytics
// endpoint requests.
type ZoneAnalyticsOptions struct {
	Since      *time.Time
	Until      *time.Time
	Continuous *bool
}

// PurgeCacheRequest represents the request format made to the purge endpoint.
type PurgeCacheRequest struct {
	Everything bool `json:"purge_everything,omitempty"`
	// Purge by filepath (exact match). Limit of 30
	Files []string `json:"files,omitempty"`
	// Purge by Tag (Enterprise only):
	// https://support.cloudflare.com/hc/en-us/articles/206596608-How-to-Purge-Cache-Using-Cache-Tags-Enterprise-only-
	Tags []string `json:"tags,omitempty"`
	// Purge by hostname - e.g. "assets.example.com"
	Hosts []string `json:"hosts,omitempty"`
}

// PurgeCacheResponse represents the response from the purge endpoint.
type PurgeCacheResponse struct {
	Response
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

// newZone describes a new zone.
type newZone struct {
	Name      string `json:"name"`
	JumpStart bool   `json:"jump_start"`
	// We use a pointer to get a nil type when the field is empty.
	// This allows us to completely omit this with json.Marshal().
	Organization *Organization `json:"organization,omitempty"`
}

// CreateZone creates a zone on an account.
//
// Setting jumpstart to true will attempt to automatically scan for existing
// DNS records. Setting this to false will create the zone with no DNS records.
//
// If Organization is non-empty, it must have at least the ID field populated.
// This will add the new zone to the specified multi-user organization.
//
// API reference: https://api.cloudflare.com/#zone-create-a-zone
func (api *API) CreateZone(name string, jumpstart bool, org Organization) (Zone, error) {
	var newzone newZone
	newzone.Name = name
	newzone.JumpStart = jumpstart
	if org.ID != "" {
		newzone.Organization = &org
	}

	res, err := api.makeRequest("POST", "/zones", newzone)
	if err != nil {
		return Zone{}, errors.Wrap(err, errMakeRequestError)
	}

	var r ZoneResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return Zone{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ZoneActivationCheck initiates another zone activation check for newly-created zones.
//
// API reference: https://api.cloudflare.com/#zone-initiate-another-zone-activation-check
func (api *API) ZoneActivationCheck(zoneID string) (Response, error) {
	res, err := api.makeRequest("PUT", "/zones/"+zoneID+"/activation_check", nil)
	if err != nil {
		return Response{}, errors.Wrap(err, errMakeRequestError)
	}
	var r Response
	err = json.Unmarshal(res, &r)
	if err != nil {
		return Response{}, errors.Wrap(err, errUnmarshalError)
	}
	return r, nil
}

// ListZones lists zones on an account. Optionally takes a list of zone names
// to filter against.
//
// API reference: https://api.cloudflare.com/#zone-list-zones
func (api *API) ListZones(z ...string) ([]Zone, error) {
	v := url.Values{}
	var res []byte
	var r ZonesResponse
	var zones []Zone
	var err error
	if len(z) > 0 {
		for _, zone := range z {
			v.Set("name", zone)
			res, err = api.makeRequest("GET", "/zones?"+v.Encode(), nil)
			if err != nil {
				return []Zone{}, errors.Wrap(err, errMakeRequestError)
			}
			err = json.Unmarshal(res, &r)
			if err != nil {
				return []Zone{}, errors.Wrap(err, errUnmarshalError)
			}
			if !r.Success {
				// TODO: Provide an actual error message instead of always returning nil
				return []Zone{}, err
			}
			for zi := range r.Result {
				zones = append(zones, r.Result[zi])
			}
		}
	} else {
		// TODO: Paginate here. We only grab the first page of results.
		// Could do this concurrently after the first request by creating a
		// sync.WaitGroup or just a channel + workers.
		res, err = api.makeRequest("GET", "/zones", nil)
		if err != nil {
			return []Zone{}, errors.Wrap(err, errMakeRequestError)
		}
		err = json.Unmarshal(res, &r)
		if err != nil {
			return []Zone{}, errors.Wrap(err, errUnmarshalError)
		}
		zones = r.Result
	}

	return zones, nil
}

// ZoneDetails fetches information about a zone.
//
// API reference: https://api.cloudflare.com/#zone-zone-details
func (api *API) ZoneDetails(zoneID string) (Zone, error) {
	res, err := api.makeRequest("GET", "/zones/"+zoneID, nil)
	if err != nil {
		return Zone{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZoneResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return Zone{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ZoneOptions is a subset of Zone, for editable options.
type ZoneOptions struct {
	Paused   *bool         `json:"paused,omitempty"`
	VanityNS []string      `json:"vanity_name_servers,omitempty"`
	Plan     *ZoneRatePlan `json:"plan,omitempty"`
}

// ZoneSetPaused pauses Cloudflare service for the entire zone, sending all
// traffic direct to the origin.
func (api *API) ZoneSetPaused(zoneID string, paused bool) (Zone, error) {
	zoneopts := ZoneOptions{Paused: &paused}
	zone, err := api.EditZone(zoneID, zoneopts)
	if err != nil {
		return Zone{}, err
	}

	return zone, nil
}

// ZoneSetVanityNS sets custom nameservers for the zone.
// These names must be within the same zone.
func (api *API) ZoneSetVanityNS(zoneID string, ns []string) (Zone, error) {
	zoneopts := ZoneOptions{VanityNS: ns}
	zone, err := api.EditZone(zoneID, zoneopts)
	if err != nil {
		return Zone{}, err
	}

	return zone, nil
}

// ZoneSetRatePlan changes the zone plan.
func (api *API) ZoneSetRatePlan(zoneID string, plan ZoneRatePlan) (Zone, error) {
	zoneopts := ZoneOptions{Plan: &plan}
	zone, err := api.EditZone(zoneID, zoneopts)
	if err != nil {
		return Zone{}, err
	}

	return zone, nil
}

// EditZone edits the given zone.
//
// This is usually called by ZoneSetPaused, ZoneSetVanityNS or ZoneSetPlan.
//
// API reference: https://api.cloudflare.com/#zone-edit-zone-properties
func (api *API) EditZone(zoneID string, zoneOpts ZoneOptions) (Zone, error) {
	res, err := api.makeRequest("PATCH", "/zones/"+zoneID, zoneOpts)
	if err != nil {
		return Zone{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZoneResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return Zone{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, nil
}

// PurgeEverything purges the cache for the given zone.
//
// Note: this will substantially increase load on the origin server for that
// zone if there is a high cached vs. uncached request ratio.
//
// API reference: https://api.cloudflare.com/#zone-purge-all-files
func (api *API) PurgeEverything(zoneID string) (PurgeCacheResponse, error) {
	uri := "/zones/" + zoneID + "/purge_cache"
	res, err := api.makeRequest("DELETE", uri, PurgeCacheRequest{true, nil, nil, nil})
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errMakeRequestError)
	}
	var r PurgeCacheResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errUnmarshalError)
	}
	return r, nil
}

// PurgeCache purges the cache using the given PurgeCacheRequest (zone/url/tag).
//
// API reference: https://api.cloudflare.com/#zone-purge-individual-files-by-url-and-cache-tags
func (api *API) PurgeCache(zoneID string, pcr PurgeCacheRequest) (PurgeCacheResponse, error) {
	uri := "/zones/" + zoneID + "/purge_cache"
	res, err := api.makeRequest("DELETE", uri, pcr)
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errMakeRequestError)
	}
	var r PurgeCacheResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errUnmarshalError)
	}
	return r, nil
}

// DeleteZone deletes the given zone.
//
// API reference: https://api.cloudflare.com/#zone-delete-a-zone
func (api *API) DeleteZone(zoneID string) (ZoneID, error) {
	res, err := api.makeRequest("DELETE", "/zones/"+zoneID, nil)
	if err != nil {
		return ZoneID{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZoneIDResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return ZoneID{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// AvailableZoneRatePlans returns information about all plans available to the specified zone.
//
// API reference: https://api.cloudflare.com/#zone-plan-available-plans
func (api *API) AvailableZoneRatePlans(zoneID string) ([]ZoneRatePlan, error) {
	uri := "/zones/" + zoneID + "/available_rate_plans"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []ZoneRatePlan{}, errors.Wrap(err, errMakeRequestError)
	}
	var r AvailableZoneRatePlansResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []ZoneRatePlan{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// encode encodes non-nil fields into URL encoded form.
func (o ZoneAnalyticsOptions) encode() string {
	v := url.Values{}
	if o.Since != nil {
		v.Set("since", (*o.Since).Format(time.RFC3339))
	}
	if o.Until != nil {
		v.Set("until", (*o.Until).Format(time.RFC3339))
	}
	if o.Continuous != nil {
		v.Set("continuous", fmt.Sprintf("%t", *o.Continuous))
	}
	return v.Encode()
}

// ZoneAnalyticsDashboard returns zone analytics information.
//
// API reference: https://api.cloudflare.com/#zone-analytics-dashboard
func (api *API) ZoneAnalyticsDashboard(zoneID string, options ZoneAnalyticsOptions) (ZoneAnalyticsData, error) {
	uri := "/zones/" + zoneID + "/analytics/dashboard" + "?" + options.encode()
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return ZoneAnalyticsData{}, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneAnalyticsDataResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return ZoneAnalyticsData{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ZoneAnalyticsByColocation returns zone analytics information by datacenter.
//
// API reference: https://api.cloudflare.com/#zone-analytics-analytics-by-co-locations
func (api *API) ZoneAnalyticsByColocation(zoneID string, options ZoneAnalyticsOptions) ([]ZoneAnalyticsColocation, error) {
	uri := "/zones/" + zoneID + "/analytics/colos" + "?" + options.encode()
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneAnalyticsColocationResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ZoneSettings returns all of the settings for a given zone.
//
// API reference: https://api.cloudflare.com/#zone-settings-get-all-zone-settings
func (api *API) ZoneSettings(zoneID string) (*ZoneSettingResponse, error) {
	uri := "/zones/" + zoneID + "/settings"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneSettingResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// UpdateZoneSettings updates the settings for a given zone.
//
// API reference: https://api.cloudflare.com/#zone-settings-edit-zone-settings-info
func (api *API) UpdateZoneSettings(zoneID string, settings []ZoneSetting) (*ZoneSettingResponse, error) {
	uri := "/zones/" + zoneID + "/settings"
	res, err := api.makeRequest("PATCH", uri, struct {
		Items []ZoneSetting `json:"items"`
	}{settings})
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneSettingResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// ZoneSSLSettings returns information about SSL setting to the specified zone.
//
// API reference: https://api.cloudflare.com/#zone-settings-get-ssl-setting
func (api *API) ZoneSSLSettings(zoneID string) (ZoneSSLSetting, error) {
	uri := "/zones/" + zoneID + "/settings/ssl"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return ZoneSSLSetting{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZoneSSLSettingResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return ZoneSSLSetting{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}
