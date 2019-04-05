# EntryPoints

Opening Connections for Incoming Requests
{: .subtitle }

![EntryPoints](../assets/img/entrypoints.png)

EntryPoints are the network entry points into Traefik.
They define the port which will receive the requests (whether HTTP or TCP).

## Configuration Examples

??? example "Port 80 only"

    ```toml
    [entrypoints]
      [entrypoints.web]
         address = ":80"
    ```

    We define an `entrypoint` called `web` that will listen on port `80`.

??? example "Port 80 & 443" 

    ```toml
    [entrypoints]
      [entrypoints.web]
        address = ":80"
    
      [entrypoints.web-secure]
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
[EntryPoints]

  [EntryPoints.EntryPoint0]
    Address = "foobar"
    [EntryPoints.EntryPoint0.Transport]
      [EntryPoints.EntryPoint0.Transport.LifeCycle]
        RequestAcceptGraceTimeout = 42
        GraceTimeOut = 42
      [EntryPoints.EntryPoint0.Transport.RespondingTimeouts]
        ReadTimeout = 42
        WriteTimeout = 42
        IdleTimeout = 42
    [EntryPoints.EntryPoint0.ProxyProtocol]
      Insecure = true
      TrustedIPs = ["foobar", "foobar"]
    [EntryPoints.EntryPoint0.ForwardedHeaders]
      Insecure = true
      TrustedIPs = ["foobar", "foobar"]
```

```ini tab="CLI"
Name:EntryPoint0
Address:foobar
Transport.LifeCycle.RequestAcceptGraceTimeout:42
Transport.LifeCycle.GraceTimeOut:42
Transport.RespondingTimeouts.ReadTimeout:42
Transport.RespondingTimeouts.WriteTimeout:42
Transport.RespondingTimeouts.IdleTimeout:42
ProxyProtocol.Insecure:true
ProxyProtocol.TrustedIPs:foobar,foobar
ForwardedHeaders.Insecure:true
ForwardedHeaders.TrustedIPs:foobar,foobar
```

??? example "Using the CLI"

    Here is an example of using the CLI to define `entrypoints`:
    
    ```shell
    --entryPoints='Name:http Address::80'
    --entryPoints='Name:https Address::443'
    ```
    
    !!! note
        The whitespace character (` `) is the option separator, and the comma (`,`) is the value separator for lists inside an option.  
        The option names are case-insensitive.
    
    !!! warning "Using Docker Compose Files"
    
        The syntax for passing arguments inside a docker compose file is a little different. Here are two examples.
        
        ```yaml
        traefik:
            image: traefik:v2.0 # The official v2.0 Traefik docker image
            command:
                - --defaultentrypoints=powpow
                - "--entryPoints=Name:powpow Address::42 Compress:true"
        ```

        or
        
        ```yaml
        traefik:
            image: traefik:v2.0 # The official v2.0 Traefik docker image
            command: --defaultentrypoints=powpow --entryPoints='Name:powpow Address::42 Compress:true'
        ```

## ProxyProtocol

Traefik supports [ProxyProtocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt).

??? example "Enabling Proxy Protocol with Trusted IPs" 

    ```toml
    [entrypoints]
      [entrypoints.web]
        address = ":80"
    
        [entrypoints.web.proxyProtocol]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```
    
    IPs in `trustedIPs` only will lead to remote client address replacement: Declare load-balancer IPs or CIDR range here.
    
??? example "Insecure Mode -- Testing Environnement Only"

    In a test environments, you can configure Traefik to trust every incoming connection. Doing so, every remote client address will be replaced (`trustedIPs` won't have any effect)

    ```toml
    [entrypoints]
      [entrypoints.web]
        address = ":80"
    
        [entrypoints.web.proxyProtocol]
          insecure = true
    ```
         
!!! warning "Queuing Traefik behind Another Load Balancer"

    When queuing Traefik behind another load-balancer, make sure to configure Proxy Protocol on both sides.
    Not doing so could introduce a security risk in your system (enabling request forgery).

## Forwarded Header

You can configure Traefik to trust the forwarded headers information (`X-Forwarded-*`)

??? example "Trusting Forwarded Headers from specific IPs"

    ```toml
    [entrypoints]
      [entrypoints.web]
        address = ":80"
    
        [entrypoints.web.forwardedHeaders]
          trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
    ```
    
??? example "Insecure Mode -- Always Trusting Forwarded Headers"

    ```toml
    [entrypoints]
      [entrypoints.web]
        address = ":80"
    
        [entrypoints.web.forwardedHeaders]
           insecure = true
    ```
