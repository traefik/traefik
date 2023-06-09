---
title: "Traefik EntryPoints Documentation"
description: "For routing and load balancing in Traefik Proxy, EntryPoints define which port will receive packets and whether in UDP or TCP. Read the technical documentation."
---

# EntryPoints

Opening Connections for Incoming Requests
{: .subtitle }

![entryPoints](../assets/img/entrypoints.png)

EntryPoints are the network entry points into Traefik.
They define the port which will receive the packets,
and whether to listen for TCP or UDP.

## Configuration Examples

??? example "Port 80 only"

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
       address: ":80"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.web.address=:80
    ```

    We define an `entrypoint` called `web` that will listen on port `80`.

??? example "Port 80 & 443"

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"

      websecure:
        address: ":443"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"

      [entryPoints.websecure]
        address = ":443"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.web.address=:80
    --entryPoints.websecure.address=:443
    ```

    - Two entrypoints are defined: one called `web`, and the other called `websecure`.
    - `web` listens on port `80`, and `websecure` on port `443`.

??? example "UDP on port 1704"

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      streaming:
        address: ":1704/udp"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.streaming]
        address = ":1704/udp"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.streaming.address=:1704/udp
    ```

## Configuration

### General

EntryPoints are part of the [static configuration](../getting-started/configuration-overview.md#the-static-configuration).
They can be defined by using a file (YAML or TOML) or CLI arguments.

??? info "See the complete reference for the list of available options"

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      name:
        address: ":8888" # same as ":8888/tcp"
        http2:
          maxConcurrentStreams: 42
        http3:
          advertisedPort: 8888
        transport:
          lifeCycle:
            requestAcceptGraceTimeout: 42
            graceTimeOut: 42
          respondingTimeouts:
            readTimeout: 42
            writeTimeout: 42
            idleTimeout: 42
        proxyProtocol:
          insecure: true
          trustedIPs:
            - "127.0.0.1"
            - "192.168.0.1"
        forwardedHeaders:
          insecure: true
          trustedIPs:
            - "127.0.0.1"
            - "192.168.0.1"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.name]
        address = ":8888" # same as ":8888/tcp"
        [entryPoints.name.http2]
          maxConcurrentStreams = 42
        [entryPoints.name.http3]
          advertisedPort = 8888
        [entryPoints.name.transport]
          [entryPoints.name.transport.lifeCycle]
            requestAcceptGraceTimeout = 42
            graceTimeOut = 42
          [entryPoints.name.transport.respondingTimeouts]
            readTimeout = 42
            writeTimeout = 42
            idleTimeout = 42
        [entryPoints.name.proxyProtocol]
          insecure = true
          trustedIPs = ["127.0.0.1", "192.168.0.1"]
        [entryPoints.name.forwardedHeaders]
          insecure = true
          trustedIPs = ["127.0.0.1", "192.168.0.1"]
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.name.address=:8888 # same as :8888/tcp
    --entryPoints.name.http2.maxConcurrentStreams=42
    --entryPoints.name.http3.advertisedport=8888
    --entryPoints.name.transport.lifeCycle.requestAcceptGraceTimeout=42
    --entryPoints.name.transport.lifeCycle.graceTimeOut=42
    --entryPoints.name.transport.respondingTimeouts.readTimeout=42
    --entryPoints.name.transport.respondingTimeouts.writeTimeout=42
    --entryPoints.name.transport.respondingTimeouts.idleTimeout=42
    --entryPoints.name.proxyProtocol.insecure=true
    --entryPoints.name.proxyProtocol.trustedIPs=127.0.0.1,192.168.0.1
    --entryPoints.name.forwardedHeaders.insecure=true
    --entryPoints.name.forwardedHeaders.trustedIPs=127.0.0.1,192.168.0.1
    ```

### Address

The address defines the port, and optionally the hostname, on which to listen for incoming connections and packets.
It also defines the protocol to use (TCP or UDP).
If no protocol is specified, the default is TCP.
The format is:

```bash
[host]:port[/tcp|/udp]
```

If both TCP and UDP are wanted for the same port, two entryPoints definitions are needed, such as in the example below.

??? example "Both TCP and UDP on Port 3179"

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      tcpep:
       address: ":3179"
      udpep:
       address: ":3179/udp"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.tcpep]
        address = ":3179"
      [entryPoints.udpep]
        address = ":3179/udp"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.tcpep.address=:3179
    --entryPoints.udpep.address=:3179/udp
    ```

??? example "Listen on Specific IP Addresses Only"

    ```yaml tab="File (yaml)"
    entryPoints:
      specificIPv4:
        address: "192.168.2.7:8888"
      specificIPv6:
        address: "[2001:db8::1]:8888"
    ```

    ```toml tab="File (TOML)"
    [entryPoints.specificIPv4]
      address = "192.168.2.7:8888"
    [entryPoints.specificIPv6]
      address = "[2001:db8::1]:8888"
    ```

    ```bash tab="CLI"
    --entrypoints.specificIPv4.address=192.168.2.7:8888
    --entrypoints.specificIPv6.address=[2001:db8::1]:8888
    ```

    Full details for how to specify `address` can be found in [net.Listen](https://golang.org/pkg/net/#Listen) (and [net.Dial](https://golang.org/pkg/net/#Dial)) of the doc for go.

### HTTP/2

#### `maxConcurrentStreams`

_Optional, Default=250_

`maxConcurrentStreams` specifies the number of concurrent streams per connection that each client is allowed to initiate.
The `maxConcurrentStreams` value must be greater than zero.

```yaml tab="File (YAML)"
entryPoints:
  foo:
    http2:
      maxConcurrentStreams: 250
```

```toml tab="File (TOML)"
[entryPoints.foo]
  [entryPoints.foo.http2]
    maxConcurrentStreams = 250
```

```bash tab="CLI"
--entryPoints.name.http2.maxConcurrentStreams=250
```

### HTTP/3

#### `http3`

`http3` enables HTTP/3 protocol on the entryPoint.
HTTP/3 requires a TCP entryPoint, as HTTP/3 always starts as a TCP connection that then gets upgraded to UDP.
In most scenarios, this entryPoint is the same as the one used for TLS traffic.

??? info "HTTP/3 uses UDP+TLS"

    As HTTP/3 uses UDP, you can't have a TCP entryPoint with HTTP/3 on the same port as a UDP entryPoint.
    Since HTTP/3 requires the use of TLS, only routers with TLS enabled will be usable with HTTP/3.

!!! warning "Enabling Experimental HTTP/3"

    As the HTTP/3 spec is still in draft, HTTP/3 support in Traefik is an experimental feature and needs to be activated 
    in the experimental section of the static configuration.
    
    ```yaml tab="File (YAML)"
    experimental:
      http3: true

    entryPoints:
      name:
        http3: {}
    ```

    ```toml tab="File (TOML)"
    [experimental]
      http3 = true
    
    [entryPoints.name.http3]
    ```
    
    ```bash tab="CLI"
    --experimental.http3=true 
    --entrypoints.name.http3
    ```

#### `advertisedPort`

`http3.advertisedPort` defines which UDP port to advertise as the HTTP/3 authority.
It defaults to the entryPoint's address port.
It can be used to override the authority in the `alt-svc` header, for example if the public facing port is different from where Traefik is listening.

!!! info "http3.advertisedPort"

    ```yaml tab="File (YAML)"
    experimental:
      http3: true

    entryPoints:
      name:
        http3:
          advertisedPort: 443
    ```

    ```toml tab="File (TOML)"
    [experimental]
      http3 = true
    
    [entryPoints.name.http3]
      advertisedPort = 443
    ```
    
    ```bash tab="CLI"
    --experimental.http3=true 
    --entrypoints.name.http3.advertisedport=443
    ```

### Forwarded Headers

You can configure Traefik to trust the forwarded headers information (`X-Forwarded-*`).

??? info "`forwardedHeaders.trustedIPs`"

    Trusting Forwarded Headers from specific IPs.

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
        forwardedHeaders:
          trustedIPs:
            - "127.0.0.1/32"
            - "192.168.1.7"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"

        [entryPoints.web.forwardedHeaders]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.web.address=:80
    --entryPoints.web.forwardedHeaders.trustedIPs=127.0.0.1/32,192.168.1.7
    ```

??? info "`forwardedHeaders.insecure`"

    Insecure Mode (Always Trusting Forwarded Headers).

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
        forwardedHeaders:
          insecure: true
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"

        [entryPoints.web.forwardedHeaders]
          insecure = true
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.web.address=:80
    --entryPoints.web.forwardedHeaders.insecure
    ```

### Transport

#### `respondingTimeouts`

`respondingTimeouts` are timeouts for incoming requests to the Traefik instance.
Setting them has no effect for UDP entryPoints.

??? info "`transport.respondingTimeouts.readTimeout`"

    _Optional, Default=0s_

    `readTimeout` is the maximum duration for reading the entire request, including the body.

    If zero, no timeout exists.  
    Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
    If no units are provided, the value is parsed assuming seconds.

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      name:
        address: ":8888"
        transport:
          respondingTimeouts:
            readTimeout: 42
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.name]
        address = ":8888"
        [entryPoints.name.transport]
          [entryPoints.name.transport.respondingTimeouts]
            readTimeout = 42
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.name.address=:8888
    --entryPoints.name.transport.respondingTimeouts.readTimeout=42
    ```

??? info "`transport.respondingTimeouts.writeTimeout`"

    _Optional, Default=0s_

    `writeTimeout` is the maximum duration before timing out writes of the response.

    It covers the time from the end of the request header read to the end of the response write.
    If zero, no timeout exists.  
    Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
    If no units are provided, the value is parsed assuming seconds.

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      name:
        address: ":8888"
        transport:
          respondingTimeouts:
            writeTimeout: 42
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.name]
        address = ":8888"
        [entryPoints.name.transport]
          [entryPoints.name.transport.respondingTimeouts]
            writeTimeout = 42
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.name.address=:8888
    --entryPoints.name.transport.respondingTimeouts.writeTimeout=42
    ```

??? info "`transport.respondingTimeouts.idleTimeout`"

    _Optional, Default=180s_

    `idleTimeout` is the maximum duration an idle (keep-alive) connection will remain idle before closing itself.

    If zero, no timeout exists.  
    Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).
    If no units are provided, the value is parsed assuming seconds.

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      name:
        address: ":8888"
        transport:
          respondingTimeouts:
            idleTimeout: 42
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.name]
        address = ":8888"
        [entryPoints.name.transport]
          [entryPoints.name.transport.respondingTimeouts]
            idleTimeout = 42
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.name.address=:8888
    --entryPoints.name.transport.respondingTimeouts.idleTimeout=42
    ```

#### `lifeCycle`

Controls the behavior of Traefik during the shutdown phase.

??? info "`lifeCycle.requestAcceptGraceTimeout`"

    _Optional, Default=0s_

    Duration to keep accepting requests prior to initiating the graceful termination period (as defined by the `graceTimeOut` option).
    This option is meant to give downstream load-balancers sufficient time to take Traefik out of rotation.

    Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).

    If no units are provided, the value is parsed assuming seconds.
    The zero duration disables the request accepting grace period, i.e., Traefik will immediately proceed to the grace period.

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      name:
        address: ":8888"
        transport:
          lifeCycle:
            requestAcceptGraceTimeout: 42
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.name]
        address = ":8888"
        [entryPoints.name.transport]
          [entryPoints.name.transport.lifeCycle]
            requestAcceptGraceTimeout = 42
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.name.address=:8888
    --entryPoints.name.transport.lifeCycle.requestAcceptGraceTimeout=42
    ```

??? info "`lifeCycle.graceTimeOut`"

    _Optional, Default=10s_

    Duration to give active requests a chance to finish before Traefik stops.

    Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration) or as raw values (digits).

    If no units are provided, the value is parsed assuming seconds.

    !!! warning "In this time frame no new requests are accepted."

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      name:
        address: ":8888"
        transport:
          lifeCycle:
            graceTimeOut: 42
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.name]
        address = ":8888"
        [entryPoints.name.transport]
          [entryPoints.name.transport.lifeCycle]
            graceTimeOut = 42
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.name.address=:8888
    --entryPoints.name.transport.lifeCycle.graceTimeOut=42
    ```

### ProxyProtocol

Traefik supports [ProxyProtocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2.

If Proxy Protocol header parsing is enabled for the entry point, this entry point can accept connections with or without Proxy Protocol headers.

If the Proxy Protocol header is passed, then the version is determined automatically.

??? info "`proxyProtocol.trustedIPs`"

    Enabling Proxy Protocol with Trusted IPs.

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
        proxyProtocol:
          trustedIPs:
            - "127.0.0.1/32"
            - "192.168.1.7"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"

        [entryPoints.web.proxyProtocol]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```

    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web.proxyProtocol.trustedIPs=127.0.0.1/32,192.168.1.7
    ```

    IPs in `trustedIPs` only will lead to remote client address replacement: Declare load-balancer IPs or CIDR range here.

??? info "`proxyProtocol.insecure`"

    Insecure Mode (Testing Environment Only).

    In a test environments, you can configure Traefik to trust every incoming connection.
    Doing so, every remote client address will be replaced (`trustedIPs` won't have any effect)

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
        proxyProtocol:
          insecure: true
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"

        [entryPoints.web.proxyProtocol]
          insecure = true
    ```

    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web.proxyProtocol.insecure
    ```

!!! warning "Queuing Traefik behind Another Load Balancer"

    When queuing Traefik behind another load-balancer, make sure to configure Proxy Protocol on both sides.
    Not doing so could introduce a security risk in your system (enabling request forgery).

## HTTP Options

This whole section is dedicated to options, keyed by entry point, that will apply only to HTTP routing.

### Redirection

??? example "HTTPS redirection (80 to 443)"

    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: :80
        http:
          redirections:
            entryPoint:
              to: websecure
              scheme: https

      websecure:
        address: :443
    ```

    ```toml tab="File (TOML)"
    [entryPoints.web]
      address = ":80"

      [entryPoints.web.http]
        [entryPoints.web.http.redirections]
          [entryPoints.web.http.redirections.entryPoint]
            to = "websecure"
            scheme = "https"

    [entryPoints.websecure]
      address = ":443"
    ```

    ```bash tab="CLI"
    --entrypoints.web.address=:80
    --entrypoints.web.http.redirections.entryPoint.to=websecure
    --entrypoints.web.http.redirections.entryPoint.scheme=https
    --entrypoints.websecure.address=:443
    ```

#### `entryPoint`

This section is a convenience to enable (permanent) redirecting of all incoming requests on an entry point (e.g. port `80`) to another entry point (e.g. port `443`) or an explicit port (`:443`).

??? info "`entryPoint.to`"

    _Required_

    The target element, it can be:

      - an entry point name (ex: `websecure`)
      - a port (`:443`)

    ```yaml tab="File (YAML)"
    entryPoints:
      foo:
        # ...
        http:
          redirections:
            entryPoint:
              to: websecure
    ```

    ```toml tab="File (TOML)"
    [entryPoints.foo]
      # ...
      [entryPoints.foo.http.redirections]
        [entryPoints.foo.http.redirections.entryPoint]
          to = "websecure"
    ```

    ```bash tab="CLI"
    --entrypoints.foo.http.redirections.entryPoint.to=websecure
    ```

??? info "`entryPoint.scheme`"

    _Optional, Default="https"_

    The redirection target scheme.

    ```yaml tab="File (YAML)"
    entryPoints:
      foo:
        # ...
        http:
          redirections:
            entryPoint:
              # ...
              scheme: https
    ```

    ```toml tab="File (TOML)"
    [entryPoints.foo]
      # ...
      [entryPoints.foo.http.redirections]
        [entryPoints.foo.http.redirections.entryPoint]
          # ...
          scheme = "https"
    ```

    ```bash tab="CLI"
    --entrypoints.foo.http.redirections.entryPoint.scheme=https
    ```

??? info "`entryPoint.permanent`"

    _Optional, Default=true_

    To apply a permanent redirection.

    ```yaml tab="File (YAML)"
    entryPoints:
      foo:
        # ...
        http:
          redirections:
            entryPoint:
              # ...
              permanent: true
    ```

    ```toml tab="File (TOML)"
    [entryPoints.foo]
      # ...
      [entryPoints.foo.http.redirections]
        [entryPoints.foo.http.redirections.entryPoint]
          # ...
          permanent = true
    ```

    ```bash tab="CLI"
    --entrypoints.foo.http.redirections.entrypoint.permanent=true
    ```

??? info "`entryPoint.priority`"

    _Optional, Default=MaxInt32-1 (2147483646)_

    Priority of the generated router.

    ```yaml tab="File (YAML)"
    entryPoints:
      foo:
        # ...
        http:
          redirections:
            entryPoint:
              # ...
              priority: 10
    ```

    ```toml tab="File (TOML)"
    [entryPoints.foo]
      # ...
      [entryPoints.foo.http.redirections]
        [entryPoints.foo.http.redirections.entryPoint]
          # ...
          priority = 10
    ```

    ```bash tab="CLI"
    --entrypoints.foo.http.redirections.entrypoint.priority=10
    ```

### EncodeQuerySemicolons

_Optional, Default=false_

The `encodeQuerySemicolons` option allows to enable query semicolons encoding.
One could use this option to avoid non-encoded semicolons to be interpreted as query parameter separators by Traefik.
When using this option, the non-encoded semicolons characters in query will be transmitted encoded to the backend.

```yaml tab="File (YAML)"
entryPoints:
  websecure:
    address: ':443'
    http:
      encodeQuerySemicolons: true
```

```toml tab="File (TOML)"
[entryPoints.websecure]
  address = ":443"

  [entryPoints.websecure.http]
    encodeQuerySemicolons = true
```

```bash tab="CLI"
--entrypoints.websecure.address=:443
--entrypoints.websecure.http.encodequerysemicolons=true
```

#### Examples

| EncodeQuerySemicolons | Request Query       | Resulting Request Query |
|-----------------------|---------------------|-------------------------|
| false                 | foo=bar;baz=bar     | foo=bar&baz=bar         |
| true                  | foo=bar;baz=bar     | foo=bar%3Bbaz=bar       |
| false                 | foo=bar&baz=bar;foo | foo=bar&baz=bar&foo     |
| true                  | foo=bar&baz=bar;foo | foo=bar&baz=bar%3Bfoo   |

### Middlewares

The list of middlewares that are prepended by default to the list of middlewares of each router associated to the named entry point.

```yaml tab="File (YAML)"
entryPoints:
  websecure:
    address: ':443'
    http:
      middlewares:
        - auth@file
        - strip@file
```

```toml tab="File (TOML)"
[entryPoints.websecure]
  address = ":443"

  [entryPoints.websecure.http]
    middlewares = ["auth@file", "strip@file"]
```

```bash tab="CLI"
--entrypoints.websecure.address=:443
--entrypoints.websecure.http.middlewares=auth@file,strip@file
```

### TLS

This section is about the default TLS configuration applied to all routers associated with the named entry point.

If a TLS section (i.e. any of its fields) is user-defined, then the default configuration does not apply at all.

The TLS section is the same as the [TLS section on HTTP routers](./routers/index.md#tls).

```yaml tab="File (YAML)"
entryPoints:
  websecure:
    address: ':443'
    http:
      tls:
        options: foobar
        certResolver: leresolver
        domains:
          - main: example.com
            sans:
              - foo.example.com
              - bar.example.com
          - main: test.com
            sans:
              - foo.test.com
              - bar.test.com
```

```toml tab="File (TOML)"
[entryPoints.websecure]
  address = ":443"

    [entryPoints.websecure.http.tls]
      options = "foobar"
      certResolver = "leresolver"
      [[entryPoints.websecure.http.tls.domains]]
        main = "example.com"
        sans = ["foo.example.com", "bar.example.com"]
      [[entryPoints.websecure.http.tls.domains]]
        main = "test.com"
        sans = ["foo.test.com", "bar.test.com"]
```

```bash tab="CLI"
--entrypoints.websecure.address=:443
--entrypoints.websecure.http.tls.options=foobar
--entrypoints.websecure.http.tls.certResolver=leresolver
--entrypoints.websecure.http.tls.domains[0].main=example.com
--entrypoints.websecure.http.tls.domains[0].sans=foo.example.com,bar.example.com
--entrypoints.websecure.http.tls.domains[1].main=test.com
--entrypoints.websecure.http.tls.domains[1].sans=foo.test.com,bar.test.com
```

??? example "Let's Encrypt"

    ```yaml tab="File (YAML)"
    entryPoints:
      websecure:
        address: ':443'
        http:
          tls:
            certResolver: leresolver
    ```

    ```toml tab="File (TOML)"
    [entryPoints.websecure]
      address = ":443"

        [entryPoints.websecure.http.tls]
          certResolver = "leresolver"
    ```

    ```bash tab="CLI"
    --entrypoints.websecure.address=:443
    --entrypoints.websecure.http.tls.certResolver=leresolver
    ```

## UDP Options

This whole section is dedicated to options, keyed by entry point, that will apply only to UDP routing.

### Timeout

_Optional, Default=3s_

Timeout defines how long to wait on an idle session before releasing the related resources.
The Timeout value must be greater than zero.

```yaml tab="File (YAML)"
entryPoints:
  foo:
    address: ':8000/udp'
    udp:
      timeout: 10s
```

```toml tab="File (TOML)"
[entryPoints.foo]
  address = ":8000/udp"

    [entryPoints.foo.udp]
      timeout = "10s"
```

```bash tab="CLI"
entrypoints.foo.address=:8000/udp
entrypoints.foo.udp.timeout=10s
```

{!traefik-for-business-applications.md!}
