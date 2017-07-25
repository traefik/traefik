# Zipkin

## Development and Testing Set-up

Great efforts have been made to make [Zipkin] easier to test, develop and
experiment against. [Zipkin] can now be run from a single Docker container or by
running its self-contained executable jar without extensive configuration. In
its default configuration you will run Zipkin with a HTTP collector, In memory
Span storage backend and web UI on port 9411.

Example:
```
docker run -d -p 9411:9411 openzipkin/zipkin
```

[zipkin]: http://zipkin.io

Instrumenting your services with Zipkin distributed tracing using the default
configuration is now possible with the latest release of [zipkin-go-opentracing]
as it includes an HTTP transport for sending spans to the [Zipkin] HTTP
Collector.

## Middleware Usage

Follow the [addsvc] example to check out how to wire the Zipkin Middleware. The
changes should be relatively minor.

The [zipkin-go-opentracing] package has support for HTTP, Kafka and Scribe
collectors as well as using Go Kit's [Log] package for logging.

### Configuring for the Zipkin HTTP Collector

To select the transport for the HTTP Collector, you configure the `Recorder`
with the appropriate collector like this:

```go
var (
  debugMode          = false
  serviceName        = "MyService"
  serviceHostPort    = "localhost:8000"
  zipkinHTTPEndpoint = "localhost:9411"
)
collector, err = zipkin.NewHTTPCollector(zipkinHTTPEndpoint)
if err != nil {
  // handle error
}
tracer, err = zipkin.NewTracer(
  zipkin.NewRecorder(collector, debugMode, serviceHostPort, serviceName),
  ...
)
```

### Span per Node vs. Span per RPC
By default Zipkin V1 considers either side of an RPC to have the same identity
and differs in that respect from many other tracing systems which consider the
caller to be the parent and the receiver to be the child. The OpenTracing
specification does not dictate one model over the other, but the Zipkin team is
looking into these [single-host-spans] to potentially bring Zipkin more in-line
with the other tracing systems.

[single-host-spans]: https://github.com/openzipkin/zipkin/issues/963

In case of a `span per node` the receiver will create a child span from the
propagated parent span like this:

```
Span per Node propagation and identities

CALLER:            RECEIVER:
---------------------------------
traceId        ->  traceId
                   spanId (new)
spanId         ->  parentSpanId
parentSpanId
```

**Note:** most tracing implementations supporting the `span per node` model
therefore do not propagate their `parentSpanID` as its not needed.

A typical Zipkin implementation will use the `span per RPC` model and recreate
the span identity from the caller on the receiver's end and then annotates its
values on top of it. Propagation will happen like this:

```
Span per RPC propagation and identities

CALLER:            RECEIVER:
---------------------------------
traceId        ->  traceId
spanId         ->  spanId
parentSpanId   ->  parentSpanId
```

The [zipkin-go-opentracing] implementation allows you to choose which model you
wish to use. Make sure you select the same model consistently for all your
services that are required to communicate with each other or you will have trace
propagation issues. If using non OpenTracing / legacy instrumentation, it's
probably best to use the `span per RPC call` model.

To adhere to the more common tracing philosophy of `span per node`, the Tracer
defaults to `span per node`. To set the `span per RPC call` mode start your
tracer like this:

```go
tracer, err = zipkin.NewTracer(
	zipkin.NewRecorder(...),
	zipkin.ClientServerSameSpan(true),
)
```

[zipkin-go-opentracing]: https://github.com/openzipkin/zipkin-go-opentracing
[addsvc]:https://github.com/go-kit/kit/tree/master/examples/addsvc
[Log]: https://github.com/go-kit/kit/tree/master/log

### Tracing Resources

In our legacy implementation we had the `NewChildSpan` method to allow
annotation of resources such as databases, caches and other services that do not
have server side tracing support. Since OpenTracing has no specific method of
dealing with these items explicitely that is compatible with Zipkin's `SA`
annotation, the [zipkin-go-opentracing] has implemented support using the
OpenTracing Tags system. Here is an example of how one would be able to record
a resource span compatible with standard OpenTracing and triggering an `SA`
annotation in [zipkin-go-opentracing]:

```go
// you need to import the ext package for the Tag helper functions
import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func (svc *Service) GetMeSomeExamples(ctx context.Context, ...) ([]Examples, error) {
	// Example of annotating a database query:
	var (
		serviceName = "MySQL"
		serviceHost = "mysql.example.com"
		servicePort = uint16(3306)
		queryLabel  = "GetExamplesByParam"
		query       = "select * from example where param = 'value'"
	)

	// retrieve the parent span, if not found create a new trace
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		parentSpan = opentracing.StartSpan(queryLabel)
    defer parentSpan.Finish()
	}

	// create a new span to record the resource interaction
	span := opentracing.StartChildSpan(parentSpan, queryLabel)

	// span.kind "resource" triggers SA annotation
	ext.SpanKind.Set(span, "resource")

	// this will label the span's service & hostPort (called Endpoint in Zipkin)
	ext.PeerService.Set(span, serviceName)
	ext.PeerHostname,Set(span, serviceHost)
	ext.PeerPort.Set(span, servicePort)

	// a Tag is the equivalent of a Zipkin Binary Annotation (key:value pair)
	span.SetTag("query", query)

	// a LogEvent is the equivalent of a Zipkin Annotation (timestamped)
	span.LogEvent("query:start")

	// do the actual query...

	// let's annotate the end...
	span.LogEvent("query:end")

	// we're done with this span.
	span.Finish()

	// do other stuff
	...
}
```
