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
| `metrics.addInternals`                     | Enables metrics for internal resources (e.g.: `ping@internal`).                                                                                                  | false                                              | No       |
| `metrics.otlp.serviceName`                 | Defines the service name resource attribute.                                                                                                                     | "traefik"                                          | No       |
| `metrics.otlp.resourceAttributes`          | Defines additional resource attributes to be sent to the collector.                                                                                              | []                                                 | No       |
| `metrics.otlp.addEntryPointsLabels`        | Enable metrics on entry points.                                                                                                                                  | true                                               | No       |
| `metrics.otlp.addRoutersLabels`            | Enable metrics on routers.                                                                                                                                       | false                                              | No       |
| `metrics.otlp.addServicesLabels`           | Enable metrics on services.                                                                                                                                      | true                                               | No       |
| `metrics.otlp.explicitBoundaries`          | Explicit boundaries for Histogram data points.                                                                                                                   | ".005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10" | No       |
| `metrics.otlp.pushInterval`                | Interval at which metrics are sent to the OpenTelemetry Collector.                                                                                               | 10s                                                | No       |
| `metrics.otlp.http`                        | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| `metrics.otlp.http.endpoint`               | URL of the OpenTelemetry Collector to send metrics to.<br /> Format="`<scheme>://<host>:<port><path>`"                                                           | "http://localhost:4318/v1/metrics"                 | Yes      |
| `metrics.otlp.http.headers`                | Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.                                                                             | -                                                  | No       |
| `metrics.otlp.http.tls.ca`                 | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector,<br />it defaults to the system bundle.                          | ""                                                 | No       |
| `metrics.otlp.http.tls.cert`               | Path to the public certificate used for the secure connection to the OpenTelemetry Collector.<br />When using this option, setting the `key` option is required. | ""                                                 | No       |
| `metrics.otlp.http.tls.key`                | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| `metrics.otlp.http.tls.insecureskipverify` | Allow the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.                   | false                                              | Yes      |
| `metrics.otlp.grpc`                        | This instructs the exporter to send metrics to the OpenTelemetry Collector using gRPC.                                                                           | null/false                                         | No       |
| `metrics.otlp.grpc.endpoint`               | Address of the OpenTelemetry Collector to send metrics to.<br /> Format="`<host>:<port>`"                                                                        | "localhost:4317"                                   | Yes      |
| `metrics.otlp.grpc.headers`                | Additional headers sent with metrics by the exporter to the OpenTelemetry Collector.                                                                             | -                                                  | No       |
| `metrics.otlp.http.grpc.insecure`          | Allows exporter to send metrics to the OpenTelemetry Collector without using a secured protocol.                                                                 | false                                              | Yes      |
| `metrics.otlp.grpc.tls.ca`                 | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector,<br />it defaults to the system bundle.                          | -                                                  | No       |
| `metrics.otlp.grpc.tls.cert`               | Path to the public certificate used for the secure connection to the OpenTelemetry Collector.<br />When using this option, setting the `key` option is required. | -                                                  | No       |
| `metrics.otlp.grpc.tls.key`                | This instructs the exporter to send the metrics to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.              | null/false                                         | No       |
| `metrics.otlp.grpc.tls.insecureskipverify` | Allow the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.                   | false                                              | Yes      |

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
| `metrics.addInternals` | Enables metrics for internal resources (e.g.: `ping@internal`). | false      | No      |
| `datadog.address` | Defines the address for the exporter to send metrics to datadog-agent. More information [here](#address)|  `127.0.0.1:8125`     | Yes   |
| `datadog.addEntryPointsLabels` | Enable metrics on entry points. |  true   | No   |
| `datadog.addRoutersLabels` | Enable metrics on routers. |  false   | No   |
| `datadog.addServicesLabels` | Enable metrics on services. |  true   | No   |
| `datadog.pushInterval` | Defines the interval used by the exporter to push metrics to datadog-agent. |  10s   | No   |
| `datadog.prefix` | Defines the prefix to use for metrics collection. |  "traefik"   | No   |

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
| `metrics.addInternal` | Enables metrics for internal resources (e.g.: `ping@internal`). | false      | No      |
| `metrics.influxDB2.addEntryPointsLabels` | Enable metrics on entry points. | true      | No      |
| `metrics.influxDB2.addRoutersLabels` | Enable metrics on routers. | false      | No      |
| `metrics.influxDB2.addServicesLabels` | Enable metrics on services.| true      | No      |
| `metrics.influxDB2.additionalLabels` | Additional labels (InfluxDB tags) on all metrics. | - | No      |
| `metrics.influxDB2.pushInterval` | The interval used by the exporter to push metrics to InfluxDB server. | 10s      | No      |
| `metrics.influxDB2.address` | Address of the InfluxDB v2 instance. | "http://localhost:8086"     | Yes      |
| `metrics.influxDB2.token` | Token with which to connect to InfluxDB v2. | - | Yes      |
| `metrics.influxDB2.org` | Organisation where metrics will be stored. | -  | Yes      |
| `metrics.influxDB2.bucket` | Bucket where metrics will be stored. | -  | Yes      |

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
| `metrics.prometheus.addInternals` | Enables metrics for internal resources (e.g.: `ping@internals`). | false      | No      |
| `metrics.prometheus.addEntryPointsLabels` | Enable metrics on entry points. | true      | No      |
| `metrics.prometheus.addRoutersLabels` | Enable metrics on routers. | false      | No      |
| `metrics.prometheus.addServicesLabels` | Enable metrics on services.| true      | No      |
| `metrics.prometheus.buckets` | Buckets for latency metrics. |"0.100000, 0.300000, 1.200000, 5.000000"  | No      |
| `metrics.prometheus.manualRouting` | Set to _true_, it disables the default internal router in order to allow creating a custom router for the `prometheus@internal` service. | false    | No      |
| `metrics.prometheus.entryPoint` | Traefik Entrypoint name used to expose metrics. | "traefik"     | No      |
| `metrics.prometheus.headerLabels` | Defines extra labels extracted from request headers for the `requests_total` metrics.<br />More information [here](#headerlabels). |       | Yes      |

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
| `metrics.addInternals` | Enables metrics for internal resources (e.g.: `ping@internals`). | false      | No      |
| `metrics.statsD.addEntryPointsLabels` | Enable metrics on entry points. | true      | No      |
| `metrics.statsD.addRoutersLabels` | Enable metrics on routers. | false      | No      |
| `metrics.statsD.addServicesLabels` | Enable metrics on services.| true      | No      |
| `metrics.statsD.pushInterval` | The interval used by the exporter to push metrics to DataDog server. | 10s      | No      |
| `metrics.statsD.address` | Address instructs exporter to send metrics to statsd at this address.  | "127.0.0.1:8125"     | Yes      |
| `metrics.statsD.prefix` | The prefix to use for metrics collection. | "traefik"      | No      |

## Metrics Provided

### Global Metrics

=== "OpenTelemetry"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | `traefik_config_reloads_total`        | Count |                          | The total count of configuration reloads.                          |
    | `traefik_config_last_reload_success` | Gauge |                          | The timestamp of the last configuration reload success.            |
    | `traefik_open_connections`           | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | `traefik_tls_certs_not_after` | Gauge |                          | The expiration date of certificates.                               |
    
=== "Prometheus"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | `traefik_config_reloads_total`        | Count |                          | The total count of configuration reloads.                          |
    | `traefik_config_last_reload_success` | Gauge |                          | The timestamp of the last configuration reload success.            |
    | `traefik_open_connections`           | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | `traefik_tls_certs_not_after` | Gauge |      | The expiration date of certificates. |

=== "Datadog"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | `config.reload.total`        | Count |                          | The total count of configuration reloads.                          |
    | `config.reload.lastSuccessTimestamp` | Gauge |                          | The timestamp of the last configuration reload success.            |
    | `open.connections`           | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | `tls.certs.notAfterTimestamp` | Gauge |                          | The expiration date of certificates.                               |

=== "InfluxDB2"
    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | `traefik.config.reload.total`        | Count |                          | The total count of configuration reloads.                          |
    | `traefik.config.reload.lastSuccessTimestamp` | Gauge |                          | The timestamp of the last configuration reload success.            |
    | `traefik.open.connections`           | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | `traefik.tls.certs.notAfterTimestamp` | Gauge |                          | The expiration date of certificates.                               |

=== "StatsD"
    | Metric       | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | `{prefix}.config.reload.total`    | Count |     | The total count of configuration reloads. |
    | `{prefix}.config.reload.lastSuccessTimestamp` | Gauge |          | The timestamp of the last configuration reload success.            |
    | `{prefix}.open.connections`    | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
    | `{prefix}.tls.certs.notAfterTimestamp` | Gauge |    | The expiration date of certificates.   |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Labels

Here is a comprehensive list of labels that are provided by the global metrics:

| Label        | Description      | example              |
|--------------|----------------------------------------|----------------------|
| `entrypoint` | Entrypoint that handled the connection | "example_entrypoint" |
| `protocol`   | Connection protocol     | "TCP"      |

### OpenTelemetry Semantic Conventions

Traefik Proxy follows [official OpenTelemetry semantic conventions v1.23.1](https://github.com/open-telemetry/semantic-conventions/blob/v1.23.1/docs/http/http-metrics.md).

#### HTTP Server

| Metric     | Type      | [Labels](#labels)       | Description   |
|----------|-----------|-------------------------|------------------|
| `http.server.request.duration`	 | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP server requests  |

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label     | Description   | example       |
|-----------------------------|--------|---------------|
| `error.type`                | Describes a class of error the operation ended with          | "500"         |
| `http.request.method`       | HTTP request method                                          | "GET"         |
| `http.response.status_code` | HTTP response status code                                    | "200"         |
| `network.protocol.name`     | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| `network.protocol.version`  | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| `server.address`            | Name of the local HTTP server that received the request      | "example.com" |
| `server.port`               | Port of the local HTTP server that received the request      | "80"          |
| `url.scheme`                | The URI scheme component identifying the used protocol       | "http"        |

#### HTTP Client

| Metric    | Type      | [Labels](#labels)    | Description  |
|-------------------------------|-----------|-----------------|--------|
| `http.client.request.duration`	 | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP client requests  |

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label  | Description     | example       |
|------  -----|------------|---------------|
| `error.type`  | Describes a class of error the operation ended with    | "500"   |
| `http.request.method`       | HTTP request method  | "GET" |
| `http.response.status_code` | HTTP response status code  | "200" |
| `network.protocol.name`     | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| `network.protocol.version`  | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| `server.address`            | Name of the local HTTP server that received the request      | "example.com" |
| `server.port`               | Port of the local HTTP server that received the request      | "80"          |
| `url.scheme`                | The URI scheme component identifying the used protocol       | "http"        |

### HTTP Metrics

On top of the official OpenTelemetry semantic conventions, Traefik provides its own metrics to monitor the incoming traffic.

#### EntryPoint Metrics

=== "OpenTelemetry"

    | Metric   | Type      | [Labels](#labels)         | Description   |
    |-----------------------|-----------|--------------------|--------------------------|
    | `traefik_entrypoint_requests_total`        | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | `traefik_entrypoint_requests_tls_total`    | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | `traefik_entrypoint_request_duration_seconds`     | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | `traefik_entrypoint_requests_bytes_total`  | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | `traefik_entrypoint_responses_bytes_total` | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |
    
=== "Prometheus"

    | Metric     | Type      | [Labels](#labels)      | Description      |
    |-----------------------|-----------|------------------------|-------------------------|
    | `traefik_entrypoint_requests_total`        | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | `traefik_entrypoint_requests_tls_total`    | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | `traefik_entrypoint_request_duration_seconds`     | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | `traefik_entrypoint_requests_bytes_total`  | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | `traefik_entrypoint_responses_bytes_total` | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "Datadog"

    | Metric   | Type      | [Labels](#labels)     | Description     |
    |-----------------------|-----------|------------------|---------------------------|
    | `entrypoint.requests.total`        | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | `entrypoint.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | `entrypoint.request.duration.seconds`     | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | `entrypoint.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | `entrypoint.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "InfluxDB2"

    | Metric    | Type      | [Labels](#labels)   | Description     |
    |------------|-----------|-------------------|-----------------|
    | `traefik.entrypoint.requests.total`        | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | `traefik.entrypoint.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | `traefik.entrypoint.request.duration.seconds`     | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | `traefik.entrypoint.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | `traefik.entrypoint.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

=== "StatsD"

    | Metric                     | Type  | [Labels](#labels)        | Description                                                        |
    |----------------------------|-------|--------------------------|--------------------------------------------------------------------|
    | `{prefix}.entrypoint.requests.total`        | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
    | `{prefix}.entrypoint.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
    | `{prefix}.entrypoint.request.duration.seconds`     | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
    | `{prefix}.entrypoint.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
    | `{prefix}.entrypoint.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Router Metrics

=== "OpenTelemetry"

    | Metric    | Type      | [Labels](#labels)         | Description           |
    |-----------------------|-----------|----------------------|--------------------------------|
    | `traefik_router_requests_total`        | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | `traefik_router_requests_tls_total`    | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | `traefik_router_request_duration_seconds`      | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | `traefik_router_requests_bytes_total`  | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | `traefik_router_responses_bytes_total` | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |
    
=== "Prometheus"

    | Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
    |-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
    | `traefik_router_requests_total`        | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | `traefik_router_requests_tls_total`    | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | `traefik_router_request_duration_seconds`      | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | `traefik_router_requests_bytes_total`  | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | `traefik_router_responses_bytes_total` | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "Datadog"

    | Metric    | Type      | [Labels](#labels)   | Description   |
    |-------------|-----------|---------------|---------------------|
    | `router.requests.total`        | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | `router.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | `router.request.duration.seconds`      | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | `router.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | `router.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "InfluxDB2"

    | Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
    |-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
    | `traefik.router.requests.total`        | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | `traefik.router.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | `traefik.router.request.duration.seconds`      | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | `traefik.router.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | `traefik.router.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

=== "StatsD"

    | Metric     | Type      | [Labels](#labels)      | Description   |
    |-----------------------|-----------|---------------|-------------|
    | `{prefix}.router.requests.total`        | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
    | `{prefix}.router.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
    | `{prefix}.router.request.duration.seconds`      | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
    | `{prefix}.router.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
    | `{prefix}.router.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

#### Service Metrics

=== "OpenTelemetry"

    | Metric    | Type      | Labels      | Description     |
    |-----------------------|-----------|------------|------------|
    | `traefik_service_requests_total`        | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | `traefik_service_requests_tls_total`    | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | `traefik_service_request_duration_seconds`      | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | `traefik_service_retries_total`         | Count     | `service`                               | The count of requests retries on a service.                 |
    | `traefik_service_server_up`             | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | `traefik_service_requests_bytes_total`  | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | `traefik_service_responses_bytes_total` | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |
    
=== "Prometheus"

    | Metric    | Type      | Labels    | Description    |
    |-----------------------|-----------|-------|------------|
    | `traefik_service_requests_total`        | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | `traefik_service_requests_tls_total`    | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | `traefik_service_request_duration_seconds`      | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | `traefik_service_retries_total`         | Count     | `service`                               | The count of requests retries on a service.                 |
    | `traefik_service_server_up`             | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | `traefik_service_requests_bytes_total`  | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | `traefik_service_responses_bytes_total` | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "Datadog"

    | Metric    | Type      | Labels    | Description |
    |-----------------------|-----------|--------|------------------|
    | `service.requests.total`        | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | `router.service.tls.total`    | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | `service.request.duration.seconds`      | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | `service.retries.total`         | Count     | `service`                               | The count of requests retries on a service.                 |
    | `service.server.up`             | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | `service.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | `service.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "InfluxDB2"

    | Metric                | Type      | Labels                                  | Description                                                 |
    |-----------------------|-----------|-----------------------------------------|-------------------------------------------------------------|
    | `traefik.service.requests.total`        | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | `traefik.service.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | `traefik.service.request.duration.seconds`      | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | `traefik.service.retries.total`         | Count     | `service`                               | The count of requests retries on a service.                 |
    | `traefik.service.server.up`             | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | `traefik.service.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | `traefik.service.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

=== "StatsD"

    | Metric                | Type      | Labels   | Description    |
    |-----------------------|-----------|-----|---------|
    | `{prefix}.service.requests.total`        | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
    | `{prefix}.service.requests.tls.total`    | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
    | `{prefix}.service.request.duration.seconds`      | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
    | `{prefix}.service.retries.total`         | Count     | `service`                               | The count of requests retries on a service.                 |
    | `{prefix}.service.server.up`             | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
    | `{prefix}.service.requests.bytes.total`  | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
    | `{prefix}.service.responses.bytes.total` | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

!!! note "\{prefix\} Default Value"
        By default, \{prefix\} value is `traefik`.

##### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label         | Description      | example      |
|---------------|-------------------|----------------------------|
| `cn`          | Certificate Common Name     | "example.com"     |
| `code`        | Request code       | "200"                      |
| `entrypoint`  | Entrypoint that handled the request   | "example_entrypoint"       |
| `method`      | Request Method     | "GET"    |
| `protocol`    | Request protocol      | "http"                     |
| `router`      | Router that handled the request       | "example_router"    |
| `sans`        | Certificate Subject Alternative NameS | "example.com"              |
| `serial`      | Certificate Serial Number   | "123..."                   |
| `service`     | Service that handled the request      | "example_service@provider" |
| `tls_cipher`  | TLS cipher used for the request       | "TLS_FALLBACK_SCSV"        |
| `tls_version` | TLS version used for the request      | "1.0"                      |
| `url`         | Service server url                    | "http://example.com"       |

!!! info "`method` label value"

    If the HTTP method verb on a request is not one defined in the set of common methods for [`HTTP/1.1`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
    or the [`PRI`](https://datatracker.ietf.org/doc/html/rfc7540#section-11.6) verb (for `HTTP/2`),
    then the value for the method label becomes `EXTENSION_METHOD`.
