package egoscale

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sort"
)

func csQuotePlus(s string) string {
	return strings.Replace(s, "+", "%20", -1)
}

func csEncode(s string) string {
	return csQuotePlus(url.QueryEscape(s))
}

func rawValue(b json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for _, v := range m {
		return v, nil
	}
	return nil, nil
}

func rawValues(b json.RawMessage) (json.RawMessage, error) {
	var i []json.RawMessage

	if err := json.Unmarshal(b, &i); err != nil {
		return nil, nil
	}

	return i[0], nil
}

func (exo *Client) ParseResponse(resp *http.Response) (json.RawMessage, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	a, err := rawValues(b)

	if a == nil {
		b, err = rawValue(b)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 400 {
		var e Error
		if err := json.Unmarshal(b, &e); err != nil {
			return nil, err
		}

		/* Need to account for differet error types */
		if e.ErrorCode != 0 {
			return nil, e.Error()
		} else {
			var de DNSError
			if err := json.Unmarshal(b, &de); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("Exoscale error (%d): %s", resp.StatusCode, strings.Join(de.Name, "\n"))
		}
	}

	return b, nil
}

func (exo *Client) Request(command string, params url.Values) (json.RawMessage, error) {

	mac := hmac.New(sha1.New, []byte(exo.apiSecret))
	keys := make([]string, 0)
	unencoded := make([]string, 0)

	params.Set("apikey", exo.apiKey)
	params.Set("command", command)
	params.Set("response", "json")

	for k, _ := range(params) {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range(keys) {
		arg := k + "=" + csEncode(params[k][0])
		unencoded = append(unencoded, arg)
	}
	sign_string := strings.ToLower(strings.Join(unencoded, "&"))

	mac.Write([]byte(sign_string))
	signature := csEncode(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	query := params.Encode()
	url := exo.endpoint + "?" + csQuotePlus(query) + "&signature=" + signature

	resp, err := exo.client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return exo.ParseResponse(resp)
}


func (exo *Client) DetailedRequest(uri string, params string, method string, header http.Header) (json.RawMessage, error) {
	url := exo.endpoint + uri

	req, err := http.NewRequest(method, url, strings.NewReader(params)); if err != nil {
		return nil, err
	}

	req.Header = header

	response, err := exo.client.Do(req); if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return exo.ParseResponse(response)
}
