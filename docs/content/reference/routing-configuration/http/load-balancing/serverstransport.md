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
      cipherSuites: 
        - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
        - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
      minVersion: VersionTLS12
      maxVersion: VersionTLS12
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
  cipherSuites = ["TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256","TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"]
  minVersion = "VersionTLS12"
  maxVersion = "VersionTLS12"

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

| Field                                                                                                                                                                                                          | Description                                                                                                                              | Default | Required |
|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| <a id="opt-serverName" href="#opt-serverName" title="#opt-serverName">`serverName`</a> | Configures the server name that will be used as the SNI.                                                                                 | ""      | No       |
| <a id="opt-certificates" href="#opt-certificates" title="#opt-certificates">`certificates`</a> | Defines the list of certificates (as file paths, or data bytes) that will be set as client certificates for mTLS.                        | []      | No       |
| <a id="opt-insecureSkipVerify" href="#opt-insecureSkipVerify" title="#opt-insecureSkipVerify">`insecureSkipVerify`</a> | Controls whether the server's certificate chain and host name is verified.                                                               | false   | No       |
| <a id="opt-rootcas" href="#opt-rootcas" title="#opt-rootcas">`rootcas`</a> | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections).                                   | []      | No       |
| <a id="opt-cipherSuites" href="#opt-cipherSuites" title="#opt-cipherSuites">`cipherSuites`</a> | Defines the cipher suites to use when contacting backend servers. | [] | No |
| <a id="opt-minVersion" href="#opt-minVersion" title="#opt-minVersion">`minVersion`</a> | Defines the minimum TLS version to use when contacting backend servers. | "" | No |
| <a id="opt-maxVersion" href="#opt-maxVersion" title="#opt-maxVersion">`maxVersion`</a> | Defines the maximum TLS version to use when contacting backend servers. | "" | No |
| <a id="opt-maxIdleConnsPerHost" href="#opt-maxIdleConnsPerHost" title="#opt-maxIdleConnsPerHost">`maxIdleConnsPerHost`</a> | Maximum idle (keep-alive) connections to keep per-host.                                                                                  | 200     | No       |
| <a id="opt-disableHTTP2" href="#opt-disableHTTP2" title="#opt-disableHTTP2">`disableHTTP2`</a> | Disables HTTP/2 for connections with servers.                                                                                            | false   | No       |
| <a id="opt-peerCertURI" href="#opt-peerCertURI" title="#opt-peerCertURI">`peerCertURI`</a> | Defines the URI used to match against SAN URIs during the server's certificate verification.                                             | ""      | No       |
| <a id="opt-forwardingTimeouts-dialTimeout" href="#opt-forwardingTimeouts-dialTimeout" title="#opt-forwardingTimeouts-dialTimeout">`forwardingTimeouts.dialTimeout`</a> | Amount of time to wait until a connection to a server can be established.<br />0 = no timeout                                            | 30s     | No       |
| <a id="opt-forwardingTimeouts-responseHeaderTimeout" href="#opt-forwardingTimeouts-responseHeaderTimeout" title="#opt-forwardingTimeouts-responseHeaderTimeout">`forwardingTimeouts.responseHeaderTimeout`</a> | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />0 = no timeout | 0s      | No       |
| <a id="opt-forwardingTimeouts-idleConnTimeout" href="#opt-forwardingTimeouts-idleConnTimeout" title="#opt-forwardingTimeouts-idleConnTimeout">`forwardingTimeouts.idleConnTimeout`</a> | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />0 = no timeout                       | 90s     | No       |
| <a id="opt-forwardingTimeouts-readIdleTimeout" href="#opt-forwardingTimeouts-readIdleTimeout" title="#opt-forwardingTimeouts-readIdleTimeout">`forwardingTimeouts.readIdleTimeout`</a> | Defines the timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection.    | 0s      | No       |
| <a id="opt-forwardingTimeouts-pingTimeout" href="#opt-forwardingTimeouts-pingTimeout" title="#opt-forwardingTimeouts-pingTimeout">`forwardingTimeouts.pingTimeout`</a> | Defines the timeout after which the HTTP/2 connection will be closed if a response to ping is not received.                              | 15s     | No       |
| <a id="opt-spiffe" href="#opt-spiffe" title="#opt-spiffe">`spiffe`</a> | Defines the SPIFFE configuration. An empty `spiffe` section enables SPIFFE (that allows any SPIFFE ID).                                  |         | No       |
| <a id="opt-spiffe-ids" href="#opt-spiffe-ids" title="#opt-spiffe-ids">`spiffe.ids`</a> | Defines the allowed SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain.                                                  | []      | No       |
| <a id="opt-spiffe-trustDomain" href="#opt-spiffe-trustDomain" title="#opt-spiffe-trustDomain">`spiffe.trustDomain`</a> | Defines the SPIFFE trust domain.                                                                                                         | ""      | No       |
