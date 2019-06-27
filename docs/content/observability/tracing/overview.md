# Tracing

Visualize the Requests Flow
{: .subtitle }

The tracing system allows developers to visualize call flows in their infrastructure.

Traefik uses OpenTracing, an open standard designed for distributed tracing.

Traefik supports five tracing backends:

- [Jaeger](./jaeger.md)
- [Zipkin](./zipkin.md)
- [DataDog](./datadog.md)
- [Instana](./instana.md)
- [Haystack](./haystack.md)

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
