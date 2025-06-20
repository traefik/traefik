---
title: "Tracing"
description: "Tracing in Traefik Proxy allows you to track the flow of operations within your system. Using traces and spans, you can identify performance bottlenecks and pinpoint applications causing slowdowns to optimize response times effectively."
---

# Tracing

Tracing in Traefik Proxy allows you to track the flow of operations within your system. Using traces and spans, you can identify performance bottlenecks and pinpoint applications causing slowdowns to optimize response times effectively.

Traefik Proxy uses [OpenTelemetry](https://opentelemetry.io/) to export traces. OpenTelemetry is an open-source observability framework. You can send traces to an OpenTelemetry collector, which can then export them to a variety of backends like Jaeger, Zipkin, or Datadog.

## Configuration

To enable tracing in Traefik Proxy, you need to configure it in your static configuration file or Helm values if you are using the [Helm chart](https://github.com/traefik/traefik-helm-chart). The following example shows how to configure the OpenTelemetry provider to send traces to a collector via HTTP.

```yaml tab="Structured (YAML)"
tracing:
  otlp:
    http:
      endpoint: http://myotlpcollector:4318/v1/traces
```

```toml tab="Structured (TOML)"
[tracing.otlp.http]
  endpoint = "http://myotlpcollector:4318/v1/traces"
```

```yaml tab="Helm Chart Values"
# values.yaml
tracing:
  otlp:
    enabled: true
    http:
      enabled: true
      endpoint: http://myotlpcollector:4318/v1/traces
```

!!! info
    For detailed configuration options, refer to the [tracing reference documentation](../reference/install-configuration/observability/tracing.md).

## Per-Router Tracing

You can enable or disable tracing for a specific router. This is useful for turning off tracing for specific routes while keeping it on globally.

Here's an example of disabling tracing on a specific router:

```yaml tab="Structured (YAML)"
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: my-service
      observability:
        tracing: false
```

```toml tab="Structured (TOML)"
[http.routers.my-router.observability]
  tracing = false
```

```yaml tab="Kubernetes"
# ingressoute.yaml
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
        tracing: false
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.my-router.observability.tracing=false"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.routers.my-router.observability.tracing=false"
  ]
}
```

When the `observability` options are not defined on a router, it inherits the behavior from the [entrypoint's observability configuration](./overview.md), or the global one.
