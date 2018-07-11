package consulcatalog

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

func (p *Provider) buildConfigurationV2(catalog []catalogUpdate) *types.Configuration {
	var funcMap = template.FuncMap{
		"getAttribute":    p.getAttribute,
		"getTag":          getTag,
		"hasTag":          hasTag,
		"getChildNames":   getChildNames,
		"hasTree":         hasTree,
		"getPrefixedName": p.getPrefixedName,

		// Backend functions
		"getNodeBackendName":    getNodeBackendName,
		"getServiceBackendName": getServiceBackendName,
		"getBackendAddress":     getBackendAddress,
		"getServerName":         getServerName,
		"getCircuitBreaker":     getCircuitBreaker,
		"getLoadBalancer":       getLoadBalancer,
		"getMaxConn":            label.GetMaxConn,
		"getHealthCheck":        label.GetHealthCheck,
		"getBuffering":          label.GetBuffering,
		"getServer":             p.getServer,

		// Frontend functions
		"getFrontendRule":        p.getFrontendRule,
		"getBasicAuth":           label.GetFuncSliceString(label.TraefikFrontendAuthBasic), // Deprecated
		"getAuth":                label.GetAuth,
		"getFrontEndEntryPoints": label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getPriority":            label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":      label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":         label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getWhiteList":           label.GetWhiteList,
		"getRedirect":            label.GetRedirect,
		"getErrorPages":          label.GetErrorPages,
		"getRateLimit":           label.GetRateLimit,
		"getHeaders":             label.GetHeaders,
		"getFrontendMap":         getFrontendMap,
	}

	var allNodes []*api.ServiceEntry
	var services []*serviceUpdate
	for _, info := range catalog {
		if len(info.Nodes) > 0 {
			services = append(services, info.Service)
			allNodes = append(allNodes, info.Nodes...)
		}
	}
	// Ensure a stable ordering of nodes so that identical configurations may be detected
	sort.Sort(nodeSorter(allNodes))

	templateObjects := struct {
		Services []*serviceUpdate
		Nodes    []*api.ServiceEntry
	}{
		Services: services,
		Nodes:    allNodes,
	}

	configuration, err := p.GetConfiguration("templates/consul_catalog.tmpl", funcMap, templateObjects)
	if err != nil {
		log.WithError(err).Error("Failed to create config")
	}

	return configuration
}

// Specific functions

func (p *Provider) getFrontendRule(service serviceUpdate, labels map[string]string) string {
	customFrontendRule := label.GetStringValue(labels, label.TraefikFrontendRule, "")
	if customFrontendRule == "" {
		customFrontendRule = p.FrontEndRule
	}

	tmpl := p.frontEndRuleTemplate
	tmpl, err := tmpl.Parse(customFrontendRule)
	if err != nil {
		log.Errorf("Failed to parse Consul Catalog custom frontend rule: %v", err)
		return ""
	}

	templateObjects := struct {
		ServiceName string
		Domain      string
		Attributes  []string
	}{
		ServiceName: service.ServiceName,
		Domain:      p.Domain,
		Attributes:  service.Attributes,
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		log.Errorf("Failed to execute Consul Catalog custom frontend rule template: %v", err)
		return ""
	}

	return buffer.String()
}

func (p *Provider) getServer(node *api.ServiceEntry) types.Server {
	scheme := p.getAttribute(label.SuffixProtocol, node.Service.Tags, label.DefaultProtocol)
	address := getBackendAddress(node)

	return types.Server{
		URL:    fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(address, strconv.Itoa(node.Service.Port))),
		Weight: p.getWeight(node.Service.Tags),
	}
}

func (p *Provider) setupFrontEndRuleTemplate() {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,
	}
	p.frontEndRuleTemplate = template.New("consul catalog frontend rule").Funcs(FuncMap)
}

// Specific functions

// Only for compatibility
// Deprecated
func getLoadBalancer(labels map[string]string) *types.LoadBalancer {
	if v, ok := labels[label.TraefikBackendLoadBalancer]; ok {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancer, label.TraefikBackendLoadBalancerMethod)
		if !label.Has(labels, label.TraefikBackendLoadBalancerMethod) {
			labels[label.TraefikBackendLoadBalancerMethod] = v
		}
	}

	return label.GetLoadBalancer(labels)
}

// Only for compatibility
// Deprecated
func getCircuitBreaker(labels map[string]string) *types.CircuitBreaker {
	if v, ok := labels[label.TraefikBackendCircuitBreaker]; ok {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendCircuitBreaker, label.TraefikBackendCircuitBreakerExpression)
		if !label.Has(labels, label.TraefikBackendCircuitBreakerExpression) {
			labels[label.TraefikBackendCircuitBreakerExpression] = v
		}
	}

	return label.GetCircuitBreaker(labels)
}

func getServiceBackendName(service *serviceUpdate) string {
	return strings.ToLower(service.ServiceName)
}

func getNodeBackendName(node *api.ServiceEntry) string {
	return strings.ToLower(node.Service.Service)
}

func getBackendAddress(node *api.ServiceEntry) string {
	if node.Service.Address != "" {
		return node.Service.Address
	}
	return node.Node.Address
}

func getServerName(node *api.ServiceEntry, index int) string {
	serviceName := node.Service.Service + node.Service.Address + strconv.Itoa(node.Service.Port)
	// TODO sort tags ?
	serviceName += strings.Join(node.Service.Tags, "")

	hash := sha1.New()
	_, err := hash.Write([]byte(serviceName))
	if err != nil {
		// Impossible case
		log.Error(err)
	} else {
		serviceName = base64.URLEncoding.EncodeToString(hash.Sum(nil))
	}

	// unique int at the end
	return provider.Normalize(node.Service.Service + "-" + strconv.Itoa(index) + "-" + serviceName)
}

func (p *Provider) getWeight(tags []string) int {
	weight := p.getIntAttribute(label.SuffixWeight, tags, label.DefaultWeight)

	// Deprecated
	deprecatedWeightTag := "backend." + label.SuffixWeight
	if p.hasAttribute(deprecatedWeightTag, tags) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.",
			p.getPrefixedName(deprecatedWeightTag), p.getPrefixedName(label.SuffixWeight))

		weight = p.getIntAttribute(deprecatedWeightTag, tags, label.DefaultWeight)
	}

	return weight
}

// Base functions

func (p *Provider) hasAttribute(name string, tags []string) bool {
	return hasTag(p.getPrefixedName(name), tags)
}

func (p *Provider) getAttribute(name string, tags []string, defaultValue string) string {
	return getTag(p.getPrefixedName(name), tags, defaultValue)
}

func (p *Provider) getPrefixedName(name string) string {
	if len(p.Prefix) > 0 && len(name) > 0 {
		return p.Prefix + "." + name
	}
	return name
}

func hasTag(name string, tags []string) bool {
	lowerName := strings.ToLower(name)

	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)

		// Given the nature of Consul tags, which could be either singular markers, or key=value pairs
		if strings.HasPrefix(lowerTag, lowerName+"=") || lowerTag == lowerName {
			return true
		}
	}
	return false
}

func getTag(name string, tags []string, defaultValue string) string {
	lowerName := strings.ToLower(name)

	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)

		// Given the nature of Consul tags, which could be either singular markers, or key=value pairs
		if strings.HasPrefix(lowerTag, lowerName+"=") || lowerTag == lowerName {
			// In case, where a tag might be a key=value, try to split it by the first '='
			kv := strings.SplitN(tag, "=", 2)

			// If the returned result is a key=value pair, return the 'value' component
			if len(kv) == 2 {
				return kv[1]
			}
			// If the returned result is a singular marker, return the 'key' component
			return kv[0]
		}
	}
	return defaultValue
}

func hasTree(name string, tags map[string]string) bool {
	// checks if a set of tags contains a particular tree with children
	lowerName := strings.ToLower(name)

	for tag := range tags {
		lowerTag := strings.ToLower(tag)

		if strings.HasPrefix(lowerTag, lowerName+".") {
			return true
		}
	}
	return false
}

func getChildNames(name string, tags map[string]string) []string {
	// gets the names of the children in a tree:
	// a.b, a.b.c=1 => "c"
	children := make([]string, 0)
	lowerName := strings.ToLower(name)

	for tag := range tags {
		lowerTag := strings.ToLower(tag)

		if strings.HasPrefix(lowerTag, lowerName+".") {
			// name.child.key=value -> ["name", "child.key=value"]
			result := strings.SplitN(lowerTag, lowerName+".", 2)

			if len(result) == 2 {
				// child.key=value -> ["child", "key=value"] or
				//   child=value -> ["child=value"]
				child := strings.SplitN(result[1], ".", 2)[0]
				child = strings.Split(child, "=")[0]
				children = append(children, child)
			}
		}
	}
	return children
}

func getFrontendMap(name string, tags map[string]string) map[string]string {
	// Generates a new $service.TraefikLabels for a specific frontend
	// This is a hack to trick other functions into thinking this
	//  `<prefix>.frontends.name` is actually `<prefix>.frontend`
	result := make(map[string]string)
	for key, value := range tags {
		// strip `name` from key
		result[strings.Replace(key, "frontends."+name, "frontend", 1)] = value
	}
	return result
}
