package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/containous/alice"
	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/config/runtime"
	"github.com/containous/traefik/pkg/healthcheck"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/metrics"
	"github.com/containous/traefik/pkg/middlewares/accesslog"
	"github.com/containous/traefik/pkg/middlewares/emptybackendhandler"
	metricsMiddle "github.com/containous/traefik/pkg/middlewares/metrics"
	"github.com/containous/traefik/pkg/middlewares/pipelining"
	"github.com/containous/traefik/pkg/server/cookie"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/vulcand/oxy/roundrobin"
)

const (
	defaultHealthCheckInterval = 30 * time.Second
	defaultHealthCheckTimeout  = 5 * time.Second
)

// NewManager creates a new Manager
func NewManager(configs map[string]*runtime.ServiceInfo, defaultRoundTripper http.RoundTripper, metricsRegistry metrics.Registry) *Manager {
	return &Manager{
		metricsRegistry:     metricsRegistry,
		bufferPool:          newBufferPool(),
		defaultRoundTripper: defaultRoundTripper,
		balancers:           make(map[string][]healthcheck.BalancerHandler),
		configs:             configs,
	}
}

// Manager The service manager
type Manager struct {
	metricsRegistry     metrics.Registry
	bufferPool          httputil.BufferPool
	defaultRoundTripper http.RoundTripper
	balancers           map[string][]healthcheck.BalancerHandler
	configs             map[string]*runtime.ServiceInfo
}

// BuildHTTP Creates a http.Handler for a service configuration.
func (m *Manager) BuildHTTP(rootCtx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error) {
	ctx := log.With(rootCtx, log.Str(log.ServiceName, serviceName))

	serviceName = internal.GetQualifiedName(ctx, serviceName)
	ctx = internal.AddProviderInContext(ctx, serviceName)

	conf, ok := m.configs[serviceName]
	if !ok {
		return nil, fmt.Errorf("the service %q does not exist", serviceName)
	}

	// TODO Should handle multiple service types
	// FIXME Check if the service is declared multiple times with different types
	if conf.LoadBalancer == nil {
		sErr := fmt.Errorf("the service %q doesn't have any load balancer", serviceName)
		conf.AddError(sErr, true)
		return nil, sErr
	}

	lb, err := m.getLoadBalancerServiceHandler(ctx, serviceName, conf.LoadBalancer, responseModifier)
	if err != nil {
		conf.AddError(err, true)
		return nil, err
	}

	return lb, nil
}

func (m *Manager) getLoadBalancerServiceHandler(
	ctx context.Context,
	serviceName string,
	service *dynamic.LoadBalancerService,
	responseModifier func(*http.Response) error,
) (http.Handler, error) {
	fwd, err := buildProxy(service.PassHostHeader, service.ResponseForwarding, m.defaultRoundTripper, m.bufferPool, responseModifier)
	if err != nil {
		return nil, err
	}

	alHandler := func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, accesslog.ServiceName, serviceName, accesslog.AddServiceFields), nil
	}
	chain := alice.New()
	if m.metricsRegistry != nil && m.metricsRegistry.IsEnabled() {
		chain = chain.Append(metricsMiddle.WrapServiceHandler(ctx, m.metricsRegistry, serviceName))
	}

	handler, err := chain.Append(alHandler).Then(pipelining.New(ctx, fwd, "pipelining"))
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

		// TODO aggregate
		balancer := balancers[0]

		// TODO Should all the services handle healthcheck? Handle different types
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

func buildHealthCheckOptions(ctx context.Context, lb healthcheck.BalancerHandler, backend string, hc *dynamic.HealthCheck) *healthcheck.Options {
	if hc == nil || hc.Path == "" {
		return nil
	}

	logger := log.FromContext(ctx)

	interval := defaultHealthCheckInterval
	if hc.Interval != "" {
		intervalOverride, err := time.ParseDuration(hc.Interval)
		switch {
		case err != nil:
			logger.Errorf("Illegal health check interval for '%s': %s", backend, err)
		case intervalOverride <= 0:
			logger.Errorf("Health check interval smaller than zero for service '%s'", backend)
		default:
			interval = intervalOverride
		}
	}

	timeout := defaultHealthCheckTimeout
	if hc.Timeout != "" {
		timeoutOverride, err := time.ParseDuration(hc.Timeout)
		switch {
		case err != nil:
			logger.Errorf("Illegal health check timeout for backend '%s': %s", backend, err)
		case timeoutOverride <= 0:
			logger.Errorf("Health check timeout smaller than zero for backend '%s', backend", backend)
		default:
			timeout = timeoutOverride
		}
	}

	if timeout >= interval {
		logger.Warnf("Health check timeout for backend '%s' should be lower than the health check interval. Interval set to timeout + 1 second (%s).", backend, interval)
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

func (m *Manager) getLoadBalancer(ctx context.Context, serviceName string, service *dynamic.LoadBalancerService, fwd http.Handler) (healthcheck.BalancerHandler, error) {
	logger := log.FromContext(ctx)
	logger.Debug("Creating load-balancer")

	var options []roundrobin.LBOption

	var cookieName string
	if stickiness := service.Stickiness; stickiness != nil {
		cookieName = cookie.GetName(stickiness.CookieName, serviceName)
		opts := roundrobin.CookieOptions{HTTPOnly: stickiness.HTTPOnlyCookie, Secure: stickiness.SecureCookie}
		options = append(options, roundrobin.EnableStickySession(roundrobin.NewStickySessionWithOptions(cookieName, opts)))
		logger.Debugf("Sticky session cookie name: %v", cookieName)
	}

	lb, err := roundrobin.New(fwd, options...)
	if err != nil {
		return nil, err
	}

	lbsu := healthcheck.NewLBStatusUpdater(lb, m.configs[serviceName])
	if err := m.upsertServers(ctx, lbsu, service.Servers); err != nil {
		return nil, fmt.Errorf("error configuring load balancer for service %s: %v", serviceName, err)
	}

	return lb, nil
}

func (m *Manager) upsertServers(ctx context.Context, lb healthcheck.BalancerHandler, servers []dynamic.Server) error {
	logger := log.FromContext(ctx)

	for name, srv := range servers {
		u, err := url.Parse(srv.URL)
		if err != nil {
			return fmt.Errorf("error parsing server URL %s: %v", srv.URL, err)
		}

		logger.WithField(log.ServerName, name).Debugf("Creating server %d %s", name, u)

		if err := lb.UpsertServer(u, roundrobin.Weight(1)); err != nil {
			return fmt.Errorf("error adding server %s to load balancer: %v", srv.URL, err)
		}

		// FIXME Handle Metrics
	}
	return nil
}
