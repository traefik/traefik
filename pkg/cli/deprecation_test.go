package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/traefik/v3/cmd"
)

// func TestDeprecationNotice(t *testing.T) {
// 	tests := []struct {
// 		desc         string
// 		config       rawConfiguration
// 		logLevel     zerolog.Level
// 		incompatible assert.BoolAssertionFunc
// 	}{
// 		{
// 			desc: "Docker provider swarmMode option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"docker": map[string]bool{
// 						"swarmMode": true,
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Docker provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"docker": map[string]interface{}{
// 						"tls": map[string]bool{
// 							"caOptional": true,
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Swarm provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"swarm": map[string]interface{}{
// 						"tls": map[string]bool{
// 							"caOptional": true,
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Consul provider namespace option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"consul": map[string]string{
// 						"namespace": "myNamespace",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Consul provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"consul": map[string]interface{}{
// 						"tls": map[string]bool{
// 							"caOptional": true,
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "ConsulCatalog provider namespace option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"consulCatalog": map[string]string{
// 						"namespace": "myNamespace",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "ConsulCatalog provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"consulCatalog": map[string]interface{}{
// 						"endpoint": map[string]interface{}{
// 							"tls": map[string]bool{
// 								"caOptional": true,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Nomad provider namespace option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"nomad": map[string]string{
// 						"namespace": "myNamespace",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Nomad provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"nomad": map[string]interface{}{
// 						"endpoint": map[string]interface{}{
// 							"tls": map[string]bool{
// 								"caOptional": true,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Marathon configuration is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"marathon": map[string]string{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Rancher configuration is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"rancher": map[string]string{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "ETCD provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"etcd": map[string]interface{}{
// 						"tls": map[string]bool{
// 							"caOptional": true,
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Redis provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"redis": map[string]interface{}{
// 						"tls": map[string]bool{
// 							"caOptional": true,
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "HTTP provider tls.CAOptional option is incompatible",
// 			config: map[string]interface{}{
// 				"providers": map[string]interface{}{
// 					"http": map[string]interface{}{
// 						"tls": map[string]bool{
// 							"caOptional": true,
// 						},
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Pilot configuration is incompatible",
// 			config: map[string]interface{}{
// 				"pilot": map[string]string{
// 					"foo": "bar",
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Experimental HTTP3 enablement configuration is incompatible",
// 			config: map[string]interface{}{
// 				"experimental": map[string]interface{}{
// 					"http3": true,
// 				},
// 			},
// 			logLevel:     zerolog.WarnLevel,
// 			incompatible: assert.False,
// 		},
// 		{
// 			desc: "Tracing SpanNameLimit option is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"spanNameLimit": 42,
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Tracing Jaeger configuration is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"jaeger": map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Tracing Zipkin configuration is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"zipkin": map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Tracing Datadog configuration is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"datadog": map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Tracing Instana configuration is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"instana": map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Tracing Haystack configuration is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"haystack": map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 		{
// 			desc: "Tracing Elastic configuration is incompatible",
// 			config: map[string]interface{}{
// 				"tracing": map[string]interface{}{
// 					"elastic": map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 			logLevel:     zerolog.ErrorLevel,
// 			incompatible: assert.True,
// 		},
// 	}
//
// 	for _, test := range tests {
// 		test := test
// 		t.Run(test.desc, func(t *testing.T) {
// 			t.Parallel()
//
// 			var gotLog bool
// 			var gotLevel zerolog.Level
// 			testHook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
// 				gotLog = true
// 				gotLevel = level
// 			})
//
// 			logger := log.With().Logger().Hook(testHook)
// 			test.incompatible(t, test.config.deprecationNotice(logger))
// 			assert.True(t, gotLog)
// 			assert.Equal(t, test.logLevel, gotLevel)
// 		})
// 	}
// }

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
			l := DeprecatedLoader{}

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
