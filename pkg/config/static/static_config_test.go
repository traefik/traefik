package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				Providers: &Providers{},
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
						MaxConcurrentStreams: 250,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{},
			},
		},
		{
			desc: "ACME simple",
			conf: &Configuration{
				Providers: &Providers{},
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
						MaxConcurrentStreams: 250,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{},
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
				Providers: &Providers{},
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
						MaxConcurrentStreams: 250,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{},
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
				Providers: &Providers{},
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
						MaxConcurrentStreams: 250,
					},
					HTTP3: nil,
					UDP: &UDPConfig{
						Timeout: 3000000000,
					},
				}},
				Providers: &Providers{},
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
