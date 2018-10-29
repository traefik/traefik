package cloudflare

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// ZoneCustomSSL represents custom SSL certificate metadata.
type ZoneCustomSSL struct {
	ID            string     `json:"id"`
	Hosts         []string   `json:"hosts"`
	Issuer        string     `json:"issuer"`
	Signature     string     `json:"signature"`
	Status        string     `json:"status"`
	BundleMethod  string     `json:"bundle_method"`
	ZoneID        string     `json:"zone_id"`
	UploadedOn    time.Time  `json:"uploaded_on"`
	ModifiedOn    time.Time  `json:"modified_on"`
	ExpiresOn     time.Time  `json:"expires_on"`
	Priority      int        `json:"priority"`
	KeylessServer KeylessSSL `json:"keyless_server"`
}

// zoneCustomSSLResponse represents the response from the zone SSL details endpoint.
type zoneCustomSSLResponse struct {
	Response
	Result ZoneCustomSSL `json:"result"`
}

// zoneCustomSSLsResponse represents the response from the zone SSL list endpoint.
type zoneCustomSSLsResponse struct {
	Response
	Result []ZoneCustomSSL `json:"result"`
}

// ZoneCustomSSLOptions represents the parameters to create or update an existing
// custom SSL configuration.
type ZoneCustomSSLOptions struct {
	Certificate  string `json:"certificate"`
	PrivateKey   string `json:"private_key"`
	BundleMethod string `json:"bundle_method,omitempty"`
}

// ZoneCustomSSLPriority represents a certificate's ID and priority. It is a
// subset of ZoneCustomSSL used for patch requests.
type ZoneCustomSSLPriority struct {
	ID       string `json:"ID"`
	Priority int    `json:"priority"`
}

// CreateSSL allows you to add a custom SSL certificate to the given zone.
//
// API reference: https://api.cloudflare.com/#custom-ssl-for-a-zone-create-ssl-configuration
func (api *API) CreateSSL(zoneID string, options ZoneCustomSSLOptions) (ZoneCustomSSL, error) {
	uri := "/zones/" + zoneID + "/custom_certificates"
	res, err := api.makeRequest("POST", uri, options)
	if err != nil {
		return ZoneCustomSSL{}, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneCustomSSLResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return ZoneCustomSSL{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ListSSL lists the custom certificates for the given zone.
//
// API reference: https://api.cloudflare.com/#custom-ssl-for-a-zone-list-ssl-configurations
func (api *API) ListSSL(zoneID string) ([]ZoneCustomSSL, error) {
	uri := "/zones/" + zoneID + "/custom_certificates"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneCustomSSLsResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// SSLDetails returns the configuration details for a custom SSL certificate.
//
// API reference: https://api.cloudflare.com/#custom-ssl-for-a-zone-ssl-configuration-details
func (api *API) SSLDetails(zoneID, certificateID string) (ZoneCustomSSL, error) {
	uri := "/zones/" + zoneID + "/custom_certificates/" + certificateID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return ZoneCustomSSL{}, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneCustomSSLResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return ZoneCustomSSL{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// UpdateSSL updates (replaces) a custom SSL certificate.
//
// API reference: https://api.cloudflare.com/#custom-ssl-for-a-zone-update-ssl-configuration
func (api *API) UpdateSSL(zoneID, certificateID string, options ZoneCustomSSLOptions) (ZoneCustomSSL, error) {
	uri := "/zones/" + zoneID + "/custom_certificates/" + certificateID
	res, err := api.makeRequest("PATCH", uri, options)
	if err != nil {
		return ZoneCustomSSL{}, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneCustomSSLResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return ZoneCustomSSL{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ReprioritizeSSL allows you to change the priority (which is served for a given
// request) of custom SSL certificates associated with the given zone.
//
// API reference: https://api.cloudflare.com/#custom-ssl-for-a-zone-re-prioritize-ssl-certificates
func (api *API) ReprioritizeSSL(zoneID string, p []ZoneCustomSSLPriority) ([]ZoneCustomSSL, error) {
	uri := "/zones/" + zoneID + "/custom_certificates/prioritize"
	params := struct {
		Certificates []ZoneCustomSSLPriority `json:"certificates"`
	}{
		Certificates: p,
	}
	res, err := api.makeRequest("PUT", uri, params)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r zoneCustomSSLsResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// DeleteSSL deletes a custom SSL certificate from the given zone.
//
// API reference: https://api.cloudflare.com/#custom-ssl-for-a-zone-delete-an-ssl-certificate
func (api *API) DeleteSSL(zoneID, certificateID string) error {
	uri := "/zones/" + zoneID + "/custom_certificates/" + certificateID
	if _, err := api.makeRequest("DELETE", uri, nil); err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	return nil
}
