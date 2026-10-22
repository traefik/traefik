package requestmirror

import (
	"context"
	"errors"
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

func TestRequestMirror(t *testing.T) {
	testCases := []struct {
		desc            string
		percent         int
		requests        int
		expectedMirrors int32
	}{
		{
			desc:            "mirror all requests",
			percent:         100,
			requests:        10,
			expectedMirrors: 10,
		},
		{
			desc:            "mirror half of the requests",
			percent:         50,
			requests:        10,
			expectedMirrors: 5,
		},
		{
			desc:            "mirror no request",
			percent:         0,
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
			handler, err := NewRequestMirror(t.Context(), next, dynamic.RequestMirror{Service: "mirror", Percent: test.percent}, builder, pool, "foo-request-mirror")
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
	handler, err := NewRequestMirror(t.Context(), headerModifier, dynamic.RequestMirror{Service: "mirror", Percent: 100}, builder, pool, "foo-request-mirror")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://foo.com/bar", http.NoBody)
	req.Header.Set("X-Foo", "original")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	wg.Wait()
	pool.Stop()

	assert.Equal(t, "modified", servedHeader)
	assert.Equal(t, "original", mirroredHeader)
}

func TestRequestMirror_unknownService(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := NewRequestMirror(t.Context(), next, dynamic.RequestMirror{Service: "unknown", Percent: 100}, serviceBuilderMock{}, safe.NewPool(t.Context()), "foo-request-mirror")
	require.Error(t, err)
}
