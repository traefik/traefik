---
title: "Traefik WebSocket Documentation"
description: "How to configure WebSocket and WebSocket Secure (WSS) connections with Traefik Proxy."
---

# WebSocket

Configuring Traefik to handle WebSocket and WebSocket Secure (WSS) connections.
{: .subtitle }

## Overview

WebSocket is a communication protocol that provides full-duplex communication channels over a single TCP connection.
WebSocket Secure (WSS) is the encrypted version of WebSocket, using TLS/SSL encryption.

Traefik supports WebSocket and WebSocket Secure (WSS) out of the box. This guide will walk through examples of how to configure Traefik for different WebSocket scenarios.

## Basic WebSocket Configuration

A basic WebSocket configuration only requires defining a router and a service that points to your WebSocket server.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.routers.my-websocket.rule=Host(`ws.example.com`)"
  - "traefik.http.routers.my-websocket.service=my-websocket-service"
  - "traefik.http.services.my-websocket-service.loadbalancer.server.port=8000"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-websocket-route
spec:
  entryPoints:
    - web
  routes:
    - match: Host(`ws.example.com`)
      kind: Rule
      services:
        - name: my-websocket-service
          port: 8000
```

```yaml tab="File (YAML)"
http:
  routers:
    my-websocket:
      rule: "Host(`ws.example.com`)"
      service: my-websocket-service
  
  services:
    my-websocket-service:
      loadBalancer:
        servers:
          - url: "http://my-websocket-server:8000"
```

```toml tab="File (TOML)"
[http.routers]
  [http.routers.my-websocket]
    rule = "Host(`ws.example.com`)"
    service = "my-websocket-service"

[http.services]
  [http.services.my-websocket-service]
    [http.services.my-websocket-service.loadBalancer]
      [[http.services.my-websocket-service.loadBalancer.servers]]
        url = "http://my-websocket-server:8000"
```

## WebSocket Secure (WSS) Configuration

WebSocket Secure (WSS) requires TLS configuration.
The client connects using the `wss://` protocol instead of `ws://`.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.routers.my-websocket-secure.rule=Host(`wss.example.com`)"
  - "traefik.http.routers.my-websocket-secure.service=my-websocket-service"
  - "traefik.http.routers.my-websocket-secure.tls=true"
  - "traefik.http.services.my-websocket-service.loadbalancer.server.port=8000"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-websocket-secure-route
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`wss.example.com`)
      kind: Rule
      services:
        - name: my-websocket-service
          port: 8000
  tls: {}
```

```yaml tab="File (YAML)"
http:
  routers:
    my-websocket-secure:
      rule: "Host(`wss.example.com`)"
      service: my-websocket-service
      tls: {}
  
  services:
    my-websocket-service:
      loadBalancer:
        servers:
          - url: "http://my-websocket-server:8000"
```

```toml tab="File (TOML)"
[http.routers]
  [http.routers.my-websocket-secure]
    rule = "Host(`wss.example.com`)"
    service = "my-websocket-service"
    [http.routers.my-websocket-secure.tls]

[http.services]
  [http.services.my-websocket-service]
    [http.services.my-websocket-service.loadBalancer]
      [[http.services.my-websocket-service.loadBalancer.servers]]
        url = "http://my-websocket-server:8000"
```

## SSL Termination for WebSockets

In this scenario, clients connect to Traefik using WSS (encrypted), but Traefik connects to your backend server using WS (unencrypted).
This is called SSL termination.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.routers.my-wss-termination.rule=Host(`wss.example.com`)"
  - "traefik.http.routers.my-wss-termination.service=my-ws-service"
  - "traefik.http.routers.my-wss-termination.tls=true"
  - "traefik.http.services.my-ws-service.loadbalancer.server.port=8000"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-wss-termination-route
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`wss.example.com`)
      kind: Rule
      services:
        - name: my-ws-service
          port: 8000
  tls: {}
```

```yaml tab="File (YAML)"
http:
  routers:
    my-wss-termination:
      rule: "Host(`wss.example.com`)"
      service: my-ws-service
      tls: {}
  
  services:
    my-ws-service:
      loadBalancer:
        servers:
          - url: "http://my-ws-server:8000"
```

```toml tab="File (TOML)"
[http.routers]
  [http.routers.my-wss-termination]
    rule = "Host(`wss.example.com`)"
    service = "my-ws-service"
    [http.routers.my-wss-termination.tls]

[http.services]
  [http.services.my-ws-service]
    [http.services.my-ws-service.loadBalancer]
      [[http.services.my-ws-service.loadBalancer.servers]]
        url = "http://my-ws-server:8000"
```

## End-to-End WebSocket Secure (WSS)

For end-to-end encryption, Traefik can be configured to connect to your backend using HTTPS.

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.routers.my-wss-e2e.rule=Host(`wss.example.com`)"
  - "traefik.http.routers.my-wss-e2e.service=my-wss-service"
  - "traefik.http.routers.my-wss-e2e.tls=true"
  - "traefik.http.services.my-wss-service.loadbalancer.server.port=8443"
  # If the backend uses a self-signed certificate
  - "traefik.http.serversTransports.insecureTransport.insecureSkipVerify=true"
  - "traefik.http.services.my-wss-service.loadBalancer.serversTransport=insecureTransport"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: ServersTransport
metadata:
  name: insecure-transport
spec:
  insecureSkipVerify: true

---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-wss-e2e-route
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`wss.example.com`)
      kind: Rule
      services:
        - name: my-wss-service
          port: 8443
          serversTransport: insecure-transport
  tls: {}
```

```yaml tab="File (YAML)"
http:
  serversTransports:
    insecureTransport:
      insecureSkipVerify: true

  routers:
    my-wss-e2e:
      rule: "Host(`wss.example.com`)"
      service: my-wss-service
      tls: {}
  
  services:
    my-wss-service:
      loadBalancer:
        serversTransport: insecureTransport
        servers:
          - url: "https://my-wss-server:8443"
```

```toml tab="File (TOML)"
[http.serversTransports]
  [http.serversTransports.insecureTransport]
    insecureSkipVerify = true

[http.routers]
  [http.routers.my-wss-e2e]
    rule = "Host(`wss.example.com`)"
    service = "my-wss-service"
    [http.routers.my-wss-e2e.tls]

[http.services]
  [http.services.my-wss-service]
    [http.services.my-wss-service.loadBalancer]
      serversTransport = "insecureTransport"
      [[http.services.my-wss-service.loadBalancer.servers]]
        url = "https://my-wss-server:8443"
```

## EntryPoints Configuration for WebSockets

In your Traefik static configuration, you'll need to define entryPoints for both WS and WSS:

```yaml tab="File (YAML)"
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"
```

```toml tab="File (TOML)"
[entryPoints]
  [entryPoints.web]
    address = ":80"
  [entryPoints.websecure]
    address = ":443"
```

## Testing WebSocket Connections

You can test your WebSocket configuration using various tools:

1. Browser Developer Tools: Most modern browsers include WebSocket debugging in their developer tools.
2. WebSocket client tools like [wscat](https://github.com/websockets/wscat) or online tools like [Piesocket's WebSocket Tester](https://www.piesocket.com/websocket-tester).

Example wscat commands:

```bash
# Test standard WebSocket
wscat -c ws://ws.example.com

# Test WebSocket Secure
wscat -c wss://wss.example.com
```

## Common Issues and Solutions

### Headers and Origin Checks

Some WebSocket servers implement origin checking. Traefik passes the original headers to your backend, including the `Origin` header.

If you need to manipulate headers for WebSocket connections, you can use Traefik's Headers middleware:

```yaml tab="Docker & Swarm"
labels:
  - "traefik.http.middlewares.my-headers.headers.customrequestheaders.Origin=https://allowed-origin.com"
  - "traefik.http.routers.my-websocket.middlewares=my-headers"
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: my-headers
spec:
  headers:
    customRequestHeaders:
      Origin: "https://allowed-origin.com"

---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-websocket-route
spec:
  routes:
    - match: Host(`ws.example.com`)
      kind: Rule
      middlewares:
        - name: my-headers
      services:
        - name: my-websocket-service
          port: 8000
```

### Certificate Issues with WSS

If you're experiencing certificate issues with WSS:

1. Ensure your certificates are valid and not expired
2. For testing with self-signed certificates, configure your clients to accept them
3. When using Let's Encrypt, ensure your domain is properly configured

For backends with self-signed certificates, use the `insecureSkipVerify` option in the ServersTransport configuration as shown in the examples above.
