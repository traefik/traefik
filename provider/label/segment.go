package label

import (
	"regexp"
	"strings"

	"github.com/containous/traefik/log"
)

var (
	// SegmentPropertiesRegexp used to extract the name of the segment and the name of the property for this segment
	// All properties are under the format traefik.<segment_name>.frontend.*= except the port/portIndex/weight/protocol/backend directly after traefik.<segment_name>.
	SegmentPropertiesRegexp = regexp.MustCompile(`^traefik\.(?P<segment_name>.+?)\.(?P<property_name>port|portIndex|portName|weight|protocol|backend|frontend\.(.+))$`)

	// PortRegexp used to extract the port label of the segment
	PortRegexp = regexp.MustCompile(`^traefik\.(?P<segment_name>.+?)\.port$`)
)

// SegmentPropertyValues is a map of segment properties
// an example value is: weight=42
type SegmentPropertyValues map[string]string

// SegmentProperties is a map of segment properties per segment,
// which we can get with label[segmentName][propertyName].
// It yields a property value.
type SegmentProperties map[string]SegmentPropertyValues

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

// ExtractServicePropertiesP Extract services labels
// Deprecated
func ExtractServicePropertiesP(labels *map[string]string) SegmentProperties {
	if labels == nil {
		return make(SegmentProperties)
	}
	return ExtractServiceProperties(*labels)
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
			// the group 0 is anonymous because it's always the root expression
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
				// the group 0 is anonymous because it's always the root expression
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

	log.Debug("originLabels", originLabels)
	log.Debug("allLabels", allLabels)

	allLabels.mergeDefault()

	return allLabels
}

func (s SegmentProperties) mergeDefault() {
	// if SegmentProperties contains the default segment, merge each segments with the default segment
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
