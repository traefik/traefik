package servicefabric

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	sf "github.com/jjcollinge/servicefabric"
)

func (p *Provider) buildConfiguration(sfClient sfClient) (*types.Configuration, error) {
	services, err := getClusterServices(sfClient)
	if err != nil {
		return nil, err
	}

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
		"getWeight":         getFuncServiceIntLabel(label.TraefikWeight, label.DefaultWeightInt),
		"getProtocol":       getFuncServiceStringLabel(label.TraefikProtocol, label.DefaultProtocol),
		"getMaxConn":        getMaxConn,
		"getHealthCheck":    getHealthCheck,
		"getCircuitBreaker": getCircuitBreaker,
		"getLoadBalancer":   getLoadBalancer,

		// Frontend Functions
		"getPriority":       getFuncServiceIntLabel(label.TraefikFrontendPriority, label.DefaultFrontendPriorityInt),
		"getPassHostHeader": getFuncServiceBoolLabel(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeaderBool),
		"getPassTLSCert":    getFuncBoolLabel(label.TraefikFrontendPassTLSCert, false),
		"getEntryPoints":    getFuncServiceSliceStringLabel(label.TraefikFrontendEntryPoints),
		"getBasicAuth":      getFuncServiceSliceStringLabel(label.TraefikFrontendAuthBasic),
		"getFrontendRules":  getFuncServiceLabelWithPrefix(label.TraefikFrontendRule),
		"getWhiteList":      getWhiteList,
		"getHeaders":        getHeaders,
		"getRedirect":       getRedirect,

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

func isStateful(service ServiceItemExtended) bool {
	return service.ServiceKind == "Stateful"
}

func isStateless(service ServiceItemExtended) bool {
	return service.ServiceKind == "Stateless"
}

func getBackendName(service ServiceItemExtended, partition PartitionItemExtended) string {
	return provider.Normalize(service.Name + partition.PartitionInformation.ID)
}

func getDefaultEndpoint(instance replicaInstance) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getReplicaDefaultEndpoint(data)
	if err != nil {
		log.Warnf("No default endpoint for replica %s in service %s endpointData: %s", id, data.Address)
		return ""
	}
	return endpoint
}

func getReplicaDefaultEndpoint(replicaData *sf.ReplicaItemBase) (string, error) {
	endpoints, err := decodeEndpointData(replicaData.Address)
	if err != nil {
		return "", err
	}

	var defaultHTTPEndpoint string
	for _, v := range endpoints {
		if strings.Contains(v, "http") {
			defaultHTTPEndpoint = v
			break
		}
	}

	if len(defaultHTTPEndpoint) == 0 {
		return "", errors.New("no default endpoint found")
	}
	return defaultHTTPEndpoint, nil
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

func decodeEndpointData(endpointData string) (map[string]string, error) {
	var endpointsMap map[string]map[string]string

	if endpointData == "" {
		return nil, errors.New("endpoint data is empty")
	}

	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		return nil, err
	}

	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
		return nil, errors.New("endpoint doesn't exist in endpoint data")
	}

	return endpoints, nil
}

func getHeaders(service ServiceItemExtended) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    label.GetMapValue(service.Labels, label.TraefikFrontendRequestHeaders),
		CustomResponseHeaders:   label.GetMapValue(service.Labels, label.TraefikFrontendResponseHeaders),
		SSLProxyHeaders:         label.GetMapValue(service.Labels, label.TraefikFrontendSSLProxyHeaders),
		AllowedHosts:            label.GetSliceStringValue(service.Labels, label.TraefikFrontendAllowedHosts),
		HostsProxyHeaders:       label.GetSliceStringValue(service.Labels, label.TraefikFrontendHostsProxyHeaders),
		STSSeconds:              label.GetInt64Value(service.Labels, label.TraefikFrontendSTSSeconds, 0),
		SSLRedirect:             label.GetBoolValue(service.Labels, label.TraefikFrontendSSLRedirect, false),
		SSLTemporaryRedirect:    label.GetBoolValue(service.Labels, label.TraefikFrontendSSLTemporaryRedirect, false),
		STSIncludeSubdomains:    label.GetBoolValue(service.Labels, label.TraefikFrontendSTSIncludeSubdomains, false),
		STSPreload:              label.GetBoolValue(service.Labels, label.TraefikFrontendSTSPreload, false),
		ForceSTSHeader:          label.GetBoolValue(service.Labels, label.TraefikFrontendForceSTSHeader, false),
		FrameDeny:               label.GetBoolValue(service.Labels, label.TraefikFrontendFrameDeny, false),
		ContentTypeNosniff:      label.GetBoolValue(service.Labels, label.TraefikFrontendContentTypeNosniff, false),
		BrowserXSSFilter:        label.GetBoolValue(service.Labels, label.TraefikFrontendBrowserXSSFilter, false),
		IsDevelopment:           label.GetBoolValue(service.Labels, label.TraefikFrontendIsDevelopment, false),
		SSLHost:                 label.GetStringValue(service.Labels, label.TraefikFrontendSSLHost, ""),
		CustomFrameOptionsValue: label.GetStringValue(service.Labels, label.TraefikFrontendCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   label.GetStringValue(service.Labels, label.TraefikFrontendContentSecurityPolicy, ""),
		PublicKey:               label.GetStringValue(service.Labels, label.TraefikFrontendPublicKey, ""),
		ReferrerPolicy:          label.GetStringValue(service.Labels, label.TraefikFrontendReferrerPolicy, ""),
		CustomBrowserXSSValue:   label.GetStringValue(service.Labels, label.TraefikFrontendCustomBrowserXSSValue, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

func getWhiteList(service ServiceItemExtended) *types.WhiteList {
	if label.Has(service.Labels, label.TraefikFrontendWhitelistSourceRange) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikFrontendWhitelistSourceRange, label.TraefikFrontendWhiteListSourceRange)
	}

	ranges := label.GetSliceStringValue(service.Labels, label.TraefikFrontendWhiteListSourceRange)
	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: label.GetBoolValue(service.Labels, label.TraefikFrontendWhiteListUseXForwardedFor, false),
		}
	}

	// TODO: Deprecated
	values := label.GetSliceStringValue(service.Labels, label.TraefikFrontendWhitelistSourceRange)
	if len(values) > 0 {
		return &types.WhiteList{
			SourceRange:      values,
			UseXForwardedFor: false,
		}
	}

	return nil
}

func getRedirect(service ServiceItemExtended) *types.Redirect {
	permanent := label.GetBoolValue(service.Labels, label.TraefikFrontendRedirectPermanent, false)

	if label.Has(service.Labels, label.TraefikFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: label.GetStringValue(service.Labels, label.TraefikFrontendRedirectEntryPoint, ""),
			Permanent:  permanent,
		}
	}

	if label.Has(service.Labels, label.TraefikFrontendRedirectRegex) &&
		label.Has(service.Labels, label.TraefikFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       label.GetStringValue(service.Labels, label.TraefikFrontendRedirectRegex, ""),
			Replacement: label.GetStringValue(service.Labels, label.TraefikFrontendRedirectReplacement, ""),
			Permanent:   permanent,
		}
	}

	return nil
}

func getMaxConn(service ServiceItemExtended) *types.MaxConn {
	amount := label.GetInt64Value(service.Labels, label.TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := label.GetStringValue(service.Labels, label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func getHealthCheck(service ServiceItemExtended) *types.HealthCheck {
	path := label.GetStringValue(service.Labels, label.TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	port := label.GetIntValue(service.Labels, label.TraefikBackendHealthCheckPort, label.DefaultBackendHealthCheckPort)
	interval := label.GetStringValue(service.Labels, label.TraefikBackendHealthCheckInterval, "")

	return &types.HealthCheck{
		Path:     path,
		Port:     port,
		Interval: interval,
	}
}

func getCircuitBreaker(service ServiceItemExtended) *types.CircuitBreaker {
	circuitBreaker := label.GetStringValue(service.Labels, label.TraefikBackendCircuitBreakerExpression, "")
	if len(circuitBreaker) == 0 {
		return nil
	}
	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func getLoadBalancer(service ServiceItemExtended) *types.LoadBalancer {
	if !label.HasPrefix(service.Labels, label.TraefikBackendLoadBalancer) {
		return nil
	}

	method := label.GetStringValue(service.Labels, label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: getSticky(service),
	}

	if label.GetBoolValue(service.Labels, label.TraefikBackendLoadBalancerStickiness, false) {
		cookieName := label.GetStringValue(service.Labels, label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getSticky(service ServiceItemExtended) bool {
	if label.Has(service.Labels, label.TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", label.TraefikBackendLoadBalancerSticky, label.TraefikBackendLoadBalancerStickiness)
	}

	return label.GetBoolValue(service.Labels, label.TraefikBackendLoadBalancerSticky, false)
}
