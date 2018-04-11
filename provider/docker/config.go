package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
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

func (p *Provider) buildConfigurationV2(containersInspected []dockerData) *types.Configuration {
	dockerFuncMap := template.FuncMap{
		"getLabelValue":    label.GetStringValue,
		"getSubDomain":     getSubDomain,
		"isBackendLBSwarm": isBackendLBSwarm,
		"getDomain":        label.GetFuncString(label.TraefikDomain, p.Domain),

		// Backend functions
		"getIPAddress":      p.getIPAddress,
		"getServers":        p.getServers,
		"getMaxConn":        label.GetMaxConn,
		"getHealthCheck":    label.GetHealthCheck,
		"getBuffering":      label.GetBuffering,
		"getCircuitBreaker": label.GetCircuitBreaker,
		"getLoadBalancer":   label.GetLoadBalancer,

		// Frontend functions
		"getBackendName":    getBackendName,
		"getPriority":       label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader": label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":    label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":    label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":      label.GetFuncSliceString(label.TraefikFrontendAuthBasic),
		"getFrontendRule":   p.getFrontendRule,
		"getRedirect":       label.GetRedirect,
		"getErrorPages":     label.GetErrorPages,
		"getRateLimit":      label.GetRateLimit,
		"getHeaders":        label.GetHeaders,
		"getWhiteList":      label.GetWhiteList,
	}

	// filter containers
	filteredContainers := fun.Filter(p.containerFilter, containersInspected).([]dockerData)

	frontends := map[string][]dockerData{}
	servers := map[string][]dockerData{}

	serviceNames := make(map[string]struct{})

	for idx, container := range filteredContainers {
		segmentProperties := label.ExtractTraefikLabels(container.Labels)
		for segmentName, labels := range segmentProperties {
			container.SegmentLabels = labels
			container.SegmentName = segmentName

			// Frontends
			if _, exists := serviceNames[container.ServiceName+segmentName]; !exists {
				frontendName := p.getFrontendName(container, idx)
				frontends[frontendName] = append(frontends[frontendName], container)
				if len(container.ServiceName+segmentName) > 0 {
					serviceNames[container.ServiceName+segmentName] = struct{}{}
				}
			}

			// Backends
			backendName := getBackendName(container)

			// Servers
			servers[backendName] = append(servers[backendName], container)
		}
	}

	templateObjects := struct {
		Containers []dockerData
		Frontends  map[string][]dockerData
		Servers    map[string][]dockerData
		Domain     string
	}{
		Containers: filteredContainers,
		Frontends:  frontends,
		Servers:    servers,
		Domain:     p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/docker.tmpl", dockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func (p *Provider) containerFilter(container dockerData) bool {
	if !label.IsEnabled(container.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	segmentProperties := label.ExtractTraefikLabels(container.Labels)

	var errPort error
	for segmentName, labels := range segmentProperties {
		errPort = checkSegmentPort(labels, segmentName)

		if len(p.getFrontendRule(container, labels)) == 0 {
			log.Debugf("Filtering container with empty frontend rule %s %s", container.Name, segmentName)
			return false
		}
	}

	if len(container.NetworkSettings.Ports) == 0 && errPort != nil {
		log.Debugf("Filtering container without port, %s: %v", container.Name, errPort)
		return false
	}

	constraintTags := label.SplitAndTrimString(container.Labels[label.TraefikTags], ",")
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Container %s pruned by %q constraint", container.Name, failingConstraint.String())
		}
		return false
	}

	if container.Health != "" && container.Health != "healthy" {
		log.Debugf("Filtering unhealthy or starting container %s", container.Name)
		return false
	}

	return true
}

func checkSegmentPort(labels map[string]string, segmentName string) error {
	if port, ok := labels[label.TraefikPort]; ok {
		_, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid port value %q for the segment %q: %v", port, segmentName, err)
		}
	} else {
		return fmt.Errorf("port label is missing, please use %s as default value or define port label for all segments ('traefik.<segment_name>.port')", label.TraefikPort)
	}
	return nil
}

func (p *Provider) getFrontendName(container dockerData, idx int) string {
	var name string
	if len(container.SegmentName) > 0 {
		name = getBackendName(container)
	} else {
		name = p.getFrontendRule(container, container.SegmentLabels) + "-" + strconv.Itoa(idx)
	}

	return provider.Normalize(name)
}

func (p *Provider) getFrontendRule(container dockerData, segmentLabels map[string]string) string {
	if value := label.GetStringValue(segmentLabels, label.TraefikFrontendRule, ""); len(value) != 0 {
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
		port := getPortV1(container)
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

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func getSubDomain(name string) string {
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func isBackendLBSwarm(container dockerData) bool {
	return label.GetBoolValue(container.Labels, labelBackendLoadBalancerSwarm, false)
}

func getSegmentBackendName(container dockerData) string {
	if value := label.GetStringValue(container.SegmentLabels, label.TraefikFrontendBackend, ""); len(value) > 0 {
		return provider.Normalize(container.ServiceName + "-" + value)
	}

	return provider.Normalize(container.ServiceName + "-" + getDefaultBackendName(container) + "-" + container.SegmentName)
}

func getDefaultBackendName(container dockerData) string {
	if value := label.GetStringValue(container.SegmentLabels, label.TraefikBackend, ""); len(value) != 0 {
		return provider.Normalize(value)
	}

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		return provider.Normalize(values[labelDockerComposeService] + "_" + values[labelDockerComposeProject])
	}

	return provider.Normalize(container.ServiceName)
}

func getBackendName(container dockerData) string {
	if len(container.SegmentName) > 0 {
		return getSegmentBackendName(container)
	}

	return getDefaultBackendName(container)
}

func getPort(container dockerData) string {
	if value := label.GetStringValue(container.SegmentLabels, label.TraefikPort, ""); len(value) != 0 {
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

func (p *Provider) getServers(containers []dockerData) map[string]types.Server {
	var servers map[string]types.Server

	for i, container := range containers {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(container.SegmentLabels, label.TraefikProtocol, label.DefaultProtocol)
		ip := p.getIPAddress(container)
		port := getPort(container)

		serverName := "server-" + container.SegmentName + "-" + container.Name
		if len(container.SegmentName) > 0 {
			serverName += "-" + strconv.Itoa(i)
		}

		servers[provider.Normalize(serverName)] = types.Server{
			URL:    fmt.Sprintf("%s://%s:%s", protocol, ip, port),
			Weight: label.GetIntValue(container.SegmentLabels, label.TraefikWeight, label.DefaultWeight),
		}
	}

	return servers
}
