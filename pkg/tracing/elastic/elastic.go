package elastic

import (
	"io"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/version"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmot"
	"go.elastic.co/apm/transport"
)

// Name sets the name of this tracer.
const Name = "elastic"

func init() {
	// The APM lib uses the init() function to create a default tracer.
	// So this default tracer must be disabled.
	// https://github.com/elastic/apm-agent-go/blob/8dd383d0d21776faad8841fe110f35633d199a03/tracer.go#L61-L65
	apm.DefaultTracer.Close()
}

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
