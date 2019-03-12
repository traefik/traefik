package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
)

// RunstatusService is a runstatus service
type RunstatusService struct {
	ID      int    `json:"id"` // missing field
	Name    string `json:"name"`
	PageURL string `json:"page_url,omitempty"` // fake field
	State   string `json:"state,omitempty"`
	URL     string `json:"url,omitempty"`
}

// FakeID fills up the ID field as it's currently missing
func (service *RunstatusService) FakeID() error {
	if service.ID > 0 {
		return nil
	}

	if service.URL == "" {
		return fmt.Errorf("empty URL for %#v", service)
	}

	u, err := url.Parse(service.URL)
	if err != nil {
		return err
	}

	s := path.Base(u.Path)
	id, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	service.ID = id
	return nil
}

// Match returns true if the other service has got similarities with itself
func (service RunstatusService) Match(other RunstatusService) bool {
	if other.Name != "" && service.Name == other.Name {
		return true
	}

	if other.ID > 0 && service.ID == other.ID {
		return true
	}

	return false
}

// RunstatusServiceList service list
type RunstatusServiceList struct {
	Next     string             `json:"next"`
	Previous string             `json:"previous"`
	Services []RunstatusService `json:"results"`
}

// DeleteRunstatusService delete runstatus service
func (client *Client) DeleteRunstatusService(ctx context.Context, service RunstatusService) error {
	if service.URL == "" {
		return fmt.Errorf("empty URL for %#v", service)
	}

	_, err := client.runstatusRequest(ctx, service.URL, nil, "DELETE")
	return err
}

// CreateRunstatusService create runstatus service
func (client *Client) CreateRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", service)
	}

	page, err := client.GetRunstatusPage(ctx, RunstatusPage{URL: service.PageURL})
	if err != nil {
		return nil, err
	}

	resp, err := client.runstatusRequest(ctx, page.ServicesURL, service, "POST")
	if err != nil {
		return nil, err
	}

	s := &RunstatusService{}
	if err := json.Unmarshal(resp, s); err != nil {
		return nil, err
	}

	return s, nil
}

// GetRunstatusService displays service detail.
func (client *Client) GetRunstatusService(ctx context.Context, service RunstatusService) (*RunstatusService, error) {
	if service.URL != "" {
		return client.getRunstatusService(ctx, service.URL)
	}

	if service.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL in %#v", service)
	}

	page, err := client.getRunstatusPage(ctx, service.PageURL)
	if err != nil {
		return nil, err
	}

	for i := range page.Services {
		s := &page.Services[i]
		if s.Match(service) {
			if err := s.FakeID(); err != nil {
				log.Printf("bad fake ID for %#v, %s", s, err)
			}
			return s, nil
		}
	}

	return nil, fmt.Errorf("%#v not found", service)
}

func (client *Client) getRunstatusService(ctx context.Context, serviceURL string) (*RunstatusService, error) {
	resp, err := client.runstatusRequest(ctx, serviceURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	s := &RunstatusService{}
	if err := json.Unmarshal(resp, &s); err != nil {
		return nil, err
	}

	if err := s.FakeID(); err != nil {
		log.Printf("bad fake ID for %#v, %s", s, err)
	}

	return s, nil
}

// ListRunstatusServices displays the list of services.
func (client *Client) ListRunstatusServices(ctx context.Context, page RunstatusPage) ([]RunstatusService, error) {
	if page.ServicesURL == "" {
		return nil, fmt.Errorf("empty Services URL for %#v", page)
	}

	results := make([]RunstatusService, 0)

	var err error
	client.PaginateRunstatusServices(ctx, page, func(service *RunstatusService, e error) bool {
		if e != nil {
			err = e
			return false
		}

		results = append(results, *service)
		return true
	})

	return results, err
}

// PaginateRunstatusServices paginates Services
func (client *Client) PaginateRunstatusServices(ctx context.Context, page RunstatusPage, callback func(*RunstatusService, error) bool) { // nolint: dupl
	if page.ServicesURL == "" {
		callback(nil, fmt.Errorf("empty Services URL for %#v", page))
		return
	}

	servicesURL := page.ServicesURL
	for servicesURL != "" {
		resp, err := client.runstatusRequest(ctx, servicesURL, nil, "GET")
		if err != nil {
			callback(nil, err)
			return
		}

		var ss *RunstatusServiceList
		if err := json.Unmarshal(resp, &ss); err != nil {
			callback(nil, err)
			return
		}

		for i := range ss.Services {
			if err := ss.Services[i].FakeID(); err != nil {
				log.Printf("bad fake ID for %#v, %s", ss.Services[i], err)
			}

			if cont := callback(&ss.Services[i], nil); !cont {
				return
			}
		}

		servicesURL = ss.Next
	}
}
