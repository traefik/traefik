package docker

import (
	"context"
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

func hasLoadBalancerLabel(container dockerData) bool {
	method := label.Has(container.Labels, label.TraefikBackendLoadBalancerMethod)
	sticky := label.Has(container.Labels, label.TraefikBackendLoadBalancerSticky)
	stickiness := label.Has(container.Labels, label.TraefikBackendLoadBalancerStickiness)
	cookieName := label.Has(container.Labels, label.TraefikBackendLoadBalancerStickinessCookieName)
	return method || sticky || stickiness || cookieName
}

func hasMaxConnLabels(container dockerData) bool {
	mca := label.Has(container.Labels, label.TraefikBackendMaxConnAmount)
	mcef := label.Has(container.Labels, label.TraefikBackendMaxConnExtractorFunc)
	return mca && mcef
}

func getBackend(container dockerData) string {
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

// TODO: Deprecated
// Deprecated replaced by Stickiness
func getSticky(container dockerData) string {
	if label.Has(container.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}

	return label.GetStringValue(container.Labels, label.TraefikBackendLoadBalancerSticky, "false")
}

func isBackendLBSwarm(container dockerData) bool {
	return label.GetBoolValue(container.Labels, labelBackendLoadBalancerSwarm, false)
}

func hasRedirect(container dockerData) bool {
	return label.Has(container.Labels, label.TraefikFrontendRedirectEntryPoint) ||
		label.Has(container.Labels, label.TraefikFrontendRedirectReplacement) && label.Has(container.Labels, label.TraefikFrontendRedirectRegex)
}

func hasErrorPages(container dockerData) bool {
	return label.HasPrefix(container.Labels, label.Prefix+label.BaseFrontendErrorPage)
}

func getErrorPages(container dockerData) map[string]*types.ErrorPage {
	prefix := label.Prefix + label.BaseFrontendErrorPage
	return label.ParseErrorPages(container.Labels, prefix, label.RegexpFrontendErrorPage)
}

func getRateLimits(container dockerData) map[string]*types.Rate {
	prefix := label.Prefix + label.BaseFrontendRateLimit
	return label.ParseRateSets(container.Labels, prefix, label.RegexpFrontendRateLimit)
}

func hasHeaders(container dockerData) bool {
	for key := range container.Labels {
		if strings.HasPrefix(key, label.TraefikFrontendHeaders) {
			return true
		}
	}
	return false
}

// Label functions

func getFuncInt64Label(labelName string, defaultValue int64) func(container dockerData) int64 {
	return func(container dockerData) int64 {
		return label.GetInt64Value(container.Labels, labelName, defaultValue)
	}
}

func getFuncMapLabel(labelName string) func(container dockerData) map[string]string {
	return func(container dockerData) map[string]string {
		return label.GetMapValue(container.Labels, labelName)
	}
}

func getFuncStringLabel(labelName string, defaultValue string) func(container dockerData) string {
	return func(container dockerData) string {
		return label.GetStringValue(container.Labels, labelName, defaultValue)
	}
}

func getFuncIntLabel(labelName string, defaultValue int) func(container dockerData) int {
	return func(container dockerData) int {
		return label.GetIntValue(container.Labels, labelName, defaultValue)
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

func hasFunc(labelName string) func(container dockerData) bool {
	return func(container dockerData) bool {
		return label.Has(container.Labels, labelName)
	}
}
