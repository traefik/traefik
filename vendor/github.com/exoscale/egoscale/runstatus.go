package egoscale

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RunstatusValidationErrorResponse represents an error in the API
type RunstatusValidationErrorResponse map[string][]string

// RunstatusErrorResponse represents the default errors
type RunstatusErrorResponse struct {
	Detail string `json:"detail"`
}

// runstatusPagesURL is the only URL that cannot be guessed
const runstatusPagesURL = "/pages"

// Error formats the DNSerror into a string
func (req RunstatusErrorResponse) Error() string {
	return fmt.Sprintf("Runstatus error: %s", req.Detail)
}

// Error formats the DNSerror into a string
func (req RunstatusValidationErrorResponse) Error() string {
	if len(req) > 0 {
		errs := []string{}
		for name, ss := range req {
			if len(ss) > 0 {
				errs = append(errs, fmt.Sprintf("%s: %s", name, strings.Join(ss, ", ")))
			}
		}
		return fmt.Sprintf("Runstatus error: %s", strings.Join(errs, "; "))
	}
	return fmt.Sprintf("Runstatus error")
}

func (client *Client) runstatusRequest(ctx context.Context, uri string, structParam interface{}, method string) (json.RawMessage, error) {
	reqURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if reqURL.Scheme == "" {
		return nil, fmt.Errorf("only absolute URI are considered valid, got %q", uri)
	}

	var params string
	if structParam != nil {
		m, err := json.Marshal(structParam)
		if err != nil {
			return nil, err
		}
		params = string(m)
	}

	req, err := http.NewRequest(method, reqURL.String(), strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	time := time.Now().Local().Format("2006-01-02T15:04:05-0700")

	payload := fmt.Sprintf("%s%s%s", req.URL.String(), time, params)

	mac := hmac.New(sha256.New, []byte(client.apiSecret))
	_, err = mac.Write([]byte(payload))
	if err != nil {
		return nil, err
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	var hdr = make(http.Header)

	hdr.Add("Authorization", fmt.Sprintf("Exoscale-HMAC-SHA256 %s:%s", client.APIKey, signature))
	hdr.Add("Exoscale-Date", time)
	hdr.Add("User-Agent", UserAgent)
	hdr.Add("Accept", "application/json")
	if params != "" {
		hdr.Add("Content-Type", "application/json")
	}
	req.Header = hdr

	req = req.WithContext(ctx)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	if resp.StatusCode == 204 {
		if method != "DELETE" {
			return nil, fmt.Errorf("only DELETE is expected to produce 204, was %q", method)
		}
		return nil, nil
	}

	contentType := resp.Header.Get("content-type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf(`response %d content-type expected to be "application/json", got %q`, resp.StatusCode, contentType)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		rerr := new(RunstatusValidationErrorResponse)
		if err := json.Unmarshal(b, rerr); err == nil {
			return nil, rerr
		}
		rverr := new(RunstatusErrorResponse)
		if err := json.Unmarshal(b, rverr); err != nil {
			return nil, err
		}

		return nil, rverr
	}

	return b, nil
}
