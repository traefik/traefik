package glesys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-acme/lego/log"
)

// types for JSON method calls, parameters, and responses

type addRecordRequest struct {
	DomainName string `json:"domainname"`
	Host       string `json:"host"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	TTL        int    `json:"ttl,omitempty"`
}

type deleteRecordRequest struct {
	RecordID int `json:"recordid"`
}

type responseStruct struct {
	Response struct {
		Status struct {
			Code int `json:"code"`
		} `json:"status"`
		Record deleteRecordRequest `json:"record"`
	} `json:"response"`
}

func (d *DNSProvider) addTXTRecord(fqdn string, domain string, name string, value string, ttl int) (int, error) {
	response, err := d.sendRequest(http.MethodPost, "addrecord", addRecordRequest{
		DomainName: domain,
		Host:       name,
		Type:       "TXT",
		Data:       value,
		TTL:        ttl,
	})

	if response != nil && response.Response.Status.Code == http.StatusOK {
		log.Infof("[%s]: Successfully created record id %d", fqdn, response.Response.Record.RecordID)
		return response.Response.Record.RecordID, nil
	}
	return 0, err
}

func (d *DNSProvider) deleteTXTRecord(fqdn string, recordid int) error {
	response, err := d.sendRequest(http.MethodPost, "deleterecord", deleteRecordRequest{
		RecordID: recordid,
	})
	if response != nil && response.Response.Status.Code == 200 {
		log.Infof("[%s]: Successfully deleted record id %d", fqdn, recordid)
	}
	return err
}

func (d *DNSProvider) sendRequest(method string, resource string, payload interface{}) (*responseStruct, error) {
	url := fmt.Sprintf("%s/%s", defaultBaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.config.APIUser, d.config.APIKey)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
	}

	var response responseStruct
	err = json.NewDecoder(resp.Body).Decode(&response)

	return &response, err
}
