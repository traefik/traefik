package linodego

import (
	"context"
	"fmt"
)

// LinodeType represents a linode type object
type LinodeType struct {
	ID         string          `json:"id"`
	Disk       int             `json:"disk"`
	Class      LinodeTypeClass `json:"class"` // enum: nanode, standard, highmem, dedicated
	Price      *LinodePrice    `json:"price"`
	Label      string          `json:"label"`
	Addons     *LinodeAddons   `json:"addons"`
	NetworkOut int             `json:"network_out"`
	Memory     int             `json:"memory"`
	Transfer   int             `json:"transfer"`
	VCPUs      int             `json:"vcpus"`
}

// LinodePrice represents a linode type price object
type LinodePrice struct {
	Hourly  float32 `json:"hourly"`
	Monthly float32 `json:"monthly"`
}

// LinodeBackupsAddon represents a linode backups addon object
type LinodeBackupsAddon struct {
	Price *LinodePrice `json:"price"`
}

// LinodeAddons represent the linode addons object
type LinodeAddons struct {
	Backups *LinodeBackupsAddon `json:"backups"`
}

// LinodeTypeClass constants start with Class and include Linode API Instance Type Classes
type LinodeTypeClass string

// LinodeTypeClass contants are the Instance Type Classes that an Instance Type can be assigned
const (
	ClassNanode    LinodeTypeClass = "nanode"
	ClassStandard  LinodeTypeClass = "standard"
	ClassHighmem   LinodeTypeClass = "highmem"
	ClassDedicated LinodeTypeClass = "dedicated"
)

// LinodeTypesPagedResponse represents a linode types API response for listing
type LinodeTypesPagedResponse struct {
	*PageOptions
	Data []LinodeType `json:"data"`
}

func (LinodeTypesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Types.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

func (resp *LinodeTypesPagedResponse) appendData(r *LinodeTypesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListTypes lists linode types
func (c *Client) ListTypes(ctx context.Context, opts *ListOptions) ([]LinodeType, error) {
	response := LinodeTypesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetType gets the type with the provided ID
func (c *Client) GetType(ctx context.Context, typeID string) (*LinodeType, error) {
	e, err := c.Types.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, typeID)

	r, err := coupleAPIErrors(c.Types.R(ctx).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*LinodeType), nil
}
