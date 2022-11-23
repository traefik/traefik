package tcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

func Test_HostSNICatchAll(t *testing.T) {
	testCases := []struct {
		desc       string
		rule       string
		isCatchAll bool
	}{
		{
			desc: "HostSNI(`example.com`) is not catchAll",
			rule: "HostSNI(`example.com`)",
		},
		{
			desc:       "HostSNI(`*`) is catchAll",
			rule:       "HostSNI(`*`)",
			isCatchAll: true,
		},
		{
			desc:       "HostSNIRegexp(`^.*$`) is not catchAll",
			rule:       "HostSNIRegexp(`.*`)",
			isCatchAll: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, tcp.HandlerFunc(func(conn tcp.WriteCloser) {}))
			require.NoError(t, err)

			handler, catchAll := muxer.Match(ConnData{
				serverName: "example.com",
			})
			require.NotNil(t, handler)
			assert.Equal(t, test.isCatchAll, catchAll)
		})
	}
}

func Test_HostSNI(t *testing.T) {
	testCases := []struct {
		desc       string
		rule       string
		serverName string
		buildErr   bool
		match      bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNI matcher (empty host)",
			rule:     "HostSNI(``)",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNI matcher (too many parameters)",
			rule:     "HostSNI(`example.com`, `example.org`)",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNI matcher (globing sub domain)",
			rule:     "HostSNI(`*.com`)",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNI matcher (non ASCII host)",
			rule:     "HostSNI(`🦭.com`)",
			buildErr: true,
		},
		{
			desc:       "Valid HostSNI matcher - puny-coded emoji",
			rule:       "HostSNI(`xn--9t9h.com`)",
			serverName: "xn--9t9h.com",
			match:      true,
		},
		{
			desc:       "Valid HostSNI matcher - puny-coded emoji but emoji in server name",
			rule:       "HostSNI(`xn--9t9h.com`)",
			serverName: "🦭.com",
		},
		{
			desc:       "Matching hosts",
			rule:       "HostSNI(`example.com`)",
			serverName: "example.com",
			match:      true,
		},
		{
			desc:       "No matching hosts",
			rule:       "HostSNI(`example.com`)",
			serverName: "example.org",
		},
		{
			desc:       "Matching globing host `*`",
			rule:       "HostSNI(`*`)",
			serverName: "example.com",
			match:      true,
		},
		{
			desc:       "Matching globing host `*` and empty server name",
			rule:       "HostSNI(`*`)",
			serverName: "",
			match:      true,
		},
		{
			desc:       "Matching host with trailing dot",
			rule:       "HostSNI(`example.com.`)",
			serverName: "example.com.",
			match:      true,
		},
		{
			desc:       "Matching host with trailing dot but not in server name",
			rule:       "HostSNI(`example.com.`)",
			serverName: "example.com",
			match:      true,
		},
		{
			desc:       "Matching hosts with subdomains",
			rule:       "HostSNI(`foo.example.com`)",
			serverName: "foo.example.com",
			match:      true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, tcp.HandlerFunc(func(conn tcp.WriteCloser) {}))
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			meta := ConnData{
				serverName: test.serverName,
			}

			handler, _ := muxer.Match(meta)
			require.Equal(t, test.match, handler != nil)
		})
	}
}

func Test_HostSNIRegexp(t *testing.T) {
	testCases := []struct {
		desc     string
		rule     string
		expected map[string]bool
		buildErr bool
		match    bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNIRegexp matcher (empty host)",
			rule:     "HostSNIRegexp(``)",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNIRegexp matcher (non ASCII host)",
			rule:     "HostSNIRegexp(`🦭.com`)",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNIRegexp matcher (invalid regexp)",
			rule:     "HostSNIRegexp(`(example.com`)",
			buildErr: true,
		},
		{
			desc:     "Invalid HostSNIRegexp matcher (too many parameters)",
			rule:     "HostSNIRegexp(`example.com`, `example.org`)",
			buildErr: true,
		},
		{
			desc: "valid HostSNIRegexp matcher",
			rule: "HostSNIRegexp(`^example\\.(com|org)$`)",
			expected: map[string]bool{
				"example.com":  true,
				"example.com.": false,
				"EXAMPLE.com":  false,
				"example.org":  true,
				"exampleuorg":  false,
				"":             false,
			},
		},
		{
			desc: "valid HostSNIRegexp matcher with Traefik v2 syntax",
			rule: "HostSNIRegexp(`example.{tld:(com|org)}`)",
			expected: map[string]bool{
				"example.com":  false,
				"example.com.": false,
				"EXAMPLE.com":  false,
				"example.org":  false,
				"exampleuorg":  false,
				"":             false,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, tcp.HandlerFunc(func(conn tcp.WriteCloser) {}))
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for serverName, match := range test.expected {
				meta := ConnData{
					serverName: serverName,
				}

				handler, _ := muxer.Match(meta)
				assert.Equal(t, match, handler != nil, serverName)
			}
		})
	}
}

func Test_ClientIP(t *testing.T) {
	testCases := []struct {
		desc     string
		rule     string
		expected map[string]bool
		buildErr bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:     "Invalid ClientIP matcher (empty host)",
			rule:     "ClientIP(``)",
			buildErr: true,
		},
		{
			desc:     "Invalid ClientIP matcher (non ASCII host)",
			rule:     "ClientIP(`🦭/32`)",
			buildErr: true,
		},
		{
			desc:     "Invalid ClientIP matcher (too many parameters)",
			rule:     "ClientIP(`127.0.0.1`, `127.0.0.2`)",
			buildErr: true,
		},
		{
			desc: "valid ClientIP matcher",
			rule: "ClientIP(`20.20.20.20`)",
			expected: map[string]bool{
				"20.20.20.20": true,
				"10.10.10.10": false,
			},
		},
		{
			desc: "valid ClientIP matcher with CIDR",
			rule: "ClientIP(`20.20.20.20/24`)",
			expected: map[string]bool{
				"20.20.20.20": true,
				"20.20.20.40": true,
				"10.10.10.10": false,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, tcp.HandlerFunc(func(conn tcp.WriteCloser) {}))
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for remoteIP, match := range test.expected {
				meta := ConnData{
					remoteIP: remoteIP,
				}

				handler, _ := muxer.Match(meta)
				assert.Equal(t, match, handler != nil, remoteIP)
			}
		})
	}
}

func Test_ALPN(t *testing.T) {
	testCases := []struct {
		desc     string
		rule     string
		expected map[string]bool
		buildErr bool
	}{
		{
			desc:     "Empty",
			buildErr: true,
		},
		{
			desc:     "Invalid ALPN matcher (TLS proto)",
			rule:     "ALPN(`acme-tls/1`)",
			buildErr: true,
		},
		{
			desc:     "Invalid ALPN matcher (empty parameters)",
			rule:     "ALPN(``)",
			buildErr: true,
		},
		{
			desc:     "Invalid ALPN matcher (too many parameters)",
			rule:     "ALPN(`h2`, `mqtt`)",
			buildErr: true,
		},
		{
			desc: "Valid ALPN matcher",
			rule: "ALPN(`h2`)",
			expected: map[string]bool{
				"h2":   true,
				"mqtt": false,
				"":     false,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, tcp.HandlerFunc(func(conn tcp.WriteCloser) {}))
			if test.buildErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for proto, match := range test.expected {
				meta := ConnData{
					alpnProtos: []string{proto},
				}

				handler, _ := muxer.Match(meta)
				assert.Equal(t, match, handler != nil, proto)
			}
		})
	}
}
