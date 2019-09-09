package versio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
)

const defaultBaseURL = "https://www.versio.nl/api/v1/"

type dnsRecordsResponse struct {
	Record dnsRecord `json:"domainInfo"`
}

type dnsRecord struct {
	DNSRecords []record `json:"dns_records"`
}

type record struct {
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Priority int    `json:"prio,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

type dnsErrorResponse struct {
	Error errorMessage `json:"error"`
}

type errorMessage struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (d *DNSProvider) postDNSRecords(domain string, msg interface{}) error {
	reqBody := &bytes.Buffer{}
	err := json.NewEncoder(reqBody).Encode(msg)
	if err != nil {
		return err
	}

	req, err := d.makeRequest(http.MethodPost, "domains/"+domain+"/update", reqBody)
	if err != nil {
		return err
	}

	return d.do(req, nil)
}

func (d *DNSProvider) getDNSRecords(domain string) (*dnsRecordsResponse, error) {
	req, err := d.makeRequest(http.MethodGet, "domains/"+domain+"?show_dns_records=true", nil)
	if err != nil {
		return nil, err
	}

	// we'll need all the dns_records to add the new TXT record
	respData := &dnsRecordsResponse{}
	err = d.do(req, respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

func (d *DNSProvider) makeRequest(method string, uri string, body io.Reader) (*http.Request, error) {
	endpoint, err := d.config.BaseURL.Parse(path.Join(d.config.BaseURL.EscapedPath(), uri))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(d.config.Username) > 0 && len(d.config.Password) > 0 {
		req.SetBasicAuth(d.config.Username, d.config.Password)
	}

	return req, nil
}

func (d *DNSProvider) do(req *http.Request, result interface{}) error {
	resp, err := d.config.HTTPClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%d: failed to read response body: %v", resp.StatusCode, err)
		}

		respError := &dnsErrorResponse{}
		err = json.Unmarshal(body, respError)
		if err != nil {
			return fmt.Errorf("%d: request failed: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("%d: request failed: %s", resp.StatusCode, respError.Error.Message)
	}

	if result != nil {
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("request failed: %v", err)
		}

		if err = json.Unmarshal(content, result); err != nil {
			return fmt.Errorf("%v: %s", err, content)
		}
	}

	return nil
}
