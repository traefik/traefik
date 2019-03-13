package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Domain represents domain name.
type Domain struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Record represents DNS record.
type Record struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"` // Record type (SOA, NS, A/AAAA, CNAME, SRV, MX, TXT, SPF)
	TTL     int    `json:"ttl,omitempty"`
	Email   string `json:"email,omitempty"`   // Email of domain's admin (only for SOA records)
	Content string `json:"content,omitempty"` // Record content (not for SRV)
}

// APIError API error message
type APIError struct {
	Description string `json:"error"`
	Code        int    `json:"code"`
	Field       string `json:"field"`
}

func (a *APIError) Error() string {
	return fmt.Sprintf("API error: %d - %s - %s", a.Code, a.Description, a.Field)
}

// ClientOpts represents options to init client.
type ClientOpts struct {
	BaseURL    string
	Token      string
	UserAgent  string
	HTTPClient *http.Client
}

// Client represents DNS client.
type Client struct {
	baseURL    string
	token      string
	userAgent  string
	httpClient *http.Client
}

// NewClient returns a client instance.
func NewClient(opts ClientOpts) *Client {
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{}
	}

	return &Client{
		token:      opts.Token,
		baseURL:    opts.BaseURL,
		httpClient: opts.HTTPClient,
		userAgent:  opts.UserAgent,
	}
}

// GetDomainByName gets Domain object by its name. If `domainName` level > 2 and there is
// no such domain on the account - it'll recursively search for the first
// which is exists in Vscale Domains API.
func (c *Client) GetDomainByName(domainName string) (*Domain, error) {
	uri := fmt.Sprintf("/%s", domainName)
	req, err := c.newRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	domain := &Domain{}
	resp, err := c.do(req, domain)
	if err != nil {
		switch {
		case resp.StatusCode == http.StatusNotFound && strings.Count(domainName, ".") > 1:
			// Look up for the next sub domain
			subIndex := strings.Index(domainName, ".")
			return c.GetDomainByName(domainName[subIndex+1:])
		default:
			return nil, err
		}
	}

	return domain, nil
}

// AddRecord adds Record for given domain.
func (c *Client) AddRecord(domainID int, body Record) (*Record, error) {
	uri := fmt.Sprintf("/%d/records/", domainID)
	req, err := c.newRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	record := &Record{}
	_, err = c.do(req, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ListRecords returns list records for specific domain.
func (c *Client) ListRecords(domainID int) ([]*Record, error) {
	uri := fmt.Sprintf("/%d/records/", domainID)
	req, err := c.newRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	var records []*Record
	_, err = c.do(req, &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// DeleteRecord deletes specific record.
func (c *Client) DeleteRecord(domainID, recordID int) error {
	uri := fmt.Sprintf("/%d/records/%d", domainID, recordID)
	req, err := c.newRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

func (c *Client) newRequest(method, uri string, body interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body with error: %v", err)
		}
	}

	req, err := http.NewRequest(method, c.baseURL+uri, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create new http request with error: %v", err)
	}

	req.Header.Add("X-Token", c.token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, to interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed with error: %v", err)
	}

	err = checkResponse(resp)
	if err != nil {
		return resp, err
	}

	if to != nil {
		if err = unmarshalBody(resp, to); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= http.StatusBadRequest &&
		resp.StatusCode <= http.StatusNetworkAuthenticationRequired {

		if resp.Body == nil {
			return fmt.Errorf("request failed with status code %d and empty body", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		apiError := APIError{}
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return fmt.Errorf("request failed with status code %d, response body: %s", resp.StatusCode, string(body))
		}

		return fmt.Errorf("request failed with status code %d: %v", resp.StatusCode, apiError)
	}

	return nil
}

func unmarshalBody(resp *http.Response, to interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, to)
	if err != nil {
		return fmt.Errorf("unmarshaling error: %v: %s", err, string(body))
	}

	return nil
}
