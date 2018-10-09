package main

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/old/log"
	"github.com/containous/traefik/old/types"
	"github.com/sirupsen/logrus"
)

var oldvalue = `
[backends]
  [backends.backend1]
    [backends.backend1.servers.server1]
    url = "http://127.0.0.1:9010"
    weight = 1
  [backends.backend2]
    [backends.backend2.servers.server1]
    url = "http://127.0.0.1:9020"
    weight = 1

[frontends]
  [frontends.frontend1]
  backend = "backend1"
    [frontends.frontend1.routes.test_1]
    rule = "Host:snitest.com"
  [frontends.frontend2]
  backend = "backend2"
    [frontends.frontend2.routes.test_2]
    rule = "Host:snitest.org"

`

// Temporary utility to convert dynamic conf v1 to v2
func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	oldconfig := &types.Configuration{}
	toml.Decode(oldvalue, oldconfig)

	newconfig := config.Configuration{
		Routers:     make(map[string]*config.Router),
		Middlewares: make(map[string]*config.Middleware),
		Services:    make(map[string]*config.Service),
	}

	for frontendName, frontend := range oldconfig.Frontends {
		newconfig.Routers[replaceFrontend(frontendName)] = convertFrontend(frontend)
		if frontend.PassHostHeader {
			log.Warn("ignore PassHostHeader")
		}
	}

	for backendName, backend := range oldconfig.Backends {
		newconfig.Services[replaceBackend(backendName)] = convertBackend(backend)
	}

	encoder := toml.NewEncoder(os.Stdout)
	encoder.Encode(newconfig)
}

func replaceBackend(name string) string {
	return strings.Replace(name, "backend", "service", -1)
}

func replaceFrontend(name string) string {
	return strings.Replace(name, "frontend", "router", -1)
}

func convertFrontend(frontend *types.Frontend) *config.Router {
	router := &config.Router{
		EntryPoints: frontend.EntryPoints,
		Middlewares: nil,
		Service:     replaceBackend(frontend.Backend),
		Priority:    frontend.Priority,
	}

	if len(frontend.Routes) > 1 {
		log.Fatal("Multiple routes")
	}

	for _, route := range frontend.Routes {
		router.Rule = route.Rule
	}

	return router
}

func convertBackend(backend *types.Backend) *config.Service {
	service := &config.Service{
		LoadBalancer: &config.LoadBalancerService{
			Stickiness:     nil,
			Servers:        nil,
			Method:         "",
			HealthCheck:    nil,
			PassHostHeader: false,
		},
	}

	if backend.Buffering != nil {
		log.Warn("Buffering not implemented")
	}

	if backend.CircuitBreaker != nil {
		log.Warn("CircuitBreaker not implemented")
	}

	if backend.MaxConn != nil {
		log.Warn("MaxConn not implemented")
	}

	for _, oldserver := range backend.Servers {
		service.LoadBalancer.Servers = append(service.LoadBalancer.Servers, config.Server{
			URL:    oldserver.URL,
			Weight: oldserver.Weight,
		})
	}

	if backend.LoadBalancer != nil {
		service.LoadBalancer.Method = backend.LoadBalancer.Method
		if backend.LoadBalancer.Stickiness != nil {
			service.LoadBalancer.Stickiness = &config.Stickiness{
				CookieName: backend.LoadBalancer.Stickiness.CookieName,
			}
		}

		if backend.HealthCheck != nil {
			service.LoadBalancer.HealthCheck = &config.HealthCheck{
				Scheme:   backend.HealthCheck.Scheme,
				Path:     backend.HealthCheck.Path,
				Port:     backend.HealthCheck.Port,
				Interval: backend.HealthCheck.Interval,
				Timeout:  backend.HealthCheck.Timeout,
				Hostname: backend.HealthCheck.Hostname,
				Headers:  backend.HealthCheck.Headers,
			}
		}

	}

	return service
}
