# EntryPoints

Opening Connections for Incoming Requests
{: .subtitle }

![entryPoints](../assets/img/entrypoints.png)

EntryPoints are the network entry points into Traefik.
They define the port which will receive the requests (whether HTTP or TCP).

## Configuration Examples

??? example "Port 80 only"

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
       address: ":80"
    ```
    
    ```bash tab="CLI"
    --entryPoints.web.address=:80
    ```

    We define an `entrypoint` called `web` that will listen on port `80`.

??? example "Port 80 & 443" 

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
      [entryPoints.web-secure]
        address = ":443"
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"
     
      web-secure:
        address: ":443"
    ```
    
    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web-secure.address=:443
    ```

    - Two entrypoints are defined: one called `web`, and the other called `web-secure`.
    - `web` listens on port `80`, and `web-secure` on port `443`. 
    
## Configuration

### General

EntryPoints are part of the [static configuration](../getting-started/configuration-overview.md#the-static-configuration).
You can define them using a toml file, CLI arguments, or a key-value store.

See the complete reference for the list of available options:

```toml tab="File (TOML)"
[entryPoints]

  [entryPoints.EntryPoint0]
    address = ":8888"
    [entryPoints.EntryPoint0.transport]
      [entryPoints.EntryPoint0.transport.lifeCycle]
        requestAcceptGraceTimeout = 42
        graceTimeOut = 42
      [entryPoints.EntryPoint0.transport.respondingTimeouts]
        readTimeout = 42
        writeTimeout = 42
        idleTimeout = 42
    [entryPoints.EntryPoint0.proxyProtocol]
      insecure = true
      trustedIPs = ["foobar", "foobar"]
    [entryPoints.EntryPoint0.forwardedHeaders]
      insecure = true
      trustedIPs = ["foobar", "foobar"]
```

```yaml tab="File (YAML)"
entryPoints:

  EntryPoint0:
    address: ":8888"
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
      - "foobar"
      - "foobar"
    forwardedHeaders:
      insecure: true
      trustedIPs:
      - "foobar"
      - "foobar"
```

```bash tab="CLI"
--entryPoints.EntryPoint0.address=:8888
--entryPoints.EntryPoint0.transport.lifeCycle.requestAcceptGraceTimeout=42
--entryPoints.EntryPoint0.transport.lifeCycle.graceTimeOut=42
--entryPoints.EntryPoint0.transport.respondingTimeouts.readTimeout=42
--entryPoints.EntryPoint0.transport.respondingTimeouts.writeTimeout=42
--entryPoints.EntryPoint0.transport.respondingTimeouts.idleTimeout=42
--entryPoints.EntryPoint0.proxyProtocol.insecure=true
--entryPoints.EntryPoint0.proxyProtocol.trustedIPs=foobar,foobar
--entryPoints.EntryPoint0.forwardedHeaders.insecure=true
--entryPoints.EntryPoint0.forwardedHeaders.trustedIPs=foobar,foobar
```

## ProxyProtocol

Traefik supports [ProxyProtocol](https://www.haproxy.org/download/2.0/doc/proxy-protocol.txt) version 1 and 2.

If proxyprotocol header parsing is enabled for the entry point, this entry point can accept connections with or without proxyprotocol headers.

If the proxyprotocol header is passed, then the version is determined automatically.

??? example "Enabling Proxy Protocol with Trusted IPs" 

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.proxyProtocol]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"
        proxyProtocol:
          trustedIPs:
          - "127.0.0.1/32"
          - "192.168.1.7"
    ```
    
    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web.proxyProtocol.trustedIPs=127.0.0.1/32,192.168.1.7
    ```

    IPs in `trustedIPs` only will lead to remote client address replacement: Declare load-balancer IPs or CIDR range here.
    
??? example "Insecure Mode -- Testing Environment Only"

    In a test environments, you can configure Traefik to trust every incoming connection.
    Doing so, every remote client address will be replaced (`trustedIPs` won't have any effect)

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.proxyProtocol]
          insecure = true
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"
        proxyProtocol:
          insecure: true
    ```
    
    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web.proxyProtocol.insecure
    ```

!!! warning "Queuing Traefik behind Another Load Balancer"

    When queuing Traefik behind another load-balancer, make sure to configure Proxy Protocol on both sides.
    Not doing so could introduce a security risk in your system (enabling request forgery).

## Forwarded Header

You can configure Traefik to trust the forwarded headers information (`X-Forwarded-*`)

??? example "Trusting Forwarded Headers from specific IPs"

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.forwardedHeaders]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"
        forwardedHeaders:
          trustedIPs:
          - "127.0.0.1/32"
          - "192.168.1.7"
    ```
    
    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web.forwardedHeaders.trustedIPs=127.0.0.1/32,192.168.1.7
    ```

??? example "Insecure Mode -- Always Trusting Forwarded Headers"

    ```toml tab="File (TOML)"
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.forwardedHeaders]
          insecure = true
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"
        forwardedHeaders:
          insecure: true
    ```
    
    ```bash tab="CLI"
    --entryPoints.web.address=:80
    --entryPoints.web.forwardedHeaders.insecure
    ```
