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
  Backend = "jaeger"

  # Service name used in Jaeger backend
  #
  # Default: "traefik"
  #
  ServiceName = "traefik"

  [tracing.jaeger]
    # SamplingServerURL is the address of jaeger-agent's HTTP sampling server
    #
    # Default: "http://localhost:5778/sampling"
    #
    SamplingServerURL = "http://localhost:5778/sampling"

    # Sampling Type specifies the type of the sampler: const, probabilistic, rateLimiting
    #
    # Default: "const"
    #
    SamplingType = "const"

    # SamplingParam Param is a value passed to the sampler.
    # Valid values for Param field are:
    #   - for "const" sampler, 0 or 1 for always false/true respectively
    #   - for "probabilistic" sampler, a probability between 0 and 1
    #   - for "rateLimiting" sampler, the number of spans per second
    #
    # Default: 1.0
    #
    SamplingParam = 1.0

    # LocalAgentHostPort instructs reporter to send spans to jaeger-agent at this address
    #
    # Default: "127.0.0.1:6832"
    #
    LocalAgentHostPort = "127.0.0.1:6832"
```

## Zipkin

```toml
# Tracing definition
[tracing]
  # Backend name used to send tracing data
  #
  # Default: "jaeger"
  #
  Backend = "zipkin"

  # Service name used in Zipkin backend
  #
  # Default: "traefik"
  #
  ServiceName = "traefik"

  [tracing.zipkin]
    # Zipking HTTP endpoint used to send data
    #
    # Default: "http://localhost:9411/api/v1/spans"
    #
    HTTPEndpoint = "http://localhost:9411/api/v1/spans"

    # Enable Zipkin debug
    #
    # Default: false
    #
    Debug = false

    # Use ZipKin SameSpan RPC style traces
    #
    # Default: false
    #
    SameSpan = false

    # Use ZipKin 128 bit root span IDs
    #
    # Default: true
    #
    ID128Bit = true
```
