// Package otc implements a DNS provider for solving the DNS-01 challenge
// using Open Telekom Cloud Managed DNS.
package otc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// OTC's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	identityEndpoint string
	otcBaseURL       string
	domainName       string
	projectName      string
	userName         string
	password         string
	token            string
}

// NewDNSProvider returns a DNSProvider instance configured for OTC DNS.
// Credentials must be passed in the environment variables: OTC_USER_NAME,
// OTC_DOMAIN_NAME, OTC_PASSWORD OTC_PROJECT_NAME and OTC_IDENTITY_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	domainName := os.Getenv("OTC_DOMAIN_NAME")
	userName := os.Getenv("OTC_USER_NAME")
	password := os.Getenv("OTC_PASSWORD")
	projectName := os.Getenv("OTC_PROJECT_NAME")
	identityEndpoint := os.Getenv("OTC_IDENTITY_ENDPOINT")
	return NewDNSProviderCredentials(domainName, userName, password, projectName, identityEndpoint)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for OTC DNS.
func NewDNSProviderCredentials(domainName, userName, password, projectName, identityEndpoint string) (*DNSProvider, error) {
	if domainName == "" || userName == "" || password == "" || projectName == "" {
		return nil, fmt.Errorf("OTC credentials missing")
	}

	if identityEndpoint == "" {
		identityEndpoint = "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens"
	}

	return &DNSProvider{
		identityEndpoint: identityEndpoint,
		domainName:       domainName,
		userName:         userName,
		password:         password,
		projectName:      projectName,
	}, nil
}

func (d *DNSProvider) SendRequest(method, resource string, payload interface{}) (io.Reader, error) {
	url := fmt.Sprintf("%s/%s", d.otcBaseURL, resource)

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

	// Workaround for keep alive bug in otc api
	tr := http.DefaultTransport.(*http.Transport)
	tr.DisableKeepAlives = true

	client := &http.Client{
		Timeout:   time.Duration(10 * time.Second),
		Transport: tr,
	}
	resp, err := client.Do(req)
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
	type nameResponse struct {
		Name string `json:"name"`
	}

	type userResponse struct {
		Name     string       `json:"name"`
		Password string       `json:"password"`
		Domain   nameResponse `json:"domain"`
	}

	type passwordResponse struct {
		User userResponse `json:"user"`
	}
	type identityResponse struct {
		Methods  []string         `json:"methods"`
		Password passwordResponse `json:"password"`
	}

	type scopeResponse struct {
		Project nameResponse `json:"project"`
	}

	type authResponse struct {
		Identity identityResponse `json:"identity"`
		Scope    scopeResponse    `json:"scope"`
	}

	type loginResponse struct {
		Auth authResponse `json:"auth"`
	}

	userResp := userResponse{
		Name:     d.userName,
		Password: d.password,
		Domain: nameResponse{
			Name: d.domainName,
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
					Name: d.projectName,
				},
			},
		},
	}

	body, err := json.Marshal(loginResp)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", d.identityEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
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

	type endpointResponse struct {
		Token struct {
			Catalog []struct {
				Type      string `json:"type"`
				Endpoints []struct {
					URL string `json:"url"`
				} `json:"endpoints"`
			} `json:"catalog"`
		} `json:"token"`
	}
	var endpointResp endpointResponse

	err = json.NewDecoder(resp.Body).Decode(&endpointResp)
	if err != nil {
		return err
	}

	for _, v := range endpointResp.Token.Catalog {
		if v.Type == "dns" {
			for _, endpoint := range v.Endpoints {
				d.otcBaseURL = fmt.Sprintf("%s/v2", endpoint.URL)
				continue
			}
		}
	}

	if d.otcBaseURL == "" {
		return fmt.Errorf("unable to get dns endpoint")
	}

	return nil
}

// Starts a new OTC API Session. Authenticates using userName, password
// and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	err := d.loginRequest()
	if err != nil {
		return err
	}

	return nil
}

func (d *DNSProvider) getZoneID(zone string) (string, error) {
	type zoneItem struct {
		ID string `json:"id"`
	}

	type zonesResponse struct {
		Zones []zoneItem `json:"zones"`
	}

	resource := fmt.Sprintf("zones?name=%s", zone)
	resp, err := d.SendRequest("GET", resource, nil)
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
	type recordSet struct {
		ID string `json:"id"`
	}

	type recordSetsResponse struct {
		RecordSets []recordSet `json:"recordsets"`
	}

	resource := fmt.Sprintf("zones/%s/recordsets?type=TXT&name=%s", zoneID, fqdn)
	resp, err := d.SendRequest("GET", resource, nil)
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

	_, err := d.SendRequest("DELETE", resource, nil)
	if err != nil {
		return err
	}
	return nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	if ttl < 300 {
		ttl = 300 // 300 is otc minimum value for ttl
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	err = d.login()
	if err != nil {
		return err
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("unable to get zone: %s", err)
	}

	resource := fmt.Sprintf("zones/%s/recordsets", zoneID)

	type recordset struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Type        string   `json:"type"`
		Ttl         int      `json:"ttl"`
		Records     []string `json:"records"`
	}

	r1 := &recordset{
		Name:        fqdn,
		Description: "Added TXT record for ACME dns-01 challenge using lego client",
		Type:        "TXT",
		Ttl:         300,
		Records:     []string{fmt.Sprintf("\"%s\"", value)},
	}
	_, err = d.SendRequest("POST", resource, r1)

	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	err = d.login()
	if err != nil {
		return err
	}

	zoneID, err := d.getZoneID(authZone)

	if err != nil {
		return err
	}

	recordID, err := d.getRecordSetID(zoneID, fqdn)
	if err != nil {
		return fmt.Errorf("unable go get record %s for zone %s: %s", fqdn, domain, err)
	}
	return d.deleteRecordSet(zoneID, recordID)
}
