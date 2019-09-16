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

func TestGetConsulServicesNames(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		clientCatalog: m,
	}

	data := make(map[string][]string)
	data["service1"] = []string{}
	data["service2"] = []string{"foo"}
	data["service3"] = []string{"foo", "bar"}

	m.On("Services", mock.Anything).Return(data, &api.QueryMeta{}, nil)

	names, err := p.getConsulServicesNames(context.Background())
	assert.NoError(t, err)
	require.Len(t, names, 3)

	s, ok := names["service1"]
	require.True(t, ok)
	assert.Len(t, s, 0)

	s, ok = names["service2"]
	require.True(t, ok)
	assert.Len(t, s, 1)
	assert.Contains(t, s, "foo")

	s, ok = names["service3"]
	require.True(t, ok)
	assert.Len(t, s, 2)
	assert.Contains(t, s, "foo")
	assert.Contains(t, s, "bar")

	m.AssertExpectations(t)
}

func TestGetConsulServiceData_APIError_Services(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{}, &api.QueryMeta{}, fmt.Errorf("api error services"))

	_, err := p.getConsulServicesData(context.Background())
	require.Error(t, err)
	assert.Equal(t, "api error services", err.Error())
	m.AssertExpectations(t)
}

func TestGetConsulServiceData_APIError_Service(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		Protocol:      "http",
		Entrypoints:   []string{"web", "api"},
		RouterRule:    "Path(`/`)",
		clientCatalog: m,
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": {"traefik.enabled", "traefik.entrypoints=foo,bar", "traefik.router.rule=bar"}}, &api.QueryMeta{}, nil)
	m.On("Service", "service1", "", mock.Anything).Return([]*api.CatalogService{}, &api.QueryMeta{}, fmt.Errorf("api error service1"))

	_, err := p.getConsulServicesData(context.Background())
	require.Error(t, err)
	assert.Equal(t, "api error service1", err.Error())
	m.AssertExpectations(t)
}

func TestGetConsulServiceData(t *testing.T) {
	m := &consulCatalogMock{}
	p := &Provider{
		Protocol:         "http",
		Entrypoints:      []string{"web", "api"},
		RouterRule:       "Path(`/`)",
		clientCatalog:    m,
		ExposedByDefault: true,
	}

	catalogService := &api.CatalogService{
		ServiceName:    "service1",
		ServiceID:      "6f1ba69a-c7ae-4c5e-b636-2b49860aab00",
		ServiceAddress: "192.168.1.1",
		ServicePort:    1000,
		ServiceTags:    []string{"foo=bar"},
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": []string{}}, &api.QueryMeta{}, nil)
	m.On("Service", "service1", mock.Anything, mock.Anything).Return([]*api.CatalogService{catalogService}, &api.QueryMeta{}, nil)

	data, err := p.getConsulServicesData(context.Background())
	require.NoError(t, err)
	m.AssertExpectations(t)

	expectedConsulCatalogData := &consulCatalogData{
		Items: []*consulCatalogItem{
			{
				ID:      "6f1ba69a-c7ae-4c5e-b636-2b49860aab00",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    1000,
				Labels:  map[string]string{"foo": "bar"},
			},
		},
	}

	assert.Equal(t, expectedConsulCatalogData, data)
}
