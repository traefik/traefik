package docker

import (
	"context"
	"io"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainertypes "github.com/docker/docker/api/types/container"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
)

const (
	// SwarmAPIVersion is a constant holding the version of the Provider API traefik will use
	SwarmAPIVersion = "1.24"
	// SwarmDefaultWatchTime is the duration of the interval when polling docker
	SwarmDefaultWatchTime = 15 * time.Second

	defaultWeight                      = "0"
	defaultProtocol                    = "http"
	defaultPassHostHeader              = "true"
	defaultFrontendPriority            = "0"
	defaultCircuitBreakerExpression    = "NetworkErrorRatio() > 1"
	defaultFrontendRedirectEntryPoint  = ""
	defaultBackendLoadBalancerMethod   = "wrr"
	defaultBackendMaxconnExtractorfunc = "request.host"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	Endpoint              string           `description:"Docker server endpoint. Can be a tcp or a unix socket endpoint"`
	Domain                string           `description:"Default domain used"`
	TLS                   *types.ClientTLS `description:"Enable Docker TLS support" export:"true"`
	ExposedByDefault      bool             `description:"Expose containers by default" export:"true"`
	UseBindPortIP         bool             `description:"Use the ip address from the bound port, rather than from the inner network" export:"true"`
	SwarmMode             bool             `description:"Use Docker on Swarm Mode" export:"true"`
}

// dockerData holds the need data to the Provider p
type dockerData struct {
	ServiceName     string
	Name            string
	Labels          map[string]string // List of labels set to container or service
	NetworkSettings networkSettings
	Health          string
	Node            *dockertypes.ContainerNode
}

// NetworkSettings holds the networks data to the Provider p
type networkSettings struct {
	NetworkMode dockercontainertypes.NetworkMode
	Ports       nat.PortMap
	Networks    map[string]*networkData
}

// Network holds the network data to the Provider p
type networkData struct {
	Name     string
	Addr     string
	Port     int
	Protocol string
	ID       string
}

func (p *Provider) createClient() (client.APIClient, error) {
	var httpClient *http.Client

	if p.TLS != nil {
		config, err := p.TLS.CreateTLSConfig()
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: config,
		}
		proto, addr, _, err := client.ParseHost(p.Endpoint)
		if err != nil {
			return nil, err
		}

		sockets.ConfigureTransport(tr, proto, addr)

		httpClient = &http.Client{
			Transport: tr,
		}
	}

	httpHeaders := map[string]string{
		"User-Agent": "Traefik " + version.Version,
	}

	var apiVersion string
	if p.SwarmMode {
		apiVersion = SwarmAPIVersion
	} else {
		apiVersion = DockerAPIVersion
	}

	return client.NewClient(p.Endpoint, apiVersion, httpClient, httpHeaders)
}

// Provide allows the docker provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.Constraints = append(p.Constraints, constraints...)
	// TODO register this routine in pool, and watch for stop channel
	safe.Go(func() {
		operation := func() error {
			var err error

			dockerClient, err := p.createClient()
			if err != nil {
				log.Errorf("Failed to create a client for docker, error: %s", err)
				return err
			}

			ctx := context.Background()
			serverVersion, err := dockerClient.ServerVersion(ctx)
			if err != nil {
				log.Errorf("Failed to retrieve information of the docker client and server host: %s", err)
				return err
			}
			log.Debugf("Provider connection established with docker %s (API %s)", serverVersion.Version, serverVersion.APIVersion)
			var dockerDataList []dockerData
			if p.SwarmMode {
				dockerDataList, err = listServices(ctx, dockerClient)
				if err != nil {
					log.Errorf("Failed to list services for docker swarm mode, error %s", err)
					return err
				}
			} else {
				dockerDataList, err = listContainers(ctx, dockerClient)
				if err != nil {
					log.Errorf("Failed to list containers for docker, error %s", err)
					return err
				}
			}

			configuration := p.loadDockerConfig(dockerDataList)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "docker",
				Configuration: configuration,
			}
			if p.Watch {
				ctx, cancel := context.WithCancel(ctx)
				if p.SwarmMode {
					errChan := make(chan error)
					// TODO: This need to be change. Linked to Swarm events docker/docker#23827
					ticker := time.NewTicker(SwarmDefaultWatchTime)
					pool.Go(func(stop chan bool) {
						defer close(errChan)
						for {
							select {
							case <-ticker.C:
								services, err := listServices(ctx, dockerClient)
								if err != nil {
									log.Errorf("Failed to list services for docker, error %s", err)
									errChan <- err
									return
								}
								configuration := p.loadDockerConfig(services)
								if configuration != nil {
									configurationChan <- types.ConfigMessage{
										ProviderName:  "docker",
										Configuration: configuration,
									}
								}

							case <-stop:
								ticker.Stop()
								cancel()
								return
							}
						}
					})
					if err, ok := <-errChan; ok {
						return err
					}
					// channel closed

				} else {
					pool.Go(func(stop chan bool) {
						for {
							select {
							case <-stop:
								cancel()
								return
							}
						}
					})
					f := filters.NewArgs()
					f.Add("type", "container")
					options := dockertypes.EventsOptions{
						Filters: f,
					}

					startStopHandle := func(m eventtypes.Message) {
						log.Debugf("Provider event received %+v", m)
						containers, err := listContainers(ctx, dockerClient)
						if err != nil {
							log.Errorf("Failed to list containers for docker, error %s", err)
							// Call cancel to get out of the monitor
							cancel()
							return
						}
						configuration := p.loadDockerConfig(containers)
						if configuration != nil {
							configurationChan <- types.ConfigMessage{
								ProviderName:  "docker",
								Configuration: configuration,
							}
						}
					}

					eventsc, errc := dockerClient.Events(ctx, options)
					for {
						select {
						case event := <-eventsc:
							if event.Action == "start" ||
								event.Action == "die" ||
								strings.HasPrefix(event.Action, "health_status") {
								startStopHandle(event)
							}
						case err := <-errc:
							if err == io.EOF {
								log.Debug("Provider event stream closed")
							}

							return err
						}
					}
				}
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to docker server %+v", err)
		}
	})

	return nil
}

func (p *Provider) loadDockerConfig(containersInspected []dockerData) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getBackend":                  getBackend,
		"getIPAddress":                p.getIPAddress,
		"getPort":                     getPort,
		"getWeight":                   getFuncStringLabel(types.LabelWeight, defaultWeight),
		"getDomain":                   getFuncStringLabel(types.LabelDomain, p.Domain),
		"getProtocol":                 getFuncStringLabel(types.LabelProtocol, defaultProtocol),
		"getPassHostHeader":           getFuncStringLabel(types.LabelFrontendPassHostHeader, defaultPassHostHeader),
		"getPriority":                 getFuncStringLabel(types.LabelFrontendPriority, defaultFrontendPriority),
		"getEntryPoints":              getFuncSliceStringLabel(types.LabelFrontendEntryPoints),
		"getBasicAuth":                getFuncSliceStringLabel(types.LabelFrontendAuthBasic),
		"getFrontendRule":             p.getFrontendRule,
		"hasCircuitBreakerLabel":      hasLabel(types.LabelBackendCircuitbreakerExpression),
		"getCircuitBreakerExpression": getFuncStringLabel(types.LabelBackendCircuitbreakerExpression, defaultCircuitBreakerExpression),
		"hasLoadBalancerLabel":        hasLoadBalancerLabel,
		"getLoadBalancerMethod":       getFuncStringLabel(types.LabelBackendLoadbalancerMethod, defaultBackendLoadBalancerMethod),
		"hasMaxConnLabels":            hasMaxConnLabels,
		"getMaxConnAmount":            getFuncInt64Label(types.LabelBackendMaxconnAmount, math.MaxInt64),
		"getMaxConnExtractorFunc":     getFuncStringLabel(types.LabelBackendMaxconnExtractorfunc, defaultBackendMaxconnExtractorfunc),
		"getSticky":                   getSticky,
		"hasStickinessLabel":          hasStickinessLabel,
		"getStickinessCookieName":     getFuncStringLabel(types.LabelBackendLoadbalancerStickinessCookieName, ""),
		"getIsBackendLBSwarm":         getIsBackendLBSwarm,
		"getServiceBackend":           getServiceBackend,
		"getWhitelistSourceRange":     getFuncSliceStringLabel(types.LabelTraefikFrontendWhitelistSourceRange),

		"hasHeaders":                        hasHeaders,
		"hasRequestHeaders":                 hasLabel(types.LabelFrontendRequestHeaders),
		"getRequestHeaders":                 getFuncMapLabel(types.LabelFrontendRequestHeaders),
		"hasResponseHeaders":                hasLabel(types.LabelFrontendResponseHeaders),
		"getResponseHeaders":                getFuncMapLabel(types.LabelFrontendResponseHeaders),
		"hasAllowedHostsHeaders":            hasLabel(types.LabelFrontendAllowedHosts),
		"getAllowedHostsHeaders":            getFuncSliceStringLabel(types.LabelFrontendAllowedHosts),
		"hasHostsProxyHeaders":              hasLabel(types.LabelFrontendHostsProxyHeaders),
		"getHostsProxyHeaders":              getFuncSliceStringLabel(types.LabelFrontendHostsProxyHeaders),
		"hasSSLRedirectHeaders":             hasLabel(types.LabelFrontendSSLRedirect),
		"getSSLRedirectHeaders":             getFuncBoolLabel(types.LabelFrontendSSLRedirect),
		"hasSSLTemporaryRedirectHeaders":    hasLabel(types.LabelFrontendSSLTemporaryRedirect),
		"getSSLTemporaryRedirectHeaders":    getFuncBoolLabel(types.LabelFrontendSSLTemporaryRedirect),
		"hasSSLHostHeaders":                 hasLabel(types.LabelFrontendSSLHost),
		"getSSLHostHeaders":                 getFuncStringLabel(types.LabelFrontendSSLHost, ""),
		"hasSSLProxyHeaders":                hasLabel(types.LabelFrontendSSLProxyHeaders),
		"getSSLProxyHeaders":                getFuncMapLabel(types.LabelFrontendSSLProxyHeaders),
		"hasSTSSecondsHeaders":              hasLabel(types.LabelFrontendSTSSeconds),
		"getSTSSecondsHeaders":              getFuncInt64Label(types.LabelFrontendSTSSeconds, 0),
		"hasSTSIncludeSubdomainsHeaders":    hasLabel(types.LabelFrontendSTSIncludeSubdomains),
		"getSTSIncludeSubdomainsHeaders":    getFuncBoolLabel(types.LabelFrontendSTSIncludeSubdomains),
		"hasSTSPreloadHeaders":              hasLabel(types.LabelFrontendSTSPreload),
		"getSTSPreloadHeaders":              getFuncBoolLabel(types.LabelFrontendSTSPreload),
		"hasForceSTSHeaderHeaders":          hasLabel(types.LabelFrontendForceSTSHeader),
		"getForceSTSHeaderHeaders":          getFuncBoolLabel(types.LabelFrontendForceSTSHeader),
		"hasFrameDenyHeaders":               hasLabel(types.LabelFrontendFrameDeny),
		"getFrameDenyHeaders":               getFuncBoolLabel(types.LabelFrontendFrameDeny),
		"hasCustomFrameOptionsValueHeaders": hasLabel(types.LabelFrontendCustomFrameOptionsValue),
		"getCustomFrameOptionsValueHeaders": getFuncStringLabel(types.LabelFrontendCustomFrameOptionsValue, ""),
		"hasContentTypeNosniffHeaders":      hasLabel(types.LabelFrontendContentTypeNosniff),
		"getContentTypeNosniffHeaders":      getFuncBoolLabel(types.LabelFrontendContentTypeNosniff),
		"hasBrowserXSSFilterHeaders":        hasLabel(types.LabelFrontendBrowserXSSFilter),
		"getBrowserXSSFilterHeaders":        getFuncBoolLabel(types.LabelFrontendBrowserXSSFilter),
		"hasContentSecurityPolicyHeaders":   hasLabel(types.LabelFrontendContentSecurityPolicy),
		"getContentSecurityPolicyHeaders":   getFuncStringLabel(types.LabelFrontendContentSecurityPolicy, ""),
		"hasPublicKeyHeaders":               hasLabel(types.LabelFrontendPublicKey),
		"getPublicKeyHeaders":               getFuncStringLabel(types.LabelFrontendPublicKey, ""),
		"hasReferrerPolicyHeaders":          hasLabel(types.LabelFrontendReferrerPolicy),
		"getReferrerPolicyHeaders":          getFuncStringLabel(types.LabelFrontendReferrerPolicy, ""),
		"hasIsDevelopmentHeaders":           hasLabel(types.LabelFrontendIsDevelopment),
		"getIsDevelopmentHeaders":           getFuncBoolLabel(types.LabelFrontendIsDevelopment),

		"hasServices":              hasServices,
		"getServiceNames":          getServiceNames,
		"getServicePort":           getServicePort,
		"getServiceWeight":         getFuncServiceStringLabel(types.SuffixWeight, defaultWeight),
		"getServiceProtocol":       getFuncServiceStringLabel(types.SuffixProtocol, defaultProtocol),
		"getServiceEntryPoints":    getFuncServiceSliceStringLabel(types.SuffixFrontendEntryPoints),
		"getServiceBasicAuth":      getFuncServiceSliceStringLabel(types.SuffixFrontendAuthBasic),
		"getServiceFrontendRule":   p.getServiceFrontendRule,
		"getServicePassHostHeader": getFuncServiceStringLabel(types.SuffixFrontendPassHostHeader, defaultPassHostHeader),
		"getServicePriority":       getFuncServiceStringLabel(types.SuffixFrontendPriority, defaultFrontendPriority),

		"hasRedirect":                   hasRedirect,
		"getRedirectEntryPoint":         getFuncStringLabel(types.LabelFrontendRedirectEntryPoint, defaultFrontendRedirectEntryPoint),
		"getRedirectRegex":              getFuncStringLabel(types.LabelFrontendRedirectRegex, ""),
		"getRedirectReplacement":        getFuncStringLabel(types.LabelFrontendRedirectReplacement, ""),
		"hasServiceRedirect":            hasServiceRedirect,
		"getServiceRedirectEntryPoint":  getFuncServiceStringLabel(types.SuffixFrontendRedirectEntryPoint, defaultFrontendRedirectEntryPoint),
		"getServiceRedirectReplacement": getFuncServiceStringLabel(types.SuffixFrontendRedirectReplacement, ""),
		"getServiceRedirectRegex":       getFuncServiceStringLabel(types.SuffixFrontendRedirectRegex, ""),
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
		serviceNameKey := getServiceNameKey(container, p.SwarmMode)
		if _, exists := serviceNames[serviceNameKey]; !exists {
			frontendName := p.getFrontendName(container, idx)
			frontends[frontendName] = append(frontends[frontendName], container)
			if len(serviceNameKey) > 0 {
				serviceNames[serviceNameKey] = struct{}{}
			}
		}
		backendName := getBackend(container)
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
		filteredContainers,
		frontends,
		backends,
		servers,
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/docker.tmpl", DockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func getServiceNameKey(container dockerData, swarmMode bool) string {
	serviceNameKey := container.ServiceName

	if len(container.Labels[labelDockerComposeProject]) > 0 && len(container.Labels[labelDockerComposeService]) > 0 && !swarmMode {
		serviceNameKey = container.Labels[labelDockerComposeService] + container.Labels[labelDockerComposeProject]
	}

	return serviceNameKey
}

// Regexp used to extract the name of the service and the name of the property for this service
// All properties are under the format traefik.<servicename>.frontent.*= except the port/weight/protocol directly after traefik.<servicename>.
var servicesPropertiesRegexp = regexp.MustCompile(`^traefik\.(?P<service_name>.+?)\.(?P<property_name>port|weight|protocol|frontend\.(.*))$`)

// Check if for the given container, we find labels that are defining services
func hasServices(container dockerData) bool {
	return len(extractServicesLabels(container.Labels)) > 0
}

// Gets array of service names for a given container
func getServiceNames(container dockerData) []string {
	labelServiceProperties := extractServicesLabels(container.Labels)
	keys := make([]string, 0, len(labelServiceProperties))
	for k := range labelServiceProperties {
		keys = append(keys, k)
	}
	return keys
}

// Extract backend from labels for a given service and a given docker container
func getServiceBackend(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, types.SuffixFrontendBackend); ok {
		return provider.Normalize(container.ServiceName + "-" + value)
	}
	return provider.Normalize(container.ServiceName + "-" + getBackend(container) + "-" + serviceName)
}

// Extract rule from labels for a given service and a given docker container
func (p Provider) getServiceFrontendRule(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, types.SuffixFrontendRule); ok {
		return value
	}
	return p.getFrontendRule(container)
}

// Extract port from labels for a given service and a given docker container
func getServicePort(container dockerData, serviceName string) string {
	if value, ok := getContainerServiceLabel(container, serviceName, types.SuffixPort); ok {
		return value
	}
	return getPort(container)
}

func hasLoadBalancerLabel(container dockerData) bool {
	_, errMethod := getLabel(container, types.LabelBackendLoadbalancerMethod)
	_, errSticky := getLabel(container, types.LabelBackendLoadbalancerSticky)
	_, errStickiness := getLabel(container, types.LabelBackendLoadbalancerStickiness)
	_, errCookieName := getLabel(container, types.LabelBackendLoadbalancerStickinessCookieName)

	return errMethod == nil || errSticky == nil || errStickiness == nil || errCookieName == nil
}

func hasMaxConnLabels(container dockerData) bool {
	if _, err := getLabel(container, types.LabelBackendMaxconnAmount); err != nil {
		return false
	}
	if _, err := getLabel(container, types.LabelBackendMaxconnExtractorfunc); err != nil {
		return false
	}
	return true
}

func (p Provider) containerFilter(container dockerData) bool {
	if !isContainerEnabled(container, p.ExposedByDefault) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	var err error
	portLabel := "traefik.port label"
	if hasServices(container) {
		portLabel = "traefik.<serviceName>.port or " + portLabel + "s"
		err = checkServiceLabelPort(container)
	} else {
		_, err = strconv.Atoi(container.Labels[types.LabelPort])
	}
	if len(container.NetworkSettings.Ports) == 0 && err != nil {
		log.Debugf("Filtering container without port and no %s %s : %s", portLabel, container.Name, err.Error())
		return false
	}

	constraintTags := strings.Split(container.Labels[types.LabelTags], ",")
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

// checkServiceLabelPort checks if all service names have a port service label
// or if port container label exists for default value
func checkServiceLabelPort(container dockerData) error {
	// If port container label is present, there is a default values for all ports, use it for the check
	_, err := strconv.Atoi(container.Labels[types.LabelPort])
	if err != nil {
		serviceLabelPorts := make(map[string]struct{})
		serviceLabels := make(map[string]struct{})
		portRegexp := regexp.MustCompile(`^traefik\.(?P<service_name>.+?)\.port$`)
		for label := range container.Labels {
			// Get all port service labels
			portLabel := portRegexp.FindStringSubmatch(label)
			if portLabel != nil && len(portLabel) > 0 {
				serviceLabelPorts[portLabel[0]] = struct{}{}
			}
			// Get only one instance of all service names from service labels
			servicesLabelNames := servicesPropertiesRegexp.FindStringSubmatch(label)
			if servicesLabelNames != nil && len(servicesLabelNames) > 0 {
				serviceLabels[strings.Split(servicesLabelNames[0], ".")[1]] = struct{}{}
			}
		}
		// If the number of service labels is different than the number of port services label
		// there is an error
		if len(serviceLabels) == len(serviceLabelPorts) {
			for labelPort := range serviceLabelPorts {
				_, err = strconv.Atoi(container.Labels[labelPort])
				if err != nil {
					break
				}
			}
		} else {
			err = errors.New("Port service labels missing, please use traefik.port as default value or define all port service labels")
		}
	}
	return err
}

func (p Provider) getFrontendName(container dockerData, idx int) string {
	// Replace '.' with '-' in quoted keys because of this issue https://github.com/BurntSushi/toml/issues/78
	return provider.Normalize(p.getFrontendRule(container) + "-" + strconv.Itoa(idx))
}

// GetFrontendRule returns the frontend rule for the specified container, using
// it's label. It returns a default one (Host) if the label is not present.
func (p Provider) getFrontendRule(container dockerData) string {
	if label, err := getLabel(container, types.LabelFrontendRule); err == nil {
		return label
	}
	if labels, err := getLabels(container, []string{labelDockerComposeProject, labelDockerComposeService}); err == nil {
		return "Host:" + getSubDomain(labels[labelDockerComposeService]+"."+labels[labelDockerComposeProject]) + "." + p.Domain
	}
	if len(p.Domain) > 0 {
		return "Host:" + getSubDomain(container.ServiceName) + "." + p.Domain
	}
	return ""
}

func getBackend(container dockerData) string {
	if label, err := getLabel(container, types.LabelBackend); err == nil {
		return provider.Normalize(label)
	}
	if labels, err := getLabels(container, []string{labelDockerComposeProject, labelDockerComposeService}); err == nil {
		return provider.Normalize(labels[labelDockerComposeService] + "_" + labels[labelDockerComposeProject])
	}
	return provider.Normalize(container.ServiceName)
}

func (p Provider) getIPAddress(container dockerData) string {
	if label, err := getLabel(container, labelDockerNetwork); err == nil && label != "" {
		networkSettings := container.NetworkSettings
		if networkSettings.Networks != nil {
			network := networkSettings.Networks[label]
			if network != nil {
				return network.Addr
			}

			log.Warnf("Could not find network named '%s' for container '%s'! Maybe you're missing the project's prefix in the label? Defaulting to first available network.", label, container.Name)
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
		ctx := context.Background()
		containerInspected, err := dockerClient.ContainerInspect(ctx, container.NetworkSettings.NetworkMode.ConnectedContainer())
		if err != nil {
			log.Warnf("Unable to get IP address for container %s : Failed to inspect container ID %s, error: %s", container.Name, container.NetworkSettings.NetworkMode.ConnectedContainer(), err)
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

func getPort(container dockerData) string {
	if label, err := getLabel(container, types.LabelPort); err == nil {
		return label
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

func hasStickinessLabel(container dockerData) bool {
	labelStickiness, errStickiness := getLabel(container, types.LabelBackendLoadbalancerStickiness)
	return errStickiness == nil && len(labelStickiness) > 0 && strings.EqualFold(strings.TrimSpace(labelStickiness), "true")
}

// Deprecated replaced by Stickiness
func getSticky(container dockerData) string {
	if label, err := getLabel(container, types.LabelBackendLoadbalancerSticky); err == nil {
		if len(label) > 0 {
			log.Warnf("Deprecated configuration found: %s. Please use %s.", types.LabelBackendLoadbalancerSticky, types.LabelBackendLoadbalancerStickiness)
		}
		return label
	}
	return "false"
}

func getIsBackendLBSwarm(container dockerData) string {
	return getStringLabel(container, labelBackendLoadBalancerSwarm, "false")
}

func isContainerEnabled(container dockerData, exposedByDefault bool) bool {
	return exposedByDefault && container.Labels[types.LabelEnable] != "false" || container.Labels[types.LabelEnable] == "true"
}

func listContainers(ctx context.Context, dockerClient client.ContainerAPIClient) ([]dockerData, error) {
	containerList, err := dockerClient.ContainerList(ctx, dockertypes.ContainerListOptions{})
	if err != nil {
		return []dockerData{}, err
	}

	var containersInspected []dockerData
	// get inspect containers
	for _, container := range containerList {
		containerInspected, err := dockerClient.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Warnf("Failed to inspect container %s, error: %s", container.ID, err)
		} else {
			// This condition is here to avoid to have empty IP https://github.com/containous/traefik/issues/2459
			// We register only container which are running
			if containerInspected.ContainerJSONBase != nil && containerInspected.ContainerJSONBase.State != nil && containerInspected.ContainerJSONBase.State.Running {
				dockerData := parseContainer(containerInspected)
				containersInspected = append(containersInspected, dockerData)
			}
		}
	}
	return containersInspected, nil
}

func parseContainer(container dockertypes.ContainerJSON) dockerData {
	dockerData := dockerData{
		NetworkSettings: networkSettings{},
	}

	if container.ContainerJSONBase != nil {
		dockerData.Name = container.ContainerJSONBase.Name
		dockerData.ServiceName = dockerData.Name //Default ServiceName to be the container's Name.
		dockerData.Node = container.ContainerJSONBase.Node

		if container.ContainerJSONBase.HostConfig != nil {
			dockerData.NetworkSettings.NetworkMode = container.ContainerJSONBase.HostConfig.NetworkMode
		}

		if container.State != nil && container.State.Health != nil {
			dockerData.Health = container.State.Health.Status
		}
	}

	if container.Config != nil && container.Config.Labels != nil {
		dockerData.Labels = container.Config.Labels
	}

	if container.NetworkSettings != nil {
		if container.NetworkSettings.Ports != nil {
			dockerData.NetworkSettings.Ports = container.NetworkSettings.Ports
		}
		if container.NetworkSettings.Networks != nil {
			dockerData.NetworkSettings.Networks = make(map[string]*networkData)
			for name, containerNetwork := range container.NetworkSettings.Networks {
				dockerData.NetworkSettings.Networks[name] = &networkData{
					ID:   containerNetwork.NetworkID,
					Name: name,
					Addr: containerNetwork.IPAddress,
				}
			}
		}
	}
	return dockerData
}

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func getSubDomain(name string) string {
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func listServices(ctx context.Context, dockerClient client.APIClient) ([]dockerData, error) {
	serviceList, err := dockerClient.ServiceList(ctx, dockertypes.ServiceListOptions{})
	if err != nil {
		return []dockerData{}, err
	}

	serverVersion, err := dockerClient.ServerVersion(ctx)

	networkListArgs := filters.NewArgs()
	// https://docs.docker.com/engine/api/v1.29/#tag/Network (Docker 17.06)
	if versions.GreaterThanOrEqualTo(serverVersion.APIVersion, "1.29") {
		networkListArgs.Add("scope", "swarm")
	} else {
		networkListArgs.Add("driver", "overlay")
	}

	networkList, err := dockerClient.NetworkList(ctx, dockertypes.NetworkListOptions{Filters: networkListArgs})

	networkMap := make(map[string]*dockertypes.NetworkResource)
	if err != nil {
		log.Debugf("Failed to network inspect on client for docker, error: %s", err)
		return []dockerData{}, err
	}
	for _, network := range networkList {
		networkToAdd := network
		networkMap[network.ID] = &networkToAdd
	}

	var dockerDataList []dockerData
	var dockerDataListTasks []dockerData

	for _, service := range serviceList {
		dockerData := parseService(service, networkMap)

		useSwarmLB, _ := strconv.ParseBool(getIsBackendLBSwarm(dockerData))

		if useSwarmLB {
			if len(dockerData.NetworkSettings.Networks) > 0 {
				dockerDataList = append(dockerDataList, dockerData)
			}
		} else {
			isGlobalSvc := service.Spec.Mode.Global != nil
			dockerDataListTasks, err = listTasks(ctx, dockerClient, service.ID, dockerData, networkMap, isGlobalSvc)

			for _, dockerDataTask := range dockerDataListTasks {
				dockerDataList = append(dockerDataList, dockerDataTask)
			}

		}
	}
	return dockerDataList, err
}

func parseService(service swarmtypes.Service, networkMap map[string]*dockertypes.NetworkResource) dockerData {
	dockerData := dockerData{
		ServiceName:     service.Spec.Annotations.Name,
		Name:            service.Spec.Annotations.Name,
		Labels:          service.Spec.Annotations.Labels,
		NetworkSettings: networkSettings{},
	}

	if service.Spec.EndpointSpec != nil {
		if service.Spec.EndpointSpec.Mode == swarmtypes.ResolutionModeDNSRR {
			useSwarmLB, _ := strconv.ParseBool(getIsBackendLBSwarm(dockerData))
			if useSwarmLB {
				log.Warnf("Ignored %s endpoint-mode not supported, service name: %s. Fallback to Træfik load balancing", swarmtypes.ResolutionModeDNSRR, service.Spec.Annotations.Name)
			}
		} else if service.Spec.EndpointSpec.Mode == swarmtypes.ResolutionModeVIP {
			dockerData.NetworkSettings.Networks = make(map[string]*networkData)
			for _, virtualIP := range service.Endpoint.VirtualIPs {
				networkService := networkMap[virtualIP.NetworkID]
				if networkService != nil {
					ip, _, _ := net.ParseCIDR(virtualIP.Addr)
					network := &networkData{
						Name: networkService.Name,
						ID:   virtualIP.NetworkID,
						Addr: ip.String(),
					}
					dockerData.NetworkSettings.Networks[network.Name] = network
				} else {
					log.Debugf("Network not found, id: %s", virtualIP.NetworkID)
				}
			}
		}
	}
	return dockerData
}

func listTasks(ctx context.Context, dockerClient client.APIClient, serviceID string,
	serviceDockerData dockerData, networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool) ([]dockerData, error) {
	serviceIDFilter := filters.NewArgs()
	serviceIDFilter.Add("service", serviceID)
	serviceIDFilter.Add("desired-state", "running")
	taskList, err := dockerClient.TaskList(ctx, dockertypes.TaskListOptions{Filters: serviceIDFilter})

	if err != nil {
		return []dockerData{}, err
	}
	var dockerDataList []dockerData

	for _, task := range taskList {
		if task.Status.State != swarmtypes.TaskStateRunning {
			continue
		}
		dockerData := parseTasks(task, serviceDockerData, networkMap, isGlobalSvc)
		if len(dockerData.NetworkSettings.Networks) > 0 {
			dockerDataList = append(dockerDataList, dockerData)
		}
	}
	return dockerDataList, err
}

func parseTasks(task swarmtypes.Task, serviceDockerData dockerData, networkMap map[string]*dockertypes.NetworkResource, isGlobalSvc bool) dockerData {
	dockerData := dockerData{
		ServiceName:     serviceDockerData.Name,
		Name:            serviceDockerData.Name + "." + strconv.Itoa(task.Slot),
		Labels:          serviceDockerData.Labels,
		NetworkSettings: networkSettings{},
	}

	if isGlobalSvc {
		dockerData.Name = serviceDockerData.Name + "." + task.ID
	}

	if task.NetworksAttachments != nil {
		dockerData.NetworkSettings.Networks = make(map[string]*networkData)
		for _, virtualIP := range task.NetworksAttachments {
			if networkService, present := networkMap[virtualIP.Network.ID]; present {
				// Not sure about this next loop - when would a task have multiple IP's for the same network?
				for _, addr := range virtualIP.Addresses {
					ip, _, _ := net.ParseCIDR(addr)
					network := &networkData{
						ID:   virtualIP.Network.ID,
						Name: networkService.Name,
						Addr: ip.String(),
					}
					dockerData.NetworkSettings.Networks[network.Name] = network
				}
			}
		}
	}
	return dockerData
}

// TODO will be rewrite when merge on master
func hasServiceRedirect(container dockerData, serviceName string) bool {
	serviceLabels, ok := extractServicesLabels(container.Labels)[serviceName]
	if !ok || len(serviceLabels) == 0 {
		return false
	}

	value, ok := serviceLabels[types.SuffixFrontendRedirectEntryPoint]
	frep := ok && len(value) > 0
	value, ok = serviceLabels[types.SuffixFrontendRedirectRegex]
	frrg := ok && len(value) > 0
	value, ok = serviceLabels[types.SuffixFrontendRedirectReplacement]
	frrp := ok && len(value) > 0

	return frep || frrg && frrp
}

// TODO will be rewrite when merge on master
func hasRedirect(container dockerData) bool {
	return hasLabel(types.LabelFrontendRedirectEntryPoint)(container) ||
		hasLabel(types.LabelFrontendRedirectReplacement)(container) && hasLabel(types.LabelFrontendRedirectRegex)(container)
}

// TODO will be rewrite when merge on master
func hasHeaders(container dockerData) bool {
	// LabelPrefix + "frontend.headers.

	for key := range container.Labels {
		if strings.HasPrefix(key, types.LabelPrefix+"frontend.headers.") {
			return true
		}
	}
	return false
}
