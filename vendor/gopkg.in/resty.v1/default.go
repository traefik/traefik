// Copyright (c) 2015-2019 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/publicsuffix"
)

// DefaultClient of resty
var DefaultClient *Client

// New method creates a new go-resty client.
func New() *Client {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return createClient(&http.Client{Jar: cookieJar})
}

// NewWithClient method create a new go-resty client with given `http.Client`.
func NewWithClient(hc *http.Client) *Client {
	return createClient(hc)
}

// R creates a new resty request object, it is used form a HTTP/RESTful request
// such as GET, POST, PUT, DELETE, HEAD, PATCH and OPTIONS.
func R() *Request {
	return DefaultClient.R()
}

// NewRequest is an alias for R(). Creates a new resty request object, it is used form a HTTP/RESTful request
// such as GET, POST, PUT, DELETE, HEAD, PATCH and OPTIONS.
func NewRequest() *Request {
	return R()
}

// SetHostURL sets Host URL. See `Client.SetHostURL for more information.
func SetHostURL(url string) *Client {
	return DefaultClient.SetHostURL(url)
}

// SetHeader sets single header. See `Client.SetHeader` for more information.
func SetHeader(header, value string) *Client {
	return DefaultClient.SetHeader(header, value)
}

// SetHeaders sets multiple headers. See `Client.SetHeaders` for more information.
func SetHeaders(headers map[string]string) *Client {
	return DefaultClient.SetHeaders(headers)
}

// SetCookieJar sets custom http.CookieJar. See `Client.SetCookieJar` for more information.
func SetCookieJar(jar http.CookieJar) *Client {
	return DefaultClient.SetCookieJar(jar)
}

// SetCookie sets single cookie object. See `Client.SetCookie` for more information.
func SetCookie(hc *http.Cookie) *Client {
	return DefaultClient.SetCookie(hc)
}

// SetCookies sets multiple cookie object. See `Client.SetCookies` for more information.
func SetCookies(cs []*http.Cookie) *Client {
	return DefaultClient.SetCookies(cs)
}

// SetQueryParam method sets single parameter and its value. See `Client.SetQueryParam` for more information.
func SetQueryParam(param, value string) *Client {
	return DefaultClient.SetQueryParam(param, value)
}

// SetQueryParams method sets multiple parameters and its value. See `Client.SetQueryParams` for more information.
func SetQueryParams(params map[string]string) *Client {
	return DefaultClient.SetQueryParams(params)
}

// SetFormData method sets Form parameters and its values. See `Client.SetFormData` for more information.
func SetFormData(data map[string]string) *Client {
	return DefaultClient.SetFormData(data)
}

// SetBasicAuth method sets the basic authentication header. See `Client.SetBasicAuth` for more information.
func SetBasicAuth(username, password string) *Client {
	return DefaultClient.SetBasicAuth(username, password)
}

// SetAuthToken method sets bearer auth token header. See `Client.SetAuthToken` for more information.
func SetAuthToken(token string) *Client {
	return DefaultClient.SetAuthToken(token)
}

// OnBeforeRequest method sets request middleware. See `Client.OnBeforeRequest` for more information.
func OnBeforeRequest(m func(*Client, *Request) error) *Client {
	return DefaultClient.OnBeforeRequest(m)
}

// OnAfterResponse method sets response middleware. See `Client.OnAfterResponse` for more information.
func OnAfterResponse(m func(*Client, *Response) error) *Client {
	return DefaultClient.OnAfterResponse(m)
}

// SetPreRequestHook method sets the pre-request hook. See `Client.SetPreRequestHook` for more information.
func SetPreRequestHook(h func(*Client, *Request) error) *Client {
	return DefaultClient.SetPreRequestHook(h)
}

// SetDebug method enables the debug mode. See `Client.SetDebug` for more information.
func SetDebug(d bool) *Client {
	return DefaultClient.SetDebug(d)
}

// SetDebugBodyLimit method sets the response body limit for debug mode. See `Client.SetDebugBodyLimit` for more information.
func SetDebugBodyLimit(sl int64) *Client {
	return DefaultClient.SetDebugBodyLimit(sl)
}

// SetAllowGetMethodPayload method allows the GET method with payload. See `Client.SetAllowGetMethodPayload` for more information.
func SetAllowGetMethodPayload(a bool) *Client {
	return DefaultClient.SetAllowGetMethodPayload(a)
}

// SetRetryCount method sets the retry count. See `Client.SetRetryCount` for more information.
func SetRetryCount(count int) *Client {
	return DefaultClient.SetRetryCount(count)
}

// SetRetryWaitTime method sets the retry wait time. See `Client.SetRetryWaitTime` for more information.
func SetRetryWaitTime(waitTime time.Duration) *Client {
	return DefaultClient.SetRetryWaitTime(waitTime)
}

// SetRetryMaxWaitTime method sets the retry max wait time. See `Client.SetRetryMaxWaitTime` for more information.
func SetRetryMaxWaitTime(maxWaitTime time.Duration) *Client {
	return DefaultClient.SetRetryMaxWaitTime(maxWaitTime)
}

// AddRetryCondition method appends check function for retry. See `Client.AddRetryCondition` for more information.
func AddRetryCondition(condition RetryConditionFunc) *Client {
	return DefaultClient.AddRetryCondition(condition)
}

// SetDisableWarn method disables warning comes from `go-resty` client. See `Client.SetDisableWarn` for more information.
func SetDisableWarn(d bool) *Client {
	return DefaultClient.SetDisableWarn(d)
}

// SetLogger method sets given writer for logging. See `Client.SetLogger` for more information.
func SetLogger(w io.Writer) *Client {
	return DefaultClient.SetLogger(w)
}

// SetContentLength method enables `Content-Length` value. See `Client.SetContentLength` for more information.
func SetContentLength(l bool) *Client {
	return DefaultClient.SetContentLength(l)
}

// SetError method is to register the global or client common `Error` object. See `Client.SetError` for more information.
func SetError(err interface{}) *Client {
	return DefaultClient.SetError(err)
}

// SetRedirectPolicy method sets the client redirect poilicy. See `Client.SetRedirectPolicy` for more information.
func SetRedirectPolicy(policies ...interface{}) *Client {
	return DefaultClient.SetRedirectPolicy(policies...)
}

// SetHTTPMode method sets go-resty mode into HTTP. See `Client.SetMode` for more information.
func SetHTTPMode() *Client {
	return DefaultClient.SetHTTPMode()
}

// SetRESTMode method sets go-resty mode into RESTful. See `Client.SetMode` for more information.
func SetRESTMode() *Client {
	return DefaultClient.SetRESTMode()
}

// Mode method returns the current client mode. See `Client.Mode` for more information.
func Mode() string {
	return DefaultClient.Mode()
}

// SetTLSClientConfig method sets TLSClientConfig for underling client Transport. See `Client.SetTLSClientConfig` for more information.
func SetTLSClientConfig(config *tls.Config) *Client {
	return DefaultClient.SetTLSClientConfig(config)
}

// SetTimeout method sets timeout for request. See `Client.SetTimeout` for more information.
func SetTimeout(timeout time.Duration) *Client {
	return DefaultClient.SetTimeout(timeout)
}

// SetProxy method sets Proxy for request. See `Client.SetProxy` for more information.
func SetProxy(proxyURL string) *Client {
	return DefaultClient.SetProxy(proxyURL)
}

// RemoveProxy method removes the proxy configuration. See `Client.RemoveProxy` for more information.
func RemoveProxy() *Client {
	return DefaultClient.RemoveProxy()
}

// SetCertificates method helps to set client certificates into resty conveniently.
// See `Client.SetCertificates` for more information and example.
func SetCertificates(certs ...tls.Certificate) *Client {
	return DefaultClient.SetCertificates(certs...)
}

// SetRootCertificate method helps to add one or more root certificates into resty client.
// See `Client.SetRootCertificate` for more information.
func SetRootCertificate(pemFilePath string) *Client {
	return DefaultClient.SetRootCertificate(pemFilePath)
}

// SetOutputDirectory method sets output directory. See `Client.SetOutputDirectory` for more information.
func SetOutputDirectory(dirPath string) *Client {
	return DefaultClient.SetOutputDirectory(dirPath)
}

// SetTransport method sets custom `*http.Transport` or any `http.RoundTripper`
// compatible interface implementation in the resty client.
// See `Client.SetTransport` for more information.
func SetTransport(transport http.RoundTripper) *Client {
	return DefaultClient.SetTransport(transport)
}

// SetScheme method sets custom scheme in the resty client.
// See `Client.SetScheme` for more information.
func SetScheme(scheme string) *Client {
	return DefaultClient.SetScheme(scheme)
}

// SetCloseConnection method sets close connection value in the resty client.
// See `Client.SetCloseConnection` for more information.
func SetCloseConnection(close bool) *Client {
	return DefaultClient.SetCloseConnection(close)
}

// SetDoNotParseResponse method instructs `Resty` not to parse the response body automatically.
// See `Client.SetDoNotParseResponse` for more information.
func SetDoNotParseResponse(parse bool) *Client {
	return DefaultClient.SetDoNotParseResponse(parse)
}

// SetPathParams method sets the Request path parameter key-value pairs. See
// `Client.SetPathParams` for more information.
func SetPathParams(params map[string]string) *Client {
	return DefaultClient.SetPathParams(params)
}

// IsProxySet method returns the true if proxy is set on client otherwise false.
// See `Client.IsProxySet` for more information.
func IsProxySet() bool {
	return DefaultClient.IsProxySet()
}

// GetClient method returns the current `http.Client` used by the default resty client.
func GetClient() *http.Client {
	return DefaultClient.httpClient
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Unexported methods
//___________________________________

func createClient(hc *http.Client) *Client {
	c := &Client{
		HostURL:            "",
		QueryParam:         url.Values{},
		FormData:           url.Values{},
		Header:             http.Header{},
		UserInfo:           nil,
		Token:              "",
		Cookies:            make([]*http.Cookie, 0),
		Debug:              false,
		Log:                getLogger(os.Stderr),
		RetryCount:         0,
		RetryWaitTime:      defaultWaitTime,
		RetryMaxWaitTime:   defaultMaxWaitTime,
		JSONMarshal:        json.Marshal,
		JSONUnmarshal:      json.Unmarshal,
		jsonEscapeHTML:     true,
		httpClient:         hc,
		debugBodySizeLimit: math.MaxInt32,
		pathParams:         make(map[string]string),
	}

	// Log Prefix
	c.SetLogPrefix("RESTY ")

	// Default redirect policy
	c.SetRedirectPolicy(NoRedirectPolicy())

	// default before request middlewares
	c.beforeRequest = []func(*Client, *Request) error{
		parseRequestURL,
		parseRequestHeader,
		parseRequestBody,
		createHTTPRequest,
		addCredentials,
	}

	// user defined request middlewares
	c.udBeforeRequest = []func(*Client, *Request) error{}

	// default after response middlewares
	c.afterResponse = []func(*Client, *Response) error{
		responseLogger,
		parseResponseBody,
		saveResponseIntoFile,
	}

	return c
}

func init() {
	DefaultClient = New()
}
