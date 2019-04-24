package blockpathregex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockPathRegex(t *testing.T) {
	testCases := []struct {
		desc            string
		path            string
		config          config.BlockPathRegex
		expectedCode    int
		expectedMessage string
		expectsError    bool
	}{
		{
			desc: "simple regex",
			path: "/whoami/and/whoami",
			config: config.BlockPathRegex{
				Regex:        `^/whoami/(.*)`,
				ResponseCode: 588,
				Message:      "foo",
			},
			expectedCode:    588,
			expectedMessage: "foo",
		},
		{
			desc: "invalid regular expression",
			path: "/invalid/regexp/test",
			config: config.BlockPathRegex{
				Regex:        `^(?err)/invalid/regexp/([^/]+)$`,
				ResponseCode: 588,
				Message:      "invalid",
			},
			expectsError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			handler, err := New(context.Background(), next, test.config, "foo-block-path-regexp")
			if test.expectsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)
				req.RequestURI = test.path

				rw := httptest.NewRecorder()
				handler.ServeHTTP(rw, req)

				if test.expectedCode != 0 {
					assert.Equal(t, test.expectedCode, rw.Code, "Unexpected response code: %d.", rw.Code)
				}

				if len(test.expectedMessage) > 0 {
					assert.Equal(t, test.expectedMessage, rw.Body.String(), "Unexpected response message: %s.", rw.Body.String())
				}

			}
		})
	}
}
