package instana

import (
	"io"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Name sets the name of this tracer.
const Name = "instana"

// Config provides configuration settings for an instana tracer.
type Config struct {
	LocalAgentHost    string `description:"Sets the Instana Agent host." json:"localAgentHost,omitempty" toml:"localAgentHost,omitempty" yaml:"localAgentHost,omitempty"`
	LocalAgentPort    int    `description:"Sets the Instana Agent port." json:"localAgentPort,omitempty" toml:"localAgentPort,omitempty" yaml:"localAgentPort,omitempty"`
	LogLevel          string `description:"Sets the log level for the Instana tracer. ('error','warn','info','debug')" json:"logLevel,omitempty" toml:"logLevel,omitempty" yaml:"logLevel,omitempty" export:"true"`
	EnableAutoProfile bool   `description:"Enables automatic profiling for the Traefik process." json:"enableAutoProfile,omitempty" toml:"enableAutoProfile,omitempty" yaml:"enableAutoProfile,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.LocalAgentPort = 42699
	c.LogLevel = "info"
}

// Setup sets up the tracer.
func (c *Config) Setup(serviceName string) (opentracing.Tracer, io.Closer, error) {
	// set default logLevel
	logLevel := instana.Info

	// check/set logLevel overrides
	switch c.LogLevel {
	case "error":
		logLevel = instana.Error
	case "warn":
		logLevel = instana.Warn
	case "debug":
		logLevel = instana.Debug
	}

	tracer := instana.NewTracerWithOptions(&instana.Options{
		Service:           serviceName,
		LogLevel:          logLevel,
		AgentPort:         c.LocalAgentPort,
		AgentHost:         c.LocalAgentHost,
		EnableAutoProfile: c.EnableAutoProfile,
	})

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("Instana tracer configured")

	return tracer, nil, nil
}
