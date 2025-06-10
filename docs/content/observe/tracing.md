# Tracing

Tracing in Traefik Proxy allows you to track the flow of operations within your system. Using traces and spans, you can identify performance bottlenecks and pinpoint applications causing slowdowns to optimize response times effectively.

Traefik Proxy uses [OpenTelemetry](https://opentelemetry.io/) to export traces. OpenTelemetry is an open-source observability framework. You can send traces to an OpenTelemetry collector, which can then export them to a variety of backends like Jaeger, Zipkin, or Datadog.

## Configuration

To enable tracing in Traefik Proxy, you need to configure it in your configuration file or helm values if you are using the [Helm chart](https://github.com/traefik/traefik-helm-chart). The following example shows how to configure the OpenTelemetry provider to send traces to a collector via HTTP.

```yaml tab="Structured (YAML)"
tracing:
  otlp:
    http:
      endpoint: http://myotlpcollector:4318/v1/traces
```

```yaml tab="Helm Values"
# values.yaml
tracing:
  otlp:
    enabled: true
    http:
      enabled: true
      endpoint: http://myotlpcollector:4318/v1/traces
```

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

When the `observability` options are not defined on a router, it inherits the behavior from the entrypoint's observability configuration, or the global one.
