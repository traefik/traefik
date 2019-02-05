# Tracing

The tracing system allows developers to visualize call flows in their infrastructure.

We use [OpenTracing](http://opentracing.io). It is an open standard designed for distributed tracing.

Traefik supports three tracing backends: Jaeger, Zipkin and DataDog.

## Jaeger

```toml
# Tracing definition
[tracing]
  # Backend name used to send tracing data
  #
  # Default: "jaeger"
  #
  backend = "jaeger"

  # Service name used in Jaeger backend
  #
  # Default: "traefik"
  #
  serviceName = "traefik"
    
  # Span name limit allows for name truncation in case of very long Frontend/Backend names
  # This can prevent certain tracing providers to drop traces that exceed their length limits
  #
  # Default: 0 - no truncation will occur
  # 
  spanNameLimit = 0

  [tracing.jaeger]
    # Sampling Server URL is the address of jaeger-agent's HTTP sampling server
    #
    # Default: "http://localhost:5778/sampling"
    #
    samplingServerURL = "http://localhost:5778/sampling"

    # Sampling Type specifies the type of the sampler: const, probabilistic, rateLimiting
    #
    # Default: "const"
    #
    samplingType = "const"

    # Sampling Param is a value passed to the sampler.
    # Valid values for Param field are:
    #   - for "const" sampler, 0 or 1 for always false/true respectively
    #   - for "probabilistic" sampler, a probability between 0 and 1
    #   - for "rateLimiting" sampler, the number of spans per second
    #
    # Default: 1.0
    #
    samplingParam = 1.0

    # Local Agent Host Port instructs reporter to send spans to jaeger-agent at this address
    #
    # Default: "127.0.0.1:6831"
    #
    localAgentHostPort = "127.0.0.1:6831"
    
    # Trace Context Header Name is the http header name used to propagate tracing context.
    # This must be in lower-case to avoid mismatches when decoding incoming headers.
    #
    # Default: "uber-trace-id"
    #
    traceContextHeaderName = "uber-trace-id"
```

!!! warning
    Traefik is only able to send data over compact thrift protocol to the [Jaeger agent](https://www.jaegertracing.io/docs/deployment/#agent).

## Zipkin

```toml
# Tracing definition
[tracing]
  # Backend name used to send tracing data
  #
  # Default: "jaeger"
  #
  backend = "zipkin"

  # Service name used in Zipkin backend
  #
  # Default: "traefik"
  #
  serviceName = "traefik"
    
  # Span name limit allows for name truncation in case of very long Frontend/Backend names
  # This can prevent certain tracing providers to drop traces that exceed their length limits
  #
  # Default: 0 - no truncation will occur
  # 
  spanNameLimit = 150

  [tracing.zipkin]
    # Zipking HTTP endpoint used to send data
    #
    # Default: "http://localhost:9411/api/v1/spans"
    #
    httpEndpoint = "http://localhost:9411/api/v1/spans"

    # Enable Zipkin debug
    #
    # Default: false
    #
    debug = false

    # Use ZipKin SameSpan RPC style traces
    #
    # Default: false
    #
    sameSpan = false

    # Use ZipKin 128 bit root span IDs
    #
    # Default: true
    #
    id128Bit = true
```

## DataDog

```toml
# Tracing definition
[tracing]
  # Backend name used to send tracing data
  #
  # Default: "jaeger"
  #
  backend = "datadog"

  # Service name used in DataDog backend
  #
  # Default: "traefik"
  #
  serviceName = "traefik"
  
  # Span name limit allows for name truncation in case of very long Frontend/Backend names
  # This can prevent certain tracing providers to drop traces that exceed their length limits
  #
  # Default: 0 - no truncation will occur
  # 
  spanNameLimit = 100

  [tracing.datadog]
    # Local Agent Host Port instructs reporter to send spans to datadog-tracing-agent at this address
    #
    # Default: "127.0.0.1:8126"
    #
    localAgentHostPort = "127.0.0.1:8126"

    # Enable DataDog debug
    #
    # Default: false
    #
    debug = false

    # Apply shared tag in a form of Key:Value to all the traces
    #
    # Default: ""
    #
    globalTag = ""

    # Enable priority sampling. When using distributed tracing, this option must be enabled in order
    # to get all the parts of a distributed trace sampled.
    #
    # Default: false
    #
    prioritySampling = false
```
