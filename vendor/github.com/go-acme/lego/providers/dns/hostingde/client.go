package hostingde

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

const defaultBaseURL = "https://secure.hosting.de/api/dns/v1/json"

// https://www.hosting.de/api/?json#list-zoneconfigs
func (d *DNSProvider) listZoneConfigs(findRequest ZoneConfigsFindRequest) (*ZoneConfigsFindResponse, error) {
	uri := defaultBaseURL + "/zoneConfigsFind"

	findResponse := &ZoneConfigsFindResponse{}

	rawResp, err := d.post(uri, findRequest, findResponse)
	if err != nil {
		return nil, err
	}

	if len(findResponse.Response.Data) == 0 {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(uri, rawResp))
	}

	if findResponse.Status != "success" && findResponse.Status != "pending" {
		return findResponse, errors.New(toUnreadableBodyMessage(uri, rawResp))
	}

	return findResponse, nil
}

// https://www.hosting.de/api/?json#updating-zones
func (d *DNSProvider) updateZone(updateRequest ZoneUpdateRequest) (*ZoneUpdateResponse, error) {
	uri := defaultBaseURL + "/zoneUpdate"

	// but we'll need the ID later to delete the record
	updateResponse := &ZoneUpdateResponse{}

	rawResp, err := d.post(uri, updateRequest, updateResponse)
	if err != nil {
		return nil, err
	}

	if updateResponse.Status != "success" && updateResponse.Status != "pending" {
		return nil, errors.New(toUnreadableBodyMessage(uri, rawResp))
	}

	return updateResponse, nil
}

func (d *DNSProvider) getZone(findRequest ZoneConfigsFindRequest) (*ZoneConfig, error) {
	ctx, cancel := context.WithCancel(context.Background())

	var zoneConfig *ZoneConfig

	operation := func() error {
		findResponse, err := d.listZoneConfigs(findRequest)
		if err != nil {
			cancel()
			return err
		}

		if findResponse.Response.Data[0].Status != "active" {
			return fmt.Errorf("unexpected status: %q", findResponse.Response.Data[0].Status)
		}

		zoneConfig = &findResponse.Response.Data[0]

		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 3 * time.Second
	bo.MaxInterval = 10 * bo.InitialInterval
	bo.MaxElapsedTime = 100 * bo.InitialInterval

	// retry in case the zone was edited recently and is not yet active
	err := backoff.Retry(operation, backoff.WithContext(bo, ctx))
	if err != nil {
		return nil, err
	}

	return zoneConfig, nil
}

func (d *DNSProvider) post(uri string, request interface{}, response interface{}) ([]byte, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(body))
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
		return nil, errors.New(toUnreadableBodyMessage(uri, content))
	}

	err = json.Unmarshal(content, response)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(uri, content))
	}

	return content, nil
}

func toUnreadableBodyMessage(uri string, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", uri, string(rawBody))
}
