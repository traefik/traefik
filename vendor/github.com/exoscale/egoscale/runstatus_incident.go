package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

//RunstatusIncident is a runstatus incident
type RunstatusIncident struct {
	EndDate    *time.Time       `json:"end_date,omitempty"`
	Events     []RunstatusEvent `json:"events,omitempty"`
	EventsURL  string           `json:"events_url,omitempty"`
	ID         int              `json:"id,omitempty"`
	PageURL    string           `json:"page_url,omitempty"` // fake field
	PostMortem string           `json:"post_mortem,omitempty"`
	RealTime   bool             `json:"real_time,omitempty"`
	Services   []string         `json:"services"`
	StartDate  *time.Time       `json:"start_date,omitempty"`
	State      string           `json:"state"`
	Status     string           `json:"status"`
	StatusText string           `json:"status_text"`
	Title      string           `json:"title"`
	URL        string           `json:"url,omitempty"`
}

// Match returns true if the other incident has got similarities with itself
func (incident RunstatusIncident) Match(other RunstatusIncident) bool {
	if other.Title != "" && incident.Title == other.Title {
		return true
	}

	if other.ID > 0 && incident.ID == other.ID {
		return true
	}

	return false
}

//RunstatusIncidentList is a list of incident
type RunstatusIncidentList struct {
	Next      string              `json:"next"`
	Previous  string              `json:"previous"`
	Incidents []RunstatusIncident `json:"results"`
}

// GetRunstatusIncident retrieves the details of a specific incident.
func (client *Client) GetRunstatusIncident(ctx context.Context, incident RunstatusIncident) (*RunstatusIncident, error) {
	if incident.URL != "" {
		return client.getRunstatusIncident(ctx, incident.URL)
	}

	if incident.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", incident)
	}

	page, err := client.getRunstatusPage(ctx, incident.PageURL)
	if err != nil {
		return nil, err
	}

	for i := range page.Incidents {
		j := &page.Incidents[i]
		if j.Match(incident) {
			return j, nil
		}
	}

	return nil, fmt.Errorf("%#v not found", incident)
}

func (client *Client) getRunstatusIncident(ctx context.Context, incidentURL string) (*RunstatusIncident, error) {
	resp, err := client.runstatusRequest(ctx, incidentURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	i := new(RunstatusIncident)
	if err := json.Unmarshal(resp, i); err != nil {
		return nil, err
	}
	return i, nil
}

// ListRunstatusIncidents lists the incidents for a specific page.
func (client *Client) ListRunstatusIncidents(ctx context.Context, page RunstatusPage) ([]RunstatusIncident, error) {
	if page.IncidentsURL == "" {
		return nil, fmt.Errorf("empty Incidents URL for %#v", page)
	}

	results := make([]RunstatusIncident, 0)

	var err error
	client.PaginateRunstatusIncidents(ctx, page, func(incident *RunstatusIncident, e error) bool {
		if e != nil {
			err = e
			return false
		}

		results = append(results, *incident)
		return true
	})

	return results, err
}

// PaginateRunstatusIncidents paginate Incidents
func (client *Client) PaginateRunstatusIncidents(ctx context.Context, page RunstatusPage, callback func(*RunstatusIncident, error) bool) {
	if page.IncidentsURL == "" {
		callback(nil, fmt.Errorf("empty Incidents URL for %#v", page))
		return
	}

	incidentsURL := page.IncidentsURL
	for incidentsURL != "" {
		resp, err := client.runstatusRequest(ctx, incidentsURL, nil, "GET")
		if err != nil {
			callback(nil, err)
			return
		}

		var is *RunstatusIncidentList
		if err := json.Unmarshal(resp, &is); err != nil {
			callback(nil, err)
			return
		}

		for i := range is.Incidents {
			if cont := callback(&is.Incidents[i], nil); !cont {
				return
			}
		}

		incidentsURL = is.Next
	}
}

// CreateRunstatusIncident create runstatus incident
func (client *Client) CreateRunstatusIncident(ctx context.Context, incident RunstatusIncident) (*RunstatusIncident, error) {
	if incident.PageURL == "" {
		return nil, fmt.Errorf("empty Page URL for %#v", incident)
	}

	page, err := client.getRunstatusPage(ctx, incident.PageURL)
	if err != nil {
		return nil, err
	}

	if page.IncidentsURL == "" {
		return nil, fmt.Errorf("empty Incidents URL for %#v", page)
	}

	resp, err := client.runstatusRequest(ctx, page.IncidentsURL, incident, "POST")
	if err != nil {
		return nil, err
	}

	i := &RunstatusIncident{}
	if err := json.Unmarshal(resp, &i); err != nil {
		return nil, err
	}

	return i, nil
}

// DeleteRunstatusIncident delete runstatus incident
func (client *Client) DeleteRunstatusIncident(ctx context.Context, incident RunstatusIncident) error {
	if incident.URL == "" {
		return fmt.Errorf("empty URL for %#v", incident)
	}

	_, err := client.runstatusRequest(ctx, incident.URL, nil, "DELETE")
	return err
}
