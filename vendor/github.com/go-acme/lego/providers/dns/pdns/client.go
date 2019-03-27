package pdns

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"
)

type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`

	// pre-v1 API
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl,omitempty"`
}

type hostedZone struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	RRSets []rrSet `json:"rrsets"`

	// pre-v1 API
	Records []Record `json:"records"`
}

type rrSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Kind       string   `json:"kind"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records"`
	TTL        int      `json:"ttl,omitempty"`
}

type rrSets struct {
	RRSets []rrSet `json:"rrsets"`
}

type apiError struct {
	ShortMsg string `json:"error"`
}

func (a apiError) Error() string {
	return a.ShortMsg
}

type apiVersion struct {
	URL     string `json:"url"`
	Version int    `json:"version"`
}

func (d *DNSProvider) getHostedZone(fqdn string) (*hostedZone, error) {
	var zone hostedZone
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, err
	}

	u := "/servers/localhost/zones"
	result, err := d.sendRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var zones []hostedZone
	err = json.Unmarshal(result, &zones)
	if err != nil {
		return nil, err
	}

	u = ""
	for _, zone := range zones {
		if dns01.UnFqdn(zone.Name) == dns01.UnFqdn(authZone) {
			u = zone.URL
			break
		}
	}

	result, err = d.sendRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(result, &zone)
	if err != nil {
		return nil, err
	}

	// convert pre-v1 API result
	if len(zone.Records) > 0 {
		zone.RRSets = []rrSet{}
		for _, record := range zone.Records {
			set := rrSet{
				Name:    record.Name,
				Type:    record.Type,
				Records: []Record{record},
			}
			zone.RRSets = append(zone.RRSets, set)
		}
	}

	return &zone, nil
}

func (d *DNSProvider) findTxtRecord(fqdn string) (*rrSet, error) {
	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return nil, err
	}

	_, err = d.sendRequest(http.MethodGet, zone.URL, nil)
	if err != nil {
		return nil, err
	}

	for _, set := range zone.RRSets {
		if (set.Name == dns01.UnFqdn(fqdn) || set.Name == fqdn) && set.Type == "TXT" {
			return &set, nil
		}
	}

	return nil, nil
}

func (d *DNSProvider) getAPIVersion() (int, error) {
	result, err := d.sendRequest(http.MethodGet, "/api", nil)
	if err != nil {
		return 0, err
	}

	var versions []apiVersion
	err = json.Unmarshal(result, &versions)
	if err != nil {
		return 0, err
	}

	latestVersion := 0
	for _, v := range versions {
		if v.Version > latestVersion {
			latestVersion = v.Version
		}
	}

	return latestVersion, err
}

func (d *DNSProvider) sendRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	req, err := d.makeRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error talking to PDNS API -> %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, fmt.Errorf("unexpected HTTP status code %d when fetching '%s'", resp.StatusCode, req.URL)
	}

	var msg json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		if err == io.EOF {
			// empty body
			return nil, nil
		}
		// other error
		return nil, err
	}

	// check for PowerDNS error message
	if len(msg) > 0 && msg[0] == '{' {
		var errInfo apiError
		err = json.Unmarshal(msg, &errInfo)
		if err != nil {
			return nil, err
		}
		if errInfo.ShortMsg != "" {
			return nil, fmt.Errorf("error talking to PDNS API -> %v", errInfo)
		}
	}
	return msg, nil
}

func (d *DNSProvider) makeRequest(method, uri string, body io.Reader) (*http.Request, error) {
	var path = ""
	if d.config.Host.Path != "/" {
		path = d.config.Host.Path
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	if d.apiVersion > 0 && !strings.HasPrefix(uri, "/api/v") {
		uri = "/api/v" + strconv.Itoa(d.apiVersion) + uri
	}

	u := d.config.Host.Scheme + "://" + d.config.Host.Host + path + uri
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", d.config.APIKey)

	return req, nil
}
