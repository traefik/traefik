package tcp

import (
	"fmt"
	"testing"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// All the tests in the suite are a copy of tcp muxer tests on branch v2.
// Only the test for route priority has not been copied here,
// because the priority computation is no longer done when calling the muxer AddRoute method.
func Test_addTCPRouteV2(t *testing.T) {
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
			desc:       "Valid negative HostSNI rule matching",
			rule:       "!HostSNI(`bar`)",
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
			desc:       "Empty HostSNIRegexp rule",
			rule:       "HostSNIRegexp()",
			serverName: "foobar",
			routeErr:   true,
		},
		{
			desc:       "Empty HostSNIRegexp rule",
			rule:       "HostSNIRegexp(``)",
			serverName: "foobar",
			routeErr:   true,
		},
		{
			desc:       "Valid HostSNIRegexp rule matching",
			rule:       "HostSNIRegexp(`{subdomain:[a-z]+}.foobar`)",
			serverName: "sub.foobar",
		},
		{
			desc:       "Valid negative HostSNIRegexp rule matching",
			rule:       "!HostSNIRegexp(`bar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNIRegexp rule matching with alternative case",
			rule:       "hostsniregexp(`foobar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNIRegexp rule matching with alternative case",
			rule:       "HOSTSNIREGEXP(`foobar`)",
			serverName: "foobar",
		},
		{
			desc:       "Valid HostSNIRegexp rule not matching",
			rule:       "HostSNIRegexp(`foobar`)",
			serverName: "bar",
			matchErr:   true,
		},
		{
			desc:       "Valid negative HostSNI rule not matching",
			rule:       "!HostSNI(`bar`)",
			serverName: "bar",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNIRegexp rule matching empty servername",
			rule:       "HostSNIRegexp(`{subdomain:[a-z]*}`)",
			serverName: "",
		},
		{
			desc:       "Valid HostSNIRegexp rule with one name",
			rule:       "HostSNIRegexp(`{dummy}`)",
			serverName: "toto",
		},
		{
			desc:       "Valid HostSNIRegexp rule with one name 2",
			rule:       "HostSNIRegexp(`{dummy}`)",
			serverName: "toto.com",
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
			desc:       "Invalid remoteAddr",
			rule:       "ClientIP(`10.0.0.1`)",
			remoteAddr: "not.an.IP:80",
			matchErr:   true,
		},
		{
			desc:       "Valid ClientIP rule matching",
			rule:       "ClientIP(`10.0.0.1`)",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative ClientIP rule matching",
			rule:       "!ClientIP(`20.0.0.1`)",
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
			desc:       "Valid negative ClientIP rule not matching",
			rule:       "!ClientIP(`10.0.0.2`)",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid ClientIP rule matching IPv6",
			rule:       "ClientIP(`10::10`)",
			remoteAddr: "[10::10]:80",
		},
		{
			desc:       "Valid negative ClientIP rule matching IPv6",
			rule:       "!ClientIP(`10::10`)",
			remoteAddr: "[::1]:80",
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
			desc:       "Valid negative HostSNI and ClientIP rule matching",
			rule:       "!HostSNI(`bar`) && ClientIP(`10.0.0.1`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and negative ClientIP rule matching",
			rule:       "HostSNI(`foobar`) && !ClientIP(`10.0.0.2`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!HostSNI(`bar`) && !ClientIP(`10.0.0.2`)",
			serverName: "foobar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI or negative ClientIP rule matching",
			rule:       "!(HostSNI(`bar`) || ClientIP(`10.0.0.2`))",
			serverName: "foobar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`bar`) && ClientIP(`10.0.0.2`))",
			serverName: "foobar",
			remoteAddr: "10.0.0.2:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`bar`) && ClientIP(`10.0.0.2`))",
			serverName: "bar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`bar`) && ClientIP(`10.0.0.2`))",
			serverName: "bar",
			remoteAddr: "10.0.0.2:80",
			matchErr:   true,
		},
		{
			desc:       "Valid negative HostSNI and negative ClientIP rule matching",
			rule:       "!(HostSNI(`bar`) && ClientIP(`10.0.0.2`))",
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
		{
			desc:       "Valid HostSNI and ClientIP complex combined rule matching",
			rule:       "(HostSNI(`foobar`) || HostSNI(`bar`)) && (ClientIP(`10.0.0.1`) || ClientIP(`10.0.0.2`))",
			serverName: "bar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:       "Valid HostSNI and ClientIP complex combined rule not matching",
			rule:       "(HostSNI(`foobar`) || HostSNI(`bar`)) && (ClientIP(`10.0.0.1`) || ClientIP(`10.0.0.2`))",
			serverName: "baz",
			remoteAddr: "10.0.0.1:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP complex combined rule not matching",
			rule:       "(HostSNI(`foobar`) || HostSNI(`bar`)) && (ClientIP(`10.0.0.1`) || ClientIP(`10.0.0.2`))",
			serverName: "bar",
			remoteAddr: "10.0.0.3:80",
			matchErr:   true,
		},
		{
			desc:       "Valid HostSNI and ClientIP more complex (but absurd) combined rule matching",
			rule:       "(HostSNI(`foobar`) || (HostSNI(`bar`) && !HostSNI(`foobar`))) && ((ClientIP(`10.0.0.1`) && !ClientIP(`10.0.0.2`)) || ClientIP(`10.0.0.2`)) ",
			serverName: "bar",
			remoteAddr: "10.0.0.1:80",
		},
		{
			desc:     "Invalid ALPN rule matching ACME-TLS/1",
			rule:     fmt.Sprintf("ALPN(`%s`)", tlsalpn01.ACMETLS1Protocol),
			protos:   []string{"foo"},
			routeErr: true,
		},
		{
			desc:   "Valid ALPN rule matching single protocol",
			rule:   "ALPN(`foo`)",
			protos: []string{"foo"},
		},
		{
			desc:     "Valid ALPN rule matching ACME-TLS/1 protocol",
			rule:     "ALPN(`foo`)",
			protos:   []string{tlsalpn01.ACMETLS1Protocol},
			matchErr: true,
		},
		{
			desc:     "Valid ALPN rule not matching single protocol",
			rule:     "ALPN(`foo`)",
			protos:   []string{"bar"},
			matchErr: true,
		},
		{
			desc:   "Valid alternative case ALPN rule matching single protocol without another being supported",
			rule:   "ALPN(`foo`) && !alpn(`h2`)",
			protos: []string{"foo", "bar"},
		},
		{
			desc:     "Valid alternative case ALPN rule not matching single protocol because of another being supported",
			rule:     "ALPN(`foo`) && !alpn(`h2`)",
			protos:   []string{"foo", "h2", "bar"},
			matchErr: true,
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule",
			rule:       "ALPN(`foo`) && (!alpn(`h2`) || hostsni(`foo`))",
			protos:     []string{"foo", "bar"},
			serverName: "foo",
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule not matching by SNI",
			rule:       "ALPN(`foo`) && (!alpn(`h2`) || hostsni(`foo`))",
			protos:     []string{"foo", "bar", "h2"},
			serverName: "bar",
			matchErr:   true,
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule matching by ALPN",
			rule:       "ALPN(`foo`) && (!alpn(`h2`) || hostsni(`foo`))",
			protos:     []string{"foo", "bar"},
			serverName: "bar",
		},
		{
			desc:       "Valid complex alternative case ALPN and HostSNI rule not matching by protos",
			rule:       "ALPN(`foo`) && (!alpn(`h2`) || hostsni(`foo`))",
			protos:     []string{"h2", "bar"},
			serverName: "bar",
			matchErr:   true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			msg := "BYTES"
			handler := tcp.HandlerFunc(func(conn tcp.WriteCloser) {
				_, err := conn.Write([]byte(msg))
				require.NoError(t, err)
			})

			router, err := NewMuxer()
			require.NoError(t, err)

			err = router.AddRoute(test.rule, "v2", 0, handler)
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
			assert.Equal(t, 1, n)
			assert.True(t, ok)
		})
	}
}

func TestParseHostSNIV2(t *testing.T) {
	testCases := []struct {
		description   string
		expression    string
		domain        []string
		errorExpected bool
	}{
		{
			description:   "Unknown rule",
			expression:    "Foobar(`foo.bar`,`test.bar`)",
			errorExpected: true,
		},
		{
			description: "Many hostSNI rules",
			expression:  "HostSNI(`foo.bar`,`test.bar`)",
			domain:      []string{"foo.bar", "test.bar"},
		},
		{
			description: "Many hostSNI rules upper",
			expression:  "HOSTSNI(`foo.bar`,`test.bar`)",
			domain:      []string{"foo.bar", "test.bar"},
		},
		{
			description: "Many hostSNI rules lower",
			expression:  "hostsni(`foo.bar`,`test.bar`)",
			domain:      []string{"foo.bar", "test.bar"},
		},
		{
			description: "No hostSNI rule",
			expression:  "ClientIP(`10.1`)",
		},
		{
			description: "HostSNI rule and another rule",
			expression:  "HostSNI(`foo.bar`) && ClientIP(`10.1`)",
			domain:      []string{"foo.bar"},
		},
		{
			description: "HostSNI rule to lower and another rule",
			expression:  "HostSNI(`Foo.Bar`) && ClientIP(`10.1`)",
			domain:      []string{"foo.bar"},
		},
		{
			description: "HostSNI rule with no domain",
			expression:  "HostSNI() && ClientIP(`10.1`)",
		},
	}

	for _, test := range testCases {
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			domains, err := ParseHostSNI(test.expression)

			if test.errorExpected {
				require.Errorf(t, err, "unable to parse correctly the domains in the HostSNI rule from %q", test.expression)
			} else {
				require.NoError(t, err, "%s: Error while parsing domain.", test.expression)
			}

			assert.Equal(t, test.domain, domains, "%s: Error parsing domains from expression.", test.expression)
		})
	}
}

func Test_HostSNICatchAllV2(t *testing.T) {
	testCases := []struct {
		desc       string
		rule       string
		isCatchAll bool
	}{
		{
			desc: "HostSNI(`foobar`) is not catchAll",
			rule: "HostSNI(`foobar`)",
		},
		{
			desc:       "HostSNI(`*`) is catchAll",
			rule:       "HostSNI(`*`)",
			isCatchAll: true,
		},
		{
			desc:       "HOSTSNI(`*`) is catchAll",
			rule:       "HOSTSNI(`*`)",
			isCatchAll: true,
		},
		{
			desc:       `HostSNI("*") is catchAll`,
			rule:       `HostSNI("*")`,
			isCatchAll: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, "v2", 0, tcp.HandlerFunc(func(conn tcp.WriteCloser) {}))
			require.NoError(t, err)

			handler, catchAll := muxer.Match(ConnData{
				serverName: "foobar",
			})
			require.NotNil(t, handler)
			assert.Equal(t, test.isCatchAll, catchAll)
		})
	}
}

func Test_HostSNIV2(t *testing.T) {
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
			desc:       "Matching globing host `*` and empty serverName",
			ruleHosts:  []string{"*"},
			serverName: "",
		},
		{
			desc:       "Matching globing host `*` and another non matching host",
			ruleHosts:  []string{"foo", "*"},
			serverName: "bar",
		},
		{
			desc:       "Matching globing host `*` and another non matching host, and empty servername",
			ruleHosts:  []string{"foo", "*"},
			serverName: "",
			matchErr:   true,
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
			ruleHosts:  []string{"foobar", "foo-bar.baz"},
			serverName: "foobar",
		},
		{
			desc:       "Matching hosts with subdomains",
			ruleHosts:  []string{"foo.bar"},
			serverName: "foo.bar",
		},
		{
			desc:       "Matching hosts with subdomains with _",
			ruleHosts:  []string{"foo_bar.example.com"},
			serverName: "foo_bar.example.com",
		},
		{
			desc:       "Matching IPv4",
			ruleHosts:  []string{"127.0.0.1"},
			serverName: "127.0.0.1",
		},
		{
			desc:       "Matching IPv6",
			ruleHosts:  []string{"10::10"},
			serverName: "10::10",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			matcherTree := &matchersTree{}
			err := hostSNIV2(matcherTree, test.ruleHosts...)
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			meta := ConnData{
				serverName: test.serverName,
			}

			assert.Equal(t, test.matchErr, !matcherTree.match(meta))
		})
	}
}

func Test_HostSNIRegexpV2(t *testing.T) {
	testCases := []struct {
		desc        string
		pattern     string
		serverNames map[string]bool
		buildErr    bool
	}{
		{
			desc:     "unbalanced braces",
			pattern:  "subdomain:(foo\\.)?bar\\.com}",
			buildErr: true,
		},
		{
			desc:     "empty group name",
			pattern:  "{:(foo\\.)?bar\\.com}",
			buildErr: true,
		},
		{
			desc:     "empty capturing group",
			pattern:  "{subdomain:}",
			buildErr: true,
		},
		{
			desc:     "malformed capturing group",
			pattern:  "{subdomain:(foo\\.?bar\\.com}",
			buildErr: true,
		},
		{
			desc:    "not interpreted as a regexp",
			pattern: "bar.com",
			serverNames: map[string]bool{
				"bar.com": true,
				"barucom": false,
			},
		},
		{
			desc:    "capturing group",
			pattern: "{subdomain:(foo\\.)?bar\\.com}",
			serverNames: map[string]bool{
				"foo.bar.com": true,
				"bar.com":     true,
				"fooubar.com": false,
				"barucom":     false,
				"barcom":      false,
			},
		},
		{
			desc:    "non capturing group",
			pattern: "{subdomain:(?:foo\\.)?bar\\.com}",
			serverNames: map[string]bool{
				"foo.bar.com": true,
				"bar.com":     true,
				"fooubar.com": false,
				"barucom":     false,
				"barcom":      false,
			},
		},
		{
			desc:    "regex insensitive",
			pattern: "{dummy:[A-Za-z-]+\\.bar\\.com}",
			serverNames: map[string]bool{
				"FOO.bar.com": true,
				"foo.bar.com": true,
				"fooubar.com": false,
				"barucom":     false,
				"barcom":      false,
			},
		},
		{
			desc:    "insensitive host",
			pattern: "{dummy:[a-z-]+\\.bar\\.com}",
			serverNames: map[string]bool{
				"FOO.bar.com": true,
				"foo.bar.com": true,
				"fooubar.com": false,
				"barucom":     false,
				"barcom":      false,
			},
		},
		{
			desc:    "insensitive host simple",
			pattern: "foo.bar.com",
			serverNames: map[string]bool{
				"FOO.bar.com": true,
				"foo.bar.com": true,
				"fooubar.com": false,
				"barucom":     false,
				"barcom":      false,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			matchersTree := &matchersTree{}
			err := hostSNIRegexpV2(matchersTree, test.pattern)
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for serverName, match := range test.serverNames {
				meta := ConnData{
					serverName: serverName,
				}

				assert.Equal(t, match, matchersTree.match(meta))
			}
		})
	}
}

func Test_ClientIPV2(t *testing.T) {
	testCases := []struct {
		desc      string
		ruleCIDRs []string
		remoteIP  string
		buildErr  bool
		matchErr  bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:      "Malformed CIDR",
			ruleCIDRs: []string{"héhé"},
			buildErr:  true,
		},
		{
			desc:      "Not matching empty remote IP",
			ruleCIDRs: []string{"20.20.20.20"},
			matchErr:  true,
		},
		{
			desc:      "Not matching IP",
			ruleCIDRs: []string{"20.20.20.20"},
			remoteIP:  "10.10.10.10",
			matchErr:  true,
		},
		{
			desc:      "Matching IP",
			ruleCIDRs: []string{"10.10.10.10"},
			remoteIP:  "10.10.10.10",
		},
		{
			desc:      "Not matching multiple IPs",
			ruleCIDRs: []string{"20.20.20.20", "30.30.30.30"},
			remoteIP:  "10.10.10.10",
			matchErr:  true,
		},
		{
			desc:      "Matching multiple IPs",
			ruleCIDRs: []string{"10.10.10.10", "20.20.20.20", "30.30.30.30"},
			remoteIP:  "20.20.20.20",
		},
		{
			desc:      "Not matching CIDR",
			ruleCIDRs: []string{"20.0.0.0/24"},
			remoteIP:  "10.10.10.10",
			matchErr:  true,
		},
		{
			desc:      "Matching CIDR",
			ruleCIDRs: []string{"20.0.0.0/8"},
			remoteIP:  "20.10.10.10",
		},
		{
			desc:      "Not matching multiple CIDRs",
			ruleCIDRs: []string{"10.0.0.0/24", "20.0.0.0/24"},
			remoteIP:  "10.10.10.10",
			matchErr:  true,
		},
		{
			desc:      "Matching multiple CIDRs",
			ruleCIDRs: []string{"10.0.0.0/8", "20.0.0.0/8"},
			remoteIP:  "20.10.10.10",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			matchersTree := &matchersTree{}
			err := clientIPV2(matchersTree, test.ruleCIDRs...)
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			meta := ConnData{
				remoteIP: test.remoteIP,
			}

			assert.Equal(t, test.matchErr, !matchersTree.match(meta))
		})
	}
}

func Test_ALPNV2(t *testing.T) {
	testCases := []struct {
		desc           string
		ruleALPNProtos []string
		connProto      string
		buildErr       bool
		matchErr       bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:           "ACME TLS proto",
			ruleALPNProtos: []string{tlsalpn01.ACMETLS1Protocol},
			buildErr:       true,
		},
		{
			desc:           "Not matching empty proto",
			ruleALPNProtos: []string{"h2"},
			matchErr:       true,
		},
		{
			desc:           "Not matching ALPN",
			ruleALPNProtos: []string{"h2"},
			connProto:      "mqtt",
			matchErr:       true,
		},
		{
			desc:           "Matching ALPN",
			ruleALPNProtos: []string{"h2"},
			connProto:      "h2",
		},
		{
			desc:           "Not matching multiple ALPNs",
			ruleALPNProtos: []string{"h2", "mqtt"},
			connProto:      "h2c",
			matchErr:       true,
		},
		{
			desc:           "Matching multiple ALPNs",
			ruleALPNProtos: []string{"h2", "h2c", "mqtt"},
			connProto:      "h2c",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			matchersTree := &matchersTree{}
			err := alpnV2(matchersTree, test.ruleALPNProtos...)
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			meta := ConnData{
				alpnProtos: []string{test.connProto},
			}

			assert.Equal(t, test.matchErr, !matchersTree.match(meta))
		})
	}
}
