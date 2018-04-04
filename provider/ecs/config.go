package ecs

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// buildConfiguration fills the config template with the given instances
func (p *Provider) buildConfigurationV2(services map[string][]ecsInstance) (*types.Configuration, error) {
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
		"getPassHostHeader": label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":    label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPriority":       label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
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

func (p *Provider) getFrontendRule(i ecsInstance) string {
	defaultRule := "Host:" + strings.ToLower(strings.Replace(i.Name, "_", "-", -1)) + "." + p.Domain
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
			URL:    fmt.Sprintf("%s://%s:%s", protocol, host, port),
			Weight: label.GetIntValue(instance.TraefikLabels, label.TraefikWeight, 0),
		}
	}

	return servers
}

func isEnabled(i ecsInstance, exposedByDefault bool) bool {
	return label.GetBoolValue(i.TraefikLabels, label.TraefikEnable, exposedByDefault)
}
