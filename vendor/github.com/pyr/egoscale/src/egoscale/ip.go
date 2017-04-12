package egoscale

import (
	"encoding/json"
	"net/url"
)

func (exo *Client) AddIpToNic(nic_id string, ip_address string) (string, error) {
	params := url.Values{}
	params.Set("nicid", nic_id)
	params.Set("ipaddress", ip_address)

	resp, err := exo.Request("addIpToNic", params)
	if err != nil {
		return "", err
	}

	var r AddIpToNicResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.Id, nil
}

func (exo *Client) RemoveIpFromNic(nic_id string) (string, error) {
	params := url.Values{}
	params.Set("id", nic_id)

	resp, err := exo.Request("removeIpFromNic", params)
	if err != nil {
		return "", err
	}

	var r RemoveIpFromNicResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}
	return r.JobID, nil
}
