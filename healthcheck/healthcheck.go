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

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/vulcand/oxy/roundrobin"
)

var singleton *HealthCheck
var once sync.Once

// GetHealthCheck returns the health check which is guaranteed to be a singleton.
func GetHealthCheck() *HealthCheck {
	once.Do(func() {
		singleton = newHealthCheck()
	})
	return singleton
}

// Options are the public health check options.
type Options struct {
	Path     string
	Port     int
	Interval time.Duration
	LB       LoadBalancer
}

func (opt Options) String() string {
	return fmt.Sprintf("[Path: %s Interval: %s]", opt.Path, opt.Interval)
}

// BackendHealthCheck HealthCheck configuration for a backend
type BackendHealthCheck struct {
	Options
	disabledURLs   []*url.URL
	requestTimeout time.Duration
}

//HealthCheck struct
type HealthCheck struct {
	Backends map[string]*BackendHealthCheck
	cancel   context.CancelFunc
}

// LoadBalancer includes functionality for load-balancing management.
type LoadBalancer interface {
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
	Servers() []*url.URL
}

func newHealthCheck() *HealthCheck {
	return &HealthCheck{
		Backends: make(map[string]*BackendHealthCheck),
	}
}

// NewBackendHealthCheck Instantiate a new BackendHealthCheck
func NewBackendHealthCheck(options Options) *BackendHealthCheck {
	return &BackendHealthCheck{
		Options:        options,
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

	for backendID, backend := range hc.Backends {
		currentBackendID := backendID
		currentBackend := backend
		safe.Go(func() {
			hc.execute(ctx, currentBackendID, currentBackend)
		})
	}
}

func (hc *HealthCheck) execute(ctx context.Context, backendID string, backend *BackendHealthCheck) {
	log.Debugf("Initial healthcheck for currentBackend %s ", backendID)
	checkBackend(backend)
	ticker := time.NewTicker(backend.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Debug("Stopping all current Healthcheck goroutines")
			return
		case <-ticker.C:
			log.Debugf("Refreshing healthcheck for currentBackend %s ", backendID)
			checkBackend(backend)
		}
	}
}

func checkBackend(currentBackend *BackendHealthCheck) {
	enabledURLs := currentBackend.LB.Servers()
	var newDisabledURLs []*url.URL
	for _, url := range currentBackend.disabledURLs {
		if checkHealth(url, currentBackend) {
			log.Debugf("HealthCheck is up [%s]: Upsert in server list", url.String())
			currentBackend.LB.UpsertServer(url, roundrobin.Weight(1))
		} else {
			log.Warnf("HealthCheck is still failing [%s]", url.String())
			newDisabledURLs = append(newDisabledURLs, url)
		}
	}
	currentBackend.disabledURLs = newDisabledURLs

	for _, url := range enabledURLs {
		if !checkHealth(url, currentBackend) {
			log.Warnf("HealthCheck has failed [%s]: Remove from server list", url.String())
			currentBackend.LB.RemoveServer(url)
			currentBackend.disabledURLs = append(currentBackend.disabledURLs, url)
		}
	}
}

func (backend *BackendHealthCheck) newRequest(serverURL *url.URL) (*http.Request, error) {
	if backend.Options.Port == 0 {
		return http.NewRequest("GET", serverURL.String()+backend.Path, nil)
	}

	// copy the url and add the port to the host
	u := &url.URL{}
	*u = *serverURL
	u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(backend.Options.Port))
	u.Path = u.Path + backend.Path

	return http.NewRequest("GET", u.String(), nil)
}

func checkHealth(serverURL *url.URL, backend *BackendHealthCheck) bool {
	client := http.Client{
		Timeout: backend.requestTimeout,
	}
	req, err := backend.newRequest(serverURL)
	if err != nil {
		log.Errorf("Failed to create HTTP request [%s] for healthcheck: %s", serverURL, err)
		return false
	}

	resp, err := client.Do(req)

	if err == nil {
		defer resp.Body.Close()
	}
	return err == nil && resp.StatusCode == 200
}
