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
| <a id="opt-tracing-addInternals" href="#opt-tracing-addInternals" title="#opt-tracing-addInternals">`tracing.addInternals`</a> | Enables tracing for internal resources (e.g.: `ping@internal`).                                                                                                             | false                              | No       |
| <a id="opt-tracing-serviceName" href="#opt-tracing-serviceName" title="#opt-tracing-serviceName">`tracing.serviceName`</a> | Defines the service name resource attribute.                                                                                                                                | "traefik"                          | No       |
| <a id="opt-tracing-resourceAttributes" href="#opt-tracing-resourceAttributes" title="#opt-tracing-resourceAttributes">`tracing.resourceAttributes`</a> | Defines additional resource attributes to be sent to the collector.                                                                                                         | []                                 | No       |
| <a id="opt-tracing-sampleRate" href="#opt-tracing-sampleRate" title="#opt-tracing-sampleRate">`tracing.sampleRate`</a> | The proportion of requests to trace, specified between 0.0 and 1.0.                                                                                                         | 1.0                                | No       |
| <a id="opt-tracing-capturedRequestHeaders" href="#opt-tracing-capturedRequestHeaders" title="#opt-tracing-capturedRequestHeaders">`tracing.capturedRequestHeaders`</a> | Defines the list of request headers to add as attributes.<br />It applies to client and server kind spans.                                                                  | []                                 | No       |
| <a id="opt-tracing-capturedResponseHeaders" href="#opt-tracing-capturedResponseHeaders" title="#opt-tracing-capturedResponseHeaders">`tracing.capturedResponseHeaders`</a> | Defines the list of response headers to add as attributes.<br />It applies to client and server kind spans.                                                                 | []                                 | False    |
| <a id="opt-tracing-safeQueryParams" href="#opt-tracing-safeQueryParams" title="#opt-tracing-safeQueryParams">`tracing.safeQueryParams`</a> | By default, all query parameters are redacted.<br />Defines the list of query parameters to not redact.                                                                     | []                                 | No       |
| <a id="opt-tracing-otlp-http" href="#opt-tracing-otlp-http" title="#opt-tracing-otlp-http">`tracing.otlp.http`</a> | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.                         | null/false                         | No       |
| <a id="opt-tracing-otlp-http-endpoint" href="#opt-tracing-otlp-http-endpoint" title="#opt-tracing-otlp-http-endpoint">`tracing.otlp.http.endpoint`</a> | URL of the OpenTelemetry Collector to send tracing to.<br /> Format="`<scheme>://<host>:<port><path>`"                                                                      | "https://localhost:4318/v1/tracing" | Yes      |
| <a id="opt-tracing-otlp-http-headers" href="#opt-tracing-otlp-http-headers" title="#opt-tracing-otlp-http-headers">`tracing.otlp.http.headers`</a> | Additional headers sent with tracing by the exporter to the OpenTelemetry Collector.                                                                                        |                                    | No       |
| <a id="opt-tracing-otlp-http-tls-ca" href="#opt-tracing-otlp-http-tls-ca" title="#opt-tracing-otlp-http-tls-ca">`tracing.otlp.http.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle.                                          | ""                                 | No       |
| <a id="opt-tracing-otlp-http-tls-cert" href="#opt-tracing-otlp-http-tls-cert" title="#opt-tracing-otlp-http-tls-cert">`tracing.otlp.http.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector. When using this option, setting the `key` option is required.                 | ""                                 | No       |
| <a id="opt-tracing-otlp-http-tls-key" href="#opt-tracing-otlp-http-tls-key" title="#opt-tracing-otlp-http-tls-key">`tracing.otlp.http.tls.key`</a> | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.                         | ""null/false ""                    | No       |
| <a id="opt-tracing-otlp-http-tls-insecureskipverify" href="#opt-tracing-otlp-http-tls-insecureskipverify" title="#opt-tracing-otlp-http-tls-insecureskipverify">`tracing.otlp.http.tls.insecureskipverify`</a> | If `insecureSkipVerify` is `true`, the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers. | false                              | Yes      |
| <a id="opt-tracing-otlp-grpc" href="#opt-tracing-otlp-grpc" title="#opt-tracing-otlp-grpc">`tracing.otlp.grpc`</a> | This instructs the exporter to send tracing to the OpenTelemetry Collector using gRPC.                                                                                      | false                              | No       |
| <a id="opt-tracing-otlp-grpc-endpoint" href="#opt-tracing-otlp-grpc-endpoint" title="#opt-tracing-otlp-grpc-endpoint">`tracing.otlp.grpc.endpoint`</a> | Address of the OpenTelemetry Collector to send tracing to.<br /> Format="`<host>:<port>`"                                                                                   | "localhost:4317"                   | Yes      |
| <a id="opt-tracing-otlp-grpc-headers" href="#opt-tracing-otlp-grpc-headers" title="#opt-tracing-otlp-grpc-headers">`tracing.otlp.grpc.headers`</a> | Additional headers sent with tracing by the exporter to the OpenTelemetry Collector.                                                                                        | []                                 | No       |
| <a id="opt-tracing-otlp-grpc-insecure" href="#opt-tracing-otlp-grpc-insecure" title="#opt-tracing-otlp-grpc-insecure">`tracing.otlp.grpc.insecure`</a> | Allows exporter to send tracing to the OpenTelemetry Collector without using a secured protocol.                                                                            | false                              | Yes      |
| <a id="opt-tracing-otlp-grpc-tls-ca" href="#opt-tracing-otlp-grpc-tls-ca" title="#opt-tracing-otlp-grpc-tls-ca">`tracing.otlp.grpc.tls.ca`</a> | Path to the certificate authority used for the secure connection to the OpenTelemetry Collector, it defaults to the system bundle.                                          | ""                                 | No       |
| <a id="opt-tracing-otlp-grpc-tls-cert" href="#opt-tracing-otlp-grpc-tls-cert" title="#opt-tracing-otlp-grpc-tls-cert">`tracing.otlp.grpc.tls.cert`</a> | Path to the public certificate used for the secure connection to the OpenTelemetry Collector. When using this option, setting the `key` option is required.                 | ""                                 | No       |
| <a id="opt-tracing-otlp-grpc-tls-key" href="#opt-tracing-otlp-grpc-tls-key" title="#opt-tracing-otlp-grpc-tls-key">`tracing.otlp.grpc.tls.key`</a> | This instructs the exporter to send the tracing to the OpenTelemetry Collector using HTTP.<br /> Setting the sub-options with their default values.                         | ""null/false ""                    | No       |
| <a id="opt-tracing-otlp-grpc-tls-insecureskipverify" href="#opt-tracing-otlp-grpc-tls-insecureskipverify" title="#opt-tracing-otlp-grpc-tls-insecureskipverify">`tracing.otlp.grpc.tls.insecureskipverify`</a> | If `insecureSkipVerify` is `true`, the TLS connection to the OpenTelemetry Collector accepts any certificate presented by the server regardless of the hostnames it covers. | false                              | Yes      |
