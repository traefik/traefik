<p align="center">
<h1 align="center">Resty</h1>
<p align="center">Simple HTTP and REST client library for Go (inspired by Ruby rest-client)</p>
<p align="center"><a href="#features">Features</a> section describes in detail about Resty capabilities</p>
</p>
<p align="center">
<p align="center"><a href="https://travis-ci.org/go-resty/resty"><img src="https://travis-ci.org/go-resty/resty.svg?branch=master" alt="Build Status"></a> <a href="https://codecov.io/gh/go-resty/resty/branch/master"><img src="https://codecov.io/gh/go-resty/resty/branch/master/graph/badge.svg" alt="Code Coverage"></a> <a href="https://goreportcard.com/report/go-resty/resty"><img src="https://goreportcard.com/badge/go-resty/resty" alt="Go Report Card"></a> <a href="https://github.com/go-resty/resty/releases/latest"><img src="https://img.shields.io/badge/version-1.9.1-blue.svg" alt="Release Version"></a> <a href="https://godoc.org/github.com/go-resty/resty"><img src="https://godoc.org/github.com/go-resty/resty?status.svg" alt="GoDoc"></a> <a href="LICENSE"><img src="https://img.shields.io/github/license/go-resty/resty.svg" alt="License"></a></p>
</p>

## News

  * [Collecting Inputs for Resty v2.0.0](https://github.com/go-resty/resty/issues/166)
  * v1.9.1 [released](https://github.com/go-resty/resty/releases/latest) and tagged on Aug 29, 2018.
  * v1.0 released - Resty's first version was released on Sep 15, 2015 then it grew gradually as a very handy and helpful library. Its been a two years; `v1.0` was released on Sep 25, 2017. I'm very thankful to Resty users and its [contributors](https://github.com/go-resty/resty/graphs/contributors).

## Features

  * GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS, etc.
  * Simple and chainable methods for settings and request
  * Request Body can be `string`, `[]byte`, `struct`, `map`, `slice` and `io.Reader` too
    * Auto detects `Content-Type`
    * Buffer less processing for `io.Reader`
  * [Response](https://godoc.org/github.com/go-resty/resty#Response) object gives you more possibility
    * Access as `[]byte` array - `response.Body()` OR Access as `string` - `response.String()`
    * Know your `response.Time()` and when we `response.ReceivedAt()`
  * Automatic marshal and unmarshal for `JSON` and `XML` content type
    * Default is `JSON`, if you supply `struct/map` without header `Content-Type`
    * For auto-unmarshal, refer to -
        - Success scenario [Request.SetResult()](https://godoc.org/github.com/go-resty/resty#Request.SetResult) and [Response.Result()](https://godoc.org/github.com/go-resty/resty#Response.Result).
        - Error scenario [Request.SetError()](https://godoc.org/github.com/go-resty/resty#Request.SetError) and [Response.Error()](https://godoc.org/github.com/go-resty/resty#Response.Error).
        - Supports [RFC7807](https://tools.ietf.org/html/rfc7807) - `application/problem+json` & `application/problem+xml`
  * Easy to upload one or more file(s) via `multipart/form-data`
    * Auto detects file content type
  * Request URL [Path Params (aka URI Params)](https://godoc.org/github.com/go-resty/resty#Request.SetPathParams)
  * Backoff Retry Mechanism with retry condition function [reference](retry_test.go)
  * resty client HTTP & REST [Request](https://godoc.org/github.com/go-resty/resty#Client.OnBeforeRequest) and [Response](https://godoc.org/github.com/go-resty/resty#Client.OnAfterResponse) middlewares
  * `Request.SetContext` supported `go1.7` and above
  * Authorization option of `BasicAuth` and `Bearer` token
  * Set request `ContentLength` value for all request or particular request
  * Choose between HTTP and REST mode. Default is `REST`
    * `HTTP` - default up to 10 redirects and no automatic response unmarshal
    * `REST` - defaults to no redirects and automatic response marshal/unmarshal for `JSON` & `XML`
  * Custom [Root Certificates](https://godoc.org/github.com/go-resty/resty#Client.SetRootCertificate) and Client [Certificates](https://godoc.org/github.com/go-resty/resty#Client.SetCertificates)
  * Download/Save HTTP response directly into File, like `curl -o` flag. See [SetOutputDirectory](https://godoc.org/github.com/go-resty/resty#Client.SetOutputDirectory) & [SetOutput](https://godoc.org/github.com/go-resty/resty#Request.SetOutput).
  * Cookies for your request and CookieJar support
  * SRV Record based request instead of Host URL
  * Client settings like `Timeout`, `RedirectPolicy`, `Proxy`, `TLSClientConfig`, `Transport`, etc.
  * Optionally allows GET request with payload, see [SetAllowGetMethodPayload](https://godoc.org/github.com/go-resty/resty#Client.SetAllowGetMethodPayload)
  * Supports registering external JSON library into resty, see [how to use](https://github.com/go-resty/resty/issues/76#issuecomment-314015250)
  * Exposes Response reader without reading response (no auto-unmarshaling) if need be, see [how to use](https://github.com/go-resty/resty/issues/87#issuecomment-322100604)
  * Option to specify expected `Content-Type` when response `Content-Type` header missing. Refer to [#92](https://github.com/go-resty/resty/issues/92)
  * Resty design
    * Have client level settings & options and also override at Request level if you want to
    * Request and Response middlewares
    * Create Multiple clients if you want to `resty.New()`
    * Supports `http.RoundTripper` implementation, see [SetTransport](https://godoc.org/github.com/go-resty/resty#Client.SetTransport)
    * goroutine concurrent safe
    * REST and HTTP modes
    * Debug mode - clean and informative logging presentation
    * Gzip - Go does it automatically also resty has fallback handling too
    * Works fine with `HTTP/2` and `HTTP/1.1`
  * [Bazel support](#bazel-support)
  * Easily mock resty for testing, [for e.g.](#mocking-http-requests-using-httpmock-library)
  * Well tested client library

Resty works with `go1.3` and above.

### Included Batteries

  * Redirect Policies - see [how to use](#redirect-policy)
    * NoRedirectPolicy
    * FlexibleRedirectPolicy
    * DomainCheckRedirectPolicy
    * etc. [more info](redirect.go)
  * Retry Mechanism [how to use](#retries)
    * Backoff Retry
    * Conditional Retry
  * SRV Record based request instead of Host URL [how to use](resty_test.go#L1412)
  * etc (upcoming - throw your idea's [here](https://github.com/go-resty/resty/issues)).

## Installation

#### Stable Version - Production Ready

Please refer section [Versioning](#versioning) for detailed info.

##### go.mod

```bash
require gopkg.in/resty.v1 v1.9.1
```

##### go get
```bash
go get -u gopkg.in/resty.v1
```

#### Latest Version - Development Edge

```bash
# install the latest & greatest library
go get -u github.com/go-resty/resty
```

## It might be beneficial for your project :smile:

Resty author also published following projects for Go Community.

  * [aah framework](https://aahframework.org) - A secure, flexible, rapid Go web framework.
  * [go-model](https://github.com/jeevatkm/go-model) - Robust & Easy to use model mapper and utility methods for Go `struct`.

## Usage

The following samples will assist you to become as comfortable as possible with resty library. Resty comes with ready to use DefaultClient.

Import resty into your code and refer it as `resty`.

```go
import "gopkg.in/resty.v1"
```

#### Simple GET

```go
// GET request
resp, err := resty.R().Get("http://httpbin.org/get")

// explore response object
fmt.Printf("\nError: %v", err)
fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
fmt.Printf("\nResponse Status: %v", resp.Status())
fmt.Printf("\nResponse Time: %v", resp.Time())
fmt.Printf("\nResponse Received At: %v", resp.ReceivedAt())
fmt.Printf("\nResponse Body: %v", resp)     // or resp.String() or string(resp.Body())
// more...

/* Output
Error: <nil>
Response Status Code: 200
Response Status: 200 OK
Response Time: 644.290186ms
Response Received At: 2015-09-15 12:05:28.922780103 -0700 PDT
Response Body: {
  "args": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "go-resty v0.1 - https://github.com/go-resty/resty"
  },
  "origin": "0.0.0.0",
  "url": "http://httpbin.org/get"
}
*/
```

#### Enhanced GET

```go
resp, err := resty.R().
      SetQueryParams(map[string]string{
          "page_no": "1",
          "limit": "20",
          "sort":"name",
          "order": "asc",
          "random":strconv.FormatInt(time.Now().Unix(), 10),
      }).
      SetHeader("Accept", "application/json").
      SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F").
      Get("/search_result")


// Sample of using Request.SetQueryString method
resp, err := resty.R().
      SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more").
      SetHeader("Accept", "application/json").
      SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F").
      Get("/show_product")
```

#### Various POST method combinations

```go
// POST JSON string
// No need to set content type, if you have client level setting
resp, err := resty.R().
      SetHeader("Content-Type", "application/json").
      SetBody(`{"username":"testuser", "password":"testpass"}`).
      SetResult(&AuthSuccess{}).    // or SetResult(AuthSuccess{}).
      Post("https://myapp.com/login")

// POST []byte array
// No need to set content type, if you have client level setting
resp, err := resty.R().
      SetHeader("Content-Type", "application/json").
      SetBody([]byte(`{"username":"testuser", "password":"testpass"}`)).
      SetResult(&AuthSuccess{}).    // or SetResult(AuthSuccess{}).
      Post("https://myapp.com/login")

// POST Struct, default is JSON content type. No need to set one
resp, err := resty.R().
      SetBody(User{Username: "testuser", Password: "testpass"}).
      SetResult(&AuthSuccess{}).    // or SetResult(AuthSuccess{}).
      SetError(&AuthError{}).       // or SetError(AuthError{}).
      Post("https://myapp.com/login")

// POST Map, default is JSON content type. No need to set one
resp, err := resty.R().
      SetBody(map[string]interface{}{"username": "testuser", "password": "testpass"}).
      SetResult(&AuthSuccess{}).    // or SetResult(AuthSuccess{}).
      SetError(&AuthError{}).       // or SetError(AuthError{}).
      Post("https://myapp.com/login")

// POST of raw bytes for file upload. For example: upload file to Dropbox
fileBytes, _ := ioutil.ReadFile("/Users/jeeva/mydocument.pdf")

// See we are not setting content-type header, since go-resty automatically detects Content-Type for you
resp, err := resty.R().
      SetBody(fileBytes).
      SetContentLength(true).          // Dropbox expects this value
      SetAuthToken("<your-auth-token>").
      SetError(&DropboxError{}).       // or SetError(DropboxError{}).
      Post("https://content.dropboxapi.com/1/files_put/auto/resty/mydocument.pdf") // for upload Dropbox supports PUT too

// Note: resty detects Content-Type for request body/payload if content type header is not set.
//   * For struct and map data type defaults to 'application/json'
//   * Fallback is plain text content type
```

#### Sample PUT

You can use various combinations of `PUT` method call like demonstrated for `POST`.

```go
// Note: This is one sample of PUT method usage, refer POST for more combination

// Request goes as JSON content type
// No need to set auth token, error, if you have client level settings
resp, err := resty.R().
      SetBody(Article{
        Title: "go-resty",
        Content: "This is my article content, oh ya!",
        Author: "Jeevanandam M",
        Tags: []string{"article", "sample", "resty"},
      }).
      SetAuthToken("C6A79608-782F-4ED0-A11D-BD82FAD829CD").
      SetError(&Error{}).       // or SetError(Error{}).
      Put("https://myapp.com/article/1234")
```

#### Sample PATCH

You can use various combinations of `PATCH` method call like demonstrated for `POST`.

```go
// Note: This is one sample of PUT method usage, refer POST for more combination

// Request goes as JSON content type
// No need to set auth token, error, if you have client level settings
resp, err := resty.R().
      SetBody(Article{
        Tags: []string{"new tag1", "new tag2"},
      }).
      SetAuthToken("C6A79608-782F-4ED0-A11D-BD82FAD829CD").
      SetError(&Error{}).       // or SetError(Error{}).
      Patch("https://myapp.com/articles/1234")
```

#### Sample DELETE, HEAD, OPTIONS

```go
// DELETE a article
// No need to set auth token, error, if you have client level settings
resp, err := resty.R().
      SetAuthToken("C6A79608-782F-4ED0-A11D-BD82FAD829CD").
      SetError(&Error{}).       // or SetError(Error{}).
      Delete("https://myapp.com/articles/1234")

// DELETE a articles with payload/body as a JSON string
// No need to set auth token, error, if you have client level settings
resp, err := resty.R().
      SetAuthToken("C6A79608-782F-4ED0-A11D-BD82FAD829CD").
      SetError(&Error{}).       // or SetError(Error{}).
      SetHeader("Content-Type", "application/json").
      SetBody(`{article_ids: [1002, 1006, 1007, 87683, 45432] }`).
      Delete("https://myapp.com/articles")

// HEAD of resource
// No need to set auth token, if you have client level settings
resp, err := resty.R().
      SetAuthToken("C6A79608-782F-4ED0-A11D-BD82FAD829CD").
      Head("https://myapp.com/videos/hi-res-video")

// OPTIONS of resource
// No need to set auth token, if you have client level settings
resp, err := resty.R().
      SetAuthToken("C6A79608-782F-4ED0-A11D-BD82FAD829CD").
      Options("https://myapp.com/servers/nyc-dc-01")
```

### Multipart File(s) upload

#### Using io.Reader

```go
profileImgBytes, _ := ioutil.ReadFile("/Users/jeeva/test-img.png")
notesBytes, _ := ioutil.ReadFile("/Users/jeeva/text-file.txt")

resp, err := dclr().
      SetFileReader("profile_img", "test-img.png", bytes.NewReader(profileImgBytes)).
      SetFileReader("notes", "text-file.txt", bytes.NewReader(notesBytes)).
      SetFormData(map[string]string{
          "first_name": "Jeevanandam",
          "last_name": "M",
      }).
      Post(t"http://myapp.com/upload")
```

#### Using File directly from Path

```go
// Single file scenario
resp, err := resty.R().
      SetFile("profile_img", "/Users/jeeva/test-img.png").
      Post("http://myapp.com/upload")

// Multiple files scenario
resp, err := resty.R().
      SetFiles(map[string]string{
        "profile_img": "/Users/jeeva/test-img.png",
        "notes": "/Users/jeeva/text-file.txt",
      }).
      Post("http://myapp.com/upload")

// Multipart of form fields and files
resp, err := resty.R().
      SetFiles(map[string]string{
        "profile_img": "/Users/jeeva/test-img.png",
        "notes": "/Users/jeeva/text-file.txt",
      }).
      SetFormData(map[string]string{
        "first_name": "Jeevanandam",
        "last_name": "M",
        "zip_code": "00001",
        "city": "my city",
        "access_token": "C6A79608-782F-4ED0-A11D-BD82FAD829CD",
      }).
      Post("http://myapp.com/profile")
```

#### Sample Form submission

```go
// just mentioning about POST as an example with simple flow
// User Login
resp, err := resty.R().
      SetFormData(map[string]string{
        "username": "jeeva",
        "password": "mypass",
      }).
      Post("http://myapp.com/login")

// Followed by profile update
resp, err := resty.R().
      SetFormData(map[string]string{
        "first_name": "Jeevanandam",
        "last_name": "M",
        "zip_code": "00001",
        "city": "new city update",
      }).
      Post("http://myapp.com/profile")

// Multi value form data
criteria := url.Values{
  "search_criteria": []string{"book", "glass", "pencil"},
}
resp, err := resty.R().
      SetMultiValueFormData(criteria).
      Post("http://myapp.com/search")
```

#### Save HTTP Response into File

```go
// Setting output directory path, If directory not exists then resty creates one!
// This is optional one, if you're planning using absoule path in
// `Request.SetOutput` and can used together.
resty.SetOutputDirectory("/Users/jeeva/Downloads")

// HTTP response gets saved into file, similar to curl -o flag
_, err := resty.R().
          SetOutput("plugin/ReplyWithHeader-v5.1-beta.zip").
          Get("http://bit.ly/1LouEKr")

// OR using absolute path
// Note: output directory path is not used for absoulte path
_, err := resty.R().
          SetOutput("/MyDownloads/plugin/ReplyWithHeader-v5.1-beta.zip").
          Get("http://bit.ly/1LouEKr")
```

#### Request URL Path Params

Resty provides easy to use dynamic request URL path params. Params can be set at client and request level. Client level params value can be overridden at request level.

```go
resty.R().SetPathParams(map[string]string{
   "userId": "sample@sample.com",
   "subAccountId": "100002",
}).
Get("/v1/users/{userId}/{subAccountId}/details")

// Result:
//   Composed URL - /v1/users/sample@sample.com/100002/details
```

#### Request and Response Middleware

Resty provides middleware ability to manipulate for Request and Response. It is more flexible than callback approach.

```go
// Registering Request Middleware
resty.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
    // Now you have access to Client and current Request object
    // manipulate it as per your need

    return nil  // if its success otherwise return error
  })

// Registering Response Middleware
resty.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
    // Now you have access to Client and current Response object
    // manipulate it as per your need

    return nil  // if its success otherwise return error
  })
```

#### Redirect Policy

Resty provides few ready to use redirect policy(s) also it supports multiple policies together.

```go
// Assign Client Redirect Policy. Create one as per you need
resty.SetRedirectPolicy(resty.FlexibleRedirectPolicy(15))

// Wanna multiple policies such as redirect count, domain name check, etc
resty.SetRedirectPolicy(resty.FlexibleRedirectPolicy(20),
                        resty.DomainCheckRedirectPolicy("host1.com", "host2.org", "host3.net"))
```

##### Custom Redirect Policy

Implement [RedirectPolicy](redirect.go#L20) interface and register it with resty client. Have a look [redirect.go](redirect.go) for more information.

```go
// Using raw func into resty.SetRedirectPolicy
resty.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
  // Implement your logic here

  // return nil for continue redirect otherwise return error to stop/prevent redirect
  return nil
}))

//---------------------------------------------------

// Using struct create more flexible redirect policy
type CustomRedirectPolicy struct {
  // variables goes here
}

func (c *CustomRedirectPolicy) Apply(req *http.Request, via []*http.Request) error {
  // Implement your logic here

  // return nil for continue redirect otherwise return error to stop/prevent redirect
  return nil
}

// Registering in resty
resty.SetRedirectPolicy(CustomRedirectPolicy{/* initialize variables */})
```

#### Custom Root Certificates and Client Certificates

```go
// Custom Root certificates, just supply .pem file.
// you can add one or more root certificates, its get appended
resty.SetRootCertificate("/path/to/root/pemFile1.pem")
resty.SetRootCertificate("/path/to/root/pemFile2.pem")
// ... and so on!

// Adding Client Certificates, you add one or more certificates
// Sample for creating certificate object
// Parsing public/private key pair from a pair of files. The files must contain PEM encoded data.
cert1, err := tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
if err != nil {
  log.Fatalf("ERROR client certificate: %s", err)
}
// ...

// You add one or more certificates
resty.SetCertificates(cert1, cert2, cert3)
```

#### Proxy Settings - Client as well as at Request Level

Default `Go` supports Proxy via environment variable `HTTP_PROXY`. Resty provides support via `SetProxy` & `RemoveProxy`.
Choose as per your need.

**Client Level Proxy** settings applied to all the request

```go
// Setting a Proxy URL and Port
resty.SetProxy("http://proxyserver:8888")

// Want to remove proxy setting
resty.RemoveProxy()
```

#### Retries

Resty uses [backoff](http://www.awsarchitectureblog.com/2015/03/backoff.html)
to increase retry intervals after each attempt.

Usage example:

```go
// Retries are configured per client
resty.
    // Set retry count to non zero to enable retries
    SetRetryCount(3).
    // You can override initial retry wait time.
    // Default is 100 milliseconds.
    SetRetryWaitTime(5 * time.Second).
    // MaxWaitTime can be overridden as well.
    // Default is 2 seconds.
    SetRetryMaxWaitTime(20 * time.Second)
```

Above setup will result in resty retrying requests returned non nil error up to
3 times with delay increased after each attempt.

You can optionally provide client with custom retry conditions:

```go
resty.AddRetryCondition(
    // Condition function will be provided with *resty.Response as a
    // parameter. It is expected to return (bool, error) pair. Resty will retry
    // in case condition returns true or non nil error.
    func(r *resty.Response) (bool, error) {
        return r.StatusCode() == http.StatusTooManyRequests, nil
    },
)
```

Above example will make resty retry requests ended with `429 Too Many Requests`
status code.

Multiple retry conditions can be added.

It is also possible to use `resty.Backoff(...)` to get arbitrary retry scenarios
implemented. [Reference](retry_test.go).

#### Choose REST or HTTP mode

```go
// REST mode. This is Default.
resty.SetRESTMode()

// HTTP mode
resty.SetHTTPMode()
```

#### Allow GET request with Payload

```go
// Allow GET request with Payload. This is disabled by default.
resty.SetAllowGetMethodPayload(true)
```

#### Wanna Multiple Clients

```go
// Here you go!
// Client 1
client1 := resty.New()
client1.R().Get("http://httpbin.org")
// ...

// Client 2
client2 := resty.New()
client2.R().Head("http://httpbin.org")
// ...

// Bend it as per your need!!!
```

#### Remaining Client Settings & its Options

```go
// Unique settings at Client level
//--------------------------------
// Enable debug mode
resty.SetDebug(true)

// Using you custom log writer
logFile, _ := os.OpenFile("/Users/jeeva/go-resty.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
resty.SetLogger(logFile)

// Assign Client TLSClientConfig
// One can set custom root-certificate. Refer: http://golang.org/pkg/crypto/tls/#example_Dial
resty.SetTLSClientConfig(&tls.Config{ RootCAs: roots })

// or One can disable security check (https)
resty.SetTLSClientConfig(&tls.Config{ InsecureSkipVerify: true })

// Set client timeout as per your need
resty.SetTimeout(1 * time.Minute)


// You can override all below settings and options at request level if you want to
//--------------------------------------------------------------------------------
// Host URL for all request. So you can use relative URL in the request
resty.SetHostURL("http://httpbin.org")

// Headers for all request
resty.SetHeader("Accept", "application/json")
resty.SetHeaders(map[string]string{
        "Content-Type": "application/json",
        "User-Agent": "My custom User Agent String",
      })

// Cookies for all request
resty.SetCookie(&http.Cookie{
      Name:"go-resty",
      Value:"This is cookie value",
      Path: "/",
      Domain: "sample.com",
      MaxAge: 36000,
      HttpOnly: true,
      Secure: false,
    })
resty.SetCookies(cookies)

// URL query parameters for all request
resty.SetQueryParam("user_id", "00001")
resty.SetQueryParams(map[string]string{ // sample of those who use this manner
      "api_key": "api-key-here",
      "api_secert": "api-secert",
    })
resty.R().SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more")

// Form data for all request. Typically used with POST and PUT
resty.SetFormData(map[string]string{
    "access_token": "BC594900-518B-4F7E-AC75-BD37F019E08F",
  })

// Basic Auth for all request
resty.SetBasicAuth("myuser", "mypass")

// Bearer Auth Token for all request
resty.SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F")

// Enabling Content length value for all request
resty.SetContentLength(true)

// Registering global Error object structure for JSON/XML request
resty.SetError(&Error{})    // or resty.SetError(Error{})
```

#### Unix Socket

```go
unixSocket := "/var/run/my_socket.sock"

// Create a Go's http.Transport so we can set it in resty.
transport := http.Transport{
	Dial: func(_, _ string) (net.Conn, error) {
		return net.Dial("unix", unixSocket)
	},
}

// Set the previous transport that we created, set the scheme of the communication to the
// socket and set the unixSocket as the HostURL.
r := resty.New().SetTransport(&transport).SetScheme("http").SetHostURL(unixSocket)

// No need to write the host's URL on the request, just the path.
r.R().Get("/index.html")
```

#### Bazel support

Resty can be built, tested and depended upon via [Bazel](https://bazel.build).
For example, to run all tests:

```shell
bazel test :go_default_test
```

#### Mocking http requests using [httpmock](https://github.com/jarcoal/httpmock) library

In order to mock the http requests when testing your application you
could use the `httpmock` library.

When using the default resty client, you should pass the client to the library as follow:

```go
httpmock.ActivateNonDefault(resty.DefaultClient.GetClient())
```

More detailed example of mocking resty http requests using ginko could be found [here](https://github.com/jarcoal/httpmock#ginkgo--resty-example).

## Versioning

resty releases versions according to [Semantic Versioning](http://semver.org)

  * `gopkg.in/resty.vX` points to appropriate tagged versions; `X` denotes version series number and it's a stable release for production use. For e.g. `gopkg.in/resty.v0`.
  * Development takes place at the master branch. Although the code in master should always compile and test successfully, it might break API's. I aim to maintain backwards compatibility, but sometimes API's and behavior might be changed to fix a bug.

## Contribution

I would welcome your contribution! If you find any improvement or issue you want to fix, feel free to send a pull request, I like pull requests that include test cases for fix/enhancement. I have done my best to bring pretty good code coverage. Feel free to write tests.

BTW, I'd like to know what you think about `Resty`. Kindly open an issue or send me an email; it'd mean a lot to me.

## Creator

[Jeevanandam M.](https://github.com/jeevatkm) (jeeva@myjeeva.com)

## Contributors

Have a look on [Contributors](https://github.com/go-resty/resty/graphs/contributors) page.

## License

Resty released under MIT license, refer [LICENSE](LICENSE) file.
