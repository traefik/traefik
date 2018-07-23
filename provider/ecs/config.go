package ecs

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// buildConfiguration fills the config template with the given instances
func (p *Provider) buildConfiguration(instances []ecsInstance) (*types.Configuration, error) {
	services := make(map[string][]ecsInstance)
	for _, instance := range instances {
		backendName := getBackendName(instance)
		if p.filterInstance(instance) {
			if serviceInstances, ok := services[backendName]; ok {
				services[backendName] = append(serviceInstances, instance)
			} else {
				services[backendName] = []ecsInstance{instance}
			}
		}
	}

	var ecsFuncMap = template.FuncMap{
		// Backend functions
		"getHost":           getHost,
		"getPort":           getPort,
		"getCircuitBreaker": label.GetCircuitBreaker,
		"getLoadBalancer":   label.GetLoadBalancer,
		"getMaxConn":        label.GetMaxConn,
		"getHealthCheck":    label.GetHealthCheck,
		"getBuffering":      label.GetBuffering,
		"getServers":        getServers,

		// Frontend functions
		"filterFrontends":   filterFrontends,
		"getFrontendRule":   p.getFrontendRule,
		"getPassHostHeader": label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":    label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":       label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getBasicAuth":      label.GetFuncSliceString(label.TraefikFrontendAuthBasic), // Deprecated
		"getAuth":           label.GetAuth,
		"getEntryPoints":    label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getRedirect":       label.GetRedirect,
		"getErrorPages":     label.GetErrorPages,
		"getRateLimit":      label.GetRateLimit,
		"getHeaders":        label.GetHeaders,
		"getWhiteList":      label.GetWhiteList,
	}

	return p.GetConfiguration("templates/ecs.tmpl", ecsFuncMap, struct {
		Services map[string][]ecsInstance
	}{
		Services: services,
	})
}

func (p *Provider) filterInstance(i ecsInstance) bool {
	if i.machine == nil {
		log.Debug("Filtering ecs instance with nil machine")
		return false
	}

	if labelPort := label.GetStringValue(i.TraefikLabels, label.TraefikPort, ""); len(i.machine.ports) == 0 && labelPort == "" {
		log.Debugf("Filtering ecs instance without port %s (%s)", i.Name, i.ID)
		return false
	}

	if strings.ToLower(i.machine.state) != ec2.InstanceStateNameRunning {
		log.Debugf("Filtering ecs instance with an incorrect state %s (%s) (state = %s)", i.Name, i.ID, i.machine.state)
		return false
	}

	if len(i.machine.privateIP) == 0 {
		log.Debugf("Filtering ecs instance without an ip address %s (%s)", i.Name, i.ID)
		return false
	}

	if !isEnabled(i, p.ExposedByDefault) {
		log.Debugf("Filtering disabled ecs instance %s (%s)", i.Name, i.ID)
		return false
	}

	constraintTags := label.GetSliceStringValue(i.TraefikLabels, label.TraefikTags)
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Filtering ecs instance pruned by constraint %s (%s) (constraint = %q)", i.Name, i.ID, failingConstraint.String())
		}
		return false
	}

	return true
}

func getBackendName(i ecsInstance) string {
	if value := label.GetStringValue(i.TraefikLabels, label.TraefikBackend, ""); len(value) > 0 {
		return value
	}
	return i.Name
}

func (p *Provider) getFrontendRule(i ecsInstance) string {
	domain := label.GetStringValue(i.TraefikLabels, label.TraefikDomain, p.Domain)
	defaultRule := "Host:" + strings.ToLower(strings.Replace(i.Name, "_", "-", -1)) + "." + domain

	return label.GetStringValue(i.TraefikLabels, label.TraefikFrontendRule, defaultRule)
}

func getHost(i ecsInstance) string {
	return i.machine.privateIP
}

func getPort(i ecsInstance) string {
	if value := label.GetStringValue(i.TraefikLabels, label.TraefikPort, ""); len(value) > 0 {
		port, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			for _, mapping := range i.machine.ports {
				if port == mapping.hostPort || port == mapping.containerPort {
					return strconv.FormatInt(mapping.hostPort, 10)
				}
			}
			return value
		}
	}
	return strconv.FormatInt(i.machine.ports[0].hostPort, 10)
}

func filterFrontends(instances []ecsInstance) []ecsInstance {
	byName := make(map[string]struct{})

	return fun.Filter(func(i ecsInstance) bool {
		backendName := getBackendName(i)
		_, found := byName[backendName]
		if !found {
			byName[backendName] = struct{}{}
		}
		return !found
	}, instances).([]ecsInstance)
}

func getServers(instances []ecsInstance) map[string]types.Server {
	var servers map[string]types.Server

	for _, instance := range instances {
		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(instance.TraefikLabels, label.TraefikProtocol, label.DefaultProtocol)
		host := getHost(instance)
		port := getPort(instance)

		serverName := provider.Normalize(fmt.Sprintf("server-%s-%s", instance.Name, instance.ID))
		servers[serverName] = types.Server{
			URL:    fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(host, port)),
			Weight: label.GetIntValue(instance.TraefikLabels, label.TraefikWeight, label.DefaultWeight),
		}
	}

	return servers
}

func isEnabled(i ecsInstance, exposedByDefault bool) bool {
	return label.GetBoolValue(i.TraefikLabels, label.TraefikEnable, exposedByDefault)
}
