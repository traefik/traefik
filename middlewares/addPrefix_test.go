package middlewares

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestAddPrefix(t *testing.T) {

	path := "/bar"
	prefix := "/foo"

	var expectedPath string
	handler := &AddPrefix{
		Prefix: prefix,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath = r.URL.Path
		}),
	}

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+path, nil)

	handler.ServeHTTP(nil, req)

	assert.Equal(t, expectedPath, "/foo/bar", "Unexpected path.")
}
