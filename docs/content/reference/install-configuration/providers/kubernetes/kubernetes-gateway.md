---
title: "Traefik Kubernetes Gateway API Documentation"
description: "Learn how to use the Kubernetes Gateway API as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Kubernetes with Gateway API

The Kubernetes Gateway provider is a Traefik implementation of the [Gateway API](https://gateway-api.sigs.k8s.io/)
specification from the Kubernetes Special Interest Groups (SIGs).

This provider supports Standard version [v1.3.0](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v1.3.0) of the Gateway API specification.

It fully supports all HTTP core and some extended features, as well as the `TCPRoute` and `TLSRoute` resources from the [Experimental channel](https://gateway-api.sigs.k8s.io/concepts/versioning/?h=#release-channels).

For more details, check out the conformance [report](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/v1.3.0/traefik-traefik).

!!! info "Using The Helm Chart"

    When using the Traefik [Helm Chart](../../../../getting-started/install-traefik.md#use-the-helm-chart), the CRDs (Custom Resource Definitions) and RBAC (Role-Based Access Control) are automatically managed for you.
    The only remaining task is to enable the `kubernetesGateway` in the chart [values](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/values.yaml#L130).

## Requirements

{!kubernetes-requirements.md!}

1. Install/update the Kubernetes Gateway API CRDs.

    ```bash
    # Install Gateway API CRDs from the Standard channel.
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
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
| `providers.providersThrottleDuration`                                 | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.**                  | 2s      | No       |
| `providers.kubernetesGateway.endpoint`                                | Server endpoint URL.<br />More information [here](#endpoint).                                                                                                                                                                                                                                                                                                                                         | ""      | No       |
| `providers.kubernetesGateway.experimentalChannel`                     | Toggles support for the Experimental Channel resources ([Gateway API release channels documentation](https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels)).<br />(ex: `TCPRoute` and `TLSRoute`)                                                                                                                                                                                    | false   | No       |
| `providers.kubernetesGateway.token`                                   | Bearer token used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                                            | ""      | No       |
| `providers.kubernetesGateway.certAuthFilePath`                        | Path to the certificate authority file.<br />Used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                            | ""      | No       |
| `providers.kubernetesGateway.namespaces`                              | Array of namespaces to watch.<br />If left empty, watch all namespaces.                                                                                                                                                                                                                                                                                                                               | []      | No       |
| `providers.kubernetesGateway.labelselector`                           | Allow filtering on specific resource objects only using label selectors.<br />Only to Traefik [Custom Resources](./kubernetes-crd.md#list-of-resources) (they all must match the filter).<br />No effect on Kubernetes `Secrets`, `EndpointSlices` and `Services`.<br />See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details. | ""      | No       |
| `providers.kubernetesGateway.throttleDuration`                        | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught.                                                                                                                            | 0s      | No       |
| `providers.kubernetesGateway.nativeLBByDefault`                       | Defines whether to use Native Kubernetes load-balancing mode by default. For more information, please check out the `traefik.io/service.nativelb` service annotation documentation.                                                                                                                                                                                                                   | false   | No       |
| `providers.kubernetesGateway.`<br />`statusAddress.hostname`          | Hostname copied to the Gateway `status.addresses`.                                                                                                                                                                                                                                                                                                                                                    | ""      | No       |
| `providers.kubernetesGateway.`<br />`statusAddress.ip`                | IP address copied to the Gateway `status.addresses`, and currently only supports one IP value (IPv4 or IPv6).                                                                                                                                                                                                                                                                                         | ""      | No       |
| `providers.kubernetesGateway.`<br />`statusAddress.service.namespace` | The namespace of the Kubernetes service to copy status addresses from.<br />When using third parties tools like External-DNS, this option can be used to copy the service `loadbalancer.status` (containing the service's endpoints IPs) to the Gateway `status.addresses`.                                                                                                                           | ""      | No       |
| `providers.kubernetesGateway.`<br />`statusAddress.service.name`      | The name of the Kubernetes service to copy status addresses from.<br />When using third parties tools like External-DNS, this option can be used to copy the service `loadbalancer.status` (containing the service's endpoints IPs) to the Gateway `status.addresses`.                                                                                                                                | ""      | No       |

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
