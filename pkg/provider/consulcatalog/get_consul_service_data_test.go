package consulcatalog

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type consulCatalogMock struct {
	mock.Mock
}

func (m *consulCatalogMock) Service(name string, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error) {
	args := m.Called(name, tag, q)
	return args.Get(0).([]*api.CatalogService), args.Get(1).(*api.QueryMeta), args.Error(2)
}

func (m *consulCatalogMock) Services(q *api.QueryOptions) (map[string][]string, *api.QueryMeta, error) {
	args := m.Called(q)
	return args.Get(0).(map[string][]string), args.Get(1).(*api.QueryMeta), args.Error(2)
}

func TestServiceNames_APIError(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{}, &api.QueryMeta{}, fmt.Errorf("api error"))

	_, err := p.serviceNames(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "api error", err.Error())
	m.AssertExpectations(t)
}

func TestServiceNames_WithoutTags(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{}, &api.QueryMeta{}, nil)

	names, err := p.serviceNames(context.Background())
	assert.NoError(t, err)
	assert.Len(t, names, 0)
	m.AssertExpectations(t)
}

func TestServiceNames(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		prefixes: prefixes{
			enabled: "traefik.enabled",
		},
		clientCatalog: m,
	}

	data := make(map[string][]string)
	data["service1"] = []string{}
	data["service2"] = []string{"traefik.enabled", "foo"}
	data["service3"] = []string{"traefik.enable"}

	m.On("Services", mock.Anything).Return(data, &api.QueryMeta{}, nil)

	names, err := p.serviceNames(context.Background())
	assert.NoError(t, err)
	require.Len(t, names, 1)

	s, ok := names["service2"]
	require.True(t, ok)
	assert.Len(t, s, 2)
	assert.Contains(t, s, "traefik.enabled")
	assert.Contains(t, s, "foo")

	m.AssertExpectations(t)
}

func TestGetConsulServiceData_APIError_Services(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		prefixes: prefixes{
			protocol:          "traefik.protocol=",
			enabled:           "traefik.enabled",
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{}, &api.QueryMeta{}, fmt.Errorf("api error service"))

	_, _, _, _, err := p.getConsulServicesData(context.Background())
	require.Error(t, err)
	assert.Equal(t, "api error service", err.Error())
	m.AssertExpectations(t)
}

func TestGetConsulServiceData_APIError_Service(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		prefixes: prefixes{
			protocol:          "traefik.protocol=",
			enabled:           "traefik.enabled",
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": {"traefik.enabled", "traefik.entrypoints=foo,bar", "traefik.router.rule=bar"}}, &api.QueryMeta{}, nil)
	m.On("Service", "service1", "", mock.Anything).Return([]*api.CatalogService{}, &api.QueryMeta{}, fmt.Errorf("api error"))

	_, _, _, _, err := p.getConsulServicesData(context.Background())
	require.Error(t, err)
	assert.Equal(t, "api error", err.Error())
	m.AssertExpectations(t)
}

func TestGetConsulServiceData_BadProtocolTag(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		prefixes: prefixes{
			protocol:          "traefik.protocol=",
			enabled:           "traefik.enabled",
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": {"traefik.enabled", "traefik.protocol=WRONG", "traefik.entrypoints=foo,bar", "traefik.router.rule=bar"}}, &api.QueryMeta{}, nil)

	_, _, _, _, err := p.getConsulServicesData(context.Background())
	require.Error(t, err)
	assert.Equal(t, "wrong protocol 'WRONG', allowed 'http' or 'tcp'", err.Error())
	m.AssertExpectations(t)
}

func TestGetConsulServiceData_HTTP(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		prefixes: prefixes{
			protocol:          "traefik.protocol=",
			enabled:           "traefik.enabled",
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	catalogService := &api.CatalogService{
		ServiceAddress: "192.168.1.1",
		ServicePort:    1000,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": {"traefik.enabled", "traefik.protocol=http", "traefik.entrypoints=foo,bar", "traefik.router.rule=baz"}}, &api.QueryMeta{}, nil)
	m.On("Service", "service1", "", mock.Anything).Return([]*api.CatalogService{catalogService}, &api.QueryMeta{}, nil)

	httpRouters, httpServices, tcpRouters, tcpServices, err := p.getConsulServicesData(context.Background())
	require.NoError(t, err)
	m.AssertExpectations(t)
	require.Len(t, httpRouters, 1)
	require.Len(t, httpServices, 1)
	require.Len(t, tcpRouters, 0)
	require.Len(t, tcpServices, 0)

	httpRouter1, ok := httpRouters["service1"]
	require.True(t, ok)

	assert.Equal(t, "service1", httpRouter1.Service)
	require.Len(t, httpRouter1.EntryPoints, 2)
	assert.Contains(t, httpRouter1.EntryPoints, "foo")
	assert.Contains(t, httpRouter1.EntryPoints, "bar")
	assert.Equal(t, "baz", httpRouter1.Rule)

	httpService1, ok := httpServices["service1"]
	require.True(t, ok)

	require.NotNil(t, httpService1.LoadBalancer)
	require.Len(t, httpService1.LoadBalancer.Servers, 1)
	assert.Equal(t, "http://192.168.1.1:1000", httpService1.LoadBalancer.Servers[0].URL)
}

func TestGetConsulServiceData_TCP(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		prefixes: prefixes{
			protocol:          "traefik.protocol=",
			enabled:           "traefik.enabled",
			routerRule:        "traefik.router.rule=",
			routerEntrypoints: "traefik.entrypoints=",
		},
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	catalogService := &api.CatalogService{
		ServiceAddress: "192.168.1.1",
		ServicePort:    1000,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": {"traefik.enabled", "traefik.protocol=tcp", "traefik.entrypoints=foo,bar", "traefik.router.rule=baz"}}, &api.QueryMeta{}, nil)
	m.On("Service", "service1", "", mock.Anything).Return([]*api.CatalogService{catalogService}, &api.QueryMeta{}, nil)

	httpRouters, httpServices, tcpRouters, tcpServices, err := p.getConsulServicesData(context.Background())
	require.NoError(t, err)
	m.AssertExpectations(t)
	require.Len(t, httpRouters, 0)
	require.Len(t, httpServices, 0)
	require.Len(t, tcpRouters, 1)
	require.Len(t, tcpServices, 1)

	tcpRouter1, ok := tcpRouters["service1"]
	require.True(t, ok)

	assert.Equal(t, "service1", tcpRouter1.Service)
	require.Len(t, tcpRouter1.EntryPoints, 2)
	assert.Contains(t, tcpRouter1.EntryPoints, "foo")
	assert.Contains(t, tcpRouter1.EntryPoints, "bar")
	assert.Equal(t, "baz", tcpRouter1.Rule)

	httpService1, ok := tcpServices["service1"]
	require.True(t, ok)

	require.NotNil(t, httpService1.LoadBalancer)
	require.Len(t, httpService1.LoadBalancer.Servers, 1)
	assert.Equal(t, "192.168.1.1:1000", httpService1.LoadBalancer.Servers[0].Address)
}
