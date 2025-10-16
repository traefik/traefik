---
title: "Traefik Kubernetes Gateway API Documentation"
description: "Learn how to use the Kubernetes Gateway API as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Kubernetes with Gateway API

The Kubernetes Gateway provider is a Traefik implementation of the [Gateway API](https://gateway-api.sigs.k8s.io/)
specification from the Kubernetes Special Interest Groups (SIGs).

This provider supports Standard version [v1.4.0](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v1.4.0) of the Gateway API specification.

It fully supports all HTTP core and some extended features, as well as the `TCPRoute` and `TLSRoute` resources from the [Experimental channel](https://gateway-api.sigs.k8s.io/concepts/versioning/?h=#release-channels).

For more details, check out the conformance [report](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/v1.4.0/traefik-traefik).

!!! info "Using The Helm Chart"

    When using the Traefik [Helm Chart](../../../../getting-started/install-traefik.md#use-the-helm-chart), the CRDs (Custom Resource Definitions) and RBAC (Role-Based Access Control) are automatically managed for you.
    The only remaining task is to enable the `kubernetesGateway` in the chart [values](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/values.yaml#L130).

## Requirements

{!kubernetes-requirements.md!}

1. Install/update the Kubernetes Gateway API CRDs.

    ```bash
    # Install Gateway API CRDs from the Standard channel.
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml
    ```

2. Install/update the Traefik [RBAC](../../../dynamic-configuration/kubernetes-gateway-rbac.yml).

    ```bash
    # Install Traefik RBACs.
    kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v3.5/docs/content/reference/dynamic-configuration/kubernetes-gateway-rbac.yml
    ```

## Configuration Example

You can enable the `kubernetesGateway` provider as detailed below:

```yaml tab="File (YAML)"
providers:
  kubernetesGateway: {}
  # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
# ...
```

```bash tab="CLI"
--providers.kubernetesgateway=true
```

```yaml tab="Helm Chart Values"
## Values file
providers:
  kubernetesGateway:
    enabled: true
```

## Configuration Options

<!-- markdownlint-disable MD013 -->

| Field                                                                 | Description                                                                                                                                                                                                                                                                                                                                                                                           | Default | Required |
|:----------------------------------------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="opt-providers-providersThrottleDuration" href="#opt-providers-providersThrottleDuration" title="#opt-providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.**                  | 2s      | No       |
| <a id="opt-providers-kubernetesGateway-endpoint" href="#opt-providers-kubernetesGateway-endpoint" title="#opt-providers-kubernetesGateway-endpoint">`providers.kubernetesGateway.endpoint`</a> | Server endpoint URL.<br />More information [here](#endpoint).                                                                                                                                                                                                                                                                                                                                         | ""      | No       |
| <a id="opt-providers-kubernetesGateway-experimentalChannel" href="#opt-providers-kubernetesGateway-experimentalChannel" title="#opt-providers-kubernetesGateway-experimentalChannel">`providers.kubernetesGateway.experimentalChannel`</a> | Toggles support for the Experimental Channel resources ([Gateway API release channels documentation](https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels)).<br />(ex: `TCPRoute` and `TLSRoute`)                                                                                                                                                                                    | false   | No       |
| <a id="opt-providers-kubernetesGateway-token" href="#opt-providers-kubernetesGateway-token" title="#opt-providers-kubernetesGateway-token">`providers.kubernetesGateway.token`</a> | Bearer token used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                                            | ""      | No       |
| <a id="opt-providers-kubernetesGateway-certAuthFilePath" href="#opt-providers-kubernetesGateway-certAuthFilePath" title="#opt-providers-kubernetesGateway-certAuthFilePath">`providers.kubernetesGateway.certAuthFilePath`</a> | Path to the certificate authority file.<br />Used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                            | ""      | No       |
| <a id="opt-providers-kubernetesGateway-namespaces" href="#opt-providers-kubernetesGateway-namespaces" title="#opt-providers-kubernetesGateway-namespaces">`providers.kubernetesGateway.namespaces`</a> | Array of namespaces to watch.<br />If left empty, watch all namespaces.                                                                                                                                                                                                                                                                                                                               | []      | No       |
| <a id="opt-providers-kubernetesGateway-labelselector" href="#opt-providers-kubernetesGateway-labelselector" title="#opt-providers-kubernetesGateway-labelselector">`providers.kubernetesGateway.labelselector`</a> | Allow filtering on specific resource objects only using label selectors.<br />Only to Traefik [Custom Resources](./kubernetes-crd.md#list-of-resources) (they all must match the filter).<br />No effect on Kubernetes `Secrets`, `EndpointSlices` and `Services`.<br />See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details. | ""      | No       |
| <a id="opt-providers-kubernetesGateway-throttleDuration" href="#opt-providers-kubernetesGateway-throttleDuration" title="#opt-providers-kubernetesGateway-throttleDuration">`providers.kubernetesGateway.throttleDuration`</a> | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught.                                                                                                                            | 0s      | No       |
| <a id="opt-providers-kubernetesGateway-nativeLBByDefault" href="#opt-providers-kubernetesGateway-nativeLBByDefault" title="#opt-providers-kubernetesGateway-nativeLBByDefault">`providers.kubernetesGateway.nativeLBByDefault`</a> | Defines whether to use Native Kubernetes load-balancing mode by default. For more information, please check out the `traefik.io/service.nativelb` service annotation documentation.                                                                                                                                                                                                                   | false   | No       |
| <a id="opt-providers-kubernetesGateway-statusAddress-hostname" href="#opt-providers-kubernetesGateway-statusAddress-hostname" title="#opt-providers-kubernetesGateway-statusAddress-hostname">`providers.kubernetesGateway.`<br />`statusAddress.hostname`</a> | Hostname copied to the Gateway `status.addresses`.                                                                                                                                                                                                                                                                                                                                                    | ""      | No       |
| <a id="opt-providers-kubernetesGateway-statusAddress-ip" href="#opt-providers-kubernetesGateway-statusAddress-ip" title="#opt-providers-kubernetesGateway-statusAddress-ip">`providers.kubernetesGateway.`<br />`statusAddress.ip`</a> | IP address copied to the Gateway `status.addresses`, and currently only supports one IP value (IPv4 or IPv6).                                                                                                                                                                                                                                                                                         | ""      | No       |
| <a id="opt-providers-kubernetesGateway-statusAddress-service-namespace" href="#opt-providers-kubernetesGateway-statusAddress-service-namespace" title="#opt-providers-kubernetesGateway-statusAddress-service-namespace">`providers.kubernetesGateway.`<br />`statusAddress.service.namespace`</a> | The namespace of the Kubernetes service to copy status addresses from.<br />When using third parties tools like External-DNS, this option can be used to copy the service `loadbalancer.status` (containing the service's endpoints IPs) to the Gateway `status.addresses`.                                                                                                                           | ""      | No       |
| <a id="opt-providers-kubernetesGateway-statusAddress-service-name" href="#opt-providers-kubernetesGateway-statusAddress-service-name" title="#opt-providers-kubernetesGateway-statusAddress-service-name">`providers.kubernetesGateway.`<br />`statusAddress.service.name`</a> | The name of the Kubernetes service to copy status addresses from.<br />When using third parties tools like External-DNS, this option can be used to copy the service `loadbalancer.status` (containing the service's endpoints IPs) to the Gateway `status.addresses`.                                                                                                                                | ""      | No       |

<!-- markdownlint-enable MD013 -->

### `endpoint`

The Kubernetes server endpoint URL.

When deployed into Kubernetes, Traefik reads the environment variables
`KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to
construct the endpoint.

The access token is looked up
in `/var/run/secrets/kubernetes.io/serviceaccount/token`
and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values
inside a cluster.

When the environment variables are not found, Traefik tries to connect to
the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a
Kubernetes cluster using the granted authentication
and authorization of the associated kubeconfig.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    endpoint: "http://localhost:8080"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  endpoint = "http://localhost:8080"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.endpoint=http://localhost:8080
```

## Routing Configuration

See the dedicated section in [routing](../../../../routing/providers/kubernetes-gateway.md).

!!! tip "Routing Configuration"

    When using the Kubernetes Gateway API provider, Traefik uses the Gateway API
    CRDs to retrieve its routing configuration.
    Check out the Gateway API concepts [documentation](https://gateway-api.sigs.k8s.io/concepts/api-overview/),
    and the dedicated [routing section](../../../../routing/providers/kubernetes-gateway.md)
    in the Traefik documentation.

{!traefik-for-business-applications.md!}
