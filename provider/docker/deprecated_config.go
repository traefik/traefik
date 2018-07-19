package docker

import (
	"context"
	"math"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// Deprecated
func (p *Provider) buildConfigurationV1(containersInspected []dockerData) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getDomain":        getFuncStringLabelV1(label.TraefikDomain, p.Domain),
		"getSubDomain":     getSubDomain,
		"isBackendLBSwarm": isBackendLBSwarm,

		// Backend functions
		"getIPAddress": p.getIPAddressV1,
		"getPort":      getPortV1,
		"getWeight":    getFuncIntLabelV1(label.TraefikWeight, label.DefaultWeight),
		"getProtocol":  getFuncStringLabelV1(label.TraefikProtocol, label.DefaultProtocol),

		"hasCircuitBreakerLabel":      hasFuncV1(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncStringLabelV1(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasLoadBalancerLabel":        hasLoadBalancerLabelV1,
		"getLoadBalancerMethod":       getFuncStringLabelV1(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"hasMaxConnLabels":            hasMaxConnLabelsV1,
		"getMaxConnAmount":            getFuncInt64LabelV1(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		"getMaxConnExtractorFunc":     getFuncStringLabelV1(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		"getSticky":                   getStickyV1,
		"hasStickinessLabel":          hasFuncV1(label.TraefikBackendLoadBalancerStickiness),
		"getStickinessCookieName":     getFuncStringLabelV1(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),

		// Frontend functions
		"getBackend":              getBackendNameV1,
		"getBackendName":          getBackendNameV1,
		"getPriority":             getFuncIntLabelV1(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":       getFuncBoolLabelV1(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":          getFuncBoolLabelV1(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":          getFuncSliceStringLabelV1(label.TraefikFrontendEntryPoints),
		"getBasicAuth":            getFuncSliceStringLabelV1(label.TraefikFrontendAuthBasic),
		"getWhitelistSourceRange": getFuncSliceStringLabelV1(label.TraefikFrontendWhitelistSourceRange),
		"getFrontendRule":         p.getFrontendRuleV1,
		"hasRedirect":             hasRedirectV1,
		"getRedirectEntryPoint":   getFuncStringLabelV1(label.TraefikFrontendRedirectEntryPoint, ""),
		"getRedirectRegex":        getFuncStringLabelV1(label.TraefikFrontendRedirectRegex, ""),
		"getRedirectReplacement":  getFuncStringLabelV1(label.TraefikFrontendRedirectReplacement, ""),

		"hasHeaders":                        hasHeadersV1,
		"hasRequestHeaders":                 hasLabelV1(label.TraefikFrontendRequestHeaders),
		"getRequestHeaders":                 getFuncMapLabelV1(label.TraefikFrontendRequestHeaders),
		"hasResponseHeaders":                hasLabelV1(label.TraefikFrontendResponseHeaders),
		"getResponseHeaders":                getFuncMapLabelV1(label.TraefikFrontendResponseHeaders),
		"hasAllowedHostsHeaders":            hasLabelV1(label.TraefikFrontendAllowedHosts),
		"getAllowedHostsHeaders":            getFuncSliceStringLabelV1(label.TraefikFrontendAllowedHosts),
		"hasHostsProxyHeaders":              hasLabelV1(label.TraefikFrontendHostsProxyHeaders),
		"getHostsProxyHeaders":              getFuncSliceStringLabelV1(label.TraefikFrontendHostsProxyHeaders),
		"hasSSLRedirectHeaders":             hasLabelV1(label.TraefikFrontendSSLRedirect),
		"getSSLRedirectHeaders":             getFuncBoolLabelV1(label.TraefikFrontendSSLRedirect, false),
		"hasSSLTemporaryRedirectHeaders":    hasLabelV1(label.TraefikFrontendSSLTemporaryRedirect),
		"getSSLTemporaryRedirectHeaders":    getFuncBoolLabelV1(label.TraefikFrontendSSLTemporaryRedirect, false),
		"hasSSLHostHeaders":                 hasLabelV1(label.TraefikFrontendSSLHost),
		"getSSLHostHeaders":                 getFuncStringLabelV1(label.TraefikFrontendSSLHost, ""),
		"hasSSLProxyHeaders":                hasLabelV1(label.TraefikFrontendSSLProxyHeaders),
		"getSSLProxyHeaders":                getFuncMapLabelV1(label.TraefikFrontendSSLProxyHeaders),
		"hasSTSSecondsHeaders":              hasLabelV1(label.TraefikFrontendSTSSeconds),
		"getSTSSecondsHeaders":              getFuncInt64LabelV1(label.TraefikFrontendSTSSeconds, 0),
		"hasSTSIncludeSubdomainsHeaders":    hasLabelV1(label.TraefikFrontendSTSIncludeSubdomains),
		"getSTSIncludeSubdomainsHeaders":    getFuncBoolLabelV1(label.TraefikFrontendSTSIncludeSubdomains, false),
		"hasSTSPreloadHeaders":              hasLabelV1(label.TraefikFrontendSTSPreload),
		"getSTSPreloadHeaders":              getFuncBoolLabelV1(label.TraefikFrontendSTSPreload, false),
		"hasForceSTSHeaderHeaders":          hasLabelV1(label.TraefikFrontendForceSTSHeader),
		"getForceSTSHeaderHeaders":          getFuncBoolLabelV1(label.TraefikFrontendForceSTSHeader, false),
		"hasFrameDenyHeaders":               hasLabelV1(label.TraefikFrontendFrameDeny),
		"getFrameDenyHeaders":               getFuncBoolLabelV1(label.TraefikFrontendFrameDeny, false),
		"hasCustomFrameOptionsValueHeaders": hasLabelV1(label.TraefikFrontendCustomFrameOptionsValue),
		"getCustomFrameOptionsValueHeaders": getFuncStringLabelV1(label.TraefikFrontendCustomFrameOptionsValue, ""),
		"hasContentTypeNosniffHeaders":      hasLabelV1(label.TraefikFrontendContentTypeNosniff),
		"getContentTypeNosniffHeaders":      getFuncBoolLabelV1(label.TraefikFrontendContentTypeNosniff, false),
		"hasBrowserXSSFilterHeaders":        hasLabelV1(label.TraefikFrontendBrowserXSSFilter),
		"getBrowserXSSFilterHeaders":        getFuncBoolLabelV1(label.TraefikFrontendBrowserXSSFilter, false),
		"hasContentSecurityPolicyHeaders":   hasLabelV1(label.TraefikFrontendContentSecurityPolicy),
		"getContentSecurityPolicyHeaders":   getFuncStringLabelV1(label.TraefikFrontendContentSecurityPolicy, ""),
		"hasPublicKeyHeaders":               hasLabelV1(label.TraefikFrontendPublicKey),
		"getPublicKeyHeaders":               getFuncStringLabelV1(label.TraefikFrontendPublicKey, ""),
		"hasReferrerPolicyHeaders":          hasLabelV1(label.TraefikFrontendReferrerPolicy),
		"getReferrerPolicyHeaders":          getFuncStringLabelV1(label.TraefikFrontendReferrerPolicy, ""),
		"hasIsDevelopmentHeaders":           hasLabelV1(label.TraefikFrontendIsDevelopment),
		"getIsDevelopmentHeaders":           getFuncBoolLabelV1(label.TraefikFrontendIsDevelopment, false),

		// Services
		"hasServices":           hasServicesV1,
		"getServiceNames":       getServiceNamesV1,
		"getServiceBackend":     getServiceBackendNameV1,
		"getServiceBackendName": getServiceBackendNameV1,
		// Services - Backend server functions
		"getServicePort":     getServicePortV1,
		"getServiceProtocol": getFuncServiceStringLabelV1(label.SuffixProtocol, label.DefaultProtocol),
		"getServiceWeight":   getFuncServiceIntLabelV1(label.SuffixWeight, label.DefaultWeight),
		// Services - Frontend functions
		"getServiceEntryPoints":          getFuncServiceSliceStringLabelV1(label.SuffixFrontendEntryPoints),
		"getServiceWhitelistSourceRange": getFuncServiceSliceStringLabelV1(label.SuffixFrontendWhiteListSourceRange),
		"getServiceBasicAuth":            getFuncServiceSliceStringLabelV1(label.SuffixFrontendAuthBasic),
		"getServiceFrontendRule":         p.getServiceFrontendRuleV1,
		"getServicePassHostHeader":       getFuncServiceBoolLabelV1(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getServicePassTLSCert":          getFuncServiceBoolLabelV1(label.SuffixFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getServicePriority":             getFuncServiceIntLabelV1(label.SuffixFrontendPriority, label.DefaultFrontendPriority),
		"hasServiceRedirect":             hasServiceRedirectV1,
		"getServiceRedirectEntryPoint":   getFuncServiceStringLabelV1(label.SuffixFrontendRedirectEntryPoint, ""),
		"getServiceRedirectReplacement":  getFuncServiceStringLabelV1(label.SuffixFrontendRedirectReplacement, ""),
		"getServiceRedirectRegex":        getFuncServiceStringLabelV1(label.SuffixFrontendRedirectRegex, ""),
	}

	// filter containers
	filteredContainers := fun.Filter(func(container dockerData) bool {
		return p.containerFilterV1(container)
	}, containersInspected).([]dockerData)

	frontends := map[string][]dockerData{}
	backends := map[string]dockerData{}
	servers := map[string][]dockerData{}
	serviceNames := make(map[string]struct{})
	for idx, container := range filteredContainers {

		serviceNamesKey := getServiceNameKey(container, p.SwarmMode, "")

		if _, exists := serviceNames[serviceNamesKey]; !exists {
			frontendName := p.getFrontendNameV1(container, idx)
			frontends[frontendName] = append(frontends[frontendName], container)
			if len(serviceNamesKey) > 0 {
				serviceNames[serviceNamesKey] = struct{}{}
			}
		}
		backendName := getBackendNameV1(container)
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

	configuration, err := p.GetConfiguration("templates/docker-v1.tmpl", DockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

// Deprecated
func (p Provider) containerFilterV1(container dockerData) bool {
	if !label.IsEnabled(container.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	var err error
	portLabel := "traefik.port label"
	if hasServicesV1(container) {
		portLabel = "traefik.<serviceName>.port or " + portLabel + "s"
		err = checkServiceLabelPortV1(container)
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

	if len(p.getFrontendRuleV1(container)) == 0 {
		log.Debugf("Filtering container with empty frontend rule %s", container.Name)
		return false
	}

	return true
}

func (p Provider) getIPAddressV1(container dockerData) string {
	if value := label.GetStringValue(container.Labels, labelDockerNetwork, p.Network); value != "" {
		networkSettings := container.NetworkSettings
		if networkSettings.Networks != nil {
			network := networkSettings.Networks[value]
			if network != nil {
				return network.Addr
			}

			log.Warnf("Could not find network named '%s' for container '%s'! Maybe you're missing the project's prefix in the label? Defaulting to first available network.", value, container.Name)
		}
	}

	if container.NetworkSettings.NetworkMode.IsHost() {
		if container.Node != nil {
			if container.Node.IPAddress != "" {
				return container.Node.IPAddress
			}
		}
		return "127.0.0.1"
	}

	if container.NetworkSettings.NetworkMode.IsContainer() {
		dockerClient, err := p.createClient()
		if err != nil {
			log.Warnf("Unable to get IP address for container %s, error: %s", container.Name, err)
			return ""
		}

		connectedContainer := container.NetworkSettings.NetworkMode.ConnectedContainer()
		containerInspected, err := dockerClient.ContainerInspect(context.Background(), connectedContainer)
		if err != nil {
			log.Warnf("Unable to get IP address for container %s : Failed to inspect container ID %s, error: %s", container.Name, connectedContainer, err)
			return ""
		}
		return p.getIPAddress(parseContainer(containerInspected))
	}

	if p.UseBindPortIP {
		port := getPortV1(container)
		for netPort, portBindings := range container.NetworkSettings.Ports {
			if strings.EqualFold(string(netPort), port+"/TCP") || strings.EqualFold(string(netPort), port+"/UDP") {
				for _, p := range portBindings {
					return p.HostIP
				}
			}
		}
	}

	for _, network := range container.NetworkSettings.Networks {
		return network.Addr
	}

	log.Warnf("Unable to find the IP address for the container %q.", container.Name)
	return ""
}
