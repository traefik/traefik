---
title: "Traefik Knative Documentation"
description: "Learn how to use the Knative as a provider for configuration discovery in Traefik Proxy. Read the 
technical documentation."
---

# Traefik & Knative

The Knative Provider.
{: .subtitle }

The Traefik Knative provider integrates with Knative to manage service access, enabling the use of Traefik Proxy as a router. It fully supports all routing types.

## Requirements

{!kubernetes-requirements.md!}

1. Install/update the Knative CRDs.

    ```bash
    # Install Knative CRDs from the Standard channel.
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-crds.yaml
    ```

2. Install the Knative Serving core components.

    ```bash
    # Install Knative Serving core components.
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.17.0/serving-core.yaml
    ```
   
3. Update the config-network configuration to use the Traefik ingress class.

   ```bash
       kubectl patch configmap/config-network \
       -n knative-serving \
       --type merge \
       -p '{"data":{"ingress.class":"traefik.ingress.networking.knative.dev"}}'
    ```
   
4. (Optional) Add a custom domain to your Knative configuration.

   ```bash
    kubectl patch configmap config-domain \
      -n knative-serving \
      --type='merge' \
      -p='{"data":{"example.com":""}}'
    ```
   
5. Deploy Traefik and enable the `knative` provider in the static configuration as detailed below:

       ```yaml tab="File (YAML)"
       providers:
         knative: {}
       ```

       ```toml tab="File (TOML)"
       [providers.knative]
       ```

       ```bash tab="CLI"
       --providers.knative=true
       ```

## Routing Configuration

See the dedicated section in [routing](../routing/providers/knative.md).

The Knative provider uses the Knative API to retrieve its routing configuration.
The provider then watches for incoming Knative events and derives the corresponding dynamic configuration from it.

## Provider Configuration

### `endpoint`

_Optional, Default=""_

The Knative server endpoint URL.

When deployed into Kubernetes, Traefik reads the environment variables `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` or `KUBECONFIG` to construct the endpoint.

The access token is looked up in `/var/run/secrets/kubernetes.io/serviceaccount/token` and the SSL CA certificate in `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`.
Both are mounted automatically when deployed inside Kubernetes.

The endpoint may be specified to override the environment variable values inside a cluster.

When the environment variables are not found, Traefik tries to connect to the Knative API server with an external-cluster client.
In this case, the endpoint is required.
Specifically, it may be set to the URL used by `kubectl proxy` to connect to a Knative cluster using the granted authentication and authorization of the associated kubeconfig.

```yaml tab="File (YAML)"
providers:
  knative:
    endpoint: "http://localhost:8080"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  endpoint = "http://localhost:8080"
  # ...
```

```bash tab="CLI"
--providers.knative.endpoint=http://localhost:8080
```

### `token`

_Optional, Default=""_

Bearer token used for the Knative client configuration.

```yaml tab="File (YAML)"
providers:
  knative:
    token: "mytoken"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  token = "mytoken"
  # ...
```

```bash tab="CLI"
--providers.knative.token=mytoken
```

### `certAuthFilePath`

_Optional, Default=""_

Path to the certificate authority file.
Used for the Knative client configuration.

```yaml tab="File (YAML)"
providers:
  knative:
    certAuthFilePath: "/my/ca.crt"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  certAuthFilePath = "/my/ca.crt"
  # ...
```

```bash tab="CLI"
--providers.knative.certauthfilepath=/my/ca.crt
```

### `namespaces`

_Optional, Default: []_

Array of namespaces to watch.
If left empty, Traefik watches all namespaces.

```yaml tab="File (YAML)"
providers:
  knative:
    namespaces:
      - "default"
      - "production"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  namespaces = ["default", "production"]
  # ...
```

```bash tab="CLI"
--providers.knative.namespaces=default,production
```

### `labelSelector`

_Optional, Default: ""_

A label selector can be defined to filter on specific Knative objects only.
If left empty, Traefik processes all Knative objects in the configured namespaces.

See [label-selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) for details.

```yaml tab="File (YAML)"
providers:
  knative:
    labelSelector: "app=traefik"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  labelSelector = "app=traefik"
  # ...
```

```bash tab="CLI"
--providers.knative.labelselector="app=traefik"
```

### `throttleDuration`

_Optional, Default: 0_

The `throttleDuration` option defines how often the provider is allowed to handle events from Knative. This prevents
a Knative cluster that updates many times per second from continuously changing your Traefik configuration.

If left empty, the provider does not apply any throttling and does not drop any Knative events.

The value of `throttleDuration` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  knative:
    throttleDuration: "10s"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  throttleDuration = "10s"
  # ...
```

```bash tab="CLI"
--providers.knative.throttleDuration=10s
```

### `entrypoints`

If no entrypoints are specified, all entrypoints are used by default.

_Optional, Default: []_

Array of entrypoints to use for the Knative provider.

```yaml tab="File (YAML)"
providers:
  knative:
    entrypoints:
      - "web"
      - "websecure"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  entrypoints = ["web", "websecure"]
  # ...
```

```bash tab="CLI"
--providers.knative.entrypoints=web,websecure
```

### `entrypointsinternal`

_Optional, Default: []_

Array of internal entrypoints to use for the Knative provider. This can be used when no external domain is configured. 

```yaml tab="File (YAML)"
providers:
  knative:
    entrypointsinternal:
      - "internal"
    # ...
```

```toml tab="File (TOML)"
[providers.knative]
  entrypointsinternal = ["internal"]
  # ...
```

```bash tab="CLI"
--providers.knative.entrypointsinternal=internal
```

{!traefik-for-business-applications.md!}
