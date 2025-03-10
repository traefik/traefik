---
title: "ServersTransport"
description: "ServersTransport allows configuring the connection between Traefik and the HTTP servers."
---

ServersTransport allows you to configure the transport between Traefik and your HTTP servers.

## Configuration Example

Declare the serversTransport:

```yaml tab="Structured (YAML)"
http:
  serversTransports:
    mytransport:
      serverName: "myhost"
      certificates:
        - "/path/to/cert1.pem"
        - "/path/to/cert2.pem"
      insecureSkipVerify: true
      rootcas:
        - "/path/to/rootca1.pem"
        - "/path/to/rootca2.pem"
      maxIdleConnsPerHost: 100
      disableHTTP2: true
      peerCertURI: "spiffe://example.org/peer"
      forwardingTimeouts:
        dialTimeout: "30s"
        responseHeaderTimeout: "10s"
        idleConnTimeout: "60s"
        readIdleTimeout: "5s"
        pingTimeout: "15s"
      spiffe:
        ids:
          - "spiffe://example.org/id1"
          - "spiffe://example.org/id2"
        trustDomain: "example.org"
```

```toml tab="Structured (TOML)"
[http.serversTransports.mytransport]
  serverName = "myhost"
  certificates = ["/path/to/cert1.pem", "/path/to/cert2.pem"]
  insecureSkipVerify = true
  rootcas = ["/path/to/rootca1.pem", "/path/to/rootca2.pem"]
  maxIdleConnsPerHost = 100
  disableHTTP2 = true
  peerCertURI = "spiffe://example.org/peer"

  [http.serversTransports.mytransport.forwardingTimeouts]
    dialTimeout = "30s"
    responseHeaderTimeout = "10s"
    idleConnTimeout = "60s"
    readIdleTimeout = "5s"
    pingTimeout = "15s"

  [http.serversTransports.mytransport.spiffe]
    ids = ["spiffe://example.org/id1", "spiffe://example.org/id2"]
    trustDomain = "example.org"
``` 

Attach the serversTransport to a service:

```yaml tab="Structured (YAML)"
## Dynamic configuration
http:
  services:
    Service01:
      loadBalancer:
        serversTransport: mytransport
```

```toml tab="Structured(TOML)"
## Dynamic configuration
[http.services]
  [http.services.Service01]
    [http.services.Service01.loadBalancer]
      serversTransport = "mytransport"
```

```yaml tab="Labels"
labels:
  - "traefik.http.services.Service01.loadBalancer.serversTransport=mytransport"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.services.Service01.loadBalancer.serversTransport=mytransport"
  ]
}
```

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `serverName` | Configures the server name that will be used as the SNI. | "" | No |
| `certificates` | Defines the list of certificates (as file paths, or data bytes) that will be set as client certificates for mTLS. | [] | No |
| `insecureSkipVerify` | Controls whether the server's certificate chain and host name is verified. | false  | No |
| `rootcas` | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections). | [] | No |
| `maxIdleConnsPerHost` | Maximum idle (keep-alive) connections to keep per-host. | 200 | No |
| `disableHTTP2` | Disables HTTP/2 for connections with servers. | false | No |
| `peerCertURI` | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| `forwardingTimeouts.dialTimeout` | Amount of time to wait until a connection to a server can be established.<br />0 = no timeout | 30s  | No |
| `forwardingTimeouts.responseHeaderTimeout` | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />0 = no timeout | 0s  | No |
| `forwardingTimeouts.idleConnTimeout` | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />0 = no timeout | 90s  | No |
| `forwardingTimeouts.readIdleTimeout` | Defines the timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection.  | 0s  | No |
| `forwardingTimeouts.pingTimeout` | Defines the timeout after which the HTTP/2 connection will be closed if a response to ping is not received. | 15s  | No |
| `spiffe.ids` | Defines the allowed SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. | []  | No |
| `spiffe.trustDomain` | Defines the SPIFFE trust domain. | ""  | No |
