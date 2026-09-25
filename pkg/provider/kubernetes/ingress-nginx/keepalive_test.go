package ingressnginx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildServersTransportKeepAlive(t *testing.T) {
	for _, test := range []struct {
		desc        string
		connections int
		reuse       bool
	}{
		{desc: "disabled", connections: 0},
		{desc: "enabled", connections: 1, reuse: true},
	} {
		t.Run(test.desc, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, _ = io.WriteString(rw, req.RemoteAddr)
			}))
			defer backend.Close()

			p := Provider{}
			p.SetDefaults()
			p.UpstreamKeepaliveConnections = test.connections
			nst, err := p.buildServersTransport(t.Context(), "default", "whoami", IngressConfig{})
			require.NoError(t, err)

			transport := &http.Transport{MaxIdleConnsPerHost: nst.MaxIdleConnsPerHost}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

			var addresses []string
			for range 2 {
				req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, backend.URL, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, resp.Body.Close())
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				addresses = append(addresses, string(body))
			}

			if test.reuse {
				assert.Equal(t, addresses[0], addresses[1])
			} else {
				assert.NotEqual(t, addresses[0], addresses[1])
			}
		})
	}
}
