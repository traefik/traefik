---
title: "Traefik Kubernetes Ingress NGINX Routing Configuration"
description: "Understand the routing configuration for the Kubernetes Ingress NGINX Controller and Traefik Proxy. Read the technical documentation."
---

# Traefik & Ingresses with NGINX Annotations

The experimental Kubernetes Controller for Ingresses with NGINX annotations.
{: .subtitle }

!!! warning "Ingress Discovery"

    The Kubernetes Ingress NGINX provider is discovering by default all Ingresses in the cluster,
    which may lead to duplicated routers if you are also using the Kubernetes Ingress provider.
    We recommend to use IngressClass for the Ingresses you want to be handled by this provider,
    or to use the `watchNamespace` or `watchNamespaceSelector` options to limit the discovery of Ingresses to a specific namespace or set of namespaces.

## Routing Configuration

The Kubernetes Ingress NGINX provider watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn will create the resulting routers, services, handlers, etc.

## Configuration Example

??? example "Configuring Kubernetes Ingress NGINX Controller"

      ```yaml tab="RBAC"
      ---
      apiVersion: rbac.authorization.k8s.io/v1
      kind: ClusterRole
      metadata:
        name: traefik-ingress-controller
      rules:
        - apiGroups:
            - ""
          resources:
            - namespaces
          verbs:
            - get
        - apiGroups:
            - ""
          resources:
            - configmaps
            - pods
            - secrets
            - endpoints
          verbs:
            - get
            - list
            - watch
        - apiGroups:
            - ""
          resources:
            - services
          verbs:
            - get
            - list
            - watch
        - apiGroups:
            - networking.k8s.io
          resources:
            - ingresses
          verbs:
            - get
            - list
            - watch
        - apiGroups:
            - networking.k8s.io
          resources:
            - ingresses/status
          verbs:
            - update
        - apiGroups:
            - networking.k8s.io
          resources:
            - ingressclasses
          verbs:
            - get
            - list
            - watch
        - apiGroups:
            - ""
          resources:
            - events
          verbs:
            - create
            - patch
        - apiGroups:
            - discovery.k8s.io
          resources:
            - endpointslices
          verbs:
            - list
            - watch
            - get
        
          ---
          apiVersion: rbac.authorization.k8s.io/v1
          kind: ClusterRoleBinding
          metadata:
            name: traefik-ingress-controller
          roleRef:
            apiGroup: rbac.authorization.k8s.io
            kind: ClusterRole
            name: traefik-ingress-controller
          subjects:
            - kind: ServiceAccount
              name: traefik-ingress-controller
              namespace: default
      ```

      ```yaml tab="Traefik"
      ---
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: traefik-ingress-controller

      ---
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: traefik
        labels:
          app: traefik

      spec:
        replicas: 1
        selector:
          matchLabels:
            app: traefik
        template:
          metadata:
            labels:
              app: traefik
          spec:
            serviceAccountName: traefik-ingress-controller
            containers:
              - name: traefik
                image: traefik:v3.5
                args:
                  - --entryPoints.web.address=:80
                  - --providers.kubernetesingressnginx
                ports:
                  - name: web
                    containerPort: 80

      ---
      apiVersion: v1
      kind: Service
      metadata:
        name: traefik
      spec:
        type: LoadBalancer
        selector:
          app: traefik
        ports:
          - name: web
            port: 80
            targetPort: 80
      ```

      ```yaml tab="Whoami"
      ---
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: whoami
        labels:
          app: whoami

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

      spec:
        selector:
          app: whoami
        ports:
          - name: http
            port: 80
      ```

      ```yaml tab="Ingress"
      ---
      apiVersion: networking.k8s.io/v1
      kind: IngressClass
      metadata:
        name: nginx
      spec:
        controller: k8s.io/ingress-nginx

      ---
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: myingress
        
      spec:
        ingressClassName: nginx
        rules:
          - host: whoami.localhost
            http:
              paths:
                - path: /bar
                  pathType: Exact
                  backend:
                    service:
                      name:  whoami
                      port:
                        number: 80
                - path: /foo
                  pathType: Exact
                  backend:
                    service:
                      name:  whoami
                      port:
                        number: 80
      ```

## Annotations Support

This section lists all known NGINX Ingress annotations, split between those currently implemented (with limitations if any) and those not implemented. 
Limitations or behavioral differences are indicated where relevant.

!!! warning "Global configuration"

    Traefik does not expose all global configuration options to control default behaviors for ingresses. 
    
    Some behaviors that are globally configurable in NGINX (such as default SSL redirect, rate limiting, or affinity) are currently not supported and cannot be overridden per-ingress as in NGINX.

### Caveats and Key Behavioral Differences

- **Authentication**: Forward auth behaves differently and session caching is not supported. NGINX supports sub-request based auth, while Traefik forwards the original request.
- **Session Affinity**: Only persistent mode is supported.
- **Leader Election**: Not supported; no cluster mode with leader election.
- **Default Backend**: Only `defaultBackend` in Ingress spec is supported; the annotation is ignored.
- **Load Balancing**: Only round_robin is supported; EWMA and IP hash are not supported.
- **CORS**: NGINX responds with all configured headers unconditionally; Traefik handles headers differently between pre-flight and regular requests.
- **TLS/Backend Protocols**: AUTO_HTTP, FCGI and some TLS options are not supported in Traefik.
- **Path Handling**: Traefik preserves trailing slashes by default; NGINX removes them unless configured otherwise.

### Supported NGINX Annotations

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| `nginx.ingress.kubernetes.io/affinity`                |                                                                                            |
| `nginx.ingress.kubernetes.io/affinity-mode`           | Only persistent mode supported; balanced/canary not supported.                             |
| `nginx.ingress.kubernetes.io/auth-type`               |                                                                                            |
| `nginx.ingress.kubernetes.io/auth-secret`             |                                                                                            |
| `nginx.ingress.kubernetes.io/auth-secret-type`        |                                                                                            |
| `nginx.ingress.kubernetes.io/auth-realm`              |                                                                                            |
| `nginx.ingress.kubernetes.io/auth-url`                | Only URL and response headers copy supported. Forward auth behaves differently than NGINX. |
| `nginx.ingress.kubernetes.io/auth-method`             |                                                                                            |
| `nginx.ingress.kubernetes.io/auth-response-headers`   |                                                                                            |
| `nginx.ingress.kubernetes.io/ssl-redirect`            | Cannot opt-out per route if enabled globally.                                              |
| `nginx.ingress.kubernetes.io/force-ssl-redirect`      | Cannot opt-out per route if enabled globally.                                              |
| `nginx.ingress.kubernetes.io/ssl-passthrough`         | Some differences in SNI/default backend handling.                                          |
| `nginx.ingress.kubernetes.io/use-regex`               |                                                                                            |
| `nginx.ingress.kubernetes.io/session-cookie-name`     |                                                                                            |
| `nginx.ingress.kubernetes.io/session-cookie-path`     |                                                                                            |
| `nginx.ingress.kubernetes.io/session-cookie-domain`   |                                                                                            |
| `nginx.ingress.kubernetes.io/session-cookie-samesite` |                                                                                            |
| `nginx.ingress.kubernetes.io/load-balance`            | Only round_robin supported; ewma and IP hash not supported.                                |
| `nginx.ingress.kubernetes.io/backend-protocol`        | FCGI and AUTO_HTTP not supported.                                                          |
| `nginx.ingress.kubernetes.io/enable-cors`             | Partial support.                                                                           |
| `nginx.ingress.kubernetes.io/cors-allow-credentials`  |                                                                                            |
| `nginx.ingress.kubernetes.io/cors-allow-headers`      |                                                                                            |
| `nginx.ingress.kubernetes.io/cors-allow-methods`      |                                                                                            |
| `nginx.ingress.kubernetes.io/cors-allow-origin`       |                                                                                            |
| `nginx.ingress.kubernetes.io/cors-max-age`            |                                                                                            |
| `nginx.ingress.kubernetes.io/proxy-ssl-server-name`   |                                                                                            |
| `nginx.ingress.kubernetes.io/proxy-ssl-name`          |                                                                                            |
| `nginx.ingress.kubernetes.io/proxy-ssl-verify`        |                                                                                            |
| `nginx.ingress.kubernetes.io/proxy-ssl-secret`        |                                                                                            |
| `nginx.ingress.kubernetes.io/service-upstream`        |                                                                                            |

### Unsupported NGINX Annotations

!!! question "Want to Add Support for More Annotations?"

    You can help extend support in two ways:

    - [**Open a PR**](../../../contributing/submitting-pull-requests.md) with the new annotation support.
    - **Reach out** to the [Traefik Labs support team](https://info.traefik.io/request-commercial-support?cta=doc).

    All contributions and suggestions are welcome — let's build this together!


| Annotation                                                                  | Notes                                                |
|-----------------------------------------------------------------------------|------------------------------------------------------|
| `nginx.ingress.kubernetes.io/app-root`                                      | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/affinity-canary-behavior`                      | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-tls-secret`                               | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-tls-verify-depth`                         | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-tls-verify-client`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-tls-error-page`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream`         | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-tls-match-cn`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-cache-key`                                | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-cache-duration`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-keepalive`                                | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-keepalive-share-vars`                     | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-keepalive-requests`                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-keepalive-timeout`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-proxy-set-headers`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/auth-snippet`                                  | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-global-auth`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary`                                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary-by-header`                              | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary-by-header-value`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary-by-header-pattern`                      | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary-by-cookie`                              | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary-weight`                                 | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/canary-weight-total`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/client-body-buffer-size`                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/configuration-snippet`                         | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/custom-http-errors`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/disable-proxy-intercept-errors`                | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/default-backend`                               | Not supported yet; use `defaultBackend` in Ingress spec. |
| `nginx.ingress.kubernetes.io/limit-rate-after`                              | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/limit-rate`                                    | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/limit-whitelist`                               | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/limit-rps`                                     | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/limit-rpm`                                     | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/limit-burst-multiplier`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/limit-connections`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/global-rate-limit`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/global-rate-limit-window`                      | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/global-rate-limit-key`                         | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/global-rate-limit-ignored-cidrs`               | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/permanent-redirect`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/permanent-redirect-code`                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/temporal-redirect`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/preserve-trailing-slash`                       | Not supported yet; Traefik preserves by default.         |
| `nginx.ingress.kubernetes.io/proxy-cookie-domain`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-cookie-path`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-connect-timeout`                         | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-send-timeout`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-read-timeout`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-next-upstream`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-next-upstream-timeout`                   | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-next-upstream-tries`                     | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-request-buffering`                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-redirect-from`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-redirect-to`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-http-version`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-ssl-ciphers`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-ssl-verify-depth`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-ssl-protocols`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-rewrite-log`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/rewrite-target`                                | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/satisfy`                                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/server-alias`                                  | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/server-snippet`                                | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none`      | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/session-cookie-expires`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/session-cookie-change-on-failure`              | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/ssl-ciphers`                                   | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers`                     | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/connection-proxy-header`                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-access-log`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-opentracing`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/opentracing-trust-incoming-span`               | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-opentelemetry`                          | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/opentelemetry-trust-incoming-span`             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-modsecurity`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/enable-owasp-core-rules`                       | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/modsecurity-transaction-id`                    | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/modsecurity-snippet`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/mirror-request-body`                           | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/mirror-target`                                 | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/mirror-host`                                   | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/x-forwarded-prefix`                            | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/upstream-hash-by`                              | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/upstream-vhost`                                | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/denylist-source-range`                         | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/whitelist-source-range`                        | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-buffering`                               | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-buffers-number`                          | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-buffer-size`                             | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/proxy-max-temp-file-size`                      | Not supported yet.                                    |
| `nginx.ingress.kubernetes.io/stream-snippet`                                | Not supported yet.                                    |
