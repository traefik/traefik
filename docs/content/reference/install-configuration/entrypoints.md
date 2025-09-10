---
title: "Traefik EntryPoints Documentation"
description: "For routing and load balancing in Traefik Proxy, EntryPoints define which port will receive packets and whether they are TCP or UDP. Read the technical documentation."
---

Listening for Incoming Connections/Requests
{: .subtitle }

### Configuration Example

```yaml tab="File (YAML)"
entryPoints:
  web:
    address: :80
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
          permanent: true
    observability:
      accessLogs: false
      metrics: false
      tracing: false

  websecure:
    address: :443
    http:
      tls: {}
      middlewares:
        - auth@kubernetescrd
        - strip@kubernetescrd
```

```toml tab="File (TOML)"
[entryPoints]
  [entryPoints.web]
    address = ":80"
    [entryPoints.web.http]
      [entryPoints.web.http.redirections]
        entryPoint = "websecure"
        scheme = "https"
        permanent = true
    [entryPoints.web.observability]
      accessLogs = false
      metrics = false
      tracing = false
      
  [entryPoints.websecure]
    address = ":443"
    [entryPoints.websecure.tls]
    [entryPoints.websecure.middlewares]
      - auth@kubernetescrd
      - strip@kubernetescrd
```

```yaml tab="Helm Chart Values"
## Values file
ports:
  web:
    port: :80
  websecure:
    port: :443
    tls:
      enabled: true
    middlewares:
      - auth@kubernetescrd
      - strip@kubernetescrd
additionalArguments:
  - --entryPoints.web.http.redirections.to=websecure
  - --entryPoints.web.http.redirections.scheme=https
  - --entryPoints.web.http.redirections.permanent=true
  - --entryPoints.web.observability.accessLogs=false
  - --entryPoints.web.observability.metrics=false
  - --entryPoints.web.observability.tracing=false
```

!!! tip 

      In the Helm Chart, the entryPoints `web` (port 80), `websecure` (port 443), `traefik` (port 8080) and `metrics` (port 9100) are created by default.
      The entryPoints `web`, `websecure` are exposed by default using a Service.

      The default behaviors can be overridden in the Helm Chart.

## Configuration Options

| Field                                                           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | Default                 | Required |
|:----------------------------------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------------|:---------|
| <a id="address" href="#address" title="#address">`address`</a> | Define the port, and optionally the hostname, on which to listen for incoming connections and packets.<br /> It also defines the protocol to use (TCP or UDP).<br /> If no protocol is specified, the default is TCP. The format is:`[host]:port[/tcp\|/udp]                                                                                                                                                                                                                                                                                                                                                                                                                        | -                       | Yes      |
| <a id="asDefault" href="#asDefault" title="#asDefault">`asDefault`</a> | Mark the `entryPoint` to be in the list of default `entryPoints`.<br /> `entryPoints`in this list are used (by default) on HTTP and TCP routers that do not define their own `entryPoints` option.<br /> More information [here](#asdefault).                                                                                                                                                                                                                                                                                                                                                                                                                                       | false                   | No       |
| <a id="forwardedHeaders-trustedIPs" href="#forwardedHeaders-trustedIPs" title="#forwardedHeaders-trustedIPs">`forwardedHeaders.trustedIPs`</a> | Set the IPs or CIDR from where Traefik trusts the forwarded headers information (`X-Forwarded-*`).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  | -                       | No       |
| <a id="forwardedHeaders-insecure" href="#forwardedHeaders-insecure" title="#forwardedHeaders-insecure">`forwardedHeaders.insecure`</a> | Set the insecure mode to always trust the forwarded headers information (`X-Forwarded-*`).<br />We recommend to use this option only for tests purposes, not in production.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | false                   | No       |
| <a id="http-redirections-entryPoint-to" href="#http-redirections-entryPoint-to" title="#http-redirections-entryPoint-to">`http.redirections.`<br />`entryPoint.to`</a> | The target element to enable (permanent) redirecting of all incoming requests on an entry point to another one. <br /> The target element can be an entry point name (ex: `websecure`), or a port (`:443`).                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | -                       | Yes      |
| <a id="http-redirections-entryPoint-scheme" href="#http-redirections-entryPoint-scheme" title="#http-redirections-entryPoint-scheme">`http.redirections.`<br />`entryPoint.scheme`</a> | The target scheme to use for (permanent) redirection of all incoming requests.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | https                   | No       |
| <a id="http-redirections-entryPoint-permanent" href="#http-redirections-entryPoint-permanent" title="#http-redirections-entryPoint-permanent">`http.redirections.`<br />`entryPoint.permanent`</a> | Enable permanent redirecting of all incoming requests on an entry point to another one changing the scheme. <br /> The target element, it can be an entry point name (ex: `websecure`), or a port (`:443`).                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | false                   | No       |
| <a id="http-redirections-entryPoint-priority" href="#http-redirections-entryPoint-priority" title="#http-redirections-entryPoint-priority">`http.redirections.`<br />`entryPoint.priority`</a> | Default priority applied to the routers attached to the `entryPoint`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | MaxInt32-1 (2147483646) | No       |
| <a id="http-encodeQuerySemicolons" href="#http-encodeQuerySemicolons" title="#http-encodeQuerySemicolons">`http.encodeQuerySemicolons`</a> | Enable query semicolons encoding. <br /> Use this option to avoid non-encoded semicolons to be interpreted as query parameter separators by Traefik. <br /> When using this option, the non-encoded semicolons characters in query will be transmitted encoded to the backend.<br /> More information [here](#encodequerysemicolons).                                                                                                                                                                                                                                                                                                                                               | false                   | No       |
| <a id="http-sanitizePath" href="#http-sanitizePath" title="#http-sanitizePath">`http.sanitizePath`</a> | Defines whether to enable the request path sanitization.<br /> More information [here](#sanitizepath).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | false                   | No       |
| <a id="http-middlewares" href="#http-middlewares" title="#http-middlewares">`http.middlewares`</a> | Set the list of middlewares that are prepended by default to the list of middlewares of each router associated to the named entry point. <br />More information [here](#httpmiddlewares).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | -                       | No       |
| <a id="http-tls" href="#http-tls" title="#http-tls">`http.tls`</a> | Enable TLS on every router attached to the `entryPoint`. <br /> If no certificate are set, a default self-signed certificate is generates by Traefik. <br /> We recommend to not use self signed certificates in production.                                                                                                                                                                                                                                                                                                                                                                                                                                                        | -                       | No       |
| <a id="http-tls-options" href="#http-tls-options" title="#http-tls-options">`http.tls.options`</a> | Apply TLS options on every router attached to the `entryPoint`. <br /> The TLS options can be overidden per router. <br /> More information in the [dedicated section](../../routing/providers/kubernetes-crd.md#kind-tlsoption).                                                                                                                                                                                                                                                                                                                                                                                                                                                   | -                       | No       |
| <a id="http-tls-certResolver" href="#http-tls-certResolver" title="#http-tls-certResolver">`http.tls.certResolver`</a> | Apply a certificate resolver on every router attached to the `entryPoint`. <br /> The TLS options can be overidden per router. <br /> More information in the [dedicated section](../install-configuration/tls/certificate-resolvers/overview.md).                                                                                                                                                                                                                                                                                                                                                                                                                                  | -                       | No       |
| <a id="http2-maxConcurrentStreams" href="#http2-maxConcurrentStreams" title="#http2-maxConcurrentStreams">`http2.maxConcurrentStreams`</a> | Set the number of concurrent streams per connection that each client is allowed to initiate. <br /> The value must be greater than zero.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | 250                     | No       |
| <a id="http3" href="#http3" title="#http3">`http3`</a> | Enable HTTP/3 protocol on the `entryPoint`. <br /> HTTP/3 requires a TCP `entryPoint`. as HTTP/3 always starts as a TCP connection that then gets upgraded to UDP. In most scenarios, this `entryPoint` is the same as the one used for TLS traffic.<br /> More information [here](#http3).                                                                                                                                                                                                                                                                                                                                                                                          | -                       | No       |
| <a id="http3-advertisedPort" href="#http3-advertisedPort" title="#http3-advertisedPort">`http3.advertisedPort`</a> | Set the UDP port to advertise as the HTTP/3 authority. <br /> It defaults to the entryPoint's address port. <br /> It can be used to override the authority in the `alt-svc` header, for example if the public facing port is different from where Traefik is listening.                                                                                                                                                                                                                                                                                                                                                                                                            | -                       | No       |
| <a id="observability-accessLogs" href="#observability-accessLogs" title="#observability-accessLogs">`observability.accessLogs`</a> | Defines whether a router attached to this EntryPoint produces access-logs by default. Nonetheless, a router defining its own observability configuration will opt-out from this default.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | true                    | No       |
| <a id="observability-metrics" href="#observability-metrics" title="#observability-metrics">`observability.metrics`</a> | Defines whether a router attached to this EntryPoint produces metrics by default. Nonetheless, a router defining its own observability configuration will opt-out from this default.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | true                    | No       |
| <a id="observability-tracing" href="#observability-tracing" title="#observability-tracing">`observability.tracing`</a> | Defines whether a router attached to this EntryPoint produces traces by default. Nonetheless, a router defining its own observability configuration will opt-out from this default.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | true                    | No       |
| <a id="observability-traceVerbosity" href="#observability-traceVerbosity" title="#observability-traceVerbosity">`observability.traceVerbosity`</a> | Defines the tracing verbosity level for routers attached to this EntryPoint. Possible values: `minimal` (default), `detailed`. Routers can override this value in their own observability configuration. <br /> More information [here](#traceverbosity).                                                                                                                                                                                                                                                                                                                                                                                                                           | minimal                 | No       |
| <a id="proxyProtocol-trustedIPs" href="#proxyProtocol-trustedIPs" title="#proxyProtocol-trustedIPs">`proxyProtocol.trustedIPs`</a> | Enable PROXY protocol with Trusted IPs. <br /> Traefik supports [PROXY protocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2. <br /> If PROXY protocol header parsing is enabled for the entry point, this entry point can accept connections with or without PROXY protocol headers. <br /> If the PROXY protocol header is passed, then the version is determined automatically.<br /> More information [here](#proxyprotocol-and-load-balancers).                                                                                                                                                                                               | -                       | No       |
| <a id="proxyProtocol-insecure" href="#proxyProtocol-insecure" title="#proxyProtocol-insecure">`proxyProtocol.insecure`</a> | Enable PROXY protocol trusting every incoming connection. <br /> Every remote client address will be replaced (`trustedIPs`) won't have any effect). <br /> Traefik supports [PROXY protocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2. <br /> If PROXY protocol header parsing is enabled for the entry point, this entry point can accept connections with or without PROXY protocol headers. <br /> If the PROXY protocol header is passed, then the version is determined automatically.<br />We recommend to use this option only for tests purposes, not in production.<br /> More information [here](#proxyprotocol-and-load-balancers). | -                       | No       |
| <a id="reusePort" href="#reusePort" title="#reusePort">`reusePort`</a> | Enable `entryPoints` from the same or different processes listening on the same TCP/UDP port by utilizing the `SO_REUSEPORT` socket option. <br /> It also allows the kernel to act like a load balancer to distribute incoming connections between entry points.<br /> More information [here](#reuseport).                                                                                                                                                                                                                                                                                                                                                                        | false                   | No       |
| <a id="transport-respondingTimeouts-readTimeout" href="#transport-respondingTimeouts-readTimeout" title="#transport-respondingTimeouts-readTimeout">`transport.`<br />`respondingTimeouts.`<br />`readTimeout`</a> | Set the timeouts for incoming requests to the Traefik instance. This is the maximum duration for reading the entire request, including the body. Setting them has no effect for UDP `entryPoints`.<br /> If zero, no timeout exists. <br />Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).<br />If no units are provided, the value is parsed assuming seconds.                                                                                                                                                                                                                                | 60s (seconds)           | No       |
| <a id="transport-respondingTimeouts-writeTimeout" href="#transport-respondingTimeouts-writeTimeout" title="#transport-respondingTimeouts-writeTimeout">`transport.`<br />`respondingTimeouts.`<br />`writeTimeout`</a> | Maximum duration before timing out writes of the response. <br /> It covers the time from the end of the request header read to the end of the response write. <br /> If zero, no timeout exists. <br />Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).<br />If no units are provided, the value is parsed assuming seconds.                                                                                                                                                                                                                                                                   | 0s (seconds)            | No       |
| <a id="transport-respondingTimeouts-idleTimeout" href="#transport-respondingTimeouts-idleTimeout" title="#transport-respondingTimeouts-idleTimeout">`transport.`<br />`respondingTimeouts.`<br />`idleTimeout`</a> | Maximum duration an idle (keep-alive) connection will remain idle before closing itself. <br /> If zero, no timeout exists <br />Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).<br />If no units are provided, the value is parsed assuming seconds                                                                                                                                                                                                                                                                                                                                           | 180s (seconds)          | No       |
| <a id="transport-lifeCycle-graceTimeOut" href="#transport-lifeCycle-graceTimeOut" title="#transport-lifeCycle-graceTimeOut">`transport.`<br />`lifeCycle.`<br />`graceTimeOut`</a> | Set the duration to give active requests a chance to finish before Traefik stops. <br />Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).<br />If no units are provided, the value is parsed assuming seconds <br /> In this time frame no new requests are accepted.                                                                                                                                                                                                                                                                                                                            | 10s (seconds)           | No       |
| <a id="transport-lifeCycle-requestAcceptGraceTimeout" href="#transport-lifeCycle-requestAcceptGraceTimeout" title="#transport-lifeCycle-requestAcceptGraceTimeout">`transport.`<br />`lifeCycle.`<br />`requestAcceptGraceTimeout`</a> | Set the duration to keep accepting requests prior to initiating the graceful termination period (as defined by the `transportlifeCycle.graceTimeOut` option). <br /> This option is meant to give downstream load-balancers sufficient time to take Traefik out of rotation. <br />Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).<br />If no units are provided, the value is parsed assuming seconds                                                                                                                                                                                         | 0s (seconds)            | No       |
| <a id="transport-keepAliveMaxRequests" href="#transport-keepAliveMaxRequests" title="#transport-keepAliveMaxRequests">`transport.`<br />`keepAliveMaxRequests`</a> | Set the maximum number of requests Traefik can handle before sending a `Connection: Close` header to the client (for HTTP2, Traefik sends a GOAWAY). <br /> Zero means no limit.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | 0                       | No       |
| <a id="transport-keepAliveMaxTime" href="#transport-keepAliveMaxTime" title="#transport-keepAliveMaxTime">`transport.`<br />`keepAliveMaxTime`</a> | Set the maximum duration Traefik can handle requests before sending a `Connection: Close` header to the client (for HTTP2, Traefik sends a GOAWAY). Zero means no limit.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | 0s (seconds)            | No       |
| <a id="udp-timeout" href="#udp-timeout" title="#udp-timeout">`udp.timeout`</a> | Define how long to wait on an idle session before releasing the related resources. <br />The Timeout value must be greater than zero.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               | 3s (seconds)            | No       |

### asDefault

If there is no entryPoint with the `asDefault` option set to `true`, then the 
list of default entryPoints includes all HTTP/TCP entryPoints.

If at least one entryPoint has the `asDefault` option set to `true`,
then the list of default entryPoints includes only entryPoints that have the
`asDefault` option set to `true`.

Some built-in entryPoints are always excluded from the list, namely: `traefik`.

The `asDefault` option has no effect on UDP entryPoints.
When a UDP router does not define the entryPoints option, it is attached to all
available UDP entryPoints.

### http.middlewares

- You can attach a list of [middlewares](../../middlewares/http/overview.md)
to each entryPoint.
- The middlewares will take effect only if the rule matches, and before forwarding
the request to the service.
- Middlewares are applied in the same order as their declaration.
- Middlewares are applied by default to every router exposed through the EntryPoint
(the Middlewares declared on the [IngressRoute](../../routing/routers/index.md#middlewares)
or the [Ingress](../../routing/providers/kubernetes-ingress.md#on-ingress)
are applied after the ones declared on the Entrypoint)
- The option allows attaching a list of middleware using the format 
`middlewarename@providername` as described in the example below:

```yaml tab="File (YAML)"
entryPoints:
  web:
    address: :80
    http:
      middlewares:
        - auth@kubernetescrd
        - strip@file
```

```yaml tab="Helm Chart Values"
ports:
  web:
    port: :80
    http:
      middlewares:
        - auth@kubernetescrd
        - strip@file
```

### encodeQuerySemicolons

Behavior examples:

| EncodeQuerySemicolons | Request Query       | Resulting Request Query |
|-----------------------|---------------------|-------------------------|
| <a id="false" href="#false" title="#false">false</a> | foo=bar;baz=bar     | foo=bar&baz=bar         |
| <a id="true" href="#true" title="#true">true</a> | foo=bar;baz=bar     | foo=bar%3Bbaz=bar       |
| <a id="false-2" href="#false-2" title="#false-2">false</a> | foo=bar&baz=bar;foo | foo=bar&baz=bar&foo     |
| <a id="true-2" href="#true-2" title="#true-2">true</a> | foo=bar&baz=bar;foo | foo=bar&baz=bar%3Bfoo   |

### SanitizePath

The `sanitizePath` option defines whether to enable the request path sanitization.
When disabled, the incoming request path is passed to the backend as is.
This can be useful when dealing with legacy clients that are not url-encoding data in the request path.
For example, as base64 uses the “/” character internally,
if it's not url encoded,
it can lead to unsafe routing when the `sanitizePath` option is set to `false`.

!!! warning "Security"

    Setting the sanitizePath option to false is not safe.
    Ensure every request is properly url encoded instead.

#### Examples

| SanitizePath | Request Path    | Resulting Request Path |
|--------------|-----------------|------------------------|
| <a id="false-3" href="#false-3" title="#false-3">false</a> | /./foo/bar      | /./foo/bar             |
| <a id="true-3" href="#true-3" title="#true-3">true</a> | /./foo/bar      | /foo/bar               |
| <a id="false-4" href="#false-4" title="#false-4">false</a> | /foo/../bar     | /foo/../bar            |
| <a id="true-4" href="#true-4" title="#true-4">true</a> | /foo/../bar     | /bar                   |
| <a id="false-5" href="#false-5" title="#false-5">false</a> | /foo/bar//      | /foo/bar//             |
| <a id="true-5" href="#true-5" title="#true-5">true</a> | /foo/bar//      | /foo/bar/              |
| <a id="false-6" href="#false-6" title="#false-6">false</a> | /./foo/../bar// | /./foo/../bar//        |
| <a id="true-6" href="#true-6" title="#true-6">true</a> | /./foo/../bar// | /bar/                  |

### HTTP3

As HTTP/3 actually uses UDP, when Traefik is configured with a TCP `entryPoint`
on port N with HTTP/3 enabled, the underlying HTTP/3 server that is started 
automatically listens on UDP port N too. As a consequence,
it means port N cannot be used by another UDP `entryPoint`.
Since HTTP/3 requires the use of TLS,
only routers with TLS enabled will be usable with HTTP/3.

### ProxyProtocol and Load-Balancers

The replacement of the remote client address will occur only for IP addresses listed in `trustedIPs`. This is where you specify your load balancer IPs or CIDR ranges.

When queuing Traefik behind another load-balancer, make sure to configure 
PROXY protocol on both sides.
Not doing so could introduce a security risk in your system (enabling request forgery).

### reusePort

#### Examples

Many processes on the same EntryPoint:

```yaml tab="File (YAML)"
  entryPoints:
    web:
      address: ":80"
      reusePort: true
```

```yaml tab="Helm Chart Values"
  ## Values file
  additionalArguments:
    - --entryPoints.web.reusePort=true
```

Many processes on the same EntryPoint on another host:

```yaml tab="File (YAML)"
entryPoints:
  web:
    address: ":80"
    reusePort: true
  privateWeb:
    address: "192.168.1.2:80"
    reusePort: true
```

```yaml tab="Helm Chart Values"
additionalArguments:
  - --entryPoints.web.reusePort=true
  - --entryPoints.privateWeb.address=192.168.1.2:80
  - --entryPoints.privateWeb.reusePort=true
```

#### Supported platforms

The `reusePort` option currently works only on Linux, FreeBSD, OpenBSD and Darwin.
It will be ignored on other platforms.

There is a known bug in the Linux kernel that may cause unintended TCP connection
failures when using the `reusePort` option. For more details, see [here](https://lwn.net/Articles/853637/).

#### Canary deployment

Use the `reusePort` option with the other option `transport.lifeCycle.gracetimeout`
to do
canary deployments against Traefik itself. Like upgrading Traefik version
or reloading the static configuration without any service downtime.

#### Trace Verbosity

`observability.traceVerbosity` defines the tracing verbosity level for routers attached to this EntryPoint.
Routers can override this value in their own observability configuration.

Possible values are:

- `minimal`: produces a single server span and one client span for each request processed by a router.
- `detailed`: enables the creation of additional spans for each middleware executed for each request processed by a router.
