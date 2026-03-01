# Exposing Services with Traefik on Kubernetes - Advanced

This guide builds on the concepts and setup from the [Basic Guide](basic.md). Make sure you've completed the basic guide and have a working Traefik setup with Kubernetes before proceeding.

In this advanced guide, you'll learn how to enhance your Traefik deployment with:

- **Middlewares** for security headers and access control
- **Let's Encrypt** for automated certificate management (IngressRoute)
- **cert-manager** for automated certificate management (Gateway API)
- **Sticky sessions** for stateful applications
- **Multi-layer routing** for hierarchical routing with complex authentication scenarios (IngressRoute only)

## Prerequisites

- Completed the [Basic Guide](basic.md)
- A Kubernetes cluster with Traefik Proxy installed
- `kubectl` configured to interact with your cluster
- Working Traefik setup from the basic guide

## Add Middlewares

Middlewares allow you to modify requests or responses as they pass through Traefik. Let's add two useful middlewares: [Headers](../../reference/routing-configuration/http/middlewares/headers.md) for security and [IP allowlisting](../../reference/routing-configuration/http/middlewares/ipallowlist.md) for access control.

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

For more advanced configuration options, see the [reference documentation](../../reference/routing-configuration/http/load-balancing/service.md).

## Setup Multi-Layer Routing

Multi-layer routing enables hierarchical relationships between routers, where parent routers can process requests through middleware before child routers make final routing decisions. This is particularly useful for authentication-based routing or staged middleware application.

!!! info "IngressRoute Support"
    Multi-layer routing is **natively supported** by Kubernetes IngressRoute (CRD) using the `spec.parentRefs` field. This feature is not available when using standard Kubernetes Ingress or Gateway API resources.

### Authentication-Based Routing Example

Let's create a multi-layer routing setup where a parent IngressRoute authenticates requests, and child IngressRoutes direct traffic based on user roles.

!!! important "Parent Router Requirements"
    Parent routers in multi-layer routing must not have a service defined. The child routers will handle the service selection based on their matching rules. Make sure all child IngressRoutes reference the parent correctly using `parentRefs`.

First, deploy your backend services:

```yaml
# whoami-backends.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: admin-backend
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: admin-backend
  template:
    metadata:
      labels:
        app: admin-backend
    spec:
      containers:
      - name: whoami
        image: traefik/whoami
        env:
        - name: WHOAMI_NAME
          value: "Admin Backend"
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: admin-backend
  namespace: default
spec:
  selector:
    app: admin-backend
  ports:
  - port: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-backend
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: user-backend
  template:
    metadata:
      labels:
        app: user-backend
    spec:
      containers:
      - name: whoami
        image: traefik/whoami
        env:
        - name: WHOAMI_NAME
          value: "User Backend"
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: user-backend
  namespace: default
spec:
  selector:
    app: user-backend
  ports:
  - port: 80
```

Apply the backend services:

```bash
kubectl apply -f whoami-backends.yaml
```

Now create the middleware and IngressRoutes for multi-layer routing:

```yaml
# mlr-ingressroute.yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-secret
  namespace: default
type: Opaque
stringData:
  users: |
    admin:$apr1$DmXR3Add$wfdbGw6RWIhFb0ffXMM4d0
    user:$apr1$GJtcIY1o$mSLdsWYeXpPHVsxGDqadI.
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: auth-middleware
  namespace: default
spec:
  basicAuth:
    secret: auth-secret
    headerField: X-Auth-User
---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: api-parent
  namespace: default
spec:
  entryPoints:
    - websecure
  routes:
  - match: Host(`api.docker.localhost`) && PathPrefix(`/api`)
    kind: Rule
    middlewares:
    - name: auth-middleware
  # Note: No services and no TLS config - this is a parent IngressRoute
---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: api-admin
  namespace: default
spec:
  parentRefs:
  - name: api-parent
    namespace: default  # Optional, defaults to same namespace
  routes:
  - match: HeadersRegexp(`X-Auth-User`, `admin`)
    kind: Rule
    services:
    - name: admin-backend
      port: 80
---
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: api-user
  namespace: default
spec:
  parentRefs:
  - name: api-parent
    namespace: default  # Optional, defaults to same namespace
  routes:
  - match: HeadersRegexp(`X-Auth-User`, `user`)
    kind: Rule
    services:
    - name: user-backend
      port: 80
```

!!! note "Generating Password Hashes"
    The password hashes above are generated using `htpasswd`. To create your own user credentials:

    ```bash
    # Using htpasswd (Apache utils)
    htpasswd -nb admin yourpassword
    ```

Apply the multi-layer routing configuration:

```bash
kubectl apply -f mlr-ingressroute.yaml
```

### Test Multi-Layer Routing

Test the routing behavior:

```bash
# Request goes through parent router → auth middleware → admin child router
curl -k -u admin:test -H "Host: api.docker.localhost" https://localhost/api
```

You should see the response from the admin-backend service when authenticating as `admin`. Try with `user:test` credentials to reach the user-backend service instead.

### How It Works

1. **Request arrives** at `api.docker.localhost/api`
2. **Parent IngressRoute** (`api-parent`) matches based on host and path
3. **BasicAuth middleware** authenticates the user and sets the `X-Auth-User` header with the username
4. **Child IngressRoute** (`api-admin` or `api-user`) matches based on the header value
5. **Request forwarded** to the appropriate Kubernetes service

### Cross-Namespace Parent References

You can reference parent IngressRoutes in different namespaces by specifying the `namespace` field:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: api-child
  namespace: app-namespace
spec:
  parentRefs:
  - name: api-parent
    namespace: shared-namespace  # Parent in different namespace
  routes:
  - match: Path(`/child`)
    kind: Rule
    services:
    - name: child-service
      port: 80
```

!!! important "Cross-Namespace Requirement"
    To use cross-namespace parent references, you must enable the `allowCrossNamespace` option in your Traefik Helm values:

    ```yaml
    providers:
      kubernetesCRD:
        allowCrossNamespace: true
    ```

### Multiple Parent References

Child IngressRoutes can reference multiple parent IngressRoutes:

```yaml
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: api-child
  namespace: default
spec:
  parentRefs:
  - name: parent-one
  - name: parent-two
  routes:
  - match: Path(`/api`)
    kind: Rule
    services:
    - name: child-service
      port: 80
```

For more details about multi-layer routing, see the [Multi-Layer Routing documentation](../../reference/routing-configuration/http/routing/multi-layer-routing.md).

## Conclusion

In this advanced guide, you've learned how to:

- Add security with middlewares like secure headers and IP allow listing
- Automate certificate management with Let's Encrypt (IngressRoute) and cert-manager (Gateway API)
- Implement sticky sessions for stateful applications
- Setup multi-layer routing for authentication-based routing (IngressRoute only)

These advanced capabilities allow you to build production-ready Traefik deployments with Kubernetes. Each of these can be further customized to meet your specific requirements.

### Next Steps

Now that you've mastered both basic and advanced Traefik features with Kubernetes, you might want to explore:

- [Advanced routing options](../../reference/routing-configuration/http/routing/rules-and-priority.md) like query parameter matching, header-based routing, and more
- [Additional middlewares](../../reference/routing-configuration/http/middlewares/overview.md) for authentication, rate limiting, and request modifications
- [Observability features](../../reference/install-configuration/observability/metrics.md) for monitoring and debugging your Traefik deployment
- [TCP services](../../reference/routing-configuration/tcp/service.md) for exposing TCP services
- [UDP services](../../reference/routing-configuration/udp/service.md) for exposing UDP services
- [Kubernetes Provider documentation](../../reference/install-configuration/providers/kubernetes/kubernetes-crd.md) for more details about the Kubernetes integration
- [Gateway API provider documentation](../../reference/install-configuration/providers/kubernetes/kubernetes-gateway.md) for more details about the Gateway API integration
