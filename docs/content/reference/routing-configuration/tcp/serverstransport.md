---
title: "ServersTransport TCP"
description: "The ServersTransport allows configuring the connection between Traefik and the TCP servers in Kubernetes."
---

ServersTransport allows to configure the transport between Traefik and your TCP servers.

## Configuration Example

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      dialTimeout: 30s
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport]
  dialTimeout = "30s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  dialTimeout: 30s
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `serverstransport.`<br />`dialTimeout` | Defines the timeout when dialing the backend TCP service. If zero, no timeout exists.  | 30s | No |
| `serverstransport.`<br />`dialKeepAlive` | Defines the interval between keep-alive probes for an active network connection. More Information [here](#dialkeepalive).  | 15s | No |
| `serverstransport.`<br />`terminationDelay` | Sets the time limit for the proxy to fully terminate connections on both sides after initiating the termination sequence, with a negative value indicating no deadline. More Information [here](#terminationdelay) | 100ms | No |
| `serverstransport.`<br />`tls` | Defines the TLS configuration. An empty `tls` section enables TLS. |  | No |
| `serverstransport.`<br />`tls`<br />`.serverName` | Configures the server name that will be used for SNI. |  | No |
| `serverstransport.`<br />`tls`<br />`.certificates` | Defines the list of certificates (as file paths, or data bytes) that will be set as client certificates for mTLS. |  | No |
| `serverstransport.`<br />`tls`<br />`.insecureSkipVerify` | Controls whether the server's certificate chain and host name is verified. | false  | No |
| `serverstransport.`<br />`tls`<br />`.rootcas` | Defines the root certificate authorities to use when verifying server certificates. (for mTLS connections). |  | No |
| `serverstransport.`<br />`tls.`<br />`peerCertURI` | Defines the URI used to match against SAN URIs during the server's certificate verification.  | false | No |
| `serverstransport.`<br />`spiffe`<br />`.ids` | Allow SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. |  | No |
| `serverstransport.`<br />`spiffe`<br />`.trustDomain` | Allow SPIFFE trust domain. | ""  | No |

!!! note "SPIFFE"

    Please note that SPIFFE must be enabled in the [static configuration](../../install-configuration/tls/spiffe.md) before using it to secure the connection between Traefik and the backends.

### dialKeepAlive

`dialKeepAlive` defines the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      dialKeepAlive: 30s
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport]
  dialKeepAlive = "30s"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  dialKeepAlive: 30s
```

### `terminationDelay`

As a proxy between a client and a server, it can happen that either side (e.g. client side) decides to terminate its writing capability on the connection (i.e. issuance of a FIN packet).
The proxy needs to propagate that intent to the other side, and so when that happens, it also does the same on its connection with the other side (e.g. backend side).

However, if for some reason (bad implementation, or malicious intent) the other side does not eventually do the same as well,
the connection would stay half-open, which would lock resources for however long.

To that end, as soon as the proxy enters this termination sequence, it sets a deadline on fully terminating the connections on both sides.

The termination delay controls that deadline.
A negative value means an infinite deadline (i.e. the connection is never fully terminated by the proxy itself).

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  serversTransports:
    mytransport:
      terminationDelay: 100ms
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.serversTransports.mytransport]
  terminationDelay = "100ms"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransportTCP
metadata:
  name: mytransport
  namespace: default

spec:
  terminationDelay: 100ms
```
