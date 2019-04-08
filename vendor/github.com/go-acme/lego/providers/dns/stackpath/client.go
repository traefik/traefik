package stackpath

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/go-acme/lego/challenge/dns01"
	"golang.org/x/net/publicsuffix"
)

// Zones is the response struct from the Stackpath api GetZones
type Zones struct {
	Zones []Zone `json:"zones"`
}

// Zone a DNS zone representation
type Zone struct {
	ID     string
	Domain string
}

// Records is the response struct from the Stackpath api GetZoneRecords
type Records struct {
	Records []Record `json:"records"`
}

// Record a DNS record representation
type Record struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl"`
	Data string `json:"data"`
}

// ErrorResponse the API error response representation
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

// https://developer.stackpath.com/en/api/dns/#operation/GetZones
func (d *DNSProvider) getZones(domain string) (*Zone, error) {
	domain = dns01.UnFqdn(domain)
	tld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return nil, err
	}

	req, err := d.newRequest(http.MethodGet, "/zones", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("page_request.filter", fmt.Sprintf("domain='%s'", tld))
	req.URL.RawQuery = query.Encode()

	var zones Zones
	err = d.do(req, &zones)
	if err != nil {
		return nil, err
	}

	if len(zones.Zones) == 0 {
		return nil, fmt.Errorf("did not find zone with domain %s", domain)
	}

	return &zones.Zones[0], nil
}

// https://developer.stackpath.com/en/api/dns/#operation/GetZoneRecords
func (d *DNSProvider) getZoneRecords(name string, zone *Zone) ([]Record, error) {
	u := fmt.Sprintf("/zones/%s/records", zone.ID)
	req, err := d.newRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("page_request.filter", fmt.Sprintf("name='%s' and type='TXT'", name))
	req.URL.RawQuery = query.Encode()

	var records Records
	err = d.do(req, &records)
	if err != nil {
		return nil, err
	}

	if len(records.Records) == 0 {
		return nil, fmt.Errorf("did not find record with name %s", name)
	}

	return records.Records, nil
}

// https://developer.stackpath.com/en/api/dns/#operation/CreateZoneRecord
func (d *DNSProvider) createZoneRecord(zone *Zone, record Record) error {
	u := fmt.Sprintf("/zones/%s/records", zone.ID)
	req, err := d.newRequest(http.MethodPost, u, record)
	if err != nil {
		return err
	}

	return d.do(req, nil)
}

// https://developer.stackpath.com/en/api/dns/#operation/DeleteZoneRecord
func (d *DNSProvider) deleteZoneRecord(zone *Zone, record Record) error {
	u := fmt.Sprintf("/zones/%s/records/%s", zone.ID, record.ID)
	req, err := d.newRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	return d.do(req, nil)
}

func (d *DNSProvider) newRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := d.BaseURL.Parse(path.Join(d.config.StackID, urlStr))
	if err != nil {
		return nil, err
	}

	if body == nil {
		var req *http.Request
		req, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return nil, err
		}

		return req, nil
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (d *DNSProvider) do(req *http.Request, v interface{}) error {
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}

	err = checkResponse(resp)
	if err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	raw, err := readBody(resp)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	err = json.Unmarshal(raw, v)
	if err != nil {
		return fmt.Errorf("unmarshaling error: %v: %s", err, string(raw))
	}

	return nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode > 299 {
		data, err := readBody(resp)
		if err != nil {
			return &ErrorResponse{Code: resp.StatusCode, Message: err.Error()}
		}

		errResp := &ErrorResponse{}
		err = json.Unmarshal(data, errResp)
		if err != nil {
			return &ErrorResponse{Code: resp.StatusCode, Message: fmt.Sprintf("unmarshaling error: %v: %s", err, string(data))}
		}
		return errResp
	}

	return nil
}

func readBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	defer resp.Body.Close()

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return rawBody, nil
}
