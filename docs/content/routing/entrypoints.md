# EntryPoints

Opening Connections for Incoming Requests
{: .subtitle }

![EntryPoints](../assets/img/entrypoints.png)

EntryPoints are the network entry points into Traefik.
They define the port which will receive the requests (whether HTTP or TCP).

## Configuration Examples

??? example "Port 80 only"

    ```toml
    [entryPoints]
      [entryPoints.web]
         address = ":80"
    ```

    We define an `entrypoint` called `web` that will listen on port `80`.

??? example "Port 80 & 443" 

    ```toml
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
      [entryPoints.web-secure]
        address = ":443"
    ```

    - Two entrypoints are defined: one called `web`, and the other called `web-secure`.
    - `web` listens on port `80`, and `web-secure` on port `443`. 
    
## Configuration

### General

EntryPoints are part of the [static configuration](../getting-started/configuration-overview.md#the-static-configuration).
You can define them using a toml file, CLI arguments, or a key-value store.

See the complete reference for the list of available options:

```toml tab="File"
[entryPoints]

  [entryPoints.EntryPoint0]
    Address = ":8888"
    [entryPoints.EntryPoint0.Transport]
      [entryPoints.EntryPoint0.Transport.LifeCycle]
        RequestAcceptGraceTimeout = 42
        GraceTimeOut = 42
      [entryPoints.EntryPoint0.Transport.RespondingTimeouts]
        ReadTimeout = 42
        WriteTimeout = 42
        IdleTimeout = 42
    [entryPoints.EntryPoint0.ProxyProtocol]
      Insecure = true
      TrustedIPs = ["foobar", "foobar"]
    [entryPoints.EntryPoint0.ForwardedHeaders]
      Insecure = true
      TrustedIPs = ["foobar", "foobar"]
```

```ini tab="CLI"
--entryPoints.EntryPoint0.Address=:8888
--entryPoints.EntryPoint0.Transport.LifeCycle.RequestAcceptGraceTimeout=42
--entryPoints.EntryPoint0.Transport.LifeCycle.GraceTimeOut=42
--entryPoints.EntryPoint0.Transport.RespondingTimeouts.ReadTimeout=42
--entryPoints.EntryPoint0.Transport.RespondingTimeouts.WriteTimeout=42
--entryPoints.EntryPoint0.Transport.RespondingTimeouts.IdleTimeout=42
--entryPoints.EntryPoint0.ProxyProtocol.Insecure=true
--entryPoints.EntryPoint0.ProxyProtocol.TrustedIPs=foobar,foobar
--entryPoints.EntryPoint0.ForwardedHeaders.Insecure=true
--entryPoints.EntryPoint0.ForwardedHeaders.TrustedIPs=foobar,foobar
```

## ProxyProtocol

Traefik supports [ProxyProtocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt).

??? example "Enabling Proxy Protocol with Trusted IPs" 

    ```toml
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.proxyProtocol]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```
    
    IPs in `trustedIPs` only will lead to remote client address replacement: Declare load-balancer IPs or CIDR range here.
    
??? example "Insecure Mode -- Testing Environment Only"

    In a test environments, you can configure Traefik to trust every incoming connection.
    Doing so, every remote client address will be replaced (`trustedIPs` won't have any effect)

    ```toml
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.proxyProtocol]
          insecure = true
    ```
         
!!! warning "Queuing Traefik behind Another Load Balancer"

    When queuing Traefik behind another load-balancer, make sure to configure Proxy Protocol on both sides.
    Not doing so could introduce a security risk in your system (enabling request forgery).

## Forwarded Header

You can configure Traefik to trust the forwarded headers information (`X-Forwarded-*`)

??? example "Trusting Forwarded Headers from specific IPs"

    ```toml
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.forwardedHeaders]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```
    
??? example "Insecure Mode -- Always Trusting Forwarded Headers"

    ```toml
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
        [entryPoints.web.forwardedHeaders]
           insecure = true
    ```
