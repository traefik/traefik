package govultr

import (
	"context"
	"net/http"
	"net/url"
)

// UserService is the interface to interact with the user management endpoints on the Vultr API
// Link: https://www.vultr.com/api/#user
type UserService interface {
	Create(ctx context.Context, email, name, password, apiEnabled string, acls []string) (*User, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, user *User) error
}

// UserServiceHandler handles interaction with the user methods for the Vultr API
type UserServiceHandler struct {
	client *Client
}

// User represents an user on Vultr
type User struct {
	UserID     string   `json:"USERID"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Password   string   `json:"password"`
	APIEnabled string   `json:"api_enabled"`
	ACL        []string `json:"acls"`
	APIKey     string   `json:"api_key"`
}

// Create will add the specified user to your Vultr account
func (u *UserServiceHandler) Create(ctx context.Context, email, name, password, apiEnabled string, acls []string) (*User, error) {

	uri := "/v1/user/create"

	values := url.Values{
		"email":    {email},
		"name":     {name},
		"password": {password},
		"acls[]":   acls,
	}

	if apiEnabled != "" {
		values.Add("api_enabled", apiEnabled)
	}

	req, err := u.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	user := new(User)

	err = u.client.DoWithContext(ctx, req, user)

	if err != nil {
		return nil, err
	}

	user.Name = name
	user.Email = email
	user.APIEnabled = apiEnabled
	user.ACL = acls

	return user, nil
}

// Delete will remove the specified user from your Vultr account
func (u *UserServiceHandler) Delete(ctx context.Context, userID string) error {

	uri := "/v1/user/delete"

	values := url.Values{
		"USERID": {userID},
	}

	req, err := u.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = u.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List will list all the users associated with your Vultr account
func (u *UserServiceHandler) List(ctx context.Context) ([]User, error) {

	uri := "/v1/user/list"

	req, err := u.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var users []User
	err = u.client.DoWithContext(ctx, req, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Update will update the given user. Empty strings will be ignored.
func (u *UserServiceHandler) Update(ctx context.Context, user *User) error {

	uri := "/v1/user/update"

	values := url.Values{
		"USERID": {user.UserID},
	}

	// Optional
	if user.Email != "" {
		values.Add("email", user.Email)
	}
	if user.Name != "" {
		values.Add("name", user.Name)
	}
	if user.Password != "" {
		values.Add("password", user.Password)
	}
	if user.APIEnabled != "" {
		values.Add("api_enabled", user.APIEnabled)
	}
	if len(user.ACL) > 0 {
		for _, acl := range user.ACL {
			values.Add("acls[]", acl)
		}
	}

	req, err := u.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = u.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
