// Package otc implements a DNS provider for solving the DNS-01 challenge
// using Open Telekom Cloud Managed DNS.
package otc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const defaultIdentityEndpoint = "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens"

// minTTL 300 is otc minimum value for ttl
const minTTL = 300

// Config is used to configure the creation of the DNSProvider
type Config struct {
	IdentityEndpoint   string
	DomainName         string
	ProjectName        string
	UserName           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		IdentityEndpoint:   env.GetOrDefaultString("OTC_IDENTITY_ENDPOINT", defaultIdentityEndpoint),
		PropagationTimeout: env.GetOrDefaultSecond("OTC_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OTC_POLLING_INTERVAL", acme.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("OTC_TTL", minTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("OTC_HTTP_TIMEOUT", 10*time.Second),
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,

				// Workaround for keep alive bug in otc api
				DisableKeepAlives: true,
			},
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// OTC's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	config  *Config
	baseURL string
	token   string
}

// NewDNSProvider returns a DNSProvider instance configured for OTC DNS.
// Credentials must be passed in the environment variables: OTC_USER_NAME,
// OTC_DOMAIN_NAME, OTC_PASSWORD OTC_PROJECT_NAME and OTC_IDENTITY_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("OTC_DOMAIN_NAME", "OTC_USER_NAME", "OTC_PASSWORD", "OTC_PROJECT_NAME")
	if err != nil {
		return nil, fmt.Errorf("otc: %v", err)
	}

	config := NewDefaultConfig()
	config.DomainName = values["OTC_DOMAIN_NAME"]
	config.UserName = values["OTC_USER_NAME"]
	config.Password = values["OTC_PASSWORD"]
	config.ProjectName = values["OTC_PROJECT_NAME"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for OTC DNS.
// Deprecated
func NewDNSProviderCredentials(domainName, userName, password, projectName, identityEndpoint string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.IdentityEndpoint = identityEndpoint
	config.DomainName = domainName
	config.UserName = userName
	config.Password = password
	config.ProjectName = projectName

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OTC DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("otc: the configuration of the DNS provider is nil")
	}

	if config.DomainName == "" || config.UserName == "" || config.Password == "" || config.ProjectName == "" {
		return nil, fmt.Errorf("otc: credentials missing")
	}

	if config.IdentityEndpoint == "" {
		config.IdentityEndpoint = defaultIdentityEndpoint
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	if d.config.TTL < minTTL {
		d.config.TTL = minTTL
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("otc: unable to get zone: %s", err)
	}

	resource := fmt.Sprintf("zones/%s/recordsets", zoneID)

	r1 := &recordset{
		Name:        fqdn,
		Description: "Added TXT record for ACME dns-01 challenge using lego client",
		Type:        "TXT",
		TTL:         d.config.TTL,
		Records:     []string{fmt.Sprintf("\"%s\"", value)},
	}

	_, err = d.sendRequest(http.MethodPost, resource, r1)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	recordID, err := d.getRecordSetID(zoneID, fqdn)
	if err != nil {
		return fmt.Errorf("otc: unable go get record %s for zone %s: %s", fqdn, domain, err)
	}

	err = d.deleteRecordSet(zoneID, recordID)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// sendRequest send request
func (d *DNSProvider) sendRequest(method, resource string, payload interface{}) (io.Reader, error) {
	url := fmt.Sprintf("%s/%s", d.baseURL, resource)

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
		req.Header.Set("X-Auth-Token", d.token)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("OTC API request %s failed with HTTP status code %d", url, resp.StatusCode)
	}

	body1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body1), nil
}

func (d *DNSProvider) loginRequest() error {
	userResp := userResponse{
		Name:     d.config.UserName,
		Password: d.config.Password,
		Domain: nameResponse{
			Name: d.config.DomainName,
		},
	}

	loginResp := loginResponse{
		Auth: authResponse{
			Identity: identityResponse{
				Methods: []string{"password"},
				Password: passwordResponse{
					User: userResp,
				},
			},
			Scope: scopeResponse{
				Project: nameResponse{
					Name: d.config.ProjectName,
				},
			},
		},
	}

	body, err := json.Marshal(loginResp)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, d.config.IdentityEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: d.config.HTTPClient.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("OTC API request failed with HTTP status code %d", resp.StatusCode)
	}

	d.token = resp.Header.Get("X-Subject-Token")

	if d.token == "" {
		return fmt.Errorf("unable to get auth token")
	}

	var endpointResp endpointResponse

	err = json.NewDecoder(resp.Body).Decode(&endpointResp)
	if err != nil {
		return err
	}

	for _, v := range endpointResp.Token.Catalog {
		if v.Type == "dns" {
			for _, endpoint := range v.Endpoints {
				d.baseURL = fmt.Sprintf("%s/v2", endpoint.URL)
				continue
			}
		}
	}

	if d.baseURL == "" {
		return fmt.Errorf("unable to get dns endpoint")
	}

	return nil
}

// Starts a new OTC API Session. Authenticates using userName, password
// and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	return d.loginRequest()
}

func (d *DNSProvider) getZoneID(zone string) (string, error) {
	resource := fmt.Sprintf("zones?name=%s", zone)
	resp, err := d.sendRequest(http.MethodGet, resource, nil)
	if err != nil {
		return "", err
	}

	var zonesRes zonesResponse
	err = json.NewDecoder(resp).Decode(&zonesRes)
	if err != nil {
		return "", err
	}

	if len(zonesRes.Zones) < 1 {
		return "", fmt.Errorf("zone %s not found", zone)
	}

	if len(zonesRes.Zones) > 1 {
		return "", fmt.Errorf("to many zones found")
	}

	if zonesRes.Zones[0].ID == "" {
		return "", fmt.Errorf("id not found")
	}

	return zonesRes.Zones[0].ID, nil
}

func (d *DNSProvider) getRecordSetID(zoneID string, fqdn string) (string, error) {
	resource := fmt.Sprintf("zones/%s/recordsets?type=TXT&name=%s", zoneID, fqdn)
	resp, err := d.sendRequest(http.MethodGet, resource, nil)
	if err != nil {
		return "", err
	}

	var recordSetsRes recordSetsResponse
	err = json.NewDecoder(resp).Decode(&recordSetsRes)
	if err != nil {
		return "", err
	}

	if len(recordSetsRes.RecordSets) < 1 {
		return "", fmt.Errorf("record not found")
	}

	if len(recordSetsRes.RecordSets) > 1 {
		return "", fmt.Errorf("to many records found")
	}

	if recordSetsRes.RecordSets[0].ID == "" {
		return "", fmt.Errorf("id not found")
	}

	return recordSetsRes.RecordSets[0].ID, nil
}

func (d *DNSProvider) deleteRecordSet(zoneID, recordID string) error {
	resource := fmt.Sprintf("zones/%s/recordsets/%s", zoneID, recordID)

	_, err := d.sendRequest(http.MethodDelete, resource, nil)
	return err
}
