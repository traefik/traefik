---
title: "Traefik Prometheus Documentation"
description: "Traefik supports several metrics backends, including Prometheus. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# Prometheus

To enable the Prometheus:

```yaml tab="File (YAML)"
metrics:
  prometheus: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
```

```bash tab="CLI"
--metrics.prometheus=true
```

#### `buckets`

_Optional, Default="0.100000, 0.300000, 1.200000, 5.000000"_

Buckets for latency metrics.

```yaml tab="File (YAML)"
metrics:
  prometheus:
    buckets:
      - 0.1
      - 0.3
      - 1.2
      - 5.0
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    buckets = [0.1,0.3,1.2,5.0]
```

```bash tab="CLI"
--metrics.prometheus.buckets=0.1,0.3,1.2,5.0
```

#### `addEntryPointsLabels`

_Optional, Default=true_

Enable metrics on entry points.

```yaml tab="File (YAML)"
metrics:
  prometheus:
    addEntryPointsLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    addEntryPointsLabels = true
```

```bash tab="CLI"
--metrics.prometheus.addEntryPointsLabels=true
```

#### `addRoutersLabels`

_Optional, Default=false_

Enable metrics on routers.

```yaml tab="File (YAML)"
metrics:
  prometheus:
    addRoutersLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    addRoutersLabels = true
```

```bash tab="CLI"
--metrics.prometheus.addrouterslabels=true
```

#### `addServicesLabels`

_Optional, Default=true_

Enable metrics on services.

```yaml tab="File (YAML)"
metrics:
  prometheus:
    addServicesLabels: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    addServicesLabels = true
```

```bash tab="CLI"
--metrics.prometheus.addServicesLabels=true
```

#### `entryPoint`

_Optional, Default=traefik_

Entry point used to expose metrics.

```yaml tab="File (YAML)"
entryPoints:
  metrics:
    address: :8082

metrics:
  prometheus:
    entryPoint: metrics
```

```toml tab="File (TOML)"
[entryPoints]
  [entryPoints.metrics]
    address = ":8082"

[metrics]
  [metrics.prometheus]
    entryPoint = "metrics"
```

```bash tab="CLI"
--entryPoints.metrics.address=:8082
--metrics.prometheus.entryPoint=metrics
```

#### `manualRouting`

_Optional, Default=false_

If `manualRouting` is `true`, it disables the default internal router in order to allow one to create a custom router for the `prometheus@internal` service.

```yaml tab="File (YAML)"
metrics:
  prometheus:
    manualRouting: true
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    manualRouting = true
```

```bash tab="CLI"
--metrics.prometheus.manualrouting=true
```

#### `headerLabels`

_Optional_

Defines the extra labels for the `requests_total` metrics, and for each of them, the request header containing the value for this label.
Please note that if the header is not present in the request it will be added nonetheless with an empty value.
In addition, the label should be a valid label name for Prometheus metrics, 
otherwise, the Prometheus metrics provider will fail to serve any Traefik-related metric.

```yaml tab="File (YAML)"
metrics:
  prometheus:
    headerLabels:
      label: headerKey
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    [metrics.prometheus.headerLabels]
      label = "headerKey"
```

```bash tab="CLI"
--metrics.prometheus.headerlabels.label=headerKey
```

##### Example

Here is an example of the entryPoint `requests_total` metric with an additional "useragent" label.

When configuring the label in Static Configuration:

```yaml tab="File (YAML)"
metrics:
  prometheus:
    headerLabels:
      useragent: User-Agent
```

```toml tab="File (TOML)"
[metrics]
  [metrics.prometheus]
    [metrics.prometheus.headerLabels]
      useragent = "User-Agent"
```

```bash tab="CLI"
--metrics.prometheus.headerlabels.useragent=User-Agent
```

And performing a request with a custom User-Agent:

```bash
curl -H "User-Agent: foobar" http://localhost
```

The following metric is produced :

```bash
traefik_entrypoint_requests_total{code="200",entrypoint="web",method="GET",protocol="http",useragent="foobar"} 1
```

!!! info "`Host` header value"

    The `Host` header is never present in the Header map of a request, as per go documentation says:
    // For incoming requests, the Host header is promoted to the
    // Request.Host field and removed from the Header map.

    As a workaround, to obtain the Host of a request as a label, one should use instead the `X-Forwarded-Host` header.
