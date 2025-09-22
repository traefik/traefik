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
| <a id="metrics-addInternals" href="#metrics-addInternals" title="#metrics-addInternals">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internal`).                                                                                                  | false                                              | No       |
| <a id="metrics-otlp-serviceName" href="#metrics-otlp-serviceName" title="#metrics-otlp-serviceName">`metrics.otlp.serviceName`</a> | Defines the service name resource attribute.                                                                                                                     | "traefik"                                          | No       |
| <a id="metrics-otlp-resourceAttributes" href="#metrics-otlp-resourceAttributes" title="#metrics-otlp-resourceAttributes">`metrics.otlp.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                                              | []                                                 | No       |
| <a id="metrics-otlp-addEntryPointsLabels" href="#metrics-otlp-addEntryPointsLabels" title="#metrics-otlp-addEntryPointsLabels">`metrics.otlp.addEntryPointsLabels`</a> | Enable metrics on entry points.                                                                                                                                  | true                                               | No       |
| <a id="metrics-otlp-addRoutersLabels" href="#metrics-otlp-addRoutersLabels" title="#metrics-otlp-addRoutersLabels">`metrics.otlp.addRoutersLabels`</a> | Enable metrics on routers.                                                                                                                                       | false                                              | No       |
| <a id="metrics-otlp-addServicesLabels" href="#metrics-otlp-addServicesLabels" title="#metrics-otlp-addServicesLabels">`metrics.otlp.addServicesLabels`</a> | Enable metrics on services.                                                                                                                                      | true                                               | No       |
| <a id="metrics-otlp-explicitBoundaries" href="#metrics-otlp-explicitBoundaries" title="#metrics-otlp-explicitBoundaries">`metrics.otlp.explicitBoundaries`</a> | Explicit boundaries for Histogram data points.                                                                                                                   | ".005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10" | No       |
| <a id="metrics-otlp-pushInterval" href="#metrics-otlp-pushInterval" title="#metrics-otlp-pushInterval">`metrics.otlp.pushInterval`</a> | Interval at which metrics are sent to the OpenTelemetry Collector.                                                                                               | 10s                                                | No       |
| <a id="metrics-otlp-http" href="#metrics-otlp-http" title="#metrics-otlp-http">`metrics.otlp.http`</a> | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| <a id="metrics-otlp-http-endpoint" href="#metrics-otlp-http-endpoint" title="#metrics-otlp-http-endpoint">`metrics.otlp.http.endpoint`</a> | URL of the OpenTelemetry Collector to send metrics to.<br /> Format="`<scheme>://<host>:<port><path>`"                                                           | "http://localhost:4318/v1/metrics"                 | Yes      |
| <a id="metrics-otlp-http-headers" href="#metrics-otlp-http-headers" title="#metrics-otlp-http-headers">`metrics.otlp.http.headers`</a> | Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.                                                                             | -                                                  | No       |
| <a id="metrics-otlp-http-tls-ca" href="#metrics-otlp-http-tls-ca" title="#metrics-otlp-http-tls-ca">`metrics.otlp.http.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector,<br />it defaults to the system bundle.                          | ""                                                 | No       |
| <a id="metrics-otlp-http-tls-cert" href="#metrics-otlp-http-tls-cert" title="#metrics-otlp-http-tls-cert">`metrics.otlp.http.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector.<br />When using this option, setting the `key` option is required. | ""                                                 | No       |
| <a id="metrics-otlp-http-tls-key" href="#metrics-otlp-http-tls-key" title="#metrics-otlp-http-tls-key">`metrics.otlp.http.tls.key`</a> | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| <a id="metrics-otlp-http-tls-insecureskipverify" href="#metrics-otlp-http-tls-insecureskipverify" title="#metrics-otlp-http-tls-insecureskipverify">`metrics.otlp.http.tls.insecureskipverify`</a> | Allow the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.                   | false                                              | Yes      |
| <a id="metrics-otlp-grpc" href="#metrics-otlp-grpc" title="#metrics-otlp-grpc">`metrics.otlp.grpc`</a> | This instructs the exporter to send metrics to the OpenTelemetry Collector using gRPC.                                                                           | null/false                                         | No       |
| <a id="metrics-otlp-grpc-endpoint" href="#metrics-otlp-grpc-endpoint" title="#metrics-otlp-grpc-endpoint">`metrics.otlp.grpc.endpoint`</a> | Address of the OpenTelemetry Collector to send metrics to.<br /> Format="`<host>:<port>`"                                                                        | "localhost:4317"                                   | Yes      |
| <a id="metrics-otlp-grpc-headers" href="#metrics-otlp-grpc-headers" title="#metrics-otlp-grpc-headers">`metrics.otlp.grpc.headers`</a> | Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.                                                                             | -                                                  | No       |
| <a id="metrics-otlp-http-grpc-insecure" href="#metrics-otlp-http-grpc-insecure" title="#metrics-otlp-http-grpc-insecure">`metrics.otlp.http.grpc.insecure`</a> | Allows exporter to send metrics to the OpenTelemetry Collector without using a secured protocol.                                                                 | false                                              | Yes      |
| <a id="metrics-otlp-grpc-tls-ca" href="#metrics-otlp-grpc-tls-ca" title="#metrics-otlp-grpc-tls-ca">`metrics.otlp.grpc.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector,<br />it defaults to the system bundle.                          | -                                                  | No       |
| <a id="metrics-otlp-grpc-tls-cert" href="#metrics-otlp-grpc-tls-cert" title="#metrics-otlp-grpc-tls-cert">`metrics.otlp.grpc.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector.<br />When using this option, setting the `key` option is required. | -                                                  | No       |
| <a id="metrics-otlp-grpc-tls-key" href="#metrics-otlp-grpc-tls-key" title="#metrics-otlp-grpc-tls-key">`metrics.otlp.grpc.tls.key`</a> | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| <a id="metrics-otlp-grpc-tls-insecureskipverify" href="#metrics-otlp-grpc-tls-insecureskipverify" title="#metrics-otlp-grpc-tls-insecureskipverify">`metrics.otlp.grpc.tls.insecureskipverify`</a> | Allow the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.                   | false                                              | Yes      |

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
| <a id="metrics-addInternals-2" href="#metrics-addInternals-2" title="#metrics-addInternals-2">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internal`). | false      | No      |
| <a id="datadog-address" href="#datadog-address" title="#datadog-address">`datadog.address`</a> | Defines the address for the exporter to send metrics to datadog-agent. More information [here](#address)|  `127.0.0.1:8125`     | Yes   |
| <a id="datadog-addEntryPointsLabels" href="#datadog-addEntryPointsLabels" title="#datadog-addEntryPointsLabels">`datadog.addEntryPointsLabels`</a> | Enable metrics on entry points. |  true   | No   |
| <a id="datadog-addRoutersLabels" href="#datadog-addRoutersLabels" title="#datadog-addRoutersLabels">`datadog.addRoutersLabels`</a> | Enable metrics on routers. |  false   | No   |
| <a id="datadog-addServicesLabels" href="#datadog-addServicesLabels" title="#datadog-addServicesLabels">`datadog.addServicesLabels`</a> | Enable metrics on services. |  true   | No   |
| <a id="datadog-pushInterval" href="#datadog-pushInterval" title="#datadog-pushInterval">`datadog.pushInterval`</a> | Defines the interval used by the exporter to push metrics to datadog-agent. |  10s   | No   |
| <a id="datadog-prefix" href="#datadog-prefix" title="#datadog-prefix">`datadog.prefix`</a> | Defines the prefix to use for metrics collection. |  "traefik"   | No   |

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
| <a id="metrics-addInternal" href="#metrics-addInternal" title="#metrics-addInternal">`metrics.addInternal`</a> | Enables metrics for internal resources (e.g.: `ping@internal`). | false      | No      |
| <a id="metrics-influxDB2-addEntryPointsLabels" href="#metrics-influxDB2-addEntryPointsLabels" title="#metrics-influxDB2-addEntryPointsLabels">`metrics.influxDB2.addEntryPointsLabels`</a> | Enable metrics on entry points. | true      | No      |
| <a id="metrics-influxDB2-addRoutersLabels" href="#metrics-influxDB2-addRoutersLabels" title="#metrics-influxDB2-addRoutersLabels">`metrics.influxDB2.addRoutersLabels`</a> | Enable metrics on routers. | false      | No      |
| <a id="metrics-influxDB2-addServicesLabels" href="#metrics-influxDB2-addServicesLabels" title="#metrics-influxDB2-addServicesLabels">`metrics.influxDB2.addServicesLabels`</a> | Enable metrics on services.| true      | No      |
| <a id="metrics-influxDB2-additionalLabels" href="#metrics-influxDB2-additionalLabels" title="#metrics-influxDB2-additionalLabels">`metrics.influxDB2.additionalLabels`</a> | Additional labels (InfluxDB tags) on all metrics. | - | No      |
| <a id="metrics-influxDB2-pushInterval" href="#metrics-influxDB2-pushInterval" title="#metrics-influxDB2-pushInterval">`metrics.influxDB2.pushInterval`</a> | The interval used by the exporter to push metrics to InfluxDB server. | 10s      | No      |
| <a id="metrics-influxDB2-address" href="#metrics-influxDB2-address" title="#metrics-influxDB2-address">`metrics.influxDB2.address`</a> | Address of the InfluxDB v2 instance. | "http://localhost:8086"     | Yes      |
| <a id="metrics-influxDB2-token" href="#metrics-influxDB2-token" title="#metrics-influxDB2-token">`metrics.influxDB2.token`</a> | Token with which to connect to InfluxDB v2. | - | Yes      |
| <a id="metrics-influxDB2-org" href="#metrics-influxDB2-org" title="#metrics-influxDB2-org">`metrics.influxDB2.org`</a> | Organisation where metrics will be stored. | -  | Yes      |
| <a id="metrics-influxDB2-bucket" href="#metrics-influxDB2-bucket" title="#metrics-influxDB2-bucket">`metrics.influxDB2.bucket`</a> | Bucket where metrics will be stored. | -  | Yes      |

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
| <a id="metrics-prometheus-addInternals" href="#metrics-prometheus-addInternals" title="#metrics-prometheus-addInternals">`metrics.prometheus.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internals`). | false      | No      |
| <a id="metrics-prometheus-addEntryPointsLabels" href="#metrics-prometheus-addEntryPointsLabels" title="#metrics-prometheus-addEntryPointsLabels">`metrics.prometheus.addEntryPointsLabels`</a> | Enable metrics on entry points. | true      | No      |
| <a id="metrics-prometheus-addRoutersLabels" href="#metrics-prometheus-addRoutersLabels" title="#metrics-prometheus-addRoutersLabels">`metrics.prometheus.addRoutersLabels`</a> | Enable metrics on routers. | false      | No      |
| <a id="metrics-prometheus-addServicesLabels" href="#metrics-prometheus-addServicesLabels" title="#metrics-prometheus-addServicesLabels">`metrics.prometheus.addServicesLabels`</a> | Enable metrics on services.| true      | No      |
| <a id="metrics-prometheus-buckets" href="#metrics-prometheus-buckets" title="#metrics-prometheus-buckets">`metrics.prometheus.buckets`</a> | Buckets for latency metrics. |"0.100000, 0.300000, 1.200000, 5.000000"  | No      |
| <a id="metrics-prometheus-manualRouting" href="#metrics-prometheus-manualRouting" title="#metrics-prometheus-manualRouting">`metrics.prometheus.manualRouting`</a> | Set to _true_, it disables the default internal router in order to allow creating a custom router for the `prometheus@internal` service. | false    | No      |
| <a id="metrics-prometheus-entryPoint" href="#metrics-prometheus-entryPoint" title="#metrics-prometheus-entryPoint">`metrics.prometheus.entryPoint`</a> | Traefik Entrypoint name used to expose metrics. | "traefik"     | No      |
| <a id="metrics-prometheus-headerLabels" href="#metrics-prometheus-headerLabels" title="#metrics-prometheus-headerLabels">`metrics.prometheus.headerLabels`</a> | Defines extra labels extracted from request headers for the `requests_total` metrics.<br />More information [here](#headerlabels). |       | Yes      |

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
| <a id="metrics-addInternals-3" href="#metrics-addInternals-3" title="#metrics-addInternals-3">`metrics.addInternals`</a> | Enables metrics for internal resources (e.g.: `ping@internals`). | false      | No      |
| <a id="metrics-statsD-addEntryPointsLabels" href="#metrics-statsD-addEntryPointsLabels" title="#metrics-statsD-addEntryPointsLabels">`metrics.statsD.addEntryPointsLabels`</a> | Enable metrics on entry points. | true      | No      |
| <a id="metrics-statsD-addRoutersLabels" href="#metrics-statsD-addRoutersLabels" title="#metrics-statsD-addRoutersLabels">`metrics.statsD.addRoutersLabels`</a> | Enable metrics on routers. | false      | No      |
| <a id="metrics-statsD-addServicesLabels" href="#metrics-statsD-addServicesLabels" title="#metrics-statsD-addServicesLabels">`metrics.statsD.addServicesLabels`</a> | Enable metrics on services.| true      | No      |
| <a id="metrics-statsD-pushInterval" href="#metrics-statsD-pushInterval" title="#metrics-statsD-pushInterval">`metrics.statsD.pushInterval`</a> | The interval used by the exporter to push metrics to DataDog server. | 10s      | No      |
| <a id="metrics-statsD-address" href="#metrics-statsD-address" title="#metrics-statsD-address">`metrics.statsD.address`</a> | Address instructs exporter to send metrics to statsd at this address.  | "127.0.0.1:8125"     | Yes      |
| <a id="metrics-statsD-prefix" href="#metrics-statsD-prefix" title="#metrics-statsD-prefix">`metrics.statsD.prefix`</a> | The prefix to use for metrics collection. | "traefik"      | No      |

## Metrics Provided

### Global Metrics

=== "OpenTelemetry"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="traefik-config-reloads-total" href="#traefik-config-reloads-total" title="#traefik-config-reloads-total">`traefik_config_reloads_total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="traefik-config-last-reload-success" href="#traefik-config-last-reload-success" title="#traefik-config-last-reload-success">`traefik_config_last_reload_success`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="traefik-open-connections" href="#traefik-open-connections" title="#traefik-open-connections">`traefik_open_connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="traefik-tls-certs-not-after" href="#traefik-tls-certs-not-after" title="#traefik-tls-certs-not-after">`traefik_tls_certs_not_after`</a> | Gauge |                          | The expiration date of certificates.                               |
    
=== "Prometheus"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="traefik-config-reloads-total-2" href="#traefik-config-reloads-total-2" title="#traefik-config-reloads-total-2">`traefik_config_reloads_total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="traefik-config-last-reload-success-2" href="#traefik-config-last-reload-success-2" title="#traefik-config-last-reload-success-2">`traefik_config_last_reload_success`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="traefik-open-connections-2" href="#traefik-open-connections-2" title="#traefik-open-connections-2">`traefik_open_connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="traefik-tls-certs-not-after-2" href="#traefik-tls-certs-not-after-2" title="#traefik-tls-certs-not-after-2">`traefik_tls_certs_not_after`</a> | Gauge |      | The expiration date of certificates. |

=== "Datadog"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="config-reload-total" href="#config-reload-total" title="#config-reload-total">`config.reload.total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="config-reload-lastSuccessTimestamp" href="#config-reload-lastSuccessTimestamp" title="#config-reload-lastSuccessTimestamp">`config.reload.lastSuccessTimestamp`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="open-connections" href="#open-connections" title="#open-connections">`open.connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="tls-certs-notAfterTimestamp" href="#tls-certs-notAfterTimestamp" title="#tls-certs-notAfterTimestamp">`tls.certs.notAfterTimestamp`</a> | Gauge |                          | The expiration date of certificates.                               |

=== "InfluxDB2"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="traefik-config-reload-total" href="#traefik-config-reload-total" title="#traefik-config-reload-total">`traefik.config.reload.total`</a> | Count |                          | The total count of configuration reloads.                          |
    | <a id="traefik-config-reload-lastSuccessTimestamp" href="#traefik-config-reload-lastSuccessTimestamp" title="#traefik-config-reload-lastSuccessTimestamp">`traefik.config.reload.lastSuccessTimestamp`</a> | Gauge |                          | The timestamp of the last configuration reload success.            |
    | <a id="traefik-open-connections-3" href="#traefik-open-connections-3" title="#traefik-open-connections-3">`traefik.open.connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="traefik-tls-certs-notAfterTimestamp" href="#traefik-tls-certs-notAfterTimestamp" title="#traefik-tls-certs-notAfterTimestamp">`traefik.tls.certs.notAfterTimestamp`</a> | Gauge |                          | The expiration date of certificates.                               |

=== "StatsD"
    | Metric       | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="prefix-config-reload-total" href="#prefix-config-reload-total" title="#prefix-config-reload-total">`{prefix}.config.reload.total`</a> | Count |     | The total count of configuration reloads. |
    | <a id="prefix-config-reload-lastSuccessTimestamp" href="#prefix-config-reload-lastSuccessTimestamp" title="#prefix-config-reload-lastSuccessTimestamp">`{prefix}.config.reload.lastSuccessTimestamp`</a> | Gauge |          | The timestamp of the last configuration reload success.            |
    | <a id="prefix-open-connections" href="#prefix-open-connections" title="#prefix-open-connections">`{prefix}.open.connections`</a> | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | <a id="prefix-tls-certs-notAfterTimestamp" href="#prefix-tls-certs-notAfterTimestamp" title="#prefix-tls-certs-notAfterTimestamp">`{prefix}.tls.certs.notAfterTimestamp`</a> | Gauge |    | The expiration date of certificates.   |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Labels

Here is a comprehensive list of labels that are provided by the global metrics:

| Label        | Description      | example              |
|--------------|----------------------------------------|----------------------|
| <a id="entrypoint" href="#entrypoint" title="#entrypoint">`entrypoint`</a> | Entrypoint that handled the connection | "example_entrypoint" |
| <a id="protocol" href="#protocol" title="#protocol">`protocol`</a> | Connection protocol     | "TCP"      |

### OpenTelemetry Semantic Conventions

Traefik Proxy follows [official OpenTelemetry semantic conventions v1.23.1](https://github.com/open-telemetry/semantic-conventions/blob/v1.23.1/docs/http/http-metrics.md).

#### HTTP Server

| Metric     | Type      | [Labels](#labels)       | Description   |
|----------|-----------|-------------------------|------------------|
| <a id="http-server-request-duration" href="#http-server-request-duration" title="#http-server-request-duration">`http.server.request.duration`</a> | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP server requests  |

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label     | Description   | example       |
|-----------------------------|--------|---------------|
| <a id="error-type" href="#error-type" title="#error-type">`error.type`</a> | Describes a class of error the operation ended with          | "500"         |
| <a id="http-request-method" href="#http-request-method" title="#http-request-method">`http.request.method`</a> | HTTP request method                                          | "GET"         |
| <a id="http-response-status-code" href="#http-response-status-code" title="#http-response-status-code">`http.response.status_code`</a> | HTTP response status code                                    | "200"         |
| <a id="network-protocol-name" href="#network-protocol-name" title="#network-protocol-name">`network.protocol.name`</a> | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| <a id="network-protocol-version" href="#network-protocol-version" title="#network-protocol-version">`network.protocol.version`</a> | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| <a id="server-address" href="#server-address" title="#server-address">`server.address`</a> | Name of the local HTTP server that received the request      | "example.com" |
| <a id="server-port" href="#server-port" title="#server-port">`server.port`</a> | Port of the local HTTP server that received the request      | "80"          |
| <a id="url-scheme" href="#url-scheme" title="#url-scheme">`url.scheme`</a> | The URI scheme component identifying the used protocol       | "http"        |

#### HTTP Client

| Metric    | Type      | [Labels](#labels)    | Description  |
|-------------------------------|-----------|-----------------|--------|
| <a id="http-client-request-duration" href="#http-client-request-duration" title="#http-client-request-duration">`http.client.request.duration`</a> | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP client requests  |

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| <a id="Label" href="#Label" title="#Label">Label</a> | Description     | example       |
| <a id="row" href="#row" title="#row">------  -----</a> |------------|---------------|
| <a id="error-type-2" href="#error-type-2" title="#error-type-2">`error.type`</a> | Describes a class of error the operation ended with    | "500"   |
| <a id="http-request-method-2" href="#http-request-method-2" title="#http-request-method-2">`http.request.method`</a> | HTTP request method  | "GET" |
| <a id="http-response-status-code-2" href="#http-response-status-code-2" title="#http-response-status-code-2">`http.response.status_code`</a> | HTTP response status code  | "200" |
| <a id="network-protocol-name-2" href="#network-protocol-name-2" title="#network-protocol-name-2">`network.protocol.name`</a> | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| <a id="network-protocol-version-2" href="#network-protocol-version-2" title="#network-protocol-version-2">`network.protocol.version`</a> | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| <a id="server-address-2" href="#server-address-2" title="#server-address-2">`server.address`</a> | Name of the local HTTP server that received the request      | "example.com" |
| <a id="server-port-2" href="#server-port-2" title="#server-port-2">`server.port`</a> | Port of the local HTTP server that received the request      | "80"          |
| <a id="url-scheme-2" href="#url-scheme-2" title="#url-scheme-2">`url.scheme`</a> | The URI scheme component identifying the used protocol       | "http"        |

### HTTP Metrics

On top of the official OpenTelemetry semantic conventions, Traefik provides its own metrics to monitor the incoming traffic.

#### EntryPoint Metrics

=== "OpenTelemetry"

    | Metric   | Type      | [Labels](#labels)         | Description   |
    |-----------------------|-----------|--------------------|--------------------------|
    | <a id="traefik-entrypoint-requests-total" href="#traefik-entrypoint-requests-total" title="#traefik-entrypoint-requests-total">`traefik_entrypoint_requests_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="traefik-entrypoint-requests-tls-total" href="#traefik-entrypoint-requests-tls-total" title="#traefik-entrypoint-requests-tls-total">`traefik_entrypoint_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="traefik-entrypoint-request-duration-seconds" href="#traefik-entrypoint-request-duration-seconds" title="#traefik-entrypoint-request-duration-seconds">`traefik_entrypoint_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="traefik-entrypoint-requests-bytes-total" href="#traefik-entrypoint-requests-bytes-total" title="#traefik-entrypoint-requests-bytes-total">`traefik_entrypoint_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="traefik-entrypoint-responses-bytes-total" href="#traefik-entrypoint-responses-bytes-total" title="#traefik-entrypoint-responses-bytes-total">`traefik_entrypoint_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |
    
=== "Prometheus"

    | Metric     | Type      | [Labels](#labels)      | Description      |
    |-----------------------|-----------|------------------------|-------------------------|
    | <a id="traefik-entrypoint-requests-total-2" href="#traefik-entrypoint-requests-total-2" title="#traefik-entrypoint-requests-total-2">`traefik_entrypoint_requests_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="traefik-entrypoint-requests-tls-total-2" href="#traefik-entrypoint-requests-tls-total-2" title="#traefik-entrypoint-requests-tls-total-2">`traefik_entrypoint_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="traefik-entrypoint-request-duration-seconds-2" href="#traefik-entrypoint-request-duration-seconds-2" title="#traefik-entrypoint-request-duration-seconds-2">`traefik_entrypoint_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="traefik-entrypoint-requests-bytes-total-2" href="#traefik-entrypoint-requests-bytes-total-2" title="#traefik-entrypoint-requests-bytes-total-2">`traefik_entrypoint_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="traefik-entrypoint-responses-bytes-total-2" href="#traefik-entrypoint-responses-bytes-total-2" title="#traefik-entrypoint-responses-bytes-total-2">`traefik_entrypoint_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "Datadog"

    | Metric   | Type      | [Labels](#labels)     | Description     |
    |-----------------------|-----------|------------------|---------------------------|
    | <a id="entrypoint-requests-total" href="#entrypoint-requests-total" title="#entrypoint-requests-total">`entrypoint.requests.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="entrypoint-requests-tls-total" href="#entrypoint-requests-tls-total" title="#entrypoint-requests-tls-total">`entrypoint.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="entrypoint-request-duration-seconds" href="#entrypoint-request-duration-seconds" title="#entrypoint-request-duration-seconds">`entrypoint.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="entrypoint-requests-bytes-total" href="#entrypoint-requests-bytes-total" title="#entrypoint-requests-bytes-total">`entrypoint.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="entrypoint-responses-bytes-total" href="#entrypoint-responses-bytes-total" title="#entrypoint-responses-bytes-total">`entrypoint.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "InfluxDB2"

    | Metric    | Type      | [Labels](#labels)   | Description     |
    |------------|-----------|-------------------|-----------------|
    | <a id="traefik-entrypoint-requests-total-3" href="#traefik-entrypoint-requests-total-3" title="#traefik-entrypoint-requests-total-3">`traefik.entrypoint.requests.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="traefik-entrypoint-requests-tls-total-3" href="#traefik-entrypoint-requests-tls-total-3" title="#traefik-entrypoint-requests-tls-total-3">`traefik.entrypoint.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="traefik-entrypoint-request-duration-seconds-3" href="#traefik-entrypoint-request-duration-seconds-3" title="#traefik-entrypoint-request-duration-seconds-3">`traefik.entrypoint.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="traefik-entrypoint-requests-bytes-total-3" href="#traefik-entrypoint-requests-bytes-total-3" title="#traefik-entrypoint-requests-bytes-total-3">`traefik.entrypoint.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="traefik-entrypoint-responses-bytes-total-3" href="#traefik-entrypoint-responses-bytes-total-3" title="#traefik-entrypoint-responses-bytes-total-3">`traefik.entrypoint.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "StatsD"

    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | <a id="prefix-entrypoint-requests-total" href="#prefix-entrypoint-requests-total" title="#prefix-entrypoint-requests-total">`{prefix}.entrypoint.requests.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | <a id="prefix-entrypoint-requests-tls-total" href="#prefix-entrypoint-requests-tls-total" title="#prefix-entrypoint-requests-tls-total">`{prefix}.entrypoint.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | <a id="prefix-entrypoint-request-duration-seconds" href="#prefix-entrypoint-request-duration-seconds" title="#prefix-entrypoint-request-duration-seconds">`{prefix}.entrypoint.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | <a id="prefix-entrypoint-requests-bytes-total" href="#prefix-entrypoint-requests-bytes-total" title="#prefix-entrypoint-requests-bytes-total">`{prefix}.entrypoint.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | <a id="prefix-entrypoint-responses-bytes-total" href="#prefix-entrypoint-responses-bytes-total" title="#prefix-entrypoint-responses-bytes-total">`{prefix}.entrypoint.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Router Metrics

=== "OpenTelemetry"

    | Metric    | Type      | [Labels](#labels)         | Description           |
    |-----------------------|-----------|----------------------|--------------------------------|
    | <a id="traefik-router-requests-total" href="#traefik-router-requests-total" title="#traefik-router-requests-total">`traefik_router_requests_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="traefik-router-requests-tls-total" href="#traefik-router-requests-tls-total" title="#traefik-router-requests-tls-total">`traefik_router_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="traefik-router-request-duration-seconds" href="#traefik-router-request-duration-seconds" title="#traefik-router-request-duration-seconds">`traefik_router_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="traefik-router-requests-bytes-total" href="#traefik-router-requests-bytes-total" title="#traefik-router-requests-bytes-total">`traefik_router_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="traefik-router-responses-bytes-total" href="#traefik-router-responses-bytes-total" title="#traefik-router-responses-bytes-total">`traefik_router_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |
    
=== "Prometheus"

    | Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
    |-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
    | <a id="traefik-router-requests-total-2" href="#traefik-router-requests-total-2" title="#traefik-router-requests-total-2">`traefik_router_requests_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="traefik-router-requests-tls-total-2" href="#traefik-router-requests-tls-total-2" title="#traefik-router-requests-tls-total-2">`traefik_router_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="traefik-router-request-duration-seconds-2" href="#traefik-router-request-duration-seconds-2" title="#traefik-router-request-duration-seconds-2">`traefik_router_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="traefik-router-requests-bytes-total-2" href="#traefik-router-requests-bytes-total-2" title="#traefik-router-requests-bytes-total-2">`traefik_router_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="traefik-router-responses-bytes-total-2" href="#traefik-router-responses-bytes-total-2" title="#traefik-router-responses-bytes-total-2">`traefik_router_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "Datadog"

    | Metric    | Type      | [Labels](#labels)   | Description   |
    |-------------|-----------|---------------|---------------------|
    | <a id="router-requests-total" href="#router-requests-total" title="#router-requests-total">`router.requests.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="router-requests-tls-total" href="#router-requests-tls-total" title="#router-requests-tls-total">`router.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="router-request-duration-seconds" href="#router-request-duration-seconds" title="#router-request-duration-seconds">`router.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="router-requests-bytes-total" href="#router-requests-bytes-total" title="#router-requests-bytes-total">`router.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="router-responses-bytes-total" href="#router-responses-bytes-total" title="#router-responses-bytes-total">`router.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "InfluxDB2"

    | Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
    |-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
    | <a id="traefik-router-requests-total-3" href="#traefik-router-requests-total-3" title="#traefik-router-requests-total-3">`traefik.router.requests.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="traefik-router-requests-tls-total-3" href="#traefik-router-requests-tls-total-3" title="#traefik-router-requests-tls-total-3">`traefik.router.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="traefik-router-request-duration-seconds-3" href="#traefik-router-request-duration-seconds-3" title="#traefik-router-request-duration-seconds-3">`traefik.router.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="traefik-router-requests-bytes-total-3" href="#traefik-router-requests-bytes-total-3" title="#traefik-router-requests-bytes-total-3">`traefik.router.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="traefik-router-responses-bytes-total-3" href="#traefik-router-responses-bytes-total-3" title="#traefik-router-responses-bytes-total-3">`traefik.router.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "StatsD"

    | Metric     | Type      | [Labels](#labels)      | Description   |
    |-----------------------|-----------|---------------|-------------|
    | <a id="prefix-router-requests-total" href="#prefix-router-requests-total" title="#prefix-router-requests-total">`{prefix}.router.requests.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | <a id="prefix-router-requests-tls-total" href="#prefix-router-requests-tls-total" title="#prefix-router-requests-tls-total">`{prefix}.router.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | <a id="prefix-router-request-duration-seconds" href="#prefix-router-request-duration-seconds" title="#prefix-router-request-duration-seconds">`{prefix}.router.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | <a id="prefix-router-requests-bytes-total" href="#prefix-router-requests-bytes-total" title="#prefix-router-requests-bytes-total">`{prefix}.router.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | <a id="prefix-router-responses-bytes-total" href="#prefix-router-responses-bytes-total" title="#prefix-router-responses-bytes-total">`{prefix}.router.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Service Metrics

=== "OpenTelemetry"

    | Metric    | Type      | Labels      | Description     |
    |-----------------------|-----------|------------|------------|
    | <a id="traefik-service-requests-total" href="#traefik-service-requests-total" title="#traefik-service-requests-total">`traefik_service_requests_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="traefik-service-requests-tls-total" href="#traefik-service-requests-tls-total" title="#traefik-service-requests-tls-total">`traefik_service_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="traefik-service-request-duration-seconds" href="#traefik-service-request-duration-seconds" title="#traefik-service-request-duration-seconds">`traefik_service_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="traefik-service-retries-total" href="#traefik-service-retries-total" title="#traefik-service-retries-total">`traefik_service_retries_total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="traefik-service-server-up" href="#traefik-service-server-up" title="#traefik-service-server-up">`traefik_service_server_up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="traefik-service-requests-bytes-total" href="#traefik-service-requests-bytes-total" title="#traefik-service-requests-bytes-total">`traefik_service_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="traefik-service-responses-bytes-total" href="#traefik-service-responses-bytes-total" title="#traefik-service-responses-bytes-total">`traefik_service_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |
    
=== "Prometheus"

    | Metric    | Type      | Labels    | Description    |
    |-----------------------|-----------|-------|------------|
    | <a id="traefik-service-requests-total-2" href="#traefik-service-requests-total-2" title="#traefik-service-requests-total-2">`traefik_service_requests_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="traefik-service-requests-tls-total-2" href="#traefik-service-requests-tls-total-2" title="#traefik-service-requests-tls-total-2">`traefik_service_requests_tls_total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="traefik-service-request-duration-seconds-2" href="#traefik-service-request-duration-seconds-2" title="#traefik-service-request-duration-seconds-2">`traefik_service_request_duration_seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="traefik-service-retries-total-2" href="#traefik-service-retries-total-2" title="#traefik-service-retries-total-2">`traefik_service_retries_total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="traefik-service-server-up-2" href="#traefik-service-server-up-2" title="#traefik-service-server-up-2">`traefik_service_server_up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="traefik-service-requests-bytes-total-2" href="#traefik-service-requests-bytes-total-2" title="#traefik-service-requests-bytes-total-2">`traefik_service_requests_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="traefik-service-responses-bytes-total-2" href="#traefik-service-responses-bytes-total-2" title="#traefik-service-responses-bytes-total-2">`traefik_service_responses_bytes_total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "Datadog"

    | Metric    | Type      | Labels    | Description |
    |-----------------------|-----------|--------|------------------|
    | <a id="service-requests-total" href="#service-requests-total" title="#service-requests-total">`service.requests.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="router-service-tls-total" href="#router-service-tls-total" title="#router-service-tls-total">`router.service.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="service-request-duration-seconds" href="#service-request-duration-seconds" title="#service-request-duration-seconds">`service.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="service-retries-total" href="#service-retries-total" title="#service-retries-total">`service.retries.total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="service-server-up" href="#service-server-up" title="#service-server-up">`service.server.up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="service-requests-bytes-total" href="#service-requests-bytes-total" title="#service-requests-bytes-total">`service.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="service-responses-bytes-total" href="#service-responses-bytes-total" title="#service-responses-bytes-total">`service.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "InfluxDB2"

    | Metric                | Type      | Labels                                  | Description                                                 |
    |-----------------------|-----------|-----------------------------------------|-------------------------------------------------------------|
    | <a id="traefik-service-requests-total-3" href="#traefik-service-requests-total-3" title="#traefik-service-requests-total-3">`traefik.service.requests.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="traefik-service-requests-tls-total-3" href="#traefik-service-requests-tls-total-3" title="#traefik-service-requests-tls-total-3">`traefik.service.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="traefik-service-request-duration-seconds-3" href="#traefik-service-request-duration-seconds-3" title="#traefik-service-request-duration-seconds-3">`traefik.service.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="traefik-service-retries-total-3" href="#traefik-service-retries-total-3" title="#traefik-service-retries-total-3">`traefik.service.retries.total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="traefik-service-server-up-3" href="#traefik-service-server-up-3" title="#traefik-service-server-up-3">`traefik.service.server.up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="traefik-service-requests-bytes-total-3" href="#traefik-service-requests-bytes-total-3" title="#traefik-service-requests-bytes-total-3">`traefik.service.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="traefik-service-responses-bytes-total-3" href="#traefik-service-responses-bytes-total-3" title="#traefik-service-responses-bytes-total-3">`traefik.service.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "StatsD"

    | Metric                | Type      | Labels   | Description    |
    |-----------------------|-----------|-----|---------|
    | <a id="prefix-service-requests-total" href="#prefix-service-requests-total" title="#prefix-service-requests-total">`{prefix}.service.requests.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | <a id="prefix-service-requests-tls-total" href="#prefix-service-requests-tls-total" title="#prefix-service-requests-tls-total">`{prefix}.service.requests.tls.total`</a> | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | <a id="prefix-service-request-duration-seconds" href="#prefix-service-request-duration-seconds" title="#prefix-service-request-duration-seconds">`{prefix}.service.request.duration.seconds`</a> | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | <a id="prefix-service-retries-total" href="#prefix-service-retries-total" title="#prefix-service-retries-total">`{prefix}.service.retries.total`</a> | Count     | `service`                               | The count of requests retries on a service.                 |
    | <a id="prefix-service-server-up" href="#prefix-service-server-up" title="#prefix-service-server-up">`{prefix}.service.server.up`</a> | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | <a id="prefix-service-requests-bytes-total" href="#prefix-service-requests-bytes-total" title="#prefix-service-requests-bytes-total">`{prefix}.service.requests.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | <a id="prefix-service-responses-bytes-total" href="#prefix-service-responses-bytes-total" title="#prefix-service-responses-bytes-total">`{prefix}.service.responses.bytes.total`</a> | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label         | Description      | example      |
|---------------|-------------------|----------------------------|
| <a id="cn" href="#cn" title="#cn">`cn`</a> | Certificate Common Name     | "example.com"     |
| <a id="code" href="#code" title="#code">`code`</a> | Request code       | "200"                      |
| <a id="entrypoint-2" href="#entrypoint-2" title="#entrypoint-2">`entrypoint`</a> | Entrypoint that handled the request   | "example_entrypoint"       |
| <a id="method" href="#method" title="#method">`method`</a> | Request Method     | "GET"    |
| <a id="protocol-2" href="#protocol-2" title="#protocol-2">`protocol`</a> | Request protocol      | "http"                     |
| <a id="router" href="#router" title="#router">`router`</a> | Router that handled the request       | "example_router"    |
| <a id="sans" href="#sans" title="#sans">`sans`</a> | Certificate Subject Alternative NameS | "example.com"              |
| <a id="serial" href="#serial" title="#serial">`serial`</a> | Certificate Serial Number   | "123..."                   |
| <a id="service" href="#service" title="#service">`service`</a> | Service that handled the request      | "example_service@provider" |
| <a id="tls-cipher" href="#tls-cipher" title="#tls-cipher">`tls_cipher`</a> | TLS cipher used for the request       | "TLS_FALLBACK_SCSV"        |
| <a id="tls-version" href="#tls-version" title="#tls-version">`tls_version`</a> | TLS version used for the request      | "1.0"                      |
| <a id="url" href="#url" title="#url">`url`</a> | Service server url                    | "http://example.com"       |

!!! info "`method` label value"

    If the HTTP method verb on a request is not one defined in the set of common methods for [`HTTP/1.1`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
    or the [`PRI`](https://datatracker.ietf.org/doc/html/rfc7540#section-11.6) verb (for `HTTP/2`),
    then the value for the method label becomes `EXTENSION_METHOD`.
