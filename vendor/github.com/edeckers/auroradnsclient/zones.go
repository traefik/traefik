package auroradnsclient

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/edeckers/auroradnsclient/zones"
)

// GetZones returns a list of all zones
func (client *AuroraDNSClient) GetZones() ([]zones.ZoneRecord, error) {
	logrus.Debugf("GetZones")
	response, err := client.requestor.Request("zones", "GET", []byte(""))

	if err != nil {
		logrus.Errorf("Failed to get zones: %s", err)
		return nil, err
	}

	var respData []zones.ZoneRecord
	err = json.Unmarshal(response, &respData)
	if err != nil {
		logrus.Errorf("Failed to unmarshall response: %s", err)
		return nil, err
	}

	logrus.Debugf("Unmarshalled response: %+v", respData)
	return respData, nil
}
