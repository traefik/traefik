package encodedcharacters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestEncodedCharacters(t *testing.T) {
	testCases := []struct {
		desc               string
		config             dynamic.EncodedCharacters
		path               string
		expectedStatusCode int
	}{
		{
			desc:               "deny encoded slash",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%2fbar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded slash",
			config: dynamic.EncodedCharacters{
				AllowEncodedSlash: true,
			},
			path:               "/foo%2fbar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "deny encoded backslash",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%5cbar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded backslash",
			config: dynamic.EncodedCharacters{
				AllowEncodedBackSlash: true,
			},
			path:               "/foo%5cbar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "deny encoded null character",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%00bar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded null character",
			config: dynamic.EncodedCharacters{
				AllowEncodedNullCharacter: true,
			},
			path:               "/foo%00bar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "deny encoded semi colon",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%3bbar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded semi colon",
			config: dynamic.EncodedCharacters{
				AllowEncodedSemicolon: true,
			},
			path:               "/foo%3bbar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "deny encoded percent",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%25bar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded percent",
			config: dynamic.EncodedCharacters{
				AllowEncodedPercent: true,
			},
			path:               "/foo%25bar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "deny encoded question mark",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%3fbar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded question mark",
			config: dynamic.EncodedCharacters{
				AllowEncodedQuestionMark: true,
			},
			path:               "/foo%3fbar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc:               "deny encoded hash",
			config:             dynamic.EncodedCharacters{},
			path:               "/foo%23bar",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			desc: "allow encoded hash",
			config: dynamic.EncodedCharacters{
				AllowEncodedHash: true,
			},
			path:               "/foo%23bar",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
			handler := NewEncodedCharacters(t.Context(), next, test.config, "test-encoded-characters")

			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			require.Equal(t, test.expectedStatusCode, recorder.Code)
		})
	}
}

func TestMapDeniedCharacters(t *testing.T) {
	testCases := []struct {
		desc               string
		config             dynamic.EncodedCharacters
		expectedDeniedChar map[string]struct{}
	}{
		{
			desc:   "deny all characters",
			config: dynamic.EncodedCharacters{},
			expectedDeniedChar: map[string]struct{}{
				"%2F": {}, "%2f": {}, // slash
				"%5C": {}, "%5c": {}, // backslash
				"%00": {},            // null
				"%3B": {}, "%3b": {}, // semicolon
				"%25": {},            // percent
				"%3F": {}, "%3f": {}, // question mark
				"%23": {}, // hash
			},
		},
		{
			desc: "allow only encoded slash",
			config: dynamic.EncodedCharacters{
				AllowEncodedSlash: true,
			},
			expectedDeniedChar: map[string]struct{}{
				"%5C": {}, "%5c": {}, // backslash
				"%00": {},            // null
				"%3B": {}, "%3b": {}, // semicolon
				"%25": {},            // percent
				"%3F": {}, "%3f": {}, // question mark
				"%23": {}, // hash
			},
		},
		{
			desc: "allow all characters",
			config: dynamic.EncodedCharacters{
				AllowEncodedSlash:         true,
				AllowEncodedBackSlash:     true,
				AllowEncodedNullCharacter: true,
				AllowEncodedSemicolon:     true,
				AllowEncodedPercent:       true,
				AllowEncodedQuestionMark:  true,
				AllowEncodedHash:          true,
			},
			expectedDeniedChar: map[string]struct{}{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			deniedMap := mapDeniedCharacters(test.config)
			require.Equal(t, test.expectedDeniedChar, deniedMap)
			require.Len(t, deniedMap, len(test.expectedDeniedChar))
		})
	}
}
