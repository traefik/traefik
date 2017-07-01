package middlewares

//Middleware tests based on https://github.com/unrolled/secure

import (
	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bar"))
})

func TestNoConfig(t *testing.T) {
	s := NewHeader()

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "bar", res.Body.String(), "Body not the expected")
}

func TestCustomResponseHeader(t *testing.T) {
	s := NewHeader(HeaderOptions{
		CustomResponseHeaders: map[string]string{
			"X-Custom-Response-Header": "test_response",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "test_response", res.Header().Get("X-Custom-Response-Header"), "Did not get expected header")
}

func TestCustomRequestHeader(t *testing.T) {
	s := NewHeader(HeaderOptions{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	s.Handler(myHandler).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"), "Did not get expected header")
}
