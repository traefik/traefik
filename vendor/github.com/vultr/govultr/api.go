package govultr

import (
	"context"
	"net/http"
)

// APIService is the interface to interact with the API endpoint on the Vultr API
// Link: https://www.vultr.com/api/#auth
type APIService interface {
	GetInfo(ctx context.Context) (*API, error)
}

// APIServiceHandler handles interaction with the API methods for the Vultr API
type APIServiceHandler struct {
	client *Client
}

// API represents Vultr API information
type API struct {
	ACL   []string `json:"acls"`
	Email string   `json:"email"`
	Name  string   `json:"name"`
}

// GetInfo Vultr API auth information
func (a *APIServiceHandler) GetInfo(ctx context.Context) (*API, error) {
	uri := "/v1/auth/info"

	req, err := a.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	api := new(API)
	err = a.client.DoWithContext(ctx, req, api)

	if err != nil {
		return nil, err
	}

	return api, nil
}
