---
title: "Traefik V3 Migration Documentation"
description: "Migrate from Traefik Proxy v2 to v3 and update all the necessary configurations to take advantage of all the improvements. Read the technical documentation."
---

# Migration Guide: From v2 to v3

How to Migrate from Traefik v2 to Traefik v3.
{: .subtitle }

The version 3 of Traefik introduces a number of breaking changes,
which require one to update their configuration when they migrate from v2 to v3.
The goal of this page is to recapitulate all of these changes,
and in particular to give examples, feature by feature, 
of how the configuration looked like in v2,
and how it now looks like in v3.

## Static configuration

### Docker & Docker Swarm

In v3, the provider Docker has been split into 2 providers:

- Docker provider (without Swarm support)
- Swarm provider  (Swarm support only)

??? example "An example usage of v2 Docker provider with Swarm"

    ```yaml tab="File (YAML)"
    providers:
      docker:
        swarmMode: true
    ```

    ```toml tab="File (TOML)"
    [providers.docker]
        swarmMode=true
    ```

    ```bash tab="CLI"
    --providers.docker.swarmMode=true
    ```

This configuration is now unsupported and would prevent Traefik to start.

#### Remediation

In v3, the `swarmMode` should not be used with the Docker provider, and, to use Swarm, the Swarm provider should be used instead.

??? example "An example usage of the Swarm provider"

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        endpoint: "tcp://127.0.0.1:2377"
    ```

    ```toml tab="File (TOML)"
    [providers.swarm]
        endpoint="tcp://127.0.0.1:2377"
    ```

    ```bash tab="CLI"
    --providers.swarm.endpoint=tcp://127.0.0.1:2377
    ```

### HTTP3 Experimental Configuration

In v3, HTTP/3 is no longer an experimental feature.
It can be enabled on entry points without the associated `experimental.http3` option, which is now removed.
It is now unsupported and would prevent Traefik to start.

??? example "An example usage of v2 Experimental `http3` option"

    ```yaml tab="File (YAML)"
    experimental:
      http3: true
    ```

    ```toml tab="File (TOML)"
    [experimental]
        http3=true
    ```

    ```bash tab="CLI"
    --experimental.http3=true
    ```

#### Remediation

The `http3` option should be removed from the static configuration experimental section.

### Consul provider

The Consul provider `namespace` option was deprecated in v2 and is now removed in v3.
It is now unsupported and would prevent Traefik to start.

??? example "An example usage of v2 Consul `namespace` option"

    ```yaml tab="File (YAML)"
    consul:
      namespace: foobar
    ```

    ```toml tab="File (TOML)"
    [consul]
        namespace=foobar
    ```

    ```bash tab="CLI"
    --consul.namespace=foobar
    ```

#### Remediation

In v3, the `namespaces` option should be used instead of the `namespace` option.

??? example "An example usage of Consul `namespaces` option"

    ```yaml tab="File (YAML)"
    consul:
      namespaces:
        - foobar
    ```

    ```toml tab="File (TOML)"
    [consul]
        namespaces=["foobar"]
    ```

    ```bash tab="CLI"
    --consul.namespaces=foobar
    ```

### ConsulCatalog provider

The ConsulCatalog provider `namespace` option was deprecated in v2 and is now removed in v3.
It is now unsupported and would prevent Traefik to start.

??? example "An example usage of v2 ConsulCatalog `namespace` option"

    ```yaml tab="File (YAML)"
    consulCatalog:
      namespace: foobar
    ```

    ```toml tab="File (TOML)"
    [consulCatalog]
        namespace=foobar
    ```

    ```bash tab="CLI"
    --consulCatalog.namespace=foobar
    ```

#### Remediation

In v3, the `namespaces` option should be used instead of the `namespace` option.

??? example "An example usage of ConsulCatalog `namespaces` option"

    ```yaml tab="File (YAML)"
    consulCatalog:
      namespaces:
        - foobar
    ```

    ```toml tab="File (TOML)"
    [consulCatalog]
        namespaces=["foobar"]
    ```

    ```bash tab="CLI"
    --consulCatalog.namespaces=foobar
    ```

### Nomad provider

The Nomad provider `namespace` option was deprecated in v2 and is now removed in v3.
It is now unsupported and would prevent Traefik to start.

??? example "An example usage of v2 Nomad `namespace` option"

    ```yaml tab="File (YAML)"
    nomad:
      namespace: foobar
    ```

    ```toml tab="File (TOML)"
    [nomad]
        namespace=foobar
    ```

    ```bash tab="CLI"
    --nomad.namespace=foobar
    ```

#### Remediation

In v3, the `namespaces` option should be used instead of the `namespace` option.

??? example "An example usage of Nomad `namespaces` option"

    ```yaml tab="File (YAML)"
    nomad:
      namespaces:
        - foobar
    ```

    ```toml tab="File (TOML)"
    [nomad]
        namespaces=["foobar"]
    ```

    ```bash tab="CLI"
    --nomad.namespaces=foobar
    ```

### Rancher v1 Provider

In v3, the Rancher v1 provider has been removed because Rancher v1 is [no longer actively maintaned](https://rancher.com/docs/os/v1.x/en/support/),
and Rancher v2 is supported as a standard Kubernetes provider.

??? example "An example of Traefik v2 Rancher v1 configuration"

    ```yaml tab="File (YAML)"
    providers:
      rancher: {}
    ```

    ```toml tab="File (TOML)"
    [providers.rancher]
    ```

    ```bash tab="CLI"
    --providers.rancher=true
    ```

This configuration is now unsupported and would prevent Traefik to start.

#### Remediation

Rancher 2.x requires Kubernetes and does not have a metadata endpoint of its own for Traefik to query.
As such, Rancher 2.x users should utilize the [Kubernetes CRD provider](../providers/kubernetes-crd.md) directly.

Also, all Rancher provider related configuration should be removed from the static configuration.

### Marathon provider

Marathon maintenance [ended on October 31, 2021](https://github.com/mesosphere/marathon/blob/master/README.md).
In v3, the Marathon provider has been removed.

??? example "An example of v2 Marathon provider configuration"

    ```yaml tab="File (YAML)"
    providers:
      marathon: {}
    ```

    ```toml tab="File (TOML)"
    [providers.marathon]
    ```

    ```bash tab="CLI"
    --providers.marathon=true
    ```

This configuration is now unsupported and would prevent Traefik to start.

#### Remediation

All Marathon provider related configuration should be removed from the static configuration.

### InfluxDB v1

InfluxDB v1.x maintenance [ended in 2021](https://www.influxdata.com/blog/influxdb-oss-and-enterprise-roadmap-update-from-influxdays-emea/).
In v3, the InfluxDB v1 metrics provider has been removed.

??? example "An example of Traefik v2 InfluxDB v1 metrics configuration"

    ```yaml tab="File (YAML)"
    metrics:
      influxDB: {}
    ```

    ```toml tab="File (TOML)"
    [metrics.influxDB]
    ```

    ```bash tab="CLI"
    --metrics.influxDB=true
    ```

This configuration is now unsupported and would prevent Traefik to start.

#### Remediation

All InfluxDB v1 metrics provider related configuration should be removed from the static configuration.

### Pilot

Traefik Pilot is no longer available since October 4th, 2022.

??? example "An example of v2 Pilot configuration"

    ```yaml tab="File (YAML)"
    pilot:
      token: foobar
    ```

    ```toml tab="File (TOML)"
    [pilot]
        token=foobar
    ```

    ```bash tab="CLI"
    --pilot.token=foobar
    ```

In v2, Pilot configuration was deprecated and ineffective,
it is now unsupported and would prevent Traefik to start.

#### Remediation

All Pilot related configuration should be removed from the static configuration.

## Dynamic configuration

### IPWhiteList

In v3, we renamed the `IPWhiteList` middleware to `IPAllowList` without changing anything to the configuration. 

### Deprecated Options Removal

- The `tracing.datadog.globaltag` option has been removed.
- The `tls.caOptional` option has been removed from the ForwardAuth middleware, as well as from the HTTP, Consul, Etcd, Redis, ZooKeeper, Consul Catalog, and Docker providers.
- `sslRedirect`, `sslTemporaryRedirect`, `sslHost`, `sslForceHost` and `featurePolicy` options of the Headers middleware have been removed.
- The `forceSlash` option of the StripPrefix middleware has been removed.
- The `preferServerCipherSuites` option has been removed.

### Matchers

In v3, the `Headers` and `HeadersRegexp` matchers have been renamed to `Header` and `HeaderRegexp` respectively.

`PathPrefix` no longer uses regular expressions to match path prefixes.

`QueryRegexp` has been introduced to match query values using a regular expression.

`HeaderRegexp`, `HostRegexp`, `PathRegexp`, `QueryRegexp`, and `HostSNIRegexp` matchers now uses the [Go regexp syntax](https://golang.org/pkg/regexp/syntax/).

All matchers now take a single value (except `Header`, `HeaderRegexp`, `Query`, and `QueryRegexp` which take two)
and should be explicitly combined using logical operators to mimic previous behavior.

`Query` can take a single value to match is the query value that has no value (e.g. `/search?mobile`).

`HostHeader` has been removed, use `Host` instead.

### TCP LoadBalancer `terminationDelay` option

The TCP LoadBalancer `terminationDelay` option has been removed.
This option can now be configured directly on the `TCPServersTransport` level, please take a look at this [documentation](../routing/services/index.md#terminationdelay)

### Kubernetes CRDs API Group `traefik.containo.us`

In v3, the Kubernetes CRDs API Group `traefik.containo.us` has been removed. 
Please use the API Group `traefik.io` instead.

### Kubernetes Ingress API Group `networking.k8s.io/v1beta1`

In v3, the Kubernetes Ingress API Group `networking.k8s.io/v1beta1` ([removed since Kubernetes v1.22](https://kubernetes.io/docs/reference/using-api/deprecation-guide/#ingress-v122)) support has been removed.

Please use the API Group `networking.k8s.io/v1` instead.

### Traefik CRD API Version `apiextensions.k8s.io/v1beta1`

In v3, the Traefik CRD API Version `apiextensions.k8s.io/v1beta1` ([removed since Kubernetes v1.22](https://kubernetes.io/docs/reference/using-api/deprecation-guide/#customresourcedefinition-v122)) support has been removed.

Please use the CRD definition with the API Version `apiextensions.k8s.io/v1` instead.

## Operations

### Traefik RBAC Update

In v3, the support of `TCPServersTransport` has been introduced.
When using the KubernetesCRD provider, it is therefore necessary to update [RBAC](../reference/dynamic-configuration/kubernetes-crd.md#rbac) and [CRD](../reference/dynamic-configuration/kubernetes-crd.md) manifests.

### Content-Type Auto-Detection

In v3, the `Content-Type` header is not auto-detected anymore when it is not set by the backend.
One should use the `ContentType` middleware to enable the `Content-Type` header value auto-detection.

### Observability

#### gRPC Metrics

In v3, the reported status code for gRPC requests is now the value of the `Grpc-Status` header.  

#### Tracing

In v3, the tracing feature has been revamped and is now powered exclusively by [OpenTelemetry](https://opentelemetry.io/ "Link to website of OTel") (OTel).
!!! warning "Important"

    Traefik v3 **no** longer supports direct output formats for specific vendors such as Instana, Jaeger, Zipkin, Haystack, Datadog, and Elastic.
Instead, it focuses on pure OpenTelemetry implementation, providing a unified and standardized approach for observability.

Here are two possible transition strategies:

1. OTLP Ingestion Endpoints:

    Most vendors now offer OpenTelemetry Protocol (OTLP) ingestion endpoints.
    You can seamlessly integrate Traefik v3 with these endpoints to continue leveraging tracing capabilities.

2. Legacy Stack Compatibility:

    For legacy stacks that cannot immediately upgrade to the latest vendor agents supporting OTLP ingestion,
    using OpenTelemetry (OTel) collectors with appropriate exporters configuration is a viable solution.
    This allows continued compatibility with the existing infrastructure.
