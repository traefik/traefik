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

## Requirements

{!kubernetes-requirements.md!}

!!! info "Helm Chart"

    When using the Traefik [Helm Chart](../getting-started/install-traefik.md#use-the-helm-chart), the CRDs (Custom Resource Definitions) and RBAC (Role-Based Access Control) are automatically managed for you.
    The only remaining task is to enable the `kubernetesGateway` in the chart [values](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/values.yaml#L130).

1. Install/update the Kubernetes Gateway API CRDs.

    ```bash
    # Install Gateway API CRDs from the Standard channel.
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
    ```

2. Install the additional Traefik RBAC required for Gateway API.

    ```bash
    # Install Traefik RBACs.
    kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v3.5/docs/content/reference/dynamic-configuration/kubernetes-gateway-rbac.yml
    ```

3. Deploy Traefik and enable the `kubernetesGateway` provider in the static configuration as detailed below:
       
       ```yaml tab="File (YAML)"
       providers:
         kubernetesGateway: {}
       ```

       ```toml tab="File (TOML)"
       [providers.kubernetesGateway]
       ```

       ```bash tab="CLI"
       --providers.kubernetesgateway=true
       ```

## Routing Configuration

When using the Kubernetes Gateway API provider, Traefik uses the Gateway API CRDs to retrieve its routing configuration.
Check out the Gateway API concepts [documentation](https://gateway-api.sigs.k8s.io/concepts/api-overview/),
and the dedicated [routing section](../routing/providers/kubernetes-gateway.md) in the Traefik documentation.

## Provider Configuration

### `endpoint`

_Optional, Default=""_

The Kubernetes server endpoint URL.

When deployed into Kubernetes, Traefik reads the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik tries to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

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

### `token`

_Optional, Default=""_

Bearer token used for the Kubernetes client configuration.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    token: "mytoken"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  token = "mytoken"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.token=mytoken
```

### `certAuthFilePath`

_Optional, Default=""_

Path to the certificate authority file.
Used for the Kubernetes client configuration.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.certauthfilepath=/my/ca.crt
```

### `namespaces`

_Optional, Default: []_

Array of namespaces to watch.
If left empty, Traefik watches all namespaces.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    namespaces:
    - "default"
    - "production"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  namespaces = ["default", "production"]
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.namespaces=default,production
```

### `statusAddress`

#### `ip`

_Optional, Default: ""_

This IP will get copied to the Gateway `status.addresses`, and currently only supports one IP value (IPv4 or IPv6).

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    statusAddress:
      ip: "1.2.3.4"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway.statusAddress]
  ip = "1.2.3.4"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.statusaddress.ip=1.2.3.4
```

#### `hostname`

_Optional, Default: ""_

This Hostname will get copied to the Gateway `status.addresses`.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    statusAddress:
      hostname: "example.net"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway.statusAddress]
  hostname = "example.net"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.statusaddress.hostname=example.net
```

#### `service`

_Optional_

The Kubernetes service to copy status addresses from.
When using third parties tools like External-DNS, this option can be used to copy the service `loadbalancer.status` (containing the service's endpoints IPs) to the gateways.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    statusAddress:
      service:
        namespace: default
        name: foo
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway.statusAddress.service]
  namespace = "default"
  name = "foo"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.statusaddress.service.namespace=default
--providers.kubernetesgateway.statusaddress.service.name=foo
```

### `experimentalChannel`

_Optional, Default: false_

Toggles support for the Experimental Channel resources ([Gateway API release channels documentation](https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels)).
This option currently enables support for `TCPRoute` and `TLSRoute`.

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

!!! info "Experimental Channel"

    When enabling experimental channel resources support, the experimental CRDs (Custom Resource Definitions) needs to be deployed too.

    ```bash
    # Install Gateway API CRDs from the Experimental channel.
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/experimental-install.yaml
    ```

### `labelselector`

_Optional, Default: ""_

A label selector can be defined to filter on specific GatewayClass objects only.
If left empty, Traefik processes all GatewayClass objects in the configured namespaces.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    labelselector: "app=traefik"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  labelselector = "app=traefik"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.labelselector="app=traefik"
```

### `nativeLBByDefault`

_Optional, Default: false_

Defines whether to use Native Kubernetes load-balancing mode by default.
For more information, please check out the `traefik.io/service.nativelb` [service annotation documentation](../routing/providers/kubernetes-gateway.md#native-load-balancing).

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    nativeLBByDefault: true
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  nativeLBByDefault = true
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.nativeLBByDefault=true
```

### `throttleDuration`

_Optional, Default: 0_

The `throttleDuration` option defines how often the provider is allowed to handle events from Kubernetes. This prevents
a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.

If left empty, the provider does not apply any throttling and does not drop any Kubernetes events.

The value of `throttleDuration` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    throttleDuration: "10s"
    # ...
```

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  throttleDuration = "10s"
  # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.throttleDuration=10s
```

{!traefik-for-business-applications.md!}
