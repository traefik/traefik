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
)

func rawValue(b json.RawMessage) (json.RawMessage, error) {
	var m map[string]json.RawMessage

	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	for _, v := range m {
		return v, nil
	}
	//return nil, fmt.Errorf("Unable to extract raw value from:\n\n%s\n\n", string(b))
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
		fmt.Printf("ERROR: %s\n", b)
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

	params.Set("apikey", exo.apiKey)
	params.Set("command", command)
	params.Set("response", "json")

	s := strings.Replace(strings.ToLower(params.Encode()), "+", "%20", -1)
	mac.Write([]byte(s))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))

	s = params.Encode()
	url := exo.endpoint + "?" + s + "&signature=" + signature

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
