package kerberos

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFn func(req *http.Request) (*http.Response, error)

func (r roundTripperFn) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}

func TestRoundTripper(t *testing.T) {
	testCases := []struct {
		desc                   string
		authorizations         []string
		expectedStatusCode     []int
		expectedDedicatedCount int
		expectedOriginalCount  int
	}{
		{
			desc:                  "without Authorization header",
			authorizations:        []string{"", "", ""},
			expectedStatusCode:    []int{http.StatusUnauthorized, http.StatusUnauthorized, http.StatusUnauthorized},
			expectedOriginalCount: 3,
		},
		{
			desc:                   "with a preauthenticated Negotiate (Kerberos) credential on the first request",
			authorizations:         []string{"Negotiate dj1jdA==", "", ""},
			expectedStatusCode:     []int{http.StatusOK, http.StatusOK, http.StatusOK},
			expectedDedicatedCount: 3,
		},
		{
			desc:                   "with a preauthenticated NTLM credential on the first request",
			authorizations:         []string{"NTLM TlRMTVNTUAAB", "", ""},
			expectedStatusCode:     []int{http.StatusOK, http.StatusOK, http.StatusOK},
			expectedDedicatedCount: 3,
		},
		{
			desc:                   "with a challenge-first Negotiate flow",
			authorizations:         []string{"", "Negotiate dj1jdA==", ""},
			expectedStatusCode:     []int{http.StatusUnauthorized, http.StatusOK, http.StatusOK},
			expectedOriginalCount:  1,
			expectedDedicatedCount: 2,
		},
		{
			desc:                   "with a lowercase negotiate scheme",
			authorizations:         []string{"negotiate dj1jdA==", "", ""},
			expectedStatusCode:     []int{http.StatusOK, http.StatusOK, http.StatusOK},
			expectedDedicatedCount: 3,
		},
		{
			desc:                   "with a lowercase ntlm scheme",
			authorizations:         []string{"ntlm TlRMTVNTUAAB", "", ""},
			expectedStatusCode:     []int{http.StatusOK, http.StatusOK, http.StatusOK},
			expectedDedicatedCount: 3,
		},
		{
			desc:                  "with a scheme that only starts like NTLM",
			authorizations:        []string{"NTLMish dj1jdA==", "", ""},
			expectedStatusCode:    []int{http.StatusUnauthorized, http.StatusUnauthorized, http.StatusUnauthorized},
			expectedOriginalCount: 3,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			origCount := 0
			dedicatedCount := 0
			rt := NewRoundTripper(
				roundTripperFn(func(req *http.Request) (*http.Response, error) {
					origCount++
					return &http.Response{StatusCode: http.StatusUnauthorized}, nil
				}),
				func() http.RoundTripper {
					return roundTripperFn(func(req *http.Request) (*http.Response, error) {
						dedicatedCount++
						return &http.Response{StatusCode: http.StatusOK}, nil
					})
				},
			)

			ctx := AddTransportOnContext(t.Context())
			for i, expected := range test.expectedStatusCode {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1", http.NoBody)
				require.NoError(t, err)
				if test.authorizations[i] != "" {
					req.Header.Set("Authorization", test.authorizations[i])
				}
				resp, err := rt.RoundTrip(req)
				require.NoError(t, err)
				require.Equal(t, expected, resp.StatusCode)
			}

			require.Equal(t, test.expectedOriginalCount, origCount)
			require.Equal(t, test.expectedDedicatedCount, dedicatedCount)
		})
	}
}

func TestStickyRoundTrippersStickOnlyOnce(t *testing.T) {
	var newCalls int

	owner := &roundTripper{
		new: func() http.RoundTripper {
			newCalls++

			return http.DefaultTransport
		},
	}

	connRoundTrippers := &stickyRoundTrippers{}
	first := connRoundTrippers.stick(owner)
	second := connRoundTrippers.stick(owner)

	assert.Equal(t, 1, newCalls)
	assert.NotNil(t, first)
	assert.Same(t, first, second)
	assert.Same(t, first, connRoundTrippers.get(owner))
}

func TestRoundTripperSticksOnceOnConcurrentRequests(t *testing.T) {
	var newCalls atomic.Int32

	rt := NewRoundTripper(
		roundTripperFn(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusUnauthorized}, nil
		}),
		func() http.RoundTripper {
			newCalls.Add(1)

			return roundTripperFn(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK}, nil
			})
		},
	)

	ctx := AddTransportOnContext(t.Context())

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1", http.NoBody)
			assert.NoError(t, err)
			req.Header.Set("Authorization", "NTLM TlRMTVNTUAAB")

			_, err = rt.RoundTrip(req)
			assert.NoError(t, err)
		})
	}

	wg.Wait()

	assert.EqualValues(t, 1, newCalls.Load())
}
