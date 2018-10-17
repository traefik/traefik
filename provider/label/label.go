package label

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/containous/traefik/log"
)

const (
	mapEntrySeparator = "||"
	mapValueSeparator = ":"
)

// Default values
const (
	DefaultWeight                                  = 1
	DefaultProtocol                                = "http"
	DefaultPassHostHeader                          = true
	DefaultPassTLSCert                             = false
	DefaultFrontendPriority                        = 0
	DefaultCircuitBreakerExpression                = "NetworkErrorRatio() > 1"
	DefaultBackendLoadBalancerMethod               = "wrr"
	DefaultBackendMaxconnExtractorFunc             = "request.host"
	DefaultBackendLoadbalancerStickinessCookieName = ""
	DefaultBackendHealthCheckPort                  = 0
)

var (
	// RegexpFrontendErrorPage used to extract error pages from label
	RegexpFrontendErrorPage = regexp.MustCompile(`^traefik\.frontend\.errors\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)

	// RegexpFrontendRateLimit used to extract rate limits from label
	RegexpFrontendRateLimit = regexp.MustCompile(`^traefik\.frontend\.rateLimit\.rateSet\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)
)

// GetStringValue get string value associated to a label
func GetStringValue(labels map[string]string, labelName string, defaultValue string) string {
	if value, ok := labels[labelName]; ok && len(value) > 0 {
		return value
	}
	return defaultValue
}

// GetBoolValue get bool value associated to a label
func GetBoolValue(labels map[string]string, labelName string, defaultValue bool) bool {
	rawValue, ok := labels[labelName]
	if ok {
		v, err := strconv.ParseBool(rawValue)
		if err == nil {
			return v
		}
		log.Errorf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
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

// HasPrefix Check if a value is associated to a less one label with a prefix
func HasPrefix(labels map[string]string, prefix string) bool {
	for name, value := range labels {
		if strings.HasPrefix(name, prefix) && len(value) > 0 {
			return true
		}
	}
	return false
}

// IsEnabled Check if a container is enabled in Traefik
func IsEnabled(labels map[string]string, exposedByDefault bool) bool {
	return GetBoolValue(labels, TraefikEnable, exposedByDefault)
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

// GetFuncString a func related to GetStringValue
func GetFuncString(labelName string, defaultValue string) func(map[string]string) string {
	return func(labels map[string]string) string {
		return GetStringValue(labels, labelName, defaultValue)
	}
}

// GetFuncInt a func related to GetIntValue
func GetFuncInt(labelName string, defaultValue int) func(map[string]string) int {
	return func(labels map[string]string) int {
		return GetIntValue(labels, labelName, defaultValue)
	}
}

// GetFuncBool a func related to GetBoolValue
func GetFuncBool(labelName string, defaultValue bool) func(map[string]string) bool {
	return func(labels map[string]string) bool {
		return GetBoolValue(labels, labelName, defaultValue)
	}
}

// GetFuncSliceString a func related to GetSliceStringValue
func GetFuncSliceString(labelName string) func(map[string]string) []string {
	return func(labels map[string]string) []string {
		return GetSliceStringValue(labels, labelName)
	}
}
