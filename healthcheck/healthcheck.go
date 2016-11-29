package healthcheck

import (
	"github.com/containous/traefik/log"
	"github.com/vulcand/oxy/roundrobin"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var singleton *HealthCheck
var once sync.Once

// GetHealthCheck Get HealtchCheck Singleton
func GetHealthCheck() *HealthCheck {
	once.Do(func() {
		singleton = newHealthCheck()
		singleton.execute()
	})
	return singleton
}

// BackendHealthCheck HealthCheck configuration for a backend
type BackendHealthCheck struct {
	URL          string
	DisabledURLs []*url.URL
	lb           loadBalancer
}

var launch = false

//HealthCheck struct
type HealthCheck struct {
	Backends map[string]*BackendHealthCheck
}

type loadBalancer interface {
	RemoveServer(u *url.URL) error
	UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error
	Servers() []*url.URL
}

func newHealthCheck() *HealthCheck {
	return &HealthCheck{make(map[string]*BackendHealthCheck)}
}

// NewBackendHealthCheck Instantiate a new BackendHealthCheck
func NewBackendHealthCheck(URL string, lb loadBalancer) *BackendHealthCheck {
	return &BackendHealthCheck{URL, nil, lb}
}

//SetBackendsConfiguration set backends configuration
func (hc *HealthCheck) SetBackendsConfiguration(backends map[string]*BackendHealthCheck) {
	hc.Backends = backends
}

func (hc *HealthCheck) execute() {
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for t := range ticker.C {
			for backendID, backend := range hc.Backends {
				log.Debugf("Refreshing Healthcheck %s for backend %s ", t.String(), backendID)
				enabledURLs := backend.lb.Servers()
				var newDisabledURLs []*url.URL
				for _, url := range backend.DisabledURLs {
					if testHealth(url, backend.URL) {
						log.Debugf("Upsert %s", url.String())
						backend.lb.UpsertServer(url, roundrobin.Weight(1))
					} else {
						newDisabledURLs = append(newDisabledURLs, url)
					}
				}
				backend.DisabledURLs = newDisabledURLs

				for _, url := range enabledURLs {
					if !testHealth(url, backend.URL) {
						log.Debugf("Remove %s", url.String())
						backend.lb.RemoveServer(url)
						backend.DisabledURLs = append(backend.DisabledURLs, url)
					}
				}
			}
		}
	}()
}

func testHealth(serverURL *url.URL, checkURL string) bool {
	resp, err := http.Get(serverURL.String() + checkURL)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}
