package consulcatalog

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

// Deprecated
func (p *Provider) buildConfigurationV1(catalog []catalogUpdate) *types.Configuration {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,

		// Backend functions
		"getBackend":              getNodeBackendName,
		"getServiceBackendName":   getServiceBackendName,
		"getBackendAddress":       getBackendAddress,
		"getBackendName":          getServerName,
		"hasMaxconnAttributes":    p.hasMaxConnAttributesV1,
		"getSticky":               p.getStickyV1,
		"hasStickinessLabel":      p.hasStickinessLabelV1,
		"getStickinessCookieName": p.getStickinessCookieNameV1,
		"getWeight":               p.getWeight,
		"getProtocol":             p.getFuncStringAttribute(label.SuffixProtocol, label.DefaultProtocol),

		// Frontend functions
		"getFrontendRule":   p.getFrontendRuleV1,
		"getBasicAuth":      p.getFuncSliceAttribute(label.SuffixFrontendAuthBasic),
		"getEntryPoints":    getEntryPointsV1,
		"getPriority":       p.getFuncIntAttribute(label.SuffixFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader": p.getFuncBoolAttribute(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":    p.getFuncBoolAttribute(label.SuffixFrontendPassTLSCert, label.DefaultPassTLSCert),
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

	configuration, err := p.GetConfiguration("templates/consul_catalog-v1.tmpl", FuncMap, templateObjects)
	if err != nil {
		log.WithError(err).Error("Failed to create config")
	}

	return configuration
}

// Specific functions

// Deprecated
func (p *Provider) getFrontendRuleV1(service serviceUpdate) string {
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

// Deprecated
func (p *Provider) hasMaxConnAttributesV1(attributes []string) bool {
	amount := p.getAttribute(label.SuffixBackendMaxConnAmount, attributes, "")
	extractorFunc := p.getAttribute(label.SuffixBackendMaxConnExtractorFunc, attributes, "")
	return amount != "" && extractorFunc != ""
}

// Deprecated
func getEntryPointsV1(list string) []string {
	return strings.Split(list, ",")
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func (p *Provider) getStickyV1(tags []string) string {
	stickyTag := p.getAttribute(label.SuffixBackendLoadBalancerSticky, tags, "")
	if len(stickyTag) > 0 {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	} else {
		stickyTag = "false"
	}
	return stickyTag
}

// Deprecated
func (p *Provider) hasStickinessLabelV1(tags []string) bool {
	stickinessTag := p.getAttribute(label.SuffixBackendLoadBalancerStickiness, tags, "")
	return len(stickinessTag) > 0 && strings.EqualFold(strings.TrimSpace(stickinessTag), "true")
}

// Deprecated
func (p *Provider) getStickinessCookieNameV1(tags []string) string {
	return p.getAttribute(label.SuffixBackendLoadBalancerStickinessCookieName, tags, "")
}

// Base functions

// Deprecated
func (p *Provider) getFuncStringAttribute(name string, defaultValue string) func(tags []string) string {
	return func(tags []string) string {
		return p.getAttribute(name, tags, defaultValue)
	}
}

// Deprecated
func (p *Provider) getFuncSliceAttribute(name string) func(tags []string) []string {
	return func(tags []string) []string {
		return p.getSliceAttribute(name, tags)
	}
}

// Deprecated
func (p *Provider) getFuncIntAttribute(name string, defaultValue int) func(tags []string) int {
	return func(tags []string) int {
		return p.getIntAttribute(name, tags, defaultValue)
	}
}

func (p *Provider) getFuncBoolAttribute(name string, defaultValue bool) func(tags []string) bool {
	return func(tags []string) bool {
		return p.getBoolAttribute(name, tags, defaultValue)
	}
}

// Deprecated
func (p *Provider) getInt64Attribute(name string, tags []string, defaultValue int64) int64 {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", name, rawValue)
		return defaultValue
	}
	return value
}

// Deprecated
func (p *Provider) getIntAttribute(name string, tags []string, defaultValue int) int {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %s: %s", name, rawValue)
		return defaultValue
	}
	return value
}

// Deprecated
func (p *Provider) getSliceAttribute(name string, tags []string) []string {
	rawValue := getTag(p.getPrefixedName(name), tags, "")

	if len(rawValue) == 0 {
		return nil
	}
	return label.SplitAndTrimString(rawValue, ",")
}

// Deprecated
func (p *Provider) getBoolAttribute(name string, tags []string, defaultValue bool) bool {
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
