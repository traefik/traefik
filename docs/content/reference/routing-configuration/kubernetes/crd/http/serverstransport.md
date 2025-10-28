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
| <a id="opt-serverstransport-serverName" href="#opt-serverstransport-serverName" title="#opt-serverstransport-serverName">`serverstransport.`<br />`serverName`</a> | Defines the server name that will be used for SNI. |  | No |
| <a id="opt-serverstransport-insecureSkipVerify" href="#opt-serverstransport-insecureSkipVerify" title="#opt-serverstransport-insecureSkipVerify">`serverstransport.`<br />`insecureSkipVerify`</a> | Controls whether the server's certificate chain and host name is verified. | false  | No |
| <a id="opt-serverstransport-rootcas" href="#opt-serverstransport-rootcas" title="#opt-serverstransport-rootcas">`serverstransport.`<br />`rootcas`</a> | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections). |  | No |
| <a id="opt-serverstransport-certificatesSecrets" href="#opt-serverstransport-certificatesSecrets" title="#opt-serverstransport-certificatesSecrets">`serverstransport.`<br />`certificatesSecrets`</a> | Certificates to present to the server for mTLS. |  | No |
| <a id="opt-serverstransport-maxIdleConnsPerHost" href="#opt-serverstransport-maxIdleConnsPerHost" title="#opt-serverstransport-maxIdleConnsPerHost">`serverstransport.`<br />`maxIdleConnsPerHost`</a> | Maximum idle (keep-alive) connections to keep per-host. | 200 | No |
| <a id="opt-serverstransport-disableHTTP2" href="#opt-serverstransport-disableHTTP2" title="#opt-serverstransport-disableHTTP2">`serverstransport.`<br />`disableHTTP2`</a> | Disables HTTP/2 for connections with servers. | false | No |
| <a id="opt-serverstransport-peerCertURI" href="#opt-serverstransport-peerCertURI" title="#opt-serverstransport-peerCertURI">`serverstransport.`<br />`peerCertURI`</a> | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| <a id="opt-serverstransport-forwardingTimeouts-dialTimeout" href="#opt-serverstransport-forwardingTimeouts-dialTimeout" title="#opt-serverstransport-forwardingTimeouts-dialTimeout">`serverstransport.`<br />`forwardingTimeouts.dialTimeout`</a> | Amount of time to wait until a connection to a server can be established.<br />Zero means no timeout. | 30s  | No |
| <a id="opt-serverstransport-forwardingTimeouts-responseHeaderTimeout" href="#opt-serverstransport-forwardingTimeouts-responseHeaderTimeout" title="#opt-serverstransport-forwardingTimeouts-responseHeaderTimeout">`serverstransport.`<br />`forwardingTimeouts.responseHeaderTimeout`</a> | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />Zero means no timeout | 0s  | No |
| <a id="opt-serverstransport-forwardingTimeouts-idleConnTimeout" href="#opt-serverstransport-forwardingTimeouts-idleConnTimeout" title="#opt-serverstransport-forwardingTimeouts-idleConnTimeout">`serverstransport.`<br />`forwardingTimeouts.idleConnTimeout`</a> | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />Zero means no timeout. | 90s  | No |
| <a id="opt-serverstransport-spiffe-ids" href="#opt-serverstransport-spiffe-ids" title="#opt-serverstransport-spiffe-ids">`serverstransport.`<br />`spiffe.ids`</a> | Allow SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. |  | No |
| <a id="opt-serverstransport-spiffe-trustDomain" href="#opt-serverstransport-spiffe-trustDomain" title="#opt-serverstransport-spiffe-trustDomain">`serverstransport.`<br />`spiffe.trustDomain`</a> | Allow SPIFFE trust domain. | ""  | No |

!!! note "CA Secret"
    The CA secret must contain a base64 encoded certificate under either a tls.ca or a ca.crt key.
