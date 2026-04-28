package ingressnginx

import "github.com/traefik/traefik/v3/pkg/config/dynamic"

// unavailableServiceName is the name of a Traefik service returning a 503 Service Unavailable.
// It is used by routers to be aligned with ingress-nginx behavior, for example when there is a
// configuration error in the custom-headers middleware.
const unavailableServiceName = "unavailable-service"

// ensureUnavailableService registers the unavailable service in conf if not already present.
func ensureUnavailableService(conf *dynamic.Configuration) {
	if _, ok := conf.HTTP.Services[unavailableServiceName]; ok {
		return
	}
	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()
	conf.HTTP.Services[unavailableServiceName] = &dynamic.Service{LoadBalancer: lb}
}
