package server

import (
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAggregator(t *testing.T) {
	testCases := []struct {
		desc     string
		given    config.Configurations
		expected *config.HTTPConfiguration
	}{
		{
			desc:  "Nil returns an empty configuration",
			given: nil,
			expected: &config.HTTPConfiguration{
				Routers:     make(map[string]*config.Router),
				Middlewares: make(map[string]*config.Middleware),
				Services:    make(map[string]*config.Service),
			},
		},
		{
			desc: "Returns fully qualified elements from a mono-provider configuration map",
			given: config.Configurations{
				"provider-1": &config.Configuration{
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"router-1": {},
						},
						Middlewares: map[string]*config.Middleware{
							"middleware-1": {},
						},
						Services: map[string]*config.Service{
							"service-1": {},
						},
					},
				},
			},
			expected: &config.HTTPConfiguration{
				Routers: map[string]*config.Router{
					"provider-1.router-1": {},
				},
				Middlewares: map[string]*config.Middleware{
					"provider-1.middleware-1": {},
				},
				Services: map[string]*config.Service{
					"provider-1.service-1": {},
				},
			},
		},
		{
			desc: "Returns fully qualified elements from a multi-provider configuration map",
			given: config.Configurations{
				"provider-1": &config.Configuration{
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"router-1": {},
						},
						Middlewares: map[string]*config.Middleware{
							"middleware-1": {},
						},
						Services: map[string]*config.Service{
							"service-1": {},
						},
					},
				},
				"provider-2": &config.Configuration{
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"router-1": {},
						},
						Middlewares: map[string]*config.Middleware{
							"middleware-1": {},
						},
						Services: map[string]*config.Service{
							"service-1": {},
						},
					},
				},
			},
			expected: &config.HTTPConfiguration{
				Routers: map[string]*config.Router{
					"provider-1.router-1": {},
					"provider-2.router-1": {},
				},
				Middlewares: map[string]*config.Middleware{
					"provider-1.middleware-1": {},
					"provider-2.middleware-1": {},
				},
				Services: map[string]*config.Service{
					"provider-1.service-1": {},
					"provider-2.service-1": {},
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
