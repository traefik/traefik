package healthcheck

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/go-kit/kit/metrics"
	"github.com/vulcand/oxy/roundrobin"
)

var singleton *HealthCheck
var once sync.Once

// GetHealthCheck returns the health check which is guaranteed to be a singleton.
func GetHealthCheck(metrics metricsRegistry) *HealthCheck {
	once.Do(func() {
		singleton = newHealthCheck(metrics)
	})
	return singleton
}

// Options are the public health check options.
type Options struct {
	Headers   map[string]string
	Path      string
	Port      int
	Transport http.RoundTripper
	Interval  time.Duration
	LB        LoadBalancer
}

func (opt Options) String() string {
	return fmt.Sprintf("[Headers: %v Path: %s Port: %d Interval: %s]", opt.Headers, opt.Path, opt.Port, opt.Interval)
}

// BackendHealthCheck HealthCheck configuration for a backend
type BackendHealthCheck struct {
	Options
	name           string
	disabledURLs   []*url.URL
	requestTimeout time.Duration
}

//HealthCheck struct
type HealthCheck struct {
	Backends map[string]*BackendHealthCheck
	metrics  metricsRegistry
	cancel   context.CancelFunc
}

// LoadBalancer includes functionality for load-balancing management.
type LoadBalancer interface {
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
	Servers() []*url.URL
}

func newHealthCheck(metrics metricsRegistry) *HealthCheck {
	return &HealthCheck{
		Backends: make(map[string]*BackendHealthCheck),
		metrics:  metrics,
	}
}

// metricsRegistry is a local interface in the healthcheck package, exposing only the required metrics
// necessary for the healthcheck package. This makes it easier for the tests.
type metricsRegistry interface {
	BackendServerUpGauge() metrics.Gauge
}

// NewBackendHealthCheck Instantiate a new BackendHealthCheck
func NewBackendHealthCheck(options Options, backendName string) *BackendHealthCheck {
	return &BackendHealthCheck{
		Options:        options,
		name:           backendName,
		requestTimeout: 5 * time.Second,
	}
}

//SetBackendsConfiguration set backends configuration
func (hc *HealthCheck) SetBackendsConfiguration(parentCtx context.Context, backends map[string]*BackendHealthCheck) {
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

func (hc *HealthCheck) execute(ctx context.Context, backend *BackendHealthCheck) {
	log.Debugf("Initial health check for backend: %q", backend.name)
	hc.checkBackend(backend)
	ticker := time.NewTicker(backend.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Debug("Stopping current health check goroutines of backend: %s", backend.name)
			return
		case <-ticker.C:
			log.Debugf("Refreshing health check for backend: %s", backend.name)
			hc.checkBackend(backend)
		}
	}
}

func (hc *HealthCheck) checkBackend(backend *BackendHealthCheck) {
	enabledURLs := backend.LB.Servers()
	var newDisabledURLs []*url.URL
	for _, url := range backend.disabledURLs {
		serverUpMetricValue := float64(0)
		if err := checkHealth(url, backend); err == nil {
			log.Warnf("Health check up: Returning to server list. Backend: %q URL: %q", backend.name, url.String())
			backend.LB.UpsertServer(url, roundrobin.Weight(1))
			serverUpMetricValue = 1
		} else {
			log.Warnf("Health check still failing. Backend: %q URL: %q Reason: %s", backend.name, url.String(), err)
			newDisabledURLs = append(newDisabledURLs, url)
		}
		labelValues := []string{"backend", backend.name, "url", url.String()}
		hc.metrics.BackendServerUpGauge().With(labelValues...).Set(serverUpMetricValue)
	}
	backend.disabledURLs = newDisabledURLs

	for _, url := range enabledURLs {
		serverUpMetricValue := float64(1)
		if err := checkHealth(url, backend); err != nil {
			log.Warnf("Health check failed: Remove from server list. Backend: %q URL: %q Reason: %s", backend.name, url.String(), err)
			backend.LB.RemoveServer(url)
			backend.disabledURLs = append(backend.disabledURLs, url)
			serverUpMetricValue = 0
		}
		labelValues := []string{"backend", backend.name, "url", url.String()}
		hc.metrics.BackendServerUpGauge().With(labelValues...).Set(serverUpMetricValue)
	}
}

func (backend *BackendHealthCheck) newRequest(serverURL *url.URL) (*http.Request, error) {
	if backend.Port == 0 {
		return http.NewRequest(http.MethodGet, serverURL.String()+backend.Path, nil)
	}

	// copy the url and add the port to the host
	u := &url.URL{}
	*u = *serverURL
	u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(backend.Port))
	u.Path = u.Path + backend.Path

	return http.NewRequest(http.MethodGet, u.String(), nil)
}

// this function adds additional httpheaders to http.request
func (backend *BackendHealthCheck) addHeaders(req *http.Request) *http.Request {
	for k, v := range backend.Options.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}
	return req
}

// checkHealth returns a nil error in case it was successful and otherwise
// a non-nil error with a meaningful description why the health check failed.
func checkHealth(serverURL *url.URL, backend *BackendHealthCheck) error {
	client := http.Client{
		Timeout:   backend.requestTimeout,
		Transport: backend.Options.Transport,
	}
	req, err := backend.newRequest(serverURL)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %s", err)
	}
	req = backend.addHeaders(req)

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
	}

	switch {
	case err != nil:
		return fmt.Errorf("HTTP request failed: %s", err)
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("received non-200 status code: %v", resp.StatusCode)
	}
	return nil
}
