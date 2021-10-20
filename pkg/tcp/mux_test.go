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
			handler := HandlerFunc(func(conn WriteCloser) {
				_, err := conn.Write([]byte(msg))
				require.NoError(t, err)
			})
			router, err := NewMuxer()
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

				connData, err := NewConnData(test.serverName, conn)
				require.NoError(t, err)

				handler := router.Match(connData)
				if test.matchErr {
					require.Nil(t, handler)
					return
				}

				require.NotNil(t, handler)

				handler.ServeTCP(conn)

				n, ok := conn.call[msg]
				assert.Equal(t, n, 1)
				assert.True(t, ok)

				// assert.Equal(t, test.expected, results)
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

func TestParseHostSNI(t *testing.T) {
	testCases := []struct {
		description   string
		expression    string
		domain        []string
		errorExpected bool
	}{
		{
			description:   "Many hostSNI rules",
			expression:    "HostSNI(`foo.bar`,`test.bar`)",
			domain:        []string{"foo.bar", "test.bar"},
			errorExpected: false,
		},
		{
			description:   "Many hostSNI rules upper",
			expression:    "HOSTSNI(`foo.bar`,`test.bar`)",
			domain:        []string{"foo.bar", "test.bar"},
			errorExpected: false,
		},
		{
			description:   "Many hostSNI rules lower",
			expression:    "hostsni(`foo.bar`,`test.bar`)",
			domain:        []string{"foo.bar", "test.bar"},
			errorExpected: false,
		},
		{
			description:   "No hostSNI rule",
			expression:    "ClientIP(`10.1`)",
			errorExpected: false,
		},
		{
			description:   "HostSNI rule and another rule",
			expression:    "HostSNI(`foo.bar`) && ClientIP(`10.1`)",
			domain:        []string{"foo.bar"},
			errorExpected: false,
		},
		{
			description:   "HostSNI rule to lower and another rule",
			expression:    "HostSNI(`Foo.Bar`) && ClientIP(`10.1`)",
			domain:        []string{"foo.bar"},
			errorExpected: false,
		},
		{
			description:   "HostSNI rule with no domain",
			expression:    "HostSNI() && ClientIP(`10.1`)",
			errorExpected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			domains, err := ParseHostSNI(test.expression)

			if test.errorExpected {
				require.Errorf(t, err, "unable to parse correctly the domains in the HostSNI rule from %q", test.expression)
			} else {
				require.NoError(t, err, "%s: Error while parsing domain.", test.expression)
			}

			assert.EqualValues(t, test.domain, domains, "%s: Error parsing domains from expression.", test.expression)
		})
	}
}

func Test_HostSNI(t *testing.T) {
	testCases := []struct {
		desc       string
		ruleHosts  []string
		serverName string
		buildErr   bool
		matchErr   bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:      "Non ASCII host",
			ruleHosts: []string{"héhé"},
			buildErr:  true,
		},
		{
			desc:       "Not Matching hosts",
			ruleHosts:  []string{"foobar"},
			serverName: "bar",
			matchErr:   true,
		},
		{
			desc:       "Matching globing host `*`",
			ruleHosts:  []string{"*"},
			serverName: "foobar",
		},
		{
			desc:      "Not Matching globing host with subdomain",
			ruleHosts: []string{"*.bar"},
			buildErr:  true,
		},
		{
			desc:       "Not Matching host with trailing dot with ",
			ruleHosts:  []string{"foobar."},
			serverName: "foobar.",
		},
		{
			desc:       "Matching host with trailing dot",
			ruleHosts:  []string{"foobar."},
			serverName: "foobar",
		},
		{
			desc:       "Matching hosts",
			ruleHosts:  []string{"foobar"},
			serverName: "foobar",
		},
		{
			desc:       "Matching hosts with subdomains",
			ruleHosts:  []string{"foo.bar"},
			serverName: "foo.bar",
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var route route
			err := hostSNI(&route, test.ruleHosts...)
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			meta := ConnData{
				serverName: test.serverName,
			}

			if test.matchErr {
				assert.False(t, route.match(meta))
			} else {
				assert.True(t, route.match(meta))
			}
		})
	}
}

// TODO have a parse client ip test?
func Test_ClientIP(t *testing.T) {
	testCases := []struct {
		desc       string
		ruleHosts  []string
		serverName string
		buildErr   bool
		matchErr   bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:      "Non ASCII host",
			ruleHosts: []string{"héhé"},
			buildErr:  true,
		},
		{
			desc:       "Not Matching hosts",
			ruleHosts:  []string{"foobar"},
			serverName: "bar",
			matchErr:   true,
		},
		{
			desc:       "Matching globing host `*`",
			ruleHosts:  []string{"*"},
			serverName: "foobar",
		},
		{
			desc:      "Not Matching globing host with subdomain",
			ruleHosts: []string{"*.bar"},
			buildErr:  true,
		},
		{
			desc:       "Not Matching host with trailing dot with ",
			ruleHosts:  []string{"foobar."},
			serverName: "foobar.",
		},
		{
			desc:       "Matching host with trailing dot",
			ruleHosts:  []string{"foobar."},
			serverName: "foobar",
		},
		{
			desc:       "Matching hosts",
			ruleHosts:  []string{"foobar"},
			serverName: "foobar",
		},
		{
			desc:       "Matching hosts with subdomains",
			ruleHosts:  []string{"foo.bar"},
			serverName: "foo.bar",
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var route route
			err := hostSNI(&route, test.ruleHosts...)
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			meta := ConnData{
				serverName: test.serverName,
			}

			if test.matchErr {
				assert.False(t, route.match(meta))
			} else {
				assert.True(t, route.match(meta))
			}
		})
	}
}
