package docker

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
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
		"getIPAddress":          p.getDeprecatedIPAddress, // TODO: Should we expose getIPPort instead?
		"getServers":            p.getServers,
		"getMaxConn":            label.GetMaxConn,
		"getHealthCheck":        label.GetHealthCheck,
		"getBuffering":          label.GetBuffering,
		"getResponseForwarding": label.GetResponseForwarding,
		"getCircuitBreaker":     label.GetCircuitBreaker,
		"getLoadBalancer":       label.GetLoadBalancer,

		// Frontend functions
		"getBackendName":       getBackendName,
		"getPriority":          label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":    label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":       label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPassTLSClientCert": label.GetTLSClientCert,
		"getEntryPoints":       label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":         label.GetFuncSliceString(label.TraefikFrontendAuthBasic), // Deprecated
		"getAuth":              label.GetAuth,
		"getFrontendRule":      p.getFrontendRule,
		"getRedirect":          label.GetRedirect,
		"getErrorPages":        label.GetErrorPages,
		"getRateLimit":         label.GetRateLimit,
		"getHeaders":           label.GetHeaders,
		"getWhiteList":         label.GetWhiteList,
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

			serviceNamesKey := getServiceNameKey(container, p.SwarmMode, segmentName)

			if _, exists := serviceNames[serviceNamesKey]; !exists {
				frontendName := p.getFrontendName(container, idx)
				frontends[frontendName] = append(frontends[frontendName], container)
				if len(serviceNamesKey) > 0 {
					serviceNames[serviceNamesKey] = struct{}{}
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

func getServiceNameKey(container dockerData, swarmMode bool, segmentName string) string {
	if swarmMode {
		return container.ServiceName + segmentName
	}

	return getServiceName(container) + segmentName
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
		name = container.SegmentName + "-" + getBackendName(container)
	} else {
		name = p.getFrontendRule(container, container.SegmentLabels) + "-" + strconv.Itoa(idx)
	}

	return provider.Normalize(name)
}

func (p *Provider) getFrontendRule(container dockerData, segmentLabels map[string]string) string {
	if value := label.GetStringValue(segmentLabels, label.TraefikFrontendRule, ""); len(value) != 0 {
		return value
	}

	domain := label.GetStringValue(segmentLabels, label.TraefikDomain, p.Domain)
	if len(domain) > 0 {
		domain = "." + domain
	}

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		return "Host:" + getSubDomain(values[labelDockerComposeService]+"."+values[labelDockerComposeProject]) + domain
	}

	if len(domain) > 0 {
		return "Host:" + getSubDomain(container.ServiceName) + domain
	}

	return ""
}

func (p Provider) getIPAddress(container dockerData) string {
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

	for _, network := range container.NetworkSettings.Networks {
		return network.Addr
	}

	log.Warnf("Unable to find the IP address for the container %q.", container.Name)
	return ""
}

// Deprecated: Please use getIPPort instead
func (p *Provider) getDeprecatedIPAddress(container dockerData) string {
	ip, _, err := p.getIPPort(container)
	if err != nil {
		log.Warn(err)
		return ""
	}
	return ip
}

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func getSubDomain(name string) string {
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func isBackendLBSwarm(container dockerData) bool {
	return label.GetBoolValue(container.Labels, labelBackendLoadBalancerSwarm, false)
}

func getBackendName(container dockerData) string {
	if len(container.SegmentName) > 0 {
		return getSegmentBackendName(container)
	}

	return getDefaultBackendName(container)
}

func getSegmentBackendName(container dockerData) string {
	serviceName := getServiceName(container)
	if value := label.GetStringValue(container.SegmentLabels, label.TraefikBackend, ""); len(value) > 0 {
		return provider.Normalize(serviceName + "-" + value)
	}

	return provider.Normalize(serviceName + "-" + container.SegmentName)
}

func getDefaultBackendName(container dockerData) string {
	if value := label.GetStringValue(container.SegmentLabels, label.TraefikBackend, ""); len(value) != 0 {
		return provider.Normalize(value)
	}

	return provider.Normalize(getServiceName(container))
}

func getServiceName(container dockerData) string {
	serviceName := container.ServiceName

	if values, err := label.GetStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		serviceName = values[labelDockerComposeService] + "_" + values[labelDockerComposeProject]
	}

	return serviceName
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

func (p *Provider) getPortBinding(container dockerData) (*nat.PortBinding, error) {
	port := getPort(container)
	for netPort, portBindings := range container.NetworkSettings.Ports {
		if strings.EqualFold(string(netPort), port+"/TCP") || strings.EqualFold(string(netPort), port+"/UDP") {
			for _, p := range portBindings {
				return &p, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to find the external IP:Port for the container %q", container.Name)
}

func (p *Provider) getIPPort(container dockerData) (string, string, error) {
	var ip, port string
	usedBound := false

	if p.UseBindPortIP {
		portBinding, err := p.getPortBinding(container)
		if err != nil {
			log.Infof("Unable to find a binding for container %q, falling back on its internal IP/Port.", container.Name)
		} else if (portBinding.HostIP == "0.0.0.0") || (len(portBinding.HostIP) == 0) {
			log.Infof("Cannot determine the IP address (got %q) for %q's binding, falling back on its internal IP/Port.", portBinding.HostIP, container.Name)
		} else {
			ip = portBinding.HostIP
			port = portBinding.HostPort
			usedBound = true
		}
	}

	if !usedBound {
		ip = p.getIPAddress(container)
		port = getPort(container)
	}

	if len(ip) == 0 {
		return "", "", fmt.Errorf("unable to find the IP address for the container %q: the server is ignored", container.Name)
	}

	return ip, port, nil
}

func (p *Provider) getServers(containers []dockerData) map[string]types.Server {
	var servers map[string]types.Server

	for _, container := range containers {
		ip, port, err := p.getIPPort(container)
		if err != nil {
			log.Warn(err)
			continue
		}

		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(container.SegmentLabels, label.TraefikProtocol, label.DefaultProtocol)

		serverURL := fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(ip, port))

		serverName := getServerName(container.Name, serverURL)
		if _, exist := servers[serverName]; exist {
			log.Debugf("Skipping server %q with the same URL.", serverName)
			continue
		}

		servers[serverName] = types.Server{
			URL:    serverURL,
			Weight: label.GetIntValue(container.SegmentLabels, label.TraefikWeight, label.DefaultWeight),
		}
	}

	return servers
}

func getServerName(containerName, url string) string {
	hash := md5.New()
	_, err := hash.Write([]byte(url))
	if err != nil {
		// Impossible case
		log.Errorf("Fail to hash server URL %q", url)
	}

	return provider.Normalize("server-" + containerName + "-" + hex.EncodeToString(hash.Sum(nil)))
}
