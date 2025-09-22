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
| <a id="serverstransport-serverName" href="#serverstransport-serverName" title="#serverstransport-serverName">`serverstransport.`<br />`serverName`</a> | Defines the server name that will be used for SNI. |  | No |
| <a id="serverstransport-insecureSkipVerify" href="#serverstransport-insecureSkipVerify" title="#serverstransport-insecureSkipVerify">`serverstransport.`<br />`insecureSkipVerify`</a> | Controls whether the server's certificate chain and host name is verified. | false  | No |
| <a id="serverstransport-rootcas" href="#serverstransport-rootcas" title="#serverstransport-rootcas">`serverstransport.`<br />`rootcas`</a> | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections). |  | No |
| <a id="serverstransport-certificatesSecrets" href="#serverstransport-certificatesSecrets" title="#serverstransport-certificatesSecrets">`serverstransport.`<br />`certificatesSecrets`</a> | Certificates to present to the server for mTLS. |  | No |
| <a id="serverstransport-maxIdleConnsPerHost" href="#serverstransport-maxIdleConnsPerHost" title="#serverstransport-maxIdleConnsPerHost">`serverstransport.`<br />`maxIdleConnsPerHost`</a> | Maximum idle (keep-alive) connections to keep per-host. | 200 | No |
| <a id="serverstransport-disableHTTP2" href="#serverstransport-disableHTTP2" title="#serverstransport-disableHTTP2">`serverstransport.`<br />`disableHTTP2`</a> | Disables HTTP/2 for connections with servers. | false | No |
| <a id="serverstransport-peerCertURI" href="#serverstransport-peerCertURI" title="#serverstransport-peerCertURI">`serverstransport.`<br />`peerCertURI`</a> | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| <a id="serverstransport-forwardingTimeouts-dialTimeout" href="#serverstransport-forwardingTimeouts-dialTimeout" title="#serverstransport-forwardingTimeouts-dialTimeout">`serverstransport.`<br />`forwardingTimeouts.dialTimeout`</a> | Amount of time to wait until a connection to a server can be established.<br />Zero means no timeout. | 30s  | No |
| <a id="serverstransport-forwardingTimeouts-responseHeaderTimeout" href="#serverstransport-forwardingTimeouts-responseHeaderTimeout" title="#serverstransport-forwardingTimeouts-responseHeaderTimeout">`serverstransport.`<br />`forwardingTimeouts.responseHeaderTimeout`</a> | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />Zero means no timeout | 0s  | No |
| <a id="serverstransport-forwardingTimeouts-idleConnTimeout" href="#serverstransport-forwardingTimeouts-idleConnTimeout" title="#serverstransport-forwardingTimeouts-idleConnTimeout">`serverstransport.`<br />`forwardingTimeouts.idleConnTimeout`</a> | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />Zero means no timeout. | 90s  | No |
| <a id="serverstransport-spiffe-ids" href="#serverstransport-spiffe-ids" title="#serverstransport-spiffe-ids">`serverstransport.`<br />`spiffe.ids`</a> | Allow SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. |  | No |
| <a id="serverstransport-spiffe-trustDomain" href="#serverstransport-spiffe-trustDomain" title="#serverstransport-spiffe-trustDomain">`serverstransport.`<br />`spiffe.trustDomain`</a> | Allow SPIFFE trust domain. | ""  | No |

!!! note "CA Secret"
    The CA secret must contain a base64 encoded certificate under either a tls.ca or a ca.crt key.
