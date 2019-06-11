package joker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
)

const defaultBaseURL = "https://dmapi.joker.com/request/"

// Joker DMAPI Response
type response struct {
	Headers    url.Values
	Body       string
	StatusCode int
	StatusText string
	AuthSid    string
}

// parseResponse parses HTTP response body
func parseResponse(message string) *response {
	r := &response{Headers: url.Values{}, StatusCode: -1}

	parts := strings.SplitN(message, "\n\n", 2)

	for _, line := range strings.Split(parts[0], "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		kv := strings.SplitN(line, ":", 2)

		val := ""
		if len(kv) == 2 {
			val = strings.TrimSpace(kv[1])
		}

		r.Headers.Add(kv[0], val)

		switch kv[0] {
		case "Status-Code":
			i, err := strconv.Atoi(val)
			if err == nil {
				r.StatusCode = i
			}
		case "Status-Text":
			r.StatusText = val
		case "Auth-Sid":
			r.AuthSid = val
		}
	}

	if len(parts) > 1 {
		r.Body = parts[1]
	}

	return r
}

// login performs a login to Joker's DMAPI
func (d *DNSProvider) login() (*response, error) {
	if d.config.AuthSid != "" {
		// already logged in
		return nil, nil
	}

	response, err := d.postRequest("login", url.Values{"api-key": {d.config.APIKey}})
	if err != nil {
		return response, err
	}

	if response == nil {
		return nil, fmt.Errorf("login returned nil response")
	}

	if response.AuthSid == "" {
		return response, fmt.Errorf("login did not return valid Auth-Sid")
	}

	d.config.AuthSid = response.AuthSid

	return response, nil
}

// logout closes authenticated session with Joker's DMAPI
func (d *DNSProvider) logout() (*response, error) {
	if d.config.AuthSid == "" {
		return nil, fmt.Errorf("already logged out")
	}

	response, err := d.postRequest("logout", url.Values{})
	if err == nil {
		d.config.AuthSid = ""
	}
	return response, err
}

// getZone returns content of DNS zone for domain
func (d *DNSProvider) getZone(domain string) (*response, error) {
	if d.config.AuthSid == "" {
		return nil, fmt.Errorf("must be logged in to get zone")
	}

	return d.postRequest("dns-zone-get", url.Values{"domain": {dns01.UnFqdn(domain)}})
}

// putZone uploads DNS zone to Joker DMAPI
func (d *DNSProvider) putZone(domain, zone string) (*response, error) {
	if d.config.AuthSid == "" {
		return nil, fmt.Errorf("must be logged in to put zone")
	}

	return d.postRequest("dns-zone-put", url.Values{"domain": {dns01.UnFqdn(domain)}, "zone": {strings.TrimSpace(zone)}})
}

// postRequest performs actual HTTP request
func (d *DNSProvider) postRequest(cmd string, data url.Values) (*response, error) {
	uri := d.config.BaseURL + cmd

	if d.config.AuthSid != "" {
		data.Set("auth-sid", d.config.AuthSid)
	}

	if d.config.Debug {
		log.Infof("postRequest:\n\tURL: %q\n\tData: %v", uri, data)
	}

	resp, err := d.config.HTTPClient.PostForm(uri, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d [%s]: %v", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}

	return parseResponse(string(body)), nil
}

// Temporary workaround, until it get fixed on API side
func fixTxtLines(line string) string {
	fields := strings.Fields(line)

	if len(fields) < 6 || fields[1] != "TXT" {
		return line
	}

	if fields[3][0] == '"' && fields[4] == `"` {
		fields[3] = strings.TrimSpace(fields[3]) + `"`
		fields = append(fields[:4], fields[5:]...)
	}

	return strings.Join(fields, " ")
}

// removeTxtEntryFromZone clean-ups all TXT records with given name
func removeTxtEntryFromZone(zone, relative string) (string, bool) {
	prefix := fmt.Sprintf("%s TXT 0 ", relative)

	modified := false
	var zoneEntries []string
	for _, line := range strings.Split(zone, "\n") {
		if strings.HasPrefix(line, prefix) {
			modified = true
			continue
		}
		zoneEntries = append(zoneEntries, line)
	}

	return strings.TrimSpace(strings.Join(zoneEntries, "\n")), modified
}

// addTxtEntryToZone returns DNS zone with added TXT record
func addTxtEntryToZone(zone, relative, value string, ttl int) string {
	var zoneEntries []string

	for _, line := range strings.Split(zone, "\n") {
		zoneEntries = append(zoneEntries, fixTxtLines(line))
	}

	newZoneEntry := fmt.Sprintf("%s TXT 0 %q %d", relative, value, ttl)
	zoneEntries = append(zoneEntries, newZoneEntry)

	return strings.TrimSpace(strings.Join(zoneEntries, "\n"))
}
