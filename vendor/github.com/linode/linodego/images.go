package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Image represents a deployable Image object for use with Linode Instances
type Image struct {
	CreatedStr  string `json:"created"`
	UpdatedStr  string `json:"updated"`
	ID          string `json:"id"`
	CreatedBy   string `json:"created_by"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Vendor      string `json:"vendor"`
	Size        int    `json:"size"`
	IsPublic    bool   `json:"is_public"`
	Deprecated  bool   `json:"deprecated"`

	Created *time.Time `json:"-"`
	Updated *time.Time `json:"-"`
}

// ImageCreateOptions fields are those accepted by CreateImage
type ImageCreateOptions struct {
	DiskID      int    `json:"disk_id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// ImageUpdateOptions fields are those accepted by UpdateImage
type ImageUpdateOptions struct {
	Label       string  `json:"label,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (i *Image) fixDates() *Image {
	i.Created, _ = parseDates(i.CreatedStr)
	i.Updated, _ = parseDates(i.UpdatedStr)
	return i
}

// GetUpdateOptions converts an Image to ImageUpdateOptions for use in UpdateImage
func (i Image) GetUpdateOptions() (iu ImageUpdateOptions) {
	iu.Label = i.Label
	iu.Description = copyString(&i.Description)
	return
}

// ImagesPagedResponse represents a linode API response for listing of images
type ImagesPagedResponse struct {
	*PageOptions
	Data []Image `json:"data"`
}

func (ImagesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Images.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

func (resp *ImagesPagedResponse) appendData(r *ImagesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListImages lists Images
func (c *Client) ListImages(ctx context.Context, opts *ListOptions) ([]Image, error) {
	response := ImagesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	for i := range response.Data {
		response.Data[i].fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil

}

// GetImage gets the Image with the provided ID
func (c *Client) GetImage(ctx context.Context, id string) (*Image, error) {
	e, err := c.Images.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)
	r, err := coupleAPIErrors(c.Images.R(ctx).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Image), nil
}

// CreateImage creates a Image
func (c *Client) CreateImage(ctx context.Context, createOpts ImageCreateOptions) (*Image, error) {
	var body string
	e, err := c.Images.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Image{})

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
	return r.Result().(*Image).fixDates(), nil
}

// UpdateImage updates the Image with the specified id
func (c *Client) UpdateImage(ctx context.Context, id string, updateOpts ImageUpdateOptions) (*Image, error) {
	var body string
	e, err := c.Images.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)

	req := c.R(ctx).SetResult(&Image{})

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
	return r.Result().(*Image).fixDates(), nil
}

// DeleteImage deletes the Image with the specified id
func (c *Client) DeleteImage(ctx context.Context, id string) error {
	e, err := c.Images.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%s", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
