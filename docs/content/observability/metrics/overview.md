---
title: "Traefik Metrics Overview"
description: "Traefik Proxy supports these metrics backend systems: Datadog, InfluxDB 2.X, Prometheus, and StatsD. Read the full documentation to get started."
---

# Metrics

Traefik provides metrics in the [OpenTelemetry](./opentelemetry.md) format as well as the following vendor specific backends:

- [Datadog](./datadog.md)
- [InfluxDB2](./influxdb2.md)
- [Prometheus](./prometheus.md)
- [StatsD](./statsd.md)

Traefik Proxy hosts an official Grafana dashboard for both [on-premises](https://grafana.com/grafana/dashboards/17346)
and [Kubernetes](https://grafana.com/grafana/dashboards/17347) deployments.

## Common Options

### `addInternals`

_Optional, Default="false"_

Enables metrics for internal resources (e.g.: `ping@internals`).

```yaml tab="File (YAML)"
metrics:
  addInternals: true
```

```toml tab="File (TOML)"
[metrics]
addInternals = true
```

```bash tab="CLI"
--metrics.addinternals
```

## Global Metrics

| Metric                     | Type  | [Labels](#labels)        | Description                                                        |
|----------------------------|-------|--------------------------|--------------------------------------------------------------------|
| Config reload total        | Count |                          | The total count of configuration reloads.                          |
| Config reload last success | Gauge |                          | The timestamp of the last configuration reload success.            |
| Open connections           | Gauge | `entrypoint`, `protocol` | The current count of open connections, by entrypoint and protocol. |
| TLS certificates not after | Gauge |                          | The expiration date of certificates.                               |

```opentelemetry tab="OpenTelemetry"
traefik_config_reloads_total
traefik_config_last_reload_success
traefik_open_connections
traefik_tls_certs_not_after
```

```prom tab="Prometheus"
traefik_config_reloads_total
traefik_config_last_reload_success
traefik_open_connections
traefik_tls_certs_not_after
```

```dd tab="Datadog"
config.reload.total
config.reload.lastSuccessTimestamp
open.connections
tls.certs.notAfterTimestamp
```

```influxdb tab="InfluxDB2"
traefik.config.reload.total
traefik.config.reload.lastSuccessTimestamp
traefik.open.connections
traefik.tls.certs.notAfterTimestamp
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.config.reload.total
{prefix}.config.reload.lastSuccessTimestamp
{prefix}.open.connections
{prefix}.tls.certs.notAfterTimestamp
```

### Labels

Here is a comprehensive list of labels that are provided by the global metrics:

| Label        | Description                            | example              |
|--------------|----------------------------------------|----------------------|
| `entrypoint` | Entrypoint that handled the connection | "example_entrypoint" |
| `protocol`   | Connection protocol                    | "TCP"                |

## OpenTelemetry Semantic Conventions

Traefik Proxy follows [official OpenTelemetry semantic conventions v1.23.1](https://github.com/open-telemetry/semantic-conventions/blob/v1.23.1/docs/http/http-metrics.md).

### HTTP Server

| Metric                        | Type      | [Labels](#labels)                                                                                                                        | Description                       |
|-------------------------------|-----------|------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|
| http.server.request.duration	 | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP server requests  |

#### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label                       | Description                                                  | example       |
|-----------------------------|--------------------------------------------------------------|---------------|
| `error.type`                | Describes a class of error the operation ended with          | "500"         |
| `http.request.method`       | HTTP request method                                          | "GET"         |
| `http.response.status_code` | HTTP response status code                                    | "200"         |
| `network.protocol.name`     | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| `network.protocol.version`  | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| `server.address`            | Name of the local HTTP server that received the request      | "example.com" |
| `server.port`               | Port of the local HTTP server that received the request      | "80"          |
| `url.scheme`                | The URI scheme component identifying the used protocol       | "http"        |

### HTTP Client

| Metric                        | Type      | [Labels](#labels)                                                                                                                        | Description                       |
|-------------------------------|-----------|------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------|
| http.client.request.duration	 | Histogram | `error.type`, `http.request.method`, `http.response.status_code`, `network.protocol.name`, `server.address`, `server.port`, `url.scheme` | Duration of HTTP client requests  |

#### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label                       | Description                                                  | example       |
|-----------------------------|--------------------------------------------------------------|---------------|
| `error.type`                | Describes a class of error the operation ended with          | "500"         |
| `http.request.method`       | HTTP request method                                          | "GET"         |
| `http.response.status_code` | HTTP response status code                                   | "200"         |
| `network.protocol.name`     | OSI application layer or non-OSI equivalent                  | "http/1.1"    |
| `network.protocol.version`  | Version of the protocol specified in `network.protocol.name` | "1.1"         |
| `server.address`            | Name of the local HTTP server that received the request      | "example.com" |
| `server.port`               | Port of the local HTTP server that received the request      | "80"          |
| `url.scheme`                | The URI scheme component identifying the used protocol       | "http"        |

## HTTP Metrics

On top of the official OpenTelemetry semantic conventions, Traefik provides its own metrics to monitor the incoming traffic.

### EntryPoint Metrics

| Metric                | Type      | [Labels](#labels)                          | Description                                                         |
|-----------------------|-----------|--------------------------------------------|---------------------------------------------------------------------|
| Requests total        | Count     | `code`, `method`, `protocol`, `entrypoint` | The total count of HTTP requests received by an entrypoint.         |
| Requests TLS total    | Count     | `tls_version`, `tls_cipher`, `entrypoint`  | The total count of HTTPS requests received by an entrypoint.        |
| Request duration      | Histogram | `code`, `method`, `protocol`, `entrypoint` | Request processing duration histogram on an entrypoint.             |
| Requests bytes total  | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP requests in bytes handled by an entrypoint.  |
| Responses bytes total | Count     | `code`, `method`, `protocol`, `entrypoint` | The total size of HTTP responses in bytes handled by an entrypoint. |

```opentelemetry tab="OpenTelemetry"
traefik_entrypoint_requests_total
traefik_entrypoint_requests_tls_total
traefik_entrypoint_request_duration_seconds
traefik_entrypoint_requests_bytes_total
traefik_entrypoint_responses_bytes_total
```

```prom tab="Prometheus"
traefik_entrypoint_requests_total
traefik_entrypoint_requests_tls_total
traefik_entrypoint_request_duration_seconds
traefik_entrypoint_requests_bytes_total
traefik_entrypoint_responses_bytes_total
```

```dd tab="Datadog"
entrypoint.request.total
entrypoint.request.tls.total
entrypoint.request.duration
entrypoint.requests.bytes.total
entrypoint.responses.bytes.total
```

```influxdb tab="InfluxDB2"
traefik.entrypoint.requests.total
traefik.entrypoint.requests.tls.total
traefik.entrypoint.request.duration
traefik.entrypoint.requests.bytes.total
traefik.entrypoint.responses.bytes.total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.entrypoint.request.total
{prefix}.entrypoint.request.tls.total
{prefix}.entrypoint.request.duration
{prefix}.entrypoint.requests.bytes.total
{prefix}.entrypoint.responses.bytes.total
```

### Router Metrics

| Metric                | Type      | [Labels](#labels)                                 | Description                                                    |
|-----------------------|-----------|---------------------------------------------------|----------------------------------------------------------------|
| Requests total        | Count     | `code`, `method`, `protocol`, `router`, `service` | The total count of HTTP requests handled by a router.          |
| Requests TLS total    | Count     | `tls_version`, `tls_cipher`, `router`, `service`  | The total count of HTTPS requests handled by a router.         |
| Request duration      | Histogram | `code`, `method`, `protocol`, `router`, `service` | Request processing duration histogram on a router.             |
| Requests bytes total  | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP requests in bytes handled by a router.  |
| Responses bytes total | Count     | `code`, `method`, `protocol`, `router`, `service` | The total size of HTTP responses in bytes handled by a router. |

```opentelemetry tab="OpenTelemetry"
traefik_router_requests_total
traefik_router_requests_tls_total
traefik_router_request_duration_seconds
traefik_router_requests_bytes_total
traefik_router_responses_bytes_total
```

```prom tab="Prometheus"
traefik_router_requests_total
traefik_router_requests_tls_total
traefik_router_request_duration_seconds
traefik_router_requests_bytes_total
traefik_router_responses_bytes_total
```

```dd tab="Datadog"
router.request.total
router.request.tls.total
router.request.duration
router.requests.bytes.total
router.responses.bytes.total
```

```influxdb tab="InfluxDB2"
traefik.router.requests.total
traefik.router.requests.tls.total
traefik.router.request.duration
traefik.router.requests.bytes.total
traefik.router.responses.bytes.total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.router.request.total
{prefix}.router.request.tls.total
{prefix}.router.request.duration
{prefix}.router.requests.bytes.total
{prefix}.router.responses.bytes.total
```

### Service Metrics

| Metric                | Type      | Labels                                  | Description                                                 |
|-----------------------|-----------|-----------------------------------------|-------------------------------------------------------------|
| Requests total        | Count     | `code`, `method`, `protocol`, `service` | The total count of HTTP requests processed on a service.    |
| Requests TLS total    | Count     | `tls_version`, `tls_cipher`, `service`  | The total count of HTTPS requests processed on a service.   |
| Request duration      | Histogram | `code`, `method`, `protocol`, `service` | Request processing duration histogram on a service.         |
| Retries total         | Count     | `service`                               | The count of requests retries on a service.                 |
| Server UP             | Gauge     | `service`, `url`                        | Current service's server status, 0 for a down or 1 for up.  |
| Requests bytes total  | Count     | `code`, `method`, `protocol`, `service` | The total size of requests in bytes received by a service.  |
| Responses bytes total | Count     | `code`, `method`, `protocol`, `service` | The total size of responses in bytes returned by a service. |

```opentelemetry tab="OpenTelemetry"
traefik_service_requests_total
traefik_service_requests_tls_total
traefik_service_request_duration_seconds
traefik_service_retries_total
traefik_service_server_up
traefik_service_requests_bytes_total
traefik_service_responses_bytes_total
```

```prom tab="Prometheus"
traefik_service_requests_total
traefik_service_requests_tls_total
traefik_service_request_duration_seconds
traefik_service_retries_total
traefik_service_server_up
traefik_service_requests_bytes_total
traefik_service_responses_bytes_total
```

```dd tab="Datadog"
service.request.total
router.service.tls.total
service.request.duration
service.retries.total
service.server.up
service.requests.bytes.total
service.responses.bytes.total
```

```influxdb tab="InfluxDB2"
traefik.service.requests.total
traefik.service.requests.tls.total
traefik.service.request.duration
traefik.service.retries.total
traefik.service.server.up
traefik.service.requests.bytes.total
traefik.service.responses.bytes.total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.request.total
{prefix}.service.request.tls.total
{prefix}.service.request.duration
{prefix}.service.retries.total
{prefix}.service.server.up
{prefix}.service.requests.bytes.total
{prefix}.service.responses.bytes.total
```

### Labels

Here is a comprehensive list of labels that are provided by the metrics:

| Label         | Description                           | example                    |
|---------------|---------------------------------------|----------------------------|
| `cn`          | Certificate Common Name               | "example.com"              |
| `code`        | Request code                          | "200"                      |
| `entrypoint`  | Entrypoint that handled the request   | "example_entrypoint"       |
| `method`      | Request Method                        | "GET"                      |
| `protocol`    | Request protocol                      | "http"                     |
| `router`      | Router that handled the request       | "example_router"           |
| `sans`        | Certificate Subject Alternative NameS | "example.com"              |
| `serial`      | Certificate Serial Number             | "123..."                   |
| `service`     | Service that handled the request      | "example_service@provider" |
| `tls_cipher`  | TLS cipher used for the request       | "TLS_FALLBACK_SCSV"        |
| `tls_version` | TLS version used for the request      | "1.0"                      |
| `url`         | Service server url                    | "http://example.com"       |

!!! info "`method` label value"

    If the HTTP method verb on a request is not one defined in the set of common methods for [`HTTP/1.1`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
    or the [`PRI`](https://datatracker.ietf.org/doc/html/rfc7540#section-11.6) verb (for `HTTP/2`),
    then the value for the method label becomes `EXTENSION_METHOD`.
