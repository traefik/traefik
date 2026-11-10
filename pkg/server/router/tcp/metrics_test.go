package tcp

import (
	"testing"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/service/tcp"
	tcp2 "github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

func TestManager_routerMetrics(t *testing.T) {
	testCases := []struct {
		desc   string
		router *dynamic.TCPRouter
	}{
		{
			desc: "non-TLS router",
			router: &dynamic.TCPRouter{
				EntryPoints: []string{"web"},
				Service:     "bar",
				Rule:        "HostSNI(`*`)",
			},
		},
		{
			desc: "TLS passthrough router",
			router: &dynamic.TCPRouter{
				EntryPoints: []string{"web"},
				Service:     "bar",
				Rule:        "HostSNI(`foo.bar`)",
				TLS:         &dynamic.RouterTCPTLSConfig{Passthrough: true},
			},
		},
		{
			desc: "TLS router",
			router: &dynamic.TCPRouter{
				EntryPoints: []string{"web"},
				Service:     "bar",
				Rule:        "HostSNI(`foo.bar`)",
				TLS:         &dynamic.RouterTCPTLSConfig{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			conf := &runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@file": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{{Address: "127.0.0.1:9999"}},
							},
						},
					},
				},
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"foo@file": {TCPRouter: test.router},
				},
			}

			dialerManager := tcp2.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
			serviceManager := tcp.NewManager(conf, dialerManager)

			tlsManager := traefiktls.NewManager(nil)
			tlsManager.UpdateConfigs(t.Context(),
				map[string]traefiktls.Store{},
				map[string]traefiktls.Options{"default": {}},
				[]*traefiktls.CertAndStores{})

			gauge := &gaugeMock{}
			observabilityMgr := middleware.NewObservabilityMgr(static.Configuration{}, &registryMock{gauge: gauge}, nil, nil, nil, nil)

			manager := NewManager(conf, serviceManager, tcpmiddleware.NewBuilder(conf.TCPMiddlewares),
				observabilityMgr, nil, nil, tlsManager, nil)

			handlers := manager.BuildHandlers(t.Context(), []string{"web"})
			require.Contains(t, handlers, "web")
			require.Empty(t, conf.TCPRouters["foo@file"].Err)

			// The service name is qualified, as for the HTTP router metrics.
			assert.Equal(t, []string{"router", "foo@file", "service", "bar@file"}, gauge.labelValues)
		})
	}
}

type registryMock struct {
	metrics.Registry

	gauge *gaugeMock
}

func (r *registryMock) IsTCPRouterEnabled() bool {
	return true
}

func (r *registryMock) RouterTCPOpenConnectionsGauge() gokitmetrics.Gauge {
	return r.gauge
}

type gaugeMock struct {
	labelValues []string
}

func (g *gaugeMock) With(labelValues ...string) gokitmetrics.Gauge {
	g.labelValues = labelValues
	return g
}

func (g *gaugeMock) Set(_ float64) {}

func (g *gaugeMock) Add(_ float64) {}
