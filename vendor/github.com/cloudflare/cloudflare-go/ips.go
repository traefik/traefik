package cloudflare

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// IPRanges contains lists of IPv4 and IPv6 CIDRs.
type IPRanges struct {
	IPv4CIDRs []string `json:"ipv4_cidrs"`
	IPv6CIDRs []string `json:"ipv6_cidrs"`
}

// IPsResponse is the API response containing a list of IPs.
type IPsResponse struct {
	Response
	Result IPRanges `json:"result"`
}

// IPs gets a list of Cloudflare's IP ranges.
//
// This does not require logging in to the API.
//
// API reference: https://api.cloudflare.com/#cloudflare-ips
func IPs() (IPRanges, error) {
	resp, err := http.Get(apiURL + "/ips")
	if err != nil {
		return IPRanges{}, errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return IPRanges{}, errors.Wrap(err, "Response body could not be read")
	}
	var r IPsResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return IPRanges{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}
