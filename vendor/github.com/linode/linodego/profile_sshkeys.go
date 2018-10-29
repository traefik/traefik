package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SSHKey represents a SSHKey object
type SSHKey struct {
	ID         int        `json:"id"`
	Label      string     `json:"label"`
	SSHKey     string     `json:"ssh_key"`
	CreatedStr string     `json:"created"`
	Created    *time.Time `json:"-"`
}

// SSHKeyCreateOptions fields are those accepted by CreateSSHKey
type SSHKeyCreateOptions struct {
	Label  string `json:"label"`
	SSHKey string `json:"ssh_key"`
}

// SSHKeyUpdateOptions fields are those accepted by UpdateSSHKey
type SSHKeyUpdateOptions struct {
	Label string `json:"label"`
}

// GetCreateOptions converts a SSHKey to SSHKeyCreateOptions for use in CreateSSHKey
func (i SSHKey) GetCreateOptions() (o SSHKeyCreateOptions) {
	o.Label = i.Label
	o.SSHKey = i.SSHKey
	return
}

// GetUpdateOptions converts a SSHKey to SSHKeyCreateOptions for use in UpdateSSHKey
func (i SSHKey) GetUpdateOptions() (o SSHKeyUpdateOptions) {
	o.Label = i.Label
	return
}

// SSHKeysPagedResponse represents a paginated SSHKey API response
type SSHKeysPagedResponse struct {
	*PageOptions
	Data []SSHKey `json:"data"`
}

// endpoint gets the endpoint URL for SSHKey
func (SSHKeysPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.SSHKeys.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends SSHKeys when processing paginated SSHKey responses
func (resp *SSHKeysPagedResponse) appendData(r *SSHKeysPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListSSHKeys lists SSHKeys
func (c *Client) ListSSHKeys(ctx context.Context, opts *ListOptions) ([]SSHKey, error) {
	response := SSHKeysPagedResponse{}
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
func (i *SSHKey) fixDates() *SSHKey {
	i.Created, _ = parseDates(i.CreatedStr)
	return i
}

// GetSSHKey gets the sshkey with the provided ID
func (c *Client) GetSSHKey(ctx context.Context, id int) (*SSHKey, error) {
	e, err := c.SSHKeys.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&SSHKey{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*SSHKey).fixDates(), nil
}

// CreateSSHKey creates a SSHKey
func (c *Client) CreateSSHKey(ctx context.Context, createOpts SSHKeyCreateOptions) (*SSHKey, error) {
	var body string
	e, err := c.SSHKeys.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&SSHKey{})

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
	return r.Result().(*SSHKey).fixDates(), nil
}

// UpdateSSHKey updates the SSHKey with the specified id
func (c *Client) UpdateSSHKey(ctx context.Context, id int, updateOpts SSHKeyUpdateOptions) (*SSHKey, error) {
	var body string
	e, err := c.SSHKeys.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&SSHKey{})

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
	return r.Result().(*SSHKey).fixDates(), nil
}

// DeleteSSHKey deletes the SSHKey with the specified id
func (c *Client) DeleteSSHKey(ctx context.Context, id int) error {
	e, err := c.SSHKeys.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err

}
