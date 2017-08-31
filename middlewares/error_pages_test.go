package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

func TestErrorPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Test Server")
	}))
	defer ts.Close()

	testErrorPage := &types.ErrorPage{Backend: "error", Query: "/test", Status: []string{"500-501", "503-599"}}

	testHandler, err := NewErrorPagesHandler(*testErrorPage, ts.URL)
	require.NoError(t, err)

	assert.Equal(t, testHandler.BackendURL, ts.URL+"/test", "Should be equal")

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "traefik")
	})
	n := negroni.New()
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
	n500 := negroni.New()
	n500.Use(testHandler)
	n500.UseHandler(handler500)

	n500.ServeHTTP(recorder500, req)

	assert.Equal(t, http.StatusInternalServerError, recorder500.Code, "HTTP status Internal Server Error")
	assert.Contains(t, recorder500.Body.String(), "Test Server")
	assert.NotContains(t, recorder500.Body.String(), "oops", "Should not return the oops page")

	handler502 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		fmt.Fprintln(w, "oops")
	})
	recorder502 := httptest.NewRecorder()
	n502 := negroni.New()
	n502.Use(testHandler)
	n502.UseHandler(handler502)

	n502.ServeHTTP(recorder502, req)

	assert.Equal(t, http.StatusBadGateway, recorder502.Code, "HTTP status Bad Gateway")
	assert.Contains(t, recorder502.Body.String(), "oops")
	assert.NotContains(t, recorder502.Body.String(), "Test Server", "Should return the oops page since we have not configured the 502 code")
}

func TestErrorPageQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/"+strconv.Itoa(503) {
			fmt.Fprintln(w, "503 Test Server")
		} else {
			fmt.Fprintln(w, "Failed")
		}

	}))
	defer ts.Close()

	testErrorPage := &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503-503"}}

	testHandler, err := NewErrorPagesHandler(*testErrorPage, ts.URL)
	require.NoError(t, err)

	assert.Equal(t, testHandler.BackendURL, ts.URL+"/{status}", "Should be equal")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		fmt.Fprintln(w, "oops")
	})

	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	require.NoError(t, err)

	n := negroni.New()
	n.Use(testHandler)
	n.UseHandler(handler)

	n.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status Service Unavailable")
	assert.Contains(t, recorder.Body.String(), "503 Test Server")
	assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
}

func TestErrorPageSingleCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == "/"+strconv.Itoa(503) {
			fmt.Fprintln(w, "503 Test Server")
		} else {
			fmt.Fprintln(w, "Failed")
		}

	}))
	defer ts.Close()

	testErrorPage := &types.ErrorPage{Backend: "error", Query: "/{status}", Status: []string{"503"}}

	testHandler, err := NewErrorPagesHandler(*testErrorPage, ts.URL)
	require.NoError(t, err)

	assert.Equal(t, testHandler.BackendURL, ts.URL+"/{status}", "Should be equal")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		fmt.Fprintln(w, "oops")
	})

	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	require.NoError(t, err)

	n := negroni.New()
	n.Use(testHandler)
	n.UseHandler(handler)

	n.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code, "HTTP status Service Unavailable")
	assert.Contains(t, recorder.Body.String(), "503 Test Server")
	assert.NotContains(t, recorder.Body.String(), "oops", "Should not return the oops page")
}
