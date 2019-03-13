package dreamhost

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/go-acme/lego/log"
)

const (
	defaultBaseURL = "https://api.dreamhost.com"

	cmdAddRecord    = "dns-add_record"
	cmdRemoveRecord = "dns-remove_record"
)

type apiResponse struct {
	Data   string `json:"data"`
	Result string `json:"result"`
}

func (d *DNSProvider) buildQuery(action, domain, txt string) (*url.URL, error) {
	u, err := url.Parse(d.config.BaseURL)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	query.Set("key", d.config.APIKey)
	query.Set("cmd", action)
	query.Set("format", "json")
	query.Set("record", domain)
	query.Set("type", "TXT")
	query.Set("value", txt)
	query.Set("comment", url.QueryEscape("Managed By lego"))
	u.RawQuery = query.Encode()

	return u, nil
}

// updateTxtRecord will either add or remove a TXT record.
// action is either cmdAddRecord or cmdRemoveRecord
func (d *DNSProvider) updateTxtRecord(u fmt.Stringer) error {
	resp, err := d.config.HTTPClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	var response apiResponse
	err = json.Unmarshal(raw, &response)
	if err != nil {
		return fmt.Errorf("unable to decode API server response: %v: %s", err, string(raw))
	}

	if response.Result == "error" {
		return fmt.Errorf("add TXT record failed: %s", response.Data)
	}

	log.Infof("dreamhost: %s", response.Data)
	return nil
}
