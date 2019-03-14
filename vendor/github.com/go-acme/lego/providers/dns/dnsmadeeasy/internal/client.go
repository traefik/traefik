package internal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Domain holds the DNSMadeEasy API representation of a Domain
type Domain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Record holds the DNSMadeEasy API representation of a Domain Record
type Record struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	SourceID int    `json:"sourceId"`
}

type recordsResponse struct {
	Records *[]Record `json:"data"`
}

// Client DNSMadeEasy client
type Client struct {
	apiKey     string
	apiSecret  string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a DNSMadeEasy client
func NewClient(apiKey string, apiSecret string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("credentials missing: API key")
	}

	if apiSecret == "" {
		return nil, fmt.Errorf("credentials missing: API secret")
	}

	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		HTTPClient: &http.Client{},
	}, nil
}

// GetDomain gets a domain
func (c *Client) GetDomain(authZone string) (*Domain, error) {
	domainName := authZone[0 : len(authZone)-1]
	resource := fmt.Sprintf("%s%s", "/dns/managed/name?domainname=", domainName)

	resp, err := c.sendRequest(http.MethodGet, resource, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	domain := &Domain{}
	err = json.NewDecoder(resp.Body).Decode(&domain)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

// GetRecords gets all TXT records
func (c *Client) GetRecords(domain *Domain, recordName, recordType string) (*[]Record, error) {
	resource := fmt.Sprintf("%s/%d/%s%s%s%s", "/dns/managed", domain.ID, "records?recordName=", recordName, "&type=", recordType)

	resp, err := c.sendRequest(http.MethodGet, resource, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	records := &recordsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}

	return records.Records, nil
}

// CreateRecord creates a TXT records
func (c *Client) CreateRecord(domain *Domain, record *Record) error {
	url := fmt.Sprintf("%s/%d/%s", "/dns/managed", domain.ID, "records")

	resp, err := c.sendRequest(http.MethodPost, url, record)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteRecord deletes a TXT records
func (c *Client) DeleteRecord(record Record) error {
	resource := fmt.Sprintf("%s/%d/%s/%d", "/dns/managed", record.SourceID, "records", record.ID)

	resp, err := c.sendRequest(http.MethodDelete, resource, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) sendRequest(method, resource string, payload interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UTC().Format(time.RFC1123)
	signature, err := computeHMAC(timestamp, c.apiSecret)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-dnsme-apiKey", c.apiKey)
	req.Header.Set("x-dnsme-requestDate", timestamp)
	req.Header.Set("x-dnsme-hmac", signature)
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("request failed with HTTP status code %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func computeHMAC(message string, secret string) (string, error) {
	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
