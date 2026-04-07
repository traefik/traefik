package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/provider/acme"
)

func pointer[T any](v T) *T { return &v }

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
						SanitizePath:   pointer(true),
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
						SanitizePath:   pointer(true),
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
						SanitizePath:   pointer(true),
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
						SanitizePath:   pointer(true),
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.conf.SetEffectiveConfiguration()

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
