package agent

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSnapshot(t *testing.T) {
	var snap io.Reader
	httpTest(t, func(srv *HTTPServer) {
		body := bytes.NewBuffer(nil)
		req, err := http.NewRequest("GET", "/v1/snapshot?token=root", body)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		_, err = srv.Snapshot(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		snap = resp.Body

		header := resp.Header().Get("X-Consul-Index")
		if header == "" {
			t.Fatalf("bad: %v", header)
		}
		header = resp.Header().Get("X-Consul-KnownLeader")
		if header != "true" {
			t.Fatalf("bad: %v", header)
		}
		header = resp.Header().Get("X-Consul-LastContact")
		if header != "0" {
			t.Fatalf("bad: %v", header)
		}
	})

	httpTest(t, func(srv *HTTPServer) {
		req, err := http.NewRequest("PUT", "/v1/snapshot?token=root", snap)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		_, err = srv.Snapshot(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	})
}

func TestSnapshot_Options(t *testing.T) {
	for _, method := range []string{"GET", "PUT"} {
		httpTest(t, func(srv *HTTPServer) {
			body := bytes.NewBuffer(nil)
			req, err := http.NewRequest(method, "/v1/snapshot?token=anonymous", body)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			resp := httptest.NewRecorder()
			_, err = srv.Snapshot(resp, req)
			if err == nil || !strings.Contains(err.Error(), "Permission denied") {
				t.Fatalf("err: %v", err)
			}
		})

		httpTest(t, func(srv *HTTPServer) {
			body := bytes.NewBuffer(nil)
			req, err := http.NewRequest(method, "/v1/snapshot?dc=nope", body)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			resp := httptest.NewRecorder()
			_, err = srv.Snapshot(resp, req)
			if err == nil || !strings.Contains(err.Error(), "No path to datacenter") {
				t.Fatalf("err: %v", err)
			}
		})

		httpTest(t, func(srv *HTTPServer) {
			body := bytes.NewBuffer(nil)
			req, err := http.NewRequest(method, "/v1/snapshot?token=root&stale", body)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			resp := httptest.NewRecorder()
			_, err = srv.Snapshot(resp, req)
			if method == "GET" {
				if err != nil {
					t.Fatalf("err: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), "stale not allowed") {
					t.Fatalf("err: %v", err)
				}
			}
		})
	}
}

func TestSnapshot_BadMethods(t *testing.T) {
	httpTest(t, func(srv *HTTPServer) {
		body := bytes.NewBuffer(nil)
		req, err := http.NewRequest("POST", "/v1/snapshot", body)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		_, err = srv.Snapshot(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if resp.Code != 405 {
			t.Fatalf("bad code: %d", resp.Code)
		}
	})

	httpTest(t, func(srv *HTTPServer) {
		body := bytes.NewBuffer(nil)
		req, err := http.NewRequest("DELETE", "/v1/snapshot", body)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		resp := httptest.NewRecorder()
		_, err = srv.Snapshot(resp, req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if resp.Code != 405 {
			t.Fatalf("bad code: %d", resp.Code)
		}
	})
}
