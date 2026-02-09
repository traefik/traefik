package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/file"
)

func TestEntryPointProtocol(t *testing.T) {
	tests := []struct {
		name             string
		address          string
		expectedAddress  string
		expectedProtocol string
		expectedError    bool
	}{
		{
			name:             "Without protocol",
			address:          "127.0.0.1:8080",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "tcp",
			expectedError:    false,
		},
		{
			name:             "With TCP protocol in upper case",
			address:          "127.0.0.1:8080/TCP",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "tcp",
			expectedError:    false,
		},
		{
			name:             "With UDP protocol in upper case",
			address:          "127.0.0.1:8080/UDP",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "udp",
			expectedError:    false,
		},
		{
			name:             "With UDP protocol in weird case",
			address:          "127.0.0.1:8080/uDp",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "udp",
			expectedError:    false,
		},

		{
			name:          "With invalid protocol",
			address:       "127.0.0.1:8080/toto/tata",
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := EntryPoint{
				Address: tt.address,
			}
			protocol, err := ep.GetProtocol()
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedProtocol, protocol)
			require.Equal(t, tt.expectedAddress, ep.GetAddress())
		})
	}
}

func TestObservabilityConfigSetDefaults(t *testing.T) {
	testCases := []struct {
		desc            string
		mutate          func(*ObservabilityConfig)
		expectedAccess  bool
		expectedMetrics bool
		expectedTracing bool
	}{
		{
			desc:            "defaults",
			mutate:          func(*ObservabilityConfig) {},
			expectedAccess:  true,
			expectedMetrics: true,
			expectedTracing: true,
		},
		{
			desc:            "set tracing to false",
			mutate:          func(o *ObservabilityConfig) { *o.Tracing = false },
			expectedAccess:  true,
			expectedMetrics: true,
			expectedTracing: false,
		},
		{
			desc:            "set metrics to false",
			mutate:          func(o *ObservabilityConfig) { *o.Metrics = false },
			expectedAccess:  true,
			expectedMetrics: false,
			expectedTracing: true,
		},
		{
			desc:            "set access logs to false",
			mutate:          func(o *ObservabilityConfig) { *o.AccessLogs = false },
			expectedAccess:  false,
			expectedMetrics: true,
			expectedTracing: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var o ObservabilityConfig
			o.SetDefaults()

			require.NotSame(t, o.AccessLogs, o.Metrics)
			require.NotSame(t, o.Metrics, o.Tracing)

			test.mutate(&o)

			assert.Equal(t, test.expectedAccess, *o.AccessLogs)
			assert.Equal(t, test.expectedMetrics, *o.Metrics)
			assert.Equal(t, test.expectedTracing, *o.Tracing)
		})
	}
}

func TestObservabilityConfigDecode(t *testing.T) {
	testCases := []struct {
		desc            string
		yaml            string
		expectedAccess  bool
		expectedMetrics bool
		expectedTracing bool
	}{
		{
			desc: "all true",
			yaml: `
entryPoints:
  web:
    address: ":80"
    observability:
      accessLogs: true
      metrics: true
      tracing: true
`,
			expectedAccess:  true,
			expectedMetrics: true,
			expectedTracing: true,
		},
		{
			desc: "access logs true others false",
			yaml: `
entryPoints:
  web:
    address: ":80"
    observability:
      accessLogs: true
      metrics: false
      tracing: false
`,
			expectedAccess:  true,
			expectedMetrics: false,
			expectedTracing: false,
		},
		{
			desc: "metrics true others false",
			yaml: `
entryPoints:
  web:
    address: ":80"
    observability:
      accessLogs: false
      metrics: true
      tracing: false
`,
			expectedAccess:  false,
			expectedMetrics: true,
			expectedTracing: false,
		},
		{
			desc: "all false",
			yaml: `
entryPoints:
  web:
    address: ":80"
    observability:
      accessLogs: false
      metrics: false
      tracing: false
`,
			expectedAccess:  false,
			expectedMetrics: false,
			expectedTracing: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var cfg Configuration
			err := file.DecodeContent(test.yaml, ".yaml", &cfg)
			require.NoError(t, err)

			ep := cfg.EntryPoints["web"]
			require.NotNil(t, ep)
			require.NotNil(t, ep.Observability)

			assert.Equal(t, test.expectedAccess, *ep.Observability.AccessLogs)
			assert.Equal(t, test.expectedMetrics, *ep.Observability.Metrics)
			assert.Equal(t, test.expectedTracing, *ep.Observability.Tracing)
		})
	}
}
