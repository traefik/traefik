// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	ok         = "ok\n"
	notAllowed = "Method not allowed\n"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(ok))
})

func newRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func TestMethodHandler(t *testing.T) {
	tests := []struct {
		req     *http.Request
		handler http.Handler
		code    int
		allow   string // Contents of the Allow header
		body    string
	}{
		// No handlers
		{newRequest("GET", "/foo"), MethodHandler{}, http.StatusMethodNotAllowed, "", notAllowed},
		{newRequest("OPTIONS", "/foo"), MethodHandler{}, http.StatusOK, "", ""},

		// A single handler
		{newRequest("GET", "/foo"), MethodHandler{"GET": okHandler}, http.StatusOK, "", ok},
		{newRequest("POST", "/foo"), MethodHandler{"GET": okHandler}, http.StatusMethodNotAllowed, "GET", notAllowed},

		// Multiple handlers
		{newRequest("GET", "/foo"), MethodHandler{"GET": okHandler, "POST": okHandler}, http.StatusOK, "", ok},
		{newRequest("POST", "/foo"), MethodHandler{"GET": okHandler, "POST": okHandler}, http.StatusOK, "", ok},
		{newRequest("DELETE", "/foo"), MethodHandler{"GET": okHandler, "POST": okHandler}, http.StatusMethodNotAllowed, "GET, POST", notAllowed},
		{newRequest("OPTIONS", "/foo"), MethodHandler{"GET": okHandler, "POST": okHandler}, http.StatusOK, "GET, POST", ""},

		// Override OPTIONS
		{newRequest("OPTIONS", "/foo"), MethodHandler{"OPTIONS": okHandler}, http.StatusOK, "", ok},
	}

	for i, test := range tests {
		rec := httptest.NewRecorder()
		test.handler.ServeHTTP(rec, test.req)
		if rec.Code != test.code {
			t.Fatalf("%d: wrong code, got %d want %d", i, rec.Code, test.code)
		}
		if allow := rec.HeaderMap.Get("Allow"); allow != test.allow {
			t.Fatalf("%d: wrong Allow, got %s want %s", i, allow, test.allow)
		}
		if body := rec.Body.String(); body != test.body {
			t.Fatalf("%d: wrong body, got %q want %q", i, body, test.body)
		}
	}
}

func TestWriteLog(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		panic(err)
	}
	ts := time.Date(1983, 05, 26, 3, 30, 45, 0, loc)

	// A typical request with an OK response
	req := newRequest("GET", "http://example.com")
	req.RemoteAddr = "192.168.100.5"

	buf := new(bytes.Buffer)
	writeLog(buf, req, *req.URL, ts, http.StatusOK, 100)
	log := buf.String()

	expected := "192.168.100.5 - - [26/May/1983:03:30:45 +0200] \"GET / HTTP/1.1\" 200 100\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}

	// Request with an unauthorized user
	req = newRequest("GET", "http://example.com")
	req.RemoteAddr = "192.168.100.5"
	req.URL.User = url.User("kamil")

	buf.Reset()
	writeLog(buf, req, *req.URL, ts, http.StatusUnauthorized, 500)
	log = buf.String()

	expected = "192.168.100.5 - kamil [26/May/1983:03:30:45 +0200] \"GET / HTTP/1.1\" 401 500\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}

	// Request with url encoded parameters
	req = newRequest("GET", "http://example.com/test?abc=hello%20world&a=b%3F")
	req.RemoteAddr = "192.168.100.5"

	buf.Reset()
	writeLog(buf, req, *req.URL, ts, http.StatusOK, 100)
	log = buf.String()

	expected = "192.168.100.5 - - [26/May/1983:03:30:45 +0200] \"GET /test?abc=hello%20world&a=b%3F HTTP/1.1\" 200 100\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}
}

func TestWriteCombinedLog(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		panic(err)
	}
	ts := time.Date(1983, 05, 26, 3, 30, 45, 0, loc)

	// A typical request with an OK response
	req := newRequest("GET", "http://example.com")
	req.RemoteAddr = "192.168.100.5"
	req.Header.Set("Referer", "http://example.com")
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) AppleWebKit/537.33 "+
			"(KHTML, like Gecko) Chrome/27.0.1430.0 Safari/537.33",
	)

	buf := new(bytes.Buffer)
	writeCombinedLog(buf, req, *req.URL, ts, http.StatusOK, 100)
	log := buf.String()

	expected := "192.168.100.5 - - [26/May/1983:03:30:45 +0200] \"GET / HTTP/1.1\" 200 100 \"http://example.com\" " +
		"\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) " +
		"AppleWebKit/537.33 (KHTML, like Gecko) Chrome/27.0.1430.0 Safari/537.33\"\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}

	// Request with an unauthorized user
	req.URL.User = url.User("kamil")

	buf.Reset()
	writeCombinedLog(buf, req, *req.URL, ts, http.StatusUnauthorized, 500)
	log = buf.String()

	expected = "192.168.100.5 - kamil [26/May/1983:03:30:45 +0200] \"GET / HTTP/1.1\" 401 500 \"http://example.com\" " +
		"\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) " +
		"AppleWebKit/537.33 (KHTML, like Gecko) Chrome/27.0.1430.0 Safari/537.33\"\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}

	// Test with remote ipv6 address
	req.RemoteAddr = "::1"

	buf.Reset()
	writeCombinedLog(buf, req, *req.URL, ts, http.StatusOK, 100)
	log = buf.String()

	expected = "::1 - kamil [26/May/1983:03:30:45 +0200] \"GET / HTTP/1.1\" 200 100 \"http://example.com\" " +
		"\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) " +
		"AppleWebKit/537.33 (KHTML, like Gecko) Chrome/27.0.1430.0 Safari/537.33\"\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}

	// Test remote ipv6 addr, with port
	req.RemoteAddr = net.JoinHostPort("::1", "65000")

	buf.Reset()
	writeCombinedLog(buf, req, *req.URL, ts, http.StatusOK, 100)
	log = buf.String()

	expected = "::1 - kamil [26/May/1983:03:30:45 +0200] \"GET / HTTP/1.1\" 200 100 \"http://example.com\" " +
		"\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_2) " +
		"AppleWebKit/537.33 (KHTML, like Gecko) Chrome/27.0.1430.0 Safari/537.33\"\n"
	if log != expected {
		t.Fatalf("wrong log, got %q want %q", log, expected)
	}
}

func TestLogPathRewrites(t *testing.T) {
	var buf bytes.Buffer

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = "/" // simulate http.StripPrefix and friends
		w.WriteHeader(200)
	})
	logger := LoggingHandler(&buf, handler)

	logger.ServeHTTP(httptest.NewRecorder(), newRequest("GET", "/subdir/asdf"))

	if !strings.Contains(buf.String(), "GET /subdir/asdf HTTP") {
		t.Fatalf("Got log %#v, wanted substring %#v", buf.String(), "GET /subdir/asdf HTTP")
	}
}

func BenchmarkWriteLog(b *testing.B) {
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		b.Fatalf(err.Error())
	}
	ts := time.Date(1983, 05, 26, 3, 30, 45, 0, loc)

	req := newRequest("GET", "http://example.com")
	req.RemoteAddr = "192.168.100.5"

	b.ResetTimer()

	buf := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		writeLog(buf, req, *req.URL, ts, http.StatusUnauthorized, 500)
	}
}

func TestContentTypeHandler(t *testing.T) {
	tests := []struct {
		Method            string
		AllowContentTypes []string
		ContentType       string
		Code              int
	}{
		{"POST", []string{"application/json"}, "application/json", http.StatusOK},
		{"POST", []string{"application/json", "application/xml"}, "application/json", http.StatusOK},
		{"POST", []string{"application/json"}, "application/json; charset=utf-8", http.StatusOK},
		{"POST", []string{"application/json"}, "application/json+xxx", http.StatusUnsupportedMediaType},
		{"POST", []string{"application/json"}, "text/plain", http.StatusUnsupportedMediaType},
		{"GET", []string{"application/json"}, "", http.StatusOK},
		{"GET", []string{}, "", http.StatusOK},
	}
	for _, test := range tests {
		r, err := http.NewRequest(test.Method, "/", nil)
		if err != nil {
			t.Error(err)
			continue
		}

		h := ContentTypeHandler(okHandler, test.AllowContentTypes...)
		r.Header.Set("Content-Type", test.ContentType)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != test.Code {
			t.Errorf("expected %d, got %d", test.Code, w.Code)
		}
	}
}

func TestHTTPMethodOverride(t *testing.T) {
	var tests = []struct {
		Method         string
		OverrideMethod string
		ExpectedMethod string
	}{
		{"POST", "PUT", "PUT"},
		{"POST", "PATCH", "PATCH"},
		{"POST", "DELETE", "DELETE"},
		{"PUT", "DELETE", "PUT"},
		{"GET", "GET", "GET"},
		{"HEAD", "HEAD", "HEAD"},
		{"GET", "PUT", "GET"},
		{"HEAD", "DELETE", "HEAD"},
	}

	for _, test := range tests {
		h := HTTPMethodOverrideHandler(okHandler)
		reqs := make([]*http.Request, 0, 2)

		rHeader, err := http.NewRequest(test.Method, "/", nil)
		if err != nil {
			t.Error(err)
		}
		rHeader.Header.Set(HTTPMethodOverrideHeader, test.OverrideMethod)
		reqs = append(reqs, rHeader)

		f := url.Values{HTTPMethodOverrideFormKey: []string{test.OverrideMethod}}
		rForm, err := http.NewRequest(test.Method, "/", strings.NewReader(f.Encode()))
		if err != nil {
			t.Error(err)
		}
		rForm.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		reqs = append(reqs, rForm)

		for _, r := range reqs {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if r.Method != test.ExpectedMethod {
				t.Errorf("Expected %s, got %s", test.ExpectedMethod, r.Method)
			}
		}
	}
}
