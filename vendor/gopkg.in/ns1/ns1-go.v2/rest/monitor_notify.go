package rest

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/monitor"
)

// NotificationsService handles 'monitoring/lists' endpoint.
type NotificationsService service

// List returns all configured notification lists.
//
// NS1 API docs: https://ns1.com/api/#lists-get
func (s *NotificationsService) List() ([]*monitor.NotifyList, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "lists", nil)
	if err != nil {
		return nil, nil, err
	}

	nl := []*monitor.NotifyList{}
	resp, err := s.client.Do(req, &nl)
	if err != nil {
		return nil, resp, err
	}

	return nl, resp, nil
}

// Get returns the details and notifiers associated with a specific notification list.
//
// NS1 API docs: https://ns1.com/api/#lists-listid-get
func (s *NotificationsService) Get(listID string) (*monitor.NotifyList, *http.Response, error) {
	path := fmt.Sprintf("%s/%s", "lists", listID)

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var nl monitor.NotifyList
	resp, err := s.client.Do(req, &nl)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "unknown notification list" {
				return nil, resp, ErrListMissing
			}
		}
		return nil, resp, err
	}

	return &nl, resp, nil
}

// Create takes a *NotifyList and creates a new notify list.
//
// NS1 API docs: https://ns1.com/api/#lists-put
func (s *NotificationsService) Create(nl *monitor.NotifyList) (*http.Response, error) {
	req, err := s.client.NewRequest("PUT", "lists", &nl)
	if err != nil {
		return nil, err
	}

	// Update notify list fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &nl)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == fmt.Sprintf("notification list with name \"%s\" exists", nl.Name) {
				return resp, ErrListExists
			}
		}
		return resp, err
	}

	return resp, nil
}

// Update adds or removes entries or otherwise update a notification list.
//
// NS1 API docs: https://ns1.com/api/#list-listid-post
func (s *NotificationsService) Update(nl *monitor.NotifyList) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s", "lists", nl.ID)

	req, err := s.client.NewRequest("POST", path, &nl)
	if err != nil {
		return nil, err
	}

	// Update mon lists' fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &nl)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Delete immediately deletes an existing notification list.
//
// NS1 API docs: https://ns1.com/api/#lists-listid-delete
func (s *NotificationsService) Delete(listID string) (*http.Response, error) {
	path := fmt.Sprintf("%s/%s", "lists", listID)

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

var (
	// ErrListExists bundles PUT create error.
	ErrListExists = errors.New("notify List already exists")
	// ErrListMissing bundles GET/POST/DELETE error.
	ErrListMissing = errors.New("notify List does not exist")
)
