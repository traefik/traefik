package internal

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
)

const defaultBaseURL = "https://www.cloudxns.net/api2/"

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Data Domain information
type Data struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
	TTL    int    `json:"ttl,omitempty"`
}

// TXTRecord a TXT record
type TXTRecord struct {
	ID       int    `json:"domain_id,omitempty"`
	RecordID string `json:"record_id,omitempty"`

	Host   string `json:"host"`
	Value  string `json:"value"`
	Type   string `json:"type"`
	LineID int    `json:"line_id,string"`
	TTL    int    `json:"ttl,string"`
}

// NewClient creates a CloudXNS client
func NewClient(apiKey string, secretKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("CloudXNS: credentials missing: apiKey")
	}

	if secretKey == "" {
		return nil, fmt.Errorf("CloudXNS: credentials missing: secretKey")
	}

	return &Client{
		apiKey:     apiKey,
		secretKey:  secretKey,
		HTTPClient: &http.Client{},
		BaseURL:    defaultBaseURL,
	}, nil
}

// Client CloudXNS client
type Client struct {
	apiKey     string
	secretKey  string
	HTTPClient *http.Client
	BaseURL    string
}

// GetDomainInformation Get domain name information for a FQDN
func (c *Client) GetDomainInformation(fqdn string) (*Data, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, err
	}

	result, err := c.doRequest(http.MethodGet, "domain", nil)
	if err != nil {
		return nil, err
	}

	var domains []Data
	if len(result) > 0 {
		err = json.Unmarshal(result, &domains)
		if err != nil {
			return nil, fmt.Errorf("CloudXNS: domains unmarshaling error: %v", err)
		}
	}

	for _, data := range domains {
		if data.Domain == authZone {
			return &data, nil
		}
	}

	return nil, fmt.Errorf("CloudXNS: zone %s not found for domain %s", authZone, fqdn)
}

// FindTxtRecord return the TXT record a zone ID and a FQDN
func (c *Client) FindTxtRecord(zoneID, fqdn string) (*TXTRecord, error) {
	result, err := c.doRequest(http.MethodGet, fmt.Sprintf("record/%s?host_id=0&offset=0&row_num=2000", zoneID), nil)
	if err != nil {
		return nil, err
	}

	var records []TXTRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: TXT record unmarshaling error: %v", err)
	}

	for _, record := range records {
		if record.Host == dns01.UnFqdn(fqdn) && record.Type == "TXT" {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("CloudXNS: no existing record found for %q", fqdn)
}

// AddTxtRecord add a TXT record
func (c *Client) AddTxtRecord(info *Data, fqdn, value string, ttl int) error {
	id, err := strconv.Atoi(info.ID)
	if err != nil {
		return fmt.Errorf("CloudXNS: invalid zone ID: %v", err)
	}

	payload := TXTRecord{
		ID:     id,
		Host:   dns01.UnFqdn(strings.TrimSuffix(fqdn, info.Domain)),
		Value:  value,
		Type:   "TXT",
		LineID: 1,
		TTL:    ttl,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("CloudXNS: record unmarshaling error: %v", err)
	}

	_, err = c.doRequest(http.MethodPost, "record", body)
	return err
}

// RemoveTxtRecord remove a TXT record
func (c *Client) RemoveTxtRecord(recordID, zoneID string) error {
	_, err := c.doRequest(http.MethodDelete, fmt.Sprintf("record/%s/%s", recordID, zoneID), nil)
	return err
}

func (c *Client) doRequest(method, uri string, body []byte) (json.RawMessage, error) {
	req, err := c.buildRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: %v", err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: %s", toUnreadableBodyMessage(req, content))
	}

	var r apiResponse
	err = json.Unmarshal(content, &r)
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: response unmashaling error: %v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if r.Code != 1 {
		return nil, fmt.Errorf("CloudXNS: invalid code (%v), error: %s", r.Code, r.Message)
	}
	return r.Data, nil
}

func (c *Client) buildRequest(method, uri string, body []byte) (*http.Request, error) {
	url := c.BaseURL + uri

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("CloudXNS: invalid request: %v", err)
	}

	requestDate := time.Now().Format(time.RFC1123Z)

	req.Header.Set("API-KEY", c.apiKey)
	req.Header.Set("API-REQUEST-DATE", requestDate)
	req.Header.Set("API-HMAC", c.hmac(url, requestDate, string(body)))
	req.Header.Set("API-FORMAT", "json")

	return req, nil
}

func (c *Client) hmac(url, date, body string) string {
	sum := md5.Sum([]byte(c.apiKey + url + body + date + c.secretKey))
	return hex.EncodeToString(sum[:])
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
