package elastic

import (
	"io"
	"net/url"

	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/version"
	"github.com/opentracing/opentracing-go"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmot"
	"go.elastic.co/apm/transport"
)

// Name sets the name of this tracer
const Name = "elastic"

// Config provides configuration settings for a elastic.co tracer
type Config struct {
	ApmServerURL          string `description:"Set the URL of the Elastic APM server." json:"apmServerURL,omitempty" toml:"apmServerURL,omitempty" yaml:"apmServerURL,omitempty"`
	ApmSecretToken        string `description:"Set the token used to connect to Elastic APM Server." json:"ApmSecretToken,omitempty" toml:"ApmSecretToken,omitempty" yaml:"ApmSecretToken,omitempty"`
	ApmServiceEnvironment string `description:"Set the name of the environment Traefik is deployed in, e.g. 'production' or 'staging'." json:"serviceEnvironment,omitempty" toml:"serviceEnvironment,omitempty" yaml:"serviceEnvironment,omitempty"`
}

// Setup sets up the tracer
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	// Create default transport
	ht, err := transport.NewHTTPTransport()
	if err != nil {
		return nil, nil, err
	}

	if c.ApmServerURL != "" {
		apmServerURL, err := url.Parse(c.ApmServerURL)
		if err != nil {
			return nil, nil, err
		}
		ht.SetServerURL(apmServerURL)
	}

	if c.ApmSecretToken != "" {
		ht.SetSecretToken(c.ApmSecretToken)
	}

	tracer, err := apm.NewTracerOptions(apm.TracerOptions{
		ServiceName:        serviceName,
		ServiceVersion:     version.Version,
		ServiceEnvironment: c.ApmServiceEnvironment,
		Transport:          ht,
	})
	if err != nil {
		return nil, nil, err
	}

	tracer.SetLogger(log.WithoutContext())
	otracer := apmot.New(apmot.WithTracer(tracer))

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(otracer)

	log.WithoutContext().Debug("Elastic tracer configured")

	return otracer, nil, nil
}
