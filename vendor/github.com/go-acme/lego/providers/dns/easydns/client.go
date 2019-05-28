package easydns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
)

const defaultEndpoint = "https://rest.easydns.net"

type zoneRecord struct {
	ID      string `json:"id,omitempty"`
	Domain  string `json:"domain"`
	Host    string `json:"host"`
	TTL     string `json:"ttl"`
	Prio    string `json:"prio"`
	Type    string `json:"type"`
	Rdata   string `json:"rdata"`
	LastMod string `json:"last_mod,omitempty"`
	Revoked int    `json:"revoked,omitempty"`
	NewHost string `json:"new_host,omitempty"`
}

type addRecordResponse struct {
	Msg    string     `json:"msg"`
	Tm     int        `json:"tm"`
	Data   zoneRecord `json:"data"`
	Status int        `json:"status"`
}

func (d *DNSProvider) addRecord(domain string, record interface{}) (string, error) {
	pathAdd := path.Join("/zones/records/add", domain, "TXT")

	response := &addRecordResponse{}
	err := d.doRequest(http.MethodPut, pathAdd, record, response)
	if err != nil {
		return "", err
	}

	recordID := response.Data.ID

	return recordID, nil
}

func (d *DNSProvider) deleteRecord(domain, recordID string) error {
	pathDelete := path.Join("/zones/records", domain, recordID)

	return d.doRequest(http.MethodDelete, pathDelete, nil, nil)
}

func (d *DNSProvider) doRequest(method, path string, requestMsg, responseMsg interface{}) error {
	reqBody := &bytes.Buffer{}
	if requestMsg != nil {
		err := json.NewEncoder(reqBody).Encode(requestMsg)
		if err != nil {
			return err
		}
	}

	endpoint, err := d.config.Endpoint.Parse(path + "?format=json")
	if err != nil {
		return err
	}

	request, err := http.NewRequest(method, endpoint.String(), reqBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.SetBasicAuth(d.config.Token, d.config.Key)

	response, err := d.config.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("%d: failed to read response body: %v", response.StatusCode, err)
		}

		return fmt.Errorf("%d: request failed: %v", response.StatusCode, string(body))
	}

	if responseMsg != nil {
		return json.NewDecoder(response.Body).Decode(responseMsg)
	}

	return nil
}
