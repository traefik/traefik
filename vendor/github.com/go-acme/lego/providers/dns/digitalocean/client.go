package digitalocean

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-acme/lego/challenge/dns01"
)

const defaultBaseURL = "https://api.digitalocean.com"

// txtRecordResponse represents a response from DO's API after making a TXT record
type txtRecordResponse struct {
	DomainRecord record `json:"domain_record"`
}

type record struct {
	ID   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Data string `json:"data,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
}

type apiError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (d *DNSProvider) removeTxtRecord(domain string, recordID int) error {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return fmt.Errorf("could not determine zone for domain: '%s'. %s", domain, err)
	}

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records/%d", d.config.BaseURL, dns01.UnFqdn(authZone), recordID)
	req, err := d.newRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return readError(req, resp)
	}

	return nil
}

func (d *DNSProvider) addTxtRecord(fqdn, value string) (*txtRecordResponse, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(fqdn))
	if err != nil {
		return nil, fmt.Errorf("could not determine zone for domain: '%s'. %s", fqdn, err)
	}

	reqData := record{Type: "TXT", Name: fqdn, Data: value, TTL: d.config.TTL}
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records", d.config.BaseURL, dns01.UnFqdn(authZone))
	req, err := d.newRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, readError(req, resp)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	respData := &txtRecordResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	return respData, nil
}

func (d *DNSProvider) newRequest(method, reqURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.config.AuthToken))

	return req, nil
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo apiError
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("apiError unmarshaling error: %v: %s", err, toUnreadableBodyMessage(req, content))
	}

	return fmt.Errorf("HTTP %d: %s: %s", resp.StatusCode, errInfo.ID, errInfo.Message)
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
