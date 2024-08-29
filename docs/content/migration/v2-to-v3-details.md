---
title: "Traefik V3 Migration Details"
description: "Configuration changes and their details to successfully migrate from Traefik v2 to v3."
---

# Configuration Details for Migrating from Traefik v2 to v3

## Static Configuration Changes

### SwarmMode

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

#### TLS.CAOptional

Docker provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      docker:
        tls: 
          caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.docker.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.docker.tls.caOptional=true
    ```

##### Remediation

The `tls.caOptional` option should be removed from the Docker provider static configuration.

### Kubernetes Gateway API

#### Experimental Channel Resources (TLSRoute and TCPRoute)

In v3, the Kubernetes Gateway API provider does not enable support for the experimental channel API resources by default.

##### Remediation

The `experimentalChannel` option should be used to enable the support for the experimental channel API resources.

??? example "An example usage of the Kubernetes Gateway API provider with experimental channel support enabled"

    ```yaml tab="File (YAML)"
    providers:
      kubernetesGateway:
        experimentalChannel: true
    ```

    ```toml tab="File (TOML)"
    [providers.kubernetesGateway]
        experimentalChannel = true
      # ...
    ```

    ```bash tab="CLI"
    --providers.kubernetesgateway.experimentalchannel=true
    ```

### Experimental Configuration

#### HTTP3

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

##### Remediation

The `http3` option should be removed from the static configuration experimental section.
To configure `http3`, please checkout the [entrypoint configuration documentation](../routing/entrypoints.md#http3_1).

### Consul provider

#### namespace

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

##### Remediation

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

#### TLS.CAOptional

Consul provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      consul:
        tls: 
          caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.consul.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.consul.tls.caOptional=true
    ```

##### Remediation

The `tls.caOptional` option should be removed from the Consul provider static configuration.

### ConsulCatalog provider

#### namespace

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

##### Remediation

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

#### Endpoint.TLS.CAOptional

ConsulCatalog provider `endpoint.tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the Endpoint.TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      consulCatalog:
        endpoint:
          tls: 
            caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.consulCatalog.endpoint.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.consulCatalog.endpoint.tls.caOptional=true
    ```

##### Remediation

The `endpoint.tls.caOptional` option should be removed from the ConsulCatalog provider static configuration.

### Nomad provider

#### namespace

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

##### Remediation

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

#### Endpoint.TLS.CAOptional

Nomad provider `endpoint.tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the Endpoint.TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      nomad:
        endpoint:
          tls: 
            caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.nomad.endpoint.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.nomad.endpoint.tls.caOptional=true
    ```

##### Remediation

The `endpoint.tls.caOptional` option should be removed from the Nomad provider static configuration.

### Rancher v1 Provider

In v3, the Rancher v1 provider has been removed because Rancher v1 is [no longer actively maintained](https://rancher.com/docs/os/v1.x/en/support/),
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

### HTTP Provider

#### TLS.CAOptional

HTTP provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      http:
        tls: 
          caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.http.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.http.tls.caOptional=true
    ```

##### Remediation

The `tls.caOptional` option should be removed from the HTTP provider static configuration.

### ETCD Provider

#### TLS.CAOptional

ETCD provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      etcd:
        tls: 
          caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.etcd.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.etcd.tls.caOptional=true
    ```

##### Remediation

The `tls.caOptional` option should be removed from the ETCD provider static configuration.

### Redis Provider

#### TLS.CAOptional

Redis provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://pkg.go.dev/crypto/tls#ClientAuthType).

??? example "An example usage of the TLS.CAOptional option"

    ```yaml tab="File (YAML)"
    providers:
      redis:
        tls: 
          caOptional: true
    ```

    ```toml tab="File (TOML)"
    [providers.redis.tls]
        caOptional=true
    ```

    ```bash tab="CLI"
    --providers.redis.tls.caOptional=true
    ```

##### Remediation

The `tls.caOptional` option should be removed from the Redis provider static configuration.

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

### Kubernetes Ingress Path Matching

In v3, the Kubernetes Ingress default path matching does not support regexes anymore.

#### Remediation

Two levels of remediation are possible:

- Interpret the default path matcher `PathPrefix` with v2 syntax.
This can done globally for all routers with the [static configuration](#configure-the-default-syntax-in-static-configuration) or on a per-router basis by using the [traefik.ingress.kubernetes.io/router.rulesyntax](../routing/providers/kubernetes-ingress.md#annotations) annotation.

- Adapt the path regex to be compatible with the Go regex syntax and change the default path matcher to use the `PathRegexp` matcher with the [`traefik.ingress.kubernetes.io/router.pathmatcher`](../routing/providers/kubernetes-ingress.md#annotations) annotation.

## Operations Changes

### Traefik RBAC Update

In v3, the support of `TCPServersTransport` has been introduced.
When using the KubernetesCRD provider, it is therefore necessary to update [RBAC](../reference/dynamic-configuration/kubernetes-crd.md#rbac) and [CRD](../reference/dynamic-configuration/kubernetes-crd.md) manifests.

### Content-Type Auto-Detection

In v3, the `Content-Type` header is not auto-detected anymore when it is not set by the backend.
One should use the `ContentType` middleware to enable the `Content-Type` header value auto-detection.

### Observability

#### Open Connections Metric

In v3, the open connections metric has been replaced with a global one because it was erroneously at the HTTP level, and providing misleading information.
While previously produced at the entryPoint, router, and service levels, it is now replaced with a global metric.
The equivalent to `traefik_entrypoint_open_connections`, `traefik_router_open_connections` and `traefik_service_open_connections` is now `traefik_open_connections`.

#### Configuration Reload Failures Metrics

In v3, the `traefik_config_reloads_failure_total` and `traefik_config_last_reload_failure` metrics have been suppressed since they could not be implemented.

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

Please check the [OpenTelemetry Tracing provider documention](../observability/tracing/opentelemetry.md) for more information.

#### Internal Resources Observability

In v3, observability for internal routers or services (e.g.: `ping@internal`) is disabled by default.
To enable it one should use the new `addInternals` option for AccessLogs, Metrics or Tracing.
Please take a look at the observability documentation for more information:

- [AccessLogs](../observability/access-logs.md#addinternals)
- [Metrics](../observability/metrics/overview.md#addinternals)
- [Tracing](../observability/tracing/overview.md#addinternals)

#### Access logs

In v3, the `ServiceURL` field is not an object anymore but a string representation.
An update may be required if you index access logs.

## Dynamic Configuration Changes

### Router Rule Matchers

In v3, a new rule matchers syntax has been introduced for HTTP and TCP routers.
The default rule matchers syntax is now the v3 one, but for backward compatibility this can be configured.
The v2 rule matchers syntax is deprecated and its support will be removed in the next major version.
For this reason, we encourage migrating to the new syntax.

By default, the `defaultRuleSyntax` static option is automatically set to `v3`, meaning that the default rule is the new one.

#### New V3 Syntax Notable Changes

The `Headers` and `HeadersRegexp` matchers have been renamed to `Header` and `HeaderRegexp` respectively.

`PathPrefix` no longer uses regular expressions to match path prefixes.

`QueryRegexp` has been introduced to match query values using a regular expression.

`HeaderRegexp`, `HostRegexp`, `PathRegexp`, `QueryRegexp`, and `HostSNIRegexp` matchers now uses the [Go regexp syntax](https://golang.org/pkg/regexp/syntax/).

All matchers now take a single value (except `Header`, `HeaderRegexp`, `Query`, and `QueryRegexp` which take two)
and should be explicitly combined using logical operators to mimic previous behavior.

`Query` can take a single value to match is the query value that has no value (e.g. `/search?mobile`).

`HostHeader` has been removed, use `Host` instead.

#### Remediation

##### Configure the Default Syntax In Static Configuration

The default rule matchers syntax is the expected syntax for any router that is not self opt-out from this default value.
It can be configured in the static configuration.

??? example "An example configuration for the default rule matchers syntax"

    ```yaml tab="File (YAML)"
    # static configuration
    core:
      defaultRuleSyntax: v2
    ```

    ```toml tab="File (TOML)"
    # static configuration
    [core]
        defaultRuleSyntax="v2"
    ```

    ```bash tab="CLI"
    # static configuration
    --core.defaultRuleSyntax=v2
    ```

##### Configure the Syntax Per Router

The rule syntax can also be configured on a per-router basis.
This allows to have heterogeneous router configurations and ease migration.

??? example "An example router with syntax configuration"

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.routers.test.ruleSyntax=v2"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: test.route
  namespace: default

spec:
  routes:
    - match: PathPrefix(`/foo`, `/bar`)
      syntax: v2
      kind: Rule
```

```yaml tab="Consul Catalog"
- "traefik.http.routers.test.ruleSyntax=v2"
```

```yaml tab="File (YAML)"
http:
  routers:
    test:
      ruleSyntax: v2
```

```toml tab="File (TOML)"
[http.routers]
  [http.routers.test]
    ruleSyntax = "v2"
```

### IPWhiteList

In v3, we renamed the `IPWhiteList` middleware to `IPAllowList` without changing anything to the configuration. 

### Deprecated Options Removal

- The `tracing.datadog.globaltag` option has been removed.
- The `tls.caOptional` option has been removed from the ForwardAuth middleware, as well as from the HTTP, Consul, Etcd, Redis, ZooKeeper, Consul Catalog, and Docker providers.
- `sslRedirect`, `sslTemporaryRedirect`, `sslHost`, `sslForceHost` and `featurePolicy` options of the Headers middleware have been removed.
- The `forceSlash` option of the StripPrefix middleware has been removed.
- The `preferServerCipherSuites` option has been removed.

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
