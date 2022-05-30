---
title: "Traefik Metrics Overview"
description: "Traefik Proxy supports four metrics backend systems: Datadog, InfluxDB, Prometheus, and StatsD. Read the full documentation to get started."
---

# Metrics

Traefik supports 4 metrics backends:

- [Datadog](./datadog.md)
- [InfluxDB](./influxdb.md)
- [InfluxDB2](./influxdb2.md)
- [Prometheus](./prometheus.md)
- [StatsD](./statsd.md)

## Global Metrics

| Metric                                                                  | DataDog | InfluxDB / InfluxDB2 | Prometheus | StatsD |
|-------------------------------------------------------------------------|---------|----------------------|------------|--------|
| [Configuration reloads](#configuration-reloads)                         | ✓       | ✓                    | ✓          | ✓      |
| [Last Configuration Reload Success](#last-configuration-reload-success) | ✓       | ✓                    | ✓          | ✓      |
| [TLS certificates expiration](#tls-certificates-expiration)             | ✓       | ✓                    | ✓          | ✓      |

### Configuration Reloads

The total count of configuration reloads.

```dd tab="Datadog"
config.reload.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.config.reload.total
```

```prom tab="Prometheus"
traefik_config_reloads_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.config.reload.total
```

### Last Configuration Reload Success

The timestamp of the last configuration reload success.

```dd tab="Datadog"
config.reload.lastSuccessTimestamp
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.config.reload.lastSuccessTimestamp
```

```prom tab="Prometheus"
traefik_config_last_reload_success
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.config.reload.lastSuccessTimestamp
```

### TLS certificates expiration

The expiration date of certificates.

[Labels](#labels): `cn`, `sans`, `serial`.

```dd tab="Datadog"
tls.certs.notAfterTimestamp
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.tls.certs.notAfterTimestamp
```

```prom tab="Prometheus"
traefik_tls_certs_not_after
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.tls.certs.notAfterTimestamp
```

## EntryPoint Metrics

| Metric                                                    | DataDog | InfluxDB / InfluxDB2 | Prometheus | StatsD |
|-----------------------------------------------------------|---------|----------------------|------------|--------|
| [HTTP Requests Count](#http-requests-count)               | ✓       | ✓                    | ✓          | ✓      |
| [HTTPS Requests Count](#https-requests-count)             | ✓       | ✓                    | ✓          | ✓      |
| [Request Duration Histogram](#request-duration-histogram) | ✓       | ✓                    | ✓          | ✓      |
| [Open Connections Count](#open-connections-count)         | ✓       | ✓                    | ✓          | ✓      |

### HTTP Requests Count

The total count of HTTP requests received by an entrypoint.

[Labels](#labels): `code`, `method`, `protocol`, `entrypoint`.

```dd tab="Datadog"
entrypoint.request.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.entrypoint.requests.total
```

```prom tab="Prometheus"
traefik_entrypoint_requests_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.entrypoint.request.total
```

### HTTPS Requests Count

The total count of HTTPS requests received by an entrypoint.

[Labels](#labels): `tls_version`, `tls_cipher`, `entrypoint`.

```dd tab="Datadog"
entrypoint.request.tls.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.entrypoint.requests.tls.total
```

```prom tab="Prometheus"
traefik_entrypoint_requests_tls_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.entrypoint.request.tls.total
```

### Request Duration Histogram

Request processing duration histogram on an entrypoint.

[Labels](#labels): `code`, `method`, `protocol`, `entrypoint`.

```dd tab="Datadog"
entrypoint.request.duration
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.entrypoint.request.duration
```

```prom tab="Prometheus"
traefik_entrypoint_request_duration_seconds
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.entrypoint.request.duration
```

### Open Connections Count

The current count of open connections on an entrypoint.

[Labels](#labels): `method`, `protocol`, `entrypoint`.

```dd tab="Datadog"
entrypoint.connections.open
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.entrypoint.connections.open
```

```prom tab="Prometheus"
traefik_entrypoint_open_connections
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.entrypoint.connections.open
```

## Router Metrics

| Metric                                                      | DataDog | InfluxDB / InfluxDB2 | Prometheus | StatsD |
|-------------------------------------------------------------|---------|----------------------|------------|--------|
| [HTTP Requests Count](#http-requests-count_1)               | ✓       | ✓                    | ✓          | ✓      |
| [HTTPS Requests Count](#https-requests-count_1)             | ✓       | ✓                    | ✓          | ✓      |
| [Request Duration Histogram](#request-duration-histogram_1) | ✓       | ✓                    | ✓          | ✓      |
| [Open Connections Count](#open-connections-count_1)         | ✓       | ✓                    | ✓          | ✓      |

### HTTP Requests Count

The total count of HTTP requests handled by a router.

[Labels](#labels): `code`, `method`, `protocol`, `router`, `service`.

```dd tab="Datadog"
router.request.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.router.requests.total
```

```prom tab="Prometheus"
traefik_router_requests_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.router.request.total
```

### HTTPS Requests Count

The total count of HTTPS requests handled by a router.

[Labels](#labels): `tls_version`, `tls_cipher`, `router`, `service`.

```dd tab="Datadog"
router.request.tls.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.router.requests.tls.total
```

```prom tab="Prometheus"
traefik_router_requests_tls_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.router.request.tls.total
```

### Request Duration Histogram

Request processing duration histogram on a router.

[Labels](#labels): `code`, `method`, `protocol`, `router`, `service`.

```dd tab="Datadog"
router.request.duration
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.router.request.duration
```

```prom tab="Prometheus"
traefik_router_request_duration_seconds
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.router.request.duration
```

### Open Connections Count

The current count of open connections on a router.

[Labels](#labels): `method`, `protocol`, `router`, `service`.

```dd tab="Datadog"
router.connections.open
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.router.connections.open
```

```prom tab="Prometheus"
traefik_router_open_connections
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.router.connections.open
```

## Service Metrics

| Metric                                                      | DataDog | InfluxDB / InfluxDB2 | Prometheus | StatsD |
|-------------------------------------------------------------|---------|----------------------|------------|--------|
| [HTTP Requests Count](#http-requests-count_2)               | ✓       | ✓                    | ✓          | ✓      |
| [HTTPS Requests Count](#https-requests-count_2)             | ✓       | ✓                    | ✓          | ✓      |
| [Request Duration Histogram](#request-duration-histogram_2) | ✓       | ✓                    | ✓          | ✓      |
| [Open Connections Count](#open-connections-count_2)         | ✓       | ✓                    | ✓          | ✓      |
| [Requests Retries Count](#requests-retries-count)           | ✓       | ✓                    | ✓          | ✓      |
| [Service Server UP](#service-server-up)                     | ✓       | ✓                    | ✓          | ✓      |

### HTTP Requests Count

The total count of HTTP requests processed on a service.

[Labels](#labels): `code`, `method`, `protocol`, `service`.

```dd tab="Datadog"
service.request.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.service.requests.total
```

```prom tab="Prometheus"
traefik_service_requests_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.request.total
```

### HTTPS Requests Count

The total count of HTTPS requests processed on a service.

[Labels](#labels): `tls_version`, `tls_cipher`, `service`.

```dd tab="Datadog"
router.service.tls.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.service.requests.tls.total
```

```prom tab="Prometheus"
traefik_service_requests_tls_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.request.tls.total
```

### Request Duration Histogram

Request processing duration histogram on a service.

[Labels](#labels): `code`, `method`, `protocol`, `service`.

```dd tab="Datadog"
service.request.duration
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.service.request.duration
```

```prom tab="Prometheus"
traefik_service_request_duration_seconds
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.request.duration
```

### Open Connections Count

The current count of open connections on a service.

[Labels](#labels): `method`, `protocol`, `service`.

```dd tab="Datadog"
service.connections.open
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.service.connections.open
```

```prom tab="Prometheus"
traefik_service_open_connections
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.connections.open
```

### Requests Retries Count

The count of requests retries on a service.

[Labels](#labels): `service`.

```dd tab="Datadog"
service.retries.total
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.service.retries.total
```

```prom tab="Prometheus"
traefik_service_retries_total
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.retries.total
```

### Service Server UP

Current service's server status, described by a gauge with a value of 0 for a down server or a value of 1 for an up server.

[Labels](#labels): `service`, `url`.

```dd tab="Datadog"
service.server.up
```

```influxdb tab="InfluxDB / InfluxDB2"
traefik.service.server.up
```

```prom tab="Prometheus"
traefik_service_server_up
```

```statsd tab="StatsD"
# Default prefix: "traefik"
{prefix}.service.server.up
```

## Labels

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
