package bluecat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// JSON body for Bluecat entity requests and responses
type bluecatEntity struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}

type entityResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}

// Starts a new Bluecat API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	queryArgs := map[string]string{
		"username": d.config.UserName,
		"password": d.config.Password,
	}

	resp, err := d.sendRequest(http.MethodGet, "login", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	authBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("bluecat: %v", err)
	}
	authResp := string(authBytes)

	if strings.Contains(authResp, "Authentication Error") {
		msg := strings.Trim(authResp, "\"")
		return fmt.Errorf("bluecat: request failed: %s", msg)
	}

	// Upon success, API responds with "Session Token-> BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM= <- for User : username"
	d.token = regexp.MustCompile("BAMAuthToken: [^ ]+").FindString(authResp)
	return nil
}

// Destroys Bluecat Session
func (d *DNSProvider) logout() error {
	if len(d.token) == 0 {
		// nothing to do
		return nil
	}

	resp, err := d.sendRequest(http.MethodGet, "logout", nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bluecat: request failed to delete session with HTTP status code %d", resp.StatusCode)
	}

	authBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	authResp := string(authBytes)

	if !strings.Contains(authResp, "successfully") {
		msg := strings.Trim(authResp, "\"")
		return fmt.Errorf("bluecat: request failed to delete session: %s", msg)
	}

	d.token = ""

	return nil
}

// Lookup the entity ID of the configuration named in our properties
func (d *DNSProvider) lookupConfID() (uint, error) {
	queryArgs := map[string]string{
		"parentId": strconv.Itoa(0),
		"name":     d.config.ConfigName,
		"type":     configType,
	}

	resp, err := d.sendRequest(http.MethodGet, "getEntityByName", nil, queryArgs)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var conf entityResponse
	err = json.NewDecoder(resp.Body).Decode(&conf)
	if err != nil {
		return 0, fmt.Errorf("bluecat: %v", err)
	}
	return conf.ID, nil
}

// Find the DNS view with the given name within
func (d *DNSProvider) lookupViewID(viewName string) (uint, error) {
	confID, err := d.lookupConfID()
	if err != nil {
		return 0, err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(confID), 10),
		"name":     viewName,
		"type":     viewType,
	}

	resp, err := d.sendRequest(http.MethodGet, "getEntityByName", nil, queryArgs)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var view entityResponse
	err = json.NewDecoder(resp.Body).Decode(&view)
	if err != nil {
		return 0, fmt.Errorf("bluecat: %v", err)
	}

	return view.ID, nil
}

// Return the entityId of the parent zone by recursing from the root view
// Also return the simple name of the host
func (d *DNSProvider) lookupParentZoneID(viewID uint, fqdn string) (uint, string, error) {
	parentViewID := viewID
	name := ""

	if fqdn != "" {
		zones := strings.Split(strings.Trim(fqdn, "."), ".")
		last := len(zones) - 1
		name = zones[0]

		for i := last; i > -1; i-- {
			zoneID, err := d.getZone(parentViewID, zones[i])
			if err != nil || zoneID == 0 {
				return parentViewID, name, err
			}
			if i > 0 {
				name = strings.Join(zones[0:i], ".")
			}
			parentViewID = zoneID
		}
	}

	return parentViewID, name, nil
}

// Get the DNS zone with the specified name under the parentId
func (d *DNSProvider) getZone(parentID uint, name string) (uint, error) {
	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentID), 10),
		"name":     name,
		"type":     zoneType,
	}

	resp, err := d.sendRequest(http.MethodGet, "getEntityByName", nil, queryArgs)

	// Return an empty zone if the named zone doesn't exist
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("bluecat: could not find zone named %s", name)
	}
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var zone entityResponse
	err = json.NewDecoder(resp.Body).Decode(&zone)
	if err != nil {
		return 0, fmt.Errorf("bluecat: %v", err)
	}

	return zone.ID, nil
}

// Deploy the DNS config for the specified entity to the authoritative servers
func (d *DNSProvider) deploy(entityID uint) error {
	queryArgs := map[string]string{
		"entityId": strconv.FormatUint(uint64(entityID), 10),
	}

	resp, err := d.sendRequest(http.MethodPost, "quickDeploy", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Send a REST request, using query parameters specified. The Authorization
// header will be set if we have an active auth token
func (d *DNSProvider) sendRequest(method, resource string, payload interface{}, queryArgs map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s/Services/REST/v1/%s", d.config.BaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("bluecat: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bluecat: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if len(d.token) > 0 {
		req.Header.Set("Authorization", d.token)
	}

	// Add all query parameters
	q := req.URL.Query()
	for argName, argVal := range queryArgs {
		q.Add(argName, argVal)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bluecat: %v", err)
	}

	if resp.StatusCode >= 400 {
		errBytes, _ := ioutil.ReadAll(resp.Body)
		errResp := string(errBytes)
		return nil, fmt.Errorf("bluecat: request failed with HTTP status code %d\n Full message: %s",
			resp.StatusCode, errResp)
	}

	return resp, nil
}
