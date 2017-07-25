// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the GitHub client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
	client.UploadURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

// openTestFile creates a new file with the given name and content for testing.
// In order to ensure the exact file name, this function will create a new temp
// directory, and create the file in that directory. It is the caller's
// responsibility to remove the directory and its contents when no longer needed.
func openTestFile(name, content string) (file *os.File, dir string, err error) {
	dir, err = ioutil.TempDir("", "go-github")
	if err != nil {
		return nil, dir, err
	}

	file, err = os.OpenFile(path.Join(dir, name), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, dir, err
	}

	fmt.Fprint(file, content)

	// close and re-open the file to keep file.Stat() happy
	file.Close()
	file, err = os.Open(file.Name())
	if err != nil {
		return nil, dir, err
	}

	return file, dir, err
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Set(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

func testURLParseError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func testBody(t *testing.T, r *http.Request, want string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Error reading request body: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("request Body is %s, want %s", got, want)
	}
}

// Helper function to test that a value is marshalled to JSON as expected.
func testJSONMarshal(t *testing.T, v interface{}, want string) {
	j, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %v", v)
	}

	w := new(bytes.Buffer)
	err = json.Compact(w, []byte(want))
	if err != nil {
		t.Errorf("String is not valid json: %s", want)
	}

	if w.String() != string(j) {
		t.Errorf("json.Marshal(%q) returned %s, want %s", v, j, w)
	}

	// now go the other direction and make sure things unmarshal as expected
	u := reflect.ValueOf(v).Interface()
	if err := json.Unmarshal([]byte(want), u); err != nil {
		t.Errorf("Unable to unmarshal JSON for %v", want)
	}

	if !reflect.DeepEqual(v, u) {
		t.Errorf("json.Unmarshal(%q) returned %s, want %s", want, u, v)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)

	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}
}

// Ensure that length of Client.rateLimits is the same as number of fields in RateLimits struct.
func TestClient_rateLimits(t *testing.T) {
	if got, want := len(Client{}.rateLimits), reflect.TypeOf(RateLimits{}).NumField(); got != want {
		t.Errorf("len(Client{}.rateLimits) is %v, want %v", got, want)
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
	inBody, outBody := &User{Login: String("l")}, `{"login":"l"}`+"\n"
	req, _ := c.NewRequest("GET", inURL, inBody)

	// test that relative URL was expanded
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
	}

	// test that body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("NewRequest(%q) Body is %v, want %v", inBody, got, want)
	}

	// test that default user-agent is attached to the request
	if got, want := req.Header.Get("User-Agent"), c.UserAgent; got != want {
		t.Errorf("NewRequest() User-Agent is %v, want %v", got, want)
	}
}

func TestNewRequest_invalidJSON(t *testing.T) {
	c := NewClient(nil)

	type T struct {
		A map[interface{}]interface{}
	}
	_, err := c.NewRequest("GET", "/", &T{})

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("Expected a JSON error; got %#v.", err)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c := NewClient(nil)
	_, err := c.NewRequest("GET", ":", nil)
	testURLParseError(t, err)
}

// ensure that no User-Agent header is set if the client's UserAgent is empty.
// This caused a problem with Google's internal http client.
func TestNewRequest_emptyUserAgent(t *testing.T) {
	c := NewClient(nil)
	c.UserAgent = ""
	req, err := c.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if _, ok := req.Header["User-Agent"]; ok {
		t.Fatal("constructed request contains unexpected User-Agent header")
	}
}

// If a nil body is passed to github.NewRequest, make sure that nil is also
// passed to http.NewRequest. In most cases, passing an io.Reader that returns
// no content is fine, since there is no difference between an HTTP request
// body that is an empty string versus one that is not set at all. However in
// certain cases, intermediate systems may treat these differently resulting in
// subtle errors.
func TestNewRequest_emptyBody(t *testing.T) {
	c := NewClient(nil)
	req, err := c.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if req.Body != nil {
		t.Fatalf("constructed request contains a non-nil Body")
	}
}

func TestResponse_populatePageValues(t *testing.T) {
	r := http.Response{
		Header: http.Header{
			"Link": {`<https://api.github.com/?page=1>; rel="first",` +
				` <https://api.github.com/?page=2>; rel="prev",` +
				` <https://api.github.com/?page=4>; rel="next",` +
				` <https://api.github.com/?page=5>; rel="last"`,
			},
		},
	}

	response := newResponse(&r)
	if got, want := response.FirstPage, 1; got != want {
		t.Errorf("response.FirstPage: %v, want %v", got, want)
	}
	if got, want := response.PrevPage, 2; want != got {
		t.Errorf("response.PrevPage: %v, want %v", got, want)
	}
	if got, want := response.NextPage, 4; want != got {
		t.Errorf("response.NextPage: %v, want %v", got, want)
	}
	if got, want := response.LastPage, 5; want != got {
		t.Errorf("response.LastPage: %v, want %v", got, want)
	}
}

func TestResponse_populatePageValues_invalid(t *testing.T) {
	r := http.Response{
		Header: http.Header{
			"Link": {`<https://api.github.com/?page=1>,` +
				`<https://api.github.com/?page=abc>; rel="first",` +
				`https://api.github.com/?page=2; rel="prev",` +
				`<https://api.github.com/>; rel="next",` +
				`<https://api.github.com/?page=>; rel="last"`,
			},
		},
	}

	response := newResponse(&r)
	if got, want := response.FirstPage, 0; got != want {
		t.Errorf("response.FirstPage: %v, want %v", got, want)
	}
	if got, want := response.PrevPage, 0; got != want {
		t.Errorf("response.PrevPage: %v, want %v", got, want)
	}
	if got, want := response.NextPage, 0; got != want {
		t.Errorf("response.NextPage: %v, want %v", got, want)
	}
	if got, want := response.LastPage, 0; got != want {
		t.Errorf("response.LastPage: %v, want %v", got, want)
	}

	// more invalid URLs
	r = http.Response{
		Header: http.Header{
			"Link": {`<https://api.github.com/%?page=2>; rel="first"`},
		},
	}

	response = newResponse(&r)
	if got, want := response.FirstPage, 0; got != want {
		t.Errorf("response.FirstPage: %v, want %v", got, want)
	}
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	body := new(foo)
	client.Do(context.Background(), req, body)

	want := &foo{"a"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_httpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

// Test handling of an error caused by the internal http client's Do()
// function. A redirect loop is pretty unlikely to occur within the GitHub
// API, but does allow us to exercise the right code path.
func TestDo_redirectLoop(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

// Test that an error caused by the internal http client's Do() function
// does not leak the client secret.
func TestDo_sanitizeURL(t *testing.T) {
	tp := &UnauthenticatedRateLimitedTransport{
		ClientID:     "id",
		ClientSecret: "secret",
	}
	unauthedClient := NewClient(tp.Client())
	unauthedClient.BaseURL = &url.URL{Scheme: "http", Host: "127.0.0.1:0", Path: "/"} // Use port 0 on purpose to trigger a dial TCP error, expect to get "dial tcp 127.0.0.1:0: connect: can't assign requested address".
	req, err := unauthedClient.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	_, err = unauthedClient.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("Expected error to be returned.")
	}
	if strings.Contains(err.Error(), "client_secret=secret") {
		t.Errorf("Do error contains secret, should be redacted:\n%q", err)
	}
}

func TestDo_rateLimit(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRateLimit, "60")
		w.Header().Set(headerRateRemaining, "59")
		w.Header().Set(headerRateReset, "1372700873")
	})

	req, _ := client.NewRequest("GET", "/", nil)
	resp, err := client.Do(context.Background(), req, nil)
	if err != nil {
		t.Errorf("Do returned unexpected error: %v", err)
	}
	if got, want := resp.Rate.Limit, 60; got != want {
		t.Errorf("Client rate limit = %v, want %v", got, want)
	}
	if got, want := resp.Rate.Remaining, 59; got != want {
		t.Errorf("Client rate remaining = %v, want %v", got, want)
	}
	reset := time.Date(2013, 7, 1, 17, 47, 53, 0, time.UTC)
	if resp.Rate.Reset.UTC() != reset {
		t.Errorf("Client rate reset = %v, want %v", resp.Rate.Reset, reset)
	}
}

// ensure rate limit is still parsed, even for error responses
func TestDo_rateLimit_errorResponse(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRateLimit, "60")
		w.Header().Set(headerRateRemaining, "59")
		w.Header().Set(headerRateReset, "1372700873")
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	resp, err := client.Do(context.Background(), req, nil)
	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if _, ok := err.(*RateLimitError); ok {
		t.Errorf("Did not expect a *RateLimitError error; got %#v.", err)
	}
	if got, want := resp.Rate.Limit, 60; got != want {
		t.Errorf("Client rate limit = %v, want %v", got, want)
	}
	if got, want := resp.Rate.Remaining, 59; got != want {
		t.Errorf("Client rate remaining = %v, want %v", got, want)
	}
	reset := time.Date(2013, 7, 1, 17, 47, 53, 0, time.UTC)
	if resp.Rate.Reset.UTC() != reset {
		t.Errorf("Client rate reset = %v, want %v", resp.Rate.Reset, reset)
	}
}

// Ensure *RateLimitError is returned when API rate limit is exceeded.
func TestDo_rateLimit_rateLimitError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRateLimit, "60")
		w.Header().Set(headerRateRemaining, "0")
		w.Header().Set(headerRateReset, "1372700873")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, `{
   "message": "API rate limit exceeded for xxx.xxx.xxx.xxx. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)",
   "documentation_url": "https://developer.github.com/v3/#rate-limiting"
}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("Expected a *RateLimitError error; got %#v.", err)
	}
	if got, want := rateLimitErr.Rate.Limit, 60; got != want {
		t.Errorf("rateLimitErr rate limit = %v, want %v", got, want)
	}
	if got, want := rateLimitErr.Rate.Remaining, 0; got != want {
		t.Errorf("rateLimitErr rate remaining = %v, want %v", got, want)
	}
	reset := time.Date(2013, 7, 1, 17, 47, 53, 0, time.UTC)
	if rateLimitErr.Rate.Reset.UTC() != reset {
		t.Errorf("rateLimitErr rate reset = %v, want %v", rateLimitErr.Rate.Reset.UTC(), reset)
	}
}

// Ensure a network call is not made when it's known that API rate limit is still exceeded.
func TestDo_rateLimit_noNetworkCall(t *testing.T) {
	setup()
	defer teardown()

	reset := time.Now().UTC().Add(time.Minute).Round(time.Second) // Rate reset is a minute from now, with 1 second precision.

	mux.HandleFunc("/first", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRateLimit, "60")
		w.Header().Set(headerRateRemaining, "0")
		w.Header().Set(headerRateReset, fmt.Sprint(reset.Unix()))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, `{
   "message": "API rate limit exceeded for xxx.xxx.xxx.xxx. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)",
   "documentation_url": "https://developer.github.com/v3/#rate-limiting"
}`)
	})

	madeNetworkCall := false
	mux.HandleFunc("/second", func(w http.ResponseWriter, r *http.Request) {
		madeNetworkCall = true
	})

	// First request is made, and it makes the client aware of rate reset time being in the future.
	req, _ := client.NewRequest("GET", "/first", nil)
	client.Do(context.Background(), req, nil)

	// Second request should not cause a network call to be made, since client can predict a rate limit error.
	req, _ = client.NewRequest("GET", "/second", nil)
	_, err := client.Do(context.Background(), req, nil)

	if madeNetworkCall {
		t.Fatal("Network call was made, even though rate limit is known to still be exceeded.")
	}

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	rateLimitErr, ok := err.(*RateLimitError)
	if !ok {
		t.Fatalf("Expected a *RateLimitError error; got %#v.", err)
	}
	if got, want := rateLimitErr.Rate.Limit, 60; got != want {
		t.Errorf("rateLimitErr rate limit = %v, want %v", got, want)
	}
	if got, want := rateLimitErr.Rate.Remaining, 0; got != want {
		t.Errorf("rateLimitErr rate remaining = %v, want %v", got, want)
	}
	if rateLimitErr.Rate.Reset.UTC() != reset {
		t.Errorf("rateLimitErr rate reset = %v, want %v", rateLimitErr.Rate.Reset.UTC(), reset)
	}
}

// Ensure *AbuseRateLimitError is returned when the response indicates that
// the client has triggered an abuse detection mechanism.
func TestDo_rateLimit_abuseRateLimitError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		// When the abuse rate limit error is of the "temporarily blocked from content creation" type,
		// there is no "Retry-After" header.
		fmt.Fprintln(w, `{
   "message": "You have triggered an abuse detection mechanism and have been temporarily blocked from content creation. Please retry your request again later.",
   "documentation_url": "https://developer.github.com/v3#abuse-rate-limits"
}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	abuseRateLimitErr, ok := err.(*AbuseRateLimitError)
	if !ok {
		t.Fatalf("Expected a *AbuseRateLimitError error; got %#v.", err)
	}
	if got, want := abuseRateLimitErr.RetryAfter, (*time.Duration)(nil); got != want {
		t.Errorf("abuseRateLimitErr RetryAfter = %v, want %v", got, want)
	}
}

// Ensure *AbuseRateLimitError.RetryAfter is parsed correctly.
func TestDo_rateLimit_abuseRateLimitError_retryAfter(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Retry-After", "123") // Retry after value of 123 seconds.
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, `{
   "message": "You have triggered an abuse detection mechanism ...",
   "documentation_url": "https://developer.github.com/v3#abuse-rate-limits"
}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	abuseRateLimitErr, ok := err.(*AbuseRateLimitError)
	if !ok {
		t.Fatalf("Expected a *AbuseRateLimitError error; got %#v.", err)
	}
	if abuseRateLimitErr.RetryAfter == nil {
		t.Fatalf("abuseRateLimitErr RetryAfter is nil, expected not-nil")
	}
	if got, want := *abuseRateLimitErr.RetryAfter, 123*time.Second; got != want {
		t.Errorf("abuseRateLimitErr RetryAfter = %v, want %v", got, want)
	}
}

func TestDo_noContent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	var body json.RawMessage

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(context.Background(), req, &body)
	if err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"/?a=b", "/?a=b"},
		{"/?a=b&client_secret=secret", "/?a=b&client_secret=REDACTED"},
		{"/?a=b&client_id=id&client_secret=secret", "/?a=b&client_id=id&client_secret=REDACTED"},
	}

	for _, tt := range tests {
		inURL, _ := url.Parse(tt.in)
		want, _ := url.Parse(tt.want)

		if got := sanitizeURL(inURL); !reflect.DeepEqual(got, want) {
			t.Errorf("sanitizeURL(%v) returned %v, want %v", tt.in, got, want)
		}
	}
}

func TestCheckResponse(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body: ioutil.NopCloser(strings.NewReader(`{"message":"m",
			"errors": [{"resource": "r", "field": "f", "code": "c"}],
			"block": {"reason": "dmca", "created_at": "2016-03-17T15:39:46Z"}}`)),
	}
	err := CheckResponse(res).(*ErrorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	want := &ErrorResponse{
		Response: res,
		Message:  "m",
		Errors:   []Error{{Resource: "r", Field: "f", Code: "c"}},
		Block: &struct {
			Reason    string     `json:"reason,omitempty"`
			CreatedAt *Timestamp `json:"created_at,omitempty"`
		}{
			Reason:    "dmca",
			CreatedAt: &Timestamp{time.Date(2016, 3, 17, 15, 39, 46, 0, time.UTC)},
		},
	}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("Error = %#v, want %#v", err, want)
	}
}

// ensure that we properly handle API errors that do not contain a response body
func TestCheckResponse_noBody(t *testing.T) {
	res := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
	err := CheckResponse(res).(*ErrorResponse)

	if err == nil {
		t.Errorf("Expected error response.")
	}

	want := &ErrorResponse{
		Response: res,
	}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("Error = %#v, want %#v", err, want)
	}
}

func TestParseBooleanResponse_true(t *testing.T) {
	result, err := parseBoolResponse(nil)
	if err != nil {
		t.Errorf("parseBoolResponse returned error: %+v", err)
	}

	if want := true; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestParseBooleanResponse_false(t *testing.T) {
	v := &ErrorResponse{Response: &http.Response{StatusCode: http.StatusNotFound}}
	result, err := parseBoolResponse(v)
	if err != nil {
		t.Errorf("parseBoolResponse returned error: %+v", err)
	}

	if want := false; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestParseBooleanResponse_error(t *testing.T) {
	v := &ErrorResponse{Response: &http.Response{StatusCode: http.StatusBadRequest}}
	result, err := parseBoolResponse(v)

	if err == nil {
		t.Errorf("Expected error to be returned.")
	}

	if want := false; result != want {
		t.Errorf("parseBoolResponse returned %+v, want: %+v", result, want)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		t.Errorf("Expected non-empty ErrorResponse.Error()")
	}
}

func TestError_Error(t *testing.T) {
	err := Error{}
	if err.Error() == "" {
		t.Errorf("Expected non-empty Error.Error()")
	}
}

func TestRateLimits(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/rate_limit", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		fmt.Fprint(w, `{"resources":{
			"core": {"limit":2,"remaining":1,"reset":1372700873},
			"search": {"limit":3,"remaining":2,"reset":1372700874}
		}}`)
	})

	rate, _, err := client.RateLimits(context.Background())
	if err != nil {
		t.Errorf("RateLimits returned error: %v", err)
	}

	want := &RateLimits{
		Core: &Rate{
			Limit:     2,
			Remaining: 1,
			Reset:     Timestamp{time.Date(2013, 7, 1, 17, 47, 53, 0, time.UTC).Local()},
		},
		Search: &Rate{
			Limit:     3,
			Remaining: 2,
			Reset:     Timestamp{time.Date(2013, 7, 1, 17, 47, 54, 0, time.UTC).Local()},
		},
	}
	if !reflect.DeepEqual(rate, want) {
		t.Errorf("RateLimits returned %+v, want %+v", rate, want)
	}

	if got, want := client.rateLimits[coreCategory], *want.Core; got != want {
		t.Errorf("client.rateLimits[coreCategory] is %+v, want %+v", got, want)
	}
	if got, want := client.rateLimits[searchCategory], *want.Search; got != want {
		t.Errorf("client.rateLimits[searchCategory] is %+v, want %+v", got, want)
	}
}

func TestUnauthenticatedRateLimitedTransport(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var v, want string
		q := r.URL.Query()
		if v, want = q.Get("client_id"), "id"; v != want {
			t.Errorf("OAuth Client ID = %v, want %v", v, want)
		}
		if v, want = q.Get("client_secret"), "secret"; v != want {
			t.Errorf("OAuth Client Secret = %v, want %v", v, want)
		}
	})

	tp := &UnauthenticatedRateLimitedTransport{
		ClientID:     "id",
		ClientSecret: "secret",
	}
	unauthedClient := NewClient(tp.Client())
	unauthedClient.BaseURL = client.BaseURL
	req, _ := unauthedClient.NewRequest("GET", "/", nil)
	unauthedClient.Do(context.Background(), req, nil)
}

func TestUnauthenticatedRateLimitedTransport_missingFields(t *testing.T) {
	// missing ClientID
	tp := &UnauthenticatedRateLimitedTransport{
		ClientSecret: "secret",
	}
	_, err := tp.RoundTrip(nil)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}

	// missing ClientSecret
	tp = &UnauthenticatedRateLimitedTransport{
		ClientID: "id",
	}
	_, err = tp.RoundTrip(nil)
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestUnauthenticatedRateLimitedTransport_transport(t *testing.T) {
	// default transport
	tp := &UnauthenticatedRateLimitedTransport{
		ClientID:     "id",
		ClientSecret: "secret",
	}
	if tp.transport() != http.DefaultTransport {
		t.Errorf("Expected http.DefaultTransport to be used.")
	}

	// custom transport
	tp = &UnauthenticatedRateLimitedTransport{
		ClientID:     "id",
		ClientSecret: "secret",
		Transport:    &http.Transport{},
	}
	if tp.transport() == http.DefaultTransport {
		t.Errorf("Expected custom transport to be used.")
	}
}

func TestBasicAuthTransport(t *testing.T) {
	setup()
	defer teardown()

	username, password, otp := "u", "p", "123456"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok {
			t.Errorf("request does not contain basic auth credentials")
		}
		if u != username {
			t.Errorf("request contained basic auth username %q, want %q", u, username)
		}
		if p != password {
			t.Errorf("request contained basic auth password %q, want %q", p, password)
		}
		if got, want := r.Header.Get(headerOTP), otp; got != want {
			t.Errorf("request contained OTP %q, want %q", got, want)
		}
	})

	tp := &BasicAuthTransport{
		Username: username,
		Password: password,
		OTP:      otp,
	}
	basicAuthClient := NewClient(tp.Client())
	basicAuthClient.BaseURL = client.BaseURL
	req, _ := basicAuthClient.NewRequest("GET", "/", nil)
	basicAuthClient.Do(context.Background(), req, nil)
}

func TestBasicAuthTransport_transport(t *testing.T) {
	// default transport
	tp := &BasicAuthTransport{}
	if tp.transport() != http.DefaultTransport {
		t.Errorf("Expected http.DefaultTransport to be used.")
	}

	// custom transport
	tp = &BasicAuthTransport{
		Transport: &http.Transport{},
	}
	if tp.transport() == http.DefaultTransport {
		t.Errorf("Expected custom transport to be used.")
	}
}
