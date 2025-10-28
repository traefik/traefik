---
title: "Traefik Metrics Overview"
description: "Traefik Proxy supports these metrics backend systems: OpenTelemetry, Datadog, InfluxDB 2.X, Prometheus, and StatsD. Read the full documentation to get started."
---

# Metrics

Traefik provides metrics in the [OpenTelemetry](#open-telemetry) format as well as the following vendor specific backends:

- [Datadog](#datadog)
- [InfluxDB2](#influxdb-v2)
- [Prometheus](#prometheus)
- [StatsD](#statsd)

Traefik Proxy has an official Grafana dashboard for both [on-premises](https://grafana.com/grafana/dashboards/17346)
and [Kubernetes](https://grafana.com/grafana/dashboards/17347) deployments.

---

## Open Telemetry

!!! info "Default protocol"

    The OpenTelemetry exporter will export metrics to the collector using HTTP by default to https://localhost:4318/v1/metrics.

### Configuration Example

To enable the OpenTelemetry metrics:

```yaml tab="File (YAML)"
metrics:
  otlp: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.otlp]
```

```bash tab="CLI"
--metrics.otlp=true
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
```

!!! tip "Helm Chart Configuration"

    Traefik can be configured to provide metrics in the OpenTelemetry format using the Helm Chart values.
    To know more about the Helm Chart options, refer to the [Helm Chart](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/VALUES.md) (Find options `metrics.otlp`).

### Configuration Options

| Field                                      | Description                                                                                                                                                      | Default                                            | Required |
|:-------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------|:---------|
| <a id="opt-metrics-addInternals" href="#opt-metrics-addInternals" title="#opt-metrics-addInternals">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internal`).                                                                                                  | false                                              | No       |
| <a id="opt-metrics-otlp-serviceName" href="#opt-metrics-otlp-serviceName" title="#opt-metrics-otlp-serviceName">`metrics.otlp.serviceName`</a> | Defines the service name resource attribute.                                                                                                                     | "traefik"                                          | No       |
| <a id="opt-metrics-otlp-resourceAttributes" href="#opt-metrics-otlp-resourceAttributes" title="#opt-metrics-otlp-resourceAttributes">`metrics.otlp.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector. See [resourceAttributes](#resourceattributes) for details.                                                                                                   | []                                                 | No       |
| <a id="opt-metrics-otlp-addEntryPointsLabels" href="#opt-metrics-otlp-addEntryPointsLabels" title="#opt-metrics-otlp-addEntryPointsLabels">`metrics.otlp.addEntryPointsLabels`</a> | Enable metrics on entry points.                                                                                                                                  | true                                               | No       |
| <a id="opt-metrics-otlp-addRoutersLabels" href="#opt-metrics-otlp-addRoutersLabels" title="#opt-metrics-otlp-addRoutersLabels">`metrics.otlp.addRoutersLabels`</a> | Enable metrics on routers.                                                                                                                                       | false                                              | No       |
| <a id="opt-metrics-otlp-addServicesLabels" href="#opt-metrics-otlp-addServicesLabels" title="#opt-metrics-otlp-addServicesLabels">`metrics.otlp.addServicesLabels`</a> | Enable metrics on services.                                                                                                                                      | true                                               | No       |
| <a id="opt-metrics-otlp-explicitBoundaries" href="#opt-metrics-otlp-explicitBoundaries" title="#opt-metrics-otlp-explicitBoundaries">`metrics.otlp.explicitBoundaries`</a> | Explicit boundaries for Histogram data points.                                                                                                                   | ".005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10" | No       |
| <a id="opt-metrics-otlp-pushInterval" href="#opt-metrics-otlp-pushInterval" title="#opt-metrics-otlp-pushInterval">`metrics.otlp.pushInterval`</a> | Interval at which metrics are sent to the OpenTelemetry Collector.                                                                                               | 10s                                                | No       |
| <a id="opt-metrics-otlp-http" href="#opt-metrics-otlp-http" title="#opt-metrics-otlp-http">`metrics.otlp.http`</a> | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| <a id="opt-metrics-otlp-http-endpoint" href="#opt-metrics-otlp-http-endpoint" title="#opt-metrics-otlp-http-endpoint">`metrics.otlp.http.endpoint`</a> | URL of the OpenTelemetry Collector to send metrics to.<br /> Format="`<scheme>://<host>:<port><path>`"                                                           | "https://localhost:4318/v1/metrics"                 | Yes      |
| <a id="opt-metrics-otlp-http-headers" href="#opt-metrics-otlp-http-headers" title="#opt-metrics-otlp-http-headers">`metrics.otlp.http.headers`</a> | Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.                                                                             | -                                                  | No       |
| <a id="opt-metrics-otlp-http-tls-ca" href="#opt-metrics-otlp-http-tls-ca" title="#opt-metrics-otlp-http-tls-ca">`metrics.otlp.http.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector,<br />it defaults to the system bundle.                          | ""                                                 | No       |
| <a id="opt-metrics-otlp-http-tls-cert" href="#opt-metrics-otlp-http-tls-cert" title="#opt-metrics-otlp-http-tls-cert">`metrics.otlp.http.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector.<br />When using this option, setting the `key` option is required. | ""                                                 | No       |
| <a id="opt-metrics-otlp-http-tls-key" href="#opt-metrics-otlp-http-tls-key" title="#opt-metrics-otlp-http-tls-key">`metrics.otlp.http.tls.key`</a> | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| <a id="opt-metrics-otlp-http-tls-insecureskipverify" href="#opt-metrics-otlp-http-tls-insecureskipverify" title="#opt-metrics-otlp-http-tls-insecureskipverify">`metrics.otlp.http.tls.insecureskipverify`</a> | Allow the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.                   | false                                              | Yes      |
| <a id="opt-metrics-otlp-grpc" href="#opt-metrics-otlp-grpc" title="#opt-metrics-otlp-grpc">`metrics.otlp.grpc`</a> | This instructs the exporter to send metrics to the OpenTelemetry Collector using gRPC.                                                                           | null/false                                         | No       |
| <a id="opt-metrics-otlp-grpc-endpoint" href="#opt-metrics-otlp-grpc-endpoint" title="#opt-metrics-otlp-grpc-endpoint">`metrics.otlp.grpc.endpoint`</a> | Address of the OpenTelemetry Collector to send metrics to.<br /> Format="`<host>:<port>`"                                                                        | "localhost:4317"                                   | Yes      |
| <a id="opt-metrics-otlp-grpc-headers" href="#opt-metrics-otlp-grpc-headers" title="#opt-metrics-otlp-grpc-headers">`metrics.otlp.grpc.headers`</a> | Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.                                                                             | -                                                  | No       |
| <a id="opt-metrics-otlp-http-grpc-insecure" href="#opt-metrics-otlp-http-grpc-insecure" title="#opt-metrics-otlp-http-grpc-insecure">`metrics.otlp.http.grpc.insecure`</a> | Allows exporter to send metrics to the OpenTelemetry Collector without using a secured protocol.                                                                 | false                                              | Yes      |
| <a id="opt-metrics-otlp-grpc-tls-ca" href="#opt-metrics-otlp-grpc-tls-ca" title="#opt-metrics-otlp-grpc-tls-ca">`metrics.otlp.grpc.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector,<br />it defaults to the system bundle.                          | -                                                  | No       |
| <a id="opt-metrics-otlp-grpc-tls-cert" href="#opt-metrics-otlp-grpc-tls-cert" title="#opt-metrics-otlp-grpc-tls-cert">`metrics.otlp.grpc.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector.<br />When using this option, setting the `key` option is required. | -                                                  | No       |
| <a id="opt-metrics-otlp-grpc-tls-key" href="#opt-metrics-otlp-grpc-tls-key" title="#opt-metrics-otlp-grpc-tls-key">`metrics.otlp.grpc.tls.key`</a> | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| <a id="opt-metrics-otlp-grpc-tls-insecureskipverify" href="#opt-metrics-otlp-grpc-tls-insecureskipverify" title="#opt-metrics-otlp-grpc-tls-insecureskipverify">`metrics.otlp.grpc.tls.insecureskipverify`</a> | Allow the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.                   | false                                              | Yes      |

### resourceAttributes

The `resourceAttributes` option allows setting the resource attributes sent along the traces.
Traefik also supports the `OTEL_RESOURCE_ATTRIBUTES` env variable to set up the resource attributes.

!!! info "Kubernetes Resource Attributes Detection"

    Additionally, Traefik automatically discovers the following [Kubernetes resource attributes](https://opentelemetry.io/docs/specs/semconv/non-normative/k8s-attributes/) when running in a Kubernetes cluster:
    
    - `k8s.namespace.name`
    - `k8s.pod.uid`
    - `k8s.pod.name`
    
    Note that this automatic detection can fail, like if the Traefik pod is running in host network mode.
    In this case, you should provide the attributes with the option or the env variable.

## Vendors

### Datadog

#### Configuration Example 

To enable the Datadog:

```yaml tab="File (YAML)"
metrics:
  datadog: {}
```

```toml tab="File (TOML)"
[metrics]
  [metrics.datadog]
```

```bash tab="CLI"
--metrics.datadog=true
```

#### Configuration Options

| Field | Description      | Default              | Required |
|:------|:-------------------------------|:---------------------|:---------|
| <a id="opt-metrics-addInternals-2" href="#opt-metrics-addInternals-2" title="#opt-metrics-addInternals-2">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internal`). | false      | No      |
| <a id="opt-datadog-address" href="#opt-datadog-address" title="#opt-datadog-address">`datadog.address`</a> | Defines the address for the exporter to send metrics to datadog-agent. More information [here](#address)|  `127.0.0.1:8125`     | Yes   |
| <a id="opt-datadog-addEntryPointsLabels" href="#opt-datadog-addEntryPointsLabels" title="#opt-datadog-addEntryPointsLabels">`datadog.addEntryPointsLabels`</a> | Enable metrics on entry points. |  true   | No   |
| <a id="opt-datadog-addRoutersLabels" href="#opt-datadog-addRoutersLabels" title="#opt-datadog-addRoutersLabels">`datadog.addRoutersLabels`</a> | Enable metrics on routers. |  false   | No   |
| <a id="opt-datadog-addServicesLabels" href="#opt-datadog-addServicesLabels" title="#opt-datadog-addServicesLabels">`datadog.addServicesLabels`</a> | Enable metrics on services. |  true   | No   |
| <a id="opt-datadog-pushInterval" href="#opt-datadog-pushInterval" title="#opt-datadog-pushInterval">`datadog.pushInterval`</a> | Defines the interval used by the exporter to push metrics to datadog-agent. |  10s   | No   |
| <a id="opt-datadog-prefix" href="#opt-datadog-prefix" title="#opt-datadog-prefix">`datadog.prefix`</a> | Defines the prefix to use for metrics collection. |  "traefik"   | No   |

##### `address`

Address instructs exporter to send metrics to datadog-agent at this address.

This address can be a Unix Domain Socket (UDS) in the following format: `unix:///path/to/datadog.socket`.
When the prefix is set to `unix`, the socket type will be automatically determined. 
To explicitly define the socket type and avoid automatic detection, you can use the prefixes `unixgram` for `SOCK_DGRAM` (datagram sockets) and `unixstream` for `SOCK_STREAM` (stream sockets), respectively.

```yaml tab="File (YAML)"
metrics:
  datadog:
    address: 127.0.0.1:8125
```

```toml tab="File (TOML)"
[metrics]
  [metrics.datadog]
    address = "127.0.0.1:8125"
```

```bash tab="CLI"
--metrics.datadog.address=127.0.0.1:8125
```

### InfluxDB v2

#### Configuration Example 

To enable the InfluxDB2:

```yaml tab="File (YAML)"
metrics:
  influxDB2:
    address: http://localhost:8086
```

```toml tab="File (TOML)"
[metrics]
  [metrics.influxDB2]
    address: http://localhost:8086
```

```bash tab="CLI"
--metrics.influxdb2=true
```

#### Configuration Options

| Field      | Description      | Default | Required |
|:-----------|-------------------------|:--------|:---------|
| <a id="opt-metrics-addInternal" href="#opt-metrics-addInternal" title="#opt-metrics-addInternal">`metrics.addInternal`</a> | Enables metrics for internal resources (e.g.: `ping@internal`). | false      | No      |
| <a id="opt-metrics-influxDB2-addEntryPointsLabels" href="#opt-metrics-influxDB2-addEntryPointsLabels" title="#opt-metrics-influxDB2-addEntryPointsLabels">`metrics.influxDB2.addEntryPointsLabels`</a> | Enable metrics on entry points. | true      | No      |
| <a id="opt-metrics-influxDB2-addRoutersLabels" href="#opt-metrics-influxDB2-addRoutersLabels" title="#opt-metrics-influxDB2-addRoutersLabels">`metrics.influxDB2.addRoutersLabels`</a> | Enable metrics on routers. | false      | No      |
| <a id="opt-metrics-influxDB2-addServicesLabels" href="#opt-metrics-influxDB2-addServicesLabels" title="#opt-metrics-influxDB2-addServicesLabels">`metrics.influxDB2.addServicesLabels`</a> | Enable metrics on services.| true      | No      |
| <a id="opt-metrics-influxDB2-additionalLabels" href="#opt-metrics-influxDB2-additionalLabels" title="#opt-metrics-influxDB2-additionalLabels">`metrics.influxDB2.additionalLabels`</a> | Additional labels (InfluxDB tags) on all metrics. | - | No      |
| <a id="opt-metrics-influxDB2-pushInterval" href="#opt-metrics-influxDB2-pushInterval" title="#opt-metrics-influxDB2-pushInterval">`metrics.influxDB2.pushInterval`</a> | The interval used by the exporter to push metrics to InfluxDB server. | 10s      | No      |
| <a id="opt-metrics-influxDB2-address" href="#opt-metrics-influxDB2-address" title="#opt-metrics-influxDB2-address">`metrics.influxDB2.address`</a> | Address of the InfluxDB v2 instance. | "http://localhost:8086"     | Yes      |
| <a id="opt-metrics-influxDB2-token" href="#opt-metrics-influxDB2-token" title="#opt-metrics-influxDB2-token">`metrics.influxDB2.token`</a> | Token with which to connect to InfluxDB v2. | - | Yes      |
| <a id="opt-metrics-influxDB2-org" href="#opt-metrics-influxDB2-org" title="#opt-metrics-influxDB2-org">`metrics.influxDB2.org`</a> | Organisation where metrics will be stored. | -  | Yes      |
| <a id="opt-metrics-influxDB2-bucket" href="#opt-metrics-influxDB2-bucket" title="#opt-metrics-influxDB2-bucket">`metrics.influxDB2.bucket`</a> | Bucket where metrics will be stored. | -  | Yes      |

### Prometheus

#### Configuration Example

To enable the Prometheus:

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
    [metrics.prometheus.buckets]
      - 0.1
      - 0.3
      - 1.2
      - 5.0
```

```bash tab="CLI"
--metrics.prometheus=true
```

#### Configuration Options

| Field      | Description         | Default | Required |
|:-----------|---------------------|:--------|:---------|
| <a id="opt-metrics-addInternals-3" href="#opt-metrics-addInternals-3" title="#opt-metrics-addInternals-3">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internals`). | false      | No      |
| <a id="opt-metrics-prometheus-addEntryPointsLabels" href="#opt-metrics-prometheus-addEntryPointsLabels" title="#opt-metrics-prometheus-addEntryPointsLabels">`metrics.prometheus.addEntryPointsLabels`</a> | Enable metrics on entry points. | true      | No      |
| <a id="opt-metrics-prometheus-addRoutersLabels" href="#opt-metrics-prometheus-addRoutersLabels" title="#opt-metrics-prometheus-addRoutersLabels">`metrics.prometheus.addRoutersLabels`</a> | Enable metrics on routers. | false      | No      |
| <a id="opt-metrics-prometheus-addServicesLabels" href="#opt-metrics-prometheus-addServicesLabels" title="#opt-metrics-prometheus-addServicesLabels">`metrics.prometheus.addServicesLabels`</a> | Enable metrics on services.| true      | No      |
| <a id="opt-metrics-prometheus-buckets" href="#opt-metrics-prometheus-buckets" title="#opt-metrics-prometheus-buckets">`metrics.prometheus.buckets`</a> | Buckets for latency metrics. |"0.100000, 0.300000, 1.200000, 5.000000"  | No      |
| <a id="opt-metrics-prometheus-manualRouting" href="#opt-metrics-prometheus-manualRouting" title="#opt-metrics-prometheus-manualRouting">`metrics.prometheus.manualRouting`</a> | Set to _true_, it disables the default internal router in order to allow creating a custom router for the `prometheus@internal` service. | false    | No      |
| <a id="opt-metrics-prometheus-entryPoint" href="#opt-metrics-prometheus-entryPoint" title="#opt-metrics-prometheus-entryPoint">`metrics.prometheus.entryPoint`</a> | Traefik Entrypoint name used to expose metrics. | "traefik"     | No      |
| <a id="opt-metrics-prometheus-headerLabels" href="#opt-metrics-prometheus-headerLabels" title="#opt-metrics-prometheus-headerLabels">`metrics.prometheus.headerLabels`</a> | Defines extra labels extracted from request headers for the `requests_total` metrics.<br />More information [here](#headerlabels). |       | Yes      |

##### headerLabels

Defines the extra labels for the `requests_total` metrics, and for each of them, the request header containing the value for this label.
If the header is not present in the request it will be added nonetheless with an empty value.
The label must be a valid label name for Prometheus metrics, otherwise, the Prometheus metrics provider will fail to serve any Traefik-related metric.

!!! note "How to provide the `Host` header value"
      The `Host` header is never present in the Header map of a request, as per go documentation says:

      ```Golang
      // For incoming requests, the Host header is promoted to the
      // Request.Host field and removed from the Header map.
      ```

      As a workaround, to obtain the Host of a request as a label, use instead the `X-Forwarded-Host` header.

###### Configuration Example

Here is an example of the entryPoint `requests_total` metric with an additional "useragent" label.

When configuring the label in Static Configuration:

```yaml tab="Configuration"
# static_configuration.yaml
metrics:
  prometheus:
    headerLabels:
      useragent: User-Agent
```

```bash tab="Request"
curl -H "User-Agent: foobar" http://localhost
```

```bash tab="Metric"
traefik_entrypoint_requests_total\{code="200",entrypoint="web",method="GET",protocol="http",useragent="foobar"\} 1
```

### StatsD

#### Configuration Example

To enable the Statsd:

```yaml tab="File (YAML)"
metrics:
  statsD:
    address: localhost:8125
```

```toml tab="File (TOML)"
[metrics]
  [metrics.statsD]
    address: localhost:8125
```

```bash tab="CLI"
--metrics.statsd=true
```

#### Configuration Options

| Field      | Description       | Default | Required |
|:-----------|:-------------------------|:--------|:---------|
| <a id="opt-metrics-addInternals-4" href="#opt-metrics-addInternals-4" title="#opt-metrics-addInternals-4">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internals`). | false      | No      |
| <a id="opt-metrics-statsD-addEntryPointsLabels" href="#opt-metrics-statsD-addEntryPointsLabels" title="#opt-metrics-statsD-addEntryPointsLabels">`metrics.statsD.addEntryPointsLabels`</a> | Enable metrics on entry points. | true      | No      |
| <a id="opt-metrics-statsD-addRoutersLabels" href="#opt-metrics-statsD-addRoutersLabels" title="#opt-metrics-statsD-addRoutersLabels">`metrics.statsD.addRoutersLabels`</a> | Enable metrics on routers. | false      | No      |
| <a id="opt-metrics-statsD-addServicesLabels" href="#opt-metrics-statsD-addServicesLabels" title="#opt-metrics-statsD-addServicesLabels">`metrics.statsD.addServicesLabels`</a> | Enable metrics on services.| true      | No      |
| <a id="opt-metrics-statsD-pushInterval" href="#opt-metrics-statsD-pushInterval" title="#opt-metrics-statsD-pushInterval">`metrics.statsD.pushInterval`</a> | The interval used by the exporter to push metrics to DataDog server. | 10s      | No      |
| <a id="opt-metrics-statsD-address" href="#opt-metrics-statsD-address" title="#opt-metrics-statsD-address">`metrics.statsD.address`</a> | Address instructs exporter to send metrics to statsd at this address.  | "127.0.0.1:8125"     | Yes      |
| <a id="opt-metrics-statsD-prefix" href="#opt-metrics-statsD-prefix" title="#opt-metrics-statsD-prefix">`metrics.statsD.prefix`</a> | The prefix to use for metrics collection. | "traefik"      | No      |

## Metrics Provided

### Global Metrics

=== "OpenTelemetry"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="opt-traefik-config-reloads-total" href="#opt-traefik-config-reloads-total" title="#opt-traefik-config-reloads-total">`traefik_config_reloads_total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="opt-traefik-config-last-reload-success" href="#opt-traefik-config-last-reload-success" title="#opt-traefik-config-last-reload-success">`traefik_config_last_reload_success`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="opt-traefik-open-connections" href="#opt-traefik-open-connections" title="#opt-traefik-open-connections">`traefik_open_connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="opt-traefik-tls-certs-not-after" href="#opt-traefik-tls-certs-not-after" title="#opt-traefik-tls-certs-not-after">`traefik_tls_certs_not_after`</a> | Gauge |                          | The expiration date of certificates.                               |
    
=== "Prometheus"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="opt-traefik-config-reloads-total-2" href="#opt-traefik-config-reloads-total-2" title="#opt-traefik-config-reloads-total-2">`traefik_config_reloads_total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="opt-traefik-config-last-reload-success-2" href="#opt-traefik-config-last-reload-success-2" title="#opt-traefik-config-last-reload-success-2">`traefik_config_last_reload_success`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="opt-traefik-open-connections-2" href="#opt-traefik-open-connections-2" title="#opt-traefik-open-connections-2">`traefik_open_connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="opt-traefik-tls-certs-not-after-2" href="#opt-traefik-tls-certs-not-after-2" title="#opt-traefik-tls-certs-not-after-2">`traefik_tls_certs_not_after`</a> | Gauge |      | The expiration date of certificates. |

=== "Datadog"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="opt-config-reload-total" href="#opt-config-reload-total" title="#opt-config-reload-total">`config.reload.total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="opt-config-reload-lastSuccessTimestamp" href="#opt-config-reload-lastSuccessTimestamp" title="#opt-config-reload-lastSuccessTimestamp">`config.reload.lastSuccessTimestamp`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="opt-open-connections" href="#opt-open-connections" title="#opt-open-connections">`open.connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="opt-tls-certs-notAfterTimestamp" href="#opt-tls-certs-notAfterTimestamp" title="#opt-tls-certs-notAfterTimestamp">`tls.certs.notAfterTimestamp`</a> | Gauge |                          | The expiration date of certificates.                               |

=== "InfluxDB2"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="opt-traefik-config-reload-total" href="#opt-traefik-config-reload-total" title="#opt-traefik-config-reload-total">`traefik.config.reload.total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="opt-traefik-config-reload-lastSuccessTimestamp" href="#opt-traefik-config-reload-lastSuccessTimestamp" title="#opt-traefik-config-reload-lastSuccessTimestamp">`traefik.config.reload.lastSuccessTimestamp`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="opt-traefik-open-connections-3" href="#opt-traefik-open-connections-3" title="#opt-traefik-open-connections-3">`traefik.open.connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="opt-traefik-tls-certs-notAfterTimestamp" href="#opt-traefik-tls-certs-notAfterTimestamp" title="#opt-traefik-tls-certs-notAfterTimestamp">`traefik.tls.certs.notAfterTimestamp`</a> | Gauge |                          | The expiration date of certificates.                               |

=== "StatsD"
    | Metric       | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="opt-prefix-config-reload-total" href="#opt-prefix-config-reload-total" title="#opt-prefix-config-reload-total">`{prefix}.config.reload.total`</a> | Count |     | The total count of configuration reloads. |
    | <a id="opt-prefix-config-reload-lastSuccessTimestamp" href="#opt-prefix-config-reload-lastSuccessTimestamp" title="#opt-prefix-config-reload-lastSuccessTimestamp">`{prefix}.config.reload.lastSuccessTimestamp`</a> | Gauge |          | The timestamp of the last configuration reload success.            |
    | <a id="opt-prefix-open-connections" href="#opt-prefix-open-connections" title="#opt-prefix-open-connections">`{prefix}.open.connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="opt-prefix-tls-certs-notAfterTimestamp" href="#opt-prefix-tls-certs-notAfterTimestamp" title="#opt-prefix-tls-certs-notAfterTimestamp">`{prefix}.tls.certs.notAfterTimestamp`</a> | Gauge |    | The expiration date of certificates.   |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Labels

Here is a comprehensive list of labels that are provided by the global metrics:

| Label        | Description      | example              |
|--------------|----------------------------------------|----------------------|
| <a id="opt-entrypoint" href="#opt-entrypoint" title="#opt-entrypoint">`entrypoint`</a> | Entrypoint that handled the connection | "example_entrypoint" |
| <a id="opt-protocol" href="#opt-protocol" title="#opt-protocol">`protocol`</a> | Connection protocol     | "TCP"      |

### OpenTelemetry Semantic Conventions

Traefik Proxy follows [official OpenTelemetry semantic conventions v1.23.1](https://github.com/open-telemetry/semantic-conventions/blob/v1.23.1/docs/http/http-metrics.md).

#### HTTP Server

| Metric     | Type      | [Labels](#labels)       | Description   |
|----------|-----------|-------------------------|------------------|
| <a id="opt-http-server-request-duration" href="#opt-http-server-request-duration" title="#opt-http-server-request-duration">`http.server.request.duration`</a> | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP server requests  |

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label     | Description   | example       |
|-----------------------------|--------|---------------|
| <a id="opt-error-type" href="#opt-error-type" title="#opt-error-type">`error.type`</a> | Describes a class of error the operation ended with          | "500"         |
| <a id="opt-http-request-method" href="#opt-http-request-method" title="#opt-http-request-method">`http.request.method`</a> | HTTP request method                                          | "GET"         |
| <a id="opt-http-response-status-code" href="#opt-http-response-status-code" title="#opt-http-response-status-code">`http.response.status_code`</a> | HTTP response status code                                    | "200"         |
| <a id="opt-network-protocol-name" href="#opt-network-protocol-name" title="#opt-network-protocol-name">`network.protocol.name`</a> | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| <a id="opt-network-protocol-version" href="#opt-network-protocol-version" title="#opt-network-protocol-version">`network.protocol.version`</a> | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| <a id="opt-server-address" href="#opt-server-address" title="#opt-server-address">`server.address`</a> | Name of the local HTTP server that received the request      | "example.com" |
| <a id="opt-server-port" href="#opt-server-port" title="#opt-server-port">`server.port`</a> | Port of the local HTTP server that received the request      | "80"          |
| <a id="opt-url-scheme" href="#opt-url-scheme" title="#opt-url-scheme">`url.scheme`</a> | The URI scheme component identifying the used protocol       | "http"        |

#### HTTP Client

| Metric    | Type      | [Labels](#labels)    | Description  |
|-------------------------------|-----------|-----------------|--------|
| <a id="opt-http-client-request-duration" href="#opt-http-client-request-duration" title="#opt-http-client-request-duration">`http.client.request.duration`</a> | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP client requests  |

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

<<<<<<< Updated upstream
| <a id="Label" href="#Label" title="#Label">Label</a>                                                                                            | Description     | example       |
|-------------------------------------------------------------------------------------------------------------------------------------------------|------------|---------------|
| <a id="opt-error-type-2" href="#opt-error-type-2" title="#opt-error-type-2">`error.type`</a> | Describes a class of error the operation ended with    | "500"   |
| <a id="opt-http-request-method-2" href="#opt-http-request-method-2" title="#opt-http-request-method-2">`http.request.method`</a> | HTTP request method  | "GET" |
| <a id="opt-http-response-status-code-2" href="#opt-http-response-status-code-2" title="#opt-http-response-status-code-2">`http.response.status_code`</a> | HTTP response status code  | "200" |
| <a id="opt-network-protocol-name-2" href="#opt-network-protocol-name-2" title="#opt-network-protocol-name-2">`network.protocol.name`</a> | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| <a id="opt-network-protocol-version-2" href="#opt-network-protocol-version-2" title="#opt-network-protocol-version-2">`network.protocol.version`</a> | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| <a id="opt-server-address-2" href="#opt-server-address-2" title="#opt-server-address-2">`server.address`</a> | Name of the local HTTP server that received the request      | "example.com" |
| <a id="opt-server-port-2" href="#opt-server-port-2" title="#opt-server-port-2">`server.port`</a> | Port of the local HTTP server that received the request      | "80"          |
| <a id="opt-url-scheme-2" href="#opt-url-scheme-2" title="#opt-url-scheme-2">`url.scheme`</a> | The URI scheme component identifying the used protocol       | "http"        |
=======
| <a id="opt-Label" href="#opt-Label" title="#opt-Label">Label</a> | Description     | example       |
| <a id="opt-row" href="#opt-row" title="#opt-row">------  -----</a> |------------|---------------|
| <a id="opt-error-type-3" href="#opt-error-type-3" title="#opt-error-type-3">`error.type`</a> | Describes a class of error the operation ended with    | "500"   |
| <a id="opt-http-request-method-3" href="#opt-http-request-method-3" title="#opt-http-request-method-3">`http.request.method`</a> | HTTP request method  | "GET" |
| <a id="opt-http-response-status-code-3" href="#opt-http-response-status-code-3" title="#opt-http-response-status-code-3">`http.response.status_code`</a> | HTTP response status code  | "200" |
| <a id="opt-network-protocol-name-3" href="#opt-network-protocol-name-3" title="#opt-network-protocol-name-3">`network.protocol.name`</a> | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| <a id="opt-network-protocol-version-3" href="#opt-network-protocol-version-3" title="#opt-network-protocol-version-3">`network.protocol.version`</a> | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| <a id="opt-server-address-3" href="#opt-server-address-3" title="#opt-server-address-3">`server.address`</a> | Name of the local HTTP server that received the request      | "example.com" |
| <a id="opt-server-port-3" href="#opt-server-port-3" title="#opt-server-port-3">`server.port`</a> | Port of the local HTTP server that received the request      | "80"          |
| <a id="opt-url-scheme-3" href="#opt-url-scheme-3" title="#opt-url-scheme-3">`url.scheme`</a> | The URI scheme component identifying the used protocol       | "http"        |
>>>>>>> Stashed changes

### HTTP Metrics

On top of the official OpenTelemetry semantic conventions, Traefik provides its own metrics to monitor the incoming traffic.

#### EntryPoint Metrics

=== "OpenTelemetry"

    | Metric   | Type      | [Labels](#labels)         | Description   |
    |-----------------------|-----------|--------------------|--------------------------|
    | <a id="opt-traefik-entrypoint-requests-total" href="#opt-traefik-entrypoint-requests-total" title="#opt-traefik-entrypoint-requests-total">`traefik_entrypoint_requests_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="opt-traefik-entrypoint-requests-tls-total" href="#opt-traefik-entrypoint-requests-tls-total" title="#opt-traefik-entrypoint-requests-tls-total">`traefik_entrypoint_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="opt-traefik-entrypoint-request-duration-seconds" href="#opt-traefik-entrypoint-request-duration-seconds" title="#opt-traefik-entrypoint-request-duration-seconds">`traefik_entrypoint_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="opt-traefik-entrypoint-requests-bytes-total" href="#opt-traefik-entrypoint-requests-bytes-total" title="#opt-traefik-entrypoint-requests-bytes-total">`traefik_entrypoint_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="opt-traefik-entrypoint-responses-bytes-total" href="#opt-traefik-entrypoint-responses-bytes-total" title="#opt-traefik-entrypoint-responses-bytes-total">`traefik_entrypoint_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |
    
=== "Prometheus"

    | Metric     | Type      | [Labels](#labels)      | Description      |
    |-----------------------|-----------|------------------------|-------------------------|
    | <a id="opt-traefik-entrypoint-requests-total-2" href="#opt-traefik-entrypoint-requests-total-2" title="#opt-traefik-entrypoint-requests-total-2">`traefik_entrypoint_requests_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="opt-traefik-entrypoint-requests-tls-total-2" href="#opt-traefik-entrypoint-requests-tls-total-2" title="#opt-traefik-entrypoint-requests-tls-total-2">`traefik_entrypoint_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="opt-traefik-entrypoint-request-duration-seconds-2" href="#opt-traefik-entrypoint-request-duration-seconds-2" title="#opt-traefik-entrypoint-request-duration-seconds-2">`traefik_entrypoint_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="opt-traefik-entrypoint-requests-bytes-total-2" href="#opt-traefik-entrypoint-requests-bytes-total-2" title="#opt-traefik-entrypoint-requests-bytes-total-2">`traefik_entrypoint_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="opt-traefik-entrypoint-responses-bytes-total-2" href="#opt-traefik-entrypoint-responses-bytes-total-2" title="#opt-traefik-entrypoint-responses-bytes-total-2">`traefik_entrypoint_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "Datadog"

    | Metric   | Type      | [Labels](#labels)     | Description     |
    |-----------------------|-----------|------------------|---------------------------|
    | <a id="opt-entrypoint-requests-total" href="#opt-entrypoint-requests-total" title="#opt-entrypoint-requests-total">`entrypoint.requests.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="opt-entrypoint-requests-tls-total" href="#opt-entrypoint-requests-tls-total" title="#opt-entrypoint-requests-tls-total">`entrypoint.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="opt-entrypoint-request-duration-seconds" href="#opt-entrypoint-request-duration-seconds" title="#opt-entrypoint-request-duration-seconds">`entrypoint.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="opt-entrypoint-requests-bytes-total" href="#opt-entrypoint-requests-bytes-total" title="#opt-entrypoint-requests-bytes-total">`entrypoint.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="opt-entrypoint-responses-bytes-total" href="#opt-entrypoint-responses-bytes-total" title="#opt-entrypoint-responses-bytes-total">`entrypoint.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "InfluxDB2"

    | Metric    | Type      | [Labels](#labels)   | Description     |
    |------------|-----------|-------------------|-----------------|
    | <a id="opt-traefik-entrypoint-requests-total-3" href="#opt-traefik-entrypoint-requests-total-3" title="#opt-traefik-entrypoint-requests-total-3">`traefik.entrypoint.requests.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="opt-traefik-entrypoint-requests-tls-total-3" href="#opt-traefik-entrypoint-requests-tls-total-3" title="#opt-traefik-entrypoint-requests-tls-total-3">`traefik.entrypoint.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="opt-traefik-entrypoint-request-duration-seconds-3" href="#opt-traefik-entrypoint-request-duration-seconds-3" title="#opt-traefik-entrypoint-request-duration-seconds-3">`traefik.entrypoint.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="opt-traefik-entrypoint-requests-bytes-total-3" href="#opt-traefik-entrypoint-requests-bytes-total-3" title="#opt-traefik-entrypoint-requests-bytes-total-3">`traefik.entrypoint.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="opt-traefik-entrypoint-responses-bytes-total-3" href="#opt-traefik-entrypoint-responses-bytes-total-3" title="#opt-traefik-entrypoint-responses-bytes-total-3">`traefik.entrypoint.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "StatsD"

    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="opt-prefix-entrypoint-requests-total" href="#opt-prefix-entrypoint-requests-total" title="#opt-prefix-entrypoint-requests-total">`{prefix}.entrypoint.requests.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="opt-prefix-entrypoint-requests-tls-total" href="#opt-prefix-entrypoint-requests-tls-total" title="#opt-prefix-entrypoint-requests-tls-total">`{prefix}.entrypoint.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="opt-prefix-entrypoint-request-duration-seconds" href="#opt-prefix-entrypoint-request-duration-seconds" title="#opt-prefix-entrypoint-request-duration-seconds">`{prefix}.entrypoint.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="opt-prefix-entrypoint-requests-bytes-total" href="#opt-prefix-entrypoint-requests-bytes-total" title="#opt-prefix-entrypoint-requests-bytes-total">`{prefix}.entrypoint.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="opt-prefix-entrypoint-responses-bytes-total" href="#opt-prefix-entrypoint-responses-bytes-total" title="#opt-prefix-entrypoint-responses-bytes-total">`{prefix}.entrypoint.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Router Metrics

=== "OpenTelemetry"

    | Metric    | Type      | [Labels](#labels)         | Description           |
    |-----------------------|-----------|----------------------|--------------------------------|
    | <a id="opt-traefik-router-requests-total" href="#opt-traefik-router-requests-total" title="#opt-traefik-router-requests-total">`traefik_router_requests_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="opt-traefik-router-requests-tls-total" href="#opt-traefik-router-requests-tls-total" title="#opt-traefik-router-requests-tls-total">`traefik_router_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="opt-traefik-router-request-duration-seconds" href="#opt-traefik-router-request-duration-seconds" title="#opt-traefik-router-request-duration-seconds">`traefik_router_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="opt-traefik-router-requests-bytes-total" href="#opt-traefik-router-requests-bytes-total" title="#opt-traefik-router-requests-bytes-total">`traefik_router_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="opt-traefik-router-responses-bytes-total" href="#opt-traefik-router-responses-bytes-total" title="#opt-traefik-router-responses-bytes-total">`traefik_router_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |
    
=== "Prometheus"

    | Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
    |-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
    | <a id="opt-traefik-router-requests-total-2" href="#opt-traefik-router-requests-total-2" title="#opt-traefik-router-requests-total-2">`traefik_router_requests_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="opt-traefik-router-requests-tls-total-2" href="#opt-traefik-router-requests-tls-total-2" title="#opt-traefik-router-requests-tls-total-2">`traefik_router_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="opt-traefik-router-request-duration-seconds-2" href="#opt-traefik-router-request-duration-seconds-2" title="#opt-traefik-router-request-duration-seconds-2">`traefik_router_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="opt-traefik-router-requests-bytes-total-2" href="#opt-traefik-router-requests-bytes-total-2" title="#opt-traefik-router-requests-bytes-total-2">`traefik_router_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="opt-traefik-router-responses-bytes-total-2" href="#opt-traefik-router-responses-bytes-total-2" title="#opt-traefik-router-responses-bytes-total-2">`traefik_router_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "Datadog"

    | Metric    | Type      | [Labels](#labels)   | Description   |
    |-------------|-----------|---------------|---------------------|
    | <a id="opt-router-requests-total" href="#opt-router-requests-total" title="#opt-router-requests-total">`router.requests.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="opt-router-requests-tls-total" href="#opt-router-requests-tls-total" title="#opt-router-requests-tls-total">`router.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="opt-router-request-duration-seconds" href="#opt-router-request-duration-seconds" title="#opt-router-request-duration-seconds">`router.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="opt-router-requests-bytes-total" href="#opt-router-requests-bytes-total" title="#opt-router-requests-bytes-total">`router.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="opt-router-responses-bytes-total" href="#opt-router-responses-bytes-total" title="#opt-router-responses-bytes-total">`router.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "InfluxDB2"

    | Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
    |-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
    | <a id="opt-traefik-router-requests-total-3" href="#opt-traefik-router-requests-total-3" title="#opt-traefik-router-requests-total-3">`traefik.router.requests.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="opt-traefik-router-requests-tls-total-3" href="#opt-traefik-router-requests-tls-total-3" title="#opt-traefik-router-requests-tls-total-3">`traefik.router.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="opt-traefik-router-request-duration-seconds-3" href="#opt-traefik-router-request-duration-seconds-3" title="#opt-traefik-router-request-duration-seconds-3">`traefik.router.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="opt-traefik-router-requests-bytes-total-3" href="#opt-traefik-router-requests-bytes-total-3" title="#opt-traefik-router-requests-bytes-total-3">`traefik.router.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="opt-traefik-router-responses-bytes-total-3" href="#opt-traefik-router-responses-bytes-total-3" title="#opt-traefik-router-responses-bytes-total-3">`traefik.router.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "StatsD"

    | Metric     | Type      | [Labels](#labels)      | Description   |
    |-----------------------|-----------|---------------|-------------|
    | <a id="opt-prefix-router-requests-total" href="#opt-prefix-router-requests-total" title="#opt-prefix-router-requests-total">`{prefix}.router.requests.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="opt-prefix-router-requests-tls-total" href="#opt-prefix-router-requests-tls-total" title="#opt-prefix-router-requests-tls-total">`{prefix}.router.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="opt-prefix-router-request-duration-seconds" href="#opt-prefix-router-request-duration-seconds" title="#opt-prefix-router-request-duration-seconds">`{prefix}.router.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="opt-prefix-router-requests-bytes-total" href="#opt-prefix-router-requests-bytes-total" title="#opt-prefix-router-requests-bytes-total">`{prefix}.router.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="opt-prefix-router-responses-bytes-total" href="#opt-prefix-router-responses-bytes-total" title="#opt-prefix-router-responses-bytes-total">`{prefix}.router.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Service Metrics

=== "OpenTelemetry"

    | Metric    | Type      | Labels      | Description     |
    |-----------------------|-----------|------------|------------|
    | <a id="opt-traefik-service-requests-total" href="#opt-traefik-service-requests-total" title="#opt-traefik-service-requests-total">`traefik_service_requests_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="opt-traefik-service-requests-tls-total" href="#opt-traefik-service-requests-tls-total" title="#opt-traefik-service-requests-tls-total">`traefik_service_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="opt-traefik-service-request-duration-seconds" href="#opt-traefik-service-request-duration-seconds" title="#opt-traefik-service-request-duration-seconds">`traefik_service_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="opt-traefik-service-retries-total" href="#opt-traefik-service-retries-total" title="#opt-traefik-service-retries-total">`traefik_service_retries_total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="opt-traefik-service-server-up" href="#opt-traefik-service-server-up" title="#opt-traefik-service-server-up">`traefik_service_server_up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="opt-traefik-service-requests-bytes-total" href="#opt-traefik-service-requests-bytes-total" title="#opt-traefik-service-requests-bytes-total">`traefik_service_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="opt-traefik-service-responses-bytes-total" href="#opt-traefik-service-responses-bytes-total" title="#opt-traefik-service-responses-bytes-total">`traefik_service_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |
    
=== "Prometheus"

    | Metric    | Type      | Labels    | Description    |
    |-----------------------|-----------|-------|------------|
    | <a id="opt-traefik-service-requests-total-2" href="#opt-traefik-service-requests-total-2" title="#opt-traefik-service-requests-total-2">`traefik_service_requests_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="opt-traefik-service-requests-tls-total-2" href="#opt-traefik-service-requests-tls-total-2" title="#opt-traefik-service-requests-tls-total-2">`traefik_service_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="opt-traefik-service-request-duration-seconds-2" href="#opt-traefik-service-request-duration-seconds-2" title="#opt-traefik-service-request-duration-seconds-2">`traefik_service_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="opt-traefik-service-retries-total-2" href="#opt-traefik-service-retries-total-2" title="#opt-traefik-service-retries-total-2">`traefik_service_retries_total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="opt-traefik-service-server-up-2" href="#opt-traefik-service-server-up-2" title="#opt-traefik-service-server-up-2">`traefik_service_server_up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="opt-traefik-service-requests-bytes-total-2" href="#opt-traefik-service-requests-bytes-total-2" title="#opt-traefik-service-requests-bytes-total-2">`traefik_service_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="opt-traefik-service-responses-bytes-total-2" href="#opt-traefik-service-responses-bytes-total-2" title="#opt-traefik-service-responses-bytes-total-2">`traefik_service_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "Datadog"

    | Metric    | Type      | Labels    | Description |
    |-----------------------|-----------|--------|------------------|
    | <a id="opt-service-requests-total" href="#opt-service-requests-total" title="#opt-service-requests-total">`service.requests.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="opt-router-service-tls-total" href="#opt-router-service-tls-total" title="#opt-router-service-tls-total">`router.service.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="opt-service-request-duration-seconds" href="#opt-service-request-duration-seconds" title="#opt-service-request-duration-seconds">`service.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="opt-service-retries-total" href="#opt-service-retries-total" title="#opt-service-retries-total">`service.retries.total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="opt-service-server-up" href="#opt-service-server-up" title="#opt-service-server-up">`service.server.up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="opt-service-requests-bytes-total" href="#opt-service-requests-bytes-total" title="#opt-service-requests-bytes-total">`service.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="opt-service-responses-bytes-total" href="#opt-service-responses-bytes-total" title="#opt-service-responses-bytes-total">`service.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "InfluxDB2"

    | Metric                | Type      | Labels                                  | Description                                                 |
    |-----------------------|-----------|-----------------------------------------|-------------------------------------------------------------|
    | <a id="opt-traefik-service-requests-total-3" href="#opt-traefik-service-requests-total-3" title="#opt-traefik-service-requests-total-3">`traefik.service.requests.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="opt-traefik-service-requests-tls-total-3" href="#opt-traefik-service-requests-tls-total-3" title="#opt-traefik-service-requests-tls-total-3">`traefik.service.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="opt-traefik-service-request-duration-seconds-3" href="#opt-traefik-service-request-duration-seconds-3" title="#opt-traefik-service-request-duration-seconds-3">`traefik.service.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="opt-traefik-service-retries-total-3" href="#opt-traefik-service-retries-total-3" title="#opt-traefik-service-retries-total-3">`traefik.service.retries.total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="opt-traefik-service-server-up-3" href="#opt-traefik-service-server-up-3" title="#opt-traefik-service-server-up-3">`traefik.service.server.up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="opt-traefik-service-requests-bytes-total-3" href="#opt-traefik-service-requests-bytes-total-3" title="#opt-traefik-service-requests-bytes-total-3">`traefik.service.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="opt-traefik-service-responses-bytes-total-3" href="#opt-traefik-service-responses-bytes-total-3" title="#opt-traefik-service-responses-bytes-total-3">`traefik.service.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "StatsD"

    | Metric                | Type      | Labels   | Description    |
    |-----------------------|-----------|-----|---------|
    | <a id="opt-prefix-service-requests-total" href="#opt-prefix-service-requests-total" title="#opt-prefix-service-requests-total">`{prefix}.service.requests.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="opt-prefix-service-requests-tls-total" href="#opt-prefix-service-requests-tls-total" title="#opt-prefix-service-requests-tls-total">`{prefix}.service.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="opt-prefix-service-request-duration-seconds" href="#opt-prefix-service-request-duration-seconds" title="#opt-prefix-service-request-duration-seconds">`{prefix}.service.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="opt-prefix-service-retries-total" href="#opt-prefix-service-retries-total" title="#opt-prefix-service-retries-total">`{prefix}.service.retries.total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="opt-prefix-service-server-up" href="#opt-prefix-service-server-up" title="#opt-prefix-service-server-up">`{prefix}.service.server.up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="opt-prefix-service-requests-bytes-total" href="#opt-prefix-service-requests-bytes-total" title="#opt-prefix-service-requests-bytes-total">`{prefix}.service.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="opt-prefix-service-responses-bytes-total" href="#opt-prefix-service-responses-bytes-total" title="#opt-prefix-service-responses-bytes-total">`{prefix}.service.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label         | Description      | example      |
|---------------|-------------------|----------------------------|
| <a id="opt-cn" href="#opt-cn" title="#opt-cn">`cn`</a> | Certificate Common Name     | "example.com"     |
| <a id="opt-code" href="#opt-code" title="#opt-code">`code`</a> | Request code       | "200"                      |
| <a id="opt-entrypoint-2" href="#opt-entrypoint-2" title="#opt-entrypoint-2">`entrypoint`</a> | Entrypoint that handled the request   | "example_entrypoint"       |
| <a id="opt-method" href="#opt-method" title="#opt-method">`method`</a> | Request Method     | "GET"    |
| <a id="opt-protocol-2" href="#opt-protocol-2" title="#opt-protocol-2">`protocol`</a> | Request protocol      | "http"                     |
| <a id="opt-router" href="#opt-router" title="#opt-router">`router`</a> | Router that handled the request       | "example_router"    |
| <a id="opt-sans" href="#opt-sans" title="#opt-sans">`sans`</a> | Certificate Subject Alternative NameS | "example.com"              |
| <a id="opt-serial" href="#opt-serial" title="#opt-serial">`serial`</a> | Certificate Serial Number   | "123..."                   |
| <a id="opt-service" href="#opt-service" title="#opt-service">`service`</a> | Service that handled the request      | "example_service@provider" |
| <a id="opt-tls-cipher" href="#opt-tls-cipher" title="#opt-tls-cipher">`tls_cipher`</a> | TLS cipher used for the request       | "TLS_FALLBACK_SCSV"        |
| <a id="opt-tls-version" href="#opt-tls-version" title="#opt-tls-version">`tls_version`</a> | TLS version used for the request      | "1.0"                      |
| <a id="opt-url" href="#opt-url" title="#opt-url">`url`</a> | Service server url                    | "http://example.com"       |

!!! info "`method` label value"

    If the HTTP method verb on a request is not one defined in the set of common methods for [`HTTP/1.1`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
    or the [`PRI`](https://datatracker.ietf.org/doc/html/rfc7540#section-11.6) verb (for `HTTP/2`),
    then the value for the method label becomes `EXTENSION_METHOD`.
