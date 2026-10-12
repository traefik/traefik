package ingressnginx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNginxTime(t *testing.T) {
	testCases := []struct {
		value    string
		expected time.Duration
		wantErr  bool
	}{
		{value: "30", expected: 30 * time.Second},
		{value: "500ms", expected: 500 * time.Millisecond},
		{value: "30s", expected: 30 * time.Second},
		{value: "10m", expected: 10 * time.Minute},
		{value: "2h", expected: 2 * time.Hour},
		{value: "360d", expected: 360 * 24 * time.Hour},
		{value: "2w", expected: 14 * 24 * time.Hour},
		{value: "1h30m", expected: 90 * time.Minute},
		{value: "1h 30m", expected: 90 * time.Minute},
		{value: "1d2h3m4s5ms", expected: 26*time.Hour + 3*time.Minute + 4*time.Second + 5*time.Millisecond},
		{value: "1m30", expected: 90 * time.Second},
		{value: " 30s ", expected: 30 * time.Second},
		{value: "0", expected: 0},

		// NGINX rejects months and years for the grpc_*_timeout directives.
		{value: "1M", wantErr: true},
		{value: "1y", wantErr: true},
		// Units must go from the largest to the smallest, each at most once.
		{value: "30m1h", wantErr: true},
		{value: "1m1m", wantErr: true},
		{value: "1s1ms1s", wantErr: true},
		{value: "", wantErr: true},
		{value: "s", wantErr: true},
		{value: "10x", wantErr: true},
		{value: "-10s", wantErr: true},
		{value: "1.5s", wantErr: true},
		{value: "1 h", wantErr: true},
		{value: "99999999999999999999s", wantErr: true},
		{value: "9999999999w", wantErr: true},
	}

	for _, test := range testCases {
		t.Run(test.value, func(t *testing.T) {
			t.Parallel()

			got, err := parseNginxTime(test.value)
			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestParseGRPCTimeouts(t *testing.T) {
	testCases := []struct {
		desc     string
		wrapper  string
		snippet  string
		expected grpcTimeouts
	}{
		{
			desc:     "empty snippet",
			wrapper:  "server",
			expected: grpcTimeouts{},
		},
		{
			desc:    "server snippet",
			wrapper: "server",
			snippet: `
grpc_connect_timeout 5s;
grpc_read_timeout 360d;
grpc_send_timeout 1h30m;
`,
			expected: grpcTimeouts{
				connect: new(5 * time.Second),
				read:    new(360 * 24 * time.Hour),
				send:    new(90 * time.Minute),
			},
		},
		{
			desc:    "configuration snippet with other directives",
			wrapper: "location /",
			snippet: `
more_set_headers "X-Foo: bar";
grpc_read_timeout 10m;
add_header X-Bar baz;
`,
			expected: grpcTimeouts{read: new(10 * time.Minute)},
		},
		{
			desc:    "invalid value is ignored",
			wrapper: "location /",
			snippet: `
grpc_read_timeout 10parsecs;
grpc_send_timeout 20s;
`,
			expected: grpcTimeouts{send: new(20 * time.Second)},
		},
		{
			desc:     "directive with two values is ignored",
			wrapper:  "location /",
			snippet:  `grpc_read_timeout 10s 20s;`,
			expected: grpcTimeouts{},
		},
		{
			desc:    "directive in a nested block is ignored",
			wrapper: "server",
			snippet: `
location /foo {
	grpc_read_timeout 10m;
}
`,
			expected: grpcTimeouts{},
		},
		{
			desc:     "snippet that cannot be parsed",
			wrapper:  "server",
			snippet:  `grpc_read_timeout 10m`,
			expected: grpcTimeouts{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := parseGRPCTimeouts(t.Context(), test.wrapper, test.snippet)
			assert.Equal(t, test.expected, got)
		})
	}
}
