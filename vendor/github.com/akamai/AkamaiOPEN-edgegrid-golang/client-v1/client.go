// Package client is a simple library for http.Client to sign Akamai OPEN Edgegrid API requests
package client

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/jsonhooks-v1"
)

var (
	libraryVersion = "0.6.2"
	// UserAgent is the User-Agent value sent for all requests
	UserAgent = "Akamai-Open-Edgegrid-golang/" + libraryVersion + " golang/" + strings.TrimPrefix(runtime.Version(), "go")
	// Client is the *http.Client to use
	Client = http.DefaultClient
)

// NewRequest creates an HTTP request that can be sent to Akamai APIs. A relative URL can be provided in path, which will be resolved to the
// Host specified in Config. If body is specified, it will be sent as the request body.
func NewRequest(config edgegrid.Config, method, path string, body io.Reader) (*http.Request, error) {
	var (
		baseURL *url.URL
		err     error
	)

	if strings.HasPrefix(config.Host, "https://") {
		baseURL, err = url.Parse(config.Host)
	} else {
		baseURL, err = url.Parse("https://" + config.Host)
	}

	if err != nil {
		return nil, err
	}

	rel, err := url.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, err
	}

	u := baseURL.ResolveReference(rel)

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", UserAgent)

	return req, nil
}

// NewJSONRequest creates an HTTP request that can be sent to the Akamai APIs with a JSON body
// The JSON body is encoded and the Content-Type/Accept headers are set automatically.
func NewJSONRequest(config edgegrid.Config, method, path string, body interface{}) (*http.Request, error) {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := jsonhooks.Marshal(body)
		if err != nil {
			return nil, err
		}

		buf := bytes.NewReader(jsonBody)
		req, err = NewRequest(config, method, path, buf)
	} else {
		req, err = NewRequest(config, method, path, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json,*/*")

	return req, nil
}

// NewMultiPartFormDataRequest creates an HTTP request that uploads a file to the Akamai API
func NewMultiPartFormDataRequest(config edgegrid.Config, uriPath, filePath string, otherFormParams map[string]string) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// TODO: make this field name configurable
	part, err := writer.CreateFormFile("importFile", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range otherFormParams {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := NewRequest(config, "POST", uriPath, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

// Do performs a given HTTP Request, signed with the Akamai OPEN Edgegrid
// Authorization header. An edgegrid.Response or an error is returned.
func Do(config edgegrid.Config, req *http.Request) (*http.Response, error) {
	Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		req = edgegrid.AddRequestHeader(config, req)
		return nil
	}

	req = edgegrid.AddRequestHeader(config, req)
	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// BodyJSON unmarshals the Response.Body into a given data structure
func BodyJSON(r *http.Response, data interface{}) error {
	if data == nil {
		return errors.New("You must pass in an interface{}")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = jsonhooks.Unmarshal(body, data)

	return err
}
