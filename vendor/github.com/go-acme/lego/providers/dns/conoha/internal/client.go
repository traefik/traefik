package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	identityBaseURL   = "https://identity.%s.conoha.io"
	dnsServiceBaseURL = "https://dns-service.%s.conoha.io"
)

// IdentityRequest is an authentication request body.
type IdentityRequest struct {
	Auth Auth `json:"auth"`
}

// Auth is an authentication information.
type Auth struct {
	TenantID            string              `json:"tenantId"`
	PasswordCredentials PasswordCredentials `json:"passwordCredentials"`
}

// PasswordCredentials is API-user's credentials.
type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// IdentityResponse is an authentication response body.
type IdentityResponse struct {
	Access Access `json:"access"`
}

// Access is an identity information.
type Access struct {
	Token Token `json:"token"`
}

// Token is an api access token.
type Token struct {
	ID string `json:"id"`
}

// DomainListResponse is a response of a domain listing request.
type DomainListResponse struct {
	Domains []Domain `json:"domains"`
}

// Domain is a hosted domain entry.
type Domain struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecordListResponse is a response of record listing request.
type RecordListResponse struct {
	Records []Record `json:"records"`
}

// Record is a record entry.
type Record struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl"`
}

// Client is a ConoHa API client.
type Client struct {
	token      string
	endpoint   string
	httpClient *http.Client
}

// NewClient returns a client instance logged into the ConoHa service.
func NewClient(region string, auth Auth, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	c := &Client{httpClient: httpClient}

	c.endpoint = fmt.Sprintf(identityBaseURL, region)

	identity, err := c.getIdentity(auth)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}

	c.token = identity.Access.Token.ID
	c.endpoint = fmt.Sprintf(dnsServiceBaseURL, region)

	return c, nil
}

func (c *Client) getIdentity(auth Auth) (*IdentityResponse, error) {
	req := &IdentityRequest{Auth: auth}

	identity := &IdentityResponse{}

	err := c.do(http.MethodPost, "/v2.0/tokens", req, identity)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

// GetDomainID returns an ID of specified domain.
func (c *Client) GetDomainID(domainName string) (string, error) {
	domainList := &DomainListResponse{}

	err := c.do(http.MethodGet, "/v1/domains", nil, domainList)
	if err != nil {
		return "", err
	}

	for _, domain := range domainList.Domains {
		if domain.Name == domainName {
			return domain.ID, nil
		}
	}
	return "", fmt.Errorf("no such domain: %s", domainName)
}

// GetRecordID returns an ID of specified record.
func (c *Client) GetRecordID(domainID, recordName, recordType, data string) (string, error) {
	recordList := &RecordListResponse{}

	err := c.do(http.MethodGet, fmt.Sprintf("/v1/domains/%s/records", domainID), nil, recordList)
	if err != nil {
		return "", err
	}

	for _, record := range recordList.Records {
		if record.Name == recordName && record.Type == recordType && record.Data == data {
			return record.ID, nil
		}
	}
	return "", errors.New("no such record")
}

// CreateRecord adds new record.
func (c *Client) CreateRecord(domainID string, record Record) error {
	return c.do(http.MethodPost, fmt.Sprintf("/v1/domains/%s/records", domainID), record, nil)
}

// DeleteRecord removes specified record.
func (c *Client) DeleteRecord(domainID, recordID string) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/v1/domains/%s/records/%s", domainID, recordID), nil, nil)
}

func (c *Client) do(method, path string, payload, result interface{}) error {
	body := bytes.NewReader(nil)

	if payload != nil {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.endpoint+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		return fmt.Errorf("HTTP request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		return json.Unmarshal(respBody, result)
	}

	return nil
}
