package middlewares

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

const (
	acceptEncodingHeader  = "Accept-Encoding"
	contentEncodingHeader = "Content-Encoding"
	contentTypeHeader     = "Content-Type"
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

func TestShouldNotCompressWhenGRPC(t *testing.T) {
	handler := &Compress{}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzipValue)
	req.Header.Add(contentTypeHeader, "application/grpc")

	baseBody := generateBytes(gziphandler.DefaultMinSize)
	next := func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(baseBody)
	}

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req, next)

	assert.Empty(t, rw.Header().Get(acceptEncodingHeader))
	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.EqualValues(t, rw.Body.Bytes(), baseBody)
}

func TestIntegrationShouldNotCompress(t *testing.T) {
	fakeCompressedBody := generateBytes(100000)
	comp := &Compress{}

	testCases := []struct {
		name               string
		handler            func(rw http.ResponseWriter, r *http.Request)
		expectedStatusCode int
	}{
		{
			name: "when content already compressed",
			handler: func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Add(contentEncodingHeader, gzipValue)
				rw.Header().Add(varyHeader, acceptEncodingHeader)
				rw.Write(fakeCompressedBody)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "when content already compressed and status code Created",
			handler: func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Add(contentEncodingHeader, gzipValue)
				rw.Header().Add(varyHeader, acceptEncodingHeader)
				rw.WriteHeader(http.StatusCreated)
				rw.Write(fakeCompressedBody)
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, test := range testCases {

		t.Run(test.name, func(t *testing.T) {
			negro := negroni.New(comp)
			negro.UseHandlerFunc(test.handler)
			ts := httptest.NewServer(negro)
			defer ts.Close()

			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.Header.Add(acceptEncodingHeader, gzipValue)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)

			assert.Equal(t, gzipValue, resp.Header.Get(contentEncodingHeader))
			assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.EqualValues(t, fakeCompressedBody, body)
		})
	}
}

func TestShouldWriteHeaderWhenFlush(t *testing.T) {
	comp := &Compress{}
	negro := negroni.New(comp)
	negro.UseHandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add(contentEncodingHeader, gzipValue)
		rw.Header().Add(varyHeader, acceptEncodingHeader)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.(http.Flusher).Flush()
		rw.Write([]byte("short"))
	})
	ts := httptest.NewServer(negro)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Add(acceptEncodingHeader, gzipValue)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	assert.Equal(t, gzipValue, resp.Header.Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))
}

func TestIntegrationShouldCompress(t *testing.T) {
	fakeBody := generateBytes(100000)

	testCases := []struct {
		name               string
		handler            func(rw http.ResponseWriter, r *http.Request)
		expectedStatusCode int
	}{
		{
			name: "when AcceptEncoding header is present",
			handler: func(rw http.ResponseWriter, r *http.Request) {
				rw.Write(fakeBody)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "when AcceptEncoding header is present and status code Created",
			handler: func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusCreated)
				rw.Write(fakeBody)
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, test := range testCases {

		t.Run(test.name, func(t *testing.T) {
			comp := &Compress{}

			negro := negroni.New(comp)
			negro.UseHandlerFunc(test.handler)
			ts := httptest.NewServer(negro)
			defer ts.Close()

			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.Header.Add(acceptEncodingHeader, gzipValue)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)

			assert.Equal(t, gzipValue, resp.Header.Get(contentEncodingHeader))
			assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			if assert.ObjectsAreEqualValues(body, fakeBody) {
				assert.Fail(t, "expected a compressed body", "got %v", body)
			}
		})
	}
}

func generateBytes(len int) []byte {
	var value []byte
	for i := 0; i < len; i++ {
		value = append(value, 0x61+byte(i))
	}
	return value
}
