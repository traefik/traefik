package docker

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/docker/go-connections/nat"
)

const (
	labelDockerNetwork            = "traefik.docker.network"
	labelBackendLoadBalancerSwarm = "traefik.backend.loadbalancer.swarm"
	labelDockerComposeProject     = "com.docker.compose.project"
	labelDockerComposeService     = "com.docker.compose.service"
)

// Specific functions

func (p Provider) getFrontendName(container dockerData, idx int) string {
	return provider.Normalize(p.getFrontendRule(container) + "-" + strconv.Itoa(idx))
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
func (p Provider) getFrontendRule(container dockerData) string {
	if value := label.GetStringValue(container.Labels, label.TraefikFrontendRule, ""); len(value) != 0 {
		return value
	}

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		return "Host:" + getSubDomain(values[labelDockerComposeService]+"."+values[labelDockerComposeProject]) + "." + p.Domain
	}

	if len(p.Domain) > 0 {
		return "Host:" + getSubDomain(container.ServiceName) + "." + p.Domain
	}

	return ""
}

func (p Provider) getIPAddress(container dockerData) string {

	if value := label.GetStringValue(container.Labels, labelDockerNetwork, ""); value != "" {
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
		port := getPort(container)
		for netPort, portBindings := range container.NetworkSettings.Ports {
			if string(netPort) == port+"/TCP" || string(netPort) == port+"/UDP" {
				for _, p := range portBindings {
					return p.HostIP
				}
			}
		}
	}

	for _, network := range container.NetworkSettings.Networks {
		return network.Addr
	}
	return ""
}

func getBackendName(container dockerData) string {
	if value := label.GetStringValue(container.Labels, label.TraefikBackend, ""); len(value) != 0 {
		return provider.Normalize(value)
	}

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		return provider.Normalize(values[labelDockerComposeService] + "_" + values[labelDockerComposeProject])
	}

	return provider.Normalize(container.ServiceName)
}

func getPort(container dockerData) string {
	if value := label.GetStringValue(container.Labels, label.TraefikPort, ""); len(value) != 0 {
		return value
	}

	// See iteration order in https://blog.golang.org/go-maps-in-action
	var ports []nat.Port
	for port := range container.NetworkSettings.Ports {
		ports = append(ports, port)
	}

	less := func(i, j nat.Port) bool {
		return i.Int() < j.Int()
	}
	nat.Sort(ports, less)

	if len(ports) > 0 {
		min := ports[0]
		return min.Port()
	}

	return ""
}

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func getSubDomain(name string) string {
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func isBackendLBSwarm(container dockerData) bool {
	return label.GetBoolValue(container.Labels, labelBackendLoadBalancerSwarm, false)
}

func getMaxConn(container dockerData) *types.MaxConn {
	amount := label.GetInt64Value(container.Labels, label.TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := label.GetStringValue(container.Labels, label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func getLoadBalancer(container dockerData) *types.LoadBalancer {
	if !label.HasPrefix(container.Labels, label.TraefikBackendLoadBalancer) {
		return nil
	}

	method := label.GetStringValue(container.Labels, label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: getSticky(container),
	}

	if label.GetBoolValue(container.Labels, label.TraefikBackendLoadBalancerStickiness, false) {
		cookieName := label.GetStringValue(container.Labels, label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getSticky(container dockerData) bool {
	if label.Has(container.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}

	return label.GetBoolValue(container.Labels, label.TraefikBackendLoadBalancerSticky, false)
}

func getCircuitBreaker(container dockerData) *types.CircuitBreaker {
	circuitBreaker := label.GetStringValue(container.Labels, label.TraefikBackendCircuitBreakerExpression, "")
	if len(circuitBreaker) == 0 {
		return nil
	}
	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func getHealthCheck(container dockerData) *types.HealthCheck {
	path := label.GetStringValue(container.Labels, label.TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	port := label.GetIntValue(container.Labels, label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort)
	interval := label.GetStringValue(container.Labels, label.TraefikBackendHealthCheckInterval, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func getBuffering(container dockerData) *types.Buffering {
	if !label.HasPrefix(container.Labels, label.TraefikBackendBuffering) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  label.GetInt64Value(container.Labels, label.TraefikBackendBufferingMaxRequestBodyBytes, 0),
		MaxResponseBodyBytes: label.GetInt64Value(container.Labels, label.TraefikBackendBufferingMaxResponseBodyBytes, 0),
		MemRequestBodyBytes:  label.GetInt64Value(container.Labels, label.TraefikBackendBufferingMemRequestBodyBytes, 0),
		MemResponseBodyBytes: label.GetInt64Value(container.Labels, label.TraefikBackendBufferingMemResponseBodyBytes, 0),
		RetryExpression:      label.GetStringValue(container.Labels, label.TraefikBackendBufferingRetryExpression, ""),
	}
}

func getRedirect(container dockerData) *types.Redirect {
	if label.Has(container.Labels, label.TraefikFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: label.GetStringValue(container.Labels, label.TraefikFrontendRedirectEntryPoint, ""),
		}
	}

	if label.Has(container.Labels, label.TraefikFrontendRedirectRegex) &&
		label.Has(container.Labels, label.TraefikFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       label.GetStringValue(container.Labels, label.TraefikFrontendRedirectRegex, ""),
			Replacement: label.GetStringValue(container.Labels, label.TraefikFrontendRedirectReplacement, ""),
		}
	}

	return nil
}

func getErrorPages(container dockerData) map[string]*types.ErrorPage {
	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(container.Labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimit(container dockerData) *types.RateLimit {
	extractorFunc := label.GetStringValue(container.Labels, label.TraefikFrontendRateLimitExtractorFunc, "")
	if len(extractorFunc) == 0 {
		return nil
	}

	prefix := label.Prefix + label.BaseFrontendRateLimit
	limits := label.ParseRateSets(container.Labels, prefix, label.RegexpFrontendRateLimit)

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func getHeaders(container dockerData) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    label.GetMapValue(container.Labels, label.TraefikFrontendRequestHeaders),
		CustomResponseHeaders:   label.GetMapValue(container.Labels, label.TraefikFrontendResponseHeaders),
		SSLProxyHeaders:         label.GetMapValue(container.Labels, label.TraefikFrontendSSLProxyHeaders),
		AllowedHosts:            label.GetSliceStringValue(container.Labels, label.TraefikFrontendAllowedHosts),
		HostsProxyHeaders:       label.GetSliceStringValue(container.Labels, label.TraefikFrontendHostsProxyHeaders),
		STSSeconds:              label.GetInt64Value(container.Labels, label.TraefikFrontendSTSSeconds, 0),
		SSLRedirect:             label.GetBoolValue(container.Labels, label.TraefikFrontendSSLRedirect, false),
		SSLTemporaryRedirect:    label.GetBoolValue(container.Labels, label.TraefikFrontendSSLTemporaryRedirect, false),
		STSIncludeSubdomains:    label.GetBoolValue(container.Labels, label.TraefikFrontendSTSIncludeSubdomains, false),
		STSPreload:              label.GetBoolValue(container.Labels, label.TraefikFrontendSTSPreload, false),
		ForceSTSHeader:          label.GetBoolValue(container.Labels, label.TraefikFrontendForceSTSHeader, false),
		FrameDeny:               label.GetBoolValue(container.Labels, label.TraefikFrontendFrameDeny, false),
		ContentTypeNosniff:      label.GetBoolValue(container.Labels, label.TraefikFrontendContentTypeNosniff, false),
		BrowserXSSFilter:        label.GetBoolValue(container.Labels, label.TraefikFrontendBrowserXSSFilter, false),
		IsDevelopment:           label.GetBoolValue(container.Labels, label.TraefikFrontendIsDevelopment, false),
		SSLHost:                 label.GetStringValue(container.Labels, label.TraefikFrontendSSLHost, ""),
		CustomFrameOptionsValue: label.GetStringValue(container.Labels, label.TraefikFrontendCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   label.GetStringValue(container.Labels, label.TraefikFrontendContentSecurityPolicy, ""),
		PublicKey:               label.GetStringValue(container.Labels, label.TraefikFrontendPublicKey, ""),
		ReferrerPolicy:          label.GetStringValue(container.Labels, label.TraefikFrontendReferrerPolicy, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// Deprecated
func hasLoadBalancerLabel(container dockerData) bool {
	method := label.Has(container.Labels, label.TraefikBackendLoadBalancerMethod)
	sticky := label.Has(container.Labels, label.TraefikBackendLoadBalancerSticky)
	stickiness := label.Has(container.Labels, label.TraefikBackendLoadBalancerStickiness)
	cookieName := label.Has(container.Labels, label.TraefikBackendLoadBalancerStickinessCookieName)
	return method || sticky || stickiness || cookieName
}

// Deprecated
func hasMaxConnLabels(container dockerData) bool {
	mca := label.Has(container.Labels, label.TraefikBackendMaxConnAmount)
	mcef := label.Has(container.Labels, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

// Label functions

func getFuncStringLabel(labelName string, defaultValue string) func(container dockerData) string {
	return func(container dockerData) string {
		return label.GetStringValue(container.Labels, labelName, defaultValue)
	}
}

func getFuncBoolLabel(labelName string, defaultValue bool) func(container dockerData) bool {
	return func(container dockerData) bool {
		return label.GetBoolValue(container.Labels, labelName, defaultValue)
	}
}

func getFuncSliceStringLabel(labelName string) func(container dockerData) []string {
	return func(container dockerData) []string {
		return label.GetSliceStringValue(container.Labels, labelName)
	}
}

func getFuncIntLabel(labelName string, defaultValue int) func(container dockerData) int {
	return func(container dockerData) int {
		return label.GetIntValue(container.Labels, labelName, defaultValue)
	}
}

func getFuncInt64Label(labelName string, defaultValue int64) func(container dockerData) int64 {
	return func(container dockerData) int64 {
		return label.GetInt64Value(container.Labels, labelName, defaultValue)
	}
}

// Deprecated
func hasFunc(labelName string) func(container dockerData) bool {
	return func(container dockerData) bool {
		return label.Has(container.Labels, labelName)
	}
}
