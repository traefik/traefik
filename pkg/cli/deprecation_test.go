package cli

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/cmd"
)

func boolP(v bool) *bool {
	return &v
}

func intP(v int) *int {
	return &v
}

func stringP(v string) *string {
	return &v
}

func TestDeprecationNotice(t *testing.T) {
	tests := []struct {
		desc   string
		config configuration
	}{
		{
			desc: "Docker provider swarmMode option is incompatible",
			config: configuration{
				Providers: &providers{
					Docker: &docker{
						SwarmMode: boolP(true),
					},
				},
			},
		},
		{
			desc: "Docker provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					Docker: &docker{
						TLS: &tls{
							CAOptional: boolP(true),
						},
					},
				},
			},
		},
		{
			desc: "Swarm provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					Swarm: &swarm{
						TLS: &tls{
							CAOptional: boolP(true),
						},
					},
				},
			},
		},
		{
			desc: "Consul provider namespace option is incompatible",
			config: configuration{
				Providers: &providers{
					Consul: &consul{
						Namespace: stringP("foobar"),
					},
				},
			},
		},
		{
			desc: "Consul provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					Consul: &consul{
						TLS: &tls{
							CAOptional: boolP(true),
						},
					},
				},
			},
		},
		{
			desc: "ConsulCatalog provider namespace option is incompatible",
			config: configuration{
				Providers: &providers{
					ConsulCatalog: &consulCatalog{
						Namespace: stringP("foobar"),
					},
				},
			},
		},
		{
			desc: "ConsulCatalog provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					ConsulCatalog: &consulCatalog{
						Endpoint: &endpointConfig{
							TLS: &tls{
								CAOptional: boolP(true),
							},
						},
					},
				},
			},
		},
		{
			desc: "Nomad provider namespace option is incompatible",
			config: configuration{
				Providers: &providers{
					Nomad: &nomad{
						Namespace: stringP("foobar"),
					},
				},
			},
		},
		{
			desc: "Nomad provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					Nomad: &nomad{
						Endpoint: &endpointConfig{
							TLS: &tls{
								CAOptional: boolP(true),
							},
						},
					},
				},
			},
		},
		{
			desc: "Marathon configuration is incompatible",
			config: configuration{
				Providers: &providers{
					Marathon: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "Rancher configuration is incompatible",
			config: configuration{
				Providers: &providers{
					Rancher: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "ETCD provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					ETCD: &etcd{
						TLS: &tls{
							CAOptional: boolP(true),
						},
					},
				},
			},
		},
		{
			desc: "Redis provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					Redis: &redis{
						TLS: &tls{
							CAOptional: boolP(true),
						},
					},
				},
			},
		},
		{
			desc: "HTTP provider tls.CAOptional option is incompatible",
			config: configuration{
				Providers: &providers{
					HTTP: &http{
						TLS: &tls{
							CAOptional: boolP(true),
						},
					},
				},
			},
		},
		{
			desc: "Pilot configuration is incompatible",
			config: configuration{
				Pilot: map[string]any{
					"foo": "bar",
				},
			},
		},
		{
			desc: "Experimental HTTP3 enablement configuration is incompatible",
			config: configuration{
				Experimental: &experimental{
					HTTP3: boolP(true),
				},
			},
		},
		{
			desc: "Tracing SpanNameLimit option is incompatible",
			config: configuration{
				Tracing: &tracing{
					SpanNameLimit: intP(42),
				},
			},
		},
		{
			desc: "Tracing Jaeger configuration is incompatible",
			config: configuration{
				Tracing: &tracing{
					Jaeger: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "Tracing Zipkin configuration is incompatible",
			config: configuration{
				Tracing: &tracing{
					Zipkin: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "Tracing Datadog configuration is incompatible",
			config: configuration{
				Tracing: &tracing{
					Datadog: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "Tracing Instana configuration is incompatible",
			config: configuration{
				Tracing: &tracing{
					Instana: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "Tracing Haystack configuration is incompatible",
			config: configuration{
				Tracing: &tracing{
					Haystack: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		{
			desc: "Tracing Elastic configuration is incompatible",
			config: configuration{
				Tracing: &tracing{
					Elastic: map[string]any{
						"foo": "bar",
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotLog bool
			var gotLevel zerolog.Level
			testHook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				gotLog = true
				gotLevel = level
			})

			logger := log.With().Logger().Hook(testHook)
			assert.True(t, test.config.deprecationNotice(logger))
			assert.True(t, gotLog)
			assert.Equal(t, zerolog.ErrorLevel, gotLevel)
		})
	}
}

func TestLoad(t *testing.T) {
	testCases := []struct {
		desc     string
		args     []string
		env      map[string]string
		expected bool
	}{
		{
			desc: "[FLAG] providers.marathon is deprecated",
			args: []string{
				"--access-log",
				"--log.level=DEBUG",
				"--entrypoints.test.http.tls",
				"--providers.marathon",
			},
			expected: true,
		},
		{
			desc: "[FLAG] multiple deprecated",
			args: []string{
				"--access-log",
				"--log.level=DEBUG",
				"--entrypoints.test.http.tls",
				"--providers.marathon",
				"--pilot.token=XXX",
			},
			expected: true,
		},
		{
			desc: "[FLAG] no deprecated",
			args: []string{
				"--access-log",
				"--log.level=DEBUG",
				"--entrypoints.test.http.tls",
			},
			expected: false,
		},
		{
			desc: "[ENV] providers.marathon is deprecated",
			env: map[string]string{
				"TRAEFIK_ACCESS_LOG":               "true",
				"TRAEFIK_LOG_LEVEL":                "DEBUG",
				"TRAEFIK_ENTRYPOINT_TEST_HTTP_TLS": "true",
				"TRAEFIK_PROVIDERS_MARATHON":       "true",
			},
			expected: true,
		},
		{
			desc: "[ENV] multiple deprecated",
			env: map[string]string{
				"TRAEFIK_ACCESS_LOG":               "true",
				"TRAEFIK_LOG_LEVEL":                "DEBUG",
				"TRAEFIK_ENTRYPOINT_TEST_HTTP_TLS": "true",
				"TRAEFIK_PROVIDERS_MARATHON":       "true",
				"TRAEFIK_PILOT_TOKEN":              "xxx",
			},
			expected: true,
		},
		{
			desc: "[ENV] no deprecated",
			env: map[string]string{
				"TRAEFIK_ACCESS_LOG":               "true",
				"TRAEFIK_LOG_LEVEL":                "DEBUG",
				"TRAEFIK_ENTRYPOINT_TEST_HTTP_TLS": "true",
			},

			expected: false,
		},
		{
			desc: "[FILE] providers.marathon is deprecated",
			args: []string{
				"--configfile=traefik_deprecated.toml",
			},
			expected: true,
		},
		{
			desc: "[FILE] multiple deprecated",
			args: []string{
				"--configfile=traefik_multiple_deprecated.toml",
			},
			expected: true,
		},
		{
			desc: "[FILE] no deprecated",
			args: []string{
				"--configfile=traefik_no_deprecated.toml",
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			tconfig := cmd.NewTraefikConfiguration()
			c := &cli.Command{Configuration: tconfig}
			l := DeprecationLoader{}

			for name, val := range test.env {
				t.Setenv(name, val)
			}
			deprecated, err := l.Load(test.args, c)

			assert.Equal(t, test.expected, deprecated)
			if !test.expected {
				require.NoError(t, err)
			}
		})
	}
}
