package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"time"
)

// RunstatusMaintenance is a runstatus maintenance
type RunstatusMaintenance struct {
	Created     *time.Time       `json:"created,omitempty"`
	Description string           `json:"description,omitempty"`
	EndDate     *time.Time       `json:"end_date"`
	Events      []RunstatusEvent `json:"events,omitempty"`
	EventsURL   string           `json:"events_url,omitempty"`
	ID          int              `json:"id,omitempty"`       // missing field
	PageURL     string           `json:"page_url,omitempty"` // fake field
	RealTime    bool             `json:"real_time,omitempty"`
	Services    []string         `json:"services"`
	StartDate   *time.Time       `json:"start_date"`
	Status      string           `json:"status"`
	Title       string           `json:"title"`
	URL         string           `json:"url,omitempty"`
}

// Match returns true if the other maintenance has got similarities with itself
func (maintenance RunstatusMaintenance) Match(other RunstatusMaintenance) bool {
	if other.Title != "" && maintenance.Title == other.Title {
		return true
	}

	if other.ID > 0 && maintenance.ID == other.ID {
		return true
	}

	return false
}

// FakeID fills up the ID field as it's currently missing
func (maintenance *RunstatusMaintenance) FakeID() error {
	if maintenance.ID > 0 {
		return nil
	}

	if maintenance.URL == "" {
		return fmt.Errorf("empty URL for %#v", maintenance)
	}

	u, err := url.Parse(maintenance.URL)
	if err != nil {
		return err
	}

	s := path.Base(u.Path)
	id, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	maintenance.ID = id
	return nil
}

// RunstatusMaintenanceList is a list of incident
type RunstatusMaintenanceList struct {
	Next         string                 `json:"next"`
	Previous     string                 `json:"previous"`
	Maintenances []RunstatusMaintenance `json:"results"`
}

// GetRunstatusMaintenance retrieves the details of a specific maintenance.
func (client *Client) GetRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance) (*RunstatusMaintenance, error) {
	if maintenance.URL != "" {
		return client.getRunstatusMaintenance(ctx, maintenance.URL)
	}

	if maintenance.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", maintenance)
	}

	page, err := client.getRunstatusPage(ctx, maintenance.PageURL)
	if err != nil {
		return nil, err
	}

	for i := range page.Maintenances {
		m := &page.Maintenances[i]
		if m.Match(maintenance) {
			if err := m.FakeID(); err != nil {
				log.Printf("bad fake ID for %#v, %s", m, err)
			}
			return m, nil
		}
	}

	return nil, fmt.Errorf("%#v not found", maintenance)
}

func (client *Client) getRunstatusMaintenance(ctx context.Context, maintenanceURL string) (*RunstatusMaintenance, error) {
	resp, err := client.runstatusRequest(ctx, maintenanceURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	m := new(RunstatusMaintenance)
	if err := json.Unmarshal(resp, m); err != nil {
		return nil, err
	}
	return m, nil
}

// ListRunstatusMaintenances returns the list of maintenances for the page.
func (client *Client) ListRunstatusMaintenances(ctx context.Context, page RunstatusPage) ([]RunstatusMaintenance, error) {
	if page.MaintenancesURL == "" {
		return nil, fmt.Errorf("empty Maintenances URL for %#v", page)
	}

	results := make([]RunstatusMaintenance, 0)

	var err error
	client.PaginateRunstatusMaintenances(ctx, page, func(maintenance *RunstatusMaintenance, e error) bool {
		if e != nil {
			err = e
			return false
		}

		results = append(results, *maintenance)
		return true
	})

	return results, err
}

// PaginateRunstatusMaintenances paginate Maintenances
func (client *Client) PaginateRunstatusMaintenances(ctx context.Context, page RunstatusPage, callback func(*RunstatusMaintenance, error) bool) { // nolint: dupl
	if page.MaintenancesURL == "" {
		callback(nil, fmt.Errorf("empty Maintenances URL for %#v", page))
		return
	}

	maintenancesURL := page.MaintenancesURL
	for maintenancesURL != "" {
		resp, err := client.runstatusRequest(ctx, maintenancesURL, nil, "GET")
		if err != nil {
			callback(nil, err)
			return
		}

		var ms *RunstatusMaintenanceList
		if err := json.Unmarshal(resp, &ms); err != nil {
			callback(nil, err)
			return
		}

		for i := range ms.Maintenances {
			if err := ms.Maintenances[i].FakeID(); err != nil {
				log.Printf("bad fake ID for %#v, %s", ms.Maintenances[i], err)
			}
			if cont := callback(&ms.Maintenances[i], nil); !cont {
				return
			}
		}

		maintenancesURL = ms.Next
	}
}

// CreateRunstatusMaintenance create runstatus Maintenance
func (client *Client) CreateRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance) (*RunstatusMaintenance, error) {
	if maintenance.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", maintenance)
	}

	page, err := client.getRunstatusPage(ctx, maintenance.PageURL)
	if err != nil {
		return nil, err
	}

	resp, err := client.runstatusRequest(ctx, page.MaintenancesURL, maintenance, "POST")
	if err != nil {
		return nil, err
	}

	m := &RunstatusMaintenance{}
	if err := json.Unmarshal(resp, &m); err != nil {
		return nil, err
	}

	if err := m.FakeID(); err != nil {
		log.Printf("bad fake ID for %#v, %s", m, err)
	}

	return m, nil
}

// DeleteRunstatusMaintenance delete runstatus Maintenance
func (client *Client) DeleteRunstatusMaintenance(ctx context.Context, maintenance RunstatusMaintenance) error {
	if maintenance.URL == "" {
		return fmt.Errorf("empty URL for %#v", maintenance)
	}

	_, err := client.runstatusRequest(ctx, maintenance.URL, nil, "DELETE")
	return err
}
