# Traefik & Kubernetes with Gateway API

The Kubernetes Gateway API, The Experimental Way.
{: .subtitle }

Gateway API is the evolution of Kubernetes APIs that relate to `Services`, e.g. `Ingress`.
The Gateway API project is part of Kubernetes, working under SIG-NETWORK.

The Kubernetes Gateway provider is a Traefik implementation of the [service apis](https://github.com/kubernetes-sigs/service-apis)
specifications from the Kubernetes SIGs.   

This provider is proposed as an experimental feature and partially supports the service apis [v0.1.0](https://github.com/kubernetes-sigs/service-apis/releases/tag/v0.1.0) specification. 

!!! warning "Enabling The Experimental Kubernetes Gateway Provider"
    
    As this provider is in experimental stage, it needs to be activated in the experimental section of the static configuration. 
    
    ```toml tab="File (TOML)"
    [experimental]
      kubernetesGateway = true
    
    [providers.kubernetesGateway]
    #...
    ```
    
    ```yaml tab="File (YAML)"
    experimental:
      kubernetesGateway: true
    
    providers:
      kubernetesGateway: {}
      #...
    ```
    
    ```bash tab="CLI"
    --experimental.kubernetesgateway=true --providers.kubernetesgateway=true #...
    ```

## Configuration Requirements

!!! tip "All Steps for a Successful Deployment"
  
    * Add/update the Kubernetes Gateway API [definitions](../reference/dynamic-configuration/kubernetes-gateway.md#definitions).
    * Add/update the [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for the Traefik custom resources.
    * Add all needed Kubernetes Gateway API [resources](../reference/dynamic-configuration/kubernetes-gateway.md#resources).

## Examples

??? example "Kubernetes Gateway Provider Basic Example"

    ```yaml tab="Gateway API"
    --8<-- "content/reference/dynamic-configuration/kubernetes-gateway-simple-https.yml"
    ```

    ```yaml tab="Whoami Service"
    --8<-- "content/reference/dynamic-configuration/kubernetes-whoami-svc.yml"
    ```
    
    ```yaml tab="Traefik Service"
    --8<-- "content/reference/dynamic-configuration/kubernetes-gateway-traefik-lb-svc.yml"
    ```
    
    ```yaml tab="Gateway API CRDs"
    # All resources definition must be declared
    --8<-- "content/reference/dynamic-configuration/networking.x-k8s.io_gatewayclasses.yaml"
    --8<-- "content/reference/dynamic-configuration/networking.x-k8s.io_gateways.yaml"
    --8<-- "content/reference/dynamic-configuration/networking.x-k8s.io_httproutes.yaml"
    ```

    ```yaml tab="RBAC"
    --8<-- "content/reference/dynamic-configuration/kubernetes-gateway-rbac.yml"
    ```
    
The Kubernetes Service APIs provides several [guides](https://kubernetes-sigs.github.io/service-apis/guides/) of how to use their API.
Those guides will help you to go further than the example above.
The [getting started](https://kubernetes-sigs.github.io/service-apis/getting-started/) show you how to install the CRDs from their repository.
Thus, keep in mind that the Traefik Gateway provider only supports the `v0.1.0`.

For now, the Traefik Gateway Provider could be used to achieve the following set-up guides:

* [Simple Gateway](https://kubernetes-sigs.github.io/service-apis/simple-gateway/)
* [HTTP routing](https://kubernetes-sigs.github.io/service-apis/http-routing/)
* [TLS](https://kubernetes-sigs.github.io/service-apis/tls/) (Partial support: only on listeners with terminate mode)

## Resource Configuration

When using Kubernetes Gateway API as a provider,
Traefik uses Kubernetes 
[Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
to retrieve its routing configuration.

All concepts can be found in the official API concepts [documentation](https://kubernetes-sigs.github.io/service-apis/api-overview/).
Traefik implements the following resources:

* `GatewayClass` defines a set of Gateways that share a common configuration and behaviour.
* `Gateway` describes how traffic can be translated to Services within the cluster.
* `HTTPRoute` define HTTP rules for mapping requests from a Gateway to Kubernetes Services.

## Provider Configuration 

### `endpoint`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  endpoint = "http://localhost:8080"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    endpoint: "http://localhost:8080"
    # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.endpoint=http://localhost:8080
```

The Kubernetes server endpoint as URL.

When deployed into Kubernetes, Traefik will read the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token will be looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik will try to connect to the Kubernetes API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Kubernetes cluster using the granted authentication and authorization of the associated kubeconfig.

### `token`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  token = "mytoken"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    token = "mytoken"
    # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.token=mytoken
```

Bearer token used for the Kubernetes client configuration.

### `certAuthFilePath`

_Optional, Default=empty_

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.certauthfilepath=/my/ca.crt
```

Path to the certificate authority file.
Used for the Kubernetes client configuration.

### `namespaces`

_Optional, Default: all namespaces (empty array)_

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  namespaces = ["default", "production"]
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    namespaces:
    - "default"
    - "production"
    # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.namespaces=default,production
```

Array of namespaces to watch.

### `labelselector`

_Optional, Default: empty (process all resources)_

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  labelselector = "app=traefik"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    labelselector: "app=traefik"
    # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.labelselector="app=traefik"
```

By default, Traefik processes all resource objects in the configured namespaces.
A label selector can be defined to filter on specific GatewayClass objects only.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

### `throttleDuration`

_Optional, Default: 0 (no throttling)_

```toml tab="File (TOML)"
[providers.kubernetesGateway]
  throttleDuration = "10s"
  # ...
```

```yaml tab="File (YAML)"
providers:
  kubernetesGateway:
    throttleDuration: "10s"
    # ...
```

```bash tab="CLI"
--providers.kubernetesgateway.throttleDuration=10s
```
