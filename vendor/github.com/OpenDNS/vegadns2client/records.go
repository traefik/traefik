package vegadns2client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Record - struct representing a Record object
type Record struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	RecordType string `json:"record_type"`
	TTL        int    `json:"ttl"`
	RecordID   int    `json:"record_id"`
	LocationID string `json:"location_id"`
	DomainID   int    `json:"domain_id"`
}

// RecordsResponse - api response list of records
type RecordsResponse struct {
	Status  string   `json:"status"`
	Total   int      `json:"total_records"`
	Domain  Domain   `json:"domain"`
	Records []Record `json:"records"`
}

// GetRecordID - helper to get the id of a record
// Input: domainID, record, recordType
// Output: int
func (vega *VegaDNSClient) GetRecordID(domainID int, record string, recordType string) (int, error) {
	params := make(map[string]string)
	params["domain_id"] = fmt.Sprintf("%d", domainID)

	resp, err := vega.Send("GET", "records", params)

	if err != nil {
		return -1, fmt.Errorf("Error sending GET to GetRecordID: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response from GetRecordID: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Got bad answer from VegaDNS on GetRecordID. Code: %d. Message: %s", resp.StatusCode, string(body))
	}

	answer := RecordsResponse{}
	if err := json.Unmarshal(body, &answer); err != nil {
		return -1, fmt.Errorf("Error unmarshalling body from GetRecordID: %s", err)
	}

	for _, r := range answer.Records {
		if r.Name == record && r.RecordType == recordType {
			return r.RecordID, nil
		}
	}

	return -1, errors.New("Couldnt find record")
}

// CreateTXT - Creates a TXT record
// Input: domainID, fqdn, value, ttl
// Output: nil or error
func (vega *VegaDNSClient) CreateTXT(domainID int, fqdn string, value string, ttl int) error {
	params := make(map[string]string)

	params["record_type"] = "TXT"
	params["ttl"] = fmt.Sprintf("%d", ttl)
	params["domain_id"] = fmt.Sprintf("%d", domainID)
	params["name"] = strings.TrimSuffix(fqdn, ".")
	params["value"] = value

	resp, err := vega.Send("POST", "records", params)

	if err != nil {
		return fmt.Errorf("Send POST error in CreateTXT: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading POST response in CreateTXT: %s", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Got bad answer from VegaDNS on CreateTXT. Code: %d. Message: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteRecord - deletes a record by id
// Input: recordID
// Output: nil or error
func (vega *VegaDNSClient) DeleteRecord(recordID int) error {
	resp, err := vega.Send("DELETE", fmt.Sprintf("records/%d", recordID), nil)
	if err != nil {
		return fmt.Errorf("Send DELETE error in DeleteTXT: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading DELETE response in DeleteTXT: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Got bad answer from VegaDNS on DeleteTXT. Code: %d. Message: %s", resp.StatusCode, string(body))
	}

	return nil
}
