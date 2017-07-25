// +build integration

package secure

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codegangsta/negroni"
)

// go test -tags=integration

func TestIntegration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "bar")
	})

	secureMiddleware := New(Options{
		ContentTypeNosniff: true,
		FrameDeny:          true,
	})

	n := negroni.New()
	n.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	n.UseHandler(mux)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/foo", nil)
	n.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")
	expect(t, res.Header().Get("X-Frame-Options"), "DENY")
	expect(t, res.Header().Get("X-Content-Type-Options"), "nosniff")
}

func TestIntegrationWithError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "bar")
	})

	secureMiddleware := New(Options{
		ContentTypeNosniff: true,
		FrameDeny:          true,
		AllowedHosts:       []string{"www.example.com", "sub.example.com"},
	})

	n := negroni.New()
	n.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	n.UseHandler(mux)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	req.Host = "www3.example.com"
	n.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
}
