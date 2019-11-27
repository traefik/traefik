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

// Name sets the name of this tracer.
const Name = "elastic"

// Config provides configuration settings for a elastic.co tracer.
type Config struct {
	ServerURL          string `description:"Set the URL of the Elastic APM server." json:"serverURL,omitempty" toml:"serverURL,omitempty" yaml:"serverURL,omitempty"`
	SecretToken        string `description:"Set the token used to connect to Elastic APM Server." json:"secretToken,omitempty" toml:"secretToken,omitempty" yaml:"secretToken,omitempty"`
	ServiceEnvironment string `description:"Set the name of the environment Traefik is deployed in, e.g. 'production' or 'staging'." json:"serviceEnvironment,omitempty" toml:"serviceEnvironment,omitempty" yaml:"serviceEnvironment,omitempty"`
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	// Create default transport.
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

	tracer.SetLogger(log.WithoutContext())
	otTracer := apmot.New(apmot.WithTracer(tracer))

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(otTracer)

	log.WithoutContext().Debug("Elastic tracer configured")

	return otTracer, nil, nil
}
