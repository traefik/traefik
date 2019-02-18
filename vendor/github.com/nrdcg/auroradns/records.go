package auroradns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Record types
const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeCNAME = "CNAME"
	RecordTypeMX    = "MX"
	RecordTypeNS    = "NS"
	RecordTypeSOA   = "SOA"
	RecordTypeSRV   = "SRV"
	RecordTypeTXT   = "TXT"
	RecordTypeDS    = "DS"
	RecordTypePTR   = "PTR"
	RecordTypeSSHFP = "SSHFP"
	RecordTypeTLSA  = "TLS"
)

// Record a DNS record
type Record struct {
	ID         string `json:"id,omitempty"`
	RecordType string `json:"type"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl,omitempty"`
}

// CreateRecord Creates a new record.
func (c *Client) CreateRecord(zoneID string, record Record) (*Record, *http.Response, error) {
	body, err := json.Marshal(record)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshall request body: %v", err)
	}

	resource := fmt.Sprintf("/zones/%s/records", zoneID)

	req, err := c.newRequest(http.MethodPost, resource, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	newRecord := new(Record)
	resp, err := c.do(req, newRecord)
	if err != nil {
		return nil, resp, err
	}

	return newRecord, resp, nil
}

// DeleteRecord Delete a record.
func (c *Client) DeleteRecord(zoneID string, recordID string) (bool, *http.Response, error) {
	resource := fmt.Sprintf("/zones/%s/records/%s", zoneID, recordID)

	req, err := c.newRequest(http.MethodDelete, resource, nil)
	if err != nil {
		return false, nil, err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return false, resp, err
	}

	return true, resp, nil
}

// ListRecords returns a list of all records in given zone
func (c *Client) ListRecords(zoneID string) ([]Record, *http.Response, error) {
	resource := fmt.Sprintf("/zones/%s/records", zoneID)

	req, err := c.newRequest(http.MethodGet, resource, nil)
	if err != nil {
		return nil, nil, err
	}

	var records []Record
	resp, err := c.do(req, &records)
	if err != nil {
		return nil, resp, err
	}

	return records, resp, nil
}
