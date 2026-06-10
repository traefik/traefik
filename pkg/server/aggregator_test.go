package server

import (
	"testing"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/tls"
)

func Test_mergeConfiguration(t *testing.T) {
	testCases := []struct {
		desc     string
		given    dynamic.Configurations
		expected *dynamic.HTTPConfiguration
	}{
		{
			desc:  "Nil returns an empty configuration",
			given: nil,
			expected: &dynamic.HTTPConfiguration{
				Routers:           make(map[string]*dynamic.Router),
				Middlewares:       make(map[string]*dynamic.Middleware),
				Services:          make(map[string]*dynamic.Service),
				Models:            make(map[string]*dynamic.Model),
				ServersTransports: make(map[string]*dynamic.ServersTransport),
			},
		},
		{
			desc: "Returns fully qualified elements from a mono-provider configuration map",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router-1": {},
						},
						Middlewares: map[string]*dynamic.Middleware{
							"middleware-1": {},
						},
						Services: map[string]*dynamic.Service{
							"service-1": {},
						},
					},
				},
			},
			expected: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"router-1@provider-1": {
						EntryPoints: []string{"defaultEP"},
					},
				},
				Middlewares: map[string]*dynamic.Middleware{
					"middleware-1@provider-1": {},
				},
				Services: map[string]*dynamic.Service{
					"service-1@provider-1": {},
				},
				Models:            make(map[string]*dynamic.Model),
				ServersTransports: make(map[string]*dynamic.ServersTransport),
			},
		},
		{
			desc: "Returns fully qualified elements from a multi-provider configuration map",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router-1": {},
						},
						Middlewares: map[string]*dynamic.Middleware{
							"middleware-1": {},
						},
						Services: map[string]*dynamic.Service{
							"service-1": {},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					HTTP: &dynamic.HTTPConfiguration{
						Routers: map[string]*dynamic.Router{
							"router-1": {},
						},
						Middlewares: map[string]*dynamic.Middleware{
							"middleware-1": {},
						},
						Services: map[string]*dynamic.Service{
							"service-1": {},
						},
					},
				},
			},
			expected: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"router-1@provider-1": {
						EntryPoints: []string{"defaultEP"},
					},
					"router-1@provider-2": {
						EntryPoints: []string{"defaultEP"},
					},
				},
				Middlewares: map[string]*dynamic.Middleware{
					"middleware-1@provider-1": {},
					"middleware-1@provider-2": {},
				},
				Services: map[string]*dynamic.Service{
					"service-1@provider-1": {},
					"service-1@provider-2": {},
				},
				Models:            make(map[string]*dynamic.Model),
				ServersTransports: make(map[string]*dynamic.ServersTransport),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeConfiguration(test.given, []string{"defaultEP"})
			assert.Equal(t, test.expected, actual.HTTP)
		})
	}
}

func Test_mergeConfiguration_tlsCertificates(t *testing.T) {
	testCases := []struct {
		desc     string
		given    dynamic.Configurations
		expected []*tls.CertAndStores
	}{
		{
			desc: "Skip temp certificates from another provider than tlsalpn",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{
							{Certificate: tls.Certificate{}, Stores: []string{tlsalpn01.ACMETLS1Protocol}},
						},
					},
				},
			},
			expected: nil,
		},
		{
			desc: "Allows tlsalpn provider to give certificates",
			given: dynamic.Configurations{
				"tlsalpn.acme": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Certificates: []*tls.CertAndStores{{
							Certificate: tls.Certificate{CertFile: "foo", KeyFile: "bar"},
							Stores:      []string{tlsalpn01.ACMETLS1Protocol},
						}},
					},
				},
			},
			expected: []*tls.CertAndStores{{
				Certificate: tls.Certificate{CertFile: "foo", KeyFile: "bar"},
				Stores:      []string{tlsalpn01.ACMETLS1Protocol},
			}},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeConfiguration(test.given, []string{"defaultEP"})
			assert.Equal(t, test.expected, actual.TLS.Certificates)
		})
	}
}

func Test_mergeConfiguration_tlsOptions(t *testing.T) {
	testCases := []struct {
		desc     string
		given    dynamic.Configurations
		expected map[string]tls.Options
	}{
		{
			desc:  "Nil returns an empty configuration",
			given: nil,
			expected: map[string]tls.Options{
				"default": tls.DefaultTLSOptions,
			},
		},
		{
			desc: "Returns fully qualified elements from a mono-provider configuration map",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS12",
							},
						},
					},
				},
			},
			expected: map[string]tls.Options{
				"default": tls.DefaultTLSOptions,
				"foo@provider-1": {
					MinVersion: "VersionTLS12",
				},
			},
		},
		{
			desc: "Returns fully qualified elements from a multi-provider configuration map",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS13",
							},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS12",
							},
						},
					},
				},
			},
			expected: map[string]tls.Options{
				"default": tls.DefaultTLSOptions,
				"foo@provider-1": {
					MinVersion: "VersionTLS13",
				},
				"foo@provider-2": {
					MinVersion: "VersionTLS12",
				},
			},
		},
		{
			desc: "Create a valid default tls option when appears only in one provider",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS13",
							},
							"default": {
								MinVersion: "VersionTLS11",
							},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS12",
							},
						},
					},
				},
			},
			expected: map[string]tls.Options{
				"default": {
					MinVersion: "VersionTLS11",
				},
				"foo@provider-1": {
					MinVersion: "VersionTLS13",
				},
				"foo@provider-2": {
					MinVersion: "VersionTLS12",
				},
			},
		},
		{
			desc: "No default tls option if it is defined in multiple providers",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS12",
							},
							"default": {
								MinVersion: "VersionTLS11",
							},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS13",
							},
							"default": {
								MinVersion: "VersionTLS12",
							},
						},
					},
				},
			},
			expected: map[string]tls.Options{
				"foo@provider-1": {
					MinVersion: "VersionTLS12",
				},
				"foo@provider-2": {
					MinVersion: "VersionTLS13",
				},
			},
		},
		{
			desc: "Create a default TLS Options configuration if none was provided",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS12",
							},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Options: map[string]tls.Options{
							"foo": {
								MinVersion: "VersionTLS13",
							},
						},
					},
				},
			},
			expected: map[string]tls.Options{
				"default": tls.DefaultTLSOptions,
				"foo@provider-1": {
					MinVersion: "VersionTLS12",
				},
				"foo@provider-2": {
					MinVersion: "VersionTLS13",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeConfiguration(test.given, []string{"defaultEP"})
			assert.Equal(t, test.expected, actual.TLS.Options)
		})
	}
}

func Test_mergeConfiguration_tlsStore(t *testing.T) {
	testCases := []struct {
		desc     string
		given    dynamic.Configurations
		expected map[string]tls.Store
	}{
		{
			desc: "Create a valid default tls store when appears only in one provider",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Stores: map[string]tls.Store{
							"default": {
								DefaultCertificate: &tls.Certificate{
									CertFile: "foo",
									KeyFile:  "bar",
								},
							},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Stores: map[string]tls.Store{
							"foo": {
								DefaultCertificate: &tls.Certificate{
									CertFile: "foo",
									KeyFile:  "bar",
								},
							},
						},
					},
				},
			},
			expected: map[string]tls.Store{
				"default": {
					DefaultCertificate: &tls.Certificate{
						CertFile: "foo",
						KeyFile:  "bar",
					},
				},
				"foo@provider-2": {
					DefaultCertificate: &tls.Certificate{
						CertFile: "foo",
						KeyFile:  "bar",
					},
				},
			},
		},
		{
			desc: "Don't default tls store when appears two times",
			given: dynamic.Configurations{
				"provider-1": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Stores: map[string]tls.Store{
							"default": {
								DefaultCertificate: &tls.Certificate{
									CertFile: "foo",
									KeyFile:  "bar",
								},
							},
						},
					},
				},
				"provider-2": &dynamic.Configuration{
					TLS: &dynamic.TLSConfiguration{
						Stores: map[string]tls.Store{
							"default": {
								DefaultCertificate: &tls.Certificate{
									CertFile: "foo",
									KeyFile:  "bar",
								},
							},
						},
					},
				},
			},
			expected: map[string]tls.Store{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeConfiguration(test.given, []string{"defaultEP"})
			assert.Equal(t, test.expected, actual.TLS.Stores)
		})
	}
}

func Test_mergeConfiguration_defaultTCPEntryPoint(t *testing.T) {
	given := dynamic.Configurations{
		"provider-1": &dynamic.Configuration{
			TCP: &dynamic.TCPConfiguration{
				Routers: map[string]*dynamic.TCPRouter{
					"router-1": {},
				},
				Services: map[string]*dynamic.TCPService{
					"service-1": {},
				},
			},
		},
	}

	expected := &dynamic.TCPConfiguration{
		Routers: map[string]*dynamic.TCPRouter{
			"router-1@provider-1": {
				EntryPoints: []string{"defaultEP"},
			},
		},
		Middlewares: map[string]*dynamic.TCPMiddleware{},
		Services: map[string]*dynamic.TCPService{
			"service-1@provider-1": {},
		},
	}

	actual := mergeConfiguration(given, []string{"defaultEP"})
	assert.Equal(t, expected, actual.TCP)
}

func Test_applyModel(t *testing.T) {
	testCases := []struct {
		desc     string
		input    dynamic.Configuration
		expected dynamic.Configuration
	}{
		{
			desc:     "empty configuration",
			input:    dynamic.Configuration{},
			expected: dynamic.Configuration{},
		},
		{
			desc: "without model",
			input: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     make(map[string]*dynamic.Router),
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models:      make(map[string]*dynamic.Model),
				},
			},
			expected: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     make(map[string]*dynamic.Router),
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models:      make(map[string]*dynamic.Model),
				},
			},
		},
		{
			desc: "with model, not used",
			input: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     make(map[string]*dynamic.Router),
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"ep@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
				},
			},
			expected: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     make(map[string]*dynamic.Router),
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"ep@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
				},
			},
		},
		{
			desc: "with model, one entry point",
			input: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"test": {
							EntryPoints: []string{"websecure"},
						},
					},
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"websecure@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
				},
			},
			expected: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"test": {
							EntryPoints: []string{"websecure"},
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"websecure@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
				},
			},
		},
		{
			desc: "with model, one entry point, and router with tls",
			input: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"test": {
							EntryPoints: []string{"websecure"},
							TLS:         &dynamic.RouterTLSConfig{CertResolver: "router"},
						},
					},
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"websecure@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{CertResolver: "ep"},
						},
					},
				},
			},
			expected: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"test": {
							EntryPoints: []string{"websecure"},
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{CertResolver: "router"},
						},
					},
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"websecure@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{CertResolver: "ep"},
						},
					},
				},
			},
		},
		{
			desc: "with model, two entry points",
			input: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"test": {
							EntryPoints: []string{"websecure", "web"},
						},
					},
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"websecure@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
				},
			},
			expected: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"test": {
							EntryPoints: []string{"web"},
						},
						"websecure-test": {
							EntryPoints: []string{"websecure"},
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
					Models: map[string]*dynamic.Model{
						"websecure@internal": {
							Middlewares: []string{"test"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := applyModel(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_resolveHTTPTLSOptions(t *testing.T) {
	testCases := []struct {
		desc              string
		routers           map[string]*dynamic.Router
		expected          map[string]string // router name -> ResolvedOptions
		unexpectedRouters []string
	}{
		{
			desc: "same host, different options, different entryPoints: no conflict",
			routers: map[string]*dynamic.Router{
				"router-a@file": {EntryPoints: []string{"ep-a"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsA"}},
				"router-b@file": {EntryPoints: []string{"ep-b"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsB"}},
			},
			expected: map[string]string{
				"router-a@file": "optsA@file",
				"router-b@file": "optsB@file",
			},
		},
		{
			desc: "same host, different options, same entryPoint: conflict falls back to default",
			routers: map[string]*dynamic.Router{
				"router-a@file": {EntryPoints: []string{"ep-a"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsA"}},
				"router-b@file": {EntryPoints: []string{"ep-a"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsB"}},
			},
			expected: map[string]string{
				"ep-a-conflicted-router-a@file": "default",
				"ep-a-conflicted-router-b@file": "default",
			},
			unexpectedRouters: []string{"router-a@file", "router-b@file"},
		},
		{
			desc: "same host, same options, same entryPoint: keeps the configured options",
			routers: map[string]*dynamic.Router{
				"router-a@file": {EntryPoints: []string{"ep-a"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsA"}},
				"router-b@file": {EntryPoints: []string{"ep-a"}, Rule: "Host(`example.com`) && PathPrefix(`/foo`)", TLS: &dynamic.RouterTLSConfig{Options: "optsA"}},
			},
			expected: map[string]string{
				"router-a@file": "optsA@file",
				"router-b@file": "optsA@file",
			},
		},
		{
			desc: "router spanning two entryPoints, conflict on one only: router is duplicated",
			routers: map[string]*dynamic.Router{
				"shared@file": {EntryPoints: []string{"ep-a", "ep-b"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsX"}},
				"other@file":  {EntryPoints: []string{"ep-a"}, Rule: "Host(`example.com`)", TLS: &dynamic.RouterTLSConfig{Options: "optsY"}},
			},
			expected: map[string]string{
				"ep-a-conflicted-shared@file": "default",    // conflicts with other@file on ep-a
				"shared@file":                 "optsX@file", // alone on ep-b
				"ep-a-conflicted-other@file":  "default",
			},
			unexpectedRouters: []string{"other@file"},
		},
		{
			desc: "no domain in rule, non-default options: forced to default and renamed",
			routers: map[string]*dynamic.Router{
				"router-a@file": {EntryPoints: []string{"ep-a"}, Rule: "PathPrefix(`/foo`)", TLS: &dynamic.RouterTLSConfig{Options: "optsA"}},
			},
			expected: map[string]string{
				"ep-a-conflicted-router-a@file": "default",
			},
			unexpectedRouters: []string{"router-a@file"},
		},
		{
			desc: "no domain in rule, default options: not conflicting, keeps its name",
			routers: map[string]*dynamic.Router{
				"router-a@file": {EntryPoints: []string{"ep-a"}, Rule: "PathPrefix(`/foo`)", TLS: &dynamic.RouterTLSConfig{}},
			},
			expected: map[string]string{
				"router-a@file": "default",
			},
			unexpectedRouters: []string{"ep-a-conflicted-router-a@file"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := resolveHTTPTLSOptions(test.routers)

			for name, want := range test.expected {
				rt, ok := got[name]

				require.True(t, ok, "router %q is missing", name)
				require.NotNil(t, rt.TLS, "router %q has no TLS config", name)
				assert.Equal(t, want, rt.TLS.ResolvedOptions, "router %q %v", name, rt.EntryPoints)
			}

			for _, name := range test.unexpectedRouters {
				_, ok := got[name]
				require.False(t, ok, "router %q is present", name)
			}
		})
	}
}
