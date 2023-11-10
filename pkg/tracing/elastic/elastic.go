package elastic

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/version"
	"go.elastic.co/apm/module/apmotel/v2"
	"go.elastic.co/apm/transport"
	"go.elastic.co/apm/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Name sets the name of this tracer.
const Name = "elastic"

// Config provides configuration settings for a elastic.co tracer.
type Config struct {
	ServerURL          string            `description:"Sets the URL of the Elastic APM server." json:"serverURL,omitempty" toml:"serverURL,omitempty" yaml:"serverURL,omitempty"`
	SecretToken        string            `description:"Sets the token used to connect to Elastic APM Server." json:"secretToken,omitempty" toml:"secretToken,omitempty" yaml:"secretToken,omitempty" loggable:"false"`
	ServiceEnvironment string            `description:"Sets the name of the environment Traefik is deployed in, e.g. 'production' or 'staging'." json:"serviceEnvironment,omitempty" toml:"serviceEnvironment,omitempty" yaml:"serviceEnvironment,omitempty" export:"true"`
	GlobalTags                 map[string]string `description:"Sets a list of key:value tags on all spans." json:"globalTags,omitempty" toml:"globalTags,omitempty" yaml:"globalTags,omitempty" export:"true"`
	SampleRate         float64           `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.SampleRate = 1.0
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (trace.Tracer, io.Closer, error) {
	tr, err := transport.NewHTTPTransport()
	if err != nil {
		return nil, nil, err
	}

	if c.ServerURL != "" {
		serverURL, err := url.Parse(c.ServerURL)
		if err != nil {
			return nil, nil, err
		}
		tr.SetServerURL(serverURL)
	}

	if c.SecretToken != "" {
		tr.SetSecretToken(c.SecretToken)
	}

	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:        serviceName,
		ServiceVersion:     version.Version,
		ServiceEnvironment: c.ServiceEnvironment,
		Transport:          tr,
	})
	if err != nil {
		return nil, nil, err
	}

	attr := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(version.Version),
	}

	for k, v := range c.GlobalTags {
		attr = append(attr, attribute.String(k, v))
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(attr...),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("building resource: %w", err)
	}

	tracerProvider, err := apmotel.NewTracerProvider(
		apmotel.WithResource(res),
		apmotel.WithAPMTracer(tracer),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error newTracerProvider: %w", err)
	}
	otel.SetTracerProvider(tracerProvider)

	log.Debug().Msg("Elastic tracer configured")

	return tracerProvider.Tracer(serviceName, trace.WithInstrumentationVersion(version.Version)), nil, nil
}
