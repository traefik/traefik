package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strings"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

func TestRetry(t *testing.T) {
	testCases := []struct {
		desc                  string
		maxRequestAttempts    int
		wantRetryAttempts     int
		wantResponseStatus    int
		amountFaultyEndpoints int
	}{
		{
			desc:                  "no retry on success",
			maxRequestAttempts:    1,
			wantRetryAttempts:     0,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 0,
		},
		{
			desc:                  "no retry when max request attempts is one",
			maxRequestAttempts:    1,
			wantRetryAttempts:     0,
			wantResponseStatus:    http.StatusInternalServerError,
			amountFaultyEndpoints: 1,
		},
		{
			desc:                  "one retry when one server is faulty",
			maxRequestAttempts:    2,
			wantRetryAttempts:     1,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 1,
		},
		{
			desc:                  "two retries when two servers are faulty",
			maxRequestAttempts:    3,
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusOK,
			amountFaultyEndpoints: 2,
		},
		{
			desc:                  "max attempts exhausted delivers the 5xx response",
			maxRequestAttempts:    3,
			wantRetryAttempts:     2,
			wantResponseStatus:    http.StatusInternalServerError,
			amountFaultyEndpoints: 3,
		},
	}

	backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK"))
	}))

	forwarder, err := forward.New()
	if err != nil {
		t.Fatalf("Error creating forwarder: %s", err)
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loadBalancer, err := roundrobin.New(forwarder)
			if err != nil {
				t.Fatalf("Error creating load balancer: %s", err)
			}

			basePort := 33444
			for i := 0; i < test.amountFaultyEndpoints; i++ {
				// 192.0.2.0 is a non-routable IP for testing purposes.
				// See: https://stackoverflow.com/questions/528538/non-routable-ip-address/18436928#18436928
				// We only use the port specification here because the URL is used as identifier
				// in the load balancer and using the exact same URL would not add a new server.
				loadBalancer.UpsertServer(testhelpers.MustParseURL("http://192.0.2.0:" + string(basePort+i)))
			}

			// add the functioning server to the end of the load balancer list
			loadBalancer.UpsertServer(testhelpers.MustParseURL(backendServer.URL))

			retryListener := &countingRetryListener{}
			retry := NewRetry(test.maxRequestAttempts, loadBalancer, retryListener)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://localhost:3000/ok", nil)

			retry.ServeHTTP(recorder, req)

			assert.Equal(t, test.wantResponseStatus, recorder.Code)
			assert.Equal(t, test.wantRetryAttempts, retryListener.timesCalled)
		})
	}
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
		t.Fatalf("Error creating forwarder: %s", err)
	}

	backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{}
		upgrader.Upgrade(rw, req, nil)
	}))

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loadBalancer, err := roundrobin.New(forwarder)
			if err != nil {
				t.Fatalf("Error creating load balancer: %s", err)
			}

			basePort := 33444
			for i := 0; i < test.amountFaultyEndpoints; i++ {
				// 192.0.2.0 is a non-routable IP for testing purposes.
				// See: https://stackoverflow.com/questions/528538/non-routable-ip-address/18436928#18436928
				// We only use the port specification here because the URL is used as identifier
				// in the load balancer and using the exact same URL would not add a new server.
				loadBalancer.UpsertServer(testhelpers.MustParseURL("http://192.0.2.0:" + string(basePort+i)))
			}

			// add the functioning server to the end of the load balancer list
			loadBalancer.UpsertServer(testhelpers.MustParseURL(backendServer.URL))

			retryListener := &countingRetryListener{}
			retry := NewRetry(test.maxRequestAttempts, loadBalancer, retryListener)

			retryServer := httptest.NewServer(retry)

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

func TestRetryEmptyServerList(t *testing.T) {
	forwarder, err := forward.New()
	if err != nil {
		t.Fatalf("Error creating forwarder: %s", err)
	}

	loadBalancer, err := roundrobin.New(forwarder)
	if err != nil {
		t.Fatalf("Error creating load balancer: %s", err)
	}

	// The EmptyBackendHandler middleware ensures that there is a 503
	// response status set when there is no backend server in the pool.
	next := NewEmptyBackendHandler(loadBalancer)

	retryListener := &countingRetryListener{}
	retry := NewRetry(3, next, retryListener)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3000/ok", nil)

	retry.ServeHTTP(recorder, req)

	const wantResponseStatus = http.StatusServiceUnavailable
	if wantResponseStatus != recorder.Code {
		t.Errorf("got status code %d, want %d", recorder.Code, wantResponseStatus)
	}
	const wantRetryAttempts = 0
	if wantRetryAttempts != retryListener.timesCalled {
		t.Errorf("retry listener called %d time(s), want %d time(s)", retryListener.timesCalled, wantRetryAttempts)
	}
}

func TestRetryListeners(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	retryListeners := RetryListeners{&countingRetryListener{}, &countingRetryListener{}}

	retryListeners.Retried(req, 1)
	retryListeners.Retried(req, 1)

	for _, retryListener := range retryListeners {
		listener := retryListener.(*countingRetryListener)
		if listener.timesCalled != 2 {
			t.Errorf("retry listener was called %d time(s), want %d time(s)", listener.timesCalled, 2)
		}
	}
}

// countingRetryListener is a RetryListener implementation to count the times the Retried fn is called.
type countingRetryListener struct {
	timesCalled int
}

func (l *countingRetryListener) Retried(req *http.Request, attempt int) {
	l.timesCalled++
}

func TestRetryWithFlush(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(200)
		rw.Write([]byte("FULL "))
		rw.(http.Flusher).Flush()
		rw.Write([]byte("DATA"))
	})

	retry := NewRetry(1, next, &countingRetryListener{})
	responseRecorder := httptest.NewRecorder()

	retry.ServeHTTP(responseRecorder, &http.Request{})

	if responseRecorder.Body.String() != "FULL DATA" {
		t.Errorf("Wrong body %q want %q", responseRecorder.Body.String(), "FULL DATA")
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

	retry := NewRetry(3, next, &countingRetryListener{})
	responseRecorder := httptest.NewRecorder()
	retry.ServeHTTP(responseRecorder, &http.Request{})

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
