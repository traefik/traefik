---
title: "Traefik Kubernetes Gateway"
description: "The Kubernetes Gateway API can be used as a provider for routing and load balancing in Traefik Proxy. View examples in the technical documentation."
---

# Traefik & Kubernetes

The Kubernetes Gateway API, The Experimental Way. {: .subtitle }

## Configuration Examples

??? example "Configuring Kubernetes Gateway provider and Deploying/Exposing Services"

    ```yaml tab="Gateway API"
    --8<-- "content/reference/dynamic-configuration/kubernetes-gateway-simple-https.yml"
    ```

    ```yaml tab="Whoami Service"
    --8<-- "content/reference/dynamic-configuration/kubernetes-whoami-svc.yml"
    ```

    ```yaml tab="Traefik Service"
    --8<-- "content/reference/dynamic-configuration/kubernetes-gateway-traefik-lb-svc.yml"
    ```

    ```yaml tab="RBAC"
    --8<-- "content/reference/dynamic-configuration/kubernetes-gateway-rbac.yml"
    ```

## Routing Configuration

### Custom Resource Definition (CRD)

* You can find an exhaustive list, of the custom resources and their attributes in
  [the reference page](../../reference/dynamic-configuration/kubernetes-gateway.md) or in the Kubernetes
  Sigs `Gateway API` [repository](https://github.com/kubernetes-sigs/gateway-api).
* Validate that [the prerequisites](../../providers/kubernetes-gateway.md#configuration-requirements) are fulfilled
  before using the Traefik Kubernetes Gateway Provider.

You can find an excerpt of the supported Kubernetes Gateway API resources in the table below:

| Kind                               | Purpose                                                                   | Concept Behind                                                                  |
|------------------------------------|---------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| [GatewayClass](#kind-gatewayclass) | Defines a set of Gateways that share a common configuration and behaviour | [GatewayClass](https://gateway-api.sigs.k8s.io/v1alpha2/api-types/gatewayclass) |
| [Gateway](#kind-gateway)           | Describes how traffic can be translated to Services within the cluster    | [Gateway](https://gateway-api.sigs.k8s.io/v1alpha2/api-types/gateway)           |
| [HTTPRoute](#kind-httproute)       | HTTP rules for mapping requests from a Gateway to Kubernetes Services     | [Route](https://gateway-api.sigs.k8s.io/v1alpha2/api-types/httproute)           |
| [TCPRoute](#kind-tcproute)         | Allows mapping TCP requests from a Gateway to Kubernetes Services         | [Route](https://gateway-api.sigs.k8s.io/v1alpha2/guides/tcp/)                   |
| [TLSRoute](#kind-tlsroute)         | Allows mapping TLS requests from a Gateway to Kubernetes Services         | [Route](https://gateway-api.sigs.k8s.io/v1alpha2/guides/tls/)                   |

### Kind: `GatewayClass`

`GatewayClass` is cluster-scoped resource defined by the infrastructure provider. This resource represents a class of
Gateways that can be instantiated. More details on the
GatewayClass [official documentation](https://gateway-api.sigs.k8s.io/v1alpha2/api-types/gatewayclass/).

The `GatewayClass` should be declared by the infrastructure provider, otherwise please register the `GatewayClass`
[definition](../../reference/dynamic-configuration/kubernetes-gateway.md#definitions) in the Kubernetes cluster before
creating `GatewayClass` objects.

!!! info "Declaring GatewayClass"

    ```yaml
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: GatewayClass
    metadata:
      name: my-gateway-class
    spec:
      # Controller is a domain/path string that indicates
      # the controller that is managing Gateways of this class.
      controllerName: traefik.io/gateway-controller
    ```

### Kind: `Gateway`

A `Gateway` is 1:1 with the life cycle of the configuration of infrastructure. When a user creates a Gateway, some load
balancing infrastructure is provisioned or configured by the GatewayClass controller. More details on the
Gateway [official documentation](https://gateway-api.sigs.k8s.io/v1alpha2/api-types/gateway/).

Register the `Gateway` [definition](../../reference/dynamic-configuration/kubernetes-gateway.md#definitions) in the
Kubernetes cluster before creating `Gateway` objects.

Depending on the Listener Protocol, different modes and Route types are supported.

| Listener Protocol | TLS Mode       | Route Type Supported                                   |
|-------------------|----------------|--------------------------------------------------------|
| TCP               | Not applicable | [TCPRoute](#kind-tcproute)                             |
| TLS               | Passthrough    | [TLSRoute](#kind-tlsroute), [TCPRoute](#kind-tcproute) |
| TLS               | Terminate      | [TLSRoute](#kind-tlsroute), [TCPRoute](#kind-tcproute) |
| HTTP              | Not applicable | [HTTPRoute](#kind-httproute)                           |
| HTTPS             | Terminate      | [HTTPRoute](#kind-httproute)                           |

!!! info "Declaring Gateway"

    ```yaml tab="HTTP Listener"
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: Gateway
    metadata:
      name: my-http-gateway
      namespace: default
    spec:
      gatewayClassName: my-gateway-class        # [1]
      listeners:                                # [2]
        - name: http                            # [3]
          protocol: HTTP                        # [4]
          port: 80                              # [5]
          allowedRoutes:                        # [9]
            kinds:
              - kind: HTTPRoute                 # [10]
            namespaces:
              from: Selector                    # [11]
              selector:                         # [12]
                matchLabels:
                  app: foo
    ```

    ```yaml tab="HTTPS Listener"
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: Gateway
    metadata:
      name: my-https-gateway
      namespace: default
    spec:
      gatewayClassName: my-gateway-class        # [1]
      listeners:                                # [2]
        - name: https                           # [3]
          protocol: HTTPS                       # [4]
          port: 443                             # [5]
          tls:                                  # [7]
            certificateRefs:                    # [8]
              - kind: "Secret"
                name: "mysecret"
          allowedRoutes:                        # [9]
            kinds:
              - kind: HTTPSRoute                # [10]
            namespaces:
              from: Selector                    # [11]
              selector:                         # [12]
                matchLabels:
                  app: foo
    ```

    ```yaml tab="TCP Listener"
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: Gateway
    metadata:
      name: my-tcp-gateway
      namespace: default
    spec:
      gatewayClassName: my-gateway-class        # [1]
      listeners:                                # [2]
        - name: tcp                             # [3]
          protocol: TCP                         # [4]
          port: 8000                            # [5]
          allowedRoutes:                        # [9]
            kinds:
              - kind: TCPRoute                  # [10]
            namespaces:
              from: Selector                    # [11]
              selector:                         # [12]
                matchLabels:
                  app: footcp
    ```

    ```yaml tab="TLS Listener"
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: Gateway
    metadata:
      name: my-tls-gateway
      namespace: default
    spec:
      gatewayClassName: my-gateway-class        # [1]
      listeners:                                # [2]
        - name: tls                             # [3]
          protocol: TLS                         # [4]
          port: 443                             # [5]
          hostname: foo.com                     # [6]
          tls:                                  # [7]
            certificateRefs:                    # [8]
              - kind: "Secret"
                name: "mysecret"
          allowedRoutes:                        # [9]
            kinds:
              - kind: TLSRoute                  # [10]
            namespaces:
              from: Selector                    # [11]
              selector:                         # [12]
                matchLabels:
                  app: footcp
    ```

| Ref  | Attribute          | Description                                                                                                                                                 |
|------|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `gatewayClassName` | GatewayClassName used for this Gateway. This is the name of a GatewayClass resource.                                                                        |
| [2]  | `listeners`        | Logical endpoints that are bound on this Gateway's addresses. At least one Listener MUST be specified.                                                      |
| [3]  | `name`             | Name of the Listener.                                                                                                                                       |
| [4]  | `protocol`         | The network protocol this listener expects to receive (only HTTP and HTTPS are implemented).                                                                |
| [5]  | `port`             | The network port.                                                                                                                                           |
| [6]  | `hostname`         | Hostname specifies the virtual hostname to match for protocol types that define this concept. When unspecified, “”, or *, all hostnames are matched.        |
| [7]  | `tls`              | TLS configuration for the Listener. This field is required if the Protocol field is "HTTPS" or "TLS" and ignored otherwise.                                 |
| [8]  | `certificateRefs`  | The references to Kubernetes objects that contains TLS certificates and private keys (only one reference to a Kubernetes Secret is supported).              |
| [9]  | `allowedRoutes`    | Defines the types of routes that MAY be attached to a Listener and the trusted namespaces where those Route resources MAY be present.                       |
| [10] | `kind`             | The kind of the Route.                                                                                                                                      |
| [11] | `from`             | From indicates in which namespaces the Routes will be selected for this Gateway. Possible values are `All`, `Same` and `Selector` (Defaults to `Same`).     |
| [12] | `selector`         | Selector must be specified when From is set to `Selector`. In that case, only Routes in Namespaces matching this Selector will be selected by this Gateway. |

### Kind: `HTTPRoute`

`HTTPRoute` defines HTTP rules for mapping requests from a `Gateway` to Kubernetes Services.

Register the `HTTPRoute` [definition](../../reference/dynamic-configuration/kubernetes-gateway.md#definitions) in the
Kubernetes cluster before creating `HTTPRoute` objects.

!!! info "Declaring HTTPRoute"

    ```yaml
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: HTTPRoute
    metadata:
      name: http-app
      namespace: default
    spec:
      parentRefs:                               # [1]
        - name: my-tcp-gateway                  # [2]
          namespace: default                    # [3]
          sectionName: tcp                      # [4]
      hostnames:                                # [5]
        - whoami
      rules:                                    # [6]
        - matches:                              # [7]
            - path:                             # [8]
                type: Exact                     # [9]
                value: /bar                     # [10]
            - headers:                          # [11]
                name: foo                       # [12]
                value: bar                      # [13]
        - backendRefs:                          # [14]
            - name: whoamitcp                   # [15]
              weight: 1                         # [16]
              port: 8080                        # [17]
            - name: api@internal
              group: traefik.io                 # [18]
              kind: TraefikService              # [19]
    ```

| Ref  | Attribute     | Description                                                                                                                                                                 |
|------|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `parentRefs`  | References the resources (usually Gateways) that a Route wants to be attached to.                                                                                           |
| [2]  | `name`        | Name of the referent.                                                                                                                                                       |
| [3]  | `namespace`   | Namespace of the referent. When unspecified (or empty string), this refers to the local namespace of the Route.                                                             |
| [4]  | `sectionName` | Name of a section within the target resource (the Listener name).                                                                                                           |
| [5]  | `hostnames`   | A set of hostname that should match against the HTTP Host header to select a HTTPRoute to process the request.                                                              |
| [6]  | `rules`       | A list of HTTP matchers, filters and actions.                                                                                                                               |
| [7]  | `matches`     | Conditions used for matching the rule against incoming HTTP requests. Each match is independent, i.e. this rule will be matched if **any** one of the matches is satisfied. |
| [8]  | `path`        | An HTTP request path matcher. If this field is not specified, a default prefix match on the "/" path is provided.                                                           |
| [9]  | `type`        | Type of match against the path Value (supported types: `Exact`, `Prefix`).                                                                                                  |
| [10] | `value`       | The value of the HTTP path to match against.                                                                                                                                |
| [11] | `headers`     | Conditions to select a HTTP route by matching HTTP request headers.                                                                                                         |
| [12] | `type`        | Type of match for the HTTP request header match against the `values` (supported types: `Exact`).                                                                            |
| [13] | `value`       | A map of HTTP Headers to be matched. It MUST contain at least one entry.                                                                                                    |
| [14] | `backendRefs` | Defines the backend(s) where matching requests should be sent.                                                                                                              |
| [15] | `name`        | The name of the referent service.                                                                                                                                           |
| [16] | `weight`      | The proportion of traffic forwarded to a targetRef, computed as weight/(sum of all weights in targetRefs).                                                                  |
| [17] | `port`        | The port of the referent service.                                                                                                                                           |
| [18] | `group`       | Group is the group of the referent. Only `traefik.io`, `traefik.containo.us` and `gateway.networking.k8s.io` values are supported.                                          |
| [19] | `kind`        | Kind is kind of the referent. Only `TraefikService` and `Service` values are supported.                                                                                     |

### Kind: `TCPRoute`

`TCPRoute` allows mapping TCP requests from a `Gateway` to Kubernetes Services.

Register the `TCPRoute` [definition](../../reference/dynamic-configuration/kubernetes-gateway.md#definitions) in the
Kubernetes cluster before creating `TCPRoute` objects.

!!! info "Declaring TCPRoute"

    ```yaml
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: TCPRoute
    metadata:
      name: tcp-app
      namespace: default
    spec:
      parentRefs:                               # [1]
        - name: my-tcp-gateway                  # [2]
          namespace: default                    # [3]
          sectionName: tcp                      # [4]
      rules:                                    # [5]
        - backendRefs:                          # [6]
            - name: whoamitcp                   # [7]
              weight: 1                         # [8]
              port: 8080                        # [9]
            - name: api@internal
              group: traefik.containo.us        # [10]
              kind: TraefikService              # [11]
    ```

| Ref  | Attribute     | Description                                                                                                                        |
|------|---------------|------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `parentRefs`  | References the resources (usually Gateways) that a Route wants to be attached to.                                                  |
| [2]  | `name`        | Name of the referent.                                                                                                              |
| [3]  | `namespace`   | Namespace of the referent. When unspecified (or empty string), this refers to the local namespace of the Route.                    |
| [4]  | `sectionName` | Name of a section within the target resource (the Listener name).                                                                  |
| [5]  | `rules`       | Rules are a list of TCP matchers and actions.                                                                                      |
| [6]  | `backendRefs` | Defines the backend(s) where matching requests should be sent.                                                                     |
| [7]  | `name`        | The name of the referent service.                                                                                                  |
| [8]  | `weight`      | The proportion of traffic forwarded to a targetRef, computed as weight/(sum of all weights in targetRefs).                         |
| [9]  | `port`        | The port of the referent service.                                                                                                  |
| [10] | `group`       | Group is the group of the referent. Only `traefik.io`, `traefik.containo.us` and `gateway.networking.k8s.io` values are supported. |
| [11] | `kind`        | Kind is kind of the referent. Only `TraefikService` and `Service` values are supported.                                            |

### Kind: `TLSRoute`

`TLSRoute` allows mapping TLS requests from a `Gateway` to Kubernetes Services.

Register the `TLSRoute` [definition](../../reference/dynamic-configuration/kubernetes-gateway.md#definitions) in the
Kubernetes cluster before creating `TLSRoute` objects.

!!! info "Declaring TLSRoute"

    ```yaml
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: TLSRoute
    metadata:
      name: tls-app
      namespace: default
    spec:
      parentRefs:                               # [1]
        - name: my-tls-gateway                  # [2]
          namespace: default                    # [3]
          sectionName: tcp                      # [4]
      hostnames:                                # [5]
        - whoami
      rules:                                    # [6]
        - backendRefs:                          # [7]
            - name: whoamitcp                   # [8]
              weight: 1                         # [9]
              port: 8080                        # [10]
            - name: api@internal
              group: traefik.containo.us        # [11]
              kind: TraefikService              # [12]
    ```

| Ref  | Attribute     | Description                                                                                                                        |
|------|---------------|------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `parentRefs`  | References the resources (usually Gateways) that a Route wants to be attached to.                                                  |
| [2]  | `name`        | Name of the referent.                                                                                                              |
| [3]  | `namespace`   | Namespace of the referent. When unspecified (or empty string), this refers to the local namespace of the Route.                    |
| [4]  | `sectionName` | Name of a section within the target resource (the Listener name).                                                                  |
| [5]  | `hostnames`   | Defines a set of SNI names that should match against the SNI attribute of TLS ClientHello message in TLS handshake.                |
| [6]  | `rules`       | Rules are a list of TCP matchers and actions.                                                                                      |
| [7]  | `backendRefs` | Defines the backend(s) where matching requests should be sent.                                                                     |
| [8]  | `name`        | The name of the referent service.                                                                                                  |
| [9]  | `weight`      | The proportion of traffic forwarded to a targetRef, computed as weight/(sum of all weights in targetRefs).                         |
| [10] | `port`        | The port of the referent service.                                                                                                  |
| [11] | `group`       | Group is the group of the referent. Only `traefik.io`, `traefik.containo.us` and `gateway.networking.k8s.io` values are supported. |
| [12] | `kind`        | Kind is kind of the referent. Only `TraefikService` and `Service` values are supported.                                            |

{!traefik-for-business-applications.md!}
