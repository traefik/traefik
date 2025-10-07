---
title: "ServersTransport TCP"
description: "The ServersTransport allows configuring the connection between Traefik and the TCP servers in Kubernetes."
---

ServersTransport allows to configure the transport between Traefik and your TCP servers.

## Configuration Example

Declare the serversTransport:

```yaml tab="Structured (YAML)"
tcp:
  serversTransports:
    mytransport:
      dialTimeout: "30s"
      dialKeepAlive: "20s"
      terminationDelay: "200ms"
      tls:
        serverName: "example.com"
        certificates:
          - "/path/to/cert1.pem"
          - "/path/to/cert2.pem"
        insecureSkipVerify: true
        rootcas:
          - "/path/to/rootca.pem"
        peerCertURI: "spiffe://example.org/peer"
      spiffe:
        ids:
          - "spiffe://example.org/id1"
          - "spiffe://example.org/id2"
        trustDomain: "example.org"
```

```toml tab="Structured (TOML)"
[tcp.serversTransports.mytransport]
  dialTimeout = "30s"
  dialKeepAlive = "20s"
  terminationDelay = "200ms"

  [tcp.serversTransports.mytransport.tls]
    serverName = "example.com"
    certificates = ["/path/to/cert1.pem", "/path/to/cert2.pem"]
    insecureSkipVerify = true
    rootcas = ["/path/to/rootca.pem"]
    peerCertURI = "spiffe://example.org/peer"

  [tcp.serversTransports.mytransport.spiffe]
    ids = ["spiffe://example.org/id1", "spiffe://example.org/id2"]
    trustDomain = "example.org"
```

Attach the serversTransport to a service:

```yaml tab="Structured (YAML)"
tcp:
  services:
    Service01:
      loadBalancer:
        serversTransport: mytransport
```

```toml tab="Structured(TOML)"
## Dynamic configuration
[tcp.services]
  [tcp.services.Service01]
    [tcp.services.Service01.loadBalancer]
      serversTransport = "mytransport"
```

```yaml tab="Labels"
labels:
  - "traefik.tcp.services.Service01.loadBalancer.serversTransport=mytransport"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.tcp.services.Service01.loadBalancer.serversTransport=mytransport"
  ]
}
```

## Configuration Options

| Field                                                                                                                                                                                                                      | Description                                                                                                                                                                                                        | Default | Required |
|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="opt-serverstransport-dialTimeout" href="#opt-serverstransport-dialTimeout" title="#opt-serverstransport-dialTimeout">`serverstransport.`<br />`dialTimeout`</a> | Defines the timeout when dialing the backend TCP service. If zero, no timeout exists.                                                                                                                              | 30s     | No       |
| <a id="opt-serverstransport-dialKeepAlive" href="#opt-serverstransport-dialKeepAlive" title="#opt-serverstransport-dialKeepAlive">`serverstransport.`<br />`dialKeepAlive`</a> | Defines the interval between keep-alive probes for an active network connection.                                                                                                                                   | 15s     | No       |
| <a id="opt-serverstransport-terminationDelay" href="#opt-serverstransport-terminationDelay" title="#opt-serverstransport-terminationDelay">`serverstransport.`<br />`terminationDelay`</a> | Sets the time limit for the proxy to fully terminate connections on both sides after initiating the termination sequence, with a negative value indicating no deadline. More Information [here](#terminationdelay) | 100ms   | No       |
| <a id="opt-serverstransport-proxyProtocol" href="#opt-serverstransport-proxyProtocol" title="#opt-serverstransport-proxyProtocol">`serverstransport.`<br />`proxyProtocol`</a> | Defines the Proxy Protocol configuration. An empty `proxyProtocol` section enables Proxy Protocol version 2.                                                                                                       |         | No       |
| <a id="opt-serverstransport-proxyProtocol-version" href="#opt-serverstransport-proxyProtocol-version" title="#opt-serverstransport-proxyProtocol-version">`serverstransport.`<br />`proxyProtocol.version`</a> | Traefik supports PROXY Protocol version 1 and 2 on TCP Services. More Information [here](#proxyprotocolversion)                                                                                                    | 2       | No       |
| <a id="opt-serverstransport-tls" href="#opt-serverstransport-tls" title="#opt-serverstransport-tls">`serverstransport.`<br />`tls`</a> | Defines the TLS configuration. An empty `tls` section enables TLS.                                                                                                                                                 |         | No       |
| <a id="opt-serverstransport-tls-serverName" href="#opt-serverstransport-tls-serverName" title="#opt-serverstransport-tls-serverName">`serverstransport.`<br />`tls`<br />`.serverName`</a> | Configures the server name that will be used for SNI.                                                                                                                                                              |         | No       |
| <a id="opt-serverstransport-tls-certificates" href="#opt-serverstransport-tls-certificates" title="#opt-serverstransport-tls-certificates">`serverstransport.`<br />`tls`<br />`.certificates`</a> | Defines the list of certificates (as file paths, or data bytes) that will be set as client certificates for mTLS.                                                                                                  |         | No       |
| <a id="opt-serverstransport-tls-insecureSkipVerify" href="#opt-serverstransport-tls-insecureSkipVerify" title="#opt-serverstransport-tls-insecureSkipVerify">`serverstransport.`<br />`tls`<br />`.insecureSkipVerify`</a> | Controls whether the server's certificate chain and host name is verified.                                                                                                                                         | false   | No       |
| <a id="opt-serverstransport-tls-rootcas" href="#opt-serverstransport-tls-rootcas" title="#opt-serverstransport-tls-rootcas">`serverstransport.`<br />`tls`<br />`.rootcas`</a> | Defines the root certificate authorities to use when verifying server certificates. (for mTLS connections).                                                                                                        |         | No       |
| <a id="opt-serverstransport-tls-peerCertURI" href="#opt-serverstransport-tls-peerCertURI" title="#opt-serverstransport-tls-peerCertURI">`serverstransport.`<br />`tls.`<br />`peerCertURI`</a> | Defines the URI used to match against SAN URIs during the server's certificate verification.                                                                                                                       | false   | No       |
| <a id="opt-serverstransport-spiffe" href="#opt-serverstransport-spiffe" title="#opt-serverstransport-spiffe">`serverstransport.`<br />`spiffe`</a> | Defines the SPIFFE configuration. An empty `spiffe` section enables SPIFFE (that allows any SPIFFE ID).                                                                                                            |         | No       |
| <a id="opt-serverstransport-spiffe-ids" href="#opt-serverstransport-spiffe-ids" title="#opt-serverstransport-spiffe-ids">`serverstransport.`<br />`spiffe`<br />`.ids`</a> | Allow SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain.                                                                                                                                          |         | No       |
| <a id="opt-serverstransport-spiffe-trustDomain" href="#opt-serverstransport-spiffe-trustDomain" title="#opt-serverstransport-spiffe-trustDomain">`serverstransport.`<br />`spiffe`<br />`.trustDomain`</a> | Allow SPIFFE trust domain.                                                                                                                                                                                         | ""      | No       |

!!! note "SPIFFE"

    Please note that SPIFFE must be enabled in the [install configuration](../../install-configuration/tls/spiffe.md) (formerly known as static configuration) before using it to secure the connection between Traefik and the backends.

### `terminationDelay`

As a proxy between a client and a server, it can happen that either side (e.g. client side) decides to terminate its writing capability on the connection (i.e. issuance of a FIN packet).
The proxy needs to propagate that intent to the other side, and so when that happens, it also does the same on its connection with the other side (e.g. backend side).

However, if for some reason (bad implementation, or malicious intent) the other side does not eventually do the same as well,
the connection would stay half-open, which would lock resources for however long.

To that end, as soon as the proxy enters this termination sequence, it sets a deadline on fully terminating the connections on both sides.

The termination delay controls that deadline.
A negative value means an infinite deadline (i.e. the connection is never fully terminated by the proxy itself).

### `proxyProtocol.version`

Traefik supports [PROXY Protocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2 on TCP Services.
It can be configured by setting `proxyProtocol.version` on the serversTransport.
The option specifies the version of the protocol to be used. Either 1 or 2.
