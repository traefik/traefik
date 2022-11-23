package tcp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

func Test_addTCPRoute(t *testing.T) {
	testCases := []struct {
		desc       string
		rule       string
		serverName string
		remoteAddr string
		protos     []string
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
			rule:       "HostSNI(``)",
			serverName: "example.org",
			routeErr:   true,
		},
		{
			desc:       "Valid HostSNI rule matching",
			rule:       "HostSNI(`example.org`)",
			serverName: "example.org",
		},
		{
			desc:       "Valid negative HostSNI rule matching",
			rule:       "!HostSNI(`example.com`)",
			serverName: "example.org",
		},
		{
			desc:       "Valid HostSNI rule matching with alternative case",
			rule:       "hostsni(`example.org`)",
			serverName: "example.org",
		},
		{
			desc:       "Valid HostSNI rule matching with alternative case",
			rule:       "HOSTSNI(`example.org`)",
			serverName: "example.org",
		},
		{
			desc:       "Valid HostSNI rule not matching",
			rule:       "HostSNI(`example.org`)",
			serverName: "example.com",
			matchErr:   true,
		},
		{
			desc:       "Valid negative HostSNI rule not matching",
			rule:       "!HostSNI(`example.com`)",
			serverName: "example.com",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP rule matching",
			rule:       "HostSNI(`example.org`) && ClientIP(`10.0.0.1`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and ClientIP rule matching",
			rule:       "!HostSNI(`example.com`) && ClientIP(`10.0.0.1`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and negative ClientIP rule matching",
			rule:       "HostSNI(`example.org`) && !ClientIP(`10.0.0.2`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!HostSNI(`example.com`) && !ClientIP(`10.0.0.2`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI or negative ClientIP rule matching",
			rule:       "!(HostSNI(`example.com`) || ClientIP(`10.0.0.2`))",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`example.com`) && ClientIP(`10.0.0.2`))",
			serverName: "example.org",
			remoteAddr: "10.0.0.2:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`example.com`) && ClientIP(`10.0.0.2`))",
			serverName: "example.com",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`example.com`) && ClientIP(`10.0.0.2`))",
			serverName: "example.com",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`example.com`) && ClientIP(`10.0.0.2`))",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP rule not matching",
			rule:       "HostSNI(`example.org`) && ClientIP(`10.0.0.1`)",
			serverName: "example.com",
			remoteAddr: "10.0.0.1:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP rule not matching",
			rule:       "HostSNI(`example.org`) && ClientIP(`10.0.0.1`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI or ClientIP rule matching",
			rule:       "HostSNI(`example.org`) || ClientIP(`10.0.0.1`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI or ClientIP rule matching",
			rule:       "HostSNI(`example.org`) || ClientIP(`10.0.0.1`)",
			serverName: "example.com",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI or ClientIP rule matching",
			rule:       "HostSNI(`example.org`) || ClientIP(`10.0.0.1`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.2:80",
		},
		{
			desc:       "Valid HostSNI or ClientIP rule not matching",
			rule:       "HostSNI(`example.org`) || ClientIP(`10.0.0.1`)",
			serverName: "example.com",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI x 3 OR rule matching",
			rule:       "HostSNI(`example.org`) || HostSNI(`example.eu`) || HostSNI(`example.com`)",
			serverName: "example.org",
		},
		{
			desc:       "Valid HostSNI x 3 OR rule not matching",
			rule:       "HostSNI(`example.org`) || HostSNI(`example.eu`) || HostSNI(`example.com`)",
			serverName: "baz",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule matching",
			rule:       "HostSNI(`example.org`) || HostSNI(`example.com`) && ClientIP(`10.0.0.1`)",
			serverName: "example.org",
			remoteAddr: "10.0.0.2:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule matching",
			rule:       "HostSNI(`example.org`) || HostSNI(`example.com`) && ClientIP(`10.0.0.1`)",
			serverName: "example.com",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule not matching",
			rule:       "HostSNI(`example.org`) || HostSNI(`example.com`) && ClientIP(`10.0.0.1`)",
			serverName: "example.com",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP Combined rule not matching",
			rule:       "HostSNI(`example.org`) || HostSNI(`example.com`) && ClientIP(`10.0.0.1`)",
			serverName: "baz",
			remoteAddr: "10.0.0.1:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP complex combined rule matching",
			rule:       "(HostSNI(`example.org`) || HostSNI(`example.com`)) && (ClientIP(`10.0.0.1`) || ClientIP(`10.0.0.2`))",
			serverName: "example.com",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP complex combined rule not matching",
			rule:       "(HostSNI(`example.org`) || HostSNI(`example.com`)) && (ClientIP(`10.0.0.1`) || ClientIP(`10.0.0.2`))",
			serverName: "baz",
			remoteAddr: "10.0.0.1:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP complex combined rule not matching",
			rule:       "(HostSNI(`example.org`) || HostSNI(`example.com`)) && (ClientIP(`10.0.0.1`) || ClientIP(`10.0.0.2`))",
			serverName: "example.com",
			remoteAddr: "10.0.0.3:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP more complex (but absurd) combined rule matching",
			rule:       "(HostSNI(`example.org`) || (HostSNI(`example.com`) && !HostSNI(`example.org`))) && ((ClientIP(`10.0.0.1`) && !ClientIP(`10.0.0.2`)) || ClientIP(`10.0.0.2`)) ",
			serverName: "example.com",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule",
			rule:       "ALPN(`h2c`) && (!ALPN(`h2`) || HostSNI(`example.eu`))",
			protos:     []string{"h2c", "mqtt"},
			serverName: "example.eu",
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule not matching by SNI",
			rule:       "ALPN(`h2c`) && (!ALPN(`h2`) || HostSNI(`example.eu`))",
			protos:     []string{"h2c", "http/1.1", "h2"},
			serverName: "example.com",
			matchErr:   true,
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule matching by ALPN",
			rule:       "ALPN(`h2c`) && (!ALPN(`h2`) || HostSNI(`example.eu`))",
			protos:     []string{"h2c", "http/1.1"},
			serverName: "example.com",
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule not matching by protos",
			rule:       "ALPN(`h2c`) && (!ALPN(`h2`) || HostSNI(`example.eu`))",
			protos:     []string{"http/1.1", "mqtt"},
			serverName: "example.com",
			matchErr:   true,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			msg := "BYTES"
			handler := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
				_, err := conn.Write([]byte(msg))
				require.NoError(t, err)
			})

			router, err := NewMuxer()
			require.NoError(t, err)

			err = router.AddRoute(test.rule, 0, handler)
			if test.routeErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			addr := "0.0.0.0:0"
			if test.remoteAddr != "" {
				addr = test.remoteAddr
			}

			conn := &fakeConn{
				call:       map[string]int{},
				remoteAddr: fakeAddr{addr: addr},
			}

			connData, err := NewConnData(test.serverName, conn, test.protos)
			require.NoError(t, err)

			matchingHandler, _ := router.Match(connData)
			if test.matchErr {
				require.Nil(t, matchingHandler)
				return
			}

			require.NotNil(t, matchingHandler)

			matchingHandler.ServeTCP(conn)

			n, ok := conn.call[msg]
			assert.Equal(t, n, 1)
			assert.True(t, ok)
		})
	}
}

func TestParseHostSNI(t *testing.T) {
	testCases := []struct {
		desc          string
		expression    string
		domain        []string
		errorExpected bool
	}{
		{
			desc:          "Unknown rule",
			expression:    "Unknown(`example.com`)",
			errorExpected: true,
		},
		{
			desc:       "HostSNI rule",
			expression: "HostSNI(`example.com`)",
			domain:     []string{"example.com"},
		},
		{
			desc:       "HostSNI rule upper",
			expression: "HOSTSNI(`example.com`)",
			domain:     []string{"example.com"},
		},
		{
			desc:       "HostSNI rule lower",
			expression: "hostsni(`example.com`)",
			domain:     []string{"example.com"},
		},
		{
			desc:       "No hostSNI rule",
			expression: "ClientIP(`10.1`)",
		},
		{
			desc:       "HostSNI rule and another rule",
			expression: "HostSNI(`example.com`) && ClientIP(`10.1`)",
			domain:     []string{"example.com"},
		},
		{
			desc:       "HostSNI rule to lower and another rule",
			expression: "HostSNI(`example.com`) && ClientIP(`10.1`)",
			domain:     []string{"example.com"},
		},
		{
			desc:       "HostSNI rule with no domain",
			expression: "HostSNI() && ClientIP(`10.1`)",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
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

func Test_Priority(t *testing.T) {
	testCases := []struct {
		desc         string
		rules        map[string]int
		serverName   string
		expectedRule string
	}{
		{
			desc: "One matching rule, calculated priority",
			rules: map[string]int{
				"HostSNI(`example.com`)": 0,
				"HostSNI(`example.org`)": 0,
			},
			expectedRule: "HostSNI(`example.com`)",
			serverName:   "example.com",
		},
		{
			desc: "One matching rule, custom priority",
			rules: map[string]int{
				"HostSNI(`example.org`)": 0,
				"HostSNI(`example.com`)": 10000,
			},
			expectedRule: "HostSNI(`example.org`)",
			serverName:   "example.org",
		},
		{
			desc: "Two matching rules, calculated priority",
			rules: map[string]int{
				"HostSNI(`example.org`)": 0,
				"HostSNI(`example.com`)": 0,
			},
			expectedRule: "HostSNI(`example.org`)",
			serverName:   "example.org",
		},
		{
			desc: "Two matching rules, custom priority",
			rules: map[string]int{
				"HostSNI(`example.com`)": 10000,
				"HostSNI(`example.org`)": 0,
			},
			expectedRule: "HostSNI(`example.com`)",
			serverName:   "example.com",
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			matchedRule := ""
			for rule, priority := range test.rules {
				rule := rule
				err := muxer.AddRoute(rule, priority, tcp.HandlerFunc(func(conn tcp.WriteCloser) {
					matchedRule = rule
				}))
				require.NoError(t, err)
			}

			handler, _ := muxer.Match(ConnData{
				serverName: test.serverName,
			})
			require.NotNil(t, handler)

			handler.ServeTCP(nil)
			assert.Equal(t, test.expectedRule, matchedRule)
		})
	}
}

type fakeConn struct {
	call       map[string]int
	remoteAddr net.Addr
}

func (f *fakeConn) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (f *fakeConn) Write(b []byte) (n int, err error) {
	f.call[string(b)]++
	return len(b), nil
}

func (f *fakeConn) Close() error {
	panic("implement me")
}

func (f *fakeConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (f *fakeConn) RemoteAddr() net.Addr {
	return f.remoteAddr
}

func (f *fakeConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

func (f *fakeConn) CloseWrite() error {
	panic("implement me")
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
