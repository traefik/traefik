package goacmedns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	// ua is a custom user-agent identifier
	ua = "goacmedns"
)

// userAgent returns a string that can be used as a HTTP request `User-Agent`
// header. It includes the `ua` string alongside the OS and architecture of the
// system.
func userAgent() string {
	return fmt.Sprintf("%s (%s; %s)", ua, runtime.GOOS, runtime.GOARCH)
}

var (
	// defaultTimeout is used for the httpClient Timeout settings
	defaultTimeout = 30 * time.Second
	// httpClient is a `http.Client` that is customized with the `defaultTimeout`
	httpClient = http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   defaultTimeout,
				KeepAlive: defaultTimeout,
			}).Dial,
			TLSHandshakeTimeout:   defaultTimeout,
			ResponseHeaderTimeout: defaultTimeout,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

// postAPI makes an HTTP POST request to the given URL, sending the given body
// and attaching the requested custom headers to the request. If there is no
// error the HTTP response body and HTTP response object are returned, otherwise
// an error is returned.. All POST requests include a `User-Agent` header
// populated with the `userAgent` function and a `Content-Type` header of
// `application/json`.
func postAPI(url string, body []byte, headers map[string]string) ([]byte, *http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Failed to make req: %s\n", err.Error())
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent())
	for h, v := range headers {
		req.Header.Set(h, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to do req: %s\n", err.Error())
		return nil, resp, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read body: %s\n", err.Error())
		return nil, resp, err
	}
	return respBody, resp, nil
}

// ClientError represents an error from the ACME-DNS server. It holds
// a `Message` describing the operation the client was doing, a `HTTPStatus`
// code returned by the server, and the `Body` of the HTTP Response from the
// server.
type ClientError struct {
	// Message is a string describing the client operation that failed
	Message string
	// HTTPStatus is the HTTP status code the ACME DNS server returned
	HTTPStatus int
	// Body is the response body the ACME DNS server returned
	Body []byte
}

// Error collects all of the ClientError fields into a single string
func (e ClientError) Error() string {
	return fmt.Sprintf("%s : status code %d response: %s",
		e.Message, e.HTTPStatus, string(e.Body))
}

// newClientError creates a ClientError instance populated with the given
// arguments
func newClientError(msg string, respCode int, respBody []byte) ClientError {
	return ClientError{
		Message:    msg,
		HTTPStatus: respCode,
		Body:       respBody,
	}
}

// Client is a struct that can be used to interact with an ACME DNS server to
// register accounts and update TXT records.
type Client struct {
	// baseURL is the address of the ACME DNS server
	baseURL string
}

// NewClient returns a Client configured to interact with the ACME DNS server at
// the given URL.
func NewClient(url string) Client {
	return Client{
		baseURL: url,
	}
}

// RegisterAccount creates an Account with the ACME DNS server. The optional
// `allowFrom` argument is used to constrain which CIDR ranges can use the
// created Account.
func (c Client) RegisterAccount(allowFrom []string) (Account, error) {
	var body []byte
	if len(allowFrom) > 0 {
		req := struct {
			AllowFrom []string
		}{
			AllowFrom: allowFrom,
		}
		reqBody, err := json.Marshal(req)
		if err != nil {
			return Account{}, err
		}
		body = reqBody
	}

	url := fmt.Sprintf("%s/register", c.baseURL)
	respBody, resp, err := postAPI(url, body, nil)
	if err != nil {
		return Account{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		return Account{}, newClientError(
			"failed to register account", resp.StatusCode, respBody)
	}

	var acct Account
	err = json.Unmarshal(respBody, &acct)
	if err != nil {
		return Account{}, err
	}

	return acct, nil
}

// UpdateTXTRecord updates a TXT record with the ACME DNS server to the `value`
// provided using the `account` specified.
func (c Client) UpdateTXTRecord(account Account, value string) error {
	update := struct {
		SubDomain string
		Txt       string
	}{
		SubDomain: account.SubDomain,
		Txt:       value,
	}
	updateBody, err := json.Marshal(update)
	if err != nil {
		fmt.Printf("Failed to marshal update: %s\n", update)
		return err
	}

	headers := map[string]string{
		"X-Api-User": account.Username,
		"X-Api-Key":  account.Password,
	}

	url := fmt.Sprintf("%s/update", c.baseURL)
	respBody, resp, err := postAPI(url, updateBody, headers)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return newClientError(
			"failed to update txt record", resp.StatusCode, respBody)
	}

	return nil
}
