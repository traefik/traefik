package rest

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// ZonesService handles 'zones' endpoint.
type ZonesService service

// List returns all active zones and basic zone configuration details for each.
//
// NS1 API docs: https://ns1.com/api/#zones-get
func (s *ZonesService) List() ([]*dns.Zone, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "zones", nil)
	if err != nil {
		return nil, nil, err
	}

	zl := []*dns.Zone{}
	resp, err := s.client.Do(req, &zl)
	if err != nil {
		return nil, resp, err
	}

	return zl, resp, nil
}

// Get takes a zone name and returns a single active zone and its basic configuration details.
//
// NS1 API docs: https://ns1.com/api/#zones-zone-get
func (s *ZonesService) Get(zone string) (*dns.Zone, *http.Response, error) {
	path := fmt.Sprintf("zones/%s", zone)

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var z dns.Zone
	resp, err := s.client.Do(req, &z)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "zone not found" {
				return nil, resp, ErrZoneMissing
			}
		default:
			return nil, resp, err
		}
	}

	return &z, resp, nil
}

// Create takes a *Zone and creates a new DNS zone.
//
// NS1 API docs: https://ns1.com/api/#zones-put
func (s *ZonesService) Create(z *dns.Zone) (*http.Response, error) {
	path := fmt.Sprintf("zones/%s", z.Zone)

	req, err := s.client.NewRequest("PUT", path, &z)
	if err != nil {
		return nil, err
	}

	// Update zones fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &z)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "zone already exists" {
				return resp, ErrZoneExists
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

// Update takes a *Zone and modifies basic details of a DNS zone.
//
// NS1 API docs: https://ns1.com/api/#zones-post
func (s *ZonesService) Update(z *dns.Zone) (*http.Response, error) {
	path := fmt.Sprintf("zones/%s", z.Zone)

	req, err := s.client.NewRequest("POST", path, &z)
	if err != nil {
		return nil, err
	}

	// Update zones fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &z)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "zone not found" {
				return resp, ErrZoneMissing
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

// Delete takes a zone and destroys an existing DNS zone and all records in the zone.
//
// NS1 API docs: https://ns1.com/api/#zones-delete
func (s *ZonesService) Delete(zone string) (*http.Response, error) {
	path := fmt.Sprintf("zones/%s", zone)

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		switch err.(type) {
		case *Error:
			if err.(*Error).Message == "zone not found" {
				return resp, ErrZoneMissing
			}
		default:
			return resp, err
		}
	}

	return resp, nil
}

var (
	// ErrZoneExists bundles PUT create error.
	ErrZoneExists = errors.New("Zone already exists.")
	// ErrZoneMissing bundles GET/POST/DELETE error.
	ErrZoneMissing = errors.New("Zone does not exist.")
)
