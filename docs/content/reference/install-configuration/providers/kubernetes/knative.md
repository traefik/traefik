---
title: "Traefik Knative Documentation"
description: "Learn how to use the Knative as a provider for configuration discovery in Traefik Proxy. Read the technical documentation."
---

# Traefik & Knative

The Traefik Knative provider integrates with Knative Serving to provide advanced traffic management and routing capabilities for serverless applications.

[Knative](https://knative.dev) is a Kubernetes-based platform that enables serverless workloads with features like scale-to-zero, 
automatic scaling, and revision management.

The provider watches Knative `Ingress` resources and automatically configures Traefik routing rules,
enabling seamless integration between Traefik's networking capabilities and Knative's serverless platform.

## Requirements

{!kubernetes-requirements.md!}

1. Install/update the Knative CRDs.

    ```bash
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.19.0/serving-crds.yaml
    ```

2. Install the Knative Serving core components.

    ```bash
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.19.0/serving-core.yaml
    ```

3. Update the config-network configuration to use the Traefik ingress class.

    ```bash
       kubectl patch configmap/config-network \
       -n knative-serving \
       --type merge \
       -p '{"data":{"ingress.class":"traefik.ingress.networking.knative.dev"}}'
    ```

4. Add a custom domain to your Knative configuration (Optional).

    ```bash
    kubectl patch configmap config-domain \
      -n knative-serving \
      --type='merge' \
      -p='{"data":{"example.com":""}}'
    ```

5. Install/update the Traefik [RBAC](../../../dynamic-configuration/kubernetes-knative-rbac.yml).

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/v3.6/docs/content/reference/dynamic-configuration/kubernetes-knative-rbac.yml
    ```

## Configuration Example

As this provider is an experimental feature, it needs to be enabled in the experimental and in the provider sections of the configuration.
You can enable the Knative provider as detailed below:

```yaml tab="File (YAML)"
experimental:
  knative: true

providers:
  knative: {}
```

```toml tab="File (TOML)"
[experimental.knative]

[providers.knative]
```

```bash tab="CLI"
--experimental.knative=true
--providers.knative=true
```

The Knative provider uses the Knative API to retrieve its routing configuration.
The provider then watches for incoming Knative events and derives the corresponding dynamic configuration from it.

## Configuration Options

<!-- markdownlint-disable MD013 -->

| Field                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                                                                                                                                                          | Default | Required |
|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="opt-providers-providersThrottleDuration" href="#opt-providers-providersThrottleDuration" title="#opt-providers-providersThrottleDuration">`providers.providersThrottleDuration`</a> | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s      | No       |
| <a id="opt-providers-knative-endpoint" href="#opt-providers-knative-endpoint" title="#opt-providers-knative-endpoint">providers.knative.endpoint</a> | Server endpoint URL.<br />More information [here](#endpoint).                                                                                                                                                                                                                                                                                                                        |         |
| <a id="opt-providers-knative-token" href="#opt-providers-knative-token" title="#opt-providers-knative-token">providers.knative.token</a> | Bearer token used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                                                           |         |
| <a id="opt-providers-knative-certauthfilepath" href="#opt-providers-knative-certauthfilepath" title="#opt-providers-knative-certauthfilepath">providers.knative.certauthfilepath</a> | Path to the certificate authority file.<br />Used for the Kubernetes client configuration.                                                                                                                                                                                                                                                                                           |         |
| <a id="opt-providers-knative-namespaces" href="#opt-providers-knative-namespaces" title="#opt-providers-knative-namespaces">providers.knative.namespaces</a> | Array of namespaces to watch.<br />If left empty, watch all namespaces.                                                                                                                                                                                                                                                                                                              |         |
| <a id="opt-providers-knative-labelselector" href="#opt-providers-knative-labelselector" title="#opt-providers-knative-labelselector">providers.knative.labelselector</a> | Allow filtering Knative Ingress objects using label selectors.                                                                                                                                                                                                                                                                                                                       |         |
| <a id="opt-providers-knative-throttleduration" href="#opt-providers-knative-throttleduration" title="#opt-providers-knative-throttleduration">providers.knative.throttleduration</a> | Minimum amount of time to wait between two Kubernetes events before producing a new configuration.<br />This prevents a Kubernetes cluster that updates many times per second from continuously changing your Traefik configuration.<br />If empty, every event is caught.                                                                                                           | 0       |
| <a id="opt-providers-knative-privateentrypoints" href="#opt-providers-knative-privateentrypoints" title="#opt-providers-knative-privateentrypoints">providers.knative.privateentrypoints</a> | Entrypoint names used to expose the Ingress privately. If empty local Ingresses are skipped.                                                                                                                                                                                                                                                                                         |         |
| <a id="opt-providers-knative-privateservice" href="#opt-providers-knative-privateservice" title="#opt-providers-knative-privateservice">providers.knative.privateservice</a> | Kubernetes service used to expose the networking controller privately.                                                                                                                                                                                                                                                                                                               |         |
| <a id="opt-providers-knative-privateservice-name" href="#opt-providers-knative-privateservice-name" title="#opt-providers-knative-privateservice-name">providers.knative.privateservice.name</a> | Name of the private Kubernetes service.                                                                                                                                                                                                                                                                                                                                              |         |
| <a id="opt-providers-knative-privateservice-namespace" href="#opt-providers-knative-privateservice-namespace" title="#opt-providers-knative-privateservice-namespace">providers.knative.privateservice.namespace</a> | Namespace of the private Kubernetes service.                                                                                                                                                                                                                                                                                                                                         |         |
| <a id="opt-providers-knative-publicentrypoints" href="#opt-providers-knative-publicentrypoints" title="#opt-providers-knative-publicentrypoints">providers.knative.publicentrypoints</a> | Entrypoint names used to expose the Ingress publicly. If empty an Ingress is exposed on all entrypoints.                                                                                                                                                                                                                                                                             |         |
| <a id="opt-providers-knative-publicservice" href="#opt-providers-knative-publicservice" title="#opt-providers-knative-publicservice">providers.knative.publicservice</a> | Kubernetes service used to expose the networking controller publicly.                                                                                                                                                                                                                                                                                                                |         |
| <a id="opt-providers-knative-publicservice-name" href="#opt-providers-knative-publicservice-name" title="#opt-providers-knative-publicservice-name">providers.knative.publicservice.name</a> | Name of the public Kubernetes service.                                                                                                                                                                                                                                                                                                                                               |         |
| <a id="opt-providers-knative-publicservice-namespace" href="#opt-providers-knative-publicservice-namespace" title="#opt-providers-knative-publicservice-namespace">providers.knative.publicservice.namespace</a> | Namespace of the public Kubernetes service.                                                                                                                                                                                                                                                                                                                                          |         |

<!-- markdownlint-enable MD013 -->

### `endpoint`

The Kubernetes server endpoint URL.

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
## Routing Configuration

See the dedicated section in [routing](../../../routing-configuration/kubernetes/knative.md).

{!traefik-for-business-applications.md!}
