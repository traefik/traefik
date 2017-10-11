package tracing

import (
	"io"

	"github.com/containous/traefik/log"
	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	jaegermet "github.com/uber/jaeger-lib/metrics"
)

// Tracing middleware
type Tracing struct {
	Enable bool `description:"Enable opentracing." export:"true"`
	opentracing.Tracer

	closer io.Closer
}

// Setup Tracing middleware
func (t *Tracing) Setup() {
	var err error
	jcfg := jaegercfg.Configuration{}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := jaegermet.NullFactory
	t.closer, err = jcfg.InitGlobalTracer(
		"traefik",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Warnf("Could not initialize jaeger tracer: %v", err)
		return
	}
	log.Debugf("jaeger tracer configured", err)
	t.Tracer = opentracing.GlobalTracer()
}

// Close tracer
func (t *Tracing) Close() {
	t.closer.Close()
}
