package egoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// RunstatusPage runstatus page
type RunstatusPage struct {
	Created          *time.Time             `json:"created,omitempty"`
	DarkTheme        bool                   `json:"dark_theme,omitempty"`
	Domain           string                 `json:"domain,omitempty"`
	GradientEnd      string                 `json:"gradient_end,omitempty"`
	GradientStart    string                 `json:"gradient_start,omitempty"`
	HeaderBackground string                 `json:"header_background,omitempty"`
	ID               int                    `json:"id,omitempty"`
	Incidents        []RunstatusIncident    `json:"incidents,omitempty"`
	IncidentsURL     string                 `json:"incidents_url,omitempty"`
	Logo             string                 `json:"logo,omitempty"`
	Maintenances     []RunstatusMaintenance `json:"maintenances,omitempty"`
	MaintenancesURL  string                 `json:"maintenances_url,omitempty"`
	Name             string                 `json:"name"` //fake field (used to post a new runstatus page)
	OkText           string                 `json:"ok_text,omitempty"`
	Plan             string                 `json:"plan,omitempty"`
	PublicURL        string                 `json:"public_url,omitempty"`
	Services         []RunstatusService     `json:"services,omitempty"`
	ServicesURL      string                 `json:"services_url,omitempty"`
	State            string                 `json:"state,omitempty"`
	Subdomain        string                 `json:"subdomain"`
	SupportEmail     string                 `json:"support_email,omitempty"`
	TimeZone         string                 `json:"time_zone,omitempty"`
	Title            string                 `json:"title,omitempty"`
	TitleColor       string                 `json:"title_color,omitempty"`
	TwitterUsername  string                 `json:"twitter_username,omitempty"`
	URL              string                 `json:"url,omitempty"`
}

// Match returns true if the other page has got similarities with itself
func (page RunstatusPage) Match(other RunstatusPage) bool {
	if other.Subdomain != "" && page.Subdomain == other.Subdomain {
		return true
	}

	if other.ID > 0 && page.ID == other.ID {
		return true
	}

	return false
}

// RunstatusPageList runstatus page list
type RunstatusPageList struct {
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Pages    []RunstatusPage `json:"results"`
}

// CreateRunstatusPage create runstatus page
func (client *Client) CreateRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, client.Endpoint+runstatusPagesURL, page, "POST")
	if err != nil {
		return nil, err
	}

	var p *RunstatusPage
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	return p, nil
}

// DeleteRunstatusPage delete runstatus page
func (client *Client) DeleteRunstatusPage(ctx context.Context, page RunstatusPage) error {
	if page.URL == "" {
		return fmt.Errorf("empty URL for %#v", page)
	}
	_, err := client.runstatusRequest(ctx, page.URL, nil, "DELETE")
	return err
}

// GetRunstatusPage fetches the runstatus page
func (client *Client) GetRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	if page.URL != "" {
		return client.getRunstatusPage(ctx, page.URL)
	}

	ps, err := client.ListRunstatusPages(ctx)
	if err != nil {
		return nil, err
	}

	for i := range ps {
		if ps[i].Match(page) {
			return client.getRunstatusPage(ctx, ps[i].URL)
		}
	}

	return nil, fmt.Errorf("%#v not found", page)
}

func (client *Client) getRunstatusPage(ctx context.Context, pageURL string) (*RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, pageURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	p := new(RunstatusPage)
	if err := json.Unmarshal(resp, p); err != nil {
		return nil, err
	}

	// NOTE: fix the missing IDs
	for i := range p.Maintenances {
		if err := p.Maintenances[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Maintenances[i], err)
		}
	}
	for i := range p.Services {
		if err := p.Services[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Services[i], err)
		}
	}

	return p, nil
}

// ListRunstatusPages list all the runstatus pages
func (client *Client) ListRunstatusPages(ctx context.Context) ([]RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, client.Endpoint+runstatusPagesURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusPageList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	return p.Pages, nil
}

//PaginateRunstatusPages paginate on runstatus pages
func (client *Client) PaginateRunstatusPages(ctx context.Context, callback func(pages []RunstatusPage, e error) bool) {
	pageURL := client.Endpoint + runstatusPagesURL
	for pageURL != "" {
		resp, err := client.runstatusRequest(ctx, pageURL, nil, "GET")
		if err != nil {
			callback(nil, err)
			return
		}

		var p *RunstatusPageList
		if err := json.Unmarshal(resp, &p); err != nil {
			callback(nil, err)
			return
		}

		if ok := callback(p.Pages, nil); ok {
			return
		}

		pageURL = p.Next
	}
}
