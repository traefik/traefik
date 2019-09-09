package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// VolumeStatus indicates the status of the Volume
type VolumeStatus string

const (
	// VolumeCreating indicates the Volume is being created and is not yet available for use
	VolumeCreating VolumeStatus = "creating"

	// VolumeActive indicates the Volume is online and available for use
	VolumeActive VolumeStatus = "active"

	// VolumeResizing indicates the Volume is in the process of upgrading its current capacity
	VolumeResizing VolumeStatus = "resizing"

	// VolumeContactSupport indicates there is a problem with the Volume. A support ticket must be opened to resolve the issue
	VolumeContactSupport VolumeStatus = "contact_support"
)

// Volume represents a linode volume object
type Volume struct {
	CreatedStr string `json:"created"`
	UpdatedStr string `json:"updated"`

	ID             int          `json:"id"`
	Label          string       `json:"label"`
	Status         VolumeStatus `json:"status"`
	Region         string       `json:"region"`
	Size           int          `json:"size"`
	LinodeID       *int         `json:"linode_id"`
	FilesystemPath string       `json:"filesystem_path"`
	Tags           []string     `json:"tags"`
	Created        time.Time    `json:"-"`
	Updated        time.Time    `json:"-"`
}

// VolumeCreateOptions fields are those accepted by CreateVolume
type VolumeCreateOptions struct {
	Label    string `json:"label,omitempty"`
	Region   string `json:"region,omitempty"`
	LinodeID int    `json:"linode_id,omitempty"`
	ConfigID int    `json:"config_id,omitempty"`
	// The Volume's size, in GiB. Minimum size is 10GiB, maximum size is 10240GiB. A "0" value will result in the default size.
	Size int `json:"size,omitempty"`
	// An array of tags applied to this object. Tags are for organizational purposes only.
	Tags               []string `json:"tags"`
	PersistAcrossBoots *bool    `json:"persist_across_boots,omitempty"`
}

// VolumeUpdateOptions fields are those accepted by UpdateVolume
type VolumeUpdateOptions struct {
	Label string    `json:"label,omitempty"`
	Tags  *[]string `json:"tags,omitempty"`
}

// VolumeAttachOptions fields are those accepted by AttachVolume
type VolumeAttachOptions struct {
	LinodeID           int   `json:"linode_id"`
	ConfigID           int   `json:"config_id,omitempty"`
	PersistAcrossBoots *bool `json:"persist_across_boots,omitempty"`
}

// VolumesPagedResponse represents a linode API response for listing of volumes
type VolumesPagedResponse struct {
	*PageOptions
	Data []Volume `json:"data"`
}

// GetUpdateOptions converts a Volume to VolumeUpdateOptions for use in UpdateVolume
func (v Volume) GetUpdateOptions() (updateOpts VolumeUpdateOptions) {
	updateOpts.Label = v.Label
	updateOpts.Tags = &v.Tags
	return
}

// GetCreateOptions converts a Volume to VolumeCreateOptions for use in CreateVolume
func (v Volume) GetCreateOptions() (createOpts VolumeCreateOptions) {
	createOpts.Label = v.Label
	createOpts.Tags = v.Tags
	createOpts.Region = v.Region
	createOpts.Size = v.Size
	if v.LinodeID != nil && *v.LinodeID > 0 {
		createOpts.LinodeID = *v.LinodeID
	}
	return
}

// endpoint gets the endpoint URL for Volume
func (VolumesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Volumes.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Volumes when processing paginated Volume responses
func (resp *VolumesPagedResponse) appendData(r *VolumesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListVolumes lists Volumes
func (c *Client) ListVolumes(ctx context.Context, opts *ListOptions) ([]Volume, error) {
	response := VolumesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (v *Volume) fixDates() *Volume {
	if parsed, err := parseDates(v.CreatedStr); err != nil {
		v.Created = *parsed
	}
	if parsed, err := parseDates(v.UpdatedStr); err != nil {
		v.Updated = *parsed
	}
	return v
}

// GetVolume gets the template with the provided ID
func (c *Client) GetVolume(ctx context.Context, id int) (*Volume, error) {
	e, err := c.Volumes.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&Volume{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Volume).fixDates(), nil
}

// AttachVolume attaches a volume to a Linode instance
func (c *Client) AttachVolume(ctx context.Context, id int, options *VolumeAttachOptions) (*Volume, error) {
	body := ""
	if bodyData, err := json.Marshal(options); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	e, err := c.Volumes.Endpoint()
	if err != nil {
		return nil, NewError(err)
	}

	e = fmt.Sprintf("%s/%d/attach", e, id)
	resp, err := coupleAPIErrors(c.R(ctx).
		SetResult(&Volume{}).
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return resp.Result().(*Volume).fixDates(), nil
}

// CreateVolume creates a Linode Volume
func (c *Client) CreateVolume(ctx context.Context, createOpts VolumeCreateOptions) (*Volume, error) {
	body := ""
	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	e, err := c.Volumes.Endpoint()
	if err != nil {
		return nil, NewError(err)
	}

	resp, err := coupleAPIErrors(c.R(ctx).
		SetResult(&Volume{}).
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return resp.Result().(*Volume).fixDates(), nil
}

// RenameVolume renames the label of a Linode volume
// DEPRECATED: use UpdateVolume
func (c *Client) RenameVolume(ctx context.Context, id int, label string) (*Volume, error) {
	updateOpts := VolumeUpdateOptions{Label: label}
	return c.UpdateVolume(ctx, id, updateOpts)
}

// UpdateVolume updates the Volume with the specified id
func (c *Client) UpdateVolume(ctx context.Context, id int, volume VolumeUpdateOptions) (*Volume, error) {
	var body string
	e, err := c.Volumes.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&Volume{})

	if bodyData, err := json.Marshal(volume); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Volume).fixDates(), nil
}

// CloneVolume clones a Linode volume
func (c *Client) CloneVolume(ctx context.Context, id int, label string) (*Volume, error) {
	body := fmt.Sprintf("{\"label\":\"%s\"}", label)

	e, err := c.Volumes.Endpoint()
	if err != nil {
		return nil, NewError(err)
	}
	e = fmt.Sprintf("%s/%d/clone", e, id)

	resp, err := coupleAPIErrors(c.R(ctx).
		SetResult(&Volume{}).
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return resp.Result().(*Volume).fixDates(), nil
}

// DetachVolume detaches a Linode volume
func (c *Client) DetachVolume(ctx context.Context, id int) error {
	body := ""

	e, err := c.Volumes.Endpoint()
	if err != nil {
		return NewError(err)
	}

	e = fmt.Sprintf("%s/%d/detach", e, id)

	_, err = coupleAPIErrors(c.R(ctx).
		SetBody(body).
		Post(e))

	return err
}

// ResizeVolume resizes an instance to new Linode type
func (c *Client) ResizeVolume(ctx context.Context, id int, size int) error {
	body := fmt.Sprintf("{\"size\": %d}", size)

	e, err := c.Volumes.Endpoint()
	if err != nil {
		return NewError(err)
	}
	e = fmt.Sprintf("%s/%d/resize", e, id)

	_, err = coupleAPIErrors(c.R(ctx).
		SetBody(body).
		Post(e))

	return err
}

// DeleteVolume deletes the Volume with the specified id
func (c *Client) DeleteVolume(ctx context.Context, id int) error {
	e, err := c.Volumes.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
