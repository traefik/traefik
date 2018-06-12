package ecs

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// buildConfiguration fills the config template with the given instances
func (p *Provider) buildConfigurationV2(instances []ecsInstance) (*types.Configuration, error) {
	services := make(map[string][]ecsInstance)
	for _, instance := range instances {
		if p.filterInstance(instance) {
			if serviceInstances, ok := services[instance.Name]; ok {
				services[instance.Name] = append(serviceInstances, instance)
			} else {
				services[instance.Name] = []ecsInstance{instance}
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
		"getBasicAuth":      label.GetFuncSliceString(label.TraefikFrontendAuthBasic),
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
	if labelPort := label.GetStringValue(i.TraefikLabels, label.TraefikPort, ""); len(i.container.NetworkBindings) == 0 && labelPort == "" {
		log.Debugf("Filtering ecs instance without port %s (%s)", i.Name, i.ID)
		return false
	}

	if i.machine == nil || i.machine.State == nil || i.machine.State.Name == nil {
		log.Debugf("Filtering ecs instance with missing ec2 information %s (%s)", i.Name, i.ID)
		return false
	}

	if aws.StringValue(i.machine.State.Name) != ec2.InstanceStateNameRunning {
		log.Debugf("Filtering ecs instance with an incorrect state %s (%s) (state = %s)", i.Name, i.ID, aws.StringValue(i.machine.State.Name))
		return false
	}

	if i.machine.PrivateIpAddress == nil {
		log.Debugf("Filtering ecs instance without an ip address %s (%s)", i.Name, i.ID)
		return false
	}

	if !isEnabled(i, p.ExposedByDefault) {
		log.Debugf("Filtering disabled ecs instance %s (%s)", i.Name, i.ID)
		return false
	}

	return true
}

func (p *Provider) getFrontendRule(i ecsInstance) string {
	domain := label.GetStringValue(i.TraefikLabels, label.TraefikDomain, p.Domain)
	defaultRule := "Host:" + strings.ToLower(strings.Replace(i.Name, "_", "-", -1)) + "." + domain

	return label.GetStringValue(i.TraefikLabels, label.TraefikFrontendRule, defaultRule)
}

func getHost(i ecsInstance) string {
	return aws.StringValue(i.machine.PrivateIpAddress)
}

func getPort(i ecsInstance) string {
	if value := label.GetStringValue(i.TraefikLabels, label.TraefikPort, ""); len(value) > 0 {
		return value
	}
	return strconv.FormatInt(aws.Int64Value(i.container.NetworkBindings[0].HostPort), 10)
}

func filterFrontends(instances []ecsInstance) []ecsInstance {
	byName := make(map[string]struct{})

	return fun.Filter(func(i ecsInstance) bool {
		_, found := byName[i.Name]
		if !found {
			byName[i.Name] = struct{}{}
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
