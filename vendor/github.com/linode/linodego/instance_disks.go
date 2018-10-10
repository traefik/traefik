package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// InstanceDisk represents an Instance Disk object
type InstanceDisk struct {
	CreatedStr string `json:"created"`
	UpdatedStr string `json:"updated"`

	ID         int            `json:"id"`
	Label      string         `json:"label"`
	Status     string         `json:"status"`
	Size       int            `json:"size"`
	Filesystem DiskFilesystem `json:"filesystem"`
	Created    time.Time      `json:"-"`
	Updated    time.Time      `json:"-"`
}

// DiskFilesystem constants start with Filesystem and include Linode API Filesystems
type DiskFilesystem string

// DiskFilesystem constants represent the filesystems types an Instance Disk may use
const (
	FilesystemRaw    DiskFilesystem = "raw"
	FilesystemSwap   DiskFilesystem = "swap"
	FilesystemExt3   DiskFilesystem = "ext3"
	FilesystemExt4   DiskFilesystem = "ext4"
	FilesystemInitrd DiskFilesystem = "initrd"
)

// InstanceDisksPagedResponse represents a paginated InstanceDisk API response
type InstanceDisksPagedResponse struct {
	*PageOptions
	Data []InstanceDisk `json:"data"`
}

// InstanceDiskCreateOptions are InstanceDisk settings that can be used at creation
type InstanceDiskCreateOptions struct {
	Label string `json:"label"`
	Size  int    `json:"size"`

	// Image is optional, but requires RootPass if provided
	Image    string `json:"image,omitempty"`
	RootPass string `json:"root_pass,omitempty"`

	Filesystem      string            `json:"filesystem,omitempty"`
	AuthorizedKeys  []string          `json:"authorized_keys,omitempty"`
	AuthorizedUsers []string          `json:"authorized_users,omitempty"`
	ReadOnly        bool              `json:"read_only,omitempty"`
	StackscriptID   int               `json:"stackscript_id,omitempty"`
	StackscriptData map[string]string `json:"stackscript_data,omitempty"`
}

// InstanceDiskUpdateOptions are InstanceDisk settings that can be used in updates
type InstanceDiskUpdateOptions struct {
	Label    string `json:"label"`
	ReadOnly bool   `json:"read_only"`
}

// endpointWithID gets the endpoint URL for InstanceDisks of a given Instance
func (InstanceDisksPagedResponse) endpointWithID(c *Client, id int) string {
	endpoint, err := c.InstanceDisks.endpointWithID(id)
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends InstanceDisks when processing paginated InstanceDisk responses
func (resp *InstanceDisksPagedResponse) appendData(r *InstanceDisksPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListInstanceDisks lists InstanceDisks
func (c *Client) ListInstanceDisks(ctx context.Context, linodeID int, opts *ListOptions) ([]InstanceDisk, error) {
	response := InstanceDisksPagedResponse{}
	err := c.listHelperWithID(ctx, &response, linodeID, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (v *InstanceDisk) fixDates() *InstanceDisk {
	if created, err := parseDates(v.CreatedStr); err == nil {
		v.Created = *created
	}
	if updated, err := parseDates(v.UpdatedStr); err == nil {
		v.Updated = *updated
	}
	return v
}

// GetInstanceDisk gets the template with the provided ID
func (c *Client) GetInstanceDisk(ctx context.Context, linodeID int, configID int) (*InstanceDisk, error) {
	e, err := c.InstanceDisks.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, configID)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&InstanceDisk{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceDisk).fixDates(), nil
}

// CreateInstanceDisk creates a new InstanceDisk for the given Instance
func (c *Client) CreateInstanceDisk(ctx context.Context, linodeID int, createOpts InstanceDiskCreateOptions) (*InstanceDisk, error) {
	var body string
	e, err := c.InstanceDisks.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&InstanceDisk{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*InstanceDisk).fixDates(), nil
}

// UpdateInstanceDisk creates a new InstanceDisk for the given Instance
func (c *Client) UpdateInstanceDisk(ctx context.Context, linodeID int, diskID int, updateOpts InstanceDiskUpdateOptions) (*InstanceDisk, error) {
	var body string
	e, err := c.InstanceDisks.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, diskID)

	req := c.R(ctx).SetResult(&InstanceDisk{})

	if bodyData, err := json.Marshal(updateOpts); err == nil {
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

	return r.Result().(*InstanceDisk).fixDates(), nil
}

// RenameInstanceDisk renames an InstanceDisk
func (c *Client) RenameInstanceDisk(ctx context.Context, linodeID int, diskID int, label string) (*InstanceDisk, error) {
	return c.UpdateInstanceDisk(ctx, linodeID, diskID, InstanceDiskUpdateOptions{Label: label})
}

// ResizeInstanceDisk resizes the size of the Instance disk
func (c *Client) ResizeInstanceDisk(ctx context.Context, linodeID int, diskID int, size int) (*InstanceDisk, error) {
	var body string
	e, err := c.InstanceDisks.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, diskID)

	req := c.R(ctx).SetResult(&InstanceDisk{})
	updateOpts := map[string]interface{}{
		"size": size,
	}

	if bodyData, err := json.Marshal(updateOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceDisk).fixDates(), nil
}

// DeleteInstanceDisk deletes a Linode Instance Disk
func (c *Client) DeleteInstanceDisk(ctx context.Context, linodeID int, diskID int) error {
	e, err := c.InstanceDisks.endpointWithID(linodeID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, diskID)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
