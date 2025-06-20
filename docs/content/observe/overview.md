---
title: "Observability Overview"
description: "Traefik Proxy provides comprehensive monitoring and observability capabilities to maintain reliability and efficiency."
---

# Observability Overview

Traefik Proxy provides comprehensive monitoring and observability capabilities to maintain reliability and efficiency:

- [Logs and Access Logs](./logs-and-access-logs.md) provide real-time insight into the health of your system. They enable swift error detection and intervention through alerts. By centralizing logs, you can streamline the debugging process during incident resolution.

- [Metrics](./metrics.md) offer a comprehensive view of your infrastructure's health. They allow you to monitor critical indicators like incoming traffic volume. Metrics graphs and visualizations are helpful during incident triage in understanding the causes and implementing proactive measures.

- [Tracing](./tracing.md) enables tracking the flow of operations within your system. Using traces and spans, you can identify performance bottlenecks and pinpoint applications causing slowdowns to optimize response times effectively.

## Configuration Example

You can enable access logs, metrics, and tracing globally:

```yaml tab="Structured (YAML)"
accessLog: {}

metrics:
  otlp: {}

tracing: {}
```

```toml tab="Structured (TOML)"
[accessLog]

[metrics.otlp]

[tracing.otlp]
```

```yaml tab="Helm Chart Values"
# values.yaml
accessLog:
  enabled: true

metrics:
  otlp:
    enabled: true

tracing:
  otlp:
    enabled: true
```

You can disable access logs, metrics, and tracing for a specific [entrypoint](../reference/install-configuration/entrypoints.md):

```yaml tab="Structured (YAML)"
entryPoints:
  EntryPoint0:
    address: ':8000/udp'
    observability:
      accessLogs: false
      tracing: false
      metrics: false
```

```toml tab="Structured (TOML)"
[entryPoints.EntryPoint0.observability]
  accessLogs = false
  tracing = false
  metrics = false
```

```yaml tab="Helm Chart Values"
additionalArguments:
  - "--entrypoints.entrypoint0.observability.accesslogs=false"
  - "--entrypoints.entrypoint0.observability.tracing=false"
  - "--entrypoints.entrypoint0.observability.metrics=false"
```

!!! note
    A router with its own observability configuration will override the global default.

{!traefik-for-business-applications.md!}
