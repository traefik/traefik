---
title: "ServersTransport"
description: "ServersTransport allows configuring the connection between Traefik and the HTTP servers."
---

ServersTransport allows you to configure the transport between Traefik and your HTTP servers.

## Configuration Example

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  serversTransports:
    mytransport:
      serverName: "myhost"
      # ....
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.serversTransports.mytransport]
  serverName = "myhost"
  # ....
``` 

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `serverstransport.`<br />`serverName` | Configures the server name that will be used as the SNI. | "" | No |
| `serverstransport.`<br />`certificates` | Defines the list of certificates (as file paths, or data bytes) that will be set as client certificates for mTLS. | [] | No |
| `serverstransport.`<br />`insecureSkipVerify` | Controls whether the server's certificate chain and host name is verified. | false  | No |
| `serverstransport.`<br />`rootcas` | Set of root certificate authorities to use when verifying server certificates. (for mTLS connections). | [] | No |
| `serverstransport.`<br />`maxIdleConnsPerHost` | Maximum idle (keep-alive) connections to keep per-host. | 200 | No |
| `serverstransport.`<br />`disableHTTP2` | Disables HTTP/2 for connections with servers. | false | No |
| `serverstransport.`<br />`peerCertURI` | Defines the URI used to match against SAN URIs during the server's certificate verification. | "" | No |
| `serverstransport.`<br />`forwardingTimeouts.dialTimeout` | Amount of time to wait until a connection to a server can be established.<br />0 = no timeout | 30s  | No |
| `serverstransport.`<br />`forwardingTimeouts.responseHeaderTimeout` | Amount of time to wait for a server's response headers after fully writing the request (including its body, if any).<br />0 = no timeout | 0s  | No |
| `serverstransport.`<br />`forwardingTimeouts.idleConnTimeout` | Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.<br />0 = no timeout | 90s  | No |
| `serverstransport.`<br />`forwardingTimeouts.readIdleTimeout` | Defines the timeout after which a health check using ping frame will be carried out if no frame is received on the HTTP/2 connection.  | 0s  | No |
| `serverstransport.`<br />`forwardingTimeouts.pingTimeout` | Defines the timeout after which the HTTP/2 connection will be closed if a response to ping is not received. | 15s  | No |
| `serverstransport.`<br />`spiffe.ids` | Defines the allowed SPIFFE IDs.<br />This takes precedence over the SPIFFE TrustDomain. |  | No |
| `serverstransport.`<br />`spiffe.trustDomain` | Defines the SPIFFE trust domain. | []  | No |
