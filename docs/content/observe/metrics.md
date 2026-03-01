---
title: "Metrics"
description: "Metrics in Traefik Proxy offer a comprehensive view of your infrastructure's health. They allow you to monitor critical indicators like incoming traffic volume. Metrics graphs and visualizations are helpful during incident triage in understanding the causes and implementing proactive measures."
---

# Metrics

Metrics in Traefik Proxy offer a comprehensive view of your infrastructure's health. They allow you to monitor critical indicators like incoming traffic volume. Metrics graphs and visualizations are helpful during incident triage in understanding the causes and implementing proactive measures.

## Available Metrics Providers

Traefik Proxy supports the following metrics providers:

- OpenTelemetry
- Prometheus
- Datadog
- InfluxDB 2.X
- StatsD

## Configuration

To enable metrics in Traefik Proxy, you need to configure the metrics provider in your static configuration file or helm values if you are using the [Helm chart](https://github.com/traefik/traefik-helm-chart). The following example shows how to configure the OpenTelemetry provider to send metrics to a collector.

```yaml tab="Structured (YAML)"
metrics:
  otlp:
    http:
      endpoint: http://myotlpcollector:4318/v1/metrics
```

```toml tab="Structured (TOML)"
[metrics.otlp.http]
  endpoint = "http://myotlpcollector:4318/v1/metrics"
```

```yaml tab="Helm Chart Values"
# values.yaml
metrics:
  # Disable Prometheus (enabled by default)
  prometheus: null
  # Enable providing OTel metrics
  otlp:
    enabled: true
    http:
      enabled: true
      endpoint: http://myotlpcollector:4318/v1/metrics
```

## Per-Router Metrics

You can enable or disable metrics collection for a specific router. This can be useful for excluding certain routes from your metrics data.

Here's an example of disabling metrics on a specific router:

```yaml tab="Structured (YAML)"
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: my-service
      observability:
        metrics: false
```

```toml tab="Structured (TOML)"
[http.routers.my-router.observability]
  metrics = false
```

```yaml tab="Kubernetes"
# ingressroute.yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-router
spec:
  routes:
    - kind: Rule
      match: Host(`example.com`)
      services:
        - name: my-service
          port: 80
      observability:
        metrics: false
```

```bash tab="Labels"
labels:
  - "traefik.http.routers.my-router.observability.metrics=false"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.routers.my-router.observability.metrics=false"
  ]
}
```

When the `observability` options are not defined on a router, it inherits the behavior from the [entrypoint's observability configuration](./overview.md), or the global one.

!!! info
    For detailed configuration options, refer to the [reference documentation](../reference/install-configuration/observability/metrics.md).
