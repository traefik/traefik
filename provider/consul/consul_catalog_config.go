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
		"getBackend":              getBackend,
		"getFrontendRule":         p.getFrontendRule,
		"getBackendName":          getBackendName,
		"getBackendAddress":       getBackendAddress,
		"getBasicAuth":            p.getBasicAuth,
		"getSticky":               p.getSticky,
		"hasStickinessLabel":      p.hasStickinessLabel,
		"getStickinessCookieName": p.getStickinessCookieName,
		"getAttribute":            p.getAttribute,
		"getTag":                  getTag,
		"hasTag":                  hasTag,
		"getEntryPoints":          getEntryPoints,
		"hasMaxconnAttributes":    p.hasMaxConnAttributes,
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

func (p *CatalogProvider) setupFrontEndTemplate() {
	var FuncMap = template.FuncMap{
		"getAttribute": p.getAttribute,
		"getTag":       getTag,
		"hasTag":       hasTag,
	}
	tmpl := template.New("consul catalog frontend rule").Funcs(FuncMap)
	p.frontEndRuleTemplate = tmpl
}

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
	list := p.getAttribute(label.SuffixFrontendAuthBasic, tags, "")
	if list != "" {
		return strings.Split(list, ",")
	}
	return []string{}
}

func (p *CatalogProvider) hasMaxConnAttributes(attributes []string) bool {
	amount := p.getAttribute(label.SuffixBackendMaxConnAmount, attributes, "")
	extractorfunc := p.getAttribute(label.SuffixBackendMaxConnExtractorFunc, attributes, "")
	return amount != "" && extractorfunc != ""
}

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
// Deprecated replaced by Stickiness
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
