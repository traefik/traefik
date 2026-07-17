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

	tests := []struct {
		name          string
		serverSnippet string
	}{
		{
			name:          "unterminated quoted string",
			serverSnippet: `add_header X-Test "unterminated;`,
		},
		{
			name:          "unterminated lua block with nested table",
			serverSnippet: `content_by_lua_block { local t = { a = 1`,
		},
		{
			name:          "include without path",
			serverSnippet: `include;`,
		},
		{
			name:          "include with multiple paths",
			serverSnippet: `include first.conf second.conf;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(t.Context(), next, &dynamic.Snippet{
				ServerSnippet: tt.serverSnippet,
			}, "test-snippet")

			require.Error(t, err)
			require.ErrorContains(t, err, "parsing server-snippet:")
			require.NotContains(t, err.Error(), "snippet parsing recover:")
		})
	}
}
