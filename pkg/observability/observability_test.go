package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func TestNewOTelResource(t *testing.T) {
	res, err := NewOTelResource(context.Background(), "traefik", map[string]string{"foo": "bar", "host.name": "custom"})
	require.NoError(t, err)

	attrs := res.Attributes()
	assert.Contains(t, attrs, semconv.ServiceName("traefik"))
	assert.Contains(t, attrs, semconv.ServiceVersion(version.Version))
	assert.Contains(t, attrs, attribute.String("foo", "bar"))
	assert.Contains(t, res.Attributes(), semconv.HostName("custom"))
}

func TestNewOTelResource_envOverridesAttributes(t *testing.T) {
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=from-env,foo=from-env")

	res, err := NewOTelResource(context.Background(), "traefik", map[string]string{"foo": "bar"})
	require.NoError(t, err)

	attrs := res.Attributes()
	assert.Contains(t, attrs, semconv.ServiceName("from-env"))
	assert.Contains(t, attrs, attribute.String("foo", "from-env"))
}
