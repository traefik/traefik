package dode

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/go-acme/lego/challenge/dns01"
)

type apiResponse struct {
	Domain  string
	Success bool
}

// updateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
func (d *DNSProvider) updateTxtRecord(fqdn, token, txt string, clear bool) error {
	u, _ := url.Parse("https://www.do.de/api/letsencrypt")

	query := u.Query()
	query.Set("token", token)
	query.Set("domain", dns01.UnFqdn(fqdn))

	// api call differs per set/delete
	if clear {
		query.Set("action", "delete")
	} else {
		query.Set("value", txt)
	}

	u.RawQuery = query.Encode()

	response, err := d.config.HTTPClient.Get(u.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var r apiResponse
	err = json.Unmarshal(bodyBytes, &r)
	if err != nil {
		return fmt.Errorf("request to change TXT record for do.de returned the following invalid json (%s); used url [%s]", string(bodyBytes), u)
	}

	body := string(bodyBytes)
	if !r.Success {
		return fmt.Errorf("request to change TXT record for do.de returned the following error result (%s); used url [%s]", body, u)
	}
	return nil
}
