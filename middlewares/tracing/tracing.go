package tracing

import (
	"io"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/tracing/jaeger"
	"github.com/containous/traefik/middlewares/tracing/zipkin"
	opentracing "github.com/opentracing/opentracing-go"
)

// Tracing middleware
type Tracing struct {
	Backend     string         `description:"Selects the tracking backend ('jaeger','zipkin')." export:"true"`
	ServiceName string         `description:"Set the name for this service" export:"true"`
	Jaeger      *jaeger.Config `description:"Settings for jaeger" export:"true"`
	Zipkin      *zipkin.Config `description:"Settings for zipkin" export:"true"`

	opentracing.Tracer
	closer io.Closer
}

// Backend describes things we can use to setup tracing
type Backend interface {
	Setup(serviceName string) (opentracing.Tracer, io.Closer, error)
}

// Setup Tracing middleware
func (t *Tracing) Setup() {
	var err error

	switch t.Backend {
	case jaeger.Name:
		t.Tracer, t.closer, err = t.Jaeger.Setup(t.ServiceName)
	case zipkin.Name:
		t.Tracer, t.closer, err = t.Zipkin.Setup(t.ServiceName)
	default:
		log.Warnf("unknown tracer %q", t.Backend)
		return
	}

	if err != nil {
		log.Warnf("Could not initialize %s tracing: %v", err)
		return
	}

	return
}

// Close tracer
func (t *Tracing) Close() {
	if t.closer != nil {
		t.closer.Close()
	}
}
