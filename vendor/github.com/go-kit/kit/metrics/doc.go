// Package metrics provides a framework for application instrumentation. All
// metrics are safe for concurrent use. Considerable design influence has been
// taken from https://github.com/codahale/metrics and https://prometheus.io.
//
// This package contains the common interfaces. Your code should take these
// interfaces as parameters. Implementations are provided for different
// instrumentation systems in the various subdirectories.
//
// Usage
//
// Metrics are dependencies and should be passed to the components that need
// them in the same way you'd construct and pass a database handle, or reference
// to another component. So, create metrics in your func main, using whichever
// concrete implementation is appropriate for your organization.
//
//    latency := prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
//        Namespace: "myteam",
//        Subsystem: "foosvc",
//        Name:      "request_latency_seconds",
//        Help:      "Incoming request latency in seconds."
//    }, []string{"method", "status_code"})
//
// Write your components to take the metrics they will use as parameters to
// their constructors. Use the interface types, not the concrete types. That is,
//
//    // NewAPI takes metrics.Histogram, not *prometheus.Summary
//    func NewAPI(s Store, logger log.Logger, latency metrics.Histogram) *API {
//        // ...
//    }
//
//    func (a *API) ServeFoo(w http.ResponseWriter, r *http.Request) {
//        begin := time.Now()
//        // ...
//        a.latency.Observe(time.Since(begin).Seconds())
//    }
//
// Finally, pass the metrics as dependencies when building your object graph.
// This should happen in func main, not in the global scope.
//
//    api := NewAPI(store, logger, latency)
//    http.ListenAndServe("/", api)
//
// Implementation details
//
// Each telemetry system has different semantics for label values, push vs.
// pull, support for histograms, etc. These properties influence the design of
// their respective packages. This table attempts to summarize the key points of
// distinction.
//
//    SYSTEM      DIM  COUNTERS               GAUGES                 HISTOGRAMS
//    dogstatsd   n    batch, push-aggregate  batch, push-aggregate  native, batch, push-each
//    statsd      1    batch, push-aggregate  batch, push-aggregate  native, batch, push-each
//    graphite    1    batch, push-aggregate  batch, push-aggregate  synthetic, batch, push-aggregate
//    expvar      1    atomic                 atomic                 synthetic, batch, in-place expose
//    influx      n    custom                 custom                 custom
//    prometheus  n    native                 native                 native
//    circonus    1    native                 native                 native
//    pcp         1    native                 native                 native
//
package metrics
