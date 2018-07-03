// Package tracer contains Datadog's core tracing client. It is used to trace
// requests as they flow across web servers, databases and microservices, giving
// developers visibility into bottlenecks and troublesome requests. To start the
// tracer, simply call the start method along with an optional set of options.
// By default, the trace agent is considered to be found at "localhost:8126". In a
// setup where this would be different (let's say 127.0.0.1:1234), we could do:
// 	tracer.Start(tracer.WithAgentAddr("127.0.0.1:1234"))
// 	defer tracer.Stop()
//
// The tracing client can perform trace sampling. While the trace agent
// already samples traces to reduce bandwidth usage, client sampling reduces
// performance overhead. To make use of it, the package comes with a ready-to-use
// rate sampler that can be passed to the tracer. To use it and keep only 30% of the
// requests, one would do:
//   s := tracer.NewRateSampler(0.3)
//   tracer.Start(tracer.WithSampler(s))
//
// All spans created by the tracer contain a context hereby referred to as the span
// context. Note that this is different from Go's context. The span context is used
// to package essential information from a span, which is needed when creating child
// spans that inherit from it. Thus, a child span is created from a span's span context.
// The span context can originate from within the same process, but also a
// different process or even a different machine in the case of distributed tracing.
//
// To make use of distributed tracing, a span's context may be injected via a carrier
// into a transport (HTTP, RPC, etc.) to be extracted on the other end and used to
// create spans that are direct descendants of it. A couple of carrier interfaces
// which should cover most of the use-case scenarios are readily provided, such as
// HTTPCarrier and TextMapCarrier. Users are free to create their own, which will work
// with our propagation algorithm as long as they implement the TextMapReader and TextMapWriter
// interfaces. An example alternate implementation is the MDCarrier in our gRPC integration.
//
// As an example, injecting a span's context into an HTTP request would look like this:
//  req, err := http.NewRequest("GET", "http://example.com", nil)
//  // ...
//  err := tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.Header))
//  // ...
//  http.DefaultClient.Do(req)
// Then, on the server side, to continue the trace one would do:
//  sctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(req.Header))
//  // ...
//  span := tracer.StartSpan("child.span", tracer.ChildOf(sctx))
// In the same manner, any means can be used as a carrier to inject a context into a transport. Go's
// context can also be used as a means to transport spans within the same process. The methods
// StartSpanFromContext, ContextWithSpan and SpanFromContext exist for this reason.
//
// Some libraries and frameworks are supported out-of-the-box by using one
// of our integrations. You can see a list of supported integrations here:
// https://godoc.org/gopkg.in/DataDog/dd-trace-go.v1/contrib
package tracer // import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
