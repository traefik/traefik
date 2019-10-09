package replacepath

import (
	"context"
	"net/http"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacePath(t *testing.T) {
	var replacementConfig = dynamic.ReplacePath{
		Path: "/replacement-path",
	}

	paths := []string{
		"/example",
		"/some/really/long/path",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			var expectedPath, actualHeader, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath = r.URL.Path
				actualHeader = r.Header.Get(ReplacedPathHeader)
				requestURI = r.RequestURI
			})

			handler, err := New(context.Background(), next, replacementConfig, "foo-replace-path")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+path, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, expectedPath, replacementConfig.Path, "Unexpected path.")
			assert.Equal(t, path, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)
			assert.Equal(t, expectedPath, requestURI, "Unexpected request URI.")
		})
	}
}
