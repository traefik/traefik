package proxy

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/server/service"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/traefik/traefik/v3/pkg/types"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func TestSmartBuilder_Build(t *testing.T) {
	tests := []struct {
		desc             string
		serversTransport dynamic.ServersTransport
		fastProxyConfig  static.FastProxyConfig
		https            bool
		h2c              bool
		wantFastProxy    bool
	}{
		{
			desc:            "fastproxy",
			fastProxyConfig: static.FastProxyConfig{Debug: true},
			wantFastProxy:   true,
		},
		{
			desc:            "fastproxy with https and without DisableHTTP2",
			https:           true,
			fastProxyConfig: static.FastProxyConfig{Debug: true},
			wantFastProxy:   false,
		},
		{
			desc:             "fastproxy with https and DisableHTTP2",
			https:            true,
			serversTransport: dynamic.ServersTransport{DisableHTTP2: true},
			fastProxyConfig:  static.FastProxyConfig{Debug: true},
			wantFastProxy:    true,
		},
		{
			desc:            "fastproxy with h2c",
			h2c:             true,
			fastProxyConfig: static.FastProxyConfig{Debug: true},
			wantFastProxy:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var callCount int
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if test.wantFastProxy {
					assert.Contains(t, r.Header, "X-Traefik-Fast-Proxy")
				} else {
					assert.NotContains(t, r.Header, "X-Traefik-Fast-Proxy")
				}
			})

			var server *httptest.Server

			if test.https {
				server = httptest.NewUnstartedServer(handler)
				server.EnableHTTP2 = false
				server.StartTLS()

				certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.TLS.Certificates[0].Certificate[0]})
				test.serversTransport.RootCAs = []types.FileOrContent{
					types.FileOrContent(certPEM),
				}
			} else {
				server = httptest.NewServer(h2c.NewHandler(handler, &http2.Server{}))
			}
			t.Cleanup(func() {
				server.Close()
			})

			targetURL := testhelpers.MustParseURL(server.URL)
			if test.h2c {
				targetURL.Scheme = "h2c"
			}

			serversTransports := map[string]*dynamic.ServersTransport{
				"test": &test.serversTransport,
			}

			transportManager := service.NewTransportManager(nil)
			transportManager.Update(serversTransports)

			httpProxyBuilder := httputil.NewProxyBuilder(transportManager, nil)
			proxyBuilder := NewSmartBuilder(transportManager, httpProxyBuilder, test.fastProxyConfig)

			proxyHandler, err := proxyBuilder.Build("test", targetURL, false, false, time.Second)
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			proxyHandler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/", http.NoBody))

			assert.Equal(t, 1, callCount)
		})
	}
}
