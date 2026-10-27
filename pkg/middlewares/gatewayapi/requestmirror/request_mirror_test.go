package requestmirror

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
)

type serviceBuilderMock map[string]http.Handler

func (s serviceBuilderMock) BuildHTTP(_ context.Context, serviceName string) (http.Handler, error) {
	handler, ok := s[serviceName]
	if !ok {
		return nil, errors.New("service not found")
	}
	return handler, nil
}

// mirrorResult is what the mirror service received.
type mirrorResult struct {
	req  *http.Request
	body []byte
	err  error
}

func TestRequestMirror(t *testing.T) {
	testCases := []struct {
		desc            string
		numerator       int64
		denominator     int64
		requests        int
		expectedMirrors int32
	}{
		{
			desc:            "mirror all requests",
			numerator:       1,
			denominator:     1,
			requests:        10,
			expectedMirrors: 10,
		},
		{
			desc:            "mirror half of the requests",
			numerator:       50,
			denominator:     100,
			requests:        10,
			expectedMirrors: 5,
		},
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

			var mirrored, served atomic.Int32
			builder := serviceBuilderMock{
				"mirror": http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					mirrored.Add(1)
				}),
			}
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				served.Add(1)
				rw.WriteHeader(http.StatusOK)
			})

			pool := safe.NewPool(t.Context())
			conf := dynamic.RequestMirror{Service: "mirror", Numerator: test.numerator, Denominator: test.denominator}
			handler, err := NewRequestMirror(t.Context(), next, conf, builder, pool, "foo-request-mirror")
			require.NoError(t, err)

			for range test.requests {
				rw := httptest.NewRecorder()
				handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "http://foo.com/bar", http.NoBody))
				assert.Equal(t, http.StatusOK, rw.Code)
			}

			pool.Stop()

			assert.Equal(t, int32(test.requests), served.Load())
			assert.Equal(t, test.expectedMirrors, mirrored.Load())
		})
	}
}

func TestRequestMirror_invalidFraction(t *testing.T) {
	testCases := []struct {
		desc        string
		numerator   int64
		denominator int64
	}{
		{
			desc:        "zero denominator",
			numerator:   0,
			denominator: 0,
		},
		{
			desc:        "negative numerator",
			numerator:   -1,
			denominator: 1,
		},
		{
			desc:        "numerator greater than denominator",
			numerator:   2,
			denominator: 1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			builder := serviceBuilderMock{"mirror": http.NotFoundHandler()}
			conf := dynamic.RequestMirror{Service: "mirror", Numerator: test.numerator, Denominator: test.denominator}

			_, err := NewRequestMirror(t.Context(), http.NotFoundHandler(), conf, builder, safe.NewPool(t.Context()), "foo-request-mirror")
			require.Error(t, err)
		})
	}
}

func TestRequestMirror_followingFiltersDoNotAlterMirroredRequest(t *testing.T) {
	var mirroredHeader, servedHeader string
	var wg sync.WaitGroup
	wg.Add(1)

	builder := serviceBuilderMock{
		"mirror": http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			defer wg.Done()
			mirroredHeader = req.Header.Get("X-Foo")
		}),
	}
	backend := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		servedHeader = req.Header.Get("X-Foo")
	})
	headerModifier := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Header.Set("X-Foo", "modified")
		backend.ServeHTTP(rw, req)
	})

	pool := safe.NewPool(t.Context())
	handler, err := NewRequestMirror(t.Context(), headerModifier, dynamic.RequestMirror{Service: "mirror", Numerator: 1, Denominator: 1}, builder, pool, "foo-request-mirror")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://foo.com/bar", http.NoBody)
	req.Header.Set("X-Foo", "original")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	wg.Wait()
	pool.Stop()

	assert.Equal(t, "modified", servedHeader)
	assert.Equal(t, "original", mirroredHeader)
}

func TestRequestMirror_body(t *testing.T) {
	body := make([]byte, 64*1024)
	_, err := rand.Read(body)
	require.NoError(t, err)

	results := make(chan mirrorResult, 1)
	builder := serviceBuilderMock{
		"mirror": http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			b, err := io.ReadAll(req.Body)
			results <- mirrorResult{req: req, body: b, err: err}
		}),
	}

	var servedBody []byte
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		servedBody, _ = io.ReadAll(req.Body)
	})

	pool := safe.NewPool(t.Context())
	handler, err := NewRequestMirror(t.Context(), next, dynamic.RequestMirror{Service: "mirror", Numerator: 1, Denominator: 1}, builder, pool, "foo-request-mirror")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "http://foo.com/bar", bytes.NewReader(body))
	req.Header.Set("Expect", "100-continue")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	result := <-results
	pool.Stop()

	assert.Equal(t, body, servedBody)
	require.NoError(t, result.err)
	assert.Equal(t, body, result.body)
	assert.Equal(t, int64(len(body)), result.req.ContentLength)
	assert.Empty(t, result.req.Header.Get("Expect"))
}

func TestRequestMirror_unknownService(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := NewRequestMirror(t.Context(), next, dynamic.RequestMirror{Service: "unknown", Numerator: 1, Denominator: 1}, serviceBuilderMock{}, safe.NewPool(t.Context()), "foo-request-mirror")
	require.Error(t, err)
}
