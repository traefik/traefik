package label

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

const (
	mapEntrySeparator = "||"
	mapValueSeparator = ":"
)

// Default values
const (
	DefaultWeight                                  = "0" // TODO [breaking] use int value
	DefaultWeightInt                               = 0   // TODO rename to DefaultWeight
	DefaultProtocol                                = "http"
	DefaultPassHostHeader                          = "true" // TODO [breaking] use bool value
	DefaultPassHostHeaderBool                      = true   // TODO rename to DefaultPassHostHeader
	DefaultPassTLSCert                             = false
	DefaultFrontendPriority                        = "0" // TODO [breaking] int value
	DefaultFrontendPriorityInt                     = 0   // TODO rename to DefaultFrontendPriority
	DefaultCircuitBreakerExpression                = "NetworkErrorRatio() > 1"
	DefaultFrontendRedirectEntryPoint              = ""
	DefaultBackendLoadBalancerMethod               = "wrr"
	DefaultBackendMaxconnExtractorFunc             = "request.host"
	DefaultBackendLoadbalancerStickinessCookieName = ""
	DefaultBackendHealthCheckPort                  = 0
)

var (
	// SegmentPropertiesRegexp used to extract the name of the segment and the name of the property for this segment
	// All properties are under the format traefik.<segment_name>.frontend.*= except the port/portIndex/weight/protocol/backend directly after traefik.<segment_name>.
	SegmentPropertiesRegexp = regexp.MustCompile(`^traefik\.(?P<segment_name>.+?)\.(?P<property_name>port|portIndex|weight|protocol|backend|frontend\.(.+))$`)

	// PortRegexp used to extract the port label of the segment
	PortRegexp = regexp.MustCompile(`^traefik\.(?P<segment_name>.+?)\.port$`)

	// RegexpBaseFrontendErrorPage used to extract error pages from service's label
	RegexpBaseFrontendErrorPage = regexp.MustCompile(`^frontend\.errors\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)

	// RegexpFrontendErrorPage used to extract error pages from label
	RegexpFrontendErrorPage = regexp.MustCompile(`^traefik\.frontend\.errors\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)

	// RegexpBaseFrontendRateLimit used to extract rate limits from service's label
	RegexpBaseFrontendRateLimit = regexp.MustCompile(`^frontend\.rateLimit\.rateSet\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)

	// RegexpFrontendRateLimit used to extract rate limits from label
	RegexpFrontendRateLimit = regexp.MustCompile(`^traefik\.frontend\.rateLimit\.rateSet\.(?P<name>[^ .]+)\.(?P<field>[^ .]+)$`)
)

// SegmentPropertyValues is a map of segment properties
// an example value is: weight=42
type SegmentPropertyValues map[string]string

// SegmentProperties is a map of segment properties per segment,
// which we can get with label[segmentName][propertyName].
// It yields a property value.
type SegmentProperties map[string]SegmentPropertyValues

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

// GetBoolValueP get bool value associated to a label
func GetBoolValueP(labels *map[string]string, labelName string, defaultValue bool) bool {
	if labels == nil {
		return defaultValue
	}
	return GetBoolValue(*labels, labelName, defaultValue)
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

// HasPrefixP Check if a value is associated to a less one label with a prefix
func HasPrefixP(labels *map[string]string, prefix string) bool {
	if labels == nil {
		return false
	}
	return HasPrefix(*labels, prefix)
}

// FindSegmentSubmatch split segment labels
func FindSegmentSubmatch(name string) []string {
	matches := SegmentPropertiesRegexp.FindStringSubmatch(name)
	if matches == nil ||
		strings.HasPrefix(name, TraefikFrontend+".") ||
		strings.HasPrefix(name, TraefikBackend+".") {
		return nil
	}
	return matches
}

// ExtractServiceProperties Extract services labels
// Deprecated
func ExtractServiceProperties(labels map[string]string) SegmentProperties {
	v := make(SegmentProperties)

	for name, value := range labels {
		matches := FindSegmentSubmatch(name)
		if matches == nil {
			continue
		}

		var segmentName string
		var propertyName string
		for i, name := range SegmentPropertiesRegexp.SubexpNames() {
			if i != 0 {
				if name == "segment_name" {
					segmentName = matches[i]
				} else if name == "property_name" {
					propertyName = matches[i]
				}
			}
		}

		if _, ok := v[segmentName]; !ok {
			v[segmentName] = make(SegmentPropertyValues)
		}
		v[segmentName][propertyName] = value
	}

	return v
}

// ExtractServicePropertiesP Extract services labels
// Deprecated
func ExtractServicePropertiesP(labels *map[string]string) SegmentProperties {
	if labels == nil {
		return make(SegmentProperties)
	}
	return ExtractServiceProperties(*labels)
}

// ParseErrorPages parse error pages to create ErrorPage struct
func ParseErrorPages(labels map[string]string, labelPrefix string, labelRegex *regexp.Regexp) map[string]*types.ErrorPage {
	var errorPages map[string]*types.ErrorPage

	for lblName, value := range labels {
		if strings.HasPrefix(lblName, labelPrefix) {
			submatch := labelRegex.FindStringSubmatch(lblName)
			if len(submatch) != 3 {
				log.Errorf("Invalid page error label: %s, sub-match: %v", lblName, submatch)
				continue
			}

			if errorPages == nil {
				errorPages = make(map[string]*types.ErrorPage)
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

// ParseRateSets parse rate limits to create Rate struct
func ParseRateSets(labels map[string]string, labelPrefix string, labelRegex *regexp.Regexp) map[string]*types.Rate {
	var rateSets map[string]*types.Rate

	for lblName, rawValue := range labels {
		if strings.HasPrefix(lblName, labelPrefix) && len(rawValue) > 0 {
			submatch := labelRegex.FindStringSubmatch(lblName)
			if len(submatch) != 3 {
				log.Errorf("Invalid rate limit label: %s, sub-match: %v", lblName, submatch)
				continue
			}

			if rateSets == nil {
				rateSets = make(map[string]*types.Rate)
			}

			limitName := submatch[1]

			ep, ok := rateSets[limitName]
			if !ok {
				ep = &types.Rate{}
				rateSets[limitName] = ep
			}

			switch submatch[2] {
			case "period":
				var d flaeg.Duration
				err := d.Set(rawValue)
				if err != nil {
					log.Errorf("Unable to parse %q: %q. %v", lblName, rawValue, err)
					continue
				}
				ep.Period = d
			case "average":
				value, err := strconv.ParseInt(rawValue, 10, 64)
				if err != nil {
					log.Errorf("Unable to parse %q: %q. %v", lblName, rawValue, err)
					continue
				}
				ep.Average = value
			case "burst":
				value, err := strconv.ParseInt(rawValue, 10, 64)
				if err != nil {
					log.Errorf("Unable to parse %q: %q. %v", lblName, rawValue, err)
					continue
				}
				ep.Burst = value
			default:
				log.Errorf("Invalid rate limit label: %s", lblName)
				continue
			}
		}
	}
	return rateSets
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
// Deprecated
func GetServiceLabel(labelName, serviceName string) string {
	if len(serviceName) > 0 {
		property := strings.TrimPrefix(labelName, Prefix)
		return Prefix + serviceName + "." + property
	}
	return labelName
}

// ExtractTraefikLabels transform labels to segment labels
func ExtractTraefikLabels(originLabels map[string]string) SegmentProperties {
	allLabels := make(SegmentProperties)

	if _, ok := allLabels[""]; !ok {
		allLabels[""] = make(SegmentPropertyValues)
	}

	for name, value := range originLabels {
		if !strings.HasPrefix(name, Prefix) {
			continue
		}

		matches := FindSegmentSubmatch(name)
		if matches == nil {
			// Classic labels
			allLabels[""][name] = value
		} else {
			// segments labels
			var segmentName string
			var propertyName string
			for i, name := range SegmentPropertiesRegexp.SubexpNames() {
				if i != 0 {
					if name == "segment_name" {
						segmentName = matches[i]
					} else if name == "property_name" {
						propertyName = matches[i]
					}
				}
			}

			if _, ok := allLabels[segmentName]; !ok {
				allLabels[segmentName] = make(SegmentPropertyValues)
			}
			allLabels[segmentName][Prefix+propertyName] = value
		}
	}
	log.Debug(originLabels, allLabels)

	allLabels.mergeDefault()

	return allLabels
}

func (s SegmentProperties) mergeDefault() {
	if defaultLabels, okDefault := s[""]; okDefault {

		segmentsNames := s.GetSegmentNames()
		if len(defaultLabels) > 0 {
			for _, name := range segmentsNames {
				segmentLabels := s[name]
				for key, value := range defaultLabels {
					if _, ok := segmentLabels[key]; !ok {
						segmentLabels[key] = value
					}
				}
			}
		}

		if len(segmentsNames) > 1 {
			delete(s, "")
		}
	}
}

// GetSegmentNames get all segment names
func (s SegmentProperties) GetSegmentNames() []string {
	var names []string
	for name := range s {
		names = append(names, name)
	}
	return names
}
