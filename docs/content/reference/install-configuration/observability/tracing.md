---
title: "Traefik Tracing Overview"
description: "The Traefik Proxy tracing system allows developers to visualize call flows in their infrastructure. Read the full documentation."
---

# Tracing

Visualize the Requests Flow
{: .subtitle }

The tracing system allows developers to visualize call flows in their infrastructure.

Traefik uses [OpenTelemetry](https://opentelemetry.io/ "Link to website of OTel"), an open standard designed for distributed tracing.

Please check our dedicated [OTel docs](./metrics/metrics.md#open-telemetry) to learn more.

## Configuration Example

To enable the tracing:

```yaml tab="File (YAML)"
tracing: {}
```

```toml tab="File (TOML)"
[tracing]
```

```bash tab="CLI"
--tracing=true
```

## Configuration Options

| Field      | Description                                                                                                                                                                                 | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `tracing.addInternals` | Enables tracing for internal resources (e.g.: `ping@internals`). | false      | No      |
| `tracing.otlp.http.enabled` | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values. | null/false      | No      |
| `tracing.otlp.http.endpoint` | URL of the OpenTelemetry Collector to send tracing to.<br /> Format="`<scheme>://<host>:<port><path>`" | "http://localhost:4318/v1/tracing"      | Yes      |
| `tracing.otlp.http.headers` | Additional headers sent with tracing by the exporter to the OpenTelemetry Collector. |       | No      |
| `tracing.otlp.http.tls.ca` | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. | ""  | No      |
| `tracing.otlp.http.tls.cert` | Path to the public certificate used for the secure connection to the OpenTelemetry Collector. When using this option, setting the `key` option is required. | ""      | No      |
| `tracing.otlp.http.tls.key` | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values. | ""null/false ""     | No      |
| `tracing.otlp.http.tls.insecureskipverify` |If `insecureSkipVerify` is `true`, the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.  | false | Yes      |
| `tracing.otlp.grpc.enabled` | This instructs the exporter to send tracing to the OpenTelemetry Collector using gRPC. | false | No      |
| `tracing.otlp.grpc.endpoint` | Address of the OpenTelemetry Collector to send tracing to.<br /> Format="`<host>:<port>`" | "localhost:4317"      | Yes      |
| `tracing.otlp.grpc.headers` | Additional headers sent with tracing by the exporter to the OpenTelemetry Collector. |       | No      |
| `tracing.otlp.grpc.insecure` |Allows exporter to send tracing to the OpenTelemetry Collector without using a secured protocol.  | false | Yes      |
| `tracing.otlp.grpc.tls.ca` | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle. | ""  | No      |
| `tracing.otlp.grpc.tls.cert` | Path to the public certificate used for the secure connection to the OpenTelemetry Collector. When using this option, setting the `key` option is required. | ""      | No      |
| `tracing.otlp.grpc.tls.key` | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values. | ""null/false ""     | No      |
| `tracing.otlp.grpc.tls.insecureskipverify` |If `insecureSkipVerify` is `true`, the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers.  | false | Yes      |
