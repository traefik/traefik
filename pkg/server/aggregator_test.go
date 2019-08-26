package server

import (
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/tls"
	"github.com/stretchr/testify/assert"
)

func TestAggregator(t *testing.T) {
	testCases := []struct {
		desc     string
		given    dynamic.Configurations
		expected *dynamic.HTTPConfiguration
	}{
		{
			desc:  "Nil returns an empty configuration",
			given: nil,
			expected: &dynamic.HTTPConfiguration{
				Routers:     make(map[string]*dynamic.Router),
				Middlewares: make(map[string]*dynamic.Middleware),
				Services:    make(map[string]*dynamic.Service),
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
					"router-1@provider-1": {},
				},
				Middlewares: map[string]*dynamic.Middleware{
					"middleware-1@provider-1": {},
				},
				Services: map[string]*dynamic.Service{
					"service-1@provider-1": {},
				},
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
					"router-1@provider-1": {},
					"router-1@provider-2": {},
				},
				Middlewares: map[string]*dynamic.Middleware{
					"middleware-1@provider-1": {},
					"middleware-1@provider-2": {},
				},
				Services: map[string]*dynamic.Service{
					"service-1@provider-1": {},
					"service-1@provider-2": {},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeConfiguration(test.given)
			assert.Equal(t, test.expected, actual.HTTP)
		})
	}
}

func TestAggregator_tlsoptions(t *testing.T) {
	testCases := []struct {
		desc     string
		given    dynamic.Configurations
		expected map[string]tls.Options
	}{
		{
			desc:  "Nil returns an empty configuration",
			given: nil,
			expected: map[string]tls.Options{
				"default": {},
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
				"default": {},
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
				"default": {},
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
				"default": {},
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
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := mergeConfiguration(test.given)
			assert.Equal(t, test.expected, actual.TLS.Options)
		})
	}
}
