// Copyright (c) 2015-2018 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	// MethodGet HTTP method
	MethodGet = "GET"

	// MethodPost HTTP method
	MethodPost = "POST"

	// MethodPut HTTP method
	MethodPut = "PUT"

	// MethodDelete HTTP method
	MethodDelete = "DELETE"

	// MethodPatch HTTP method
	MethodPatch = "PATCH"

	// MethodHead HTTP method
	MethodHead = "HEAD"

	// MethodOptions HTTP method
	MethodOptions = "OPTIONS"
)

var (
	hdrUserAgentKey       = http.CanonicalHeaderKey("User-Agent")
	hdrAcceptKey          = http.CanonicalHeaderKey("Accept")
	hdrContentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	hdrContentLengthKey   = http.CanonicalHeaderKey("Content-Length")
	hdrContentEncodingKey = http.CanonicalHeaderKey("Content-Encoding")
	hdrAuthorizationKey   = http.CanonicalHeaderKey("Authorization")

	plainTextType   = "text/plain; charset=utf-8"
	jsonContentType = "application/json; charset=utf-8"
	formContentType = "application/x-www-form-urlencoded"

	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json)(;|$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(xml|.*\+xml)(;|$))`)

	hdrUserAgentValue = "go-resty/%s (https://github.com/go-resty/resty)"
	bufPool           = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

// Client type is used for HTTP/RESTful global values
// for all request raised from the client
type Client struct {
	HostURL               string
	QueryParam            url.Values
	FormData              url.Values
	Header                http.Header
	UserInfo              *User
	Token                 string
	Cookies               []*http.Cookie
	Error                 reflect.Type
	Debug                 bool
	DisableWarn           bool
	AllowGetMethodPayload bool
	Log                   *log.Logger
	RetryCount            int
	RetryWaitTime         time.Duration
	RetryMaxWaitTime      time.Duration
	RetryConditions       []RetryConditionFunc
	JSONMarshal           func(v interface{}) ([]byte, error)
	JSONUnmarshal         func(data []byte, v interface{}) error

	jsonEscapeHTML     bool
	httpClient         *http.Client
	setContentLength   bool
	isHTTPMode         bool
	outputDirectory    string
	scheme             string
	proxyURL           *url.URL
	closeConnection    bool
	notParseResponse   bool
	debugBodySizeLimit int64
	logPrefix          string
	pathParams         map[string]string
	beforeRequest      []func(*Client, *Request) error
	udBeforeRequest    []func(*Client, *Request) error
	preReqHook         func(*Client, *Request) error
	afterResponse      []func(*Client, *Response) error
}

// User type is to hold an username and password information
type User struct {
	Username, Password string
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Client methods
//___________________________________

// SetHostURL method is to set Host URL in the client instance. It will be used with request
// raised from this client with relative URL
//		// Setting HTTP address
//		resty.SetHostURL("http://myjeeva.com")
//
//		// Setting HTTPS address
//		resty.SetHostURL("https://myjeeva.com")
//
func (c *Client) SetHostURL(url string) *Client {
	c.HostURL = strings.TrimRight(url, "/")
	return c
}

// SetHeader method sets a single header field and its value in the client instance.
// These headers will be applied to all requests raised from this client instance.
// Also it can be overridden at request level header options, see `resty.R().SetHeader`
// or `resty.R().SetHeaders`.
//
// Example: To set `Content-Type` and `Accept` as `application/json`
//
// 		resty.
// 			SetHeader("Content-Type", "application/json").
// 			SetHeader("Accept", "application/json")
//
func (c *Client) SetHeader(header, value string) *Client {
	c.Header.Set(header, value)
	return c
}

// SetHeaders method sets multiple headers field and its values at one go in the client instance.
// These headers will be applied to all requests raised from this client instance. Also it can be
// overridden at request level headers options, see `resty.R().SetHeaders` or `resty.R().SetHeader`.
//
// Example: To set `Content-Type` and `Accept` as `application/json`
//
// 		resty.SetHeaders(map[string]string{
//				"Content-Type": "application/json",
//				"Accept": "application/json",
//			})
//
func (c *Client) SetHeaders(headers map[string]string) *Client {
	for h, v := range headers {
		c.Header.Set(h, v)
	}

	return c
}

// SetCookieJar method sets custom http.CookieJar in the resty client. Its way to override default.
// Example: sometimes we don't want to save cookies in api contacting, we can remove the default
// CookieJar in resty client.
//
//		resty.SetCookieJar(nil)
//
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.httpClient.Jar = jar
	return c
}

// SetCookie method appends a single cookie in the client instance.
// These cookies will be added to all the request raised from this client instance.
// 		resty.SetCookie(&http.Cookie{
// 					Name:"go-resty",
//					Value:"This is cookie value",
//					Path: "/",
// 					Domain: "sample.com",
// 					MaxAge: 36000,
// 					HttpOnly: true,
//					Secure: false,
// 				})
//
func (c *Client) SetCookie(hc *http.Cookie) *Client {
	c.Cookies = append(c.Cookies, hc)
	return c
}

// SetCookies method sets an array of cookies in the client instance.
// These cookies will be added to all the request raised from this client instance.
// 		cookies := make([]*http.Cookie, 0)
//
//		cookies = append(cookies, &http.Cookie{
// 					Name:"go-resty-1",
//					Value:"This is cookie 1 value",
//					Path: "/",
// 					Domain: "sample.com",
// 					MaxAge: 36000,
// 					HttpOnly: true,
//					Secure: false,
// 				})
//
//		cookies = append(cookies, &http.Cookie{
// 					Name:"go-resty-2",
//					Value:"This is cookie 2 value",
//					Path: "/",
// 					Domain: "sample.com",
// 					MaxAge: 36000,
// 					HttpOnly: true,
//					Secure: false,
// 				})
//
//		// Setting a cookies into resty
// 		resty.SetCookies(cookies)
//
func (c *Client) SetCookies(cs []*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cs...)
	return c
}

// SetQueryParam method sets single parameter and its value in the client instance.
// It will be formed as query string for the request. For example: `search=kitchen%20papers&size=large`
// in the URL after `?` mark. These query params will be added to all the request raised from
// this client instance. Also it can be overridden at request level Query Param options,
// see `resty.R().SetQueryParam` or `resty.R().SetQueryParams`.
// 		resty.
//			SetQueryParam("search", "kitchen papers").
//			SetQueryParam("size", "large")
//
func (c *Client) SetQueryParam(param, value string) *Client {
	c.QueryParam.Set(param, value)
	return c
}

// SetQueryParams method sets multiple parameters and their values at one go in the client instance.
// It will be formed as query string for the request. For example: `search=kitchen%20papers&size=large`
// in the URL after `?` mark. These query params will be added to all the request raised from this
// client instance. Also it can be overridden at request level Query Param options,
// see `resty.R().SetQueryParams` or `resty.R().SetQueryParam`.
// 		resty.SetQueryParams(map[string]string{
//				"search": "kitchen papers",
//				"size": "large",
//			})
//
func (c *Client) SetQueryParams(params map[string]string) *Client {
	for p, v := range params {
		c.SetQueryParam(p, v)
	}

	return c
}

// SetFormData method sets Form parameters and their values in the client instance.
// It's applicable only HTTP method `POST` and `PUT` and requets content type would be set as
// `application/x-www-form-urlencoded`. These form data will be added to all the request raised from
// this client instance. Also it can be overridden at request level form data, see `resty.R().SetFormData`.
// 		resty.SetFormData(map[string]string{
//				"access_token": "BC594900-518B-4F7E-AC75-BD37F019E08F",
//				"user_id": "3455454545",
//			})
//
func (c *Client) SetFormData(data map[string]string) *Client {
	for k, v := range data {
		c.FormData.Set(k, v)
	}

	return c
}

// SetBasicAuth method sets the basic authentication header in the HTTP request. Example:
//		Authorization: Basic <base64-encoded-value>
//
// Example: To set the header for username "go-resty" and password "welcome"
// 		resty.SetBasicAuth("go-resty", "welcome")
//
// This basic auth information gets added to all the request rasied from this client instance.
// Also it can be overridden or set one at the request level is supported, see `resty.R().SetBasicAuth`.
//
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.UserInfo = &User{Username: username, Password: password}
	return c
}

// SetAuthToken method sets bearer auth token header in the HTTP request. Example:
// 		Authorization: Bearer <auth-token-value-comes-here>
//
// Example: To set auth token BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F
//
// 		resty.SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F")
//
// This bearer auth token gets added to all the request rasied from this client instance.
// Also it can be overridden or set one at the request level is supported, see `resty.R().SetAuthToken`.
//
func (c *Client) SetAuthToken(token string) *Client {
	c.Token = token
	return c
}

// R method creates a request instance, its used for Get, Post, Put, Delete, Patch, Head and Options.
func (c *Client) R() *Request {
	r := &Request{
		QueryParam: url.Values{},
		FormData:   url.Values{},
		Header:     http.Header{},

		client:          c,
		multipartFiles:  []*File{},
		multipartFields: []*multipartField{},
		pathParams:      map[string]string{},
		jsonEscapeHTML:  true,
	}

	return r
}

// NewRequest is an alias for R(). Creates a request instance, its used for
// Get, Post, Put, Delete, Patch, Head and Options.
func (c *Client) NewRequest() *Request {
	return c.R()
}

// OnBeforeRequest method appends request middleware into the before request chain.
// Its gets applied after default `go-resty` request middlewares and before request
// been sent from `go-resty` to host server.
// 		resty.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
//				// Now you have access to Client and Request instance
//				// manipulate it as per your need
//
//				return nil 	// if its success otherwise return error
//			})
//
func (c *Client) OnBeforeRequest(m func(*Client, *Request) error) *Client {
	c.udBeforeRequest = append(c.udBeforeRequest, m)
	return c
}

// OnAfterResponse method appends response middleware into the after response chain.
// Once we receive response from host server, default `go-resty` response middleware
// gets applied and then user assigened response middlewares applied.
// 		resty.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
//				// Now you have access to Client and Response instance
//				// manipulate it as per your need
//
//				return nil 	// if its success otherwise return error
//			})
//
func (c *Client) OnAfterResponse(m func(*Client, *Response) error) *Client {
	c.afterResponse = append(c.afterResponse, m)
	return c
}

// SetPreRequestHook method sets the given pre-request function into resty client.
// It is called right before the request is fired.
//
// Note: Only one pre-request hook can be registered. Use `resty.OnBeforeRequest` for mutilple.
func (c *Client) SetPreRequestHook(h func(*Client, *Request) error) *Client {
	if c.preReqHook != nil {
		c.Log.Printf("Overwriting an existing pre-request hook: %s", functionName(h))
	}
	c.preReqHook = h
	return c
}

// SetDebug method enables the debug mode on `go-resty` client. Client logs details of every request and response.
// For `Request` it logs information such as HTTP verb, Relative URL path, Host, Headers, Body if it has one.
// For `Response` it logs information such as Status, Response Time, Headers, Body if it has one.
//		resty.SetDebug(true)
//
func (c *Client) SetDebug(d bool) *Client {
	c.Debug = d
	return c
}

// SetDebugBodyLimit sets the maximum size for which the response body will be logged in debug mode.
//		resty.SetDebugBodyLimit(1000000)
//
func (c *Client) SetDebugBodyLimit(sl int64) *Client {
	c.debugBodySizeLimit = sl
	return c
}

// SetDisableWarn method disables the warning message on `go-resty` client.
// For example: go-resty warns the user when BasicAuth used on HTTP mode.
//		resty.SetDisableWarn(true)
//
func (c *Client) SetDisableWarn(d bool) *Client {
	c.DisableWarn = d
	return c
}

// SetAllowGetMethodPayload method allows the GET method with payload on `go-resty` client.
// For example: go-resty allows the user sends request with a payload on HTTP GET method.
//		resty.SetAllowGetMethodPayload(true)
//
func (c *Client) SetAllowGetMethodPayload(a bool) *Client {
	c.AllowGetMethodPayload = a
	return c
}

// SetLogger method sets given writer for logging go-resty request and response details.
// Default is os.Stderr
// 		file, _ := os.OpenFile("/Users/jeeva/go-resty.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
//
//		resty.SetLogger(file)
//
func (c *Client) SetLogger(w io.Writer) *Client {
	c.Log = getLogger(w)
	return c
}

// SetContentLength method enables the HTTP header `Content-Length` value for every request.
// By default go-resty won't set `Content-Length`.
// 		resty.SetContentLength(true)
//
// Also you have an option to enable for particular request. See `resty.R().SetContentLength`
//
func (c *Client) SetContentLength(l bool) *Client {
	c.setContentLength = l
	return c
}

// SetTimeout method sets timeout for request raised from client.
//		resty.SetTimeout(time.Duration(1 * time.Minute))
//
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// SetError method is to register the global or client common `Error` object into go-resty.
// It is used for automatic unmarshalling if response status code is greater than 399 and
// content type either JSON or XML. Can be pointer or non-pointer.
// 		resty.SetError(&Error{})
//		// OR
//		resty.SetError(Error{})
//
func (c *Client) SetError(err interface{}) *Client {
	c.Error = typeOf(err)
	return c
}

// SetRedirectPolicy method sets the client redirect poilicy. go-resty provides ready to use
// redirect policies. Wanna create one for yourself refer `redirect.go`.
//
//		resty.SetRedirectPolicy(FlexibleRedirectPolicy(20))
//
// 		// Need multiple redirect policies together
//		resty.SetRedirectPolicy(FlexibleRedirectPolicy(20), DomainCheckRedirectPolicy("host1.com", "host2.net"))
//
func (c *Client) SetRedirectPolicy(policies ...interface{}) *Client {
	for _, p := range policies {
		if _, ok := p.(RedirectPolicy); !ok {
			c.Log.Printf("ERORR: %v does not implement resty.RedirectPolicy (missing Apply method)",
				functionName(p))
		}
	}

	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for _, p := range policies {
			if err := p.(RedirectPolicy).Apply(req, via); err != nil {
				return err
			}
		}
		return nil // looks good, go ahead
	}

	return c
}

// SetRetryCount method enables retry on `go-resty` client and allows you
// to set no. of retry count. Resty uses a Backoff mechanism.
func (c *Client) SetRetryCount(count int) *Client {
	c.RetryCount = count
	return c
}

// SetRetryWaitTime method sets default wait time to sleep before retrying
// request.
// Default is 100 milliseconds.
func (c *Client) SetRetryWaitTime(waitTime time.Duration) *Client {
	c.RetryWaitTime = waitTime
	return c
}

// SetRetryMaxWaitTime method sets max wait time to sleep before retrying
// request.
// Default is 2 seconds.
func (c *Client) SetRetryMaxWaitTime(maxWaitTime time.Duration) *Client {
	c.RetryMaxWaitTime = maxWaitTime
	return c
}

// AddRetryCondition method adds a retry condition function to array of functions
// that are checked to determine if the request is retried. The request will
// retry if any of the functions return true and error is nil.
func (c *Client) AddRetryCondition(condition RetryConditionFunc) *Client {
	c.RetryConditions = append(c.RetryConditions, condition)
	return c
}

// SetHTTPMode method sets go-resty mode to 'http'
func (c *Client) SetHTTPMode() *Client {
	return c.SetMode("http")
}

// SetRESTMode method sets go-resty mode to 'rest'
func (c *Client) SetRESTMode() *Client {
	return c.SetMode("rest")
}

// SetMode method sets go-resty client mode to given value such as 'http' & 'rest'.
//	'rest':
//		- No Redirect
//		- Automatic response unmarshal if it is JSON or XML
//	'http':
//		- Up to 10 Redirects
//		- No automatic unmarshall. Response will be treated as `response.String()`
//
// If you want more redirects, use FlexibleRedirectPolicy
//		resty.SetRedirectPolicy(FlexibleRedirectPolicy(20))
//
func (c *Client) SetMode(mode string) *Client {
	// HTTP
	if mode == "http" {
		c.isHTTPMode = true
		c.SetRedirectPolicy(FlexibleRedirectPolicy(10))
		c.afterResponse = []func(*Client, *Response) error{
			responseLogger,
			saveResponseIntoFile,
		}
		return c
	}

	// RESTful
	c.isHTTPMode = false
	c.SetRedirectPolicy(NoRedirectPolicy())
	c.afterResponse = []func(*Client, *Response) error{
		responseLogger,
		parseResponseBody,
		saveResponseIntoFile,
	}
	return c
}

// Mode method returns the current client mode. Typically its a "http" or "rest".
// Default is "rest"
func (c *Client) Mode() string {
	if c.isHTTPMode {
		return "http"
	}
	return "rest"
}

// SetTLSClientConfig method sets TLSClientConfig for underling client Transport.
//
// Example:
// 		// One can set custom root-certificate. Refer: http://golang.org/pkg/crypto/tls/#example_Dial
//		resty.SetTLSClientConfig(&tls.Config{ RootCAs: roots })
//
// 		// or One can disable security check (https)
//		resty.SetTLSClientConfig(&tls.Config{ InsecureSkipVerify: true })
// Note: This method overwrites existing `TLSClientConfig`.
//
func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	transport, err := c.getTransport()
	if err != nil {
		c.Log.Printf("ERROR %v", err)
		return c
	}
	transport.TLSClientConfig = config
	return c
}

// SetProxy method sets the Proxy URL and Port for resty client.
//		resty.SetProxy("http://proxyserver:8888")
//
// Alternatives: At request level proxy, see `Request.SetProxy`.  OR Without this `SetProxy` method,
// you can also set Proxy via environment variable. By default `Go` uses setting from `HTTP_PROXY`.
//
func (c *Client) SetProxy(proxyURL string) *Client {
	transport, err := c.getTransport()
	if err != nil {
		c.Log.Printf("ERROR %v", err)
		return c
	}

	if pURL, err := url.Parse(proxyURL); err == nil {
		c.proxyURL = pURL
		transport.Proxy = http.ProxyURL(c.proxyURL)
	} else {
		c.Log.Printf("ERROR %v", err)
		c.RemoveProxy()
	}
	return c
}

// RemoveProxy method removes the proxy configuration from resty client
//		resty.RemoveProxy()
//
func (c *Client) RemoveProxy() *Client {
	transport, err := c.getTransport()
	if err != nil {
		c.Log.Printf("ERROR %v", err)
		return c
	}
	c.proxyURL = nil
	transport.Proxy = nil
	return c
}

// SetCertificates method helps to set client certificates into resty conveniently.
//
func (c *Client) SetCertificates(certs ...tls.Certificate) *Client {
	config, err := c.getTLSConfig()
	if err != nil {
		c.Log.Printf("ERROR %v", err)
		return c
	}
	config.Certificates = append(config.Certificates, certs...)
	return c
}

// SetRootCertificate method helps to add one or more root certificates into resty client
// 		resty.SetRootCertificate("/path/to/root/pemFile.pem")
//
func (c *Client) SetRootCertificate(pemFilePath string) *Client {
	rootPemData, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		c.Log.Printf("ERROR %v", err)
		return c
	}

	config, err := c.getTLSConfig()
	if err != nil {
		c.Log.Printf("ERROR %v", err)
		return c
	}
	if config.RootCAs == nil {
		config.RootCAs = x509.NewCertPool()
	}

	config.RootCAs.AppendCertsFromPEM(rootPemData)

	return c
}

// SetOutputDirectory method sets output directory for saving HTTP response into file.
// If the output directory not exists then resty creates one. This setting is optional one,
// if you're planning using absoule path in `Request.SetOutput` and can used together.
// 		resty.SetOutputDirectory("/save/http/response/here")
//
func (c *Client) SetOutputDirectory(dirPath string) *Client {
	c.outputDirectory = dirPath
	return c
}

// SetTransport method sets custom `*http.Transport` or any `http.RoundTripper`
// compatible interface implementation in the resty client.
//
// NOTE:
//
// - If transport is not type of `*http.Transport` then you may not be able to
// take advantage of some of the `resty` client settings.
//
// - It overwrites the resty client transport instance and it's configurations.
//
//		transport := &http.Transport{
//			// somthing like Proxying to httptest.Server, etc...
//			Proxy: func(req *http.Request) (*url.URL, error) {
//				return url.Parse(server.URL)
//			},
//		}
//
//		resty.SetTransport(transport)
//
func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	if transport != nil {
		c.httpClient.Transport = transport
	}
	return c
}

// SetScheme method sets custom scheme in the resty client. It's way to override default.
// 		resty.SetScheme("http")
//
func (c *Client) SetScheme(scheme string) *Client {
	if !IsStringEmpty(scheme) {
		c.scheme = scheme
	}

	return c
}

// SetCloseConnection method sets variable `Close` in http request struct with the given
// value. More info: https://golang.org/src/net/http/request.go
func (c *Client) SetCloseConnection(close bool) *Client {
	c.closeConnection = close
	return c
}

// SetDoNotParseResponse method instructs `Resty` not to parse the response body automatically.
// Resty exposes the raw response body as `io.ReadCloser`. Also do not forget to close the body,
// otherwise you might get into connection leaks, no connection reuse.
//
// Please Note: Response middlewares are not applicable, if you use this option. Basically you have
// taken over the control of response parsing from `Resty`.
func (c *Client) SetDoNotParseResponse(parse bool) *Client {
	c.notParseResponse = parse
	return c
}

// SetLogPrefix method sets the Resty logger prefix value.
func (c *Client) SetLogPrefix(prefix string) *Client {
	c.logPrefix = prefix
	c.Log.SetPrefix(prefix)
	return c
}

// SetPathParams method sets multiple URL path key-value pairs at one go in the
// resty client instance.
// 		resty.SetPathParams(map[string]string{
// 		   "userId": "sample@sample.com",
// 		   "subAccountId": "100002",
// 		})
//
// 		Result:
// 		   URL - /v1/users/{userId}/{subAccountId}/details
// 		   Composed URL - /v1/users/sample@sample.com/100002/details
// It replace the value of the key while composing request URL. Also it can be
// overridden at request level Path Params options, see `Request.SetPathParams`.
func (c *Client) SetPathParams(params map[string]string) *Client {
	for p, v := range params {
		c.pathParams[p] = v
	}
	return c
}

// SetJSONEscapeHTML method is to enable/disable the HTML escape on JSON marshal.
//
// NOTE: This option only applicable to standard JSON Marshaller.
func (c *Client) SetJSONEscapeHTML(b bool) *Client {
	c.jsonEscapeHTML = b
	return c
}

// IsProxySet method returns the true if proxy is set on client otherwise false.
func (c *Client) IsProxySet() bool {
	return c.proxyURL != nil
}

// GetClient method returns the current http.Client used by the resty client.
func (c *Client) GetClient() *http.Client {
	return c.httpClient
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Client Unexported methods
//___________________________________

// executes the given `Request` object and returns response
func (c *Client) execute(req *Request) (*Response, error) {
	defer releaseBuffer(req.bodyBuf)
	// Apply Request middleware
	var err error

	// user defined on before request methods
	// to modify the *resty.Request object
	for _, f := range c.udBeforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	// resty middlewares
	for _, f := range c.beforeRequest {
		if err = f(c, req); err != nil {
			return nil, err
		}
	}

	// call pre-request if defined
	if c.preReqHook != nil {
		if err = c.preReqHook(c, req); err != nil {
			return nil, err
		}
	}

	if hostHeader := req.Header.Get("Host"); hostHeader != "" {
		req.RawRequest.Host = hostHeader
	}

	req.Time = time.Now()
	resp, err := c.httpClient.Do(req.RawRequest)

	response := &Response{
		Request:     req,
		RawResponse: resp,
		receivedAt:  time.Now(),
	}

	if err != nil || req.notParseResponse || c.notParseResponse {
		return response, err
	}

	if !req.isSaveResponse {
		defer closeq(resp.Body)
		body := resp.Body

		// GitHub #142
		if strings.EqualFold(resp.Header.Get(hdrContentEncodingKey), "gzip") && resp.ContentLength > 0 {
			if _, ok := body.(*gzip.Reader); !ok {
				body, err = gzip.NewReader(body)
				if err != nil {
					return response, err
				}
				defer closeq(body)
			}
		}

		if response.body, err = ioutil.ReadAll(body); err != nil {
			return response, err
		}

		response.size = int64(len(response.body))
	}

	// Apply Response middleware
	for _, f := range c.afterResponse {
		if err = f(c, response); err != nil {
			break
		}
	}

	return response, err
}

// enables a log prefix
func (c *Client) enableLogPrefix() {
	c.Log.SetFlags(log.LstdFlags)
	c.Log.SetPrefix(c.logPrefix)
}

// disables a log prefix
func (c *Client) disableLogPrefix() {
	c.Log.SetFlags(0)
	c.Log.SetPrefix("")
}

// getting TLS client config if not exists then create one
func (c *Client) getTLSConfig() (*tls.Config, error) {
	transport, err := c.getTransport()
	if err != nil {
		return nil, err
	}
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	return transport.TLSClientConfig, nil
}

// returns `*http.Transport` currently in use or error
// in case currently used `transport` is not an `*http.Transport`
func (c *Client) getTransport() (*http.Transport, error) {
	if c.httpClient.Transport == nil {
		c.SetTransport(new(http.Transport))
	}

	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		return transport, nil
	}
	return nil, errors.New("current transport is not an *http.Transport instance")
}

//
// File
//

// File represent file information for multipart request
type File struct {
	Name      string
	ParamName string
	io.Reader
}

// String returns string value of current file details
func (f *File) String() string {
	return fmt.Sprintf("ParamName: %v; FileName: %v", f.ParamName, f.Name)
}

// multipartField represent custom data part for multipart request
type multipartField struct {
	Param       string
	FileName    string
	ContentType string
	io.Reader
}
