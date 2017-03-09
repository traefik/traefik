package auroradnsclient

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/edeckers/auroradnsclient/records"
)

// GetRecords returns a list of all records in given zone
func (client *AuroraDNSClient) GetRecords(zoneID string) ([]records.GetRecordsResponse, error) {
	logrus.Debugf("GetRecords(%s)", zoneID)
	relativeURL := fmt.Sprintf("zones/%s/records", zoneID)

	response, err := client.requestor.Request(relativeURL, "GET", []byte(""))
	if err != nil {
		logrus.Errorf("Failed to receive records: %s", err)
		return nil, err
	}

	var respData []records.GetRecordsResponse
	err = json.Unmarshal(response, &respData)
	if err != nil {
		logrus.Errorf("Failed to unmarshall response: %s", err)
		return nil, err
	}

	return respData, nil
}

// CreateRecord creates a new record in given zone
func (client *AuroraDNSClient) CreateRecord(zoneID string, data records.CreateRecordRequest) (*records.CreateRecordResponse, error) {
	logrus.Debugf("CreateRecord(%s, %+v)", zoneID, data)
	body, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("Failed to marshall request body: %s", err)

		return nil, err
	}

	relativeURL := fmt.Sprintf("zones/%s/records", zoneID)

	response, err := client.requestor.Request(relativeURL, "POST", body)
	if err != nil {
		logrus.Errorf("Failed to create record: %s", err)

		return nil, err
	}

	var respData *records.CreateRecordResponse
	err = json.Unmarshal(response, &respData)
	if err != nil {
		logrus.Errorf("Failed to unmarshall response: %s", err)

		return nil, err
	}

	return respData, nil
}

// RemoveRecord removes a record corresponding to a particular id in a given zone
func (client *AuroraDNSClient) RemoveRecord(zoneID string, recordID string) (*records.RemoveRecordResponse, error) {
	logrus.Debugf("RemoveRecord(%s, %s)", zoneID, recordID)
	relativeURL := fmt.Sprintf("zones/%s/records/%s", zoneID, recordID)

	_, err := client.requestor.Request(relativeURL, "DELETE", nil)
	if err != nil {
		logrus.Errorf("Failed to remove record: %s", err)

		return nil, err
	}

	return &records.RemoveRecordResponse{}, nil
}
