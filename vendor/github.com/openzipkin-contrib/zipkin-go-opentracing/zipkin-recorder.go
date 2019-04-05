package zipkintracer

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	otext "github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	"github.com/openzipkin-contrib/zipkin-go-opentracing/flag"
	"github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

var (
	// SpanKindResource will be regarded as a SA annotation by Zipkin.
	SpanKindResource = otext.SpanKindEnum("resource")
)

// Recorder implements the SpanRecorder interface.
type Recorder struct {
	collector    Collector
	debug        bool
	endpoint     *zipkincore.Endpoint
	materializer func(logFields []log.Field) ([]byte, error)
}

// RecorderOption allows for functional options.
type RecorderOption func(r *Recorder)

// WithLogFmtMaterializer will convert OpenTracing Log fields to a LogFmt representation.
func WithLogFmtMaterializer() RecorderOption {
	return func(r *Recorder) {
		r.materializer = MaterializeWithLogFmt
	}
}

// WithJSONMaterializer will convert OpenTracing Log fields to a JSON representation.
func WithJSONMaterializer() RecorderOption {
	return func(r *Recorder) {
		r.materializer = MaterializeWithJSON
	}
}

// WithStrictMaterializer will only record event Log fields and discard the rest.
func WithStrictMaterializer() RecorderOption {
	return func(r *Recorder) {
		r.materializer = StrictZipkinMaterializer
	}
}

// NewRecorder creates a new Zipkin Recorder backed by the provided Collector.
//
// hostPort and serviceName allow you to set the default Zipkin endpoint
// information which will be added to the application's standard core
// annotations. hostPort will be resolved into an IPv4 and/or IPv6 address and
// Port number, serviceName will be used as the application's service
// identifier.
//
// If application does not listen for incoming requests or an endpoint Context
// does not involve network address and/or port these cases can be solved like
// this:
//  # port is not applicable:
//  NewRecorder(c, debug, "192.168.1.12:0", "ServiceA")
//
//  # network address and port are not applicable:
//  NewRecorder(c, debug, "0.0.0.0:0", "ServiceB")
func NewRecorder(c Collector, debug bool, hostPort, serviceName string, options ...RecorderOption) SpanRecorder {
	r := &Recorder{
		collector:    c,
		debug:        debug,
		endpoint:     makeEndpoint(hostPort, serviceName),
		materializer: MaterializeWithLogFmt,
	}
	for _, opts := range options {
		opts(r)
	}
	return r
}

// RecordSpan converts a RawSpan into the Zipkin representation of a span
// and records it to the underlying collector.
func (r *Recorder) RecordSpan(sp RawSpan) {
	if !sp.Context.Sampled {
		return
	}

	var parentSpanID *int64
	if sp.Context.ParentSpanID != nil {
		id := int64(*sp.Context.ParentSpanID)
		parentSpanID = &id
	}

	var traceIDHigh *int64
	if sp.Context.TraceID.High > 0 {
		tidh := int64(sp.Context.TraceID.High)
		traceIDHigh = &tidh
	}

	span := &zipkincore.Span{
		Name:        sp.Operation,
		ID:          int64(sp.Context.SpanID),
		TraceID:     int64(sp.Context.TraceID.Low),
		TraceIDHigh: traceIDHigh,
		ParentID:    parentSpanID,
		Debug:       r.debug || (sp.Context.Flags&flag.Debug == flag.Debug),
	}
	// only send timestamp and duration if this process owns the current span.
	if sp.Context.Owner {
		timestamp := sp.Start.UnixNano() / 1e3
		duration := sp.Duration.Nanoseconds() / 1e3
		// since we always time our spans we will round up to 1 microsecond if the
		// span took less.
		if duration == 0 {
			duration = 1
		}
		span.Timestamp = &timestamp
		span.Duration = &duration
	}
	if kind, ok := sp.Tags[string(otext.SpanKind)]; ok {
		switch kind {
		case otext.SpanKindRPCClient, otext.SpanKindRPCClientEnum:
			annotate(span, sp.Start, zipkincore.CLIENT_SEND, r.endpoint)
			annotate(span, sp.Start.Add(sp.Duration), zipkincore.CLIENT_RECV, r.endpoint)
		case otext.SpanKindRPCServer, otext.SpanKindRPCServerEnum:
			annotate(span, sp.Start, zipkincore.SERVER_RECV, r.endpoint)
			annotate(span, sp.Start.Add(sp.Duration), zipkincore.SERVER_SEND, r.endpoint)
		case SpanKindResource:
			serviceName, ok := sp.Tags[string(otext.PeerService)]
			if !ok {
				serviceName = r.endpoint.GetServiceName()
			}
			host, ok := sp.Tags[string(otext.PeerHostname)].(string)
			if !ok {
				if r.endpoint.GetIpv4() > 0 {
					ip := make([]byte, 4)
					binary.BigEndian.PutUint32(ip, uint32(r.endpoint.GetIpv4()))
					host = net.IP(ip).To4().String()
				} else {
					ip := r.endpoint.GetIpv6()
					host = net.IP(ip).String()
				}
			}
			var sPort string
			port, ok := sp.Tags[string(otext.PeerPort)]
			if !ok {
				sPort = strconv.FormatInt(int64(r.endpoint.GetPort()), 10)
			} else {
				sPort = strconv.FormatInt(int64(port.(uint16)), 10)
			}
			re := makeEndpoint(net.JoinHostPort(host, sPort), serviceName.(string))
			if re != nil {
				annotateBinary(span, zipkincore.SERVER_ADDR, serviceName, re)
			} else {
				fmt.Printf("endpoint creation failed: host: %q port: %q", host, sPort)
			}
			annotate(span, sp.Start, zipkincore.CLIENT_SEND, r.endpoint)
			annotate(span, sp.Start.Add(sp.Duration), zipkincore.CLIENT_RECV, r.endpoint)
		default:
			annotateBinary(span, zipkincore.LOCAL_COMPONENT, r.endpoint.GetServiceName(), r.endpoint)
		}
		delete(sp.Tags, string(otext.SpanKind))
	} else {
		annotateBinary(span, zipkincore.LOCAL_COMPONENT, r.endpoint.GetServiceName(), r.endpoint)
	}

	for key, value := range sp.Tags {
		annotateBinary(span, key, value, r.endpoint)
	}

	for _, spLog := range sp.Logs {
		if len(spLog.Fields) == 1 && spLog.Fields[0].Key() == "event" {
			// proper Zipkin annotation
			annotate(span, spLog.Timestamp, fmt.Sprintf("%+v", spLog.Fields[0].Value()), r.endpoint)
			continue
		}
		// OpenTracing Log with key-value pair(s). Try to materialize using the
		// materializer chosen for the recorder.
		if logs, err := r.materializer(spLog.Fields); err != nil {
			fmt.Printf("Materialization of OpenTracing LogFields failed: %+v", err)
		} else {
			annotate(span, spLog.Timestamp, string(logs), r.endpoint)
		}
	}
	_ = r.collector.Collect(span)
}

// annotate annotates the span with the given value.
func annotate(span *zipkincore.Span, timestamp time.Time, value string, host *zipkincore.Endpoint) {
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	span.Annotations = append(span.Annotations, &zipkincore.Annotation{
		Timestamp: timestamp.UnixNano() / 1e3,
		Value:     value,
		Host:      host,
	})
}

// annotateBinary annotates the span with a key and a value that will be []byte
// encoded.
func annotateBinary(span *zipkincore.Span, key string, value interface{}, host *zipkincore.Endpoint) {
	if b, ok := value.(bool); ok {
		if b {
			value = "true"
		} else {
			value = "false"
		}
	}
	span.BinaryAnnotations = append(span.BinaryAnnotations, &zipkincore.BinaryAnnotation{
		Key:            key,
		Value:          []byte(fmt.Sprintf("%+v", value)),
		AnnotationType: zipkincore.AnnotationType_STRING,
		Host:           host,
	})
}
