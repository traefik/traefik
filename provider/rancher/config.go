package rancher

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfigurationV2(services []rancherData) *types.Configuration {
	var RancherFuncMap = template.FuncMap{
		"getLabelValue": label.GetStringValue,
		"getDomain":     label.GetFuncString(label.TraefikDomain, p.Domain),

		// Backend functions
		"getCircuitBreaker": label.GetCircuitBreaker,
		"getLoadBalancer":   label.GetLoadBalancer,
		"getMaxConn":        label.GetMaxConn,
		"getHealthCheck":    label.GetHealthCheck,
		"getBuffering":      label.GetBuffering,
		"getServers":        getServers,

		// Frontend functions
		"getBackendName":    getBackendName,
		"getFrontendRule":   p.getFrontendRule,
		"getPriority":       label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
		"getPassHostHeader": label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":    label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":    label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":      label.GetFuncSliceString(label.TraefikFrontendAuthBasic),
		"getErrorPages":     label.GetErrorPages,
		"getRateLimit":      label.GetRateLimit,
		"getRedirect":       label.GetRedirect,
		"getHeaders":        label.GetHeaders,
		"getWhiteList":      label.GetWhiteList,
	}

	// filter services
	filteredServices := fun.Filter(p.serviceFilter, services).([]rancherData)

	frontends := map[string]rancherData{}
	backends := map[string]rancherData{}

	for _, service := range filteredServices {
		fmt.Println(service)

		segmentProperties := label.ExtractTraefikLabels(service.Labels)
		for segmentName, labels := range segmentProperties {
			service.SegmentLabels = labels
			service.SegmentName = segmentName

			frontendName := p.getFrontendName(service)
			frontends[frontendName] = service
			backendName := getBackendName(service)
			backends[backendName] = service
		}
	}

	templateObjects := struct {
		Frontends map[string]rancherData
		Backends  map[string]rancherData
		Domain    string
	}{
		Frontends: frontends,
		Backends:  backends,
		Domain:    p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/rancher.tmpl", RancherFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func (p *Provider) serviceFilter(service rancherData) bool {
	segmentProperties := label.ExtractTraefikLabels(service.Labels)

	for segmentName, labels := range segmentProperties {
		_, err := checkSegmentPort(labels, segmentName)
		if err != nil {
			log.Debugf("Filtering service %s %s without traefik.port label", service.Name, segmentName)
			return false
		}

		if len(p.getFrontendRule(service)) == 0 {
			log.Debugf("Filtering container with empty frontend rule %s %s", service.Name, segmentName)
			return false
		}
	}

	if !label.IsEnabled(service.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled service %s", service.Name)
		return false
	}

	constraintTags := label.GetSliceStringValue(service.Labels, label.TraefikTags)
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Filtering service %s with constraint %s", service.Name, failingConstraint.String())
		}
		return false
	}

	// Only filter services by Health (HealthState) and State if EnableServiceHealthFilter is true
	if p.EnableServiceHealthFilter {

		if service.Health != "" && service.Health != healthy && service.Health != updatingHealthy {
			log.Debugf("Filtering service %s with healthState of %s", service.Name, service.Health)
			return false
		}
		if service.State != "" && service.State != active && service.State != updatingActive && service.State != upgraded && service.State != upgrading {
			log.Debugf("Filtering service %s with state of %s", service.Name, service.State)
			return false
		}
	}

	return true
}

func (p *Provider) getFrontendRule(service rancherData) string {
	defaultRule := "Host:" + strings.ToLower(strings.Replace(service.Name, "/", ".", -1)) + "." + p.Domain
	return label.GetStringValue(service.Labels, label.TraefikFrontendRule, defaultRule)
}

func (p *Provider) getFrontendName(service rancherData) string {
	var name string
	if len(service.SegmentName) > 0 {
		name = getBackendName(service)
	} else {
		name = p.getFrontendRule(service)
	}

	return provider.Normalize(name)
}

func getBackendName(service rancherData) string {
	if len(service.SegmentName) > 0 {
		return getSegmentBackendName(service)
	}

	return getDefaultBackendName(service)
}

func getSegmentBackendName(service rancherData) string {
	if value := label.GetStringValue(service.SegmentLabels, label.TraefikFrontendBackend, ""); len(value) > 0 {
		return provider.Normalize(service.Name + "-" + value)
	}

	return provider.Normalize(service.Name + "-" + getDefaultBackendName(service) + "-" + service.SegmentName)
}

func getDefaultBackendName(service rancherData) string {
	backend := label.GetStringValue(service.SegmentLabels, label.TraefikBackend, service.Name)
	return provider.Normalize(backend)
}

func getServers(service rancherData) map[string]types.Server {
	var servers map[string]types.Server

	for index, ip := range service.Containers {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(service.SegmentLabels, label.TraefikProtocol, label.DefaultProtocol)
		port := label.GetStringValue(service.SegmentLabels, label.TraefikPort, "")
		weight := label.GetIntValue(service.SegmentLabels, label.TraefikWeight, label.DefaultWeightInt)

		serverName := "server-" + strconv.Itoa(index)
		servers[serverName] = types.Server{
			URL:    fmt.Sprintf("%s://%s:%s", protocol, ip, port),
			Weight: weight,
		}
	}

	return servers
}

func checkSegmentPort(labels map[string]string, segmentName string) (int, error) {
	if rawPort, ok := labels[label.TraefikPort]; ok {
		port, err := strconv.Atoi(rawPort)
		if err != nil {
			return port, fmt.Errorf("invalid port value %q for the segment %q: %v", rawPort, segmentName, err)
		}
	} else {
		return 0, fmt.Errorf("port label is missing, please use %s as default value or define port label for all segments ('traefik.<segment_name>.port')", label.TraefikPort)
	}
	return 0, nil
}
