---
title: "Traefik Tracing Overview"
description: "The Traefik Proxy tracing system allows developers to visualize call flows in their infrastructure. Read the full documentation."
---

# Tracing

Visualize the Requests Flow
{: .subtitle }

The tracing system allows developers to visualize call flows in their infrastructure.

Traefik uses [OpenTelemetry](https://opentelemetry.io/ "Link to website of OTel"), an open standard designed for distributed tracing.

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

```yaml tab="Helm Chart Values"
  tracing:
    otlp:
        enabled: true
```

## Configuration Options

| Field                                      | Description                                                                                                                                                                 | Default                            | Required |
|:-------------------------------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-----------------------------------|:---------|
| <a id="tracing-addInternals" href="#tracing-addInternals" title="#tracing-addInternals">`tracing.addInternals`</a> | Enables tracing for internal resources (e.g.: `ping@internal`).                                                                                                             | false                              | No       |
| <a id="tracing-serviceName" href="#tracing-serviceName" title="#tracing-serviceName">`tracing.serviceName`</a> | Defines the service name resource attribute.                                                                                                                                | "traefik"                          | No       |
| <a id="tracing-resourceAttributes" href="#tracing-resourceAttributes" title="#tracing-resourceAttributes">`tracing.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                                                         | []                                 | No       |
| <a id="tracing-sampleRate" href="#tracing-sampleRate" title="#tracing-sampleRate">`tracing.sampleRate`</a> | The proportion of requests to trace, specified between 0.0 and 1.0.                                                                                                         | 1.0                                | No       |
| <a id="tracing-capturedRequestHeaders" href="#tracing-capturedRequestHeaders" title="#tracing-capturedRequestHeaders">`tracing.capturedRequestHeaders`</a> | Defines the list of request headers to add as attributes.<br />It applies to client and server kind spans.                                                                  | []                                 | No       |
| <a id="tracing-capturedResponseHeaders" href="#tracing-capturedResponseHeaders" title="#tracing-capturedResponseHeaders">`tracing.capturedResponseHeaders`</a> | Defines the list of response headers to add as attributes.<br />It applies to client and server kind spans.                                                                 | []                                 | False    |
| <a id="tracing-safeQueryParams" href="#tracing-safeQueryParams" title="#tracing-safeQueryParams">`tracing.safeQueryParams`</a> | By default, all query parameters are redacted.<br />Defines the list of query parameters to not redact.                                                                     | []                                 | No       |
| <a id="tracing-otlp-http" href="#tracing-otlp-http" title="#tracing-otlp-http">`tracing.otlp.http`</a> | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.                         | null/false                         | No       |
| <a id="tracing-otlp-http-endpoint" href="#tracing-otlp-http-endpoint" title="#tracing-otlp-http-endpoint">`tracing.otlp.http.endpoint`</a> | URL of the OpenTelemetry Collector to send tracing to.<br /> Format="`<scheme>://<host>:<port><path>`"                                                                      | "http://localhost:4318/v1/tracing" | Yes      |
| <a id="tracing-otlp-http-headers" href="#tracing-otlp-http-headers" title="#tracing-otlp-http-headers">`tracing.otlp.http.headers`</a> | Additional headers sent with tracing by the exporter to the OpenTelemetry Collector.                                                                                        |                                    | No       |
| <a id="tracing-otlp-http-tls-ca" href="#tracing-otlp-http-tls-ca" title="#tracing-otlp-http-tls-ca">`tracing.otlp.http.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle.                                          | ""                                 | No       |
| <a id="tracing-otlp-http-tls-cert" href="#tracing-otlp-http-tls-cert" title="#tracing-otlp-http-tls-cert">`tracing.otlp.http.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector. When using this option, setting the `key` option is required.                 | ""                                 | No       |
| <a id="tracing-otlp-http-tls-key" href="#tracing-otlp-http-tls-key" title="#tracing-otlp-http-tls-key">`tracing.otlp.http.tls.key`</a> | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.                         | ""null/false ""                    | No       |
| <a id="tracing-otlp-http-tls-insecureskipverify" href="#tracing-otlp-http-tls-insecureskipverify" title="#tracing-otlp-http-tls-insecureskipverify">`tracing.otlp.http.tls.insecureskipverify`</a> | If `insecureSkipVerify` is `true`, the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers. | false                              | Yes      |
| <a id="tracing-otlp-grpc" href="#tracing-otlp-grpc" title="#tracing-otlp-grpc">`tracing.otlp.grpc`</a> | This instructs the exporter to send tracing to the OpenTelemetry Collector using gRPC.                                                                                      | false                              | No       |
| <a id="tracing-otlp-grpc-endpoint" href="#tracing-otlp-grpc-endpoint" title="#tracing-otlp-grpc-endpoint">`tracing.otlp.grpc.endpoint`</a> | Address of the OpenTelemetry Collector to send tracing to.<br /> Format="`<host>:<port>`"                                                                                   | "localhost:4317"                   | Yes      |
| <a id="tracing-otlp-grpc-headers" href="#tracing-otlp-grpc-headers" title="#tracing-otlp-grpc-headers">`tracing.otlp.grpc.headers`</a> | Additional headers sent with tracing by the exporter to the OpenTelemetry Collector.                                                                                        | []                                 | No       |
| <a id="tracing-otlp-grpc-insecure" href="#tracing-otlp-grpc-insecure" title="#tracing-otlp-grpc-insecure">`tracing.otlp.grpc.insecure`</a> | Allows exporter to send tracing to the OpenTelemetry Collector without using a secured protocol.                                                                            | false                              | Yes      |
| <a id="tracing-otlp-grpc-tls-ca" href="#tracing-otlp-grpc-tls-ca" title="#tracing-otlp-grpc-tls-ca">`tracing.otlp.grpc.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle.                                          | ""                                 | No       |
| <a id="tracing-otlp-grpc-tls-cert" href="#tracing-otlp-grpc-tls-cert" title="#tracing-otlp-grpc-tls-cert">`tracing.otlp.grpc.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector. When using this option, setting the `key` option is required.                 | ""                                 | No       |
| <a id="tracing-otlp-grpc-tls-key" href="#tracing-otlp-grpc-tls-key" title="#tracing-otlp-grpc-tls-key">`tracing.otlp.grpc.tls.key`</a> | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.                         | ""null/false ""                    | No       |
| <a id="tracing-otlp-grpc-tls-insecureskipverify" href="#tracing-otlp-grpc-tls-insecureskipverify" title="#tracing-otlp-grpc-tls-insecureskipverify">`tracing.otlp.grpc.tls.insecureskipverify`</a> | If `insecureSkipVerify` is `true`, the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers. | false                              | Yes      |
