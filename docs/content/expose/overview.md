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

- [Kubernetes](./kubernetes/basic.md)
- [Docker](./docker/basic.md)
- [Docker Swarm](./swarm/basic.md)

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

### Combining Middlewares with the Chain Middleware

The [Chain middleware](../../reference/routing-configuration/http/middlewares/chain.md) lets you combine multiple middlewares into a single, reusable stack. This is the most common production pattern for API gateways: combining authentication, rate limiting, and CORS headers.

??? example "API Gateway Middleware Stack (YAML)"

    ```yaml
    http:
      routers:
        api-router:
          service: api-service
          middlewares:
            - api-stack
          rule: "Host(`api.example.com`)"

      middlewares:
        api-stack:
          chain:
            middlewares:
              - api-auth
              - api-ratelimit
              - api-cors

        api-auth:
          basicAuth:
            users:
              - "admin:$2y$10$..."  # bcrypt hash

        api-ratelimit:
          rateLimit:
            average: 100
            burst: 50

        api-cors:
          headers:
            accessControlAllowMethods:
              - GET
              - POST
              - OPTIONS
            accessControlAllowOriginList:
              - "https://app.example.com"

      services:
        api-service:
          loadBalancer:
            servers:
              - url: "http://127.0.0.1:8080"
    ```

??? example "API Gateway Middleware Stack (TOML)"

    ```toml
    [http.routers]
      [http.routers.api-router]
        service = "api-service"
        middlewares = ["api-stack"]
        rule = "Host(`api.example.com`)"

    [http.middlewares]
      [http.middlewares.api-stack.chain]
        middlewares = ["api-auth", "api-ratelimit", "api-cors"]

      [http.middlewares.api-auth.basicAuth]
        users = ["admin:$2y$10$..."]

      [http.middlewares.api-ratelimit.rateLimit]
        average = 100
        burst = 50

      [http.middlewares.api-cors.headers]
        accessControlAllowMethods = ["GET", "POST", "OPTIONS"]
        accessControlAllowOriginList = ["https://app.example.com"]

    [http.services]
      [http.services.api-service]
        [http.services.api-service.loadBalancer]
          [[http.services.api-service.loadBalancer.servers]]
            url = "http://127.0.0.1:8080"
    ```

??? example "API Gateway Middleware Stack (Docker Labels)"

    ```yaml
    labels:
      - "traefik.http.routers.api-router.service=api-service"
      - "traefik.http.routers.api-router.middlewares=api-stack"
      - "traefik.http.routers.api-router.rule=Host(`api.example.com`)"
      - "traefik.http.middlewares.api-stack.chain.middlewares=api-auth,api-ratelimit,api-cors"
      - "traefik.http.middlewares.api-auth.basicauth.users=admin:$2y$10$..."
      - "traefik.http.middlewares.api-ratelimit.ratelimit.average=100"
      - "traefik.http.middlewares.api-ratelimit.ratelimit.burst=50"
      - "traefik.http.middlewares.api-cors.headers.accesscontrolallowmethods=GET,POST,OPTIONS"
      - "traefik.http.middlewares.api-cors.headers.accesscontrolalloworiginlist=https://app.example.com"
      - "traefik.http.services.api-service.loadbalancer.server.port=8080"
    ```

??? example "API Gateway Middleware Stack (Kubernetes)"

    ```yaml
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: api-ingress
      namespace: default
    spec:
      entryPoints:
        - web
      routes:
        - match: Host(`api.example.com`)
          kind: Rule
          services:
            - name: api-service
              port: 8080
          middlewares:
            - name: api-stack
    ---
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: api-stack
    spec:
      chain:
        middlewares:
        - name: api-auth
        - name: api-ratelimit
        - name: api-cors
    ---
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: api-auth
    spec:
      basicAuth:
        users:
        - admin:$2y$10$...
    ---
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: api-ratelimit
    spec:
      rateLimit:
        average: 100
        burst: 50
    ---
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: api-cors
    spec:
      headers:
        accessControlAllowMethods:
          - GET
          - POST
          - OPTIONS
        accessControlAllowOriginList:
          - https://app.example.com
    ```

This pattern keeps your routing rules clean by abstracting security concerns into reusable middleware stacks. For a complete reference of all available middlewares, see the [Middlewares reference](../../reference/routing-configuration/http/middlewares/overview.md).
