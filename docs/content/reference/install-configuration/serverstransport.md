---
title: "Traefik Default ServersTransport Documentation"
description: "ServersTransport allows configuring the connection between Traefik and the HTTP servers."
---

Defining the default transport between Traefik and your HTTP or TCP servers.
{: .subtitle }

## HTTP ServersTransport

This ServersTransport is applied to every [HTTP service](../routing-configuration/http/load-balancing/service.md#opt-serversTransport) with no explicit [ServersTransport](../routing-configuration/http/load-balancing/serverstransport.md).

### Configuration Example

```yaml tab="File (YAML)"
serverstransport:
  forwardingTimeouts:
    dialTimeout: "30s"
    responseHeaderTimeout: "10s"
  rootcas:
  - "/path/to/rootca1.pem"
  - "/path/to/rootca2.pem"
  maxIdleConnsPerHost: 100
```

```toml tab="File (TOML)"
[serverstransport]
  [serverstransport.forwardingTimeouts]
    dialTimeout = "30s"
    responseHeaderTimeout= "10s"
  rootcas = ["/path/to/rootca1.pem", "/path/to/rootca2.pem"]
  maxIdleConnsPerHost = 100
```

```yaml tab="Helm Chart Values"
## Values file
additionalArguments:
  - --serverstransport.forwardingTimeouts.dialTimeout = "30s"
  - --serverstransport.forwardingTimeouts.responseHeaderTimeout= "10s"
  - --serverstransport.rootcas = ["/path/to/rootca1.pem", "/path/to/rootca2.pem"]
  - --serverstransport.maxIdleConnsPerHost = 100
```

### Configuration Options

| Field                                                           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Default                 | Required |
|:----------------------------------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------------|:---------|
| <a id="opt-serverstransport-forwardingtimeouts-dialtimeout" href="#opt-serverstransport-forwardingtimeouts-dialtimeout" title="#opt-serverstransport-forwardingtimeouts-dialtimeout">serverstransport.forwardingtimeouts.dialtimeout</a> | The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists. | 30 |
| <a id="opt-serverstransport-forwardingtimeouts-idleconntimeout" href="#opt-serverstransport-forwardingtimeouts-idleconntimeout" title="#opt-serverstransport-forwardingtimeouts-idleconntimeout">serverstransport.forwardingtimeouts.idleconntimeout</a> | The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself | 90 |
| <a id="opt-serverstransport-forwardingtimeouts-responseheadertimeout" href="#opt-serverstransport-forwardingtimeouts-responseheadertimeout" title="#opt-serverstransport-forwardingtimeouts-responseheadertimeout">serverstransport.forwardingtimeouts.responseheadertimeout</a> | The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists. | 0 |
| <a id="opt-serverstransport-insecureskipverify" href="#opt-serverstransport-insecureskipverify" title="#opt-serverstransport-insecureskipverify">serverstransport.insecureskipverify</a> | Disable SSL certificate verification. | false |
| <a id="opt-serverstransport-maxidleconnsperhost" href="#opt-serverstransport-maxidleconnsperhost" title="#opt-serverstransport-maxidleconnsperhost">serverstransport.maxidleconnsperhost</a> | If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used. If negative, disables connection reuse. | 200 |
| <a id="opt-serverstransport-rootcas" href="#opt-serverstransport-rootcas" title="#opt-serverstransport-rootcas">serverstransport.rootcas</a> | Add cert file for self-signed certificate. | |
| <a id="opt-serverstransport-spiffe" href="#opt-serverstransport-spiffe" title="#opt-serverstransport-spiffe">serverstransport.spiffe</a> | Defines the SPIFFE configuration. | false |
| <a id="opt-serverstransport-spiffe-ids" href="#opt-serverstransport-spiffe-ids" title="#opt-serverstransport-spiffe-ids">serverstransport.spiffe.ids</a> | Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain). | |
| <a id="opt-serverstransport-spiffe-trustdomain" href="#opt-serverstransport-spiffe-trustdomain" title="#opt-serverstransport-spiffe-trustdomain">serverstransport.spiffe.trustdomain</a> | Defines the allowed SPIFFE trust domain. | |
| <a id="opt-spiffe-workloadapiaddr" href="#opt-spiffe-workloadapiaddr" title="#opt-spiffe-workloadapiaddr">spiffe.workloadapiaddr</a> | Defines the workload API address. | |

## TCP ServersTransport

This ServersTransport is applied to every [TCP service](../routing-configuration/tcp/service.md#opt-serversTransport) with no explicit [ServersTransport](../routing-configuration/tcp/serverstransport.md).

### Configuration Example

```yaml tab="File (YAML)"
tcpserverstransport:
  dialkeepalive: "30s"
  dialtimeout: "10s"
```

```toml tab="File (TOML)"
[tcpserverstransport]
  dialkeepalive = "30s"
  dialtimeout= "10s"
```

```yaml tab="Helm Chart Values"
## Values file
additionalArguments:
  - --tcpserverstransport.dialkeepalive = "30s"
  - --tcpserverstransport.dialtimeout= "10s"
```

### Configuration Options

| Field                                                           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Default                 | Required |
|:----------------------------------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------------|:---------|
| <a id="opt-tcpserverstransport-dialkeepalive" href="#opt-tcpserverstransport-dialkeepalive" title="#opt-tcpserverstransport-dialkeepalive">tcpserverstransport.dialkeepalive</a> | Defines the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled | 15 |
| <a id="opt-tcpserverstransport-dialtimeout" href="#opt-tcpserverstransport-dialtimeout" title="#opt-tcpserverstransport-dialtimeout">tcpserverstransport.dialtimeout</a> | Defines the amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists. | 30 |
| <a id="opt-tcpserverstransport-terminationdelay" href="#opt-tcpserverstransport-terminationdelay" title="#opt-tcpserverstransport-terminationdelay">tcpserverstransport.terminationdelay</a> | Defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability. | 0 |
| <a id="opt-tcpserverstransport-tls" href="#opt-tcpserverstransport-tls" title="#opt-tcpserverstransport-tls">tcpserverstransport.tls</a> | Defines the TLS configuration. | false |
| <a id="opt-tcpserverstransport-tls-insecureskipverify" href="#opt-tcpserverstransport-tls-insecureskipverify" title="#opt-tcpserverstransport-tls-insecureskipverify">tcpserverstransport.tls.insecureskipverify</a> | Disables SSL certificate verification. | false |
| <a id="opt-tcpserverstransport-tls-rootcas" href="#opt-tcpserverstransport-tls-rootcas" title="#opt-tcpserverstransport-tls-rootcas">tcpserverstransport.tls.rootcas</a> | Defines a list of CA secret used to validate self-signed certificate | |
| <a id="opt-tcpserverstransport-tls-spiffe" href="#opt-tcpserverstransport-tls-spiffe" title="#opt-tcpserverstransport-tls-spiffe">tcpserverstransport.tls.spiffe</a> | Defines the SPIFFE TLS configuration. | false |
| <a id="opt-tcpserverstransport-tls-spiffe-ids" href="#opt-tcpserverstransport-tls-spiffe-ids" title="#opt-tcpserverstransport-tls-spiffe-ids">tcpserverstransport.tls.spiffe.ids</a> | Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain). | |
| <a id="opt-tcpserverstransport-tls-spiffe-trustdomain" href="#opt-tcpserverstransport-tls-spiffe-trustdomain" title="#opt-tcpserverstransport-tls-spiffe-trustdomain">tcpserverstransport.tls.spiffe.trustdomain</a> | Defines the allowed SPIFFE trust domain. | |
