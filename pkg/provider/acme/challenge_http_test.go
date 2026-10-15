package acme

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeHTTP_ServeHTTP_unansweredRequestIsNotAnError(t *testing.T) {
	testCases := []struct {
		desc string
		path string
	}{
		{
			desc: "token that was never presented",
			path: "/.well-known/acme-challenge/unknown-token",
		},
		{
			desc: "path that carries no token",
			path: "/.well-known/acme-challenge/",
		},
		{
			desc: "scanner probing for a PHP file",
			path: "/.well-known/acme-challenge/xmrlpc.php",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Anyone can reach this handler on the entry point serving the HTTP-01
			// challenge, so a request it cannot answer must not be logged as an error:
			// that lets a stranger drive the volume of the operator's error log.
			var buf bytes.Buffer

			logger := zerolog.New(&buf).Level(zerolog.ErrorLevel)

			challenge := NewChallengeHTTP()

			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			req.Host = "example.com"
			req = req.WithContext(logger.WithContext(t.Context()))

			recorder := httptest.NewRecorder()

			challenge.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusNotFound, recorder.Code)
			assert.Empty(t, buf.String(), "an unanswerable challenge request must not log at error level")
		})
	}
}

func TestChallengeHTTP_ServeHTTP_presentedTokenIsStillServed(t *testing.T) {
	challenge := NewChallengeHTTP()

	require.NoError(t, challenge.Present(t.Context(), "example.com", "known-token", "key-auth"))

	req := httptest.NewRequest(http.MethodGet, "/.well-known/acme-challenge/known-token", nil)
	req.Host = "example.com"
	req = req.WithContext(log.Logger.WithContext(t.Context()))

	recorder := httptest.NewRecorder()

	challenge.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "key-auth", recorder.Body.String())
}
