package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/provider/acme"
	ingressnginx "github.com/traefik/traefik/v3/pkg/provider/kubernetes/ingress-nginx"
)

func TestHasEntrypoint(t *testing.T) {
	tests := []struct {
		desc        string
		entryPoints map[string]*EntryPoint
		assert      assert.BoolAssertionFunc
	}{
		{
			desc:   "no user defined entryPoints",
			assert: assert.False,
		},
		{
			desc: "user defined entryPoints",
			entryPoints: map[string]*EntryPoint{
				"foo": {},
			},
			assert: assert.True,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := &Configuration{
				EntryPoints: test.entryPoints,
			}

			test.assert(t, cfg.hasUserDefinedEntrypoint())
		})
	}
}

func TestConfiguration_SetEffectiveConfiguration(t *testing.T) {
	testCases := []struct {
		desc     string
		conf     *Configuration
		expected *Configuration
	}{
		{
			desc: "empty",
			conf: &Configuration{
				Providers: &Providers{Precedence: providerNames},
			},
			expected: &Configuration{
				EntryPoints: EntryPoints{"http": &EntryPoint{
					Address:         ":80",
					AllowACMEByPass: false,
					ReusePort:       false,
					AsDefault:       false,
					Transport: &EntryPointsTransport{
						LifeCycle: &LifeCycle{
							GraceTimeOut: 10000000000,
						},
						RespondingTimeouts: &RespondingTimeouts{
							ReadTimeout: 60000000000,
							IdleTimeout: 180000000000,
						},
					},
					ProxyProtocol:    nil,
					ForwardedHeaders: &ForwardedHeaders{},
					HTTP: HTTPConfig{
						SanitizePath:   new(true),
						MaxHeaderBytes: 1048576,
					},
					HTTP2: &HTTP2Config{
						MaxConcurrentStreams:      250,
						MaxDecoderHeaderTableSize: 4096,
						MaxEncoderHeaderTableSize: 4096,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{Precedence: providerNames},
			},
		},
		{
			desc: "ACME simple",
			conf: &Configuration{
				Providers: &Providers{Precedence: providerNames},
				CertificatesResolvers: map[string]CertificateResolver{
					"foo": {
						ACME: &acme.Configuration{
							DNSChallenge: &acme.DNSChallenge{
								Provider: "bar",
							},
						},
					},
				},
			},
			expected: &Configuration{
				EntryPoints: EntryPoints{"http": &EntryPoint{
					Address:         ":80",
					AllowACMEByPass: false,
					ReusePort:       false,
					AsDefault:       false,
					Transport: &EntryPointsTransport{
						LifeCycle: &LifeCycle{
							GraceTimeOut: 10000000000,
						},
						RespondingTimeouts: &RespondingTimeouts{
							ReadTimeout: 60000000000,
							IdleTimeout: 180000000000,
						},
					},
					ProxyProtocol:    nil,
					ForwardedHeaders: &ForwardedHeaders{},
					HTTP: HTTPConfig{
						SanitizePath:   new(true),
						MaxHeaderBytes: 1048576,
					},
					HTTP2: &HTTP2Config{
						MaxConcurrentStreams:      250,
						MaxDecoderHeaderTableSize: 4096,
						MaxEncoderHeaderTableSize: 4096,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{Precedence: providerNames},
				CertificatesResolvers: map[string]CertificateResolver{
					"foo": {
						ACME: &acme.Configuration{
							CAServer: "https://acme-v02.api.letsencrypt.org/directory",
							DNSChallenge: &acme.DNSChallenge{
								Provider: "bar",
							},
						},
					},
				},
			},
		},
		{
			desc: "ACME deprecation DelayBeforeCheck",
			conf: &Configuration{
				Providers: &Providers{Precedence: providerNames},
				CertificatesResolvers: map[string]CertificateResolver{
					"foo": {
						ACME: &acme.Configuration{
							DNSChallenge: &acme.DNSChallenge{
								Provider:         "bar",
								DelayBeforeCheck: 123,
							},
						},
					},
				},
			},
			expected: &Configuration{
				EntryPoints: EntryPoints{"http": &EntryPoint{
					Address:         ":80",
					AllowACMEByPass: false,
					ReusePort:       false,
					AsDefault:       false,
					Transport: &EntryPointsTransport{
						LifeCycle: &LifeCycle{
							GraceTimeOut: 10000000000,
						},
						RespondingTimeouts: &RespondingTimeouts{
							ReadTimeout: 60000000000,
							IdleTimeout: 180000000000,
						},
					},
					ProxyProtocol:    nil,
					ForwardedHeaders: &ForwardedHeaders{},
					HTTP: HTTPConfig{
						SanitizePath:   new(true),
						MaxHeaderBytes: 1048576,
					},
					HTTP2: &HTTP2Config{
						MaxConcurrentStreams:      250,
						MaxDecoderHeaderTableSize: 4096,
						MaxEncoderHeaderTableSize: 4096,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{Precedence: providerNames},
				CertificatesResolvers: map[string]CertificateResolver{
					"foo": {
						ACME: &acme.Configuration{
							CAServer: "https://acme-v02.api.letsencrypt.org/directory",
							DNSChallenge: &acme.DNSChallenge{
								Provider:         "bar",
								DelayBeforeCheck: 123,
								Propagation: &acme.Propagation{
									DelayBeforeChecks: 123,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "ACME deprecation DisablePropagationCheck",
			conf: &Configuration{
				Providers: &Providers{Precedence: providerNames},
				CertificatesResolvers: map[string]CertificateResolver{
					"foo": {
						ACME: &acme.Configuration{
							DNSChallenge: &acme.DNSChallenge{
								Provider:                "bar",
								DisablePropagationCheck: true,
							},
						},
					},
				},
			},
			expected: &Configuration{
				EntryPoints: EntryPoints{"http": &EntryPoint{
					Address:         ":80",
					AllowACMEByPass: false,
					ReusePort:       false,
					AsDefault:       false,
					Transport: &EntryPointsTransport{
						LifeCycle: &LifeCycle{
							GraceTimeOut: 10000000000,
						},
						RespondingTimeouts: &RespondingTimeouts{
							ReadTimeout: 60000000000,
							IdleTimeout: 180000000000,
						},
					},
					ProxyProtocol:    nil,
					ForwardedHeaders: &ForwardedHeaders{},
					HTTP: HTTPConfig{
						SanitizePath:   new(true),
						MaxHeaderBytes: 1048576,
					},
					HTTP2: &HTTP2Config{
						MaxConcurrentStreams:      250,
						MaxDecoderHeaderTableSize: 4096,
						MaxEncoderHeaderTableSize: 4096,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{Precedence: providerNames},
				CertificatesResolvers: map[string]CertificateResolver{
					"foo": {
						ACME: &acme.Configuration{
							CAServer: "https://acme-v02.api.letsencrypt.org/directory",
							DNSChallenge: &acme.DNSChallenge{
								Provider:                "bar",
								DisablePropagationCheck: true,
								Propagation: &acme.Propagation{
									DisableChecks: true,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress NGINX provider, no asDefault, all non-TLS non-internal entrypoints included",
			conf: &Configuration{
				Providers: &Providers{
					KubernetesIngressNGINX: &ingressnginx.Provider{},
				},
				EntryPoints: EntryPoints{
					"web":       {Address: ":80"},
					"admin":     {Address: ":8081"},
					"traefik":   {Address: ":8080"},
					"websecure": {Address: ":443", HTTP: HTTPConfig{TLS: &TLSConfig{}}},
				},
			},
			expected: &Configuration{
				Providers: &Providers{
					KubernetesIngressNGINX: &ingressnginx.Provider{
						NonTLSEntryPoints: []string{"admin", "web"},
					},
				},
				EntryPoints: EntryPoints{
					"web":       {Address: ":80"},
					"admin":     {Address: ":8081"},
					"traefik":   {Address: ":8080"},
					"websecure": {Address: ":443", HTTP: HTTPConfig{TLS: &TLSConfig{}}},
				},
			},
		},
		{
			desc: "Ingress NGINX provider, asDefault set, only marked entrypoint included",
			conf: &Configuration{
				Providers: &Providers{
					KubernetesIngressNGINX: &ingressnginx.Provider{},
				},
				EntryPoints: EntryPoints{
					"web":       {Address: ":80", AsDefault: true},
					"admin":     {Address: ":8081"},
					"traefik":   {Address: ":8080"},
					"websecure": {Address: ":443", HTTP: HTTPConfig{TLS: &TLSConfig{}}},
				},
			},
			expected: &Configuration{
				Providers: &Providers{
					KubernetesIngressNGINX: &ingressnginx.Provider{
						NonTLSEntryPoints: []string{"web"},
					},
				},
				EntryPoints: EntryPoints{
					"web":       {Address: ":80", AsDefault: true},
					"admin":     {Address: ":8081"},
					"traefik":   {Address: ":8080"},
					"websecure": {Address: ":443", HTTP: HTTPConfig{TLS: &TLSConfig{}}},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.conf.SetEffectiveConfiguration()

			// NonTLSEntryPoints is built from a map iteration, so its order isn't deterministic.
			if p := test.conf.Providers.KubernetesIngressNGINX; p != nil {
				assert.ElementsMatch(t, test.expected.Providers.KubernetesIngressNGINX.NonTLSEntryPoints, p.NonTLSEntryPoints)
				p.NonTLSEntryPoints = nil
				test.expected.Providers.KubernetesIngressNGINX.NonTLSEntryPoints = nil
			}

			assert.Equal(t, test.expected, test.conf)
		})
	}
}

func TestValidateConfiguration_BasePath(t *testing.T) {
	tests := []struct {
		desc      string
		basePath  string
		expectErr bool
	}{
		{
			desc:      "valid simple path",
			basePath:  "/api",
			expectErr: false,
		},
		{
			desc:      "valid path with segments",
			basePath:  "/my/base/path",
			expectErr: false,
		},
		{
			desc:      "valid path with allowed special chars",
			basePath:  "/valid/path-123",
			expectErr: false,
		},
		{
			desc:      "relative path",
			basePath:  "api/path",
			expectErr: true,
		},
		{
			desc:      "XSS payload",
			basePath:  `/api/"></script><script>alert("XSS")</script>`,
			expectErr: true,
		},
		{
			desc:      "path with spaces",
			basePath:  "/path with spaces",
			expectErr: true,
		},
		{
			desc:      "path with angle brackets",
			basePath:  "/path/<evil>",
			expectErr: true,
		},
		{
			desc:      "path with query string",
			basePath:  "/api?foo=bar",
			expectErr: true,
		},
		{
			desc:      "path with fragment",
			basePath:  "/api#section",
			expectErr: true,
		},
		{
			desc:      "valid root path",
			basePath:  "/",
			expectErr: false,
		},
		{
			desc:      "path with quote",
			basePath:  "/api/'onclick=alert(1)",
			expectErr: true,
		},
		{
			desc:      "path with encoded character",
			basePath:  "/api%2Ftoto",
			expectErr: true,
		},
		{
			desc:      "valid path with colons",
			basePath:  "/k8s/clusters/c-abcd0/api/v1/namespaces/my-ns/services/http:traefik:8080/proxy",
			expectErr: false,
		},
		{
			desc:      "valid path with tilde",
			basePath:  "/~user/dashboard",
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := &Configuration{
				API: &API{BasePath: test.basePath},
			}

			err := cfg.ValidateConfiguration()
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvidersPrecedence(t *testing.T) {
	testCases := []struct {
		desc          string
		cfg           *Configuration
		expectedError bool
		expected      []string
	}{
		{
			desc: "No precedence",
			cfg: &Configuration{
				Providers: &Providers{
					Precedence: providerNames,
				},
			},
			expected: providerNames,
		},
		{
			desc: "Precedence with non existing provider",
			cfg: &Configuration{
				Providers: &Providers{
					Precedence: []string{"unknown"},
				},
			},
			expectedError: true,
		},
		{
			desc: "Precedence with upper case provider",
			cfg: &Configuration{
				Providers: &Providers{
					Precedence: []string{"DOCKER"},
				},
			},
			expected: []string{"docker"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.cfg.SetEffectiveConfiguration()
			err := test.cfg.ValidateConfiguration()
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, test.cfg.Providers.Precedence)
			}
		})
	}
}

func TestValidateConfiguration_aliasHeadersStrategy(t *testing.T) {
	testCases := []struct {
		desc        string
		underscore  string
		alias       string
		expectError bool
	}{
		{
			desc: "no strategy configured",
		},
		{
			desc:  "only the new option configured",
			alias: AliasHeadersStrategyDelete,
		},
		{
			desc:       "only the deprecated option configured",
			underscore: AliasHeadersStrategyDelete,
		},
		{
			desc:       "only the deprecated option configured, set to keep",
			underscore: AliasHeadersStrategyKeep,
		},
		{
			desc:       "both options configured with the same value",
			underscore: AliasHeadersStrategyDelete,
			alias:      AliasHeadersStrategyDelete,
		},
		{
			desc:        "both options configured with different values",
			underscore:  AliasHeadersStrategyDelete,
			alias:       AliasHeadersStrategyReject,
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := &Configuration{
				Providers: &Providers{},
				EntryPoints: EntryPoints{
					"web": &EntryPoint{
						Address: ":80",
						HTTP: HTTPConfig{
							AliasHeadersStrategy:      test.alias,
							UnderscoreHeadersStrategy: test.underscore,
						},
					},
				},
			}

			err := cfg.ValidateConfiguration()
			if test.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
