package label

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

const (
	mapEntrySeparator = "||"
	mapValueSeparator = ":"
)

// Default values
const (
	DefaultWeight                                  = "0"
	DefaultProtocol                                = "http"
	DefaultPassHostHeader                          = "true"
	DefaultPassTLSCert                             = false
	DefaultFrontendPriority                        = "0"
	DefaultCircuitBreakerExpression                = "NetworkErrorRatio() > 1"
	DefaultFrontendRedirectEntryPoint              = ""
	DefaultBackendLoadBalancerMethod               = "wrr"
	DefaultBackendMaxconnExtractorFunc             = "request.host"
	DefaultBackendLoadbalancerStickinessCookieName = ""
	DefaultBackendHealthCheckPort                  = 0
)

var (
	// ServicesPropertiesRegexp used to extract the name of the service and the name of the property for this service
	// All properties are under the format traefik.<servicename>.frontend.*= except the port/portIndex/weight/protocol/backend directly after traefik.<servicename>.
	ServicesPropertiesRegexp = regexp.MustCompile(`^traefik\.(?P<service_name>.+?)\.(?P<property_name>port|portIndex|weight|protocol|backend|frontend\.(.+))$`)

	// PortRegexp used to extract the port label of the service
	PortRegexp = regexp.MustCompile(`^traefik\.(?P<service_name>.+?)\.port$`)

	// RegexpBaseFrontendErrorPage used to extract error pages from service's label
	RegexpBaseFrontendErrorPage = regexp.MustCompile(`^frontend\.errors\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)

	// RegexpFrontendErrorPage used to extract error pages from label
	RegexpFrontendErrorPage = regexp.MustCompile(`^traefik\.frontend\.errors\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)
)

// ServicePropertyValues is a map of services properties
// an example value is: weight=42
type ServicePropertyValues map[string]string

// ServiceProperties is a map of service properties per service,
// which we can get with label[serviceName][propertyName].
// It yields a property value.
type ServiceProperties map[string]ServicePropertyValues

// GetStringValue get string value associated to a label
func GetStringValue(labels map[string]string, labelName string, defaultValue string) string {
	if value, ok := labels[labelName]; ok && len(value) > 0 {
		return value
	}
	return defaultValue
}

// GetStringValueP get string value associated to a label
func GetStringValueP(labels *map[string]string, labelName string, defaultValue string) string {
	if labels == nil {
		return defaultValue
	}
	return GetStringValue(*labels, labelName, defaultValue)
}

// GetBoolValue get bool value associated to a label
func GetBoolValue(labels map[string]string, labelName string, defaultValue bool) bool {
	rawValue, ok := labels[labelName]
	if ok {
		v, err := strconv.ParseBool(rawValue)
		if err == nil {
			return v
		}
	}
	return defaultValue
}

// GetIntValue get int value associated to a label
func GetIntValue(labels map[string]string, labelName string, defaultValue int) int {
	if rawValue, ok := labels[labelName]; ok {
		value, err := strconv.Atoi(rawValue)
		if err == nil {
			return value
		}
		log.Errorf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
	}
	return defaultValue
}

// GetIntValueP get int value associated to a label
func GetIntValueP(labels *map[string]string, labelName string, defaultValue int) int {
	if labels == nil {
		return defaultValue
	}
	return GetIntValue(*labels, labelName, defaultValue)
}

// GetInt64Value get int64 value associated to a label
func GetInt64Value(labels map[string]string, labelName string, defaultValue int64) int64 {
	if rawValue, ok := labels[labelName]; ok {
		value, err := strconv.ParseInt(rawValue, 10, 64)
		if err == nil {
			return value
		}
		log.Errorf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
	}
	return defaultValue
}

// GetInt64ValueP get int64 value associated to a label
func GetInt64ValueP(labels *map[string]string, labelName string, defaultValue int64) int64 {
	if labels == nil {
		return defaultValue
	}
	return GetInt64Value(*labels, labelName, defaultValue)
}

// GetSliceStringValue get a slice of string associated to a label
func GetSliceStringValue(labels map[string]string, labelName string) []string {
	var value []string

	if values, ok := labels[labelName]; ok {
		value = SplitAndTrimString(values, ",")

		if len(value) == 0 {
			log.Debugf("Could not load %q.", labelName)
		}
	}
	return value
}

// GetSliceStringValueP get a slice of string value associated to a label
func GetSliceStringValueP(labels *map[string]string, labelName string) []string {
	if labels == nil {
		return nil
	}
	return GetSliceStringValue(*labels, labelName)
}

// ParseMapValue get Map value for a label value
func ParseMapValue(labelName, values string) map[string]string {
	mapValue := make(map[string]string)

	for _, parts := range strings.Split(values, mapEntrySeparator) {
		pair := strings.SplitN(parts, mapValueSeparator, 2)
		if len(pair) != 2 {
			log.Warnf("Could not load %q: %q, skipping...", labelName, parts)
		} else {
			mapValue[http.CanonicalHeaderKey(strings.TrimSpace(pair[0]))] = strings.TrimSpace(pair[1])
		}
	}

	if len(mapValue) == 0 {
		log.Errorf("Could not load %q, skipping...", labelName)
		return nil
	}
	return mapValue
}

// GetMapValue get Map value associated to a label
func GetMapValue(labels map[string]string, labelName string) map[string]string {
	if values, ok := labels[labelName]; ok {

		if len(values) == 0 {
			log.Errorf("Missing value for %q, skipping...", labelName)
			return nil
		}

		return ParseMapValue(labelName, values)
	}

	return nil
}

// GetStringMultipleStrict get multiple string values associated to several labels
// Fail if one label is missing
func GetStringMultipleStrict(labels map[string]string, labelNames ...string) (map[string]string, error) {
	foundLabels := map[string]string{}
	for _, name := range labelNames {
		value := GetStringValue(labels, name, "")
		// Error out only if one of them is not defined.
		if len(value) == 0 {
			return nil, fmt.Errorf("label not found: %s", name)
		}
		foundLabels[name] = value
	}
	return foundLabels, nil
}

// Has Check if a value is associated to a label
func Has(labels map[string]string, labelName string) bool {
	value, ok := labels[labelName]
	return ok && len(value) > 0
}

// HasP Check if a value is associated to a label
func HasP(labels *map[string]string, labelName string) bool {
	if labels == nil {
		return false
	}
	return Has(*labels, labelName)
}

// HasPrefix Check if a value is associated to a less one label with a prefix
func HasPrefix(labels map[string]string, prefix string) bool {
	for name, value := range labels {
		if strings.HasPrefix(name, prefix) && len(value) > 0 {
			return true
		}
	}
	return false
}

// ExtractServiceProperties Extract services labels
func ExtractServiceProperties(labels map[string]string) ServiceProperties {
	v := make(ServiceProperties)

	for name, value := range labels {
		matches := ServicesPropertiesRegexp.FindStringSubmatch(name)
		if matches == nil || strings.HasPrefix(name, TraefikFrontend) {
			continue
		}

		var serviceName string
		var propertyName string
		for i, name := range ServicesPropertiesRegexp.SubexpNames() {
			if i != 0 {
				if name == "service_name" {
					serviceName = matches[i]
				} else if name == "property_name" {
					propertyName = matches[i]
				}
			}
		}

		if _, ok := v[serviceName]; !ok {
			v[serviceName] = make(ServicePropertyValues)
		}
		v[serviceName][propertyName] = value
	}

	return v
}

// ExtractServicePropertiesP Extract services labels
func ExtractServicePropertiesP(labels *map[string]string) ServiceProperties {
	if labels == nil {
		return make(ServiceProperties)
	}
	return ExtractServiceProperties(*labels)
}

// ParseErrorPages parse error pages to create ErrorPage struct
func ParseErrorPages(labels map[string]string, labelPrefix string, labelRegex *regexp.Regexp) map[string]*types.ErrorPage {
	errorPages := make(map[string]*types.ErrorPage)

	for lblName, value := range labels {
		if strings.HasPrefix(lblName, labelPrefix) {
			submatch := labelRegex.FindStringSubmatch(lblName)
			if len(submatch) != 3 {
				log.Errorf("Invalid page error label: %s, sub-match: %v", lblName, submatch)
				continue
			}

			pageName := submatch[1]

			ep, ok := errorPages[pageName]
			if !ok {
				ep = &types.ErrorPage{}
				errorPages[pageName] = ep
			}

			switch submatch[2] {
			case SuffixErrorPageStatus:
				ep.Status = SplitAndTrimString(value, ",")
			case SuffixErrorPageQuery:
				ep.Query = value
			case SuffixErrorPageBackend:
				ep.Backend = value
			default:
				log.Errorf("Invalid page error label: %s", lblName)
				continue
			}
		}
	}

	return errorPages
}

// IsEnabled Check if a container is enabled in Træfik
func IsEnabled(labels map[string]string, exposedByDefault bool) bool {
	return GetBoolValue(labels, TraefikEnable, exposedByDefault)
}

// IsEnabledP Check if a container is enabled in Træfik
func IsEnabledP(labels *map[string]string, exposedByDefault bool) bool {
	if labels == nil {
		return exposedByDefault
	}
	return IsEnabled(*labels, exposedByDefault)
}

// SplitAndTrimString splits separatedString at the separator character and trims each
// piece, filtering out empty pieces. Returns the list of pieces or nil if the input
// did not contain a non-empty piece.
func SplitAndTrimString(base string, sep string) []string {
	var trimmedStrings []string

	for _, s := range strings.Split(base, sep) {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			trimmedStrings = append(trimmedStrings, s)
		}
	}

	return trimmedStrings
}

// GetServiceLabel converts a key value of Label*, given a serviceName,
// into a pattern <LabelPrefix>.<serviceName>.<property>
// i.e. For LabelFrontendRule and serviceName=app it will return "traefik.app.frontend.rule"
func GetServiceLabel(labelName, serviceName string) string {
	if len(serviceName) > 0 {
		property := strings.TrimPrefix(labelName, Prefix)
		return Prefix + serviceName + "." + property
	}
	return labelName
}
