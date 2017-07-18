package middlewares

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NYTimes/gziphandler"
	"github.com/codegangsta/negroni"
	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

const (
	acceptEncodingHeader  = "Accept-Encoding"
	contentEncodingHeader = "Content-Encoding"
	varyHeader            = "Vary"
	gzipValue             = "gzip"
)

func TestShouldCompressWhenNoContentEncodingHeader(t *testing.T) {
	handler := &Compress{}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzipValue)

	baseBody := generateBytes(gziphandler.DefaultMinSize)
	next := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(baseBody)
	}

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req, next)

	assert.Equal(t, gzipValue, rw.Header().Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, rw.Header().Get(varyHeader))

	if assert.ObjectsAreEqualValues(rw.Body.Bytes(), baseBody) {
		assert.Fail(t, "expected a compressed body", "got %v", rw.Body.Bytes())
	}
}

func TestShouldNotCompressWhenContentEncodingHeader(t *testing.T) {
	handler := &Compress{}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzipValue)

	fakeCompressedBody := generateBytes(gziphandler.DefaultMinSize)
	next := func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add(contentEncodingHeader, gzipValue)
		rw.Header().Add(varyHeader, acceptEncodingHeader)
		rw.Write(fakeCompressedBody)
	}

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req, next)

	assert.Equal(t, gzipValue, rw.Header().Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, rw.Header().Get(varyHeader))

	assert.EqualValues(t, rw.Body.Bytes(), fakeCompressedBody)
}

func TestShouldNotCompressWhenNoAcceptEncodingHeader(t *testing.T) {
	handler := &Compress{}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)

	fakeBody := generateBytes(gziphandler.DefaultMinSize)
	next := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(fakeBody)
	}

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req, next)

	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.EqualValues(t, rw.Body.Bytes(), fakeBody)
}

func TestIntegrationShouldNotCompressWhenContentAlreadyCompressed(t *testing.T) {
	fakeCompressedBody := generateBytes(100000)

	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add(contentEncodingHeader, gzipValue)
		rw.Header().Add(varyHeader, acceptEncodingHeader)
		rw.Write(fakeCompressedBody)
	}

	comp := &Compress{}

	negro := negroni.New(comp)
	negro.UseHandlerFunc(handler)
	ts := httptest.NewServer(negro)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Add(acceptEncodingHeader, gzipValue)

	resp, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")

	assert.Equal(t, gzipValue, resp.Header.Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))

	body, err := ioutil.ReadAll(resp.Body)
	assert.EqualValues(t, fakeCompressedBody, body)
}

func TestIntegrationShouldCompressWhenAcceptEncodingHeaderIsPresent(t *testing.T) {
	fakeBody := generateBytes(100000)

	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(fakeBody)
	}

	comp := &Compress{}

	negro := negroni.New(comp)
	negro.UseHandlerFunc(handler)
	ts := httptest.NewServer(negro)
	defer ts.Close()

	client := &http.Client{}
	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Add(acceptEncodingHeader, gzipValue)

	resp, err := client.Do(req)
	assert.NoError(t, err, "there should be no error")

	assert.Equal(t, gzipValue, resp.Header.Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))

	body, err := ioutil.ReadAll(resp.Body)
	if assert.ObjectsAreEqualValues(body, fakeBody) {
		assert.Fail(t, "expected a compressed body", "got %v", body)
	}
}

func generateBytes(len int) []byte {
	var value []byte
	for i := 0; i < len; i++ {
		value = append(value, 0x61+byte(i))
	}
	return value
}
