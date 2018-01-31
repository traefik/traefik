package docker

import (
	"errors"
	"strconv"
	"strings"

	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

// Specific functions

// Extract rule from labels for a given service and a given docker container
func (p Provider) getServiceFrontendRule(container dockerData, serviceName string) string {
	if value, ok := getServiceLabels(container, serviceName)[label.SuffixFrontendRule]; ok {
		return value
	}
	return p.getFrontendRule(container)
}

// Check if for the given container, we find labels that are defining services
func hasServices(container dockerData) bool {
	return len(label.ExtractServiceProperties(container.Labels)) > 0
}

// Gets array of service names for a given container
func getServiceNames(container dockerData) []string {
	labelServiceProperties := label.ExtractServiceProperties(container.Labels)
	keys := make([]string, 0, len(labelServiceProperties))
	for k := range labelServiceProperties {
		keys = append(keys, k)
	}
	return keys
}

// checkServiceLabelPort checks if all service names have a port service label
// or if port container label exists for default value
func checkServiceLabelPort(container dockerData) error {
	// If port container label is present, there is a default values for all ports, use it for the check
	_, err := strconv.Atoi(container.Labels[label.TraefikPort])
	if err != nil {
		serviceLabelPorts := make(map[string]struct{})
		serviceLabels := make(map[string]struct{})
		for lbl := range container.Labels {
			// Get all port service labels
			portLabel := extractServicePort(lbl)
			if len(portLabel) > 0 {
				serviceLabelPorts[portLabel[0]] = struct{}{}
			}
			// Get only one instance of all service names from service labels
			servicesLabelNames := label.FindServiceSubmatch(lbl)

			if len(servicesLabelNames) > 0 {
				serviceLabels[strings.Split(servicesLabelNames[0], ".")[1]] = struct{}{}
			}
		}
		// If the number of service labels is different than the number of port services label
		// there is an error
		if len(serviceLabels) == len(serviceLabelPorts) {
			for labelPort := range serviceLabelPorts {
				_, err = strconv.Atoi(container.Labels[labelPort])
				if err != nil {
					break
				}
			}
		} else {
			err = errors.New("port service labels missing, please use traefik.port as default value or define all port service labels")
		}
	}
	return err
}

func extractServicePort(labelName string) []string {
	if strings.HasPrefix(labelName, label.TraefikFrontend+".") ||
		strings.HasPrefix(labelName, label.TraefikBackend+".") {
		return nil
	}

	return label.PortRegexp.FindStringSubmatch(labelName)
}

// Extract backend from labels for a given service and a given docker container
func getServiceBackendName(container dockerData, serviceName string) string {
	if value, ok := getServiceLabels(container, serviceName)[label.SuffixFrontendBackend]; ok {
		return provider.Normalize(container.ServiceName + "-" + value)
	}
	return provider.Normalize(container.ServiceName + "-" + getBackendName(container) + "-" + serviceName)
}

// Extract port from labels for a given service and a given docker container
func getServicePort(container dockerData, serviceName string) string {
	if value, ok := getServiceLabels(container, serviceName)[label.SuffixPort]; ok {
		return value
	}
	return getPort(container)
}

func getServiceRedirect(container dockerData, serviceName string) *types.Redirect {
	serviceLabels := getServiceLabels(container, serviceName)

	permanent := getServiceBoolValue(container, serviceLabels, label.SuffixFrontendRedirectPermanent, false)

	if hasStrictServiceLabel(serviceLabels, label.SuffixFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: getStrictServiceStringValue(serviceLabels, label.SuffixFrontendRedirectEntryPoint, label.DefaultFrontendRedirectEntryPoint),
			Permanent:  permanent,
		}
	}

	if hasStrictServiceLabel(serviceLabels, label.SuffixFrontendRedirectRegex) &&
		hasStrictServiceLabel(serviceLabels, label.SuffixFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       getStrictServiceStringValue(serviceLabels, label.SuffixFrontendRedirectRegex, ""),
			Replacement: getStrictServiceStringValue(serviceLabels, label.SuffixFrontendRedirectReplacement, ""),
			Permanent:   permanent,
		}
	}

	return getRedirect(container)
}

func getServiceErrorPages(container dockerData, serviceName string) map[string]*types.ErrorPage {
	serviceLabels := getServiceLabels(container, serviceName)

	if label.HasPrefix(serviceLabels, label.BaseFrontendErrorPage) {
		return label.ParseErrorPages(serviceLabels, label.BaseFrontendErrorPage, label.RegexpBaseFrontendErrorPage)
	}

	return getErrorPages(container)
}

func getServiceRateLimit(container dockerData, serviceName string) *types.RateLimit {
	serviceLabels := getServiceLabels(container, serviceName)

	if hasStrictServiceLabel(serviceLabels, label.SuffixFrontendRateLimitExtractorFunc) {
		extractorFunc := getStrictServiceStringValue(serviceLabels, label.SuffixFrontendRateLimitExtractorFunc, label.DefaultBackendMaxconnExtractorFunc)
		return &types.RateLimit{
			ExtractorFunc: extractorFunc,
			RateSet:       label.ParseRateSets(serviceLabels, label.BaseFrontendRateLimit, label.RegexpBaseFrontendRateLimit),
		}
	}

	return getRateLimit(container)
}

func getServiceHeaders(container dockerData, serviceName string) *types.Headers {
	serviceLabels := getServiceLabels(container, serviceName)

	headers := &types.Headers{
		CustomRequestHeaders:    getServiceMapValue(container, serviceLabels, serviceName, label.SuffixFrontendRequestHeaders),
		CustomResponseHeaders:   getServiceMapValue(container, serviceLabels, serviceName, label.SuffixFrontendResponseHeaders),
		SSLProxyHeaders:         getServiceMapValue(container, serviceLabels, serviceName, label.SuffixFrontendHeadersSSLProxyHeaders),
		AllowedHosts:            getServiceSliceValue(container, serviceLabels, label.SuffixFrontendHeadersAllowedHosts),
		HostsProxyHeaders:       getServiceSliceValue(container, serviceLabels, label.SuffixFrontendHeadersHostsProxyHeaders),
		STSSeconds:              getServiceInt64Value(container, serviceLabels, label.SuffixFrontendHeadersSTSSeconds, 0),
		SSLRedirect:             getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersSSLRedirect, false),
		SSLTemporaryRedirect:    getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersSSLTemporaryRedirect, false),
		STSIncludeSubdomains:    getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersSTSIncludeSubdomains, false),
		STSPreload:              getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersSTSPreload, false),
		ForceSTSHeader:          getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersForceSTSHeader, false),
		FrameDeny:               getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersFrameDeny, false),
		ContentTypeNosniff:      getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersContentTypeNosniff, false),
		BrowserXSSFilter:        getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersBrowserXSSFilter, false),
		IsDevelopment:           getServiceBoolValue(container, serviceLabels, label.SuffixFrontendHeadersIsDevelopment, false),
		SSLHost:                 getServiceStringValue(container, serviceLabels, label.SuffixFrontendHeadersSSLHost, ""),
		CustomFrameOptionsValue: getServiceStringValue(container, serviceLabels, label.SuffixFrontendHeadersCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   getServiceStringValue(container, serviceLabels, label.SuffixFrontendHeadersContentSecurityPolicy, ""),
		PublicKey:               getServiceStringValue(container, serviceLabels, label.SuffixFrontendHeadersPublicKey, ""),
		ReferrerPolicy:          getServiceStringValue(container, serviceLabels, label.SuffixFrontendHeadersReferrerPolicy, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// Service label functions

func getFuncServiceSliceStringLabel(labelSuffix string) func(container dockerData, serviceName string) []string {
	return func(container dockerData, serviceName string) []string {
		serviceLabels := getServiceLabels(container, serviceName)
		return getServiceSliceValue(container, serviceLabels, labelSuffix)
	}
}

func getFuncServiceStringLabel(labelSuffix string, defaultValue string) func(container dockerData, serviceName string) string {
	return func(container dockerData, serviceName string) string {
		serviceLabels := getServiceLabels(container, serviceName)
		return getServiceStringValue(container, serviceLabels, labelSuffix, defaultValue)
	}
}

func getFuncServiceBoolLabel(labelSuffix string, defaultValue bool) func(container dockerData, serviceName string) bool {
	return func(container dockerData, serviceName string) bool {
		serviceLabels := getServiceLabels(container, serviceName)
		return getServiceBoolValue(container, serviceLabels, labelSuffix, defaultValue)
	}
}

func getFuncServiceIntLabel(labelSuffix string, defaultValue int) func(container dockerData, serviceName string) int {
	return func(container dockerData, serviceName string) int {
		return getServiceIntLabel(container, serviceName, labelSuffix, defaultValue)
	}
}

func hasStrictServiceLabel(serviceLabels map[string]string, labelSuffix string) bool {
	value, ok := serviceLabels[labelSuffix]
	return ok && len(value) > 0
}

func getServiceStringValue(container dockerData, serviceLabels map[string]string, labelSuffix string, defaultValue string) string {
	if value, ok := serviceLabels[labelSuffix]; ok {
		return value
	}
	return label.GetStringValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

func getStrictServiceStringValue(serviceLabels map[string]string, labelSuffix string, defaultValue string) string {
	if value, ok := serviceLabels[labelSuffix]; ok {
		return value
	}
	return defaultValue
}

func getServiceMapValue(container dockerData, serviceLabels map[string]string, serviceName string, labelSuffix string) map[string]string {
	if value, ok := serviceLabels[labelSuffix]; ok {
		lblName := label.GetServiceLabel(labelSuffix, serviceName)
		return label.ParseMapValue(lblName, value)
	}
	return label.GetMapValue(container.Labels, label.Prefix+labelSuffix)
}

func getServiceSliceValue(container dockerData, serviceLabels map[string]string, labelSuffix string) []string {
	if value, ok := serviceLabels[labelSuffix]; ok {
		return label.SplitAndTrimString(value, ",")
	}
	return label.GetSliceStringValue(container.Labels, label.Prefix+labelSuffix)
}

func getServiceBoolValue(container dockerData, serviceLabels map[string]string, labelSuffix string, defaultValue bool) bool {
	if rawValue, ok := serviceLabels[labelSuffix]; ok {
		value, err := strconv.ParseBool(rawValue)
		if err == nil {
			return value
		}
	}
	return label.GetBoolValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

func getServiceIntLabel(container dockerData, serviceName string, labelSuffix string, defaultValue int) int {
	if rawValue, ok := getServiceLabels(container, serviceName)[labelSuffix]; ok {
		value, err := strconv.Atoi(rawValue)
		if err == nil {
			return value
		}
	}
	return label.GetIntValue(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

func getServiceInt64Value(container dockerData, serviceLabels map[string]string, labelSuffix string, defaultValue int64) int64 {
	if rawValue, ok := serviceLabels[labelSuffix]; ok {
		value, err := strconv.ParseInt(rawValue, 10, 64)
		if err == nil {
			return value
		}
	}
	return label.GetInt64Value(container.Labels, label.Prefix+labelSuffix, defaultValue)
}

func getServiceLabels(container dockerData, serviceName string) label.ServicePropertyValues {
	return label.ExtractServiceProperties(container.Labels)[serviceName]
}
