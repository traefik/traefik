package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// InstanceConfig represents all of the settings that control the boot and run configuration of a Linode Instance
type InstanceConfig struct {
	CreatedStr string `json:"created"`
	UpdatedStr string `json:"updated"`

	ID          int                      `json:"id"`
	Label       string                   `json:"label"`
	Comments    string                   `json:"comments"`
	Devices     *InstanceConfigDeviceMap `json:"devices"`
	Helpers     *InstanceConfigHelpers   `json:"helpers"`
	MemoryLimit int                      `json:"memory_limit"`
	Kernel      string                   `json:"kernel"`
	InitRD      *int                     `json:"init_rd"`
	RootDevice  string                   `json:"root_device"`
	RunLevel    string                   `json:"run_level"`
	VirtMode    string                   `json:"virt_mode"`
	Created     *time.Time               `json:"-"`
	Updated     *time.Time               `json:"-"`
}

// InstanceConfigDevice contains either the DiskID or VolumeID assigned to a Config Device
type InstanceConfigDevice struct {
	DiskID   int `json:"disk_id,omitempty"`
	VolumeID int `json:"volume_id,omitempty"`
}

// InstanceConfigDeviceMap contains SDA-SDH InstanceConfigDevice settings
type InstanceConfigDeviceMap struct {
	SDA *InstanceConfigDevice `json:"sda,omitempty"`
	SDB *InstanceConfigDevice `json:"sdb,omitempty"`
	SDC *InstanceConfigDevice `json:"sdc,omitempty"`
	SDD *InstanceConfigDevice `json:"sdd,omitempty"`
	SDE *InstanceConfigDevice `json:"sde,omitempty"`
	SDF *InstanceConfigDevice `json:"sdf,omitempty"`
	SDG *InstanceConfigDevice `json:"sdg,omitempty"`
	SDH *InstanceConfigDevice `json:"sdh,omitempty"`
}

// InstanceConfigHelpers are Instance Config options that control Linux distribution specific tweaks
type InstanceConfigHelpers struct {
	UpdateDBDisabled  bool `json:"updatedb_disabled"`
	Distro            bool `json:"distro"`
	ModulesDep        bool `json:"modules_dep"`
	Network           bool `json:"network"`
	DevTmpFsAutomount bool `json:"devtmpfs_automount"`
}

// InstanceConfigsPagedResponse represents a paginated InstanceConfig API response
type InstanceConfigsPagedResponse struct {
	*PageOptions
	Data []InstanceConfig `json:"data"`
}

// InstanceConfigCreateOptions are InstanceConfig settings that can be used at creation
type InstanceConfigCreateOptions struct {
	Label       string                  `json:"label,omitempty"`
	Comments    string                  `json:"comments,omitempty"`
	Devices     InstanceConfigDeviceMap `json:"devices"`
	Helpers     *InstanceConfigHelpers  `json:"helpers,omitempty"`
	MemoryLimit int                     `json:"memory_limit,omitempty"`
	Kernel      string                  `json:"kernel,omitempty"`
	InitRD      int                     `json:"init_rd,omitempty"`
	RootDevice  *string                 `json:"root_device,omitempty"`
	RunLevel    string                  `json:"run_level,omitempty"`
	VirtMode    string                  `json:"virt_mode,omitempty"`
}

// InstanceConfigUpdateOptions are InstanceConfig settings that can be used in updates
type InstanceConfigUpdateOptions struct {
	Label    string                   `json:"label,omitempty"`
	Comments string                   `json:"comments"`
	Devices  *InstanceConfigDeviceMap `json:"devices,omitempty"`
	Helpers  *InstanceConfigHelpers   `json:"helpers,omitempty"`
	// MemoryLimit 0 means unlimitted, this is not omitted
	MemoryLimit int    `json:"memory_limit"`
	Kernel      string `json:"kernel,omitempty"`
	// InitRD is nullable, permit the sending of null
	InitRD     *int   `json:"init_rd"`
	RootDevice string `json:"root_device,omitempty"`
	RunLevel   string `json:"run_level,omitempty"`
	VirtMode   string `json:"virt_mode,omitempty"`
}

// GetCreateOptions converts a InstanceConfig to InstanceConfigCreateOptions for use in CreateInstanceConfig
func (i InstanceConfig) GetCreateOptions() InstanceConfigCreateOptions {
	initrd := 0
	if i.InitRD != nil {
		initrd = *i.InitRD
	}
	return InstanceConfigCreateOptions{
		Label:       i.Label,
		Comments:    i.Comments,
		Devices:     *i.Devices,
		Helpers:     i.Helpers,
		MemoryLimit: i.MemoryLimit,
		Kernel:      i.Kernel,
		InitRD:      initrd,
		RootDevice:  copyString(&i.RootDevice),
		RunLevel:    i.RunLevel,
		VirtMode:    i.VirtMode,
	}
}

// GetUpdateOptions converts a InstanceConfig to InstanceConfigUpdateOptions for use in UpdateInstanceConfig
func (i InstanceConfig) GetUpdateOptions() InstanceConfigUpdateOptions {
	return InstanceConfigUpdateOptions{
		Label:       i.Label,
		Comments:    i.Comments,
		Devices:     i.Devices,
		Helpers:     i.Helpers,
		MemoryLimit: i.MemoryLimit,
		Kernel:      i.Kernel,
		InitRD:      copyInt(i.InitRD),
		RootDevice:  i.RootDevice,
		RunLevel:    i.RunLevel,
		VirtMode:    i.VirtMode,
	}
}

// endpointWithID gets the endpoint URL for InstanceConfigs of a given Instance
func (InstanceConfigsPagedResponse) endpointWithID(c *Client, id int) string {
	endpoint, err := c.InstanceConfigs.endpointWithID(id)
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends InstanceConfigs when processing paginated InstanceConfig responses
func (resp *InstanceConfigsPagedResponse) appendData(r *InstanceConfigsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListInstanceConfigs lists InstanceConfigs
func (c *Client) ListInstanceConfigs(ctx context.Context, linodeID int, opts *ListOptions) ([]InstanceConfig, error) {
	response := InstanceConfigsPagedResponse{}
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
func (i *InstanceConfig) fixDates() *InstanceConfig {
	i.Created, _ = parseDates(i.CreatedStr)
	i.Updated, _ = parseDates(i.UpdatedStr)
	return i
}

// GetInstanceConfig gets the template with the provided ID
func (c *Client) GetInstanceConfig(ctx context.Context, linodeID int, configID int) (*InstanceConfig, error) {
	e, err := c.InstanceConfigs.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, configID)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&InstanceConfig{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceConfig).fixDates(), nil
}

// CreateInstanceConfig creates a new InstanceConfig for the given Instance
func (c *Client) CreateInstanceConfig(ctx context.Context, linodeID int, createOpts InstanceConfigCreateOptions) (*InstanceConfig, error) {
	var body string
	e, err := c.InstanceConfigs.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&InstanceConfig{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, err
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*InstanceConfig).fixDates(), nil
}

// UpdateInstanceConfig update an InstanceConfig for the given Instance
func (c *Client) UpdateInstanceConfig(ctx context.Context, linodeID int, configID int, updateOpts InstanceConfigUpdateOptions) (*InstanceConfig, error) {
	var body string
	e, err := c.InstanceConfigs.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, configID)
	req := c.R(ctx).SetResult(&InstanceConfig{})

	if bodyData, err := json.Marshal(updateOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, err
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*InstanceConfig).fixDates(), nil
}

// RenameInstanceConfig renames an InstanceConfig
func (c *Client) RenameInstanceConfig(ctx context.Context, linodeID int, configID int, label string) (*InstanceConfig, error) {
	return c.UpdateInstanceConfig(ctx, linodeID, configID, InstanceConfigUpdateOptions{Label: label})
}

// DeleteInstanceConfig deletes a Linode InstanceConfig
func (c *Client) DeleteInstanceConfig(ctx context.Context, linodeID int, configID int) error {
	e, err := c.InstanceConfigs.endpointWithID(linodeID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, configID)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
