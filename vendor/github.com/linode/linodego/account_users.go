package linodego

import (
	"context"
	"encoding/json"
	"fmt"
)

// User represents a User object
type User struct {
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Restricted bool     `json:"restricted"`
	SSHKeys    []string `json:"ssh_keys"`
}

// UserCreateOptions fields are those accepted by CreateUser
type UserCreateOptions struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Restricted bool   `json:"restricted,omitempty"`
}

// UserUpdateOptions fields are those accepted by UpdateUser
type UserUpdateOptions struct {
	Username   string    `json:"username,omitempty"`
	Email      string    `json:"email,omitempty"`
	Restricted *bool     `json:"restricted,omitempty"`
	SSHKeys    *[]string `json:"ssh_keys,omitempty"`
}

// GetCreateOptions converts a User to UserCreateOptions for use in CreateUser
func (i User) GetCreateOptions() (o UserCreateOptions) {
	o.Username = i.Username
	o.Email = i.Email
	o.Restricted = i.Restricted
	return
}

// GetUpdateOptions converts a User to UserUpdateOptions for use in UpdateUser
func (i User) GetUpdateOptions() (o UserUpdateOptions) {
	o.Username = i.Username
	o.Email = i.Email
	o.Restricted = copyBool(&i.Restricted)
	return
}

// UsersPagedResponse represents a paginated User API response
type UsersPagedResponse struct {
	*PageOptions
	Data []User `json:"data"`
}

// endpoint gets the endpoint URL for User
func (UsersPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Users.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Users when processing paginated User responses
func (resp *UsersPagedResponse) appendData(r *UsersPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListUsers lists Users on the account
func (c *Client) ListUsers(ctx context.Context, opts *ListOptions) ([]User, error) {
	response := UsersPagedResponse{}
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
func (i *User) fixDates() *User {
	return i
}

// GetUser gets the user with the provided ID
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	e, err := c.Users.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&User{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*User).fixDates(), nil
}

// CreateUser creates a User.  The email address must be confirmed before the
// User account can be accessed.
func (c *Client) CreateUser(ctx context.Context, createOpts UserCreateOptions) (*User, error) {
	var body string
	e, err := c.Users.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&User{})

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
	return r.Result().(*User).fixDates(), nil
}

// UpdateUser updates the User with the specified id
func (c *Client) UpdateUser(ctx context.Context, id string, updateOpts UserUpdateOptions) (*User, error) {
	var body string
	e, err := c.Users.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)

	req := c.R(ctx).SetResult(&User{})

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
	return r.Result().(*User).fixDates(), nil
}

// DeleteUser deletes the User with the specified id
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	e, err := c.Users.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%s", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
