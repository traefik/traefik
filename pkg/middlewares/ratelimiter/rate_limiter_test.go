package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulcand/oxy/utils"
)

func TestNewRateLimiter(t *testing.T) {
	testCases := []struct {
		desc             string
		config           dynamic.RateLimit
		expectedMaxDelay time.Duration
		expectedSourceIP string
	}{
		{
			desc: "maxDelay computation",
			config: dynamic.RateLimit{
				Average: 200,
				Burst:   10,
			},
			expectedMaxDelay: 2500 * time.Microsecond,
		},
		{
			desc: "default SourceMatcher is remote address ip strategy",
			config: dynamic.RateLimit{
				Average: 200,
				Burst:   10,
			},
			expectedSourceIP: "127.0.0.1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			h, err := New(context.Background(), next, test.config, "rate-limiter")
			require.NoError(t, err)
			rtl, _ := h.(*rateLimiter)
			if test.expectedMaxDelay != 0 {
				assert.Equal(t, test.expectedMaxDelay, rtl.maxDelay)
			}
			if test.expectedSourceIP != "" {
				extractor, ok := rtl.sourceMatcher.(utils.ExtractorFunc)
				if !ok {
					t.Fatal("Not an ExtractorFunc")
				}
				req := http.Request{
					RemoteAddr: fmt.Sprintf("%s:1234", test.expectedSourceIP),
				}
				ip, _, err := extractor(&req)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedSourceIP, ip)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	testCases := []struct {
		desc     string
		config   dynamic.RateLimit
		reqCount int
	}{
		{
			desc: "Average is respected",
			config: dynamic.RateLimit{
				Average: 100,
				Burst:   1,
			},
			reqCount: 200,
		},
		{
			desc: "Burst is taken into account",
			config: dynamic.RateLimit{
				Average: 100,
				Burst:   200,
			},
			reqCount: 300,
		},
		{
			desc: "Zero average ==> no rate limiting",
			config: dynamic.RateLimit{
				Average: 0,
				Burst:   1,
			},
			reqCount: 100,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			reqCount := 0
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reqCount++
			})
			h, err := New(context.Background(), next, test.config, "rate-limiter")
			require.NoError(t, err)
			start := time.Now()
			for {
				if reqCount >= test.reqCount {
					break
				}
				req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
				req.RemoteAddr = "127.0.0.1:1234"
				w := httptest.NewRecorder()
				h.ServeHTTP(w, req)
				// TODO(mpl): predict and count the 200 VS the 429?
			}
			stop := time.Now()
			elapsed := stop.Sub(start)
			if test.config.Average == 0 {
				if elapsed > time.Millisecond {
					t.Fatalf("rate should not have been limited, but: %d requests in %v", reqCount, elapsed)
				}
				return
			}
			// Assume allowed burst is initially consumed in an infinitesimal period of time
			var expectedDuration time.Duration
			if test.config.Average != 0 {
				expectedDuration = time.Duration((int64(test.reqCount)-test.config.Burst+1)/test.config.Average) * time.Second
			}
			// Allow for a 1% leeway
			minDuration := expectedDuration * 98 / 100
			maxDuration := expectedDuration * 102 / 100
			if elapsed < minDuration {
				t.Fatalf("rate was faster than expected: %d requests in %v", reqCount, elapsed)
			}
			if elapsed > maxDuration {
				t.Fatalf("rate was slower than expected: %d requests in %v", reqCount, elapsed)
			}
		})
	}
}
