package hostingde

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const defaultBaseURL = "https://secure.hosting.de/api/dns/v1/json"

// RecordsAddRequest represents a DNS record to add
type RecordsAddRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// RecordsDeleteRequest represents a DNS record to remove
type RecordsDeleteRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
}

// ZoneConfigObject represents the ZoneConfig-section of a hosting.de API response.
type ZoneConfigObject struct {
	AccountID      string `json:"accountId"`
	EmailAddress   string `json:"emailAddress"`
	ID             string `json:"id"`
	LastChangeDate string `json:"lastChangeDate"`
	MasterIP       string `json:"masterIp"`
	Name           string `json:"name"`
	NameUnicode    string `json:"nameUnicode"`
	SOAValues      struct {
		Expire      int    `json:"expire"`
		NegativeTTL int    `json:"negativeTtl"`
		Refresh     int    `json:"refresh"`
		Retry       int    `json:"retry"`
		Serial      string `json:"serial"`
		TTL         int    `json:"ttl"`
	} `json:"soaValues"`
	Status                string   `json:"status"`
	TemplateValues        string   `json:"templateValues"`
	Type                  string   `json:"type"`
	ZoneTransferWhitelist []string `json:"zoneTransferWhitelist"`
}

// ZoneUpdateError represents an error in a ZoneUpdateResponse
type ZoneUpdateError struct {
	Code          int      `json:"code"`
	ContextObject string   `json:"contextObject"`
	ContextPath   string   `json:"contextPath"`
	Details       []string `json:"details"`
	Text          string   `json:"text"`
	Value         string   `json:"value"`
}

// ZoneUpdateMetadata represents the metadata in a ZoneUpdateResponse
type ZoneUpdateMetadata struct {
	ClientTransactionID string `json:"clientTransactionId"`
	ServerTransactionID string `json:"serverTransactionId"`
}

// ZoneUpdateResponse represents a response from hosting.de API
type ZoneUpdateResponse struct {
	Errors   []ZoneUpdateError  `json:"errors"`
	Metadata ZoneUpdateMetadata `json:"metadata"`
	Warnings []string           `json:"warnings"`
	Status   string             `json:"status"`
	Response struct {
		Records []struct {
			Content          string `json:"content"`
			Type             string `json:"type"`
			ID               string `json:"id"`
			Name             string `json:"name"`
			LastChangeDate   string `json:"lastChangeDate"`
			Priority         int    `json:"priority"`
			RecordTemplateID string `json:"recordTemplateId"`
			ZoneConfigID     string `json:"zoneConfigId"`
			TTL              int    `json:"ttl"`
		} `json:"records"`
		ZoneConfig ZoneConfigObject `json:"zoneConfig"`
	} `json:"response"`
}

// ZoneConfigSelector represents a "minimal" ZoneConfig object used in hosting.de API requests
type ZoneConfigSelector struct {
	Name string `json:"name"`
}

// ZoneUpdateRequest represents a hosting.de API ZoneUpdate request
type ZoneUpdateRequest struct {
	AuthToken          string `json:"authToken"`
	ZoneConfigSelector `json:"zoneConfig"`
	RecordsToAdd       []RecordsAddRequest    `json:"recordsToAdd"`
	RecordsToDelete    []RecordsDeleteRequest `json:"recordsToDelete"`
}

func (d *DNSProvider) updateZone(updateRequest ZoneUpdateRequest) (*ZoneUpdateResponse, error) {
	body, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, defaultBaseURL+"/zoneUpdate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying API: %v", err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	updateResponse := &ZoneUpdateResponse{}
	err = json.Unmarshal(content, updateResponse)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if updateResponse.Status != "success" && updateResponse.Status != "pending" {
		return updateResponse, errors.New(toUnreadableBodyMessage(req, content))
	}

	return updateResponse, nil
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
