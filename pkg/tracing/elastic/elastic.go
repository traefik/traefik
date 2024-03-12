package elastic

import (
	"io"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/version"
	"go.elastic.co/apm/module/apmot/v2"
	"go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/transport"
)

// Config provides configuration settings for an elastic.co tracer.
type Config struct {
	ServerURL          string `description:"Sets the URL of the Elastic APM server." json:"serverURL,omitempty" toml:"serverURL,omitempty" yaml:"serverURL,omitempty"`
	SecretToken        string `description:"Sets the token used to connect to Elastic APM Server." json:"secretToken,omitempty" toml:"secretToken,omitempty" yaml:"secretToken,omitempty" loggable:"false"`
	ServiceEnvironment string `description:"Sets the name of the environment Traefik is deployed in, e.g. 'production' or 'staging'." json:"serviceEnvironment,omitempty" toml:"serviceEnvironment,omitempty" yaml:"serviceEnvironment,omitempty" export:"true"`
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	transportOptions := transport.HTTPTransportOptions{
		SecretToken: c.SecretToken,
	}

	if c.ServerURL != "" {
		serverURL, err := url.Parse(c.ServerURL)
		if err != nil {
			return nil, nil, err
		}

		transportOptions.ServerURLs = append(transportOptions.ServerURLs, serverURL)
	}

	tr, err := transport.NewHTTPTransport(transportOptions)
	if err != nil {
		return nil, nil, err
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

	tracer.SetLogger(log.WithoutContext())
	otTracer := apmot.New(apmot.WithTracer(tracer))

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(otTracer)

	log.WithoutContext().Debug("Elastic tracer configured")

	return otTracer, nil, nil
}
