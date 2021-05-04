package healthcheck

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	gokitmetrics "github.com/go-kit/kit/metrics"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/vulcand/oxy/roundrobin"
)

const (
	serverUp   = "UP"
	serverDown = "DOWN"
)

var (
	singleton *HealthCheck
	once      sync.Once
)

// Balancer is the set of operations required to manage the list of servers in a load-balancer.
type Balancer interface {
	Servers() []*url.URL
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
}

// BalancerHandler includes functionality for load-balancing management.
type BalancerHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	Balancer
}

type metricsHealthcheck struct {
	serverUpGauge gokitmetrics.Gauge
}

// Options are the public health check options.
type Options struct {
	Headers         map[string]string
	Hostname        string
	Scheme          string
	Path            string
	Port            int
	FollowRedirects bool
	Transport       http.RoundTripper
	Interval        time.Duration
	Timeout         time.Duration
	LB              Balancer
}

func (opt Options) String() string {
	return fmt.Sprintf("[Hostname: %s Headers: %v Path: %s Port: %d Interval: %s Timeout: %s FollowRedirects: %v]", opt.Hostname, opt.Headers, opt.Path, opt.Port, opt.Interval, opt.Timeout, opt.FollowRedirects)
}

type backendURL struct {
	url    *url.URL
	weight int
}

// BackendConfig HealthCheck configuration for a backend.
type BackendConfig struct {
	Options
	name         string
	disabledURLs []backendURL
}

func (b *BackendConfig) newRequest(serverURL *url.URL) (*http.Request, error) {
	u, err := serverURL.Parse(b.Path)
	if err != nil {
		return nil, err
	}

	if len(b.Scheme) > 0 {
		u.Scheme = b.Scheme
	}

	if b.Port != 0 {
		u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(b.Port))
	}

	return http.NewRequest(http.MethodGet, u.String(), http.NoBody)
}

// this function adds additional http headers and hostname to http.request.
func (b *BackendConfig) addHeadersAndHost(req *http.Request) *http.Request {
	if b.Options.Hostname != "" {
		req.Host = b.Options.Hostname
	}

	for k, v := range b.Options.Headers {
		req.Header.Set(k, v)
	}
	return req
}

// HealthCheck struct.
type HealthCheck struct {
	Backends map[string]*BackendConfig
	metrics  metricsHealthcheck
	cancel   context.CancelFunc
}

// SetBackendsConfiguration set backends configuration.
func (hc *HealthCheck) SetBackendsConfiguration(parentCtx context.Context, backends map[string]*BackendConfig) {
	hc.Backends = backends
	if hc.cancel != nil {
		hc.cancel()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	hc.cancel = cancel

	for _, backend := range backends {
		currentBackend := backend
		safe.Go(func() {
			hc.execute(ctx, currentBackend)
		})
	}
}

func (hc *HealthCheck) execute(ctx context.Context, backend *BackendConfig) {
	logger := log.FromContext(ctx)

	logger.Debugf("Initial health check for backend: %q", backend.name)
	hc.checkServersLB(ctx, backend)

	ticker := time.NewTicker(backend.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debugf("Stopping current health check goroutines of backend: %s", backend.name)
			return
		case <-ticker.C:
			logger.Debugf("Routine health check refresh for backend: %s", backend.name)
			hc.checkServersLB(ctx, backend)
		}
	}
}

func (hc *HealthCheck) checkServersLB(ctx context.Context, backend *BackendConfig) {
	logger := log.FromContext(ctx)

	enabledURLs := backend.LB.Servers()

	var newDisabledURLs []backendURL
	for _, disabledURL := range backend.disabledURLs {
		serverUpMetricValue := float64(0)

		if err := checkHealth(disabledURL.url, backend); err == nil {
			logger.Warnf("Health check up: returning to server list. Backend: %q URL: %q Weight: %d",
				backend.name, disabledURL.url.String(), disabledURL.weight)
			if err = backend.LB.UpsertServer(disabledURL.url, roundrobin.Weight(disabledURL.weight)); err != nil {
				logger.Error(err)
			}
			serverUpMetricValue = 1
		} else {
			logger.Warnf("Health check still failing. Backend: %q URL: %q Reason: %s", backend.name, disabledURL.url.String(), err)
			newDisabledURLs = append(newDisabledURLs, disabledURL)
		}

		labelValues := []string{"service", backend.name, "url", disabledURL.url.String()}
		hc.metrics.serverUpGauge.With(labelValues...).Set(serverUpMetricValue)
	}

	backend.disabledURLs = newDisabledURLs

	for _, enabledURL := range enabledURLs {
		serverUpMetricValue := float64(1)

		if err := checkHealth(enabledURL, backend); err != nil {
			weight := 1
			rr, ok := backend.LB.(*roundrobin.RoundRobin)
			if ok {
				var gotWeight bool
				weight, gotWeight = rr.ServerWeight(enabledURL)
				if !gotWeight {
					weight = 1
				}
			}

			logger.Warnf("Health check failed, removing from server list. Backend: %q URL: %q Weight: %d Reason: %s",
				backend.name, enabledURL.String(), weight, err)
			if err := backend.LB.RemoveServer(enabledURL); err != nil {
				logger.Error(err)
			}

			backend.disabledURLs = append(backend.disabledURLs, backendURL{enabledURL, weight})
			serverUpMetricValue = 0
		}

		labelValues := []string{"service", backend.name, "url", enabledURL.String()}
		hc.metrics.serverUpGauge.With(labelValues...).Set(serverUpMetricValue)
	}
}

// GetHealthCheck returns the health check which is guaranteed to be a singleton.
func GetHealthCheck(registry metrics.Registry) *HealthCheck {
	once.Do(func() {
		singleton = newHealthCheck(registry)
	})
	return singleton
}

func newHealthCheck(registry metrics.Registry) *HealthCheck {
	return &HealthCheck{
		Backends: make(map[string]*BackendConfig),
		metrics: metricsHealthcheck{
			serverUpGauge: registry.ServiceServerUpGauge(),
		},
	}
}

// NewBackendConfig Instantiate a new BackendConfig.
func NewBackendConfig(options Options, backendName string) *BackendConfig {
	return &BackendConfig{
		Options: options,
		name:    backendName,
	}
}

// checkHealth returns a nil error in case it was successful and otherwise
// a non-nil error with a meaningful description why the health check failed.
func checkHealth(serverURL *url.URL, backend *BackendConfig) error {
	req, err := backend.newRequest(serverURL)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req = backend.addHeadersAndHost(req)

	client := http.Client{
		Timeout:   backend.Options.Timeout,
		Transport: backend.Options.Transport,
	}

	if !backend.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("received error status code: %v", resp.StatusCode)
	}

	return nil
}

// StatusUpdater should be implemented by a service that, when its status
// changes (e.g. all if its children are down), needs to propagate upwards (to
// their parent(s)) that change.
type StatusUpdater interface {
	RegisterStatusUpdater(fn func(up bool)) error
}

// NewLBStatusUpdater returns a new LbStatusUpdater.
func NewLBStatusUpdater(bh BalancerHandler, info *runtime.ServiceInfo) *LbStatusUpdater {
	return &LbStatusUpdater{
		BalancerHandler: bh,
		serviceInfo:     info,
	}
}

// LbStatusUpdater wraps a BalancerHandler and a ServiceInfo,
// so it can keep track of the status of a server in the ServiceInfo.
type LbStatusUpdater struct {
	BalancerHandler
	serviceInfo *runtime.ServiceInfo // can be nil
	updaters    []func(up bool)
}

// RegisterStatusUpdater adds fn to the list of hooks that are run when the
// status of the Balancer changes.
// Not thread safe.
func (lb *LbStatusUpdater) RegisterStatusUpdater(fn func(up bool)) error {
	// TODO: in theory, LbStatusUpdater should we aware of whether the healthcheck
	// was enabled in the config for this balancer. However, since at the moment it is
	// always wrapped in an emptyBackend, which already deals with that. If
	// LbStatusUpdater gets used in another context, we might have to deal with that
	// here too.
	lb.updaters = append(lb.updaters, fn)
	return nil
}

// RemoveServer removes the given server from the BalancerHandler,
// and updates the status of the server to "DOWN".
func (lb *LbStatusUpdater) RemoveServer(u *url.URL) error {
	// TODO(mpl): pass a context around, for better logging?
	upBefore := len(lb.BalancerHandler.Servers()) > 0
	err := lb.BalancerHandler.RemoveServer(u)
	if err != nil {
		return err
	}
	if lb.serviceInfo != nil {
		lb.serviceInfo.UpdateServerStatus(u.String(), serverDown)
	}

	if !upBefore {
		// we were already down, and we still are, no need to propagate.
		log.WithoutContext().Debugf("child %s now DOWN, but we were already DOWN, so no need to propagate.", u.String())
		return nil
	}
	if len(lb.BalancerHandler.Servers()) > 0 {
		// we were up, and we still are, no need to propagate
		log.WithoutContext().Debugf("child %s now DOWN, but we still are UP, so no need to propagate.", u.String())
		return nil
	}

	log.WithoutContext().Debugf("child %s now DOWN, and so are we, updating parent(s).", u.String())
	for _, fn := range lb.updaters {
		fn(false)
	}
	return nil
}

// UpsertServer adds the given server to the BalancerHandler,
// and updates the status of the server to "UP".
func (lb *LbStatusUpdater) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	upBefore := len(lb.BalancerHandler.Servers()) > 0
	err := lb.BalancerHandler.UpsertServer(u, options...)
	if err != nil {
		return err
	}
	if lb.serviceInfo != nil {
		lb.serviceInfo.UpdateServerStatus(u.String(), serverUp)
	}

	if upBefore {
		// we were up, and we still are, no need to propagate
		log.WithoutContext().Debugf("child %s now UP, but we were already UP, so no need to propagate.", u.String())
		return nil
	}

	log.WithoutContext().Debugf("child %s now UP, and so are we, updating parent(s).", u.String())
	for _, fn := range lb.updaters {
		fn(true)
	}
	return nil
}

// Balancers is a list of Balancers(s) that implements the Balancer interface.
type Balancers []Balancer

// Servers returns the servers url from all the BalancerHandler.
func (b Balancers) Servers() []*url.URL {
	var servers []*url.URL
	for _, lb := range b {
		servers = append(servers, lb.Servers()...)
	}

	return servers
}

// RemoveServer removes the given server from all the BalancerHandler,
// and updates the status of the server to "DOWN".
func (b Balancers) RemoveServer(u *url.URL) error {
	for _, lb := range b {
		if err := lb.RemoveServer(u); err != nil {
			return err
		}
	}
	return nil
}

// UpsertServer adds the given server to all the BalancerHandler,
// and updates the status of the server to "UP".
func (b Balancers) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	for _, lb := range b {
		if err := lb.UpsertServer(u, options...); err != nil {
			return err
		}
	}
	return nil
}
