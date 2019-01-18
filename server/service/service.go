package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/containous/alice"
	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/emptybackendhandler"
	"github.com/containous/traefik/old/middlewares/pipelining"
	"github.com/containous/traefik/server/cookie"
	"github.com/containous/traefik/server/internal"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

const (
	defaultHealthCheckInterval = 30 * time.Second
	defaultHealthCheckTimeout  = 5 * time.Second
)

// See oxy/roundrobin/rr.go
type balancerHandler interface {
	Servers() []*url.URL
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	ServerWeight(u *url.URL) (int, bool)
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
	NextServer() (*url.URL, error)
	Next() http.Handler
}

// NewManager creates a new Manager
func NewManager(configs map[string]*config.Service, defaultRoundTripper http.RoundTripper) *Manager {
	return &Manager{
		bufferPool:          newBufferPool(),
		defaultRoundTripper: defaultRoundTripper,
		balancers:           make(map[string][]healthcheck.BalancerHandler),
		configs:             configs,
	}
}

// Manager The service manager
type Manager struct {
	bufferPool          httputil.BufferPool
	defaultRoundTripper http.RoundTripper
	balancers           map[string][]healthcheck.BalancerHandler
	configs             map[string]*config.Service
}

// Build Creates a http.Handler for a service configuration.
func (m *Manager) Build(rootCtx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error) {
	ctx := log.With(rootCtx, log.Str(log.ServiceName, serviceName))

	serviceName = internal.GetQualifiedName(ctx, serviceName)
	ctx = internal.AddProviderInContext(ctx, serviceName)

	if conf, ok := m.configs[serviceName]; ok {
		// TODO Should handle multiple service types
		if conf.LoadBalancer != nil {
			return m.getLoadBalancerServiceHandler(ctx, serviceName, conf.LoadBalancer, responseModifier)
		}
		return nil, fmt.Errorf("the service %q doesn't have any load balancer", serviceName)
	}
	return nil, fmt.Errorf("the service %q does not exits", serviceName)
}

func (m *Manager) getLoadBalancerServiceHandler(
	ctx context.Context,
	serviceName string,
	service *config.LoadBalancerService,
	responseModifier func(*http.Response) error,
) (http.Handler, error) {

	fwd, err := m.buildForwarder(service.PassHostHeader, service.ResponseForwarding, responseModifier)
	if err != nil {
		return nil, err
	}

	alHandler := func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, accesslog.ServiceName, serviceName, accesslog.AddServiceFields), nil
	}

	handler, err := alice.New().Append(alHandler).Then(pipelining.NewPipelining(fwd))
	if err != nil {
		return nil, err
	}

	balancer, err := m.getLoadBalancer(ctx, serviceName, service, handler)
	if err != nil {
		return nil, err
	}

	// TODO rename and checks
	m.balancers[serviceName] = append(m.balancers[serviceName], balancer)

	// Empty (backend with no servers)
	return emptybackendhandler.New(balancer), nil
}

// LaunchHealthCheck Launches the health checks.
func (m *Manager) LaunchHealthCheck() {
	backendConfigs := make(map[string]*healthcheck.BackendConfig)

	for serviceName, balancers := range m.balancers {
		ctx := log.With(context.Background(), log.Str(log.ServiceName, serviceName))

		// FIXME aggregate
		balancer := balancers[0]

		// FIXME Should all the services handle healthcheck? Handle different types
		service := m.configs[serviceName].LoadBalancer

		// Health Check
		var backendHealthCheck *healthcheck.BackendConfig
		if hcOpts := buildHealthCheckOptions(ctx, balancer, serviceName, service.HealthCheck); hcOpts != nil {
			log.FromContext(ctx).Debugf("Setting up healthcheck for service %s with %s", serviceName, *hcOpts)

			hcOpts.Transport = m.defaultRoundTripper
			backendHealthCheck = healthcheck.NewBackendConfig(*hcOpts, serviceName)
		}

		if backendHealthCheck != nil {
			backendConfigs[serviceName] = backendHealthCheck
		}
	}

	// FIXME metrics and context
	healthcheck.GetHealthCheck().SetBackendsConfiguration(context.TODO(), backendConfigs)
}

func buildHealthCheckOptions(ctx context.Context, lb healthcheck.BalancerHandler, backend string, hc *config.HealthCheck) *healthcheck.Options {
	if hc == nil || hc.Path == "" {
		return nil
	}

	logger := log.FromContext(ctx)

	interval := defaultHealthCheckInterval
	if hc.Interval != "" {
		intervalOverride, err := time.ParseDuration(hc.Interval)
		if err != nil {
			logger.Errorf("Illegal health check interval for '%s': %s", backend, err)
		} else if intervalOverride <= 0 {
			logger.Errorf("Health check interval smaller than zero for service '%s'", backend)
		} else {
			interval = intervalOverride
		}
	}

	timeout := defaultHealthCheckTimeout
	if hc.Timeout != "" {
		timeoutOverride, err := time.ParseDuration(hc.Timeout)
		if err != nil {
			logger.Errorf("Illegal health check timeout for backend '%s': %s", backend, err)
		} else if timeoutOverride <= 0 {
			logger.Errorf("Health check timeout smaller than zero for backend '%s', backend", backend)
		} else {
			timeout = timeoutOverride
		}
	}

	if timeout >= interval {
		logger.Warnf("Health check timeout for backend '%s' should be lower than the health check interval. Interval set to timeout + 1 second (%s).", backend)
	}

	return &healthcheck.Options{
		Scheme:   hc.Scheme,
		Path:     hc.Path,
		Port:     hc.Port,
		Interval: interval,
		Timeout:  timeout,
		LB:       lb,
		Hostname: hc.Hostname,
		Headers:  hc.Headers,
	}
}

func (m *Manager) getLoadBalancer(ctx context.Context, serviceName string, service *config.LoadBalancerService, fwd http.Handler) (healthcheck.BalancerHandler, error) {
	logger := log.FromContext(ctx)

	var stickySession *roundrobin.StickySession
	var cookieName string
	if stickiness := service.Stickiness; stickiness != nil {
		cookieName = cookie.GetName(stickiness.CookieName, serviceName)
		stickySession = roundrobin.NewStickySession(cookieName)
	}

	var lb healthcheck.BalancerHandler

	if service.Method == "drr" {
		logger.Debug("Creating drr load-balancer")
		rr, err := roundrobin.New(fwd)
		if err != nil {
			return nil, err
		}

		if stickySession != nil {
			logger.Debugf("Sticky session cookie name: %v", cookieName)

			lb, err = roundrobin.NewRebalancer(rr, roundrobin.RebalancerStickySession(stickySession))
			if err != nil {
				return nil, err
			}
		} else {
			lb, err = roundrobin.NewRebalancer(rr)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if service.Method != "wrr" {
			logger.Warnf("Invalid load-balancing method %q, fallback to 'wrr' method", service.Method)
		}

		logger.Debug("Creating wrr load-balancer")

		if stickySession != nil {
			logger.Debugf("Sticky session cookie name: %v", cookieName)

			var err error
			lb, err = roundrobin.New(fwd, roundrobin.EnableStickySession(stickySession))
			if err != nil {
				return nil, err
			}
		} else {
			var err error
			lb, err = roundrobin.New(fwd)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := m.upsertServers(ctx, lb, service.Servers); err != nil {
		return nil, fmt.Errorf("error configuring load balancer for service %s: %v", serviceName, err)
	}

	return lb, nil
}

func (m *Manager) upsertServers(ctx context.Context, lb healthcheck.BalancerHandler, servers []config.Server) error {
	logger := log.FromContext(ctx)

	for name, srv := range servers {
		u, err := url.Parse(srv.URL)
		if err != nil {
			return fmt.Errorf("error parsing server URL %s: %v", srv.URL, err)
		}

		logger.WithField(log.ServerName, name).Debugf("Creating server %d at %s with weight %d", name, u, srv.Weight)

		if err := lb.UpsertServer(u, roundrobin.Weight(srv.Weight)); err != nil {
			return fmt.Errorf("error adding server %s to load balancer: %v", srv.URL, err)
		}

		// FIXME Handle Metrics
	}
	return nil
}

func (m *Manager) buildForwarder(passHostHeader bool, responseForwarding *config.ResponseForwarding, responseModifier func(*http.Response) error) (http.Handler, error) {

	var flushInterval parse.Duration
	if responseForwarding != nil {
		err := flushInterval.Set(responseForwarding.FlushInterval)
		if err != nil {
			return nil, fmt.Errorf("error creating flush interval: %v", err)
		}
	}

	return forward.New(
		forward.Stream(true),
		forward.PassHostHeader(passHostHeader),
		forward.RoundTripper(m.defaultRoundTripper),
		forward.ResponseModifier(responseModifier),
		forward.BufferPool(m.bufferPool),
		forward.StreamingFlushInterval(time.Duration(flushInterval)),
		forward.WebsocketConnectionClosedHook(func(req *http.Request, conn net.Conn) {
			server := req.Context().Value(http.ServerContextKey).(*http.Server)
			if server != nil {
				connState := server.ConnState
				if connState != nil {
					connState(conn, http.StateClosed)
				}
			}
		}),
	)
}
