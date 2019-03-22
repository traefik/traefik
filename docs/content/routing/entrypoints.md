# EntryPoints

Opening Connections for Incomming Requests
{: .subtitle }

![EntryPoints](../assets/img/entrypoints.png)

Entrypoints are the network entry points into Traefik.
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
        
        [entrypoints.web-secure.tls]
          [[entrypoints.web-secure.tls.certificates]]
            certFile = "tests/traefik.crt"
            keyFile = "tests/traefik.key"
    ```

    - Two entrypoints are defined: one called `web`, and the other called `web-secure`.
    - `web` listens on port `80`, and `web-secure` on port `443`. 
    
## Configuration

### General

Entrypoints are part of the [static configuration](../getting-started/configuration-overview.md#the-static-configuration). You can define them using a toml file, CLI arguments, or a key-value store. See the [complete reference](../reference/entrypoints.md) for the list of available options. 

??? example "Using the CLI"

    Here is an example of using the CLI to define `entrypoints`:
    
    ```shell
    --entryPoints='Name:http Address::80'
    --entryPoints='Name:https Address::443 TLS'
    ```
    
    !!! note
        The whitespace character (` `) is the option separator, and the comma (`,`) is the value separator for lists.  
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

    In a test environments, you can configure Traefik to trust every incomming connection. Doing so, every remote client address will be replaced (`trustedIPs` won't have any effect)

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
