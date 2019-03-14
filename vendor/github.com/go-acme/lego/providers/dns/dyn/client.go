package dyn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://api.dynect.net/REST"

type dynResponse struct {
	// One of 'success', 'failure', or 'incomplete'
	Status string `json:"status"`

	// The structure containing the actual results of the request
	Data json.RawMessage `json:"data"`

	// The ID of the job that was created in response to a request.
	JobID int `json:"job_id"`

	// A list of zero or more messages
	Messages json.RawMessage `json:"msgs"`
}

type credentials struct {
	Customer string `json:"customer_name"`
	User     string `json:"user_name"`
	Pass     string `json:"password"`
}

type session struct {
	Token   string `json:"token"`
	Version string `json:"version"`
}

type publish struct {
	Publish bool   `json:"publish"`
	Notes   string `json:"notes"`
}

// Starts a new Dyn API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	payload := &credentials{Customer: d.config.CustomerName, User: d.config.UserName, Pass: d.config.Password}
	dynRes, err := d.sendRequest(http.MethodPost, "Session", payload)
	if err != nil {
		return err
	}

	var s session
	err = json.Unmarshal(dynRes.Data, &s)
	if err != nil {
		return err
	}

	d.token = s.Token

	return nil
}

// Destroys Dyn Session
func (d *DNSProvider) logout() error {
	if len(d.token) == 0 {
		// nothing to do
		return nil
	}

	url := fmt.Sprintf("%s/Session", defaultBaseURL)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed to delete session with HTTP status code %d", resp.StatusCode)
	}

	d.token = ""

	return nil
}

func (d *DNSProvider) publish(zone, notes string) error {
	pub := &publish{Publish: true, Notes: notes}
	resource := fmt.Sprintf("Zone/%s/", zone)

	_, err := d.sendRequest(http.MethodPut, resource, pub)
	return err
}

func (d *DNSProvider) sendRequest(method, resource string, payload interface{}) (*dynResponse, error) {
	url := fmt.Sprintf("%s/%s", defaultBaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(d.token) > 0 {
		req.Header.Set("Auth-Token", d.token)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("API request failed with HTTP status code %d", resp.StatusCode)
	}

	var dynRes dynResponse
	err = json.NewDecoder(resp.Body).Decode(&dynRes)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with HTTP status code %d: %s", resp.StatusCode, dynRes.Messages)
	} else if resp.StatusCode == 307 {
		// TODO add support for HTTP 307 response and long running jobs
		return nil, fmt.Errorf("API request returned HTTP 307. This is currently unsupported")
	}

	if dynRes.Status == "failure" {
		// TODO add better error handling
		return nil, fmt.Errorf("API request failed: %s", dynRes.Messages)
	}

	return &dynRes, nil
}
