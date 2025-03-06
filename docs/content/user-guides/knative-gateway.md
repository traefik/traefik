---
title: "Traefik Proxy Knative Gateway API Integration"
description: "This section of the Traefik Proxy documentation explains how to use Traefik as reverse proxy with Knative, via the Gateway API integration provided by Knative."
---
# Traefik Knative Gateway API Integration

Make use of [Knative][kn], a popular serverless application platform built for Kubernetes.
{: .subtitle }

[kn]: https://knative.dev/docs/

## How Knative integrates with Traefik via Gateway API

Traefik integrates with [Knative Serving][kn-serving] by providing a powerful and flexible routing layer,
via the [Kubernetes Gateway API][k8s-gateway], with the use of [`net-gateway-api`][kn-net-gateway-api].

Traefik supports recent versions of [Gateway API][traefik-gateway-api] along with extended features like `TCPRoute` and `TLSRoute`.

!!! info
    To get started quickly with Gateway API support in Traefik, see ["Getting started with Kubernetes Gateway API and Traefik][community-getting-started-traefik-gateway-api].

[k8s-gateway]: https://gateway-api.sigs.k8s.io/
[traefik-gateway-api]: https://doc.traefik.io/traefik/providers/kubernetes-gateway/
[kn-serving]: https://knative.dev/docs/serving/
[kn-net-gateway-api]: https://github.com/knative-extensions/net-gateway-api
[community-getting-started-traefik-gateway-api]: https://community.traefik.io/t/getting-started-with-kubernetes-gateway-api-and-traefik/23601

## Pre-requisites

To get started, you'll need:

1. A Kubernetes cluster with [Gateway API][k8s-gateway] enabled
2. A working Traefik installation with [Gateway API support][traefik-gateway-api] enabled
3. A working Knative installation with Knative serving installed
  - Knative serving must be configured with the correct `ingress.class` which enables Gateway API (see [`net-gateway-api` instructions][kn-net-gateway-api-instr])
  - `net-gateway-api` must be installed

!!! info
    Traefik integration with Knative works via the Gateway API. Consider ensuring that your Gateway API support is working properly independently of Knative first.

[kn-net-gateway-api-instr]: https://github.com/knative-extensions/net-gateway-api?tab=readme-ov-file#getting-started

## Example configuration

This section details an configuration for using Traefik along with Knative and the [`net-gateway-api`][kn-net-gateway-api] (a [Knative networking layer][kn-networking] for Knative) to deploy an example workload.

[kn-networking]: https://github.com/knative/networking

### Configuring Knative to use `net-gateway-api`

To ensure that Knative serving uses `net-gateway-api` (and attempts to create appropriate `HTTPRoute` objects), you must add the following `ConfigMap`:

```yaml tab="Kubernetes"
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-network
  namespace: knative-serving
data:
  ingress.class: gateway-api.ingress.networking.knative.dev
```

### Configuring Traefik as the external gateway for Knative serving (via `net-gateway-api`)

To ensure that when `HTTPRoute`s are created by `net-gateway-api` they are configured to use Traefik, you'll need the `config-gateway` ConfigMap:

```yaml tab="Kubernetes"
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-gateway
  namespace: knative-serving
data:
  external-gateways: |
    - class: traefik
      gateway: ingress/traefik
      service: ingress/traefik
      supported-features:
        - HTTPRouteRequestTimeout
```

!!! warning
    You may need to modify the gateway namespace (`ingress`) and gateway name (`traefik`) above

### Deploying the example workload Knative Serving workload

To start a Knative workload, which will trigger the creation of a dedicated `HTTPRoute`  uses the provided ingress, use:

```yaml tab="Kubernetes"
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: example-go
spec:
  template:
    spec:
      containers:
        - image: ghcr.io/knative/helloworld-go@sha256:584dae6fbc79fbf2cd8f94168b75b200d16dd36121eea55845e775aa547b00c8
          ports:
            - containerPort: 8080
          env:
            - name: TARGET
              value: Go Sample v1
```

Creating this [Knative `Service`][kn-service] will trigger the creation of a `HTTPRoute` with Traefik configured as the upstream gateway (via `parentRef`).

For more information on more advanced topics like ([configuring ingress classes][kn-docs-configure-ingress-class]), see the [Knative serving documentation][kn-docs-serving].

[kn-docs-configure-ingress-class]: https://knative.dev/docs/serving/services/ingress-class/
[kn-service]: https://knative.dev/docs/serving/services/
[kn-docs-serving]: https://knative.dev/docs/serving/

### Configuring Knative to use a custom domain for the example workload

At this point an *internal* `HTTPRoute` for the Knative service should be available:

```
kubectl get httproute
NAME                                   HOSTNAMES                                                                                AGE
example-go.ingress.svc.cluster.local   ["example-go.ingress","example-go.ingress.svc","example-go.ingress.svc.cluster.local"]   8s
```

To configure an [external domain for a Knative Service][knative-docs-external-domain], create the following resources:

```yaml tab="Kubernetes"
apiVersion: networking.internal.knative.dev/v1alpha1
kind: ClusterDomainClaim
metadata:
  name: example.localhost
spec:
  namespace: default
```

Once you have a claim on the domain, you can create a mapping for the service we'll create

```yaml tab="Kubernetes"
apiVersion: serving.knative.dev/v1beta1
kind: DomainMapping
metadata:
  name: example.localhost
spec:
  ref:
    name: example-go
    kind: Service
    apiVersion: serving.knative.dev/v1
```

[knative-docs-external-domain]: https://knative.dev/docs/serving/services/custom-domains/
