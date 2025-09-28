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
| <a id="serverName" href="#serverName" title="#serverName">`serverName`</a> | Configures the server name that will be used as the SNI. | "" | No |
| <a id="certificates" href="#certificates" title="#certificates">`certificates`</a> | Defines the list of certificates (as file paths, or data bytes) that will be set as client certificates for mTLS. | [] | No |
| <a id="insecureSkipVerify" href="#insecureSkipVerify" title="#insecureSkipVerify">`insecureSkipVerify`</a> | Controls whether the server's certificate chain and host name is verified. | false  | No |
| <a id="rootcas" href="#rootcas" title="#rootcas">`rootcas`</a> | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections). | [] | No |
| <a id="maxIdleConnsPerHost" href="#maxIdleConnsPerHost" title="#maxIdleConnsPerHost">`maxIdleConnsPerHost`</a> | Maximum idle (keep-alive) connections to keep per-host. | 200 | No |
| <a id="disableHTTP2" href="#disableHTTP2" title="#disableHTTP2">`disableHTTP2`</a> | Disables HTTP/2 for connections with servers. | false | No |
| <a id="peerCertURI" href="#peerCertURI" title="#peerCertURI">`peerCertURI`</a> | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| <a id="forwardingTimeouts-dialTimeout" href="#forwardingTimeouts-dialTimeout" title="#forwardingTimeouts-dialTimeout">`forwardingTimeouts.dialTimeout`</a> | Amount of time to wait until a connection to a server can be established.<br />0 = no timeout | 30s  | No |
| <a id="forwardingTimeouts-responseHeaderTimeout" href="#forwardingTimeouts-responseHeaderTimeout" title="#forwardingTimeouts-responseHeaderTimeout">`forwardingTimeouts.responseHeaderTimeout`</a> | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />0 = no timeout | 0s  | No |
| <a id="forwardingTimeouts-idleConnTimeout" href="#forwardingTimeouts-idleConnTimeout" title="#forwardingTimeouts-idleConnTimeout">`forwardingTimeouts.idleConnTimeout`</a> | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />0 = no timeout | 90s  | No |
| <a id="forwardingTimeouts-readIdleTimeout" href="#forwardingTimeouts-readIdleTimeout" title="#forwardingTimeouts-readIdleTimeout">`forwardingTimeouts.readIdleTimeout`</a> | Defines the timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection.  | 0s  | No |
| <a id="forwardingTimeouts-pingTimeout" href="#forwardingTimeouts-pingTimeout" title="#forwardingTimeouts-pingTimeout">`forwardingTimeouts.pingTimeout`</a> | Defines the timeout after which the HTTP/2 connection will be closed if a response to ping is not received. | 15s  | No |
| <a id="spiffe-ids" href="#spiffe-ids" title="#spiffe-ids">`spiffe.ids`</a> | Defines the allowed SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. | []  | No |
| <a id="spiffe-trustDomain" href="#spiffe-trustDomain" title="#spiffe-trustDomain">`spiffe.trustDomain`</a> | Defines the SPIFFE trust domain. | ""  | No |
