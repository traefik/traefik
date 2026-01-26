# Exposing Services with Traefik Proxy

This section guides you through exposing services securely with Traefik Proxy. You'll learn how to route HTTP and HTTPS traffic to your services, add security features, and implement advanced load balancing.

## What You'll Accomplish

Following these guides, you'll learn how to:

- Route HTTP traffic to your services with [Gateway API](../reference/routing-configuration/kubernetes/gateway-api.md) and [IngressRoute](../reference/routing-configuration/kubernetes/crd/http/ingressroute.md)
- Configure routing rules to direct requests
- Enable HTTPS with TLS
- Add security middlewares
- Generate certificates automatically with Let's Encrypt
- Implement sticky sessions for session persistence

## Platform-Specific Guides

For detailed steps tailored to your environment, follow the guide for your platform:

- [Kubernetes](./kubernetes.md)
- [Docker](./docker.md)
- [Docker Swarm](./swarm.md)

## Advanced Use Cases

### Exposing gRPC Services

Traefik Proxy supports gRPC applications without requiring specific configuration. You can expose gRPC services using either HTTP (h2c) or HTTPS.

??? example "Using HTTP (h2c)"

    For unencrypted gRPC communication, configure your service to use the `h2c://` protocol scheme:

    ```yaml
    http:
      routers:
        grpc-router:
          service: grpc-service
          rule: Host(`grpc.example.com`)
      
      services:
        grpc-service:
          loadBalancer:
            servers:
              - url: h2c://backend:8080
    ```

    !!! note
        For providers with labels (Docker, Kubernetes), specify the scheme using:
        `traefik.http.services.<service-name>.loadbalancer.server.scheme=h2c`

??? example "Using HTTPS"

    For encrypted gRPC communication, use standard HTTPS URLs. Traefik will use HTTP/2 over TLS to communicate with your gRPC backend:

    ```yaml
    http:
      routers:
        grpc-router:
          service: grpc-service
          rule: Host(`grpc.example.com`)
          tls: {}
      
      services:
        grpc-service:
          loadBalancer:
            servers:
              - url: https://backend:8080
    ```

    Traefik handles the protocol negotiation automatically. Configure TLS certificates for your backends using [ServersTransport](../reference/routing-configuration/http/load-balancing/serverstransport.md) if needed.

### Exposing WebSocket Services

Traefik Proxy supports WebSocket (WS) and WebSocket Secure (WSS) connections out of the box. No special configuration is required beyond standard HTTP routing.

??? example "Basic WebSocket"

    Configure a router and service pointing to your WebSocket server. Traefik automatically detects and handles the WebSocket upgrade:

    ```yaml
    http:
      routers:
        websocket-router:
          rule: Host(`ws.example.com`)
          service: websocket-service
      
      services:
        websocket-service:
          loadBalancer:
            servers:
              - url: http://websocket-backend:8000
    ```

??? example "WebSocket Secure (WSS)"

    For encrypted WebSocket connections, enable TLS on your router. Clients connect using `wss://` while you can choose whether backends use encrypted or unencrypted connections:

    ```yaml
    http:
      routers:
        websocket-secure-router:
          rule: Host(`wss.example.com`)
          service: websocket-service
          tls: {}
      
      services:
        websocket-service:
          loadBalancer:
            servers:
              - url: http://websocket-backend:8000  # SSL termination at Traefik
              # OR
              # - url: https://websocket-backend:8443  # End-to-end encryption
    ```

    Traefik preserves WebSocket headers including `Origin`, `Sec-WebSocket-Key`, and `Sec-WebSocket-Version`. Use the [Headers middleware](../reference/routing-configuration/http/middlewares/headers.md) if you need to modify headers for origin checking or other requirements.
