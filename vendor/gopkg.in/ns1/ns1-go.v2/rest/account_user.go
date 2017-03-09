package rest

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/account"
)

// UsersService handles 'account/users' endpoint.
type UsersService service

// List returns all users in the account.
//
// NS1 API docs: https://ns1.com/api/#users-get
func (s *UsersService) List() ([]*account.User, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "account/users", nil)
	if err != nil {
		return nil, nil, err
	}

	ul := []*account.User{}
	resp, err := s.client.Do(req, &ul)
	if err != nil {
		return nil, resp, err
	}

	return ul, resp, nil
}

// Get returns details of a single user.
//
// NS1 API docs: https://ns1.com/api/#users-user-get
func (s *UsersService) Get(username string) (*account.User, *http.Response, error) {
	path := fmt.Sprintf("account/users/%s", username)

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var u account.User
	resp, err := s.client.Do(req, &u)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "Unknown user" {
				return nil, resp, ErrUserMissing
			}
		default:
			return nil, resp, err
		}
	}

	return &u, resp, nil
}

// Create takes a *User and creates a new account user.
//
// NS1 API docs: https://ns1.com/api/#users-put
func (s *UsersService) Create(u *account.User) (*http.Response, error) {
	req, err := s.client.NewRequest("PUT", "account/users", &u)
	if err != nil {
		return nil, err
	}

	// Update user fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &u)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "request failed:Login Name is already in use." {
				return resp, ErrUserExists
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

// Update change contact details, notification settings, or access rights for a user.
//
// NS1 API docs: https://ns1.com/api/#users-user-post
func (s *UsersService) Update(u *account.User) (*http.Response, error) {
	path := fmt.Sprintf("account/users/%s", u.Username)

	req, err := s.client.NewRequest("POST", path, &u)
	if err != nil {
		return nil, err
	}

	// Update user fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &u)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "Unknown user" {
				return resp, ErrUserMissing
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

// Delete deletes a user.
//
// NS1 API docs: https://ns1.com/api/#users-user-delete
func (s *UsersService) Delete(username string) (*http.Response, error) {
	path := fmt.Sprintf("account/users/%s", username)

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "Unknown user" {
				return resp, ErrUserMissing
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

var (
	// ErrUserExists bundles PUT create error.
	ErrUserExists = errors.New("User already exists.")
	// ErrUserMissing bundles GET/POST/DELETE error.
	ErrUserMissing = errors.New("User does not exist.")
)
