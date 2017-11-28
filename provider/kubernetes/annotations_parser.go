package kubernetes

import (
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/types"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func getBoolAnnotation(meta *v1beta1.Ingress, name string, defaultValue bool) bool {
	annotationValue := defaultValue
	annotationStringValue, ok := meta.Annotations[name]
	switch {
	case !ok:
		// No op.
	case annotationStringValue == "false":
		annotationValue = false
	case annotationStringValue == "true":
		annotationValue = true
	default:
		log.Warnf("Unknown value '%s' for %s, falling back to %s", name, types.LabelFrontendPassTLSCert, defaultValue)
	}
	return annotationValue
}

func getStringAnnotation(meta *v1beta1.Ingress, name string) string {
	value := meta.Annotations[name]
	return value
}

func getSliceAnnotation(meta *v1beta1.Ingress, name string) []string {
	var value []string
	if annotation, ok := meta.Annotations[name]; ok && annotation != "" {
		value = provider.SplitAndTrimString(annotation)
	}
	if len(value) == 0 {
		log.Debugf("Could not load %v annotation, skipping...", name)
		return nil
	}
	return value
}

func getMapAnnotation(meta *v1beta1.Ingress, name string) map[string]string {
	value := make(map[string]string)
	if annotation := meta.Annotations[name]; annotation != "" {
		for _, v := range strings.Split(annotation, ",") {
			pair := strings.Split(v, ":")
			if len(pair) != 2 {
				log.Debugf("Could not load annotation (%v) with value: %v, skipping...", name, pair)
			} else {
				value[pair[0]] = pair[1]
			}
		}
	}
	if len(value) == 0 {
		log.Debugf("Could not load %v annotation, skipping...", name)
		return nil
	}
	return value
}
