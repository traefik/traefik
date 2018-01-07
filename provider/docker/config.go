package docker

import (
	"math"
	"strconv"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfiguration(containersInspected []dockerData) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getDomain":        getFuncStringLabel(label.TraefikDomain, p.Domain),
		"getSubDomain":     getSubDomain,
		"isBackendLBSwarm": isBackendLBSwarm, // FIXME dead ?

		// Backend functions
		"getIPAddress":      p.getIPAddress,
		"getPort":           getPort,
		"getWeight":         getFuncIntLabel(label.TraefikWeight, label.DefaultWeightInt),
		"getProtocol":       getFuncStringLabel(label.TraefikProtocol, label.DefaultProtocol),
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,

		// TODO Deprecated [breaking]
		"hasCircuitBreakerLabel": hasFunc(label.TraefikBackendCircuitBreakerExpression),
		// TODO Deprecated [breaking]
		"getCircuitBreakerExpression": getFuncStringLabel(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		// TODO Deprecated [breaking]
		"hasLoadBalancerLabel": hasLoadBalancerLabel,
		// TODO Deprecated [breaking]
		"getLoadBalancerMethod": getFuncStringLabel(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		// TODO Deprecated [breaking]
		"hasMaxConnLabels": hasMaxConnLabels,
		// TODO Deprecated [breaking]
		"getMaxConnAmount": getFuncInt64Label(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		// TODO Deprecated [breaking]
		"getMaxConnExtractorFunc": getFuncStringLabel(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		// TODO Deprecated [breaking]
		"getSticky": getSticky,
		// TODO Deprecated [breaking]
		"hasStickinessLabel": hasFunc(label.TraefikBackendLoadBalancerStickiness),
		// TODO Deprecated [breaking]
		"getStickinessCookieName": getFuncStringLabel(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),

		// Frontend functions
		"getBackend":              getBackendName, // TODO Deprecated [breaking] replaced by getBackendName
		"getBackendName":          getBackendName,
		"getPriority":             getFuncIntLabel(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
		"getPassHostHeader":       getFuncBoolLabel(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":          getFuncBoolLabel(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":          getFuncSliceStringLabel(label.TraefikFrontendEntryPoints),
		"getBasicAuth":            getFuncSliceStringLabel(label.TraefikFrontendAuthBasic),
		"getWhitelistSourceRange": getFuncSliceStringLabel(label.TraefikFrontendWhitelistSourceRange),
		"getFrontendRule":         p.getFrontendRule,

		"getRedirect":   getRedirect,
		"getErrorPages": getErrorPages,
		"getRateLimit":  getRateLimit,
		"getHeaders":    getHeaders,

		// Services
		"hasServices":           hasServices,
		"getServiceNames":       getServiceNames,
		"getServiceBackend":     getServiceBackendName, // TODO Deprecated [breaking] replaced by getServiceBackendName
		"getServiceBackendName": getServiceBackendName,
		// Services - Backend server functions
		"getServicePort":     getServicePort,
		"getServiceProtocol": getFuncServiceStringLabel(label.SuffixProtocol, label.DefaultProtocol),
		"getServiceWeight":   getFuncServiceStringLabel(label.SuffixWeight, label.DefaultWeight),
		// Services - Frontend functions
		"getServiceEntryPoints":          getFuncServiceSliceStringLabel(label.SuffixFrontendEntryPoints),
		"getServiceWhitelistSourceRange": getFuncServiceSliceStringLabel(label.SuffixFrontendWhitelistSourceRange),
		"getServiceBasicAuth":            getFuncServiceSliceStringLabel(label.SuffixFrontendAuthBasic),
		"getServiceFrontendRule":         p.getServiceFrontendRule,
		"getServicePassHostHeader":       getFuncServiceBoolLabel(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getServicePassTLSCert":          getFuncServiceBoolLabel(label.SuffixFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getServicePriority":             getFuncServiceIntLabel(label.SuffixFrontendPriority, label.DefaultFrontendPriorityInt),

		"getServiceRedirect":   getServiceRedirect,
		"getServiceErrorPages": getServiceErrorPages,
		"getServiceRateLimit":  getServiceRateLimit,
		"getServiceHeaders":    getServiceHeaders,
	}
	// filter containers
	filteredContainers := fun.Filter(func(container dockerData) bool {
		return p.containerFilter(container)
	}, containersInspected).([]dockerData)

	frontends := map[string][]dockerData{}
	backends := map[string]dockerData{}
	servers := map[string][]dockerData{}
	serviceNames := make(map[string]struct{})
	for idx, container := range filteredContainers {
		if _, exists := serviceNames[container.ServiceName]; !exists {
			frontendName := p.getFrontendName(container, idx)
			frontends[frontendName] = append(frontends[frontendName], container)
			if len(container.ServiceName) > 0 {
				serviceNames[container.ServiceName] = struct{}{}
			}
		}
		backendName := getBackendName(container)
		backends[backendName] = container
		servers[backendName] = append(servers[backendName], container)
	}

	templateObjects := struct {
		Containers []dockerData
		Frontends  map[string][]dockerData
		Backends   map[string]dockerData
		Servers    map[string][]dockerData
		Domain     string
	}{
		Containers: filteredContainers,
		Frontends:  frontends,
		Backends:   backends,
		Servers:    servers,
		Domain:     p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/docker.tmpl", DockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func (p Provider) containerFilter(container dockerData) bool {
	if !label.IsEnabled(container.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	var err error
	portLabel := "traefik.port label"
	if hasServices(container) {
		portLabel = "traefik.<serviceName>.port or " + portLabel + "s"
		err = checkServiceLabelPort(container)
	} else {
		_, err = strconv.Atoi(container.Labels[label.TraefikPort])
	}
	if len(container.NetworkSettings.Ports) == 0 && err != nil {
		log.Debugf("Filtering container without port and no %s %s : %s", portLabel, container.Name, err.Error())
		return false
	}

	constraintTags := label.SplitAndTrimString(container.Labels[label.TraefikTags], ",")
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Container %v pruned by '%v' constraint", container.Name, failingConstraint.String())
		}
		return false
	}

	if container.Health != "" && container.Health != "healthy" {
		log.Debugf("Filtering unhealthy or starting container %s", container.Name)
		return false
	}

	if len(p.getFrontendRule(container)) == 0 {
		log.Debugf("Filtering container with empty frontend rule %s", container.Name)
		return false
	}

	return true
}
