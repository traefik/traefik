package middlewares

// Middleware tests based on https://github.com/unrolled/secure

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("bar"))
})

// newHeader constructs a new header instance with supplied options.
func newHeader(options ...HeaderOptions) *HeaderStruct {
	var opt HeaderOptions
	if len(options) == 0 {
		opt = HeaderOptions{}
	} else {
		opt = options[0]
	}

	return &HeaderStruct{opt: opt}
}

func TestNoConfig(t *testing.T) {
	header := newHeader()

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://example.com/foo", nil)

	header.ServeHTTP(res, req, myHandler)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "bar", res.Body.String(), "Body not the expected")
}

func TestModifyResponseHeaders(t *testing.T) {
	header := newHeader(HeaderOptions{
		CustomResponseHeaders: map[string]string{
			"X-Custom-Response-Header": "test_response",
		},
	})

	res := httptest.NewRecorder()
	res.HeaderMap.Add("X-Custom-Response-Header", "test_response")

	header.ModifyResponseHeaders(res.Result())

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "test_response", res.Header().Get("X-Custom-Response-Header"), "Did not get expected header")

	res = httptest.NewRecorder()
	res.HeaderMap.Add("X-Custom-Response-Header", "")

	header.ModifyResponseHeaders(res.Result())

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "", res.Header().Get("X-Custom-Response-Header"), "Did not get expected header")

	res = httptest.NewRecorder()
	res.HeaderMap.Add("X-Custom-Response-Header", "test_override")

	header.ModifyResponseHeaders(res.Result())

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "test_override", res.Header().Get("X-Custom-Response-Header"), "Did not get expected header")
}

func TestCustomRequestHeader(t *testing.T) {
	header := newHeader(HeaderOptions{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	header.ServeHTTP(res, req, nil)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"), "Did not get expected header")
}

func TestCustomRequestHeaderEmptyValue(t *testing.T) {
	header := newHeader(HeaderOptions{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "test_request",
		},
	})

	res := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "/foo", nil)

	header.ServeHTTP(res, req, nil)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "test_request", req.Header.Get("X-Custom-Request-Header"), "Did not get expected header")

	header = newHeader(HeaderOptions{
		CustomRequestHeaders: map[string]string{
			"X-Custom-Request-Header": "",
		},
	})

	header.ServeHTTP(res, req, nil)

	assert.Equal(t, http.StatusOK, res.Code, "Status not OK")
	assert.Equal(t, "", req.Header.Get("X-Custom-Request-Header"), "This header is not expected")
}
