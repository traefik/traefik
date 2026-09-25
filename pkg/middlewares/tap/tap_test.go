package tap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// sink collects the records sent by the middleware, and stands for a Traefik service.
type sink struct {
	mu      sync.Mutex
	records []*record
	paths   []string

	status int
}

func (s *sink) BuildHTTP(_ context.Context, serviceName string) (http.Handler, error) {
	if serviceName == "unknown" {
		return nil, errors.New("service not found")
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rec := &record{}
		if err := json.Unmarshal(body, rec); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.mu.Lock()
		s.records = append(s.records, rec)
		s.paths = append(s.paths, req.URL.Path)
		s.mu.Unlock()

		if s.status != 0 {
			rw.WriteHeader(s.status)
		}
	}), nil
}

func (s *sink) recorded() []*record {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.records
}

func newTap(t *testing.T, config dynamic.Tap, next http.Handler, builder serviceBuilder) http.Handler {
	t.Helper()

	config.Timeout = ptypes.Duration(dynamic.TapDefaultTimeout)

	handler, err := New(context.Background(), next, config, builder, "tapTest")
	require.NoError(t, err)

	return handler
}

func requestRecordConfig() *dynamic.TapRecord {
	config := &dynamic.TapRecord{}
	config.SetDefaults()
	config.Service = "requests"

	return config
}

func responseRecordConfig() *dynamic.TapRecord {
	config := &dynamic.TapRecord{}
	config.SetDefaults()
	config.Service = "responses"

	return config
}

func TestNew(t *testing.T) {
	testCases := []struct {
		desc        string
		config      dynamic.Tap
		expectedErr string
	}{
		{
			desc:        "no destination",
			config:      dynamic.Tap{},
			expectedErr: "at least one of request or response must be defined",
		},
		{
			desc:        "destination without service",
			config:      dynamic.Tap{Request: &dynamic.TapRecord{}},
			expectedErr: "building request destination: service must be defined",
		},
		{
			desc:        "unknown service",
			config:      dynamic.Tap{Response: &dynamic.TapRecord{Service: "unknown"}},
			expectedErr: "building response destination: service not found",
		},
		{
			desc: "request headers on the request records",
			config: dynamic.Tap{
				Request: &dynamic.TapRecord{Service: "requests", RequestHeaders: []string{"X-Client"}},
			},
			expectedErr: "requestHeaders is only allowed on the response records",
		},
		{
			desc: "request only",
			config: dynamic.Tap{
				Request: requestRecordConfig(),
			},
		},
		{
			desc: "request and response",
			config: dynamic.Tap{
				Request:  requestRecordConfig(),
				Response: responseRecordConfig(),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := New(context.Background(), http.NotFoundHandler(), test.config, &sink{}, "tapTest")
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestServeHTTP_records(t *testing.T) {
	s := &sink{}

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, "ping", string(body))

		rw.Header().Set("X-Backend", "yes")
		rw.WriteHeader(http.StatusCreated)
		_, err = rw.Write([]byte("pong"))
		require.NoError(t, err)
	})

	handler := newTap(t, dynamic.Tap{
		Request:  requestRecordConfig(),
		Response: responseRecordConfig(),
	}, next, s)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo?bar=baz", strings.NewReader("ping"))
	req.Header.Set("X-Client", "yes")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "pong", recorder.Body.String())
	assert.Equal(t, "yes", recorder.Header().Get("X-Backend"))

	records := s.recorded()
	require.Len(t, records, 2)

	assert.Equal(t, kindRequest, records[0].Kind)
	require.NotNil(t, records[0].Request)
	assert.Equal(t, http.MethodPost, records[0].Request.Method)
	assert.Equal(t, "/foo?bar=baz", records[0].Request.URL)
	assert.Equal(t, "example.com", records[0].Request.Host)
	assert.Equal(t, "yes", records[0].Request.Headers.Get("X-Client"))
	assert.Equal(t, "ping", string(records[0].Request.Body))
	assert.False(t, records[0].Request.BodyTruncated)

	assert.Equal(t, kindResponse, records[1].Kind)
	require.NotNil(t, records[1].Response)
	assert.Equal(t, http.StatusCreated, records[1].Response.Status)
	assert.Equal(t, "yes", records[1].Response.Headers.Get("X-Backend"))
	assert.Equal(t, "pong", string(records[1].Response.Body))
	assert.Positive(t, records[1].Response.Duration)

	// Both records belong to the same exchange.
	assert.Equal(t, records[0].ID, records[1].ID)
	assert.Equal(t, []string{"/", "/"}, s.paths)
}

func TestServeHTTP_requestOnly(t *testing.T) {
	s := &sink{}

	handler := newTap(t, dynamic.Tap{Request: requestRecordConfig()}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, err := rw.Write([]byte("pong"))
		require.NoError(t, err)
	}), s)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.Equal(t, "pong", recorder.Body.String())

	records := s.recorded()
	require.Len(t, records, 1)
	assert.Equal(t, kindRequest, records[0].Kind)
	assert.Nil(t, records[0].Response)
}

func TestServeHTTP_bodyOptions(t *testing.T) {
	testCases := []struct {
		desc              string
		body              *bool
		maxBodySize       *int64
		expectedBody      string
		expectedTruncated bool
	}{
		{
			desc:         "body kept whole by default",
			expectedBody: "ping",
		},
		{
			desc: "body disabled",
			body: new(false),
		},
		{
			desc:              "body truncated",
			maxBodySize:       new(int64(2)),
			expectedBody:      "pi",
			expectedTruncated: true,
		},
		{
			desc:        "body at the limit",
			maxBodySize: new(int64(4)),
			// A body exactly at the limit is not flagged as truncated.
			expectedBody: "ping",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := &sink{}

			requestConfig := requestRecordConfig()
			if test.body != nil {
				requestConfig.Body = test.body
			}
			if test.maxBodySize != nil {
				requestConfig.MaxBodySize = test.maxBodySize
			}

			responseConfig := responseRecordConfig()
			responseConfig.Body = requestConfig.Body
			responseConfig.MaxBodySize = requestConfig.MaxBodySize

			var forwarded string
			handler := newTap(t, dynamic.Tap{
				Request:  requestConfig,
				Response: responseConfig,
			}, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				forwarded = string(body)

				// The response body matches the request one, to assert both sides at once.
				_, err = rw.Write(body)
				require.NoError(t, err)
			}), s)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader("ping")))

			// Truncating a record must not truncate what the backend and the client get.
			assert.Equal(t, "ping", forwarded)
			assert.Equal(t, "ping", recorder.Body.String())

			records := s.recorded()
			require.Len(t, records, 2)

			assert.Equal(t, test.expectedBody, string(records[0].Request.Body))
			assert.Equal(t, test.expectedTruncated, records[0].Request.BodyTruncated)
			assert.Equal(t, test.expectedBody, string(records[1].Response.Body))
			assert.Equal(t, test.expectedTruncated, records[1].Response.BodyTruncated)
		})
	}
}

func TestServeHTTP_failClosed(t *testing.T) {
	testCases := []struct {
		desc           string
		failClosed     bool
		sinkStatus     int
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			desc:           "sink accepts the records",
			failClosed:     true,
			expectedStatus: http.StatusCreated,
			expectedBody:   "pong",
			expectedHeader: "yes",
		},
		{
			desc:           "failing closed, sink rejects the records",
			failClosed:     true,
			sinkStatus:     http.StatusServiceUnavailable,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError) + "\n",
		},
		{
			desc:           "not failing closed, sink rejects the records",
			sinkStatus:     http.StatusServiceUnavailable,
			expectedStatus: http.StatusCreated,
			expectedBody:   "pong",
			expectedHeader: "yes",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := &sink{status: test.sinkStatus}

			responseConfig := responseRecordConfig()
			responseConfig.FailClosed = test.failClosed

			handler := newTap(t, dynamic.Tap{
				Response: responseConfig,
			}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Set("X-Backend", "yes")
				rw.WriteHeader(http.StatusCreated)
				_, err := rw.Write([]byte("pong"))
				require.NoError(t, err)
			}), s)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

			assert.Equal(t, test.expectedStatus, recorder.Code)
			assert.Equal(t, test.expectedBody, recorder.Body.String())
			// A rejected response must not leak the backend headers.
			assert.Equal(t, test.expectedHeader, recorder.Header().Get("X-Backend"))
		})
	}
}

func TestServeHTTP_failClosedRejectsRequest(t *testing.T) {
	s := &sink{status: http.StatusServiceUnavailable}

	requestConfig := requestRecordConfig()
	requestConfig.FailClosed = true

	var called bool
	handler := newTap(t, dynamic.Tap{
		Request: requestConfig,
	}, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}), s)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.False(t, called, "the request must not reach the backend when its record cannot be sent")
}

// TestServeHTTP_failClosedPerDirection asserts that the directions fail independently:
// a response record that cannot be sent does not withhold the response when only the
// request is failing closed.
func TestServeHTTP_failClosedPerDirection(t *testing.T) {
	requestConfig := requestRecordConfig()
	requestConfig.FailClosed = true

	builder := serviceBuilderFunc(func(_ context.Context, _ string) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get(headerRecordKind) == kindResponse {
				rw.WriteHeader(http.StatusServiceUnavailable)
			}
		}), nil
	})

	var called bool
	handler := newTap(t, dynamic.Tap{
		Request:  requestConfig,
		Response: responseRecordConfig(),
	}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		called = true

		rw.WriteHeader(http.StatusCreated)
		_, err := rw.Write([]byte("pong"))
		require.NoError(t, err)
	}), builder)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "pong", recorder.Body.String())
}

func TestServeHTTP_noResponseFromNext(t *testing.T) {
	s := &sink{}

	handler := newTap(t, dynamic.Tap{Response: responseRecordConfig()}, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}), s)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.Equal(t, http.StatusOK, recorder.Code)

	records := s.recorded()
	require.Len(t, records, 1)
	assert.Equal(t, http.StatusOK, records[0].Response.Status)
	assert.Empty(t, records[0].Response.Body)
}

func TestReadBody(t *testing.T) {
	testCases := []struct {
		desc              string
		body              string
		maxBodySize       int64
		expectedRecorded  string
		expectedTruncated bool
	}{
		{
			desc:             "no limit",
			body:             "ping",
			maxBodySize:      -1,
			expectedRecorded: "ping",
		},
		{
			desc:        "no body kept",
			body:        "ping",
			maxBodySize: 0,
		},
		{
			desc:              "body over the limit",
			body:              "ping",
			maxBodySize:       3,
			expectedRecorded:  "pin",
			expectedTruncated: true,
		},
		{
			desc:             "body under the limit",
			body:             "ping",
			maxBodySize:      10,
			expectedRecorded: "ping",
		},
		{
			desc:             "empty body",
			maxBodySize:      10,
			expectedRecorded: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader(test.body))

			recorded, truncated, err := readBody(req, test.maxBodySize)
			require.NoError(t, err)

			assert.Equal(t, test.expectedRecorded, string(recorded))
			assert.Equal(t, test.expectedTruncated, truncated)

			// The next handler must still read the body whole.
			forwarded, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Equal(t, test.body, string(forwarded))
			assert.NoError(t, req.Body.Close())
		})
	}
}

// TestServeHTTP_failClosedHoldsResponse asserts the guarantee of the failClosed mode:
// the response record is accepted before any byte of the response reaches the client.
func TestServeHTTP_failClosedHoldsResponse(t *testing.T) {
	recorder := httptest.NewRecorder()

	builder := serviceBuilderFunc(func(_ context.Context, _ string) (http.Handler, error) {
		return http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			assert.Empty(t, recorder.Body.String())
			assert.Empty(t, recorder.Header().Get("X-Backend"))
		}), nil
	})

	responseConfig := responseRecordConfig()
	responseConfig.FailClosed = true

	handler := newTap(t, dynamic.Tap{
		Response: responseConfig,
	}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("X-Backend", "yes")
		_, err := rw.Write([]byte("pong"))
		require.NoError(t, err)
	}), builder)

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.Equal(t, "pong", recorder.Body.String())
	assert.Equal(t, "yes", recorder.Header().Get("X-Backend"))
}

func TestServeHTTP_requestHeaders(t *testing.T) {
	s := &sink{}

	response := responseRecordConfig()
	response.RequestHeaders = []string{"X-Request-Id", "X-Absent"}

	handler := newTap(t, dynamic.Tap{Response: response}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}), s)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo?bar=baz", strings.NewReader("ping"))
	req.Header.Set("X-Request-Id", "42")
	req.Header.Set("X-Client", "yes")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	records := s.recorded()
	require.Len(t, records, 1)
	require.NotNil(t, records[0].Request)

	assert.Equal(t, http.MethodPost, records[0].Request.Method)
	assert.Equal(t, "/foo?bar=baz", records[0].Request.URL)
	assert.Equal(t, "example.com", records[0].Request.Host)
	assert.Equal(t, "42", records[0].Request.Headers.Get("X-Request-Id"))

	// Only the configured headers are described, and the body is never part of a response record.
	assert.Empty(t, records[0].Request.Headers.Get("X-Client"))
	assert.Len(t, records[0].Request.Headers, 1)
	assert.Empty(t, records[0].Request.Body)
}

func TestServeHTTP_failClosedRecordsHeadersSetUpstream(t *testing.T) {
	s := &sink{}

	responseConfig := responseRecordConfig()
	responseConfig.FailClosed = true

	handler := newTap(t, dynamic.Tap{
		Response: responseConfig,
	}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}), s)

	recorder := httptest.NewRecorder()
	// A middleware standing before the tap in the chain sets a response header.
	recorder.Header().Set("X-Upstream", "yes")

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.Equal(t, "yes", recorder.Header().Get("X-Upstream"))

	records := s.recorded()
	require.Len(t, records, 1)
	require.NotNil(t, records[0].Response)
	assert.Equal(t, "yes", records[0].Response.Headers.Get("X-Upstream"))
}

type serviceBuilderFunc func(ctx context.Context, serviceName string) (http.Handler, error)

func (f serviceBuilderFunc) BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error) {
	return f(ctx, serviceName)
}

// TestServeHTTP_expectContinue asserts that the Expect the middleware honored by reading
// the body is not forwarded: the backend would answer a second informational response,
// which the client has no reason to see.
func TestServeHTTP_expectContinue(t *testing.T) {
	testCases := []struct {
		desc            string
		expect          string
		body            *bool
		expectedForward string
	}{
		{
			desc:   "the honored expectation is dropped",
			expect: "100-continue",
		},
		{
			desc:   "case insensitive",
			expect: "100-Continue",
		},
		{
			desc:            "an expectation the middleware did not honor is forwarded",
			expect:          "other-expectation",
			expectedForward: "other-expectation",
		},
		{
			desc:            "nothing is read, nothing is dropped",
			expect:          "100-continue",
			body:            new(false),
			expectedForward: "100-continue",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			requestConfig := requestRecordConfig()
			if test.body != nil {
				requestConfig.Body = test.body
			}

			var forwarded string
			handler := newTap(t, dynamic.Tap{Request: requestConfig}, http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
				forwarded = req.Header.Get("Expect")
			}), &sink{})

			req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader("ping"))
			req.Header.Set("Expect", test.expect)

			handler.ServeHTTP(httptest.NewRecorder(), req)

			assert.Equal(t, test.expectedForward, forwarded)
		})
	}
}
