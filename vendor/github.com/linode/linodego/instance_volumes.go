package linodego

import (
	"context"
)

// InstanceVolumesPagedResponse represents a paginated InstanceVolume API response
type InstanceVolumesPagedResponse struct {
	*PageOptions
	Data []Volume `json:"data"`
}

// endpoint gets the endpoint URL for InstanceVolume
func (InstanceVolumesPagedResponse) endpointWithID(c *Client, id int) string {
	endpoint, err := c.InstanceVolumes.endpointWithID(id)
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends InstanceVolumes when processing paginated InstanceVolume responses
func (resp *InstanceVolumesPagedResponse) appendData(r *InstanceVolumesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListInstanceVolumes lists InstanceVolumes
func (c *Client) ListInstanceVolumes(ctx context.Context, linodeID int, opts *ListOptions) ([]Volume, error) {
	response := InstanceVolumesPagedResponse{}
	err := c.listHelperWithID(ctx, &response, linodeID, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}
