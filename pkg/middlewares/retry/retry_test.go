package retry

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestRetry(t *testing.T) {
	testCases := []struct {
		desc                  string
		config                dynamic.Retry
		wantRetryAttempts     int
		wantResponseStatus    int
		amountFaultyEndpoints int
	}{
		{
			desc:                  "no retry on success",
			config:                dynamic.Retry{Attempts: 5},
			wantRetryAttempts:     0,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 0,
		},
		{
			desc:                  "no retry on success with backoff",
			config:                dynamic.Retry{Attempts: 5, InitialInterval: ptypes.Duration(time.Microsecond * 50)},
			wantRetryAttempts:     0,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 0,
		},
		{
			desc:                  "no retry when max request attempts is one",
			config:                dynamic.Retry{Attempts: 1},
			wantRetryAttempts:     0,
			wantResponseStatus:    http.StatusBadGateway,
			amountFaultyEndpoints: 1,
		},
		{
			desc:                  "no retry when max request attempts is one with backoff",
			config:                dynamic.Retry{Attempts: 1, InitialInterval: ptypes.Duration(time.Microsecond * 50)},
			wantRetryAttempts:     0,
			wantResponseStatus:    http.StatusBadGateway,
			amountFaultyEndpoints: 1,
		},
		{
			desc:                  "one retry when one server is faulty",
			config:                dynamic.Retry{Attempts: 2},
			wantRetryAttempts:     1,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 1,
		},
		{
			desc:                  "one retry when one server is faulty with backoff",
			config:                dynamic.Retry{Attempts: 2, InitialInterval: ptypes.Duration(time.Microsecond * 50)},
			wantRetryAttempts:     1,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 1,
		},
		{
			desc:                  "two retries when two servers are faulty",
			config:                dynamic.Retry{Attempts: 3},
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 2,
		},
		{
			desc:                  "two retries when two servers are faulty with backoff",
			config:                dynamic.Retry{Attempts: 3, InitialInterval: ptypes.Duration(time.Microsecond * 50)},
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 2,
		},
		{
			desc:                  "max attempts exhausted delivers the 5xx response",
			config:                dynamic.Retry{Attempts: 3},
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusBadGateway,
			amountFaultyEndpoints: 3,
		},
		{
			desc:                  "max attempts exhausted delivers the 5xx response with backoff",
			config:                dynamic.Retry{Attempts: 3, InitialInterval: ptypes.Duration(time.Microsecond * 50)},
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusBadGateway,
			amountFaultyEndpoints: 3,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			retryAttempts := 0
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// This signals that a connection will be established with the backend
				// to enable the Retry middleware mechanism.
				shouldRetry := ContextShouldRetry(req.Context())
				if shouldRetry != nil {
					shouldRetry(true)
				}

				retryAttempts++

				if retryAttempts > test.amountFaultyEndpoints {
					// This signals that request headers have been sent to the backend.
					if shouldRetry != nil {
						shouldRetry(false)
					}

					rw.WriteHeader(http.StatusOK)
					return
				}

				rw.WriteHeader(http.StatusBadGateway)
			})

			retryListener := &countingRetryListener{}
			retry, err := New(t.Context(), next, test.config, retryListener, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://localhost:3000/ok", nil)

			retry.ServeHTTP(recorder, req)

			assert.Equal(t, test.wantResponseStatus, recorder.Code)
			assert.Equal(t, test.wantRetryAttempts, retryListener.timesCalled)
		})
	}
}

func TestRetryEmptyServerList(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusServiceUnavailable)
	})

	retryListener := &countingRetryListener{}
	retry, err := New(t.Context(), next, dynamic.Retry{Attempts: 3}, retryListener, "traefikTest")
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3000/ok", nil)

	retry.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, 0, retryListener.timesCalled)
}

func TestMultipleRetriesShouldNotLooseHeaders(t *testing.T) {
	attempt := 0
	expectedHeaderValue := "bar"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		shouldRetry := ContextShouldRetry(req.Context())
		if shouldRetry != nil {
			shouldRetry(true)
		}

		headerName := fmt.Sprintf("X-Foo-Test-%d", attempt)
		rw.Header().Add(headerName, expectedHeaderValue)
		if attempt < 2 {
			attempt++
			return
		}

		// Request has been successfully written to backend
		shouldRetry(false)

		// And we decide to answer to client.
		rw.WriteHeader(http.StatusNoContent)
	})

	retry, err := New(t.Context(), next, dynamic.Retry{Attempts: 3}, &countingRetryListener{}, "traefikTest")
	require.NoError(t, err)

	res := httptest.NewRecorder()
	retry.ServeHTTP(res, testhelpers.MustNewRequest(http.MethodGet, "http://test", http.NoBody))

	// The third header attempt is kept.
	headerValue := res.Header().Get("X-Foo-Test-2")
	assert.Equal(t, expectedHeaderValue, headerValue)

	// Validate that we don't have headers from previous attempts
	for i := range attempt {
		headerName := fmt.Sprintf("X-Foo-Test-%d", i)
		headerValue = res.Header().Get(headerName)
		if headerValue != "" {
			t.Errorf("Expected no value for header %s, got %s", headerName, headerValue)
		}
	}
}

func TestRetryShouldNotLooseHeadersOnWrite(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("X-Foo-Test", "bar")

		// Request has been successfully written to backend.
		shouldRetry := ContextShouldRetry(req.Context())
		if shouldRetry != nil {
			shouldRetry(false)
		}
		// And we decide to answer to client without calling WriteHeader.
		_, err := rw.Write([]byte("bar"))
		require.NoError(t, err)
	})

	retry, err := New(t.Context(), next, dynamic.Retry{Attempts: 3}, &countingRetryListener{}, "traefikTest")
	require.NoError(t, err)

	res := httptest.NewRecorder()
	retry.ServeHTTP(res, testhelpers.MustNewRequest(http.MethodGet, "http://test", http.NoBody))

	headerValue := res.Header().Get("X-Foo-Test")
	assert.Equal(t, "bar", headerValue)
}

func TestRetryWithFlush(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte("FULL "))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		rw.(http.Flusher).Flush()
		_, err = rw.Write([]byte("DATA"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})

	retry, err := New(t.Context(), next, dynamic.Retry{Attempts: 1}, &countingRetryListener{}, "traefikTest")
	require.NoError(t, err)

	responseRecorder := httptest.NewRecorder()

	retry.ServeHTTP(responseRecorder, &http.Request{})

	assert.Equal(t, "FULL DATA", responseRecorder.Body.String())
}

func TestRetryWebsocket(t *testing.T) {
	testCases := []struct {
		desc                   string
		maxRequestAttempts     int
		expectedRetryAttempts  int
		expectedResponseStatus int
		expectedError          bool
		amountFaultyEndpoints  int
	}{
		{
			desc:                   "Switching ok after 2 retries",
			maxRequestAttempts:     3,
			expectedRetryAttempts:  2,
			amountFaultyEndpoints:  2,
			expectedResponseStatus: http.StatusSwitchingProtocols,
		},
		{
			desc:                   "Switching failed",
			maxRequestAttempts:     2,
			expectedRetryAttempts:  1,
			amountFaultyEndpoints:  2,
			expectedResponseStatus: http.StatusBadGateway,
			expectedError:          true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			retryAttempts := 0
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// This signals that a connection will be established with the backend
				// to enable the Retry middleware mechanism.
				shouldRetry := ContextShouldRetry(req.Context())
				if shouldRetry != nil {
					shouldRetry(true)
				}

				retryAttempts++

				if retryAttempts > test.amountFaultyEndpoints {
					// This signals that request headers have been sent to the backend.
					if shouldRetry != nil {
						shouldRetry(false)
					}

					upgrader := websocket.Upgrader{}
					_, err := upgrader.Upgrade(rw, req, nil)
					if err != nil {
						http.Error(rw, err.Error(), http.StatusInternalServerError)
					}
					return
				}

				rw.WriteHeader(http.StatusBadGateway)
			})

			retryListener := &countingRetryListener{}
			retryH, err := New(t.Context(), next, dynamic.Retry{Attempts: test.maxRequestAttempts}, retryListener, "traefikTest")
			require.NoError(t, err)

			retryServer := httptest.NewServer(retryH)

			url := strings.Replace(retryServer.URL, "http", "ws", 1)
			_, response, err := websocket.DefaultDialer.Dial(url, nil)

			if !test.expectedError {
				require.NoError(t, err)
			}

			assert.Equal(t, test.expectedResponseStatus, response.StatusCode)
			assert.Equal(t, test.expectedRetryAttempts, retryListener.timesCalled)
		})
	}
}

// This test is an adapted version of net/http/httputil.Test1xxResponses test.
func Test1xxResponses(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Link", "</style.css>; rel=preload; as=style")
		h.Add("Link", "</script.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusEarlyHints)

		h.Add("Link", "</foo.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusProcessing)

		_, _ = w.Write([]byte("Hello"))
	})

	retryListener := &countingRetryListener{}
	retry, err := New(t.Context(), next, dynamic.Retry{Attempts: 1}, retryListener, "traefikTest")
	require.NoError(t, err)

	server := httptest.NewServer(retry)
	t.Cleanup(server.Close)
	frontendClient := server.Client()

	checkLinkHeaders := func(t *testing.T, expected, got []string) {
		t.Helper()

		if len(expected) != len(got) {
			t.Errorf("Expected %d link headers; got %d", len(expected), len(got))
		}

		for i := range expected {
			if i >= len(got) {
				t.Errorf("Expected %q link header; got nothing", expected[i])

				continue
			}

			if expected[i] != got[i] {
				t.Errorf("Expected %q link header; got %q", expected[i], got[i])
			}
		}
	}

	var respCounter uint8
	trace := &httptrace.ClientTrace{
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			switch code {
			case http.StatusEarlyHints:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script"}, header["Link"])
			case http.StatusProcessing:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, header["Link"])
			default:
				t.Error("Unexpected 1xx response")
			}

			respCounter++

			return nil
		},
	}
	req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(t.Context(), trace), http.MethodGet, server.URL, nil)

	res, err := frontendClient.Do(req)
	assert.NoError(t, err)

	defer res.Body.Close()

	if respCounter != 2 {
		t.Errorf("Expected 2 1xx responses; got %d", respCounter)
	}
	checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, res.Header["Link"])

	body, _ := io.ReadAll(res.Body)
	if string(body) != "Hello" {
		t.Errorf("Read body %q; want Hello", body)
	}

	assert.Equal(t, 0, retryListener.timesCalled)
}

// countingRetryListener is a Listener implementation to count the times the Retried fn is called.
type countingRetryListener struct {
	timesCalled int
}

func (l *countingRetryListener) Retried(req *http.Request, attempt int) {
	l.timesCalled++
}
