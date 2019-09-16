package consulcatalog

import (
	"context"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConsulCatalog(t *testing.T) {

	m := &consulCatalogMock{}

	p := &Provider{
		getConsulClientFunc: func(cfg *EndpointConfig) (catalog consulCatalog, e error) {
			return m, nil
		},
		RefreshInterval:  types.Duration(time.Millisecond),
		Entrypoints:      []string{"web"},
		ExposedByDefault: true,
		Protocol:         "http",
	}

	catalogService := &api.CatalogService{
		ServiceName:    "service1",
		ServiceID:      "6f1ba69a-c7ae-4c5e-b636-2b49860aab00",
		ServiceAddress: "192.168.1.1",
		ServicePort:    1000,
		ServiceTags:    []string{"foo=bar"},
	}

	m.On("Services", mock.Anything).Return(map[string][]string{"service1": []string{"foo=bar"}}, &api.QueryMeta{}, nil)
	m.On("Service", "service1", mock.Anything, mock.Anything).Return([]*api.CatalogService{catalogService}, &api.QueryMeta{}, nil)

	ch := make(chan dynamic.Message)

	err := p.Provide(ch, safe.NewPool(context.Background()))
	require.NoError(t, err)

	expectedMessage := dynamic.Message{
		ProviderName: "consulcatalog",
		Configuration: &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"service1-router": {
						EntryPoints: []string{"web"},
						Middlewares: nil,
						Service:     "service1",
						Rule:        "",
						Priority:    0,
						TLS:         nil,
					},
				},
				Middlewares: map[string]*dynamic.Middleware{},
				Services: map[string]*dynamic.Service{
					"service1": {
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Sticky: nil,
							Servers: []dynamic.Server{
								{
									URL:    "http://192.168.1.1:1000",
									Scheme: "",
									Port:   "",
								},
							},
							HealthCheck:        nil,
							PassHostHeader:     false,
							ResponseForwarding: nil,
						},
						Weighted:  nil,
						Mirroring: nil,
					},
				},
			},
			TCP: &dynamic.TCPConfiguration{
				Routers:  map[string]*dynamic.TCPRouter{},
				Services: map[string]*dynamic.TCPService{},
			},
			TLS: nil,
		},
	}

	select {
	case msg := <-ch:
		assert.Equal(t, expectedMessage, msg)
	case <-time.After(time.Millisecond * 100):
		t.Fatalf("missing message")
	}

	m.AssertExpectations(t)
}
