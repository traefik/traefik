package retry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/middlewares/emptybackendhandler"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
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
			config:                dynamic.Retry{Attempts: 1},
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
			desc:                  "one retry when one server is faulty",
			config:                dynamic.Retry{Attempts: 2},
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
			desc:                  "max attempts exhausted delivers the 5xx response",
			config:                dynamic.Retry{Attempts: 3},
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusBadGateway,
			amountFaultyEndpoints: 3,
		},
	}

	backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte("OK"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}))

	forwarder, err := forward.New()
	require.NoError(t, err)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loadBalancer, err := roundrobin.New(forwarder)
			require.NoError(t, err)

			// out of range port
			basePort := 1133444
			for i := 0; i < test.amountFaultyEndpoints; i++ {
				// 192.0.2.0 is a non-routable IP for testing purposes.
				// See: https://stackoverflow.com/questions/528538/non-routable-ip-address/18436928#18436928
				// We only use the port specification here because the URL is used as identifier
				// in the load balancer and using the exact same URL would not add a new server.
				err = loadBalancer.UpsertServer(testhelpers.MustParseURL("http://192.0.2.0:" + strconv.Itoa(basePort+i)))
				require.NoError(t, err)
			}

			// add the functioning server to the end of the load balancer list
			err = loadBalancer.UpsertServer(testhelpers.MustParseURL(backendServer.URL))
			require.NoError(t, err)

			retryListener := &countingRetryListener{}
			retry, err := New(context.Background(), loadBalancer, test.config, retryListener, "traefikTest")
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
	forwarder, err := forward.New()
	require.NoError(t, err)

	loadBalancer, err := roundrobin.New(forwarder)
	require.NoError(t, err)

	// The EmptyBackend middleware ensures that there is a 503
	// response status set when there is no backend server in the pool.
	next := emptybackendhandler.New(loadBalancer)

	retryListener := &countingRetryListener{}
	retry, err := New(context.Background(), next, dynamic.Retry{Attempts: 3}, retryListener, "traefikTest")
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3000/ok", nil)

	retry.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, 0, retryListener.timesCalled)
}

func TestRetryListeners(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	retryListeners := Listeners{&countingRetryListener{}, &countingRetryListener{}}

	retryListeners.Retried(req, 1)
	retryListeners.Retried(req, 1)

	for _, retryListener := range retryListeners {
		listener := retryListener.(*countingRetryListener)
		if listener.timesCalled != 2 {
			t.Errorf("retry listener was called %d time(s), want %d time(s)", listener.timesCalled, 2)
		}
	}
}

func TestMultipleRetriesShouldNotLooseHeaders(t *testing.T) {
	attempt := 0
	expectedHeaderName := "X-Foo-Test-2"
	expectedHeaderValue := "bar"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		headerName := fmt.Sprintf("X-Foo-Test-%d", attempt)
		rw.Header().Add(headerName, expectedHeaderValue)
		if attempt < 2 {
			attempt++
			return
		}

		// Request has been successfully written to backend
		trace := httptrace.ContextClientTrace(req.Context())
		trace.WroteHeaders()

		// And we decide to answer to client
		rw.WriteHeader(http.StatusNoContent)
	})

	retry, err := New(context.Background(), next, dynamic.Retry{Attempts: 3}, &countingRetryListener{}, "traefikTest")
	require.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	retry.ServeHTTP(responseRecorder, testhelpers.MustNewRequest(http.MethodGet, "http://test", http.NoBody))

	headerValue := responseRecorder.Header().Get(expectedHeaderName)

	// Validate if we have the correct header
	if headerValue != expectedHeaderValue {
		t.Errorf("Expected to have %s for header %s, got %s", expectedHeaderValue, expectedHeaderName, headerValue)
	}

	// Validate that we don't have headers from previous attempts
	for i := 0; i < attempt; i++ {
		headerName := fmt.Sprintf("X-Foo-Test-%d", i)
		headerValue = responseRecorder.Header().Get("headerName")
		if headerValue != "" {
			t.Errorf("Expected no value for header %s, got %s", headerName, headerValue)
		}
	}
}

// countingRetryListener is a Listener implementation to count the times the Retried fn is called.
type countingRetryListener struct {
	timesCalled int
}

func (l *countingRetryListener) Retried(req *http.Request, attempt int) {
	l.timesCalled++
}

func TestRetryWithFlush(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(200)
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

	retry, err := New(context.Background(), next, dynamic.Retry{Attempts: 1}, &countingRetryListener{}, "traefikTest")
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

	forwarder, err := forward.New()
	if err != nil {
		t.Fatalf("Error creating forwarder: %v", err)
	}

	backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{}
		_, err := upgrader.Upgrade(rw, req, nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}))

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loadBalancer, err := roundrobin.New(forwarder)
			if err != nil {
				t.Fatalf("Error creating load balancer: %v", err)
			}

			// out of range port
			basePort := 1133444
			for i := 0; i < test.amountFaultyEndpoints; i++ {
				// 192.0.2.0 is a non-routable IP for testing purposes.
				// See: https://stackoverflow.com/questions/528538/non-routable-ip-address/18436928#18436928
				// We only use the port specification here because the URL is used as identifier
				// in the load balancer and using the exact same URL would not add a new server.
				_ = loadBalancer.UpsertServer(testhelpers.MustParseURL("http://192.0.2.0:" + strconv.Itoa(basePort+i)))
			}

			// add the functioning server to the end of the load balancer list
			err = loadBalancer.UpsertServer(testhelpers.MustParseURL(backendServer.URL))
			if err != nil {
				t.Fatalf("Fail to upsert server: %v", err)
			}

			retryListener := &countingRetryListener{}
			retryH, err := New(context.Background(), loadBalancer, dynamic.Retry{Attempts: test.maxRequestAttempts}, retryListener, "traefikTest")
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
