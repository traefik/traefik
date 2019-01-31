package mydnsjp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func (d *DNSProvider) doRequest(domain, value string, cmd string) error {
	req, err := d.buildRequest(domain, value, cmd)
	if err != nil {
		return err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error querying API: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var content []byte
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("request %s failed [status code %d]: %s", req.URL, resp.StatusCode, string(content))
	}

	return nil
}

func (d *DNSProvider) buildRequest(domain, value string, cmd string) (*http.Request, error) {
	params := url.Values{}
	params.Set("CERTBOT_DOMAIN", domain)
	params.Set("CERTBOT_VALIDATION", value)
	params.Set("EDIT_CMD", cmd)

	req, err := http.NewRequest(http.MethodPost, defaultBaseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("invalid request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(d.config.MasterID, d.config.Password)

	return req, nil
}
