package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

const (
	acceptEncodingHeader = "Accept-Encoding"
	varyHeader           = "Vary"
	gzip                 = "gzip"
)

func TestShouldCompressWhenNoContentEncodingHeader(t *testing.T) {
	handler := &Compress{}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzip)

	baseBody := generateBytes(gziphandler.DefaultMinSize)
	next := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(baseBody)
	}
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req, next)

	assert.Equal(t, gzip, rw.Header().Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, rw.Header().Get(varyHeader))

	if assert.ObjectsAreEqualValues(rw.Body.Bytes(), baseBody) {
		assert.Fail(t, "expected a compressed body", "got %v", rw.Body.Bytes())
	}
}

func TestShouldNotCompressWhenContentEncodingHeader(t *testing.T) {
	handler := &Compress{}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzip)
	req.Header.Add(contentEncodingHeader, gzip)

	baseBody := generateBytes(gziphandler.DefaultMinSize)

	next := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(baseBody)
	}

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req, next)

	assert.Equal(t, "", rw.Header().Get(contentEncodingHeader))
	assert.Equal(t, "", rw.Header().Get(varyHeader))

	assert.EqualValues(t, rw.Body.Bytes(), baseBody)
}

func generateBytes(len int) []byte {
	var value []byte
	for i := 0; i < len; i++ {
		value = append(value, 0x61)
	}
	return value
}
