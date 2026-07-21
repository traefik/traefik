package failover

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type responseRecorder struct {
	*httptest.ResponseRecorder

	save     map[string]int
	sequence []string
	status   []int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
	r.sequence = append(r.sequence, r.Header().Get("server"))
	r.status = append(r.status, statusCode)
	r.ResponseRecorder.WriteHeader(statusCode)
}

func TestFailover(t *testing.T) {
	failover, err := New(&dynamic.Failover{
		HealthCheck: &dynamic.HealthCheck{},
	})
	require.NoError(t, err)

	status := true
	require.NoError(t, failover.RegisterStatusUpdater(func(up bool) {
		status = up
	}))

	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))

	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)
	assert.True(t, status)

	failover.SetHandlerStatus(t.Context(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 1, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)
	assert.True(t, status)

	failover.SetFallbackHandlerStatus(t.Context(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{503}, recorder.status)
	assert.False(t, status)
}

func TestFailoverDownThenUp(t *testing.T) {
	failover, err := New(&dynamic.Failover{})
	require.NoError(t, err)

	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))

	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetHandlerStatus(t.Context(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 1, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetHandlerStatus(t.Context(), true)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)
}

func TestFailoverPropagate(t *testing.T) {
	failover, err := New(&dynamic.Failover{
		HealthCheck: &dynamic.HealthCheck{},
	})
	require.NoError(t, err)
	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))
	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	topFailover, err := New(&dynamic.Failover{})
	require.NoError(t, err)
	topFailover.SetHandler(failover)
	topFailover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "topFailover")
		rw.WriteHeader(http.StatusOK)
	}))
	err = failover.RegisterStatusUpdater(func(up bool) {
		topFailover.SetHandlerStatus(t.Context(), up)
	})
	require.NoError(t, err)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	topFailover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, 0, recorder.save["topFailover"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetHandlerStatus(t.Context(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	topFailover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 1, recorder.save["fallback"])
	assert.Equal(t, 0, recorder.save["topFailover"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetFallbackHandlerStatus(t.Context(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	topFailover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, 1, recorder.save["topFailover"])
	assert.Equal(t, []int{200}, recorder.status)
}

func TestFailoverStatusCode(t *testing.T) {
	testCases := []struct {
		desc                 string
		statusCode           []string
		handlerStatusCode    int
		expectedHandler      string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			desc:                 "main handler returns 503, failover triggered",
			statusCode:           []string{"503"},
			handlerStatusCode:    503,
			expectedHandler:      "fallback",
			expectedStatusCode:   200,
			expectedResponseBody: "fallback response",
		},
		{
			desc:                 "main handler returns 200, failover not triggered",
			statusCode:           []string{"503"},
			handlerStatusCode:    200,
			expectedHandler:      "handler",
			expectedStatusCode:   200,
			expectedResponseBody: "main response",
		},
		{
			desc:                 "main handler returns 500, failover not triggered",
			statusCode:           []string{"503"},
			handlerStatusCode:    500,
			expectedHandler:      "handler",
			expectedStatusCode:   500,
			expectedResponseBody: "main response",
		},
		{
			desc:                 "multiple status codes, 503 triggers failover",
			statusCode:           []string{"500", "502", "503", "504"},
			handlerStatusCode:    503,
			expectedHandler:      "fallback",
			expectedStatusCode:   200,
			expectedResponseBody: "fallback response",
		},
		{
			desc:                 "multiple status codes, 502 triggers failover",
			statusCode:           []string{"500", "502", "503", "504"},
			handlerStatusCode:    502,
			expectedHandler:      "fallback",
			expectedStatusCode:   200,
			expectedResponseBody: "fallback response",
		},
		{
			desc:                 "multiple status codes, 404 does not trigger failover",
			statusCode:           []string{"500", "502", "503", "504"},
			handlerStatusCode:    404,
			expectedHandler:      "handler",
			expectedStatusCode:   404,
			expectedResponseBody: "main response",
		},
		{
			desc:                 "status code range 500-504, 503 triggers failover",
			statusCode:           []string{"500-504"},
			handlerStatusCode:    503,
			expectedHandler:      "fallback",
			expectedStatusCode:   200,
			expectedResponseBody: "fallback response",
		},
		{
			desc:                 "status code range 500-504, 404 does not trigger failover",
			statusCode:           []string{"500-504"},
			handlerStatusCode:    404,
			expectedHandler:      "handler",
			expectedStatusCode:   404,
			expectedResponseBody: "main response",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			failover, err := New(&dynamic.Failover{
				Errors: &dynamic.FailoverError{
					Status: test.statusCode,
				},
			})
			require.NoError(t, err)

			failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("server", "handler")
				rw.WriteHeader(test.handlerStatusCode)
				_, err := rw.Write([]byte("main response"))
				require.NoError(t, err)
			}))

			failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("server", "fallback")
				rw.WriteHeader(http.StatusOK)
				_, err := rw.Write([]byte("fallback response"))
				require.NoError(t, err)
			}))

			recorder := httptest.NewRecorder()
			failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.Equal(t, test.expectedHandler, recorder.Header().Get("server"))
			assert.Equal(t, test.expectedStatusCode, recorder.Code)
			assert.Equal(t, test.expectedResponseBody, recorder.Body.String())
		})
	}
}

func TestFailoverStatusCodeWithRequestBody(t *testing.T) {
	testCases := []struct {
		desc              string
		statusCode        []string
		handlerStatusCode int
		requestBody       string
		expectedHandler   string
	}{
		{
			desc:              "request body replayed to fallback handler",
			statusCode:        []string{"503"},
			handlerStatusCode: 503,
			requestBody:       "test request body",
			expectedHandler:   "fallback",
		},
		{
			desc:              "request body used by main handler",
			statusCode:        []string{"503"},
			handlerStatusCode: 200,
			requestBody:       "test request body",
			expectedHandler:   "handler",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var receivedBody string

			maxBody := int64(-1)
			failover, err := New(&dynamic.Failover{
				Errors: &dynamic.FailoverError{
					Status:              test.statusCode,
					MaxRequestBodyBytes: &maxBody,
				},
			})
			require.NoError(t, err)

			failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				body, _ := io.ReadAll(req.Body)
				receivedBody = string(body)
				rw.Header().Set("server", "handler")
				rw.WriteHeader(test.handlerStatusCode)
			}))

			failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				body, _ := io.ReadAll(req.Body)
				receivedBody = string(body)
				rw.Header().Set("server", "fallback")
				rw.WriteHeader(http.StatusOK)
			}))

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.requestBody))
			failover.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedHandler, recorder.Header().Get("server"))
			assert.Equal(t, test.requestBody, receivedBody)
		})
	}
}

func TestFailoverStatusCodeMaxBodySize(t *testing.T) {
	testCases := []struct {
		desc               string
		maxBodySize        int64
		requestBody        string
		expectedStatusCode int
		expectedMessage    string
	}{
		{
			desc:               "request body within limit",
			maxBodySize:        100,
			requestBody:        "small body",
			expectedStatusCode: 200,
			expectedMessage:    "",
		},
		{
			desc:               "request body exceeds limit",
			maxBodySize:        5,
			requestBody:        "this body is too large",
			expectedStatusCode: http.StatusRequestEntityTooLarge,
			expectedMessage:    "Request body too large\n",
		},
		{
			desc:               "zero body size limit",
			maxBodySize:        0,
			requestBody:        "any size body",
			expectedStatusCode: http.StatusRequestEntityTooLarge,
			expectedMessage:    "Request body too large\n",
		},
		{
			desc:               "no body size limit",
			maxBodySize:        -1,
			requestBody:        "any size body should work",
			expectedStatusCode: 200,
			expectedMessage:    "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			maxBody := test.maxBodySize
			failover, err := New(&dynamic.Failover{
				Errors: &dynamic.FailoverError{
					Status:              []string{"503"},
					MaxRequestBodyBytes: &maxBody,
				},
			})
			require.NoError(t, err)

			failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusServiceUnavailable)
			}))

			failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.requestBody))
			req.ContentLength = int64(len(test.requestBody))
			failover.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedStatusCode, recorder.Code)
			if test.expectedMessage != "" {
				assert.Equal(t, test.expectedMessage, recorder.Body.String())
			}
		})
	}
}

func TestFailoverStatusCodeWithHealthCheck(t *testing.T) {
	failover, err := New(&dynamic.Failover{
		HealthCheck: &dynamic.HealthCheck{},
		Errors: &dynamic.FailoverError{
			Status: []string{"503"},
		},
	})
	require.NoError(t, err)

	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusServiceUnavailable)
	}))

	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	// Test 1: Handler is up, returns 503, should failover based on status code
	recorder := httptest.NewRecorder()
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "fallback", recorder.Header().Get("server"))
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Test 2: Handler is marked down via health check, should failover
	failover.SetHandlerStatus(t.Context(), false)

	recorder = httptest.NewRecorder()
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "fallback", recorder.Header().Get("server"))
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Test 3: Handler is back up but returns non-503 status, should not failover
	failover.SetHandlerStatus(t.Context(), true)
	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))

	recorder = httptest.NewRecorder()
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "handler", recorder.Header().Get("server"))
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestFailoverInvalidStatusCodeRange(t *testing.T) {
	_, err := New(&dynamic.Failover{
		Errors: &dynamic.FailoverError{
			Status: []string{"invalid"},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strconv")
}
