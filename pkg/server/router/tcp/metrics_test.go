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
			"foo@file": {
				TCPRouter: &dynamic.TCPRouter{
					EntryPoints: []string{"web"},
					Service:     "bar",
					Rule:        "HostSNI(`*`)",
				},
			},
		},
	}

	dialerManager := tcp2.NewDialerManager(nil)
	dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})

	gauge := &gaugeMock{}
	observabilityMgr := middleware.NewObservabilityMgr(static.Configuration{}, &registryMock{gauge: gauge}, nil, nil, nil, nil)

	manager := NewManager(conf, tcp.NewManager(conf, dialerManager), tcpmiddleware.NewBuilder(conf.TCPMiddlewares),
		observabilityMgr, nil, nil, traefiktls.NewManager(nil), nil)

	handlers := manager.BuildHandlers(t.Context(), []string{"web"})
	require.Contains(t, handlers, "web")

	// The service name is qualified, as for the HTTP router metrics.
	assert.Equal(t, []string{"traefik.router.name", "foo@file", "traefik.service.name", "bar@file", "traefik.tls.mode", "none"}, gauge.labelValues)
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

func (r *registryMock) RouterTCPOpenTLSConnectionsGauge() gokitmetrics.Gauge {
	return &gaugeMock{}
}

type gaugeMock struct {
	gokitmetrics.Gauge

	labelValues []string
}

func (g *gaugeMock) With(labelValues ...string) gokitmetrics.Gauge {
	g.labelValues = labelValues
	return g
}
