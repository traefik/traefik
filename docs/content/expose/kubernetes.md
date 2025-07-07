# Exposing Services with Traefik on Kubernetes

This guide will help you expose your services securely through Traefik Proxy on Kubernetes. We'll cover routing HTTP and HTTPS traffic, implementing TLS, adding security middleware, and configuring sticky sessions. For routing, this guide gives you two options:

- [Gateway API](../reference/routing-configuration/kubernetes/gateway-api.md)
- [IngressRoute](../reference/routing-configuration/kubernetes/crd/http/ingressroute.md)

Feel free to choose the one that fits your needs best.

## Prerequisites

- A Kubernetes cluster with Traefik Proxy installed
- `kubectl` configured to interact with your cluster
- Traefik deployed using the Traefik Kubernetes Setup guide

## Expose Your First HTTP Service

Let's expose a simple HTTP service using the [whoami](https://github.com/traefik/whoami) application. This will demonstrate basic routing to a backend service.

First, create the deployment and service:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - name: whoami
        image: traefik/whoami
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: default
spec:
  selector:
    app: whoami
  ports:
  - port: 80
```

Save this as `whoami.yaml` and apply it:

```bash
kubectl apply -f whoami.yaml
```

Now, let's create routes using either Gateway API or IngressRoute.

### Using Gateway API

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami
  namespace: default
spec:
  parentRefs:
  - name: traefik-gateway  # This Gateway is automatically created by Traefik 
  hostnames:
  - "whoami.docker.localhost"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: whoami
      port: 80
```

Save this as `whoami-route.yaml` and apply it:

```bash
kubectl apply -f whoami-route.yaml
```

### Using IngressRoute

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami
  namespace: default
spec:
  entryPoints:
    - web
  routes:
  - match: Host(`whoami.docker.localhost`)
    kind: Rule
    services:
    - name: whoami
      port: 80
```

Save this as `whoami-ingressroute.yaml` and apply it:

```bash
kubectl apply -f whoami-ingressroute.yaml
```

### Verify Your Service

Your service is now available at http://whoami.docker.localhost/. Test that it works:

```bash
curl -H "Host: whoami.docker.localhost" http://localhost/
```

!!! info
    Make sure to remove the `ports.web.redirections` block from the `values.yaml` file if you followed the Kubernetes Setup Guide to install Traefik otherwise you will be redirected to the HTTPS entrypoint:

    ```yaml
    redirections:
      entryPoint:
        to: websecure
    ```

You should see output similar to:

```bash
Hostname: whoami-6d5d964cb-8pv4k
IP: 127.0.0.1
IP: ::1
IP: 10.42.0.18
IP: fe80::d4c0:3bff:fe20:b0a3
RemoteAddr: 10.42.0.17:39872
GET / HTTP/1.1
Host: whoami.docker.localhost
User-Agent: curl/7.68.0
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 10.42.0.1
X-Forwarded-Host: whoami.docker.localhost
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: traefik-76cbd5b89c-rx5xn
X-Real-Ip: 10.42.0.1
```

This confirms that Traefik is successfully routing requests to your whoami application.

## Add Routing Rules

Now we'll enhance our routing by directing traffic to different services based on URL paths. This is useful for API versioning, frontend/backend separation, or organizing microservices.

First, deploy a second service to represent an API:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami-api
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: whoami-api
  template:
    metadata:
      labels:
        app: whoami-api
    spec:
      containers:
      - name: whoami
        image: traefik/whoami
        env:
        - name: WHOAMI_NAME
          value: "API Service"
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: whoami-api
  namespace: default
spec:
  selector:
    app: whoami-api
  ports:
  - port: 80
```

Save this as `whoami-api.yaml` and apply it:

```bash
kubectl apply -f whoami-api.yaml
```

Now set up path-based routing:

### Gateway API with Path Rules

Update your existing `HTTPRoute` to include path-based routing:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami
  namespace: default
spec:
  parentRefs:
  - name: traefik-gateway
  hostnames:
  - "whoami.docker.localhost"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: whoami-api
      port: 80
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: whoami
      port: 80
```

Update the file `whoami-route.yaml` and apply it:

```bash
kubectl apply -f whoami-route.yaml
```

### IngressRoute with Path Rules

Update your existing IngressRoute to include path-based routing:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami
  namespace: default
spec:
  entryPoints:
    - web
  routes:
  - match: Host(`whoami.docker.localhost`) && Path(`/api`)
    kind: Rule
    services:
    - name: whoami-api
      port: 80
  - match: Host(`whoami.docker.localhost`)
    kind: Rule
    services:
    - name: whoami
      port: 80
```

Save this as `whoami-ingressroute.yaml` and apply it:

```bash
kubectl apply -f whoami-ingressroute.yaml
```

### Test the Path-Based Routing

Verify that different paths route to different services:

```bash
# Root path should go to the main whoami service
curl -H "Host: whoami.docker.localhost" http://localhost/

# /api path should go to the whoami-api service
curl -H "Host: whoami.docker.localhost" http://localhost/api
```

For the `/api` requests, you should see the response showing "API Service" in the environment variables section, confirming that your path-based routing is working correctly:

```bash
{"hostname":"whoami-api-67d97b4868-dvvll","ip":["127.0.0.1","::1","10.42.0.9","fe80::10aa:37ff:fe74:31f2"],"headers":{"Accept":["*/*"],"Accept-Encoding":["gzip"],"User-Agent":["curl/8.7.1"],"X-Forwarded-For":["10.42.0.1"],"X-Forwarded-Host":["whoami.docker.localhost"],"X-Forwarded-Port":["80"],"X-Forwarded-Proto":["http"],"X-Forwarded-Server":["traefik-669c479df8-vkj22"],"X-Real-Ip":["10.42.0.1"]},"url":"/api","host":"whoami.docker.localhost","method":"GET","name":"API Service","remoteAddr":"10.42.0.13:36592"}
```

## Enable TLS

Let's secure our service with HTTPS by adding TLS. We'll start with a self-signed certificate for local development.

### Create a Self-Signed Certificate
 
Generate a self-signed certificate:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt \
  -subj "/CN=whoami.docker.localhost"
```

Create a TLS secret in Kubernetes:

```bash
kubectl create secret tls whoami-tls --cert=tls.crt --key=tls.key
```

!!! important "Prerequisite for Gateway API with TLS"
    Before using the Gateway API with TLS, you must define the `websecure` listener in your Traefik installation. This is typically done in your Helm values.
    
    Example configuration in `values.yaml`:
    ```yaml
    gateway:
      listeners:
        web:
          port: 80
          protocol: HTTP
          namespacePolicy: All
        websecure:
          port: 443
          protocol: HTTPS
          namespacePolicy: All
          mode: Terminate
          certificateRefs:
            - kind: Secret
              name: local-selfsigned-tls
              group: ""
    ```
    
    See the Traefik Kubernetes Setup Guide for complete installation details.

### Gateway API with TLS

Update your existing `HTTPRoute` to use the secured gateway listener:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami
  namespace: default
spec:
  parentRefs:
  - name: traefik-gateway
    sectionName: websecure  # The HTTPS listener
  hostnames:
  - "whoami.docker.localhost"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: whoami-api
      port: 80
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: whoami
      port: 80
```

Update the file `whoami-route.yaml` and apply it:

```bash
kubectl apply -f whoami-route.yaml
```

### IngressRoute with TLS

Update your existing IngressRoute to use TLS:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami
  namespace: default
spec:
  entryPoints:
    - websecure  # Changed from 'web' to 'websecure'
  routes:
  - match: Host(`whoami.docker.localhost`) && Path(`/api`)
    kind: Rule
    services:
    - name: whoami-api
      port: 80
  - match: Host(`whoami.docker.localhost`)
    kind: Rule
    services:
    - name: whoami
      port: 80
  tls:
    secretName: whoami-tls  # Added TLS configuration
```

Update the file `whoami-ingressroute.yaml` and apply it:

```bash
kubectl apply -f whoami-ingressroute.yaml
```

### Verify HTTPS Access

Now you can access your service securely. Since we're using a self-signed certificate, you'll need to skip certificate verification:

```bash
curl -k -H "Host: whoami.docker.localhost" https://localhost/
```

Your browser can also access https://whoami.docker.localhost/ (you'll need to accept the security warning for the self-signed certificate).

## Add Middlewares

Middlewares allow you to modify requests or responses as they pass through Traefik. Let's add two useful middlewares: [Headers](../reference/routing-configuration/http/middlewares/headers.md) for security and [IP allowlisting](../reference/routing-configuration/http/middlewares/ipallowlist.md) for access control.

### Create Middlewares

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: secure-headers
  namespace: default
spec:
  headers:
    frameDeny: true
    sslRedirect: true
    browserXssFilter: true
    contentTypeNosniff: true
    stsIncludeSubdomains: true
    stsPreload: true
    stsSeconds: 31536000
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: ip-allowlist
  namespace: default
spec:
  ipAllowList:
    sourceRange:
      - 127.0.0.1/32
      - 10.0.0.0/8  # Typical cluster network range
      - 192.168.0.0/16  # Common local network range
```

Save this as `middlewares.yaml` and apply it:

```bash
kubectl apply -f middlewares.yaml
```

### Apply Middlewares with Gateway API

In Gateway API, you can apply middlewares using the `ExtensionRef` filter type. This is the preferred and standard way to use Traefik middlewares with Gateway API, as it integrates directly with the HTTPRoute specification.

First, make sure you have the same middlewares defined:

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: secure-headers
  namespace: default
spec:
  headers:
    frameDeny: true
    sslRedirect: true
    browserXssFilter: true
    contentTypeNosniff: true
    stsIncludeSubdomains: true
    stsPreload: true
    stsSeconds: 31536000
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: ip-allowlist
  namespace: default
spec:
  ipAllowList:
    sourceRange:
      - 127.0.0.1/32
      - 10.0.0.0/8  # Typical cluster network range
      - 192.168.0.0/16  # Common local network range
```

Now, update your `HTTPRoute` to reference these middlewares using the `ExtensionRef` filter:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami
  namespace: default
spec:
  parentRefs:
  - name: traefik-gateway
    sectionName: websecure
  hostnames:
  - "whoami.docker.localhost"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    filters:
    - type: ExtensionRef
      extensionRef:  # Headers Middleware Definition
        group: traefik.io
        kind: Middleware
        name: secure-headers
    - type: ExtensionRef 
      extensionRef: # IP AllowList Middleware Definition
        group: traefik.io
        kind: Middleware
        name: ip-allowlist
    backendRefs:
    - name: whoami-api
      port: 80
  - matches:
    - path:
        type: PathPrefix
        value: /
    filters:
    - type: ExtensionRef
      extensionRef:  # Headers Middleware Definition
        group: traefik.io
        kind: Middleware
        name: secure-headers
    - type: ExtensionRef 
      extensionRef: # IP AllowList Middleware Definition
        group: traefik.io
        kind: Middleware
        name: ip-allowlist
    backendRefs:
    - name: whoami
      port: 80
```

Update the file `whoami-route.yaml` and apply it:

```bash
kubectl apply -f whoami-route.yaml
```

This approach uses the Gateway API's native filter mechanism rather than annotations. The `ExtensionRef` filter type allows you to reference Traefik middlewares directly within the HTTPRoute specification, which is more consistent with the Gateway API design principles.

### Apply Middlewares with IngressRoute

Update your existing IngressRoute to include middlewares:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami
  namespace: default
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`whoami.docker.localhost`) && Path(`/api`)
    kind: Rule
    middlewares: # Middleware Definition
    - name: secure-headers
    - name: ip-allowlist
    services:
    - name: whoami-api
      port: 80
  - match: Host(`whoami.docker.localhost`)
    kind: Rule
    middlewares: # Middleware Definition
    - name: secure-headers
    - name: ip-allowlist
    services:
    - name: whoami
      port: 80
  tls:
    certResolver: le
```

Update the file `whoami-ingressroute.yaml` and apply it:

```bash
kubectl apply -f whoami-ingressroute.yaml
```

### Verify Middleware Effects

Check that the security headers are being applied:

```bash
curl -k -I -H "Host: whoami.docker.localhost" https://localhost/
```

You should see security headers in the response, such as:

```bash
HTTP/2 200
x-content-type-options: nosniff
x-frame-options: DENY
x-xss-protection: 1; mode=block
strict-transport-security: max-age=31536000; includeSubDomains; preload
content-type: text/plain; charset=utf-8
content-length: 403
```

To test the IP allowlist, you can modify the `sourceRange` in the middleware to exclude your IP and verify that access is blocked.

## Generate Certificates with Let's Encrypt

!!! info
    Traefik's built-in Let's Encrypt integration works with IngressRoute but does not automatically issue certificates for Gateway API listeners. For Gateway API, you should use cert-manager or another certificate controller.

### Using IngressRoute with Let's Encrypt

Configure a certificate resolver in your Traefik values.yaml:

```yaml
additionalArguments:
  - "--certificatesresolvers.le.acme.email=your-email@example.com" #replace with your email
  - "--certificatesresolvers.le.acme.storage=/data/acme.json"
  - "--certificatesresolvers.le.acme.httpchallenge.entrypoint=web"
```

!!! important "Public DNS Required"
    Let's Encrypt may require a publicly accessible domain to validate domain ownership. For testing with local domains like `whoami.docker.localhost`, the certificate will remain self-signed. In production, replace it with a real domain that has a publicly accessible DNS record pointing to your Traefik instance.

Update your Traefik installation with this configuration:

```bash
helm upgrade traefik traefik/traefik -n traefik --reuse-values -f values.yaml
```

Update your IngressRoute with the Let's Encrypt certificate:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami
  namespace: default
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`whoami.docker.localhost`) && Path(`/api`)
    kind: Rule
    middlewares:
    - name: secure-headers
    - name: ip-allowlist
    services:
    - name: whoami-api
      port: 80
  - match: Host(`whoami.docker.localhost`)
    kind: Rule
    middlewares:
    - name: secure-headers
    - name: ip-allowlist
    services:
    - name: whoami
      port: 80
  tls:
    certResolver: le
```

Apply it:

```bash
kubectl apply -f whoami-ingressroute.yaml
```

### Using Gateway API with cert-manager

For Gateway API, install cert-manager:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
```

Create an Issuer & Certificate:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt
spec:
  acme:
    email: your-email@example.com # replace with your email
    server: https://acme-v02-staging.api.letsencrypt.org/directory # Replace with the production server in production
    privateKeySecretRef:
      name: letsencrypt-account-key
    solvers:
      - http01:
          gatewayHTTPRoute:
            parentRefs:
              - name: traefik
                namespace: default
                kind: Gateway
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: whoami
  namespace: default
spec:
  secretName: whoami-tls-le        # Name of secret where the generated certificate will be stored.
  dnsNames:
    - "whoami.docker.localhost" # Replace a real domain
  issuerRef:
    name: letsencrypt
    kind: Issuer
```

!!! important "Public DNS Required"
    Let's Encrypt requires a publicly accessible domain to verify ownership. When using a local domain like `whoami.docker.localhost`, cert-manager will attempt the challenge but it will fail, and the certificate will remain self-signed. For production use, replace the domain with one that has a public DNS record pointing to your cluster's ingress point.

Save the YAML file and apply:

```bash
kubectl apply -f letsencrypt-issuer-andwhoami-certificate.yaml
```

Now, update your Gateway to use the generated certificate:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: traefik-gateway
  namespace: default
spec:
  gatewayClassName: traefik
  listeners:
  - name: web
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: All
  - name: websecure
    port: 443
    protocol: HTTPS
    allowedRoutes:
      namespaces:
        from: All
    tls:
      certificateRefs:
      - name: whoami-tls-le  # References the secret created by cert-manager
```

Apply the updated Gateway:

```bash
kubectl apply -f gateway.yaml
```

Your existing `HTTPRoute` will now use this certificate when connecting to the secured gateway listener.

### Verify the Let's Encrypt Certificate

Once the certificate is issued, you can verify it:

```bash
# Check certificate status
kubectl get certificate -n default

# Verify the certificate chain
curl -v https://whoami.docker.localhost/ 2>&1 | grep -i "server certificate"
```

You should see that your certificate is issued by Let's Encrypt.

## Configure Sticky Sessions

Sticky sessions ensure that a user's requests always go to the same backend server, which is essential for applications that maintain session state. Let's implement sticky sessions for our whoami service.

### First, Scale Up the Deployment

To demonstrate sticky sessions, first scale up the deployment to 3 replicas:

```bash
kubectl scale deployment whoami --replicas=3
```

### Using Gateway API with TraefikService

First, create the `TraefikService` for sticky sessions:

```yaml
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: whoami-sticky
  namespace: default
spec:
  weighted:
    services:
      - name: whoami
        port: 80
        weight: 1
    sticky:
      cookie:
        name: sticky_cookie
        secure: true
        httpOnly: true
```

Save this as `whoami-sticky-service.yaml` and apply it:

```bash
kubectl apply -f whoami-sticky-service.yaml
```

Now update your `HTTPRoute` with an annotation referencing the `TraefikService`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: whoami
  namespace: default
spec:
  parentRefs:
  - name: traefik-gateway
    sectionName: websecure
  hostnames:
  - "whoami.docker.localhost"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    filters:
    - type: ExtensionRef
      extensionRef:  # Headers Middleware Definition
        group: traefik.io
        kind: Middleware
        name: secure-headers
    - type: ExtensionRef 
      extensionRef: # IP AllowList Middleware Definition
        group: traefik.io
        kind: Middleware
        name: ip-allowlist
    backendRefs:
    - name: whoami-api
      port: 80
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - group: traefik.io          # <── tell Gateway this is a TraefikService
      kind: TraefikService
      name: whoami-sticky
    filters:
    - type: ExtensionRef
      extensionRef:  # Headers Middleware Definition
        group: traefik.io
        kind: Middleware
        name: secure-headers
    - type: ExtensionRef 
      extensionRef: # IP AllowList Middleware Definition
        group: traefik.io
        kind: Middleware
        name: ip-allowlist
    backendRefs:
    - name: whoami
      port: 80
```

Update the file `whoami-route.yaml` and apply it:

```bash
kubectl apply -f whoami-route.yaml
```

### Using IngressRoute with TraefikService

First, create the `TraefikService` for sticky sessions:

```yaml
apiVersion: traefik.io/v1alpha1
kind: TraefikService
metadata:
  name: whoami-sticky
  namespace: default
spec:
  weighted:
    services:
    - name: whoami
      port: 80
      sticky:
        cookie:
          name: sticky_cookie
          secure: true
          httpOnly: true
```

Save this as `whoami-sticky-service.yaml` and apply it:

```bash
kubectl apply -f whoami-sticky-service.yaml
```

Now update your IngressRoute to use this `TraefikService`:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: whoami
  namespace: default
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`whoami.docker.localhost`) && Path(`/api`)
    kind: Rule
    middlewares: # Middleware Definition
    - name: secure-headers
    - name: ip-allowlist
    services:
    - name: whoami-api
      port: 80
  - match: Host(`whoami.docker.localhost`)
    kind: Rule
    middlewares: # Middleware Definition
    - name: secure-headers
    - name: ip-allowlist
    services:
    - name: whoami-sticky  # Changed from whoami to whoami-sticky
      kind: TraefikService  # Added kind: TraefikService
  tls:
    certResolver: le
```

Update the file `whoami-ingressroute.yaml` and apply it:

```bash
kubectl apply -f whoami-ingressroute.yaml
```

### Test Sticky Sessions

You can test the sticky sessions by making multiple requests and observing that they all go to the same backend pod:

```bash
# First request - save cookies to a file
curl -k -c cookies.txt -H "Host: whoami.docker.localhost" https://localhost/

# Subsequent requests - use the cookies
curl -k -b cookies.txt -H "Host: whoami.docker.localhost" https://localhost/
curl -k -b cookies.txt -H "Host: whoami.docker.localhost" https://localhost/
```

Pay attention to the `Hostname` field in each response - it should remain the same across all requests when using the cookie file, confirming that sticky sessions are working.

For comparison, try making requests without the cookie:

```bash
# Requests without cookies should be load-balanced across different pods
curl -k -H "Host: whoami.docker.localhost" https://localhost/
curl -k -H "Host: whoami.docker.localhost" https://localhost/
```

You should see different `Hostname` values in these responses, as each request is load-balanced to a different pod.

!!! important "Browser Testing"
    When testing in browsers, you need to use the same browser session to maintain the cookie. The cookie is set with `httpOnly` and `secure` flags for security, so it will only be sent over HTTPS connections and won't be accessible via JavaScript.

For more advanced configuration options, see the [reference documentation](../reference/routing-configuration/http/load-balancing/service.md).

## Conclusion

In this guide, you've learned how to:

- Expose HTTP services through Traefik in Kubernetes using both Gateway API and IngressRoute
- Set up path-based routing to direct traffic to different backend services
- Secure your services with TLS using self-signed certificates
- Add security with middlewares like secure headers and IP allow listing
- Automate certificate management with Let's Encrypt
- Implement sticky sessions for stateful applications

These fundamental capabilities provide a solid foundation for exposing any application through Traefik Proxy in Kubernetes. Each of these can be further customized to meet your specific requirements.

### Next Steps

Now that you understand the basics of exposing services with Traefik Proxy, you might want to explore:

- [Advanced routing options](../reference/routing-configuration/http/router/rules-and-priority.md) like query parameter matching, header-based routing, and more
- [Additional middlewares](../reference/routing-configuration/http/middlewares/overview.md) for authentication, rate limiting, and request modifications
- [Observability features](../reference/install-configuration/observability/metrics.md) for monitoring and debugging your Traefik deployment
- [TCP services](../reference/routing-configuration/tcp/service.md) for exposing TCP services
- [UDP services](../reference/routing-configuration/udp/service.md) for exposing UDP services
- [Kubernetes Provider documentation](../reference/install-configuration/providers/kubernetes/kubernetes-crd.md) for more details about the Kubernetes integration.
- [Gateway API provider documentation](../reference/install-configuration/providers/kubernetes/kubernetes-gateway.md) for more details about the Gateway API integration.
