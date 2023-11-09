---
title: "Traefik Zipkin Documentation"
description: "Traefik supports several tracing backends, including Zipkin. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# Zipkin

To enable the Zipkin tracer:

```yaml tab="File (YAML)"
tracing:
  zipkin: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.zipkin]
```

```bash tab="CLI"
--tracing.zipkin=true
```

#### `httpEndpoint`

_Required, Default="http://localhost:9411/api/v2/spans"_

HTTP endpoint used to send data.

```yaml tab="File (YAML)"
tracing:
  zipkin:
    httpEndpoint: http://localhost:9411/api/v2/spans
```

```toml tab="File (TOML)"
[tracing]
  [tracing.zipkin]
    httpEndpoint = "http://localhost:9411/api/v2/spans"
```

```bash tab="CLI"
--tracing.zipkin.httpEndpoint=http://localhost:9411/api/v2/spans
```

#### `globalTags`

_Optional, Default=empty_

Applies a list of shared key:value tags on all spans.

```yaml tab="File (YAML)"
tracing:
  zipkin:
    globalTags:
      tag1: foo
      tag2: bar
```

```toml tab="File (TOML)"
[tracing]
  [tracing.zipkin]
    [tracing.zipkin.globalTags]
      tag1 = "foo"
      tag2 = "bar"
```

```bash tab="CLI"
--tracing.zipkin.globalTags.tag1=foo
--tracing.zipkin.globalTags.tag2=bar
```

#### `sampleRate`

_Optional, Default=1.0_

The proportion of requests to trace, specified between 0.0 and 1.0.

```yaml tab="File (YAML)"
tracing:
  zipkin:
    sampleRate: 0.2
```

```toml tab="File (TOML)"
[tracing]
  [tracing.zipkin]
    sampleRate = 0.2
```

```bash tab="CLI"
--tracing.zipkin.sampleRate=0.2
```
