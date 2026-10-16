/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpexpect

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"sync/atomic"
)

// RequestCount counts executed requests, so that the Traefik harness can detect specs without HTTP assertions.
var RequestCount atomic.Int64

type HTTPRequest struct {
	chain        chain
	reporter     Reporter
	baseURL      string
	client       *http.Client
	query        url.Values
	Request      *http.Request
	HTTPResponse *HTTPResponse
}

// NewRequest returns an HTTPRequest object.
func NewRequest(baseURL string, client *http.Client, reporter Reporter) *HTTPRequest {
	response := NewResponse(reporter)
	return &HTTPRequest{
		baseURL:      baseURL,
		client:       client,
		reporter:     reporter,
		chain:        makeChain(reporter),
		HTTPResponse: response,
	}
}

// GET creates a new HTTP request with GET method.
func (h *HTTPRequest) GET(rpath string) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	return h.DoRequest("GET", rpath)
}

// DoRequest creates a new HTTP request object.
func (h *HTTPRequest) DoRequest(method, rpath string) *HTTPRequest {
	uri, err := url.Parse(h.baseURL)
	if err != nil {
		h.chain.fail(err.Error())
	}

	var request *http.Request
	uri.Path = path.Join(uri.Path, rpath)
	if request, err = http.NewRequest(method, uri.String(), http.NoBody); err != nil {
		h.chain.fail(err.Error())
	}

	h.Request = request
	return h
}

// ForceResolve forces the test resolver to point to a specific endpoint
func (h *HTTPRequest) ForceResolve(ip string, port uint16) *HTTPRequest {
	addr := net.ParseIP(ip)
	if addr == nil {
		h.chain.fail(fmt.Sprintf("invalid ip address: %s", ip))
		return h
	}
	dialer := &net.Dialer{
		Timeout:   h.client.Timeout,
		KeepAlive: h.client.Timeout,
		DualStack: true,
	}
	resolveAddr := fmt.Sprintf("%s:%d", ip, int(port))

	oldTransport, ok := h.client.Transport.(*http.Transport)
	if !ok {
		h.chain.fail("invalid old transport address")
		return h
	}
	newTransport := oldTransport.Clone()
	newTransport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, resolveAddr)
	}
	h.client.Transport = newTransport
	return h
}

// Expect executes the request and returns an HTTP response.
func (h *HTTPRequest) Expect() *HTTPResponse {
	RequestCount.Add(1)

	if h.query != nil {
		h.Request.URL.RawQuery = h.query.Encode()
	}

	response, err := h.client.Do(h.Request)
	if err != nil {
		h.chain.fail(err.Error())
	}
	defer response.Body.Close()

	h.HTTPResponse.Response = response // set the HTTP response

	var content []byte
	if content, err = getContent(response); err != nil {
		h.chain.fail(err.Error())
	}
	// set content and cookies from HTTPResponse
	h.HTTPResponse.content = content
	h.HTTPResponse.cookies = h.HTTPResponse.Response.Cookies()
	return h.HTTPResponse
}

// WithURL sets the request URL appending paths when already exist.
func (h *HTTPRequest) WithURL(urlStr string) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	if u, err := url.Parse(urlStr); err != nil {
		h.chain.fail(err.Error())
	} else {
		u.Path = path.Join(h.Request.URL.Path, u.Path)
		h.Request.URL = u
	}
	return h
}

// WithHeader adds given header to request.
func (h *HTTPRequest) WithHeader(key, value string) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	switch http.CanonicalHeaderKey(key) {
	case "Host":
		h.Request.Host = value
	default:
		h.Request.Header.Add(key, value)
	}
	return h
}

// WithCookies adds given cookies to request.
func (h *HTTPRequest) WithCookies(cookies map[string]string) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	for k, v := range cookies {
		h.WithCookie(k, v)
	}
	return h
}

// WithCookie adds given single cookie to request.
func (h *HTTPRequest) WithCookie(k, v string) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	h.Request.AddCookie(&http.Cookie{Name: k, Value: v})
	return h
}

// WithBasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password
// are not encrypted.
func (h *HTTPRequest) WithBasicAuth(username, password string) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	h.Request.SetBasicAuth(username, password)
	return h
}

// WithQuery adds query parameter to request URL.
func (h *HTTPRequest) WithQuery(key string, value interface{}) *HTTPRequest {
	if h.chain.failed() {
		return h
	}
	if h.query == nil {
		h.query = make(url.Values)
	}
	h.query.Add(key, fmt.Sprint(value))
	return h
}

// getContent returns the content from the body response.
func getContent(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return []byte{}, nil
	}
	return io.ReadAll(resp.Body)
}
