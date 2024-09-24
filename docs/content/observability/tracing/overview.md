---
title: "Traefik Tracing Overview"
description: "The Traefik Proxy tracing system allows developers to visualize call flows in their infrastructure. Read the full documentation."
---

# Tracing

Visualize the Requests Flow
{: .subtitle }

The tracing system allows developers to visualize call flows in their infrastructure.

Traefik uses [OpenTelemetry](https://opentelemetry.io/ "Link to website of OTel"), an open standard designed for distributed tracing.

Please check our dedicated [OTel docs](./opentelemetry.md) to learn more.

## Configuration

To enable the tracing:

```yaml tab="File (YAML)"
tracing: {}
```

```toml tab="File (TOML)"
[tracing]
```

```bash tab="CLI"
--tracing=true
```

### Common Options

#### `addInternals`

_Optional, Default="false"_

Enables tracing for internal resources (e.g.: `ping@internal`).

```yaml tab="File (YAML)"
tracing:
  addInternals: true
```

```toml tab="File (TOML)"
[tracing]
  addInternals = true
```

```bash tab="CLI"
--tracing.addinternals
```

#### `serviceName`

_Required, Default="traefik"_

Service name used in selected backend.

```yaml tab="File (YAML)"
tracing:
  serviceName: traefik
```

```toml tab="File (TOML)"
[tracing]
  serviceName = "traefik"
```

```bash tab="CLI"
--tracing.serviceName=traefik
```

#### `sampleRate`

_Optional, Default=1.0_

The proportion of requests to trace, specified between 0.0 and 1.0.

```yaml tab="File (YAML)"
tracing:
  sampleRate: 0.2
```

```toml tab="File (TOML)"
[tracing]
  sampleRate = 0.2
```

```bash tab="CLI"
--tracing.sampleRate=0.2
```

#### `globalAttributes`

_Optional, Default=empty_

Applies a list of shared key:value attributes on all spans.

```yaml tab="File (YAML)"
tracing:
  globalAttributes:
    attr1: foo
    attr2: bar
```

```toml tab="File (TOML)"
[tracing]
  [tracing.globalAttributes]
    attr1 = "foo"
    attr2 = "bar"
```

```bash tab="CLI"
--tracing.globalAttributes.attr1=foo
--tracing.globalAttributes.attr2=bar
```

#### `capturedRequestHeaders`

_Optional, Default=empty_

Defines the list of request headers to add as attributes.
It applies to client and server kind spans.

```yaml tab="File (YAML)"
tracing:
  capturedRequestHeaders:
    - X-CustomHeader
```

```toml tab="File (TOML)"
[tracing]
  capturedRequestHeaders = ["X-CustomHeader"]
```

```bash tab="CLI"
--tracing.capturedRequestHeaders[0]=X-CustomHeader
```

#### `capturedResponseHeaders`

_Optional, Default=empty_

Defines the list of response headers to add as attributes.
It applies to client and server kind spans.

```yaml tab="File (YAML)"
tracing:
  capturedResponseHeaders:
    - X-CustomHeader
```

```toml tab="File (TOML)"
[tracing]
  capturedResponseHeaders = ["X-CustomHeader"]
```

```bash tab="CLI"
--tracing.capturedResponseHeaders[0]=X-CustomHeader
```

#### `safeQueryParams`

_Optional, Default={}_

By default, all query parameters are redacted.
Defines the list of query parameters to not redact.

```yaml tab="File (YAML)"
tracing:
  safeQueryParams:
    - bar
    - buz
```

```toml tab="File (TOML)"
[tracing]
  safeQueryParams = ["bar", "buz"]
```

```bash tab="CLI"
--tracing.safeQueryParams=bar,buz
```
