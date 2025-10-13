---
title: "Per-Router Observability"
description: "You can disable access logs, metrics, and tracing for a specific entrypoint attached to a HTTP Router. Read the technical documentation."
---

Traefik's observability features include logs, access logs, metrics, and tracing. You can configure these options globally or at more specific levels, such as per router or per entry point.

By default, the router observability configuration is inherited from the attached EntryPoints and can be configured with the observability [options](../../../install-configuration/entrypoints.md#configuration-options).
However, a router defining its own observability configuration will opt-out from these defaults.

!!! info
    To enable router-level observability, you must first enable
    [access-logs](../../../install-configuration/observability/logs-and-accesslogs.md#accesslogs),
    [tracing](../../../install-configuration/observability/tracing.md),
    and [metrics](../../../install-configuration/observability/metrics.md).

    When metrics layers are not enabled with the `addEntryPointsLabels`, `addRoutersLabels` and/or `addServicesLabels` options,
    enabling metrics for a router will not enable them.

!!! warning "AddInternals option"

    By default, and for any type of signal (access-logs, metrics and tracing),
    Traefik disables observability for internal resources.
    The observability options described below cannot interfere with the `AddInternals` ones,
    and will be ignored.

    For instance, if a router exposes the `api@internal` service and `metrics.AddInternals` is false,
    it will never produces metrics, even if the router observability configuration enables metrics.

## Configuration Example

```yaml tab="Structured (YAML)"
http:
  routers:
    my-router:
      rule: "Path(`/foo`)"
      service: service-foo
      observability:
        metrics: false
        accessLogs: false
        tracing: false
        traceVerbosity: detailed
```

```yaml tab="Structured (TOML)"
[http.routers.my-router]
  rule = "Path(`/foo`)"
  service = "service-foo"

  [http.routers.my-router.observability]
    metrics = false
    accessLogs = false
    tracing = false
    traceVerbosity = "detailed"
```

```yaml tab="Labels"
labels:
  - "traefik.http.routers.my-router.rule=Path(`/foo`)"
  - "traefik.http.routers.my-router.service=service-foo"
  - "traefik.http.routers.my-router.observability.metrics=false"
  - "traefik.http.routers.my-router.observability.accessLogs=false"
  - "traefik.http.routers.my-router.observability.tracing=false"
  - "traefik.http.routers.my-router.observability.traceVerbosity=detailed"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.routers.my-router.rule=Path(`/foo`)",
    "traefik.http.routers.my-router.service=service-foo",
    "traefik.http.routers.my-router.observability.metrics=false",
    "traefik.http.routers.my-router.observability.accessLogs=false",
    "traefik.http.routers.my-router.observability.tracing=false",
    "traefik.http.routers.my-router.observability.traceVerbosity=detailed"
  ]
}
```

## Configuration Options

| Field            | Description                                                                                                                                                                                | Default   | Required |
|:-----------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------|:---------|
| <a id="opt-accessLogs" href="#opt-accessLogs" title="#opt-accessLogs">`accessLogs`</a> | The `accessLogs` option controls whether the router will produce access-logs.                                                                                                              | `true`    | No       |
| <a id="opt-metrics" href="#opt-metrics" title="#opt-metrics">`metrics`</a> | The `metrics` option controls whether the router will produce metrics.                                                                                                                     | `true`    | No       |
| <a id="opt-tracing" href="#opt-tracing" title="#opt-tracing">`tracing`</a> | The `tracing` option controls whether the router will produce traces.                                                                                                                      | `true`    | No       |
| <a id="opt-traceVerbosity" href="#opt-traceVerbosity" title="#opt-traceVerbosity">`traceVerbosity`</a> | The `traceVerbosity` option controls the tracing verbosity level for the router. Possible values: `minimal` (default), `detailed`. If not set, the value is inherited from the entryPoint. | `minimal` | No       |

#### traceVerbosity

`observability.traceVerbosity` defines the tracing verbosity level for the router.

Possible values are:

- `minimal`: produces a single server span and one client span for each request processed by a router.
- `detailed`: enables the creation of additional spans for each middleware executed for each request processed by a router.
