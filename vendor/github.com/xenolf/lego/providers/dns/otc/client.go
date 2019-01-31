package otc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type recordset struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	TTL         int      `json:"ttl"`
	Records     []string `json:"records"`
}

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

type endpointResponse struct {
	Token token `json:"token"`
}

type token struct {
	Catalog []catalog `json:"catalog"`
}

type catalog struct {
	Type      string     `json:"type"`
	Endpoints []endpoint `json:"endpoints"`
}

type endpoint struct {
	URL string `json:"url"`
}

type zoneItem struct {
	ID string `json:"id"`
}

type zonesResponse struct {
	Zones []zoneItem `json:"zones"`
}

type recordSet struct {
	ID string `json:"id"`
}

type recordSetsResponse struct {
	RecordSets []recordSet `json:"recordsets"`
}

// Starts a new OTC API Session. Authenticates using userName, password
// and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	return d.loginRequest()
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

	var endpoints []endpoint
	for _, v := range endpointResp.Token.Catalog {
		if v.Type == "dns" {
			endpoints = append(endpoints, v.Endpoints...)
		}
	}

	if len(endpoints) > 0 {
		d.baseURL = fmt.Sprintf("%s/v2", endpoints[0].URL)
	} else {
		return fmt.Errorf("unable to get dns endpoint")
	}

	return nil
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
