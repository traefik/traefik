package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestGetUDPRoutersByEntryPoints(t *testing.T) {
	testCases := []struct {
		desc        string
		conf        dynamic.Configuration
		entryPoints []string
		expected    map[string]map[string]*UDPRouterInfo
	}{
		{
			desc:        "Empty Configuration without entrypoint",
			conf:        dynamic.Configuration{},
			entryPoints: []string{""},
			expected:    map[string]map[string]*UDPRouterInfo{},
		},
		{
			desc:        "Empty Configuration with unknown entrypoints",
			conf:        dynamic.Configuration{},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*UDPRouterInfo{},
		},
		{
			desc: "Valid configuration with an unknown entrypoint",
			conf: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`bar.foo`)",
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
					},
				},
			},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*UDPRouterInfo{},
		},
		{
			desc: "Valid configuration with a known entrypoint",
			conf: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: map[string]map[string]*UDPRouterInfo{
				"web": {
					"foo": {
						UDPRouter: &dynamic.UDPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
						Status: "enabled",
						Using:  []string{"web"},
					},
					"foobar": {
						UDPRouter: &dynamic.UDPRouter{
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
						Status: "warning",
						Err:    []string{`entryPoint "webs" doesn't exist`},
						Using:  []string{"web"},
					},
				},
			},
		},
		{
			desc: "Valid configuration with multiple known entrypoints",
			conf: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
					},
				},
			},
			entryPoints: []string{"web", "webs"},
			expected: map[string]map[string]*UDPRouterInfo{
				"web": {
					"foo": {
						UDPRouter: &dynamic.UDPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
						},
						Status: "enabled",
						Using:  []string{"web"},
					},
					"foobar": {
						UDPRouter: &dynamic.UDPRouter{
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
						Status: "enabled",
						Using:  []string{"web", "webs"},
					},
				},
				"webs": {
					"bar": {
						UDPRouter: &dynamic.UDPRouter{
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
						},
						Status: "enabled",
						Using:  []string{"webs"},
					},
					"foobar": {
						UDPRouter: &dynamic.UDPRouter{
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
						},
						Status: "enabled",
						Using:  []string{"web", "webs"},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			runtimeConfig := NewConfig(test.conf)
			actual := runtimeConfig.GetUDPRoutersByEntryPoints(context.Background(), test.entryPoints)
			assert.Equal(t, test.expected, actual)
		})
	}
}
