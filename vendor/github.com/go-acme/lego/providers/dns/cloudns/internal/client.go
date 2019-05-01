package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"
)

const defaultBaseURL = "https://api.cloudns.net/dns/"

type apiResponse struct {
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
}

type Zone struct {
	Name   string
	Type   string
	Zone   string
	Status string // is an integer, but cast as string
}

// TXTRecord a TXT record
type TXTRecord struct {
	ID       int    `json:"id,string"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Record   string `json:"record"`
	Failover int    `json:"failover,string"`
	TTL      int    `json:"ttl,string"`
	Status   int    `json:"status"`
}

type TXTRecords map[string]TXTRecord

// NewClient creates a ClouDNS client
func NewClient(authID string, authPassword string) (*Client, error) {
	if authID == "" {
		return nil, fmt.Errorf("credentials missing: authID")
	}

	if authPassword == "" {
		return nil, fmt.Errorf("credentials missing: authPassword")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		authID:       authID,
		authPassword: authPassword,
		HTTPClient:   &http.Client{},
		BaseURL:      baseURL,
	}, nil
}

// Client ClouDNS client
type Client struct {
	authID       string
	authPassword string
	HTTPClient   *http.Client
	BaseURL      *url.URL
}

// GetZone Get domain name information for a FQDN
func (c *Client) GetZone(authFQDN string) (*Zone, error) {
	authZone, err := dns01.FindZoneByFqdn(authFQDN)
	if err != nil {
		return nil, err
	}

	authZoneName := dns01.UnFqdn(authZone)

	reqURL := *c.BaseURL
	reqURL.Path += "get-zone-info.json"

	q := reqURL.Query()
	q.Add("domain-name", authZoneName)
	reqURL.RawQuery = q.Encode()

	result, err := c.doRequest(http.MethodGet, &reqURL)
	if err != nil {
		return nil, err
	}

	var zone Zone

	if len(result) > 0 {
		if err = json.Unmarshal(result, &zone); err != nil {
			return nil, fmt.Errorf("zone unmarshaling error: %v", err)
		}
	}

	if zone.Name == authZoneName {
		return &zone, nil
	}

	return nil, fmt.Errorf("zone %s not found for authFQDN %s", authZoneName, authFQDN)
}

// FindTxtRecord return the TXT record a zone ID and a FQDN
func (c *Client) FindTxtRecord(zoneName, fqdn string) (*TXTRecord, error) {
	host := dns01.UnFqdn(strings.TrimSuffix(dns01.UnFqdn(fqdn), zoneName))

	reqURL := *c.BaseURL
	reqURL.Path += "records.json"

	q := reqURL.Query()
	q.Add("domain-name", zoneName)
	q.Add("host", host)
	q.Add("type", "TXT")
	reqURL.RawQuery = q.Encode()

	result, err := c.doRequest(http.MethodGet, &reqURL)
	if err != nil {
		return nil, err
	}

	// the API returns [] when there is no records.
	if string(result) == "[]" {
		return nil, nil
	}

	var records TXTRecords
	if err = json.Unmarshal(result, &records); err != nil {
		return nil, fmt.Errorf("TXT record unmarshaling error: %v: %s", err, string(result))
	}

	for _, record := range records {
		if record.Host == host && record.Type == "TXT" {
			return &record, nil
		}
	}

	return nil, nil
}

// AddTxtRecord add a TXT record
func (c *Client) AddTxtRecord(zoneName string, fqdn, value string, ttl int) error {
	host := dns01.UnFqdn(strings.TrimSuffix(dns01.UnFqdn(fqdn), zoneName))

	reqURL := *c.BaseURL
	reqURL.Path += "add-record.json"

	q := reqURL.Query()
	q.Add("domain-name", zoneName)
	q.Add("host", host)
	q.Add("record", value)
	q.Add("ttl", strconv.Itoa(ttlRounder(ttl)))
	q.Add("record-type", "TXT")
	reqURL.RawQuery = q.Encode()

	raw, err := c.doRequest(http.MethodPost, &reqURL)
	if err != nil {
		return err
	}

	resp := apiResponse{}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("apiResponse unmarshaling error: %v: %s", err, string(raw))
	}

	if resp.Status != "Success" {
		return fmt.Errorf("fail to add TXT record: %s %s", resp.Status, resp.StatusDescription)
	}

	return nil
}

// RemoveTxtRecord remove a TXT record
func (c *Client) RemoveTxtRecord(recordID int, zoneName string) error {
	reqURL := *c.BaseURL
	reqURL.Path += "delete-record.json"

	q := reqURL.Query()
	q.Add("domain-name", zoneName)
	q.Add("record-id", strconv.Itoa(recordID))
	reqURL.RawQuery = q.Encode()

	raw, err := c.doRequest(http.MethodPost, &reqURL)
	if err != nil {
		return err
	}

	resp := apiResponse{}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("apiResponse unmarshaling error: %v: %s", err, string(raw))
	}

	if resp.Status != "Success" {
		return fmt.Errorf("fail to add TXT record: %s %s", resp.Status, resp.StatusDescription)
	}

	return nil
}

func (c *Client) doRequest(method string, url *url.URL) (json.RawMessage, error) {
	req, err := c.buildRequest(method, url)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid code (%v), error: %s", resp.StatusCode, content)
	}
	return content, nil
}

func (c *Client) buildRequest(method string, url *url.URL) (*http.Request, error) {
	q := url.Query()
	q.Add("auth-id", c.authID)
	q.Add("auth-password", c.authPassword)
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %v", err)
	}

	return req, nil
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}

// https://www.cloudns.net/wiki/article/58/
// Available TTL's:
// 60 = 1 minute
// 300 = 5 minutes
// 900 = 15 minutes
// 1800 = 30 minutes
// 3600 = 1 hour
// 21600 = 6 hours
// 43200 = 12 hours
// 86400 = 1 day
// 172800 = 2 days
// 259200 = 3 days
// 604800 = 1 week
// 1209600 = 2 weeks
// 2592000 = 1 month
func ttlRounder(ttl int) int {
	for _, validTTL := range []int{60, 300, 900, 1800, 3600, 21600, 43200, 86400, 172800, 259200, 604800, 1209600} {
		if ttl <= validTTL {
			return validTTL
		}
	}

	return 2592000
}
