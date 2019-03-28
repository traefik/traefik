package servicefabric

import (
	"errors"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	sf "github.com/jjcollinge/servicefabric"
)

func (p *Provider) buildConfiguration(services []ServiceItemExtended) (*types.Configuration, error) {
	var sfFuncMap = template.FuncMap{
		// Services
		"getServices":                getServices,
		"hasLabel":                   hasService,
		"getLabelValue":              getServiceStringLabel,
		"getLabelsWithPrefix":        getServiceLabelsWithPrefix,
		"isPrimary":                  isPrimary,
		"isStateful":                 isStateful,
		"isStateless":                isStateless,
		"isEnabled":                  getFuncBoolLabel(label.TraefikEnable, false),
		"getBackendName":             getBackendName,
		"getDefaultEndpoint":         getDefaultEndpoint,
		"getNamedEndpoint":           getNamedEndpoint,           // FIXME unused
		"getApplicationParameter":    getApplicationParameter,    // FIXME unused
		"doesAppParamContain":        doesAppParamContain,        // FIXME unused
		"filterServicesByLabelValue": filterServicesByLabelValue, // FIXME unused

		// Backend functions
		"getWeight":         getFuncServiceIntLabel(label.TraefikWeight, label.DefaultWeight),
		"getProtocol":       getFuncServiceStringLabel(label.TraefikProtocol, label.DefaultProtocol),
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,

		// Frontend Functions
		"getPriority":       getFuncServiceIntLabel(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader": getFuncServiceBoolLabel(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":    getFuncBoolLabel(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getEntryPoints":    getFuncServiceSliceStringLabel(label.TraefikFrontendEntryPoints),
		"getBasicAuth":      getFuncServiceSliceStringLabel(label.TraefikFrontendAuthBasic),
		"getFrontendRules":  getFuncServiceLabelWithPrefix(label.TraefikFrontendRule),
		"getWhiteList":      getWhiteList,
		"getHeaders":        getHeaders,
		"getRedirect":       getRedirect,
		"getErrorPages":     getErrorPages,

		// SF Service Grouping
		"getGroupedServices": getFuncServicesGroupedByLabel(traefikSFGroupName),
		"getGroupedWeight":   getFuncServiceStringLabel(traefikSFGroupWeight, "1"),
	}

	templateObjects := struct {
		Services []ServiceItemExtended
	}{
		Services: services,
	}

	return p.GetConfiguration(tmpl, sfFuncMap, templateObjects)
}

func isPrimary(instance replicaInstance) bool {
	_, data := instance.GetReplicaData()
	return data.ReplicaRole == "Primary"
}

func getBackendName(service ServiceItemExtended, partition PartitionItemExtended) string {
	return provider.Normalize(service.Name + partition.PartitionInformation.ID)
}

func getDefaultEndpoint(instance replicaInstance) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getReplicaDefaultEndpoint(data)
	if err != nil {
		log.Warnf("No default endpoint for replica %s in service %s endpointData: %s", id, data.Address, err)
		return ""
	}
	return endpoint
}

func getNamedEndpoint(instance replicaInstance, endpointName string) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getReplicaNamedEndpoint(data, endpointName)
	if err != nil {
		log.Warnf("No names endpoint of %s for replica %s in endpointData: %s. Error: %v", endpointName, id, data.Address, err)
		return ""
	}
	return endpoint
}

func getReplicaNamedEndpoint(replicaData *sf.ReplicaItemBase, endpointName string) (string, error) {
	endpoints, err := decodeEndpointData(replicaData.Address)
	if err != nil {
		return "", err
	}

	endpoint, exists := endpoints[endpointName]
	if !exists {
		return "", errors.New("endpoint doesn't exist")
	}
	return endpoint, nil
}

func getApplicationParameter(app sf.ApplicationItem, key string) string {
	for _, param := range app.Parameters {
		if param.Key == key {
			return param.Value
		}
	}
	log.Errorf("Parameter %s doesn't exist in app %s", key, app.Name)
	return ""
}

func getServices(services []ServiceItemExtended, key string) map[string][]ServiceItemExtended {
	result := map[string][]ServiceItemExtended{}
	for _, service := range services {
		if value, exists := service.Labels[key]; exists {
			if matchingServices, hasKeyAlready := result[value]; hasKeyAlready {
				result[value] = append(matchingServices, service)
			} else {
				result[value] = []ServiceItemExtended{service}
			}
		}
	}
	return result
}

func doesAppParamContain(app sf.ApplicationItem, key, shouldContain string) bool {
	value := getApplicationParameter(app, key)
	return strings.Contains(value, shouldContain)
}

func filterServicesByLabelValue(services []ServiceItemExtended, key, expectedValue string) []ServiceItemExtended {
	var srvWithLabel []ServiceItemExtended
	for _, service := range services {
		value, exists := service.Labels[key]
		if exists && value == expectedValue {
			srvWithLabel = append(srvWithLabel, service)
		}
	}
	return srvWithLabel
}

func getHeaders(service ServiceItemExtended) *types.Headers {
	return label.GetHeaders(service.Labels)
}

func getWhiteList(service ServiceItemExtended) *types.WhiteList {
	return label.GetWhiteList(service.Labels)
}

func getRedirect(service ServiceItemExtended) *types.Redirect {
	return label.GetRedirect(service.Labels)
}

func getMaxConn(service ServiceItemExtended) *types.MaxConn {
	return label.GetMaxConn(service.Labels)
}

func getHealthCheck(service ServiceItemExtended) *types.HealthCheck {
	return label.GetHealthCheck(service.Labels)
}

func getCircuitBreaker(service ServiceItemExtended) *types.CircuitBreaker {
	return label.GetCircuitBreaker(service.Labels)
}

func getLoadBalancer(service ServiceItemExtended) *types.LoadBalancer {
	return label.GetLoadBalancer(service.Labels)
}

func getErrorPages(service ServiceItemExtended) map[string]*types.ErrorPage {
	return label.GetErrorPages(service.Labels)
}
