---
title: "Kubernetes serversTransport"
description: "The Kubernetes ServersTransport allows configuring the connection between Traefik and the HTTP servers in Kubernetes."
---

A `ServersTransport` allows you to configure the connection between Traefik and the HTTP servers in Kubernetes.

Before creating `ServersTransport` objects, you need to apply the [Traefik Kubernetes CRDs](https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions) to your Kubernetes cluster.

This registers the `ServersTransport` kind and other Traefik-specific resources.

It can be applied on a service using:

- The option `services.serverstransport` on a [`IngressRoute`](./ingressroute.md) (if the service is a Kubernetes Service)
- The option `serverstransport` on a [`TraefikService`](./traefikservice.md) (if the service is a Kubernetes Service)

!!! note "Reference a ServersTransport CRD from another namespace"

    The value must be of form `namespace-name@kubernetescrd`, and the `allowCrossNamespace` option must be enabled at the provider level.

## Configuration Example

```yaml tab="serversTransport"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: mytransport
  namespace: default

spec:
  serverName: example.org
  insecureSkipVerify: true
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: testroute
  namespace: default

spec:
  entryPoints:
    - web
  routes:
  - match: Host(`example.com`)
    kind: Rule
    services:
    - name: whoami
      port: 80
      serversTransport: mytransport
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `serverstransport.`<br />`serverName` | Defines the server name that will be used for SNI. |  | No |
| `serverstransport.`<br />`insecureSkipVerify` | Controls whether the server's certificate chain and host name is verified. | false  | No |
| `serverstransport.`<br />`rootcas` | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections). |  | No |
| `serverstransport.`<br />`certificatesSecrets` | Certificates to present to the server for mTLS. |  | No |
| `serverstransport.`<br />`maxIdleConnsPerHost` | Maximum idle (keep-alive) connections to keep per-host. | 200 | No |
| `serverstransport.`<br />`disableHTTP2` | Disables HTTP/2 for connections with servers. | false | No |
| `serverstransport.`<br />`peerCertURI` | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| `serverstransport.`<br />`forwardingTimeouts.dialTimeout` | Amount of time to wait until a connection to a server can be established.<br />Zero means no timeout. | 30s  | No |
| `serverstransport.`<br />`forwardingTimeouts.responseHeaderTimeout` | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />Zero means no timeout | 0s  | No |
| `serverstransport.`<br />`forwardingTimeouts.idleConnTimeout` | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />Zero means no timeout. | 90s  | No |
| `serverstransport.`<br />`spiffe.ids` | Allow SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. |  | No |
| `serverstransport.`<br />`spiffe.trustDomain` | Allow SPIFFE trust domain. | ""  | No |

!!! note "CA Secret"
    The CA secret must contain a base64 encoded certificate under either a tls.ca or a ca.crt key.
