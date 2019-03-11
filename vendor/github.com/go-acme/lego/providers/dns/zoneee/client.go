package zoneee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
)

const defaultEndpoint = "https://api.zone.eu/v2/dns/"

type txtRecord struct {
	// Identifier (identificator)
	ID string `json:"id,omitempty"`
	// Hostname
	Name string `json:"name"`
	// TXT content value
	Destination string `json:"destination"`
	// Can this record be deleted
	Delete bool `json:"delete,omitempty"`
	// Can this record be modified
	Modify bool `json:"modify,omitempty"`
	// API url to get this entity
	ResourceURL string `json:"resource_url,omitempty"`
}

func (d *DNSProvider) addTxtRecord(domain string, record txtRecord) ([]txtRecord, error) {
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(record); err != nil {
		return nil, err
	}

	req, err := d.makeRequest(http.MethodPost, path.Join(domain, "txt"), reqBody)
	if err != nil {
		return nil, err
	}

	var resp []txtRecord
	if err := d.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *DNSProvider) getTxtRecords(domain string) ([]txtRecord, error) {
	req, err := d.makeRequest(http.MethodGet, path.Join(domain, "txt"), nil)
	if err != nil {
		return nil, err
	}

	var resp []txtRecord
	if err := d.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *DNSProvider) removeTxtRecord(domain, id string) error {
	req, err := d.makeRequest(http.MethodDelete, path.Join(domain, "txt", id), nil)
	if err != nil {
		return err
	}

	return d.sendRequest(req, nil)
}

func (d *DNSProvider) makeRequest(method, resource string, body io.Reader) (*http.Request, error) {
	uri, err := d.config.Endpoint.Parse(resource)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.config.Username, d.config.APIKey)

	return req, nil
}

func (d *DNSProvider) sendRequest(req *http.Request, result interface{}) error {
	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if err = checkResponse(resp); err != nil {
		return err
	}

	defer resp.Body.Close()

	if result == nil {
		return nil
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf("unmarshaling %T error [status code=%d]: %v: %s", result, resp.StatusCode, err, string(raw))
	}
	return err
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	if resp.Body == nil {
		return fmt.Errorf("response body is nil, status code=%d", resp.StatusCode)
	}

	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: status code=%d, error=%v", resp.StatusCode, err)
	}

	return fmt.Errorf("status code=%d: %s", resp.StatusCode, string(raw))
}
