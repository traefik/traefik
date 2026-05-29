# Exposing Services with Traefik on Kubernetes - Basic

This guide will help you get started with exposing your services through Traefik Proxy on Kubernetes. You'll learn the fundamentals of routing HTTP traffic, setting up path-based routing, and securing your services with TLS.

For routing, this guide gives you two options:

- [Gateway API](../../reference/routing-configuration/kubernetes/gateway-api.md)
- [IngressRoute](../../reference/routing-configuration/kubernetes/crd/http/ingressroute.md)

Feel free to choose the one that fits your needs best.

## Prerequisites

- A Kubernetes cluster with Traefik Proxy installed
- `kubectl` configured to interact with your cluster
- Traefik deployed using the [Traefik Kubernetes Setup guide](../../setup/kubernetes.md)

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
          namespacePolicy:
            from: All
        websecure:
          port: 443
          protocol: HTTPS
          namespacePolicy:
            from: All
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

## Next Steps

Now that you've mastered the basics of exposing services with Traefik on Kubernetes, you're ready to explore more advanced features like middlewares, Let's Encrypt certificates, sticky sessions, and multi-layer routing.

Continue to the [Advanced Guide](advanced.md) to learn about:

- Adding middlewares for security and access control
- Generating certificates with Let's Encrypt (IngressRoute) or cert-manager (Gateway API)
- Configuring sticky sessions for stateful applications
- Setting up multi-layer routing for authentication-based routing with IngressRoute
