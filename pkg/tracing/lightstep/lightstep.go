package lightstep

import (
	"io"

	"github.com/lightstep/lightstep-tracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Name sets the name of this tracer.
const Name = "lightstep"

// Config provides configuration settings for a lightstep tracer.
type Config struct {
	ServerHost  string `description:"Set the URL of the Lightstep host." json:"serverHost,omitempty" toml:"serverHost,omitempty" yaml:"serverHost,omitempty"`
	ServerPort  int    `description:"Set the port of the Lightstep host." json:"serverPort,omitempty" toml:"serverPort,omitempty" yaml:"serverPort,omitempty"`
	Plaintext   bool   `description:"Set if the communication is plaintext to your local Lightstep host." json:"plaintext,omitempty" toml:"plaintext,omitempty" yaml:"plaintext,omitempty"`
	AccessToken string `description:"Set the token used to connect to Lightstep Server." json:"accessToken,omitempty" toml:"accessToken,omitempty" yaml:"accessToken,omitempty"`
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	var options lightstep.Options
	log.Infof("Setting up Lightstep tracer with service name %s", serviceName)

	if c.ServerHost != "" {
		options = lightstep.Options{
			AccessToken: c.AccessToken,
			Collector:   lightstep.Endpoint{Host: c.ServerHost, Port: c.ServerPort, Plaintext: c.Plaintext},
			Tags: map[string]interface{}{
				lightstep.ComponentNameKey: serviceName,
			},
		}
	} else {
		options = lightstep.Options{
			AccessToken: c.AccessToken,
			Tags: map[string]interface{}{
				lightstep.ComponentNameKey: serviceName,
			},
		}
	}

	lightStepTracer := lightstep.NewTracer(options)

	opentracing.SetGlobalTracer(lightStepTracer)

	log.WithoutContext().Debug("Lightstep tracer configured")

	return lightStepTracer, nil, nil
}
