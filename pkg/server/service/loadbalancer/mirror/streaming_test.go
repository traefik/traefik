package mirror

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// mirrorResult is what the mirror received.
type mirrorResult struct {
	req  *http.Request
	body []byte
	err  error
}

func readingMirror(results chan<- mirrorResult) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		results <- mirrorResult{req: req, body: b, err: err}
	})
}

func TestMirroringFraction(t *testing.T) {
	testCases := []struct {
		desc            string
		numerator       int64
		denominator     int64
		requests        int
		expectedMirrors int32
	}{
		{
			desc:            "mirror a third of the requests",
			numerator:       1,
			denominator:     3,
			requests:        9,
			expectedMirrors: 3,
		},
		{
			desc:            "mirror less than one percent of the requests",
			numerator:       5,
			denominator:     1000,
			requests:        2000,
			expectedMirrors: 10,
		},
		{
			desc:            "mirror no request",
			numerator:       0,
			denominator:     100,
			requests:        10,
			expectedMirrors: 0,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var mirrored atomic.Int32
			pool := safe.NewPool(t.Context())
			mirror := New(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pool, true, -1, true, nil)
			err := mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				mirrored.Add(1)
			}), test.numerator, test.denominator)
			require.NoError(t, err)

			for range test.requests {
				mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://foo.com/bar", http.NoBody))
			}

			pool.Stop()

			assert.Equal(t, test.expectedMirrors, mirrored.Load())
		})
	}
}

func TestStreaming_handlersDoNotAlterMirroredRequest(t *testing.T) {
	results := make(chan mirrorResult, 1)
	var servedHeader string
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Header.Set("X-Foo", "modified")
		servedHeader = req.Header.Get("X-Foo")
	})

	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(readingMirror(results), 1, 1))

	req := httptest.NewRequest(http.MethodGet, "http://foo.com/bar", http.NoBody)
	req.Header.Set("X-Foo", "original")
	mirror.ServeHTTP(httptest.NewRecorder(), req)

	result := <-results
	pool.Stop()

	assert.Equal(t, "modified", servedHeader)
	assert.Equal(t, "original", result.req.Header.Get("X-Foo"))
}

func TestStreaming_body(t *testing.T) {
	body := make([]byte, 64*1024)
	_, err := rand.Read(body)
	require.NoError(t, err)

	var servedBody []byte
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		servedBody, _ = io.ReadAll(req.Body)
	})

	results := make(chan mirrorResult, 2)
	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(readingMirror(results), 1, 1))
	require.NoError(t, mirror.AddMirror(readingMirror(results), 1, 1))

	req := httptest.NewRequest(http.MethodPost, "http://foo.com/bar", bytes.NewReader(body))
	req.Header.Set("Expect", "100-continue")
	mirror.ServeHTTP(httptest.NewRecorder(), req)

	results1, results2 := <-results, <-results
	pool.Stop()

	assert.Equal(t, body, servedBody)
	for _, result := range []mirrorResult{results1, results2} {
		require.NoError(t, result.err)
		assert.Equal(t, body, result.body)
		assert.Equal(t, int64(len(body)), result.req.ContentLength)
		assert.Empty(t, result.req.Header.Get("Expect"))
	}
}

// TestStreaming_bodyNotBuffered checks that the main handler reads the body while the client is still sending it.
func TestStreaming_bodyNotBuffered(t *testing.T) {
	firstChunk := make(chan string)
	var servedBody string
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		chunk := make([]byte, 5)
		_, err := io.ReadFull(req.Body, chunk)
		require.NoError(t, err)
		firstChunk <- string(chunk)

		rest, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		servedBody = string(chunk) + string(rest)
	})

	results := make(chan mirrorResult, 1)
	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(readingMirror(results), 1, 1))

	bodyReader, bodyWriter := io.Pipe()
	req := httptest.NewRequest(http.MethodPost, "http://foo.com/bar", bodyReader)
	req.ContentLength = -1

	done := make(chan struct{})
	go func() {
		defer close(done)
		mirror.ServeHTTP(httptest.NewRecorder(), req)
	}()

	_, err := bodyWriter.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, "hello", <-firstChunk)

	_, err = bodyWriter.Write([]byte(" world"))
	require.NoError(t, err)
	require.NoError(t, bodyWriter.Close())

	<-done
	result := <-results
	pool.Stop()

	assert.Equal(t, "hello world", servedBody)
	require.NoError(t, result.err)
	assert.Equal(t, "hello world", string(result.body))
}

func TestStreaming_slowMirror(t *testing.T) {
	unblock := make(chan struct{})
	results := make(chan mirrorResult, 1)
	slowMirror := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-unblock
		readingMirror(results).ServeHTTP(rw, req)
	})

	var servedBody string
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		servedBody = string(b)
	})

	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(slowMirror, 1, 1))

	body := strings.Repeat("a", 2*streamingBufferSize)
	mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://foo.com/bar", strings.NewReader(body)))

	// The main request is served while the mirror has not read anything yet.
	assert.Equal(t, body, servedBody)

	close(unblock)
	result := <-results
	pool.Stop()

	require.ErrorIs(t, result.err, errMirrorTooSlow)
	assert.Empty(t, result.body)
}

// TestStreaming_readLargerThanBuffer checks that a mirror keeping up is not aborted
// when the main handler reads more than the buffer size at once.
func TestStreaming_readLargerThanBuffer(t *testing.T) {
	body := strings.Repeat("a", 2*streamingBufferSize)

	var servedBody string
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		chunk := make([]byte, len(body))
		n, err := io.ReadFull(req.Body, chunk)
		require.NoError(t, err)
		servedBody = string(chunk[:n])

		_, err = req.Body.Read(make([]byte, 1))
		require.ErrorIs(t, err, io.EOF)
	})

	results := make(chan mirrorResult, 1)
	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(readingMirror(results), 1, 1))

	mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://foo.com/bar", strings.NewReader(body)))

	result := <-results
	pool.Stop()

	assert.Equal(t, body, servedBody)
	require.NoError(t, result.err)
	assert.Equal(t, body, string(result.body))
}

func TestStreaming_bodyNotFullyRead(t *testing.T) {
	served := make(chan struct{})
	results := make(chan mirrorResult, 1)
	waitingMirror := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-served
		readingMirror(results).ServeHTTP(rw, req)
	})

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadFull(req.Body, make([]byte, 2))
		require.NoError(t, err)
	})

	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(waitingMirror, 1, 1))

	mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://foo.com/bar", strings.NewReader("12345")))
	close(served)

	result := <-results
	pool.Stop()

	// The mirrored request is aborted rather than sent with a truncated body.
	require.ErrorIs(t, result.err, errBodyNotFullyRead)
	assert.Equal(t, "12", string(result.body))
}

// TestStreaming_unknownLengthEmptyBody checks that a bodyless request with an unknown length,
// as sent over HTTP/3 when the client does not set a Content-Length header, is mirrored.
func TestStreaming_unknownLengthEmptyBody(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = io.ReadAll(req.Body)
	})

	results := make(chan mirrorResult, 1)
	pool := safe.NewPool(t.Context())
	mirror := New(next, pool, true, -1, true, nil)
	require.NoError(t, mirror.AddMirror(readingMirror(results), 1, 1))

	req := httptest.NewRequest(http.MethodGet, "http://foo.com/bar", io.NopCloser(strings.NewReader("")))
	req.ContentLength = -1
	mirror.ServeHTTP(httptest.NewRecorder(), req)

	result := <-results
	pool.Stop()

	require.NoError(t, result.err)
	assert.Empty(t, result.body)
}

// TestStreaming_proxiedMirror checks, with a mirror proxying to a real backend,
// that the backend receives the whole body, or a failed request when the mirrored request is aborted.
func TestStreaming_proxiedMirror(t *testing.T) {
	testCases := []struct {
		desc         string
		contentLen   bool
		readBytes    int
		expectedBody string
		expectedErr  bool
	}{
		{
			desc:         "body fully read",
			contentLen:   true,
			readBytes:    -1,
			expectedBody: "hello world",
		},
		{
			desc:        "body not fully read with a known length",
			contentLen:  true,
			readBytes:   5,
			expectedErr: true,
		},
		{
			desc:        "body not fully read with an unknown length",
			readBytes:   5,
			expectedErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			type backendResult struct {
				body string
				err  error
			}
			results := make(chan backendResult, 1)
			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				b, err := io.ReadAll(req.Body)
				results <- backendResult{body: string(b), err: err}
			}))
			t.Cleanup(backend.Close)

			backendURL, err := url.Parse(backend.URL)
			require.NoError(t, err)
			proxy := httputil.NewSingleHostReverseProxy(backendURL)

			var served sync.WaitGroup
			served.Add(1)
			proxyMirror := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Lets the main handler stop reading before the mirror starts, to make the abort deterministic.
				served.Wait()
				proxy.ServeHTTP(rw, req)
			})

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if test.readBytes < 0 {
					_, _ = io.ReadAll(req.Body)
					return
				}
				_, _ = io.ReadFull(req.Body, make([]byte, test.readBytes))
			})

			pool := safe.NewPool(t.Context())
			mirror := New(next, pool, true, -1, true, nil)
			require.NoError(t, mirror.AddMirror(proxyMirror, 1, 1))

			req := httptest.NewRequest(http.MethodPost, "http://foo.com/bar", io.NopCloser(strings.NewReader("hello world")))
			req.ContentLength = -1
			if test.contentLen {
				req.ContentLength = int64(len("hello world"))
			}
			mirror.ServeHTTP(httptest.NewRecorder(), req)
			served.Done()
			pool.Stop()

			if !test.expectedErr {
				result := <-results
				require.NoError(t, result.err)
				assert.Equal(t, test.expectedBody, result.body)
				return
			}

			// The backend either never receives the request, or fails to read a truncated body.
			select {
			case result := <-results:
				assert.Error(t, result.err)
			case <-time.After(time.Second):
			}
		})
	}
}
