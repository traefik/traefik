# Tracing

Visualize the Requests Flow
{: .subtitle }

The tracing system allows developers to visualize call flows in their infrastructure.

Traefik uses OpenTracing, an open standard designed for distributed tracing.

Traefik supports five tracing backends: [Jaeger](./tracing.md#jaeger), [Zipkin](./tracing.md#zipkin), [DataDog](./tracing.md#datadog), [Instana](./tracing.md#instana), and [Haystack](./tracing.md#haystack).

## Configuration

By default, Traefik uses Jaeger as tracing backend.

To enable the tracing:

```toml tab="File"
[tracing]
```

```bash tab="CLI"
--tracing
```

### Common Options

#### `serviceName`

_Required, Default="traefik"_

Service name used in selected backend.

```toml tab="File"
[tracing]
  serviceName = "traefik"
```

```bash tab="CLI"
--tracing
--tracing.serviceName="traefik"
```

#### `spanNameLimit`

_Required, Default=0_

Span name limit allows for name truncation in case of very long names.
This can prevent certain tracing providers to drop traces that exceed their length limits.

`0` means no truncation will occur.

```toml tab="File"
[tracing]
  spanNameLimit = 150
```

```bash tab="CLI"
--tracing
--tracing.spanNameLimit=150
```

### Jaeger

To enable the Jaeger:

```toml tab="File"
[tracing]
  [tracing.jaeger]
```

```bash tab="CLI"
--tracing
--tracing.jaeger
```

!!! warning
    Traefik is only able to send data over the compact thrift protocol to the [Jaeger agent](https://www.jaegertracing.io/docs/deployment/#agent).

#### `samplingServerURL`

_Required, Default="http://localhost:5778/sampling"_

Sampling Server URL is the address of jaeger-agent's HTTP sampling server.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    samplingServerURL = "http://localhost:5778/sampling"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.samplingServerURL="http://localhost:5778/sampling"
```

#### `samplingType`

_Required, Default="const"_

Sampling Type specifies the type of the sampler: `const`, `probabilistic`, `rateLimiting`.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    samplingType = "const"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.samplingType="const"
```

#### `samplingParam`

_Required, Default=1.0_

Sampling Param is a value passed to the sampler.

Valid values for Param field are:

- for `const` sampler, 0 or 1 for always false/true respectively
- for `probabilistic` sampler, a probability between 0 and 1
- for `rateLimiting` sampler, the number of spans per second

```toml tab="File"
[tracing]
  [tracing.jaeger]
    samplingParam = 1.0
```

```bash tab="CLI"
--tracing
--tracing.jaeger.samplingParam="1.0"
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:6831"_

Local Agent Host Port instructs reporter to send spans to jaeger-agent at this address.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    localAgentHostPort = "127.0.0.1:6831"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.localAgentHostPort="127.0.0.1:6831"
```

#### `gen128Bit`

_Optional, Default=false_

Generate 128-bit trace IDs, compatible with OpenCensus.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    gen128Bit = true
```

```bash tab="CLI"
--tracing
--tracing.jaeger.gen128Bit
```

#### `propagation`

_Required, Default="jaeger"_

Set the propagation header type.
This can be either:

- `jaeger`, jaeger's default trace header.
- `b3`, compatible with OpenZipkin

```toml tab="File"
[tracing]
  [tracing.jaeger]
    propagation = "jaeger"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.propagation="jaeger"
```

#### `traceContextHeaderName`

_Required, Default="uber-trace-id"_

Trace Context Header Name is the http header name used to propagate tracing context.
This must be in lower-case to avoid mismatches when decoding incoming headers.

```toml tab="File"
[tracing]
  [tracing.jaeger]
    traceContextHeaderName = "uber-trace-id"
```

```bash tab="CLI"
--tracing
--tracing.jaeger.traceContextHeaderName="uber-trace-id"
```

### Zipkin

To enable the Zipkin:

```toml tab="File"
[tracing]
  [tracing.zipkin]
```

```bash tab="CLI"
--tracing
--tracing.zipkin
```

#### `httpEndpoint`

_Required, Default="http://localhost:9411/api/v1/spans"_

Zipkin HTTP endpoint used to send data.

```toml tab="File"
[tracing]
  [tracing.zipkin]
    httpEndpoint = "http://localhost:9411/api/v1/spans"
```

```bash tab="CLI"
--tracing
--tracing.zipkin.httpEndpoint="http://localhost:9411/api/v1/spans"
```

#### `debug`

_Optional, Default=false_

Enable Zipkin debug.

```toml tab="File"
[tracing]
  [tracing.zipkin]
    debug = true
```

```bash tab="CLI"
--tracing
--tracing.zipkin.debug=true
```

#### `sameSpan`

_Optional, Default=false_

Use Zipkin SameSpan RPC style traces.

```toml tab="File"
[tracing]
  [tracing.zipkin]
    sameSpan = true
```

```bash tab="CLI"
--tracing
--tracing.zipkin.sameSpan=true
```

#### `id128Bit`

_Optional, Default=true_

Use Zipkin 128 bit root span IDs.

```toml tab="File"
[tracing]
  [tracing.zipkin]
    id128Bit = false
```

```bash tab="CLI"
--tracing
--tracing.zipkin.id128Bit=false
```

#### `sampleRate`

_Required, Default=1.0_

The rate between 0.0 and 1.0 of requests to trace.

```toml tab="File"
[tracing]
  [tracing.zipkin]
    sampleRate = 0.2
```

```bash tab="CLI"
--tracing
--tracing.zipkin.sampleRate="0.2"
```

### DataDog

To enable the DataDog:

```toml tab="File"
[tracing]
  [tracing.datadog]
```

```bash tab="CLI"
--tracing
--tracing.datadog
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:8126"_

Local Agent Host Port instructs reporter to send spans to datadog-tracing-agent at this address.

```toml tab="File"
[tracing]
  [tracing.datadog]
    localAgentHostPort = "127.0.0.1:8126"
```

```bash tab="CLI"
--tracing
--tracing.datadog.localAgentHostPort="127.0.0.1:8126"
```

#### `debug`

_Optional, Default=false_

Enable DataDog debug.

```toml tab="File"
[tracing]
  [tracing.datadog]
    debug = true
```

```bash tab="CLI"
--tracing
--tracing.datadog.debug=true
```

#### `globalTag`

_Optional, Default=empty_

Apply shared tag in a form of Key:Value to all the traces.

```toml tab="File"
[tracing]
  [tracing.datadog]
    globalTag = "sample"
```

```bash tab="CLI"
--tracing
--tracing.datadog.globalTag="sample"
```

#### `prioritySampling`

_Optional, Default=false_

Enable priority sampling. When using distributed tracing,
this option must be enabled in order to get all the parts of a distributed trace sampled.

```toml tab="File"
[tracing]
  [tracing.datadog]
    prioritySampling = true
```

```bash tab="CLI"
--tracing
--tracing.datadog.prioritySampling=true
```

### Instana

To enable the Instana:

```toml tab="File"
[tracing]
  [tracing.instana]
```

```bash tab="CLI"
--tracing
--tracing.instana
```

#### `localAgentHost`

_Require, Default="127.0.0.1"_

Local Agent Host instructs reporter to send spans to instana-agent at this address.

```toml tab="File"
[tracing]
  [tracing.instana]
    localAgentHost = "127.0.0.1"
```

```bash tab="CLI"
--tracing
--tracing.instana.localAgentHost="127.0.0.1"
```

#### `localAgentPort`

_Require, Default=42699_

Local Agent port instructs reporter to send spans to the instana-agent at this port.

```toml tab="File"
[tracing]
  [tracing.instana]
    localAgentPort = 42699
```

```bash tab="CLI"
--tracing
--tracing.instana.localAgentPort=42699
```

#### `logLevel`

_Require, Default="info"_

Set Instana tracer log level.

Valid values for logLevel field are:

- `error`
- `warn`
- `debug`
- `info`

```toml tab="File"
[tracing]
  [tracing.instana]
    logLevel = "info"
```

```bash tab="CLI"
--tracing
--tracing.instana.logLevel="info"
```

### Haystack

To enable the Haystack:

```toml tab="File"
[tracing]
  [tracing.haystack]
```

```bash tab="CLI"
--tracing
--tracing.haystack
```

#### `localAgentHost`

_Require, Default="127.0.0.1"_

Local Agent Host instructs reporter to send spans to haystack-agent at this address.

```toml tab="File"
[tracing]
  [tracing.haystack]
    localAgentHost = "127.0.0.1"
```

```bash tab="CLI"
--tracing
--tracing.haystack.localAgentHost="127.0.0.1"
```

#### `localAgentPort`

_Require, Default=42699_

Local Agent port instructs reporter to send spans to the haystack-agent at this port.

```toml tab="File"
[tracing]
  [tracing.haystack]
    localAgentPort = 42699
```

```bash tab="CLI"
--tracing
--tracing.haystack.localAgentPort=42699
```

#### `globalTag`

_Optional, Default=empty_

Apply shared tag in a form of Key:Value to all the traces.

```toml tab="File"
[tracing]
  [tracing.haystack]
    globalTag = "sample:test"
```

```bash tab="CLI"
--tracing
--tracing.haystack.globalTag="sample:test"
```

#### `traceIDHeaderName`

_Optional, Default=empty_

Specifies the header name that will be used to store the trace ID.

```toml tab="File"
[tracing]
  [tracing.haystack]
    traceIDHeaderName = "sample"
```

```bash tab="CLI"
--tracing
--tracing.haystack.traceIDHeaderName="sample"
```

#### `parentIDHeaderName`

_Optional, Default=empty_

Specifies the header name that will be used to store the span ID.

```toml tab="File"
[tracing]
  [tracing.haystack]
    parentIDHeaderName = "sample"
```

```bash tab="CLI"
--tracing
--tracing.haystack.parentIDHeaderName="sample"
```

#### `spanIDHeaderName`

_Optional, Default=empty_

Apply shared tag in a form of Key:Value to all the traces.

```toml tab="File"
[tracing]
  [tracing.haystack]
    spanIDHeaderName = "sample:test"
```

```bash tab="CLI"
--tracing
--tracing.haystack.spanIDHeaderName="sample:test"
```

#### `baggagePrefixHeaderName`

_Optional, Default=empty_

Specifies the header name prefix that will be used to store baggage items in a map.

```toml tab="File"
[tracing]
  [tracing.haystack]
    baggagePrefixHeaderName = "sample"
```

```bash tab="CLI"
--tracing
--tracing.haystack.baggagePrefixHeaderName="sample"
```
