package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		desc             string
		scheme           string
		serversTransport *dynamic.ServersTransport
		fastProxyConfig  *static.FastProxyConfig
		wantFastProxy    bool
	}{
		{
			desc:             "httputil",
			serversTransport: &dynamic.ServersTransport{},
			wantFastProxy:    false,
		},
		{
			desc:             "fastproxy",
			serversTransport: &dynamic.ServersTransport{},
			fastProxyConfig:  &static.FastProxyConfig{Debug: true},
			wantFastProxy:    true,
		},
		{
			desc:             "fastproxy with disable HTTP2",
			serversTransport: &dynamic.ServersTransport{DisableHTTP2: true},
			fastProxyConfig:  &static.FastProxyConfig{Debug: true},
			wantFastProxy:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewTLSServer()

			var callCount int
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++

				if test.wantFastProxy {
					assert.Contains(t, r.Header, "X-Traefik-Fast-Proxy")
				} else {
					assert.NotContains(t, r.Header, "X-Traefik-Fast-Proxy")
				}
			}))
			t.Cleanup(func() {
				s.Close()
			})

			transportManager := transportManagerMock{ServersTransport: test.serversTransport}

			b := NewBuilder(transportManager, nil, test.fastProxyConfig)
			handler, err := b.Build("foo", testhelpers.MustParseURL(s.URL), false, false, time.Second)
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/", http.NoBody))

			assert.Equal(t, 1, callCount)
		})
	}
}

type transportManagerMock struct {
	TLSConfig        *tls.Config
	ServersTransport *dynamic.ServersTransport
}

func (t transportManagerMock) Get(_ string) (*dynamic.ServersTransport, error) {
	return t.ServersTransport, nil
}

func (t transportManagerMock) GetRoundTripper(_ string) (http.RoundTripper, error) {
	return http.DefaultTransport, nil
}

func (t transportManagerMock) GetTLSConfig(_ string) (*tls.Config, error) {
	return t.TLSConfig, nil
}
