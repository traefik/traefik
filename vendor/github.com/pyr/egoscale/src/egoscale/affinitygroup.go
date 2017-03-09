package egoscale

import (
	"encoding/json"
	"net/url"
)

func (exo *Client) CreateAffinityGroup(name string) (string, error) {
	params := url.Values{}
	params.Set("name", name)
	params.Set("type", "host anti-affinity")

	resp, err := exo.Request("createAffinityGroup", params)
	if err != nil {
		return "", err
	}

	var r CreateAffinityGroupResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobId, nil
}

func (exo *Client) DeleteAffinityGroup(name string) (string, error) {
	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("deleteAffinityGroup", params)
	if err != nil {
		return "", err
	}

	var r DeleteAffinityGroupResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobId, nil

}
