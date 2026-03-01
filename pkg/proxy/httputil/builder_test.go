package httputil

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestEscapedPath(t *testing.T) {
	var gotEscapedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		gotEscapedPath = req.URL.EscapedPath()
	}))

	transportManager := &transportManagerMock{
		roundTrippers: map[string]http.RoundTripper{"default": &http.Transport{}},
	}

	p, err := NewProxyBuilder(transportManager, nil).Build("default", testhelpers.MustParseURL(srv.URL), true, false, 0)
	require.NoError(t, err)

	proxy := httptest.NewServer(http.HandlerFunc(p.ServeHTTP))

	_, err = http.Get(proxy.URL + "/%3A%2F%2F")
	require.NoError(t, err)

	assert.Equal(t, "/%3A%2F%2F", gotEscapedPath)
}

type transportManagerMock struct {
	roundTrippers map[string]http.RoundTripper
}

func (t *transportManagerMock) GetRoundTripper(name string) (http.RoundTripper, error) {
	roundTripper, ok := t.roundTrippers[name]
	if !ok {
		return nil, errors.New("no transport for " + name)
	}

	return roundTripper, nil
}

func (t *transportManagerMock) GetTLSConfig(_ string) (*tls.Config, error) {
	panic("implement me")
}

func (t *transportManagerMock) Get(_ string) (*dynamic.ServersTransport, error) {
	panic("implement me")
}
