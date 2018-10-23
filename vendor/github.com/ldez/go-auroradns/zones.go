package auroradns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Zone a DNS zone
type Zone struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

// CreateZone Creates a zone.
func (c *Client) CreateZone(domain string) (*Zone, *http.Response, error) {
	body, err := json.Marshal(Zone{Name: domain})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshall request body: %v", err)
	}

	req, err := c.newRequest(http.MethodPost, "/zones", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	zone := new(Zone)
	resp, err := c.do(req, zone)
	if err != nil {
		return nil, resp, err
	}

	return zone, resp, nil
}

// DeleteZone Delete a zone.
func (c *Client) DeleteZone(zoneID string) (bool, *http.Response, error) {
	resource := fmt.Sprintf("/zones/%s", zoneID)

	req, err := c.newRequest(http.MethodDelete, resource, nil)
	if err != nil {
		return false, nil, err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return false, resp, err
	}

	return true, resp, nil

}

// ListZones returns a list of all zones.
func (c *Client) ListZones() ([]Zone, *http.Response, error) {
	req, err := c.newRequest(http.MethodGet, "/zones", nil)
	if err != nil {
		return nil, nil, err
	}

	var zones []Zone
	resp, err := c.do(req, &zones)
	if err != nil {
		return nil, resp, err
	}

	return zones, resp, nil
}
