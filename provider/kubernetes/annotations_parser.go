package kubernetes

import (
	"strings"

	"github.com/containous/traefik/log"
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
		log.Warnf("Unknown value %q for %q, falling back to %v", annotationStringValue, name, defaultValue)
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
		value = types.SplitAndTrimString(annotation)
	}
	if len(value) == 0 {
		log.Debugf("Could not load %v annotation, skipping...", name)
		return nil
	}
	return value
}

func getMapAnnotation(meta *v1beta1.Ingress, annotName string) map[string]string {
	if values, ok := meta.Annotations[annotName]; ok {

		if len(values) == 0 {
			log.Errorf("Missing value for annotation %q", annotName)
			return nil
		}

		mapValue := make(map[string]string)
		for _, parts := range strings.Split(values, ",") {
			pair := strings.Split(parts, ":")
			if len(pair) != 2 {
				log.Warnf("Could not load %q: %v, skipping...", annotName, pair)
			} else {
				mapValue[pair[0]] = pair[1]
			}
		}

		if len(mapValue) == 0 {
			log.Errorf("Could not load %q, skipping...", annotName)
			return nil
		}
		return mapValue
	}

	return nil
}
