package instana

import (
	"io"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Name sets the name of this tracer.
const Name = "instana"

// Config provides configuration settings for the Instana tracer.
type Config struct {
	LocalAgentHost  string `description:"Set the Instana Agent host." json:"localAgentHost,omitempty" toml:"localAgentHost,omitempty" yaml:"localAgentHost,omitempty"`
	LocalAgentPort  int    `description:"Set the Instana Agent port." json:"localAgentPort,omitempty" toml:"localAgentPort,omitempty" yaml:"localAgentPort,omitempty"`
	LogLevel        string `description:"Set the log level for the Instana tracer. ('error','warn','info','debug')" json:"logLevel,omitempty" toml:"logLevel,omitempty" yaml:"logLevel,omitempty" export:"true"`
	EnableProfiling bool   `description:"Whether or not to enable automatic profiling for the traefik process." json:"enableProfiling,omitempty" toml:"enableProfiling,omitempty" yaml:"enableProfiling,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (c *Config) SetDefaults() {
	c.LocalAgentPort = 42699
	c.LogLevel = "info"
	c.EnableProfiling = false
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
		EnableAutoProfile: c.EnableProfiling,
	})

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("Instana tracer configured")

	return tracer, nil, nil
}
