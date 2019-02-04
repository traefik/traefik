package configuration

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/middlewares/tracing/jaeger"
	"github.com/containous/traefik/middlewares/tracing/zipkin"
	"github.com/containous/traefik/provider"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/provider/file"
	"github.com/stretchr/testify/assert"
)

const defaultConfigFile = "traefik.toml"

func TestSetEffectiveConfigurationGraceTimeout(t *testing.T) {
	testCases := []struct {
		desc                  string
		legacyGraceTimeout    time.Duration
		lifeCycleGraceTimeout time.Duration
		wantGraceTimeout      time.Duration
	}{
		{
			desc:               "legacy grace timeout given only",
			legacyGraceTimeout: 5 * time.Second,
			wantGraceTimeout:   5 * time.Second,
		},
		{
			desc:                  "legacy and life cycle grace timeouts given",
			legacyGraceTimeout:    5 * time.Second,
			lifeCycleGraceTimeout: 12 * time.Second,
			wantGraceTimeout:      5 * time.Second,
		},
		{
			desc:                  "legacy grace timeout omitted",
			legacyGraceTimeout:    0,
			lifeCycleGraceTimeout: 12 * time.Second,
			wantGraceTimeout:      12 * time.Second,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gc := &GlobalConfiguration{
				GraceTimeOut: flaeg.Duration(test.legacyGraceTimeout),
			}
			if test.lifeCycleGraceTimeout > 0 {
				gc.LifeCycle = &LifeCycle{
					GraceTimeOut: flaeg.Duration(test.lifeCycleGraceTimeout),
				}
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			assert.Equal(t, test.wantGraceTimeout, time.Duration(gc.LifeCycle.GraceTimeOut))
		})
	}
}

func TestSetEffectiveConfigurationFileProviderFilename(t *testing.T) {
	testCases := []struct {
		desc                        string
		fileProvider                *file.Provider
		wantFileProviderFilename    string
		wantFileProviderTraefikFile string
	}{
		{
			desc:                        "no filename for file provider given",
			fileProvider:                &file.Provider{},
			wantFileProviderFilename:    "",
			wantFileProviderTraefikFile: defaultConfigFile,
		},
		{
			desc:                        "filename for file provider given",
			fileProvider:                &file.Provider{BaseProvider: provider.BaseProvider{Filename: "other.toml"}},
			wantFileProviderFilename:    "other.toml",
			wantFileProviderTraefikFile: defaultConfigFile,
		},
		{
			desc:                        "directory for file provider given",
			fileProvider:                &file.Provider{Directory: "/"},
			wantFileProviderFilename:    "",
			wantFileProviderTraefikFile: defaultConfigFile,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gc := &GlobalConfiguration{
				File: test.fileProvider,
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			assert.Equal(t, test.wantFileProviderFilename, gc.File.Filename)
			assert.Equal(t, test.wantFileProviderTraefikFile, gc.File.TraefikFile)
		})
	}
}

func TestSetEffectiveConfigurationTracing(t *testing.T) {
	testCases := []struct {
		desc     string
		tracing  *tracing.Tracing
		expected *tracing.Tracing
	}{
		{
			desc:     "no tracing configuration",
			tracing:  &tracing.Tracing{},
			expected: &tracing.Tracing{},
		},
		{
			desc: "tracing bad backend name",
			tracing: &tracing.Tracing{
				Backend: "powpow",
			},
			expected: &tracing.Tracing{
				Backend: "powpow",
			},
		},
		{
			desc: "tracing jaeger backend name",
			tracing: &tracing.Tracing{
				Backend: "jaeger",
				Zipkin: &zipkin.Config{
					HTTPEndpoint: "http://localhost:9411/api/v1/spans",
					SameSpan:     false,
					ID128Bit:     true,
					Debug:        false,
				},
			},
			expected: &tracing.Tracing{
				Backend: "jaeger",
				Jaeger: &jaeger.Config{
					SamplingServerURL:      "http://localhost:5778/sampling",
					SamplingType:           "const",
					SamplingParam:          1.0,
					LocalAgentHostPort:     "127.0.0.1:6831",
					TraceContextHeaderName: "uber-trace-id",
				},
				Zipkin: nil,
			},
		},
		{
			desc: "tracing zipkin backend name",
			tracing: &tracing.Tracing{
				Backend: "zipkin",
				Jaeger: &jaeger.Config{
					SamplingServerURL:      "http://localhost:5778/sampling",
					SamplingType:           "const",
					SamplingParam:          1.0,
					LocalAgentHostPort:     "127.0.0.1:6831",
					TraceContextHeaderName: "uber-trace-id",
				},
			},
			expected: &tracing.Tracing{
				Backend: "zipkin",
				Jaeger:  nil,
				Zipkin: &zipkin.Config{
					HTTPEndpoint: "http://localhost:9411/api/v1/spans",
					SameSpan:     false,
					ID128Bit:     true,
					Debug:        false,
				},
			},
		},
		{
			desc: "tracing zipkin backend name value override",
			tracing: &tracing.Tracing{
				Backend: "zipkin",
				Jaeger: &jaeger.Config{
					SamplingServerURL:      "http://localhost:5778/sampling",
					SamplingType:           "const",
					SamplingParam:          1.0,
					LocalAgentHostPort:     "127.0.0.1:6831",
					TraceContextHeaderName: "uber-trace-id",
				},
				Zipkin: &zipkin.Config{
					HTTPEndpoint: "http://powpow:9411/api/v1/spans",
					SameSpan:     true,
					ID128Bit:     true,
					Debug:        true,
				},
			},
			expected: &tracing.Tracing{
				Backend: "zipkin",
				Jaeger:  nil,
				Zipkin: &zipkin.Config{
					HTTPEndpoint: "http://powpow:9411/api/v1/spans",
					SameSpan:     true,
					ID128Bit:     true,
					Debug:        true,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gc := &GlobalConfiguration{
				Tracing: test.tracing,
			}

			gc.SetEffectiveConfiguration(defaultConfigFile)

			assert.Equal(t, test.expected, gc.Tracing)
		})
	}
}

func TestInitACMEProvider(t *testing.T) {
	testCases := []struct {
		desc                  string
		acmeConfiguration     *acme.ACME
		expectedConfiguration *acmeprovider.Provider
		noError               bool
	}{
		{
			desc:                  "No ACME configuration",
			acmeConfiguration:     nil,
			expectedConfiguration: nil,
			noError:               true,
		},
		{
			desc:                  "ACME configuration with storage",
			acmeConfiguration:     &acme.ACME{Storage: "foo/acme.json"},
			expectedConfiguration: &acmeprovider.Provider{Configuration: &acmeprovider.Configuration{Storage: "foo/acme.json"}},
			noError:               true,
		},
		{
			desc:                  "ACME configuration with no storage",
			acmeConfiguration:     &acme.ACME{},
			expectedConfiguration: nil,
			noError:               false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gc := &GlobalConfiguration{
				ACME: test.acmeConfiguration,
			}

			configuration, err := gc.InitACMEProvider()

			assert.True(t, (err == nil) == test.noError)

			if test.expectedConfiguration == nil {
				assert.Nil(t, configuration)
			} else {
				assert.Equal(t, test.expectedConfiguration.Storage, configuration.Storage)
			}
		})
	}
}
