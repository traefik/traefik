package ready

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc        string
		setReady    bool
		terminating bool
		termCode    int
		wantStatus  int
	}{
		{
			desc:       "not ready returns 503",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			desc:       "ready returns 200",
			setReady:   true,
			wantStatus: http.StatusOK,
		},
		{
			desc:        "terminating wins over ready",
			setReady:    true,
			terminating: true,
			termCode:    http.StatusServiceUnavailable,
			wantStatus:  http.StatusServiceUnavailable,
		},
		{
			desc:        "terminating uses configured status code",
			setReady:    true,
			terminating: true,
			termCode:    http.StatusGone,
			wantStatus:  http.StatusGone,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			h := &Handler{}
			h.SetDefaults()
			if test.termCode != 0 {
				h.TerminatingStatusCode = test.termCode
			}
			if test.setReady {
				h.SetReady()
			}
			if test.terminating {
				h.terminating.Store(true)
			}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/ready", http.NoBody)
			h.ServeHTTP(rr, req)

			assert.Equal(t, test.wantStatus, rr.Code)
			assert.Equal(t, http.StatusText(test.wantStatus), rr.Body.String())
		})
	}
}

func TestHandler_SetDefaults(t *testing.T) {
	h := &Handler{}
	h.SetDefaults()

	assert.Equal(t, "traefik", h.EntryPoint)
	assert.Equal(t, http.StatusServiceUnavailable, h.TerminatingStatusCode)
}

func TestHandler_WithContext(t *testing.T) {
	h := &Handler{}
	h.SetDefaults()
	h.SetReady()

	ctx, cancel := context.WithCancel(context.Background())
	h.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", http.NoBody))
	require.Equal(t, http.StatusOK, rr.Code)

	cancel()

	assert.Eventually(t, func() bool {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", http.NoBody))
		return rr.Code == http.StatusServiceUnavailable
	}, time.Second, 10*time.Millisecond)
}
