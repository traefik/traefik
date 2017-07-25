package dnspod

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the dnspod client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// This method of testing http client APIs is borrowed from
// Will Norris's work in go-github @ https://github.com/google/go-github
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	params := CommonParams{LoginToken: "dnspod login token"}

	client = NewClient(params)
	client.BaseURL = server.URL + "/"
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if want != r.Method {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func testRequestJSON(t *testing.T, r *http.Request, values map[string]interface{}) {
	var dat map[string]interface{}

	body, _ := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(body, &dat); err != nil {
		t.Errorf("Could not decode json body: %v", err)
	}

	if !reflect.DeepEqual(values, dat) {
		t.Errorf("Request parameters = %v, want %v", dat, values)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if !reflect.DeepEqual(want, r.Form) {
		t.Errorf("Request parameters = %v, want %v", r.Form, want)
	}
}

func testString(t *testing.T, test, value, want string) {
	if value != want {
		t.Errorf("%s returned %+v, want %+v", test, value, want)
	}
}

func TestNewClient(t *testing.T) {
	params := CommonParams{LoginToken: "dnspod login token"}

	c := NewClient(params)

	if c.BaseURL != baseURL {
		t.Errorf("NewClient BaseURL = %v, want %v", c.BaseURL, baseURL)
	}
}

func TestNewRequest(t *testing.T) {
	params := CommonParams{LoginToken: "dnspod login token"}
	c := NewClient(params)
	c.BaseURL = "https://go.example.com/"

	inURL, outURL := "foo", "https://go.example.com/foo"
	req, _ := c.NewRequest("GET", inURL, nil)

	// test that relative URL was expanded with the proper BaseURL
	if req.URL.String() != outURL {
		t.Errorf("makeRequest(%v) URL = %v, want %v", inURL, req.URL, outURL)
	}

	// test that default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("makeRequest() User-Agent = %v, want %v", userAgent, c.UserAgent)
	}
}
