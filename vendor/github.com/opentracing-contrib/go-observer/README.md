# An Observer API for OpenTracing-go Tracers

OTObserver can be used to watch the span events like StartSpan(),
SetOperationName(), SetTag() and Finish(). A need for observers
arose when information (metrics) more than just the latency information was
required for the spans, in the distributed tracers. But, there can be a lot
of metrics in different domains and adding such metrics to any one (client)
tracer breaks cross-platform compatibility. There are various ways to
avoid such issues, however, an observer pattern is cleaner and provides loose
coupling between the packages exporting metrics (on span events) and the
tracer.

This information can be in the form of hardware metrics, RPC metrics,
useful metrics exported out of the kernel or other metrics, profiled for a
span. These additional metrics can help us in getting better Root-cause
analysis. With that being said, its not just for calculation of metrics,
it can be used for anything which needs watching the span events.

## Installation and Usage

The `otobserver` package provides an API to watch span's events and define
callbacks for these events. This API would be a functional option to a
tracer constructor that passes an Observer. 3rd party packages (who want to
watch the span events) should actually implement this observer API.
To do that, first fetch the package using go get :

```
   go get -v github.com/opentracing-contrib/go-observer
```

and say :

```go
    import "github.com/opentracing-contrib/go-observer"
```

and then, define the required span event callbacks. These registered
callbacks would then be called on span events if an observer is created.
Tracer may allow registering multiple observers. Have a look at the [jaeger's observer](https://github.com/uber/jaeger-client-go/blob/master/observer.go).

With the required setup implemented in the backend tracers, packages
watching the span events need to implement the observer api defining what
they need to do for the observed span events.

## Span events

An observer registered with this api, can observe for the following four
span events :

```go
    StartSpan()
    SetOperationName()
    SetTag()
    Finish()
```

### Tradeoffs

As noble as our thoughts might be in fetching additional metrics (other than
latency) for a span using an observer, there are some overhead costs. Not all
observers need to observe all the span events, in which case, we may have
to keep some callback functions empty. In effect, we will still call these
functions, and that will incur unnecessary overhead. To know more about this
and other tradeoffs, see this [discussion](https://github.com/opentracing/opentracing-go/pull/135#discussion_r105497329).
