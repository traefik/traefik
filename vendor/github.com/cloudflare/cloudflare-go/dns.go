package cloudflare

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// DNSRecord represents a DNS record in a zone.
type DNSRecord struct {
	ID         string      `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Content    string      `json:"content,omitempty"`
	Proxiable  bool        `json:"proxiable,omitempty"`
	Proxied    bool        `json:"proxied,omitempty"`
	TTL        int         `json:"ttl,omitempty"`
	Locked     bool        `json:"locked,omitempty"`
	ZoneID     string      `json:"zone_id,omitempty"`
	ZoneName   string      `json:"zone_name,omitempty"`
	CreatedOn  time.Time   `json:"created_on,omitempty"`
	ModifiedOn time.Time   `json:"modified_on,omitempty"`
	Data       interface{} `json:"data,omitempty"` // data returned by: SRV, LOC
	Meta       interface{} `json:"meta,omitempty"`
	Priority   int         `json:"priority,omitempty"`
}

// DNSRecordResponse represents the response from the DNS endpoint.
type DNSRecordResponse struct {
	Result DNSRecord `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// DNSListResponse represents the response from the list DNS records endpoint.
type DNSListResponse struct {
	Result []DNSRecord `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// CreateDNSRecord creates a DNS record for the zone identifier.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-create-dns-record
func (api *API) CreateDNSRecord(zoneID string, rr DNSRecord) (*DNSRecordResponse, error) {
	uri := "/zones/" + zoneID + "/dns_records"
	res, err := api.makeRequest("POST", uri, rr)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	var recordResp *DNSRecordResponse
	err = json.Unmarshal(res, &recordResp)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return recordResp, nil
}

// DNSRecords returns a slice of DNS records for the given zone identifier.
//
// This takes a DNSRecord to allow filtering of the results returned.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records
func (api *API) DNSRecords(zoneID string, rr DNSRecord) ([]DNSRecord, error) {
	// Construct a query string
	v := url.Values{}
	// Request as many records as possible per page - API max is 50
	v.Set("per_page", "50")
	if rr.Name != "" {
		v.Set("name", rr.Name)
	}
	if rr.Type != "" {
		v.Set("type", rr.Type)
	}
	if rr.Content != "" {
		v.Set("content", rr.Content)
	}

	var query string
	var records []DNSRecord
	page := 1

	// Loop over makeRequest until what we've fetched all records
	for {
		v.Set("page", strconv.Itoa(page))
		query = "?" + v.Encode()
		uri := "/zones/" + zoneID + "/dns_records" + query
		res, err := api.makeRequest("GET", uri, nil)
		if err != nil {
			return []DNSRecord{}, errors.Wrap(err, errMakeRequestError)
		}
		var r DNSListResponse
		err = json.Unmarshal(res, &r)
		if err != nil {
			return []DNSRecord{}, errors.Wrap(err, errUnmarshalError)
		}
		records = append(records, r.Result...)
		if r.ResultInfo.Page >= r.ResultInfo.TotalPages {
			break
		}
		// Loop around and fetch the next page
		page++
	}
	return records, nil
}

// DNSRecord returns a single DNS record for the given zone & record
// identifiers.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-dns-record-details
func (api *API) DNSRecord(zoneID, recordID string) (DNSRecord, error) {
	uri := "/zones/" + zoneID + "/dns_records/" + recordID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return DNSRecord{}, errors.Wrap(err, errMakeRequestError)
	}
	var r DNSRecordResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return DNSRecord{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// UpdateDNSRecord updates a single DNS record for the given zone & record
// identifiers.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-update-dns-record
func (api *API) UpdateDNSRecord(zoneID, recordID string, rr DNSRecord) error {
	rec, err := api.DNSRecord(zoneID, recordID)
	if err != nil {
		return err
	}
	// Populate the record name from the existing one if the update didn't
	// specify it.
	if rr.Name == "" {
		rr.Name = rec.Name
	}
	rr.Type = rec.Type
	uri := "/zones/" + zoneID + "/dns_records/" + recordID
	res, err := api.makeRequest("PUT", uri, rr)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	var r DNSRecordResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}
	return nil
}

// DeleteDNSRecord deletes a single DNS record for the given zone & record
// identifiers.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-delete-dns-record
func (api *API) DeleteDNSRecord(zoneID, recordID string) error {
	uri := "/zones/" + zoneID + "/dns_records/" + recordID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	var r DNSRecordResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}
	return nil
}
