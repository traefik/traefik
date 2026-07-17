package snippet

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func Test_NewReturnsParseErrorForMalformedGonginxInputs(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testCases := []struct {
		desc    string
		snippet dynamic.Snippet
	}{
		{
			desc: "unterminated quoted string",
			snippet: dynamic.Snippet{
				ServerSnippet: `add_header X-Test "unterminated;`,
			},
		},
		{
			desc: "unterminated lua block with nested table",
			snippet: dynamic.Snippet{
				ServerSnippet: `content_by_lua_block { local t = { a = 1`,
			},
		},
		{
			desc: "include without path",
			snippet: dynamic.Snippet{
				ServerSnippet: `include;`,
			},
		},
		{
			desc: "include with multiple paths",
			snippet: dynamic.Snippet{
				ServerSnippet: `include first.conf second.conf;`,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			_, err := New(t.Context(), next, &test.snippet, "test-snippet")

			require.Error(t, err)
			require.ErrorContains(t, err, "parsing server-snippet:")
			require.NotContains(t, err.Error(), "snippet parsing recover:")
		})
	}
}
