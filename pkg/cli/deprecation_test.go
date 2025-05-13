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

func ptr[T any](t T) *T {
	return &t
}

func TestDeprecationNotice(t *testing.T) {
	tests := []struct {
		desc           string
		config         configuration
		wantCompatible bool
	}{
		{
			desc: "Docker provider swarmMode option is incompatible",
			config: configuration{
				Providers: &providers{
					Docker: &docker{
						SwarmMode: ptr(true),
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
							CAOptional: ptr(true),
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
							CAOptional: ptr(true),
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
						Namespace: ptr("foobar"),
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
							CAOptional: ptr(true),
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
						Namespace: ptr("foobar"),
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
								CAOptional: ptr(true),
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
						Namespace: ptr("foobar"),
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
								CAOptional: ptr(true),
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
							CAOptional: ptr(true),
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
							CAOptional: ptr(true),
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
							CAOptional: ptr(true),
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
					HTTP3: ptr(true),
				},
			},
		},
		{
			desc: "Experimental KubernetesGateway enablement configuration is compatible",
			config: configuration{
				Experimental: &experimental{
					KubernetesGateway: ptr(true),
				},
			},
			wantCompatible: true,
		},
		{
			desc: "Tracing SpanNameLimit option is incompatible",
			config: configuration{
				Tracing: &tracing{
					SpanNameLimit: ptr(42),
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
		{
			desc: "Core DefaultRuleSyntax configuration is compatible",
			config: configuration{
				Core: &core{
					DefaultRuleSyntax: "foobar",
				},
			},
			wantCompatible: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotLog bool
			var gotLevel zerolog.Level
			testHook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				gotLog = true
				gotLevel = level
			})

			logger := log.With().Logger().Hook(testHook)

			assert.Equal(t, !test.wantCompatible, test.config.deprecationNotice(logger))
			assert.True(t, gotLog)
			assert.Equal(t, zerolog.ErrorLevel, gotLevel)
		})
	}
}

func TestLoad(t *testing.T) {
	testCases := []struct {
		desc           string
		args           []string
		env            map[string]string
		wantDeprecated bool
	}{
		{
			desc:           "Empty",
			args:           []string{},
			wantDeprecated: false,
		},
		{
			desc: "[FLAG] providers.marathon is deprecated",
			args: []string{
				"--access-log",
				"--log.level=DEBUG",
				"--entrypoints.test.http.tls",
				"--providers.nomad.endpoint.tls.insecureskipverify=true",
				"--providers.marathon",
			},
			wantDeprecated: true,
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
			wantDeprecated: true,
		},
		{
			desc: "[FLAG] no deprecated",
			args: []string{
				"--access-log",
				"--log.level=DEBUG",
				"--entrypoints.test.http.tls",
				"--providers.docker",
			},
			wantDeprecated: false,
		},
		{
			desc: "[ENV] providers.marathon is deprecated",
			env: map[string]string{
				"TRAEFIK_ACCESS_LOG":               "",
				"TRAEFIK_LOG_LEVEL":                "DEBUG",
				"TRAEFIK_ENTRYPOINT_TEST_HTTP_TLS": "true",
				"TRAEFIK_PROVIDERS_MARATHON":       "true",
			},
			wantDeprecated: true,
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
			wantDeprecated: true,
		},
		{
			desc: "[ENV] no deprecated",
			env: map[string]string{
				"TRAEFIK_ACCESS_LOG":               "true",
				"TRAEFIK_LOG_LEVEL":                "DEBUG",
				"TRAEFIK_ENTRYPOINT_TEST_HTTP_TLS": "true",
			},

			wantDeprecated: false,
		},
		{
			desc: "[FILE] providers.marathon is deprecated",
			args: []string{
				"--configfile=./fixtures/traefik_deprecated.toml",
			},
			wantDeprecated: true,
		},
		{
			desc: "[FILE] multiple deprecated",
			args: []string{
				"--configfile=./fixtures/traefik_multiple_deprecated.toml",
			},
			wantDeprecated: true,
		},
		{
			desc: "[FILE] no deprecated",
			args: []string{
				"--configfile=./fixtures/traefik_no_deprecated.toml",
			},
			wantDeprecated: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			tconfig := cmd.NewTraefikConfiguration()
			c := &cli.Command{Configuration: tconfig}
			l := DeprecationLoader{}

			for name, val := range test.env {
				t.Setenv(name, val)
			}
			deprecated, err := l.Load(test.args, c)
			assert.Equal(t, test.wantDeprecated, deprecated)
			if !test.wantDeprecated {
				require.NoError(t, err)
			}
		})
	}
}
