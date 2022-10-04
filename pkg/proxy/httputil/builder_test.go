package httputil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func BenchmarkProxy(b *testing.B) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/", nil)
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	builder := NewProxyBuilder()
	builder.roundTrippers = map[string]http.RoundTripper{"bench": &staticTransport{res: res}}

	proxy, err := builder.Build("bench", &dynamic.HTTPClientConfig{}, nil, req.URL)
	require.NoError(b, err)

	w := httptest.NewRecorder()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		proxy.ServeHTTP(w, req)
	}
}

func TestEscapedPath(t *testing.T) {
	var gotEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		gotEscapedPath = req.URL.EscapedPath()
	}))

	p, err := NewProxyBuilder().Build("default", &dynamic.HTTPClientConfig{PassHostHeader: true}, nil, testhelpers.MustParseURL(srv.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		p.ServeHTTP(rw, req)
	}))

	_, err = http.Get(proxy.URL + "/%3A%2F%2F")
	require.NoError(t, err)

	assert.Equal(t, "/%3A%2F%2F", gotEscapedPath)
}

type staticTransport struct {
	res *http.Response
}

func (t *staticTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.res, nil
}
