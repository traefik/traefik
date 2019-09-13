package consulcatalog

import (
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildTCPServiceConfiguration(t *testing.T) {
	p := &Provider{
		Entrypoints: []string{"web", "api"},
		RouterRule:  "Path(`/`)",
	}

	consulServices := []*api.CatalogService{
		{ServiceAddress: "192.168.1.1", ServicePort: 1000},
		{ServiceAddress: "192.168.1.2", ServicePort: 2000},
	}

	name, service := p.buildTCPServiceConfiguration("service1", consulServices)

	assert.Equal(t, "service1", name)
	require.NotNil(t, service.LoadBalancer)
	require.Len(t, service.LoadBalancer.Servers, 2)
	assert.Equal(t, "192.168.1.1:1000", service.LoadBalancer.Servers[0].Address)
	assert.Equal(t, "192.168.1.2:2000", service.LoadBalancer.Servers[1].Address)
}
