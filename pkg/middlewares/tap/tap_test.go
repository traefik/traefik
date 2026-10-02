package tap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

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
			config:      dynamic.Tap{Request: &dynamic.TapRequest{}},
			expectedErr: "building request destination: service must be defined",
		},
		{
			desc:        "unknown service",
			config:      dynamic.Tap{Response: &dynamic.TapResponse{TapRequest: dynamic.TapRequest{Service: "unknown"}}},
			expectedErr: "building response destination: building tap service handler: service not found",
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

	requestConfig := requestRecordConfig()
	requestConfig.RecordBody = true

	responseConfig := responseRecordConfig()
	responseConfig.RecordBody = true

	handler := newTap(t, dynamic.Tap{
		Request:  requestConfig,
		Response: responseConfig,
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
		body              bool
		maxBodySize       *int64
		expectedBody      string
		expectedTruncated bool
	}{
		{
			desc: "body dropped by default",
		},
		{
			desc:         "body enabled and kept whole",
			body:         true,
			expectedBody: "ping",
		},
		{
			desc:              "body truncated",
			body:              true,
			maxBodySize:       new(int64(2)),
			expectedBody:      "pi",
			expectedTruncated: true,
		},
		{
			desc:        "body at the limit",
			body:        true,
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
			requestConfig.RecordBody = test.body
			if test.maxBodySize != nil {
				requestConfig.MaxRecordBodySize = test.maxBodySize
			}

			responseConfig := responseRecordConfig()
			responseConfig.RecordBody = requestConfig.RecordBody
			responseConfig.MaxRecordBodySize = requestConfig.MaxRecordBodySize

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

// TestServeHTTP_interimRecordStatus asserts that an informational status from a tap service
// is not taken for the acceptance of the record.
func TestServeHTTP_interimRecordStatus(t *testing.T) {
	builder := serviceBuilderFunc(func(_ context.Context, _ string) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusEarlyHints)
			rw.WriteHeader(http.StatusServiceUnavailable)
		}), nil
	})

	requestConfig := requestRecordConfig()
	requestConfig.FailClosed = true

	var called bool
	handler := newTap(t, dynamic.Tap{Request: requestConfig}, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}), builder)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	assert.False(t, called, "the request must not reach the backend when its record is rejected")
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// TestServeHTTP_earlyHints asserts that an interim response reaches the client, headers included,
// only when the response is streamed.
func TestServeHTTP_earlyHints(t *testing.T) {
	testCases := []struct {
		desc          string
		failClosed    bool
		expectedHints []string
	}{
		{
			desc:          "streamed response",
			expectedHints: []string{"</app.css>; rel=preload"},
		},
		{
			desc:       "withheld response",
			failClosed: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			responseConfig := responseRecordConfig()
			responseConfig.FailClosed = test.failClosed

			handler := newTap(t, dynamic.Tap{Response: responseConfig}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Add("Link", "</app.css>; rel=preload")
				rw.WriteHeader(http.StatusEarlyHints)

				rw.Header().Set("X-Backend", "yes")
				rw.WriteHeader(http.StatusCreated)
			}), &sink{})

			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			var hints []string
			ctx := httptrace.WithClientTrace(t.Context(), &httptrace.ClientTrace{
				Got1xxResponse: func(_ int, header textproto.MIMEHeader) error {
					hints = append(hints, header.Get("Link"))
					return nil
				},
			})

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, http.NoBody)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			assert.Equal(t, test.expectedHints, hints)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			assert.Equal(t, "</app.css>; rel=preload", resp.Header.Get("Link"))
			assert.Equal(t, "yes", resp.Header.Get("X-Backend"))
		})
	}
}

// TestServeHTTP_responseTrailers asserts that the trailers of a response, declared or set with the
// http.TrailerPrefix, reach the client as trailers only, whether the response is withheld or not.
func TestServeHTTP_responseTrailers(t *testing.T) {
	testCases := []struct {
		desc       string
		failClosed bool
	}{
		{
			desc: "streamed response",
		},
		{
			desc:       "withheld response",
			failClosed: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			responseConfig := responseRecordConfig()
			responseConfig.FailClosed = test.failClosed

			handler := newTap(t, dynamic.Tap{Response: responseConfig}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Set("Content-Type", "application/grpc")
				rw.Header().Set("Trailer", "X-Checksum")
				rw.Header().Set(http.TrailerPrefix+"X-Early", "yes")
				rw.WriteHeader(http.StatusOK)

				_, err := rw.Write([]byte("pong"))
				require.NoError(t, err)

				rw.Header().Set("X-Checksum", "abc")
				rw.Header().Set(http.TrailerPrefix+"Grpc-Status", "0")
			}), &sink{})

			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			assert.Equal(t, "pong", string(body))
			assert.Equal(t, "application/grpc", resp.Header.Get("Content-Type"))
			assert.Empty(t, resp.Header.Values("X-Checksum"))
			assert.Equal(t, []string{"abc"}, resp.Trailer.Values("X-Checksum"))
			assert.Equal(t, []string{"0"}, resp.Trailer.Values("Grpc-Status"))
			assert.Equal(t, []string{"yes"}, resp.Trailer.Values("X-Early"))
		})
	}
}

// TestServeHTTP_earlyHintsRejected asserts that the headers set for an informational response
// reach the client neither with it nor with the error of a rejected response.
func TestServeHTTP_earlyHintsRejected(t *testing.T) {
	responseConfig := responseRecordConfig()
	responseConfig.FailClosed = true

	handler := newTap(t, dynamic.Tap{Response: responseConfig}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Add("Link", "</app.css>; rel=preload")
		rw.Header().Set("X-Backend", "yes")
		rw.WriteHeader(http.StatusEarlyHints)

		rw.WriteHeader(http.StatusCreated)
	}), &sink{status: http.StatusServiceUnavailable})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	var hints []string
	ctx := httptrace.WithClientTrace(t.Context(), &httptrace.ClientTrace{
		Got1xxResponse: func(_ int, header textproto.MIMEHeader) error {
			hints = append(hints, header.Get("Link"))
			return nil
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Empty(t, hints)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Link"))
	assert.Empty(t, resp.Header.Get("X-Backend"))
}

// TestServeHTTP_withheldExpectContinue asserts that a withheld response does not hold back the
// 100 Continue, which the server sends when the body is read.
func TestServeHTTP_withheldExpectContinue(t *testing.T) {
	responseConfig := responseRecordConfig()
	responseConfig.FailClosed = true

	handler := newTap(t, dynamic.Tap{Response: responseConfig}, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusContinue)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		_, err = rw.Write(body)
		require.NoError(t, err)
	}), &sink{})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	var continues int
	ctx := httptrace.WithClientTrace(t.Context(), &httptrace.ClientTrace{
		Got100Continue: func() { continues++ },
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, strings.NewReader("ping"))
	require.NoError(t, err)
	req.Header.Set("Expect", "100-continue")

	// The client would rather time out than send its body without the 100 Continue.
	client := &http.Client{
		Transport: &http.Transport{ExpectContinueTimeout: time.Hour},
		Timeout:   5 * time.Second,
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Equal(t, 1, continues)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ping", string(body))
}

// TestServeHTTP_timeoutPerDirection asserts that the request and response records are each sent
// within their own timeout, a zero timeout meaning no limit.
func TestServeHTTP_timeoutPerDirection(t *testing.T) {
	requestConfig := requestRecordConfig()
	requestConfig.Timeout = ptypes.Duration(time.Second)

	responseConfig := responseRecordConfig()
	responseConfig.Timeout = ptypes.Duration(time.Hour)

	noLimitConfig := responseRecordConfig()
	noLimitConfig.Service = "unlimited"
	noLimitConfig.Timeout = 0

	type deadline struct {
		remaining time.Duration
		ok        bool
	}

	deadlines := make(map[string]deadline)
	builder := serviceBuilderFunc(func(_ context.Context, serviceName string) (http.Handler, error) {
		return http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
			d, ok := req.Context().Deadline()
			deadlines[serviceName] = deadline{remaining: time.Until(d), ok: ok}
		}), nil
	})

	handler := newTap(t, dynamic.Tap{Request: requestConfig, Response: responseConfig}, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}), builder)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	handler = newTap(t, dynamic.Tap{Response: noLimitConfig}, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}), builder)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

	require.Contains(t, deadlines, "requests")
	assert.True(t, deadlines["requests"].ok)
	assert.Positive(t, deadlines["requests"].remaining)
	assert.LessOrEqual(t, deadlines["requests"].remaining, time.Second)

	require.Contains(t, deadlines, "responses")
	assert.True(t, deadlines["responses"].ok)
	assert.Greater(t, deadlines["responses"].remaining, time.Second)
	assert.LessOrEqual(t, deadlines["responses"].remaining, time.Hour)

	require.Contains(t, deadlines, "unlimited")
	assert.False(t, deadlines["unlimited"].ok)
}

// TestServeHTTP_failClosedPerDirection asserts that the directions fail independently:
// a response record failing to be sent does not withhold the response when only the request fails closed.
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
		maxRecordBodySize int64
		expectedRecorded  string
		expectedTruncated bool
	}{
		{
			desc:              "no limit",
			body:              "ping",
			maxRecordBodySize: -1,
			expectedRecorded:  "ping",
		},
		{
			desc:              "no body kept",
			body:              "ping",
			maxRecordBodySize: 0,
		},
		{
			desc:              "body over the limit",
			body:              "ping",
			maxRecordBodySize: 3,
			expectedRecorded:  "pin",
			expectedTruncated: true,
		},
		{
			desc:              "body under the limit",
			body:              "ping",
			maxRecordBodySize: 10,
			expectedRecorded:  "ping",
		},
		{
			desc:              "empty body",
			maxRecordBodySize: 10,
			expectedRecorded:  "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader(test.body))

			recorded, truncated, err := readBody(req, test.maxRecordBodySize, -1)
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

func TestReadBody_maxBodySize(t *testing.T) {
	testCases := []struct {
		desc              string
		body              io.Reader
		maxRecordBodySize int64
		maxBodySize       int64
		expectedErr       error
		expectedUnread    bool
		expectedRecorded  string
		expectedTruncated bool
	}{
		{
			desc:              "body at the limit",
			body:              strings.NewReader("ping"),
			maxRecordBodySize: -1,
			maxBodySize:       4,
			expectedRecorded:  "ping",
		},
		{
			// A body announced too large is not read, so that no 100 Continue is sent.
			desc:              "body announced over the limit",
			body:              strings.NewReader("ping"),
			maxRecordBodySize: -1,
			maxBodySize:       3,
			expectedErr:       errBodyTooLarge,
			expectedUnread:    true,
		},
		{
			// A reader of unknown length leaves the content length unset, as a chunked body does.
			desc:              "body of unknown length over the limit",
			body:              io.MultiReader(strings.NewReader("ping")),
			maxRecordBodySize: -1,
			maxBodySize:       3,
			expectedErr:       errBodyTooLarge,
		},
		{
			desc:              "no limit",
			body:              strings.NewReader("ping"),
			maxRecordBodySize: -1,
			maxBodySize:       -1,
			expectedRecorded:  "ping",
		},
		{
			// The body is not held whole with a record limit, so there is nothing to bound.
			desc:              "record limited",
			body:              strings.NewReader("ping"),
			maxRecordBodySize: 2,
			maxBodySize:       3,
			expectedRecorded:  "pi",
			expectedTruncated: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", test.body)

			recorded, truncated, err := readBody(req, test.maxRecordBodySize, test.maxBodySize)
			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)

				if test.expectedUnread {
					unread, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					assert.Equal(t, "ping", string(unread))
				}

				return
			}

			require.NoError(t, err)

			assert.Equal(t, test.expectedRecorded, string(recorded))
			assert.Equal(t, test.expectedTruncated, truncated)

			// The next handler must still read the body whole.
			forwarded, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Equal(t, "ping", string(forwarded))
		})
	}
}

// TestServeHTTP_requestBodyTooLarge asserts that a request body held whole and over maxBodySize
// is rejected, without reaching the backend nor being recorded.
func TestServeHTTP_requestBodyTooLarge(t *testing.T) {
	s := &sink{}

	requestConfig := requestRecordConfig()
	requestConfig.RecordBody = true
	requestConfig.MaxBodySize = new(int64(3))

	var called bool
	handler := newTap(t, dynamic.Tap{Request: requestConfig}, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}), s)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader("ping")))

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.False(t, called)
	assert.Empty(t, s.recorded())
}

// TestServeHTTP_responseBodyTooLarge asserts that a withheld response over maxBodySize is replaced
// by an error and not recorded, while a streamed response is not limited.
func TestServeHTTP_responseBodyTooLarge(t *testing.T) {
	testCases := []struct {
		desc           string
		failClosed     bool
		expectedStatus int
		expectedBody   string
		expectedHeader string
		expectedRecord bool
	}{
		{
			desc:           "streamed response",
			expectedStatus: http.StatusCreated,
			expectedBody:   "pong",
			expectedHeader: "yes",
			expectedRecord: true,
		},
		{
			desc:           "withheld response",
			failClosed:     true,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   http.StatusText(http.StatusInternalServerError) + "\n",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := &sink{}

			responseConfig := responseRecordConfig()
			responseConfig.FailClosed = test.failClosed
			responseConfig.MaxBodySize = new(int64(3))

			handler := newTap(t, dynamic.Tap{Response: responseConfig}, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.Header().Set("X-Backend", "yes")
				rw.WriteHeader(http.StatusCreated)

				// The writes going over the limit are not refused, so that the next handler completes.
				for _, chunk := range []string{"po", "ng"} {
					n, err := rw.Write([]byte(chunk))
					require.NoError(t, err)
					assert.Equal(t, len(chunk), n)
				}
			}), s)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://example.com/foo", http.NoBody))

			assert.Equal(t, test.expectedStatus, recorder.Code)
			assert.Equal(t, test.expectedBody, recorder.Body.String())
			assert.Equal(t, test.expectedHeader, recorder.Header().Get("X-Backend"))
			assert.Equal(t, test.expectedRecord, len(s.recorded()) == 1)
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

// TestServeHTTP_expectContinue asserts that the Expect honored by reading the body is not forwarded,
// so that the backend does not answer a second informational response.
func TestServeHTTP_expectContinue(t *testing.T) {
	testCases := []struct {
		desc            string
		expect          string
		body            bool
		expectedForward string
	}{
		{
			desc:   "the honored expectation is dropped",
			expect: "100-continue",
			body:   true,
		},
		{
			desc:   "case insensitive",
			expect: "100-Continue",
			body:   true,
		},
		{
			desc:            "an expectation the middleware did not honor is forwarded",
			expect:          "other-expectation",
			body:            true,
			expectedForward: "other-expectation",
		},
		{
			desc:            "nothing is read, nothing is dropped",
			expect:          "100-continue",
			expectedForward: "100-continue",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			requestConfig := requestRecordConfig()
			requestConfig.RecordBody = test.body

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

	handler, err := New(context.Background(), next, config, builder, "tapTest")
	require.NoError(t, err)

	return handler
}

func requestRecordConfig() *dynamic.TapRequest {
	config := &dynamic.TapRequest{}
	config.SetDefaults()
	config.Service = "requests"

	return config
}

func responseRecordConfig() *dynamic.TapResponse {
	config := &dynamic.TapResponse{}
	config.SetDefaults()
	config.Service = "responses"

	return config
}

type serviceBuilderFunc func(ctx context.Context, serviceName string) (http.Handler, error)

func (f serviceBuilderFunc) BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error) {
	return f(ctx, serviceName)
}
