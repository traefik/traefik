package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestGetRoutersByEntryPoints(t *testing.T) {
	testCases := []struct {
		desc        string
		conf        dynamic.Configuration
		entryPoints []string
		expected    map[string]map[string]*RouterInfo
	}{
		{
			desc:        "Empty Configuration without entrypoint",
			conf:        dynamic.Configuration{},
			entryPoints: []string{""},
			expected:    map[string]map[string]*RouterInfo{},
		},
		{
			desc:        "Empty Configuration with unknown entrypoints",
			conf:        dynamic.Configuration{},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*RouterInfo{},
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
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "HostSNI(`bar.foo`)",
						},
					},
				},
			},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*RouterInfo{},
		},
		{
			desc: "Valid configuration with a known entrypoint",
			conf: dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "HostSNI(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
							Rule:        "HostSNI(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: map[string]map[string]*RouterInfo{
				"web": {
					"foo": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`bar.foo`)",
						},
						Status: "enabled",
						Using:  []string{"web"},
					},
					"foobar": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "Host(`bar.foobar`)",
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
							Rule:        "Host(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "HostSNI(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
							Rule:        "HostSNI(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
			entryPoints: []string{"web", "webs"},
			expected: map[string]map[string]*RouterInfo{
				"web": {
					"foo": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`bar.foo`)",
						},
						Status: "enabled",
						Using:  []string{"web"},
					},
					"foobar": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "Host(`bar.foobar`)",
						},
						Status: "enabled",
						Using:  []string{"web", "webs"},
					},
				},
				"webs": {
					"bar": {
						Router: &dynamic.Router{
							EntryPoints: []string{"webs"},
							Service:     "bar-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: "enabled",
						Using:  []string{"webs"},
					},
					"foobar": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web", "webs"},
							Service:     "foobar-service@myprovider",
							Rule:        "Host(`bar.foobar`)",
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
			actual := runtimeConfig.GetRoutersByEntryPoints(context.Background(), test.entryPoints, false)
			assert.Equal(t, test.expected, actual)
		})
	}
}
