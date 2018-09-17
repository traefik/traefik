package cloudflare

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/xenolf/lego/acme"
)

// defaultBaseURL represents the API endpoint to call.
const defaultBaseURL = "https://api.cloudflare.com/client/v4"

// APIError contains error details for failed requests
type APIError struct {
	Code       int        `json:"code,omitempty"`
	Message    string     `json:"message,omitempty"`
	ErrorChain []APIError `json:"error_chain,omitempty"`
}

// APIResponse represents a response from Cloudflare API
type APIResponse struct {
	Success bool            `json:"success"`
	Errors  []*APIError     `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

// TxtRecord represents a Cloudflare DNS record
type TxtRecord struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	ZoneID  string `json:"zone_id,omitempty"`
}

// HostedZone represents a Cloudflare DNS zone
type HostedZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Client Cloudflare API client
type Client struct {
	authEmail  string
	authKey    string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient create a Cloudflare API client
func NewClient(authEmail string, authKey string) (*Client, error) {
	if authEmail == "" {
		return nil, errors.New("cloudflare: some credentials information are missing: email")
	}

	if authKey == "" {
		return nil, errors.New("cloudflare: some credentials information are missing: key")
	}

	return &Client{
		authEmail:  authEmail,
		authKey:    authKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
	}, nil
}

// GetHostedZoneID get hosted zone
func (c *Client) GetHostedZoneID(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	result, err := c.doRequest(http.MethodGet, "/zones?name="+acme.UnFqdn(authZone), nil)
	if err != nil {
		return "", err
	}

	var hostedZone []HostedZone
	err = json.Unmarshal(result, &hostedZone)
	if err != nil {
		return "", fmt.Errorf("cloudflare: HostedZone unmarshaling error: %v", err)
	}

	count := len(hostedZone)
	if count == 0 {
		return "", fmt.Errorf("cloudflare: zone %s not found for domain %s", authZone, fqdn)
	} else if count > 1 {
		return "", fmt.Errorf("cloudflare: zone %s cannot be find for domain %s: too many hostedZone: %v", authZone, fqdn, hostedZone)
	}

	return hostedZone[0].ID, nil
}

// FindTxtRecord Find a TXT record
func (c *Client) FindTxtRecord(zoneID, fqdn string) (*TxtRecord, error) {
	result, err := c.doRequest(
		http.MethodGet,
		fmt.Sprintf("/zones/%s/dns_records?per_page=1000&type=TXT&name=%s", zoneID, acme.UnFqdn(fqdn)),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var records []TxtRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: record unmarshaling error: %v", err)
	}

	for _, rec := range records {
		fmt.Println(rec.Name, acme.UnFqdn(fqdn))
		if rec.Name == acme.UnFqdn(fqdn) {
			return &rec, nil
		}
	}

	return nil, fmt.Errorf("cloudflare: no existing record found for %s", fqdn)
}

// AddTxtRecord add a TXT record
func (c *Client) AddTxtRecord(fqdn string, record TxtRecord) error {
	zoneID, err := c.GetHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("cloudflare: record marshaling error: %v", err)
	}

	_, err = c.doRequest(http.MethodPost, fmt.Sprintf("/zones/%s/dns_records", zoneID), bytes.NewReader(body))
	return err
}

// RemoveTxtRecord Remove a TXT record
func (c *Client) RemoveTxtRecord(fqdn string) error {
	zoneID, err := c.GetHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	record, err := c.FindTxtRecord(zoneID, fqdn)
	if err != nil {
		return err
	}

	_, err = c.doRequest(http.MethodDelete, fmt.Sprintf("/zones/%s/dns_records/%s", record.ZoneID, record.ID), nil)
	return err
}

func (c *Client) doRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Email", c.authEmail)
	req.Header.Set("X-Auth-Key", c.authKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: error querying API: %v", err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: %s", toUnreadableBodyMessage(req, content))
	}

	var r APIResponse
	err = json.Unmarshal(content, &r)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: APIResponse unmarshaling error: %v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if !r.Success {
		if len(r.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare: error \n%s", toError(r))
		}

		return nil, fmt.Errorf("cloudflare: %s", toUnreadableBodyMessage(req, content))
	}

	return r.Result, nil
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}

func toError(r APIResponse) error {
	errStr := ""
	for _, apiErr := range r.Errors {
		errStr += fmt.Sprintf("\t Error: %d: %s", apiErr.Code, apiErr.Message)
		for _, chainErr := range apiErr.ErrorChain {
			errStr += fmt.Sprintf("<- %d: %s", chainErr.Code, chainErr.Message)
		}
	}
	return fmt.Errorf("cloudflare: error \n%s", errStr)
}
