package acme

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPHeadUserAgent(t *testing.T) {
	var ua, method string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		method = r.Method
	}))
	defer ts.Close()

	_, err := httpHead(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if method != "HEAD" {
		t.Errorf("Expected method to be HEAD, got %s", method)
	}
	if !strings.Contains(ua, ourUserAgent) {
		t.Errorf("Expected User-Agent to contain '%s', got: '%s'", ourUserAgent, ua)
	}
}

func TestHTTPGetUserAgent(t *testing.T) {
	var ua, method string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		method = r.Method
	}))
	defer ts.Close()

	res, err := httpGet(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if method != "GET" {
		t.Errorf("Expected method to be GET, got %s", method)
	}
	if !strings.Contains(ua, ourUserAgent) {
		t.Errorf("Expected User-Agent to contain '%s', got: '%s'", ourUserAgent, ua)
	}
}

func TestHTTPPostUserAgent(t *testing.T) {
	var ua, method string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		method = r.Method
	}))
	defer ts.Close()

	res, err := httpPost(ts.URL, "text/plain", strings.NewReader("falalalala"))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if method != "POST" {
		t.Errorf("Expected method to be POST, got %s", method)
	}
	if !strings.Contains(ua, ourUserAgent) {
		t.Errorf("Expected User-Agent to contain '%s', got: '%s'", ourUserAgent, ua)
	}
}

func TestUserAgent(t *testing.T) {
	ua := userAgent()

	if !strings.Contains(ua, defaultGoUserAgent) {
		t.Errorf("Expected UA to contain %s, got '%s'", defaultGoUserAgent, ua)
	}
	if !strings.Contains(ua, ourUserAgent) {
		t.Errorf("Expected UA to contain %s, got '%s'", ourUserAgent, ua)
	}
	if strings.HasSuffix(ua, " ") {
		t.Errorf("UA should not have trailing spaces; got '%s'", ua)
	}

	// customize the UA by appending a value
	UserAgent = "MyApp/1.2.3"
	ua = userAgent()
	if !strings.Contains(ua, defaultGoUserAgent) {
		t.Errorf("Expected UA to contain %s, got '%s'", defaultGoUserAgent, ua)
	}
	if !strings.Contains(ua, ourUserAgent) {
		t.Errorf("Expected UA to contain %s, got '%s'", ourUserAgent, ua)
	}
	if !strings.Contains(ua, UserAgent) {
		t.Errorf("Expected custom UA to contain %s, got '%s'", UserAgent, ua)
	}
}
