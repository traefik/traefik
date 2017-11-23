package egoscale

import (
	"encoding/json"
	"net/url"
)

func (exo *Client) PollAsyncJob(jobid string) (*QueryAsyncJobResultResponse, error) {
	params := url.Values{}

	params.Set("jobid", jobid)

	resp, err := exo.Request("queryAsyncJobResult", params)

	if err != nil {
		return nil, err
	}

	var r QueryAsyncJobResultResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (exo *Client) AsyncToVirtualMachine(resp QueryAsyncJobResultResponse) (*DeployVirtualMachineResponse, error) {
	var r DeployVirtualMachineWrappedResponse

	if err := json.Unmarshal(resp.Jobresult, &r); err != nil {
		return nil, err
	}

	return &r.Wrapped, nil
}
