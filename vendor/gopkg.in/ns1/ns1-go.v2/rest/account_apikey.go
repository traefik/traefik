package rest

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/account"
)

// APIKeysService handles 'account/apikeys' endpoint.
type APIKeysService service

// List returns all api keys in the account.
//
// NS1 API docs: https://ns1.com/api/#apikeys-get
func (s *APIKeysService) List() ([]*account.APIKey, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "account/apikeys", nil)
	if err != nil {
		return nil, nil, err
	}

	kl := []*account.APIKey{}
	resp, err := s.client.Do(req, &kl)
	if err != nil {
		return nil, resp, err
	}

	return kl, resp, nil
}

// Get returns details of an api key, including permissions, for a single API Key.
// Note: do not use the API Key itself as the keyid in the URL — use the id of the key.
//
// NS1 API docs: https://ns1.com/api/#apikeys-id-get
func (s *APIKeysService) Get(keyID string) (*account.APIKey, *http.Response, error) {
	path := fmt.Sprintf("account/apikeys/%s", keyID)

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var a account.APIKey
	resp, err := s.client.Do(req, &a)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "unknown api key" {
				return nil, resp, ErrKeyMissing
			}

		}
		return nil, resp, err
	}

	return &a, resp, nil
}

// Create takes a *APIKey and creates a new account apikey.
//
// NS1 API docs: https://ns1.com/api/#apikeys-put
func (s *APIKeysService) Create(a *account.APIKey) (*http.Response, error) {
	req, err := s.client.NewRequest("PUT", "account/apikeys", &a)
	if err != nil {
		return nil, err
	}

	// Update account fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &a)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == fmt.Sprintf("api key with name \"%s\" exists", a.Name) {
				return resp, ErrKeyExists
			}
		}
		return resp, err
	}

	return resp, nil
}

// Update changes the name or access rights for an API Key.
//
// NS1 API docs: https://ns1.com/api/#apikeys-id-post
func (s *APIKeysService) Update(a *account.APIKey) (*http.Response, error) {
	path := fmt.Sprintf("account/apikeys/%s", a.ID)

	req, err := s.client.NewRequest("POST", path, &a)
	if err != nil {
		return nil, err
	}

	// Update apikey fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &a)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "unknown api key" {
				return resp, ErrKeyMissing
			}
		}
		return resp, err
	}

	return resp, nil
}

// Delete deletes an apikey.
//
// NS1 API docs: https://ns1.com/api/#apikeys-id-delete
func (s *APIKeysService) Delete(keyID string) (*http.Response, error) {
	path := fmt.Sprintf("account/apikeys/%s", keyID)

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "unknown api key" {
				return resp, ErrKeyMissing
			}
		}
		return resp, err
	}

	return resp, nil
}

var (
	// ErrKeyExists bundles PUT create error.
	ErrKeyExists = errors.New("key already exists")
	// ErrKeyMissing bundles GET/POST/DELETE error.
	ErrKeyMissing = errors.New("key does not exist")
)
