package gateway

import (
	"fmt"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/label"
)

const annotationsPrefix = "traefik.io/"

// ServiceConfig is the service's root configuration from annotations.
type ServiceConfig struct {
	Service Service `json:"service"`
}

// Service is the service's configuration from annotations.
type Service struct {
	NativeLB *bool `json:"nativeLB"`
}

func parseServiceAnnotations(annotations map[string]string) (ServiceConfig, error) {
	var svcConf ServiceConfig

	labels := convertAnnotations(annotations)
	if len(labels) == 0 {
		return svcConf, nil
	}

	if err := label.Decode(labels, &svcConf, "traefik.service."); err != nil {
		return svcConf, fmt.Errorf("decoding labels: %w", err)
	}

	return svcConf, nil
}

func convertAnnotations(annotations map[string]string) map[string]string {
	if len(annotations) == 0 {
		return nil
	}

	result := make(map[string]string)

	for key, value := range annotations {
		if !strings.HasPrefix(key, annotationsPrefix) {
			continue
		}

		newKey := strings.ReplaceAll(key, "io/", "")
		result[newKey] = value
	}

	return result
}
