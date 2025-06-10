# Logs and Access Logs

Logs and Access Logs in Traefik Proxy provide real-time insight into the health of your system. They enable swift error detection and intervention through alerts. By centralizing logs, you can streamline the debugging process during incident resolution.

## Configuration

To enable and configure access logs in Traefik Proxy, you can use the static configuration file or Helm values if you are using the [Helm chart](https://github.com/traefik/traefik-helm-chart).

The following example enables access logs in JSON format, filters them to only include specific status codes, and customizes the fields that are kept or dropped.

```yaml tab="Structured (YAML)"
accessLog:
  format: json
  filters:
    statusCodes:
      - "200"
      - "400-404"
      - "500-503"
  fields:
    names:
      ClientUsername: drop
    headers:
      defaultMode: keep
      names:
        User-Agent: redact
        Content-Type: keep
```

```yaml tab="Helm Values"
# values.yaml
logs:
  access:
    enabled: true
    format: json
    filters:
      statusCodes:
        - "200"
        - "400-404"
        - "500-503"
    fields:
      general:
        names:
          ClientUsername: drop
      headers:
        names:
          User-Agent: redact
          Content-Type: keep
```

## Per-Router Access Logs

You can enable or disable access logs for a specific router. This is useful for turning off logging for noisy routes while keeping it on globally.

Here's an example of disabling access logs on a specific router:

```yaml tab="Structured (YAML)"
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: my-service
      observability:
        accessLogs: false
```

When the `observability` options are not defined on a router, it inherits the behavior from the entrypoint's observability configuration, or the global one.

## Log Formats

Traefik Proxy supports the following log formats:

- Common Log Format (CLF)
- JSON

## Access Log Filters

You can configure Traefik Proxy to only record access logs for requests that match certain criteria. This is useful for reducing the volume of logs and focusing on specific events.

The available filters are:

- **Status Codes:** Keep logs only for requests with specific HTTP status codes or ranges (e.g., `200`, `400-404`).
- **Retry Attempts:** Keep logs only when a request retry has occurred.
- **Minimum Duration:** Keep logs only for requests that take longer than a specified duration.

## Log Fields Customization

When using the `json` format, you can customize which fields are included in your access logs.

- **Request Fields:** You can choose to `keep`, `drop`, or `redact` any of the standard request fields. A complete list of available fields like `ClientHost`, `RequestMethod`, and `Duration` can be found in the [reference documentation](../reference/install-configuration/observability/logs-and-accesslogs.md#available-fields).
- **Request Headers:** You can also specify which request headers should be included in the logs, and whether their values should be `kept`, `dropped`, or `redacted`.

For detailed configuration options, refer to the [reference documentation](../reference/install-configuration/observability/logs-and-accesslogs.md).
