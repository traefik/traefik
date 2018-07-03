package vegadns2client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// VegaDNSClient - Struct for holding VegaDNSClient specific attributes
type VegaDNSClient struct {
	client    http.Client
	baseurl   string
	version   string
	User      string
	Pass      string
	APIKey    string
	APISecret string
	token     Token
}

// Send - Central place for sending requests
// Input: method, endpoint, params
// Output: *http.Response
func (vega *VegaDNSClient) Send(method string, endpoint string, params map[string]string) (*http.Response, error) {
	vegaURL := vega.getURL(endpoint)

	p := url.Values{}
	for k, v := range params {
		p.Set(k, v)
	}

	var err error
	var req *http.Request

	if (method == "GET") || (method == "DELETE") {
		vegaURL = fmt.Sprintf("%s?%s", vegaURL, p.Encode())
		req, err = http.NewRequest(method, vegaURL, nil)
	} else {
		req, err = http.NewRequest(method, vegaURL, strings.NewReader(p.Encode()))
	}

	if err != nil {
		return nil, fmt.Errorf("Error preparing request: %s", err)
	}

	if vega.User != "" && vega.Pass != "" {
		// Basic Auth
		req.SetBasicAuth(vega.User, vega.Pass)
	} else if vega.APIKey != "" && vega.APISecret != "" {
		// OAuth
		vega.getAuthToken()
		err = vega.token.valid()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", vega.getBearer())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return vega.client.Do(req)
}

func (vega *VegaDNSClient) getURL(endpoint string) string {
	return fmt.Sprintf("%s/%s/%s", vega.baseurl, vega.version, endpoint)
}

func (vega *VegaDNSClient) stillAuthorized() error {
	return vega.token.valid()
}
