package observability

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func EnsureUserEnvVar() error {
	if os.Getenv("USER") == "" {
		if err := os.Setenv("USER", "traefik"); err != nil {
			return fmt.Errorf("could not set USER environment variable: %w", err)
		}
	}
	return nil
}

// We only derive the base resource once so as to not repeatedly bug the OS, runtime, and k8s.
var detectedResource = sync.OnceValues(func() (*resource.Resource, error) {
	return resource.New(context.Background(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(types.K8sAttributesDetector{}),
	)
})

// NewOTelResource builds the OpenTelemetry resource. Detection runs once per process.
// serviceName, then attrs, then the OTEL_ environment variables override the detected
// resource attributes.
func NewOTelResource(ctx context.Context, serviceName string, attrs map[string]string) (*resource.Resource, error) {
	var resAttrs []attribute.KeyValue
	for k, v := range attrs {
		resAttrs = append(resAttrs, attribute.String(k, v))
	}

	base, err := detectedResource()
	if err != nil {
		return nil, fmt.Errorf("detecting resource: %w", err)
	}

	// We apply the overlay over the base to allow the user to override the service name and
	// version, as well as any other attributes set by the above detectors.
	overlay, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version.Version),
		),
		resource.WithAttributes(resAttrs...),
		// Use the environment variables to allow overriding above resource attributes.
		// We apply this last to give it the highest priority.
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("building resource overlay: %w", err)
	}

	res, err := resource.Merge(base, overlay)
	if err != nil {
		return nil, fmt.Errorf("merging resource: %w", err)
	}

	return res, nil
}
