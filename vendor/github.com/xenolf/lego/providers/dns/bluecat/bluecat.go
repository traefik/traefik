// Package bluecat implements a DNS provider for solving the DNS-01 challenge
// using a self-hosted Bluecat Address Manager.
package bluecat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"io/ioutil"

	"github.com/xenolf/lego/acmev2"
)

const bluecatUrlTemplate = "%s/Services/REST/v1"
const configType = "Configuration"
const viewType = "View"
const txtType = "TXTRecord"
const zoneType = "Zone"

type entityResponse struct {
	Id         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}

// DNSProvider is an implementation of the acmev2.ChallengeProvider interface that uses
// Bluecat's Address Manager REST API to manage TXT records for a domain.
type DNSProvider struct {
	baseUrl    string
	userName   string
	password   string
	configName string
	dnsView    string
	token      string
	httpClient *http.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Bluecat DNS.
// Credentials must be passed in the environment variables: BLUECAT_SERVER_URL,
// BLUECAT_USER_NAME and BLUECAT_PASSWORD. BLUECAT_SERVER_URL should have the
// scheme, hostname, and port (if required) of the authoritative Bluecat BAM
// server. The REST endpoint will be appended. In addition, the Configuration name
// and external DNS View Name must be passed in BLUECAT_CONFIG_NAME and
// BLUECAT_DNS_VIEW
func NewDNSProvider() (*DNSProvider, error) {
	server := os.Getenv("BLUECAT_SERVER_URL")
	userName := os.Getenv("BLUECAT_USER_NAME")
	password := os.Getenv("BLUECAT_PASSWORD")
	configName := os.Getenv("BLUECAT_CONFIG_NAME")
	dnsView := os.Getenv("BLUECAT_DNS_VIEW")
	httpClient := http.Client{Timeout: time.Duration(30 * time.Second)}
	return NewDNSProviderCredentials(server, userName, password, configName, dnsView, httpClient)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Bluecat DNS.
func NewDNSProviderCredentials(server, userName, password, configName, dnsView string, httpClient http.Client) (*DNSProvider, error) {
	if server == "" || userName == "" || password == "" || configName == "" || dnsView == "" {
		return nil, fmt.Errorf("Bluecat credentials missing")
	}

	return &DNSProvider{
		baseUrl:    fmt.Sprintf(bluecatUrlTemplate, server),
		userName:   userName,
		password:   password,
		configName: configName,
		dnsView:    dnsView,
		httpClient: http.DefaultClient,
	}, nil
}

// Send a REST request, using query parameters specified. The Authorization
// header will be set if we have an active auth token
func (d *DNSProvider) sendRequest(method, resource string, payload interface{}, queryArgs map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", d.baseUrl, resource)

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
		req.Header.Set("Authorization", d.token)
	}

	// Add all query parameters
	q := req.URL.Query()
	for argName, argVal := range queryArgs {
		q.Add(argName, argVal)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		errBytes, _ := ioutil.ReadAll(resp.Body)
		errResp := string(errBytes)
		return nil, fmt.Errorf("Bluecat API request failed with HTTP status code %d\n Full message: %s",
			resp.StatusCode, errResp)
	}

	return resp, nil
}

// Starts a new Bluecat API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	queryArgs := map[string]string{
		"username": d.userName,
		"password": d.password,
	}

	resp, err := d.sendRequest("GET", "login", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	authBytes, _ := ioutil.ReadAll(resp.Body)
	authResp := string(authBytes)

	if strings.Contains(authResp, "Authentication Error") {
		msg := strings.Trim(authResp, "\"")
		return fmt.Errorf("Bluecat API request failed: %s", msg)
	}
	// Upon success, API responds with "Session Token-> BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM= <- for User : username"
	re := regexp.MustCompile("BAMAuthToken: [^ ]+")
	token := re.FindString(authResp)
	d.token = token
	return nil
}

// Destroys Bluecat Session
func (d *DNSProvider) logout() error {
	if len(d.token) == 0 {
		// nothing to do
		return nil
	}

	resp, err := d.sendRequest("GET", "logout", nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Bluecat API request failed to delete session with HTTP status code %d", resp.StatusCode)
	} else {
		authBytes, _ := ioutil.ReadAll(resp.Body)
		authResp := string(authBytes)

		if !strings.Contains(authResp, "successfully") {
			msg := strings.Trim(authResp, "\"")
			return fmt.Errorf("Bluecat API request failed to delete session: %s", msg)
		}
	}

	d.token = ""

	return nil
}

// Lookup the entity ID of the configuration named in our properties
func (d *DNSProvider) lookupConfId() (uint, error) {
	queryArgs := map[string]string{
		"parentId": strconv.Itoa(0),
		"name":     d.configName,
		"type":     configType,
	}

	resp, err := d.sendRequest("GET", "getEntityByName", nil, queryArgs)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var conf entityResponse
	err = json.NewDecoder(resp.Body).Decode(&conf)
	if err != nil {
		return 0, err
	}
	return conf.Id, nil
}

// Find the DNS view with the given name within
func (d *DNSProvider) lookupViewId(viewName string) (uint, error) {
	confId, err := d.lookupConfId()
	if err != nil {
		return 0, err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(confId), 10),
		"name":     d.dnsView,
		"type":     viewType,
	}

	resp, err := d.sendRequest("GET", "getEntityByName", nil, queryArgs)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var view entityResponse
	err = json.NewDecoder(resp.Body).Decode(&view)
	if err != nil {
		return 0, err
	}

	return view.Id, nil
}

// Return the entityId of the parent zone by recursing from the root view
// Also return the simple name of the host
func (d *DNSProvider) lookupParentZoneId(viewId uint, fqdn string) (uint, string, error) {
	parentViewId := viewId
	name := ""

	if fqdn != "" {
		zones := strings.Split(strings.Trim(fqdn, "."), ".")
		last := len(zones) - 1
		name = zones[0]

		for i := last; i > -1; i-- {
			zoneId, err := d.getZone(parentViewId, zones[i])
			if err != nil || zoneId == 0 {
				return parentViewId, name, err
			}
			if i > 0 {
				name = strings.Join(zones[0:i], ".")
			}
			parentViewId = zoneId
		}
	}

	return parentViewId, name, nil
}

// Get the DNS zone with the specified name under the parentId
func (d *DNSProvider) getZone(parentId uint, name string) (uint, error) {

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentId), 10),
		"name":     name,
		"type":     zoneType,
	}

	resp, err := d.sendRequest("GET", "getEntityByName", nil, queryArgs)
	// Return an empty zone if the named zone doesn't exist
	if resp != nil && resp.StatusCode == 404 {
		return 0, fmt.Errorf("Bluecat API could not find zone named %s", name)
	}
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var zone entityResponse
	err = json.NewDecoder(resp.Body).Decode(&zone)
	if err != nil {
		return 0, err
	}

	return zone.Id, nil
}

// Present creates a TXT record using the specified parameters
// This will *not* create a subzone to contain the TXT record,
// so make sure the FQDN specified is within an extant zone.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acmev2.DNS01Record(domain, keyAuth)

	err := d.login()
	if err != nil {
		return err
	}

	viewId, err := d.lookupViewId(d.dnsView)
	if err != nil {
		return err
	}

	parentZoneId, name, err := d.lookupParentZoneId(viewId, fqdn)

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentZoneId), 10),
	}

	body := bluecatEntity{
		Name:       name,
		Type:       "TXTRecord",
		Properties: fmt.Sprintf("ttl=%d|absoluteName=%s|txt=%s|", ttl, fqdn, value),
	}

	resp, err := d.sendRequest("POST", "addEntity", body, queryArgs)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	addTxtBytes, _ := ioutil.ReadAll(resp.Body)
	addTxtResp := string(addTxtBytes)
	// addEntity responds only with body text containing the ID of the created record
	_, err = strconv.ParseUint(addTxtResp, 10, 64)
	if err != nil {
		return fmt.Errorf("Bluecat API addEntity request failed: %s", addTxtResp)
	}

	err = d.deploy(uint(parentZoneId))
	if err != nil {
		return err
	}

	err = d.logout()
	if err != nil {
		return err
	}

	return nil
}

// Deploy the DNS config for the specified entity to the authoritative servers
func (d *DNSProvider) deploy(entityId uint) error {
	queryArgs := map[string]string{
		"entityId": strconv.FormatUint(uint64(entityId), 10),
	}

	resp, err := d.sendRequest("POST", "quickDeploy", nil, queryArgs)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acmev2.DNS01Record(domain, keyAuth)

	err := d.login()
	if err != nil {
		return err
	}

	viewId, err := d.lookupViewId(d.dnsView)
	if err != nil {
		return err
	}

	parentId, name, err := d.lookupParentZoneId(viewId, fqdn)
	if err != nil {
		return err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentId), 10),
		"name":     name,
		"type":     txtType,
	}

	resp, err := d.sendRequest("GET", "getEntityByName", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var txtRec entityResponse
	err = json.NewDecoder(resp.Body).Decode(&txtRec)
	if err != nil {
		return err
	}
	queryArgs = map[string]string{
		"objectId": strconv.FormatUint(uint64(txtRec.Id), 10),
	}

	resp, err = d.sendRequest("DELETE", "delete", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = d.deploy(parentId)
	if err != nil {
		return err
	}

	err = d.logout()
	if err != nil {
		return err
	}

	return nil
}

//JSON body for Bluecat entity requests and responses
type bluecatEntity struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties string `json:"properties"`
}
