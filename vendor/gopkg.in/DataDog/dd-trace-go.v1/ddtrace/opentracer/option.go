package opentracer // import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"

	opentracing "github.com/opentracing/opentracing-go"
)

// ServiceName can be used with opentracing.StartSpan to set the
// service name of a span.
func ServiceName(name string) opentracing.StartSpanOption {
	return opentracing.Tag{Key: ext.ServiceName, Value: name}
}

// ResourceName can be used with opentracing.StartSpan to set the
// resource name of a span.
func ResourceName(name string) opentracing.StartSpanOption {
	return opentracing.Tag{Key: ext.ResourceName, Value: name}
}

// SpanName sets the Datadog operation name for the span.
func SpanName(name string) opentracing.StartSpanOption {
	return opentracing.Tag{Key: ext.SpanName, Value: name}
}

// SpanType can be used with opentracing.StartSpan to set the type of a span.
func SpanType(name string) opentracing.StartSpanOption {
	return opentracing.Tag{Key: ext.SpanType, Value: name}
}
