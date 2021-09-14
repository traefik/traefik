package tcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_addTCPRoute(t *testing.T) {
	testCases := []struct {
		desc       string
		rule       string
		serverName string
		remoteAddr string
		routeErr   bool
		matchErr   bool
	}{
		{
			desc:     "no tree",
			routeErr: true,
		},
		{
			desc:     "Rule with no matcher",
			rule:     "rulewithnotmatcher",
			routeErr: true,
		},
		{
			desc:       "Empty HostSNI rule",
			rule:       "HostSNI()",
			serverName: "foobar",
			routeErr:   true,
		},
		{
			desc:       "Empty HostSNI rule",
			rule:       "HostSNI(``)",
			serverName: "foobar",
			routeErr:   true,
		},
		{
			desc:       "Valid HostSNI rule matching",
			rule:       "HostSNI(`foobar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNI rule matching with alternative case",
			rule:       "hostsni(`foobar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNI rule matching with alternative case",
			rule:       "HOSTSNI(`foobar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNI rule not matching",
			rule:       "HostSNI(`foobar`)",
			serverName: "bar",
			matchErr:   true,
		},
		{
			desc:     "Empty ClientIP rule",
			rule:     "ClientIP()",
			routeErr: true,
		},
		{
			desc:     "Empty ClientIP rule",
			rule:     "ClientIP(``)",
			routeErr: true,
		},
		{
			desc:     "Invalid ClientIP",
			rule:     "ClientIP(`invalid`)",
			routeErr: true,
		},
		{
			desc:       "Valid ClientIP rule matching",
			rule:       "ClientIP(`10.0.0.1`)",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid ClientIP rule matching with alternative case",
			rule:       "clientip(`10.0.0.1`)",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid ClientIP rule matching with alternative case",
			rule:       "CLIENTIP(`10.0.0.1`)",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid ClientIP rule not matching",
			rule:       "ClientIP(`10.0.0.1`)",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid ClientIP rule matching IPv6",
			rule:       "ClientIP(`10::10`)",
			remoteAddr: "[10::10]:80",
		},
		{
			desc:       "Valid ClientIP rule not matching IPv6",
			rule:       "ClientIP(`10::10`)",
			remoteAddr: "[::1]:80",
			matchErr:   true,
		},
		{
			desc:       "Valid ClientIP rule matching multiple IPs",
			rule:       "ClientIP(`10.0.0.1`, `10.0.0.0`)",
			remoteAddr: "10.0.0.0:80",
		},
		{
			desc:       "Valid ClientIP rule matching CIDR",
			rule:       "ClientIP(`11.0.0.0/24`)",
			remoteAddr: "11.0.0.0:80",
		},
		{
			desc:       "Valid ClientIP rule not matching CIDR",
			rule:       "ClientIP(`11.0.0.0/24`)",
			remoteAddr: "10.0.0.0:80",
			matchErr:   true,
		},
		{
			desc:       "Valid ClientIP rule matching CIDR IPv6",
			rule:       "ClientIP(`11::/16`)",
			remoteAddr: "[11::]:80",
		},
		{
			desc:       "Valid ClientIP rule not matching CIDR IPv6",
			rule:       "ClientIP(`11::/16`)",
			remoteAddr: "[10::]:80",
			matchErr:   true,
		},
		{
			desc:       "Valid ClientIP rule matching multiple CIDR",
			rule:       "ClientIP(`11.0.0.0/16`, `10.0.0.0/16`)",
			remoteAddr: "10.0.0.0:80",
		},
		{
			desc:       "Valid ClientIP rule not matching CIDR and matching IP",
			rule:       "ClientIP(`11.0.0.0/16`, `10.0.0.0`)",
			remoteAddr: "10.0.0.0:80",
		},
		{
			desc:       "Valid ClientIP rule matching CIDR and not matching IP",
			rule:       "ClientIP(`11.0.0.0`, `10.0.0.0/16`)",
			remoteAddr: "10.0.0.0:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP rule matching",
			rule:       "HostSNI(`foobar`) && ClientIP(`10.0.0.1`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP rule not matching",
			rule:       "HostSNI(`foobar`) && ClientIP(`10.0.0.1`)",
			serverName: "bar",
			remoteAddr: "10.0.0.1:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP rule not matching",
			rule:       "HostSNI(`foobar`) && ClientIP(`10.0.0.1`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI or ClientIP rule matching",
			rule:       "HostSNI(`foobar`) || ClientIP(`10.0.0.1`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI or ClientIP rule matching",
			rule:       "HostSNI(`foobar`) || ClientIP(`10.0.0.1`)",
			serverName: "bar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI or ClientIP rule matching",
			rule:       "HostSNI(`foobar`) || ClientIP(`10.0.0.1`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.2:80",
		},
		{
			desc:       "Valid HostSNI or ClientIP rule not matching",
			rule:       "HostSNI(`foobar`) || ClientIP(`10.0.0.1`)",
			serverName: "bar",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI x 3 OR rule matching",
			rule:       "HostSNI(`foobar`) || HostSNI(`foo`) || HostSNI(`bar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNI x 3 OR rule not matching",
			rule:       "HostSNI(`foobar`) || HostSNI(`foo`) || HostSNI(`bar`)",
			serverName: "baz",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule matching",
			rule:       "HostSNI(`foobar`) || HostSNI(`bar`) && ClientIP(`10.0.0.1`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.2:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule matching",
			rule:       "HostSNI(`foobar`) || HostSNI(`bar`) && ClientIP(`10.0.0.1`)",
			serverName: "bar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule not matching",
			rule:       "HostSNI(`foobar`) || HostSNI(`bar`) && ClientIP(`10.0.0.1`)",
			serverName: "bar",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule not matching",
			rule:       "HostSNI(`foobar`) || HostSNI(`bar`) && ClientIP(`10.0.0.1`)",
			serverName: "baz",
			remoteAddr: "10.0.0.1:80",
			matchErr:   true,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			msg := "BYTES"
			var handler HandlerFunc
			handler = func(conn WriteCloser) {
				conn.Write([]byte(msg))
			}
			router, err := NewTCPRouterMux()
			require.NoError(t, err)

			err = router.AddRoute(test.rule, handler)
			if test.routeErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				addr := "0.0.0.0:0"
				if test.remoteAddr != "" {
					addr = test.remoteAddr
				}

				conn := &fakeConn{
					call:       map[string]int{},
					remoteAddr: fakeAddr{addr: addr},
				}

				metaTCP, err := NewMetaTCP(test.serverName, conn)
				require.NoError(t, err)

				handler := router.Match(metaTCP)
				if test.matchErr {
					require.Nil(t, handler)
					return
				}

				require.NotNil(t, handler)

				handler.ServeTCP(conn)

				n, ok := conn.call[msg]
				assert.Equal(t, n, 1)
				assert.True(t, ok)

				//assert.Equal(t, test.expected, results)
			}
		})
	}
}

type fakeAddr struct {
	addr string
}

func (f fakeAddr) String() string {
	return f.addr
}

func (f fakeAddr) Network() string {
	panic("Implement me")
}

func Test_MaRoute(t *testing.T) {

}
