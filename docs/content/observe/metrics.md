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

```yaml tab="Helm Values"
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

When the `observability` options are not defined on a router, it inherits the behavior from the entrypoint's observability configuration, or the global one.

## Available Metrics

Traefik Proxy provides a comprehensive set of metrics to monitor its health and the traffic it manages. These include internal metrics as well as standard metrics following the OpenTelemetry semantic conventions.

### Global Metrics

- **Configuration Reloads:** Total count of configuration reloads and the timestamp of the last successful reload (`traefik_config_reloads_total`, `traefik_config_last_reload_success`).
- **Open Connections:** The current number of open connections (`traefik_open_connections`).
- **TLS Certificate Expiry:** The expiration date of TLS certificates (`traefik_tls_certs_not_after`).

### EntryPoint Metrics

- **Requests Total:** Total count of requests (`traefik_entrypoint_requests_total`).
- **TLS Requests Total:** Total count of TLS requests (`traefik_entrypoint_requests_tls_total`).
- **Request Duration:** A histogram of the request durations (`traefik_entrypoint_request_duration_seconds`).
- **Request Size:** Total size in bytes of incoming requests (`traefik_entrypoint_requests_bytes_total`).
- **Response Size:** Total size in bytes of outgoing responses (`traefik_entrypoint_responses_bytes_total`).

### Router Metrics

- **Requests Total:** Total count of requests (`traefik_router_requests_total`).
- **Request Duration:** A histogram of the request durations (`traefik_router_request_duration_seconds`).

### Service Metrics

- **Requests Total:** Total count of requests (`traefik_service_requests_total`).
- **Request Duration:** A histogram of the request durations (`traefik_service_request_duration_seconds`).
- **Request Retries:** Total count of request retries (`traefik_service_retries_total`).
- **Server Health:** The health status of the backend servers, where 1 is up and 0 is down (`traefik_service_server_up`).

## Metric Labels

Metrics are enriched with labels that allow for detailed filtering and aggregation. Common labels include:

- **entrypoint:** The name of the entrypoint that received the request.
- **router:** The name of the router that handled the request.
- **service:** The name of the service that the request was forwarded to.
- **method:** The HTTP request method (e.g., `GET`, `POST`).
- **protocol:** The request protocol (e.g., `http`, `https`).
- **code:** The HTTP status code of the response (e.g., `200`, `404`).
- **tls_version:** The TLS version used for the connection.
- **tls_cipher:** The TLS cipher suite used for the connection.
- **url:** The URL of the backend server.

For detailed configuration options, refer to the [reference documentation](../reference/install-configuration/observability/metrics.md).
