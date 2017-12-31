package consul

import (
	"bytes"
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

func (p *CatalogProvider) buildConfiguration(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,

		// Backend functions
		"getBackend":              getBackend,
		"getBackendAddress":       getBackendAddress,
		"hasMaxconnAttributes":    p.hasMaxConnAttributes,
		"getSticky":               p.getSticky,
		"hasStickinessLabel":      p.hasStickinessLabel,
		"getStickinessCookieName": p.getStickinessCookieName,

		// Frontend functions
		"getBackendName":  getBackendName,
		"getFrontendRule": p.getFrontendRule,
		"getBasicAuth":    p.getBasicAuth,
		"getEntryPoints":  getEntryPoints,
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

	configuration, err := p.GetConfiguration("templates/consul_catalog.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.WithError(err).Error("Failed to create config")
	}

	return configuration
}

func (p *CatalogProvider) setupFrontEndRuleTemplate() {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,
	}
	tmpl := template.New("consul catalog frontend rule").Funcs(FuncMap)
	p.frontEndRuleTemplate = tmpl
}

// Specific functions

func (p *CatalogProvider) getFrontendRule(service serviceUpdate) string {
	customFrontendRule := p.getAttribute(label.SuffixFrontendRule, service.Attributes, "")
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

func (p *CatalogProvider) getBasicAuth(tags []string) []string {
	return p.getSliceAttribute(label.SuffixFrontendAuthBasic, tags)
}

func (p *CatalogProvider) hasMaxConnAttributes(attributes []string) bool {
	amount := p.getAttribute(label.SuffixBackendMaxConnAmount, attributes, "")
	extractorFunc := p.getAttribute(label.SuffixBackendMaxConnExtractorFunc, attributes, "")
	return amount != "" && extractorFunc != ""
}

// Deprecated
func getEntryPoints(list string) []string {
	return strings.Split(list, ",")
}

func getBackend(node *api.ServiceEntry) string {
	return strings.ToLower(node.Service.Service)
}

func getBackendAddress(node *api.ServiceEntry) string {
	if node.Service.Address != "" {
		return node.Service.Address
	}
	return node.Node.Address
}

func getBackendName(node *api.ServiceEntry, index int) string {
	serviceName := strings.ToLower(node.Service.Service) + "--" + node.Service.Address + "--" + strconv.Itoa(node.Service.Port)

	for _, tag := range node.Service.Tags {
		serviceName += "--" + provider.Normalize(tag)
	}

	serviceName = strings.Replace(serviceName, ".", "-", -1)
	serviceName = strings.Replace(serviceName, "=", "-", -1)

	// unique int at the end
	serviceName += "--" + strconv.Itoa(index)
	return serviceName
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func (p *CatalogProvider) getSticky(tags []string) string {
	stickyTag := p.getAttribute(label.SuffixBackendLoadBalancerSticky, tags, "")
	if len(stickyTag) > 0 {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	} else {
		stickyTag = "false"
	}
	return stickyTag
}

func (p *CatalogProvider) hasStickinessLabel(tags []string) bool {
	stickinessTag := p.getAttribute(label.SuffixBackendLoadBalancerStickiness, tags, "")
	return len(stickinessTag) > 0 && strings.EqualFold(strings.TrimSpace(stickinessTag), "true")
}

func (p *CatalogProvider) getStickinessCookieName(tags []string) string {
	return p.getAttribute(label.SuffixBackendLoadBalancerStickinessCookieName, tags, "")
}

// Base functions

func (p *CatalogProvider) getSliceAttribute(name string, tags []string) []string {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return nil
	}
	return label.SplitAndTrimString(rawValue, ",")
}

func (p *CatalogProvider) getBoolAttribute(name string, tags []string, defaultValue bool) bool {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", name, rawValue)
		return defaultValue
	}
	return value
}

func (p *CatalogProvider) getAttribute(name string, tags []string, defaultValue string) string {
	return getTag(p.getPrefixedName(name), tags, defaultValue)
}

func (p *CatalogProvider) getPrefixedName(name string) string {
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
