package brotlihandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gzipLength   = 306 // Lenght of 1400-byte response compressed with gzip
	brotliLength = 269 // Lenght of 1400-byte response compressed with Brotli
)

func TestShouldCompressWhenNoContentEncoding(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", gzipEncoding)

	baseBody := generateBytes(1400)

	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(baseBody)
		assert.NoError(t, err)
	})
	handler := CompressHandler(next)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, gzipEncoding, rw.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", rw.Header().Get("Vary"))
	assert.Equal(t, gzipLength, len(rw.Body.Bytes()))

	if assert.ObjectsAreEqualValues(rw.Body.Bytes(), baseBody) {
		assert.Fail(t, "expected a compressed body", "got %v", rw.Body.Bytes())
	}
}

func TestShouldCompressBrotliWhenNoContentEncoding(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", brEncoding)

	baseBody := generateBytes(1400)

	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(baseBody)
		assert.NoError(t, err)
	})
	handler := CompressHandler(next)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, brEncoding, rw.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", rw.Header().Get("Vary"))
	assert.Equal(t, brotliLength, len(rw.Body.Bytes()))

	if assert.ObjectsAreEqualValues(rw.Body.Bytes(), baseBody) {
		assert.Fail(t, "expected a compressed body", "got %v", rw.Body.Bytes())
	}
}

func TestShouldCompressBrotliWhenAcceptBrotliAndGzip(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", "gzip,br")

	baseBody := generateBytes(1400)

	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(baseBody)
		assert.NoError(t, err)
	})
	handler := CompressHandler(next)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, brEncoding, rw.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", rw.Header().Get("Vary"))
	assert.Equal(t, brotliLength, len(rw.Body.Bytes()))

	if assert.ObjectsAreEqualValues(rw.Body.Bytes(), baseBody) {
		assert.Fail(t, "expected a compressed body", "got %v", rw.Body.Bytes())
	}
}

func TestShouldNotCompressCompressed(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", deflateEncoding)

	fakeCompressedBody := generateBytes(1400)
	handler := CompressHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Encoding", deflateEncoding)
		rw.Header().Add("Vary", "Accept-Encoding")
		_, err := rw.Write(fakeCompressedBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}))

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, deflateEncoding, rw.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", rw.Header().Get("Vary"))

	assert.EqualValues(t, rw.Body.Bytes(), fakeCompressedBody)
}

func TestShouldNotCompressWhenNotAccepted(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", "deflate")

	fakeBody := generateBytes(1400)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler := CompressHandler(next)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get("Content-Encoding"))
	assert.EqualValues(t, rw.Body.Bytes(), fakeBody)
}

func TestShouldNotCompressWhenShort(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	assert.Nil(t, err, "Request error")

	fakeBody := generateBytes(1000)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler := CompressHandler(next)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get("Content-Encoding"))
	assert.EqualValues(t, rw.Body.Bytes(), fakeBody)
}

func TestShouldWriteHeader(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Encoding", gzipEncoding)
		rw.Header().Add("Vary", "Accept-Encoding")
		rw.WriteHeader(http.StatusUnauthorized)
		_, err := rw.Write([]byte("short"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler := CompressHandler(next)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", gzipEncoding)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	assert.Equal(t, gzipEncoding, resp.Header.Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", resp.Header.Get("Vary"))
}

func TestShouldWriteHeaderWhenFlush(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Encoding", gzipEncoding)
		rw.Header().Add("Vary", "Accept-Encoding")
		rw.WriteHeader(http.StatusUnauthorized)
		rw.(http.Flusher).Flush()
		_, err := rw.Write([]byte("short"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler := CompressHandler(next)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	assert.Nil(t, err, "Request error")
	req.Header.Add("Accept-Encoding", gzipEncoding)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	assert.Equal(t, gzipEncoding, resp.Header.Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", resp.Header.Get("Vary"))
}

func generateBytes(len int) []byte {
	var value []byte
	for i := 0; i < len; i++ {
		value = append(value, 0x61+byte(i))
	}
	return value
}
