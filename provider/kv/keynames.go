package kv

const (
	pathBackends                                = "/backends/"
	pathBackendCircuitBreakerExpression         = "/circuitbreaker/expression"
	pathBackendHealthCheckPath                  = "/healthcheck/path"
	pathBackendHealthCheckInterval              = "/healthcheck/interval"
	pathBackendLoadBalancerMethod               = "/loadbalancer/method"
	pathBackendLoadBalancerSticky               = "/loadbalancer/sticky"
	pathBackendLoadBalancerStickiness           = "/loadbalancer/stickiness"
	pathBackendLoadBalancerStickinessCookieName = "/loadbalancer/stickiness/cookiename"
	pathBackendServers                          = "/servers/"
	pathBackendServerURL                        = "/url"

	pathFrontends              = "/frontends/"
	pathFrontendBackend        = "/backend"
	pathFrontendPriority       = "/priority"
	pathFrontendPassHostHeader = "/passHostHeader"
	pathFrontendEntryPoints    = "/entrypoints"

	pathTags      = "/tags"
	pathAlias     = "/alias"
	pathSeparator = "/"
)
