// Package bluecat implements a DNS provider for solving the DNS-01 challenge
// using a self-hosted Bluecat Address Manager.
package bluecat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const configType = "Configuration"
const viewType = "View"
const txtType = "TXTRecord"
const zoneType = "Zone"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	UserName           string
	Password           string
	ConfigName         string
	DNSView            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("BLUECAT_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("BLUECAT_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("BLUECAT_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("BLUECAT_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// Bluecat's Address Manager REST API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
	token  string
}

// NewDNSProvider returns a DNSProvider instance configured for Bluecat DNS.
// Credentials must be passed in the environment variables: BLUECAT_SERVER_URL, BLUECAT_USER_NAME and BLUECAT_PASSWORD.
// BLUECAT_SERVER_URL should have the scheme, hostname, and port (if required) of the authoritative Bluecat BAM server.
// The REST endpoint will be appended. In addition, the Configuration name
// and external DNS View Name must be passed in BLUECAT_CONFIG_NAME and BLUECAT_DNS_VIEW
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("BLUECAT_SERVER_URL", "BLUECAT_USER_NAME", "BLUECAT_CONFIG_NAME", "BLUECAT_CONFIG_NAME", "BLUECAT_DNS_VIEW")
	if err != nil {
		return nil, fmt.Errorf("bluecat: %v", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = values["BLUECAT_SERVER_URL"]
	config.UserName = values["BLUECAT_USER_NAME"]
	config.Password = values["BLUECAT_PASSWORD"]
	config.ConfigName = values["BLUECAT_CONFIG_NAME"]
	config.DNSView = values["BLUECAT_DNS_VIEW"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Bluecat DNS.
// Deprecated
func NewDNSProviderCredentials(baseURL, userName, password, configName, dnsView string, httpClient *http.Client) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.BaseURL = baseURL
	config.UserName = userName
	config.Password = password
	config.ConfigName = configName
	config.DNSView = dnsView

	if httpClient != nil {
		config.HTTPClient = httpClient
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Bluecat DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("bluecat: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" || config.UserName == "" || config.Password == "" || config.ConfigName == "" || config.DNSView == "" {
		return nil, fmt.Errorf("bluecat: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters
// This will *not* create a subzone to contain the TXT record,
// so make sure the FQDN specified is within an extant zone.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	err := d.login()
	if err != nil {
		return err
	}

	viewID, err := d.lookupViewID(d.config.DNSView)
	if err != nil {
		return err
	}

	parentZoneID, name, err := d.lookupParentZoneID(viewID, fqdn)
	if err != nil {
		return err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentZoneID), 10),
	}

	body := bluecatEntity{
		Name:       name,
		Type:       "TXTRecord",
		Properties: fmt.Sprintf("ttl=%d|absoluteName=%s|txt=%s|", d.config.TTL, fqdn, value),
	}

	resp, err := d.sendRequest(http.MethodPost, "addEntity", body, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	addTxtBytes, _ := ioutil.ReadAll(resp.Body)
	addTxtResp := string(addTxtBytes)
	// addEntity responds only with body text containing the ID of the created record
	_, err = strconv.ParseUint(addTxtResp, 10, 64)
	if err != nil {
		return fmt.Errorf("bluecat: addEntity request failed: %s", addTxtResp)
	}

	err = d.deploy(parentZoneID)
	if err != nil {
		return err
	}

	return d.logout()
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	err := d.login()
	if err != nil {
		return err
	}

	viewID, err := d.lookupViewID(d.config.DNSView)
	if err != nil {
		return err
	}

	parentID, name, err := d.lookupParentZoneID(viewID, fqdn)
	if err != nil {
		return err
	}

	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentID), 10),
		"name":     name,
		"type":     txtType,
	}

	resp, err := d.sendRequest(http.MethodGet, "getEntityByName", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var txtRec entityResponse
	err = json.NewDecoder(resp.Body).Decode(&txtRec)
	if err != nil {
		return fmt.Errorf("bluecat: %v", err)
	}
	queryArgs = map[string]string{
		"objectId": strconv.FormatUint(uint64(txtRec.ID), 10),
	}

	resp, err = d.sendRequest(http.MethodDelete, http.MethodDelete, nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = d.deploy(parentID)
	if err != nil {
		return err
	}

	return d.logout()
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
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
		"name":     d.config.DNSView,
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
	if resp != nil && resp.StatusCode == 404 {
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
