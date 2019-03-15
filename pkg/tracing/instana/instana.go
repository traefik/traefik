package instana

import (
	"io"

	"github.com/containous/traefik/pkg/log"
	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
)

// Name sets the name of this tracer
const Name = "instana"

// Config provides configuration settings for a instana tracer
type Config struct {
	LocalAgentHost string `description:"Set instana-agent's host that the reporter will used." export:"false"`
	LocalAgentPort int    `description:"Set instana-agent's port that the reporter will used." export:"false"`
	LogLevel       string `description:"Set instana-agent's log level. ('error','warn','info','debug')" export:"false"`
}

// Setup sets up the tracer
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
		Service:   serviceName,
		LogLevel:  logLevel,
		AgentPort: c.LocalAgentPort,
		AgentHost: c.LocalAgentHost,
	})

	// Without this, child spans are getting the NOOP tracer
	opentracing.SetGlobalTracer(tracer)

	log.WithoutContext().Debug("Instana tracer configured")

	return tracer, nil, nil
}
