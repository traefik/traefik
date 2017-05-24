package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestErrorPage(t *testing.T) {
	testErrorPage := &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}}
	testHandler, err := NewErrorPagesHandler(*testErrorPage, "http://example.com")

	assert.Equal(t, nil, err, "Should be no error")
	assert.Equal(t, testHandler.BackendURL, "http://example.com/test", "Should be equal")

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://example.com/test", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	var n = negroni.New()
	n.Use(testHandler)
	n.UseHandler(handler)

	n.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "HTTP statusOK")
	assert.Contains(t, recorder.Body.String(), "traefik")

	handler500 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprintln(w, "oops")
	})
	recorder500 := httptest.NewRecorder()
	var n500 = negroni.New()
	n500.Use(testHandler)
	n500.UseHandler(handler500)

	n500.ServeHTTP(recorder500, req)

	assert.Equal(t, http.StatusInternalServerError, recorder500.Code, "HTTP status Internal Server Error")
	assert.Contains(t, recorder500.Body.String(), "This domain is established to be used for illustrative examples in documents")
	assert.NotContains(t, recorder500.Body.String(), "oops", "Should not return the oops page")

	handler502 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		fmt.Fprintln(w, "oops")
	})
	recorder502 := httptest.NewRecorder()
	var n502 = negroni.New()
	n502.Use(testHandler)
	n502.UseHandler(handler502)

	n502.ServeHTTP(recorder502, req)

	assert.Equal(t, http.StatusBadGateway, recorder502.Code, "HTTP status Internal Server Error")
	assert.Contains(t, recorder502.Body.String(), "oops")
	assert.NotContains(t, recorder502.Body.String(), "This domain is established to be used for illustrative examples in documents", "Should return the oops page")

}
