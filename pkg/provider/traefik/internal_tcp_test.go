package traefik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/file"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestTCPServersTransportProxyProtocol(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc          string
		args          []string
		content       string
		expected      *dynamic.ProxyProtocol
		expectedError string
	}{
		{
			desc: "omitted",
			args: []string{"--tcpserverstransport.dialtimeout=30s"},
		},
		{
			desc: "disabled",
			args: []string{"--tcpserverstransport.proxyprotocol.version=0"},
		},
		{
			desc:    "empty section",
			content: "tcpServersTransport:\n  proxyProtocol: {}\n",
		},
		{
			desc:     "version 1",
			args:     []string{"--tcpserverstransport.proxyprotocol.version=1"},
			expected: &dynamic.ProxyProtocol{Version: 1},
		},
		{
			desc:     "version 2",
			args:     []string{"--tcpserverstransport.proxyprotocol.version=2"},
			expected: &dynamic.ProxyProtocol{Version: 2},
		},
		{
			desc:          "negative version",
			args:          []string{"--tcpserverstransport.proxyprotocol.version=-1"},
			expected:      &dynamic.ProxyProtocol{Version: -1},
			expectedError: "unknown proxyProtocol version: -1",
		},
		{
			desc:          "unsupported version",
			args:          []string{"--tcpserverstransport.proxyprotocol.version=3"},
			expected:      &dynamic.ProxyProtocol{Version: 3},
			expectedError: "unknown proxyProtocol version: 3",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var staticCfg static.Configuration
			if test.content != "" {
				require.NoError(t, file.DecodeContent(test.content, ".yaml", &staticCfg))
			} else {
				require.NoError(t, flag.Decode(test.args, &staticCfg))
			}

			cfg := &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: make(map[string]*dynamic.TCPServersTransport),
				},
			}
			New(staticCfg).serverTransportTCP(cfg)
			transport := cfg.TCP.ServersTransports["default"]
			require.NotNil(t, transport)
			assert.Equal(t, test.expected, transport.ProxyProtocol)

			dialerManager := tcp.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": transport})
			_, err := dialerManager.Build(&dynamic.TCPServersLoadBalancer{}, false)
			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}
