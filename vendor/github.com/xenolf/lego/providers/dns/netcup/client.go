package netcup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
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

// LoginMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#login
type LoginMsg struct {
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APIPassword     string `json:"apipassword"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// LogoutMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#logout
type LogoutMsg struct {
	CustomerNumber  string `json:"customernumber"`
	APIKey          string `json:"apikey"`
	APISessionID    string `json:"apisessionid"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
}

// UpdateDNSRecordsMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#updateDnsRecords
type UpdateDNSRecordsMsg struct {
	DomainName      string       `json:"domainname"`
	CustomerNumber  string       `json:"customernumber"`
	APIKey          string       `json:"apikey"`
	APISessionID    string       `json:"apisessionid"`
	ClientRequestID string       `json:"clientrequestid,omitempty"`
	DNSRecordSet    DNSRecordSet `json:"dnsrecordset"`
}

// DNSRecordSet as specified in netcup WSDL
// needed in UpdateDNSRecordsMsg
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#Dnsrecordset
type DNSRecordSet struct {
	DNSRecords []DNSRecord `json:"dnsrecords"`
}

// InfoDNSRecordsMsg as specified in netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php#infoDnsRecords
type InfoDNSRecordsMsg struct {
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
	ServerRequestID string       `json:"serverrequestid"`
	ClientRequestID string       `json:"clientrequestid,omitempty"`
	Action          string       `json:"action"`
	Status          string       `json:"status"`
	StatusCode      int          `json:"statuscode"`
	ShortMessage    string       `json:"shortmessage"`
	LongMessage     string       `json:"longmessage"`
	ResponseData    ResponseData `json:"responsedata,omitempty"`
}

// LogoutResponseMsg similar to ResponseMsg
// allows empty ResponseData field whilst unmarshaling
type LogoutResponseMsg struct {
	ServerRequestID string `json:"serverrequestid"`
	ClientRequestID string `json:"clientrequestid,omitempty"`
	Action          string `json:"action"`
	Status          string `json:"status"`
	StatusCode      int    `json:"statuscode"`
	ShortMessage    string `json:"shortmessage"`
	LongMessage     string `json:"longmessage"`
	ResponseData    string `json:"responsedata,omitempty"`
}

// ResponseData to enable correct unmarshaling of ResponseMsg
type ResponseData struct {
	APISessionID string      `json:"apisessionid"`
	DNSRecords   []DNSRecord `json:"dnsrecords"`
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
func NewClient(customerNumber string, apiKey string, apiPassword string) *Client {
	return &Client{
		customerNumber: customerNumber,
		apiKey:         apiKey,
		apiPassword:    apiPassword,
		BaseURL:        defaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Login performs the login as specified by the netcup WSDL
// returns sessionID needed to perform remaining actions
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) Login() (string, error) {
	payload := &Request{
		Action: "login",
		Param: &LoginMsg{
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APIPassword:     c.apiPassword,
			ClientRequestID: "",
		},
	}

	response, err := c.sendRequest(payload)
	if err != nil {
		return "", fmt.Errorf("error sending request to DNS-API, %v", err)
	}

	var r ResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return "", fmt.Errorf("error decoding response of DNS-API, %v", err)
	}
	if r.Status != success {
		return "", fmt.Errorf("error logging into DNS-API, %v", r.LongMessage)
	}
	return r.ResponseData.APISessionID, nil
}

// Logout performs the logout with the supplied sessionID as specified by the netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) Logout(sessionID string) error {
	payload := &Request{
		Action: "logout",
		Param: &LogoutMsg{
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    sessionID,
			ClientRequestID: "",
		},
	}

	response, err := c.sendRequest(payload)
	if err != nil {
		return fmt.Errorf("error logging out of DNS-API: %v", err)
	}

	var r LogoutResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return fmt.Errorf("error logging out of DNS-API: %v", err)
	}

	if r.Status != success {
		return fmt.Errorf("error logging out of DNS-API: %v", r.ShortMessage)
	}
	return nil
}

// UpdateDNSRecord performs an update of the DNSRecords as specified by the netcup WSDL
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) UpdateDNSRecord(sessionID, domainName string, record DNSRecord) error {
	payload := &Request{
		Action: "updateDnsRecords",
		Param: UpdateDNSRecordsMsg{
			DomainName:      domainName,
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    sessionID,
			ClientRequestID: "",
			DNSRecordSet:    DNSRecordSet{DNSRecords: []DNSRecord{record}},
		},
	}

	response, err := c.sendRequest(payload)
	if err != nil {
		return err
	}

	var r ResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return err
	}

	if r.Status != success {
		return fmt.Errorf("%s: %+v", r.ShortMessage, r)
	}
	return nil
}

// GetDNSRecords retrieves all dns records of an DNS-Zone as specified by the netcup WSDL
// returns an array of DNSRecords
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) GetDNSRecords(hostname, apiSessionID string) ([]DNSRecord, error) {
	payload := &Request{
		Action: "infoDnsRecords",
		Param: InfoDNSRecordsMsg{
			DomainName:      hostname,
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    apiSessionID,
			ClientRequestID: "",
		},
	}

	response, err := c.sendRequest(payload)
	if err != nil {
		return nil, err
	}

	var r ResponseMsg

	err = json.Unmarshal(response, &r)
	if err != nil {
		return nil, err
	}

	if r.Status != success {
		return nil, fmt.Errorf("%s", r.ShortMessage)
	}
	return r.ResponseData.DNSRecords, nil

}

// sendRequest marshals given body to JSON, send the request to netcup API
// and returns body of response
func (c *Client) sendRequest(payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Close = true

	req.Header.Set("content-type", "application/json")
	req.Header.Set("User-Agent", acme.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("API request failed with HTTP Status code %d", resp.StatusCode)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read of response body failed, %v", err)
	}
	defer resp.Body.Close()

	return body, nil
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

// CreateTxtRecord uses the supplied values to return a DNSRecord of type TXT for the dns-01 challenge
func CreateTxtRecord(hostname, value string, ttl int) DNSRecord {
	return DNSRecord{
		ID:           0,
		Hostname:     hostname,
		RecordType:   "TXT",
		Priority:     "",
		Destination:  value,
		DeleteRecord: false,
		State:        "",
		TTL:          ttl,
	}
}
