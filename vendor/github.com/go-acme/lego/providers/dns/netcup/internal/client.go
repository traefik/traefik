package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// defaultBaseURL for reaching the jSON-based API-Endpoint of netcup
const defaultBaseURL = "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON"

// success response status
const success = "success"

// Request wrapper as specified in netcup wiki
// needed for every request to netcup API around *Msg
// https://www.netcup-wiki.de/wiki/CCP_API#Anmerkungen_zu_JSON-Requests
type Request struct {
	Action string      `json:"action"`
	Param  interface{} `json:"param"`
}

// LoginRequest as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#login
type LoginRequest struct {
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APIPassword     string `json:"apipassword"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// LogoutRequest as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#logout
type LogoutRequest struct {
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APISessionID    string `json:"apisessionid"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// UpdateDNSRecordsRequest as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#updateDnsRecords
type UpdateDNSRecordsRequest struct {
	DomainName      string       `json:"domainname"`
	CustomerNumber  string       `json:"customernumber"`
	APIKey          string       `json:"apikey"`
	APISessionID    string       `json:"apisessionid"`
	ClientRequestID string       `json:"clientrequestid,omitempty"`
	DNSRecordSet    DNSRecordSet `json:"dnsrecordset"`
}

// DNSRecordSet as specified in netcup WSDL
// needed in UpdateDNSRecordsRequest
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecordset
type DNSRecordSet struct {
	DNSRecords []DNSRecord `json:"dnsrecords"`
}

// InfoDNSRecordsRequest as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#infoDnsRecords
type InfoDNSRecordsRequest struct {
	DomainName      string `json:"domainname"`
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APISessionID    string `json:"apisessionid"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// DNSRecord as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecord
type DNSRecord struct {
	ID           int    `json:"id,string,omitempty"`
	Hostname     string `json:"hostname"`
	RecordType   string `json:"type"`
	Priority     string `json:"priority,omitempty"`
	Destination  string `json:"destination"`
	DeleteRecord bool   `json:"deleterecord,omitempty"`
	State        string `json:"state,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
}

// ResponseMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Responsemessage
type ResponseMsg struct {
	ServerRequestID string          `json:"serverrequestid"`
	ClientRequestID string          `json:"clientrequestid,omitempty"`
	Action          string          `json:"action"`
	Status          string          `json:"status"`
	StatusCode      int             `json:"statuscode"`
	ShortMessage    string          `json:"shortmessage"`
	LongMessage     string          `json:"longmessage"`
	ResponseData    json.RawMessage `json:"responsedata,omitempty"`
}

func (r *ResponseMsg) Error() string {
	return fmt.Sprintf("an error occurred during the action %s: [Status=%s, StatusCode=%d, ShortMessage=%s, LongMessage=%s]",
		r.Action, r.Status, r.StatusCode, r.ShortMessage, r.LongMessage)
}

// LoginResponse response to login action.
type LoginResponse struct {
	APISessionID string `json:"apisessionid"`
}

// InfoDNSRecordsResponse response to infoDnsRecords action.
type InfoDNSRecordsResponse struct {
	APISessionID string      `json:"apisessionid"`
	DNSRecords   []DNSRecord `json:"dnsrecords,omitempty"`
}

// Client netcup DNS client
type Client struct {
	customerNumber string
	apiKey         string
	apiPassword    string
	HTTPClient     *http.Client
	BaseURL        string
}

// NewClient creates a netcup DNS client
func NewClient(customerNumber string, apiKey string, apiPassword string) (*Client, error) {
	if customerNumber == "" || apiKey == "" || apiPassword == "" {
		return nil, fmt.Errorf("credentials missing")
	}

	return &Client{
		customerNumber: customerNumber,
		apiKey:         apiKey,
		apiPassword:    apiPassword,
		BaseURL:        defaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Login performs the login as specified by the netcup WSDL
// returns sessionID needed to perform remaining actions
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) Login() (string, error) {
	payload := &Request{
		Action: "login",
		Param: &LoginRequest{
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APIPassword:     c.apiPassword,
			ClientRequestID: "",
		},
	}

	var responseData LoginResponse
	err := c.doRequest(payload, &responseData)
	if err != nil {
		return "", fmt.Errorf("loging error: %v", err)
	}

	return responseData.APISessionID, nil
}

// Logout performs the logout with the supplied sessionID as specified by the netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) Logout(sessionID string) error {
	payload := &Request{
		Action: "logout",
		Param: &LogoutRequest{
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    sessionID,
			ClientRequestID: "",
		},
	}

	err := c.doRequest(payload, nil)
	if err != nil {
		return fmt.Errorf("logout error: %v", err)
	}

	return nil
}

// UpdateDNSRecord performs an update of the DNSRecords as specified by the netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) UpdateDNSRecord(sessionID, domainName string, records []DNSRecord) error {
	payload := &Request{
		Action: "updateDnsRecords",
		Param: UpdateDNSRecordsRequest{
			DomainName:      domainName,
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    sessionID,
			ClientRequestID: "",
			DNSRecordSet:    DNSRecordSet{DNSRecords: records},
		},
	}

	err := c.doRequest(payload, nil)
	if err != nil {
		return fmt.Errorf("error when sending the request: %v", err)
	}

	return nil
}

// GetDNSRecords retrieves all dns records of an DNS-Zone as specified by the netcup WSDL
// returns an array of DNSRecords
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) GetDNSRecords(hostname, apiSessionID string) ([]DNSRecord, error) {
	payload := &Request{
		Action: "infoDnsRecords",
		Param: InfoDNSRecordsRequest{
			DomainName:      hostname,
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    apiSessionID,
			ClientRequestID: "",
		},
	}

	var responseData InfoDNSRecordsResponse
	err := c.doRequest(payload, &responseData)
	if err != nil {
		return nil, fmt.Errorf("error when sending the request: %v", err)
	}

	return responseData.DNSRecords, nil

}

// doRequest marshals given body to JSON, send the request to netcup API
// and returns body of response
func (c *Client) doRequest(payload interface{}, responseData interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Close = true
	req.Header.Set("content-type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if err = checkResponse(resp); err != nil {
		return err
	}

	respMsg, err := decodeResponseMsg(resp)
	if err != nil {
		return err
	}

	if respMsg.Status != success {
		return respMsg
	}

	if responseData != nil {
		err = json.Unmarshal(respMsg.ResponseData, responseData)
		if err != nil {
			return fmt.Errorf("%v: unmarshaling %T error: %v: %s",
				respMsg, responseData, err, string(respMsg.ResponseData))
		}
	}

	return nil
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode > 299 {
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

	return nil
}

func decodeResponseMsg(resp *http.Response) (*ResponseMsg, error) {
	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil, status code=%d", resp.StatusCode)
	}

	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: status code=%d, error=%v", resp.StatusCode, err)
	}

	var respMsg ResponseMsg
	err = json.Unmarshal(raw, &respMsg)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling %T error [status code=%d]: %v: %s", respMsg, resp.StatusCode, err, string(raw))
	}

	return &respMsg, nil
}

// GetDNSRecordIdx searches a given array of DNSRecords for a given DNSRecord
// equivalence is determined by Destination and RecortType attributes
// returns index of given DNSRecord in given array of DNSRecords
func GetDNSRecordIdx(records []DNSRecord, record DNSRecord) (int, error) {
	for index, element := range records {
		if record.Destination == element.Destination && record.RecordType == element.RecordType {
			return index, nil
		}
	}
	return -1, fmt.Errorf("no DNS Record found")
}
