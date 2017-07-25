# package tracing

`package tracing` provides [Dapper][]-style request tracing to services.

## Rationale

Request tracing is a fundamental building block for large distributed
applications. It's instrumental in understanding request flows, identifying
hot spots, and diagnosing errors. All microservice infrastructures will
benefit from request tracing; sufficiently large infrastructures will require
it.

## OpenTracing

Go kit builds on top of the [OpenTracing] API and uses the [opentracing-go]
package to provide tracing middlewares for its servers and clients. Currently
`kit/transport/http` and `kit/transport/grpc` transports are supported.

Since [OpenTracing] is an upcoming standard API, Go kit should support a
multitude of tracing backends. If a Tracer implementation in Go for your
back-end exists, it should work out of the box. The following tracing back-ends
are known to work with Go kit through the OpenTracing interface and are
highlighted in the [addsvc] example.


### LightStep

[LightStep] support is available through their standard Go package
[lightstep-tracer-go].

### AppDash

[Appdash] support is available straight from their system repository in the
[appdash/opentracing] directory.

### Zipkin

[Zipkin] support is now available from the [zipkin-go-opentracing] package which
can be found at the [Open Zipkin GitHub] page. This means our old custom
`tracing/zipkin` package is now deprecated. In the `kit/tracing/zipkin`
directory you can still find the `docker-compose` script to bootstrap a Zipkin
development environment and a [README] detailing how to transition from the
old package to the new.

[Dapper]: http://research.google.com/pubs/pub36356.html
[addsvc]:https://github.com/go-kit/kit/tree/master/examples/addsvc
[README]: https://github.com/go-kit/kit/blob/master/tracing/zipkin/README.md

[OpenTracing]: http://opentracing.io
[opentracing-go]: https://github.com/opentracing/opentracing-go

[Zipkin]: http://zipkin.io/
[Open Zipkin GitHub]: https://github.com/openzipkin
[zipkin-go-opentracing]: https://github.com/openzipkin/zipkin-go-opentracing

[Appdash]: https://github.com/sourcegraph/appdash
[appdash/opentracing]: https://github.com/sourcegraph/appdash/tree/master/opentracing

[LightStep]: http://lightstep.com/
[lightstep-tracer-go]: https://github.com/lightstep/lightstep-tracer-go
