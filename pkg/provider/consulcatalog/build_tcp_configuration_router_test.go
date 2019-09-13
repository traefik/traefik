package consulcatalog

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildTCPRouterConfiguration_WithoutTags(t *testing.T) {
	p := &Provider{
		Entrypoints: []string{"web", "api"},
		RouterRule:  "Path(`/`)",
	}

	name, router := p.buildTCPRouterConfiguration("router1", []string{})
	assert.Equal(t, "router1", name)
	require.Len(t, router.EntryPoints, 2)
	assert.Contains(t, router.EntryPoints, "web")
	assert.Contains(t, router.EntryPoints, "api")
	assert.Equal(t, "Path(`/`)", router.Rule)
}

func TestBuildTCPRouterConfiguration_WithEndpointsTags(t *testing.T) {
	p := &Provider{
		prefixes: prefixes{
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Entrypoints: []string{"web", "api"},
		RouterRule:  "Path(`/`)",
	}

	_, router := p.buildTCPRouterConfiguration("router1", []string{"traefik.entrypoints=foo,bar,baz"})
	require.Len(t, router.EntryPoints, 3)
	assert.Contains(t, router.EntryPoints, "foo")
	assert.Contains(t, router.EntryPoints, "bar")
	assert.Contains(t, router.EntryPoints, "baz")
}

func TestBuildTCPRouterConfiguration_WithEndpointsTags_Empty(t *testing.T) {
	p := &Provider{
		prefixes: prefixes{
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Entrypoints: []string{"web", "api"},
		RouterRule:  "Path(`/`)",
	}

	_, router := p.buildTCPRouterConfiguration("router1", []string{"traefik.entrypoints="})
	require.Len(t, router.EntryPoints, 1)
	assert.Contains(t, router.EntryPoints, "")
}

func TestBuildTCPRouterConfiguration_WithRouterRuleTag(t *testing.T) {
	p := &Provider{
		prefixes: prefixes{
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Entrypoints: []string{"web", "api"},
		RouterRule:  "Path(`/`)",
	}

	_, router := p.buildTCPRouterConfiguration("router1", []string{"traefik.router.rule=Path(`/foo`)"})
	assert.Equal(t, "Path(`/foo`)", router.Rule)
}
