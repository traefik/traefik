package healthcheck

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/vulcand/oxy/roundrobin"
)

var singleton *HealthCheck
var once sync.Once

// GetHealthCheck Get HealtchCheck Singleton
func GetHealthCheck() *HealthCheck {
	once.Do(func() {
		singleton = newHealthCheck()
	})
	return singleton
}

// BackendHealthCheck HealthCheck configuration for a backend
type BackendHealthCheck struct {
	Path         string
	Interval     time.Duration
	DisabledURLs []*url.URL
	lb           loadBalancer
}

var launch = false

//HealthCheck struct
type HealthCheck struct {
	Backends map[string]*BackendHealthCheck
	cancel   context.CancelFunc
}

type loadBalancer interface {
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
	Servers() []*url.URL
}

func newHealthCheck() *HealthCheck {
	return &HealthCheck{make(map[string]*BackendHealthCheck), nil}
}

// NewBackendHealthCheck Instantiate a new BackendHealthCheck
func NewBackendHealthCheck(URL string, interval time.Duration, lb loadBalancer) *BackendHealthCheck {
	return &BackendHealthCheck{URL, interval, nil, lb}
}

//SetBackendsConfiguration set backends configuration
func (hc *HealthCheck) SetBackendsConfiguration(parentCtx context.Context, backends map[string]*BackendHealthCheck) {
	hc.Backends = backends
	if hc.cancel != nil {
		hc.cancel()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	hc.cancel = cancel
	hc.execute(ctx)
}

func (hc *HealthCheck) execute(ctx context.Context) {
	for backendID, backend := range hc.Backends {
		currentBackend := backend
		currentBackendID := backendID
		safe.Go(func() {
			for {
				ticker := time.NewTicker(currentBackend.Interval)
				select {
				case <-ctx.Done():
					log.Debugf("Stopping all current Healthcheck goroutines")
					return
				case <-ticker.C:
					log.Debugf("Refreshing Healthcheck for currentBackend %s ", currentBackendID)
					enabledURLs := currentBackend.lb.Servers()
					var newDisabledURLs []*url.URL
					for _, url := range currentBackend.DisabledURLs {
						if checkHealth(url, currentBackend.Path) {
							log.Debugf("HealthCheck is up [%s]: Upsert in server list", url.String())
							currentBackend.lb.UpsertServer(url, roundrobin.Weight(1))
						} else {
							newDisabledURLs = append(newDisabledURLs, url)
						}
					}
					currentBackend.DisabledURLs = newDisabledURLs

					for _, url := range enabledURLs {
						if !checkHealth(url, currentBackend.Path) {
							log.Debugf("HealthCheck has failed [%s]: Remove from server list", url.String())
							currentBackend.lb.RemoveServer(url)
							currentBackend.DisabledURLs = append(currentBackend.DisabledURLs, url)
						}
					}

				}
			}
		})
	}
}

func checkHealth(serverURL *url.URL, path string) bool {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(serverURL.String() + path)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}
