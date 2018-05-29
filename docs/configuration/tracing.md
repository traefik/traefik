# Tracing

Tracing system allows developers to visualize call flows in there infrastructures.

We use [OpenTracing](http://opentracing.io). It is an open standard designed for distributed tracing.

Træfik supports two backends: Jaeger and Zipkin.

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
```

!!! warning
    Træfik is only able to send data over compact thrift protocol to the [Jaeger agent](https://www.jaegertracing.io/docs/deployment/#agent).

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
