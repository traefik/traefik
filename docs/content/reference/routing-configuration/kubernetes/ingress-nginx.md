---
title: "Traefik Kubernetes Ingress NGINX Routing Configuration"
description: "Understand the routing configuration for the Kubernetes Ingress NGINX Controller and Traefik Proxy. Read the technical documentation."
---

# Traefik & Ingresses with NGINX Annotations

Enable seamless migration from NGINX Ingress Controller to Traefik with NGINX annotation compatibility.
{: .subtitle }

!!! warning "NGINX Ingress Controller Retirement"

    The Kubernetes NGINX Ingress Controller project has announced its retirement in **March 2026** and will no longer receive updates or security patches.
    Traefik provides a migration path by supporting NGINX annotations, allowing you to transition your workloads without rewriting all your Ingress configurations.

    **→ See the [NGINX to Traefik Migration Guide](../../../migrate/nginx-to-traefik.md) for step-by-step instructions.**

    For more information about the NGINX Ingress Controller retirement, see the [official Kubernetes blog announcement](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement).

## Ingress Discovery

This provider discovers all Ingresses in the cluster by default, which may lead to duplicated routers if you are also using the standard Kubernetes Ingress provider.

**Best Practices:**

- Use IngressClass to specify which Ingresses should be handled by this provider
- Configure `watchNamespace` to limit discovery to specific namespaces
- Use `watchNamespaceSelector` to target Ingresses based on namespace labels

## Routing Configuration

This provider watches for incoming Ingress events and automatically translates NGINX annotations into Traefik's dynamic configuration, creating the corresponding routers, services, middlewares, and other components needed to handle your traffic.

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
                image: traefik:v3.6
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

This section lists all known NGINX Ingress annotations.
The following annotations are organized by category for easier navigation.

### Authentication

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-ioauth-type" href="#opt-nginx-ingress-kubernetes-ioauth-type" title="#opt-nginx-ingress-kubernetes-ioauth-type">`nginx.ingress.kubernetes.io/auth-type`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioauth-secret" href="#opt-nginx-ingress-kubernetes-ioauth-secret" title="#opt-nginx-ingress-kubernetes-ioauth-secret">`nginx.ingress.kubernetes.io/auth-secret`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioauth-secret-type" href="#opt-nginx-ingress-kubernetes-ioauth-secret-type" title="#opt-nginx-ingress-kubernetes-ioauth-secret-type">`nginx.ingress.kubernetes.io/auth-secret-type`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioauth-realm" href="#opt-nginx-ingress-kubernetes-ioauth-realm" title="#opt-nginx-ingress-kubernetes-ioauth-realm">`nginx.ingress.kubernetes.io/auth-realm`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioauth-url" href="#opt-nginx-ingress-kubernetes-ioauth-url" title="#opt-nginx-ingress-kubernetes-ioauth-url">`nginx.ingress.kubernetes.io/auth-url`</a> | Only URL and response headers copy supported. Forward auth behaves differently than NGINX. |
| <a id="opt-nginx-ingress-kubernetes-ioauth-method" href="#opt-nginx-ingress-kubernetes-ioauth-method" title="#opt-nginx-ingress-kubernetes-ioauth-method">`nginx.ingress.kubernetes.io/auth-method`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioauth-response-headers" href="#opt-nginx-ingress-kubernetes-ioauth-response-headers" title="#opt-nginx-ingress-kubernetes-ioauth-response-headers">`nginx.ingress.kubernetes.io/auth-response-headers`</a> |                                                                                            |

### SSL/TLS

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-iossl-redirect" href="#opt-nginx-ingress-kubernetes-iossl-redirect" title="#opt-nginx-ingress-kubernetes-iossl-redirect">`nginx.ingress.kubernetes.io/ssl-redirect`</a> | Cannot opt-out per route if enabled globally.                                              |
| <a id="opt-nginx-ingress-kubernetes-ioforce-ssl-redirect" href="#opt-nginx-ingress-kubernetes-ioforce-ssl-redirect" title="#opt-nginx-ingress-kubernetes-ioforce-ssl-redirect">`nginx.ingress.kubernetes.io/force-ssl-redirect`</a> | Cannot opt-out per route if enabled globally.                                              |
| <a id="opt-nginx-ingress-kubernetes-iossl-passthrough" href="#opt-nginx-ingress-kubernetes-iossl-passthrough" title="#opt-nginx-ingress-kubernetes-iossl-passthrough">`nginx.ingress.kubernetes.io/ssl-passthrough`</a> | Some differences in SNI/default backend handling.                                          |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-server-name" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-server-name" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-server-name">`nginx.ingress.kubernetes.io/proxy-ssl-server-name`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-name" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-name" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-name">`nginx.ingress.kubernetes.io/proxy-ssl-name`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-verify" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-verify" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-verify">`nginx.ingress.kubernetes.io/proxy-ssl-verify`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-secret" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-secret" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-secret">`nginx.ingress.kubernetes.io/proxy-ssl-secret`</a> |                                                                                            |

### Session Affinity

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-ioaffinity" href="#opt-nginx-ingress-kubernetes-ioaffinity" title="#opt-nginx-ingress-kubernetes-ioaffinity">`nginx.ingress.kubernetes.io/affinity`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-ioaffinity-mode" href="#opt-nginx-ingress-kubernetes-ioaffinity-mode" title="#opt-nginx-ingress-kubernetes-ioaffinity-mode">`nginx.ingress.kubernetes.io/affinity-mode`</a> | Only persistent mode supported; balanced/canary not supported.                             |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-name" href="#opt-nginx-ingress-kubernetes-iosession-cookie-name" title="#opt-nginx-ingress-kubernetes-iosession-cookie-name">`nginx.ingress.kubernetes.io/session-cookie-name`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-secure" href="#opt-nginx-ingress-kubernetes-iosession-cookie-secure" title="#opt-nginx-ingress-kubernetes-iosession-cookie-secure">`nginx.ingress.kubernetes.io/session-cookie-secure`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-path" href="#opt-nginx-ingress-kubernetes-iosession-cookie-path" title="#opt-nginx-ingress-kubernetes-iosession-cookie-path">`nginx.ingress.kubernetes.io/session-cookie-path`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-domain" href="#opt-nginx-ingress-kubernetes-iosession-cookie-domain" title="#opt-nginx-ingress-kubernetes-iosession-cookie-domain">`nginx.ingress.kubernetes.io/session-cookie-domain`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-samesite" href="#opt-nginx-ingress-kubernetes-iosession-cookie-samesite" title="#opt-nginx-ingress-kubernetes-iosession-cookie-samesite">`nginx.ingress.kubernetes.io/session-cookie-samesite`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-max-age" href="#opt-nginx-ingress-kubernetes-iosession-cookie-max-age" title="#opt-nginx-ingress-kubernetes-iosession-cookie-max-age">`nginx.ingress.kubernetes.io/session-cookie-max-age`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-expires" href="#opt-nginx-ingress-kubernetes-iosession-cookie-expires" title="#opt-nginx-ingress-kubernetes-iosession-cookie-expires">`nginx.ingress.kubernetes.io/session-cookie-expires`</a> |                                                      |

### Load Balancing & Backend

| Annotation                                            | Limitations / Notes                                                                              |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-ioload-balance" href="#opt-nginx-ingress-kubernetes-ioload-balance" title="#opt-nginx-ingress-kubernetes-ioload-balance">`nginx.ingress.kubernetes.io/load-balance`</a> | Only round_robin supported; ewma and IP hash not supported.                                      |
| <a id="opt-nginx-ingress-kubernetes-iobackend-protocol" href="#opt-nginx-ingress-kubernetes-iobackend-protocol" title="#opt-nginx-ingress-kubernetes-iobackend-protocol">`nginx.ingress.kubernetes.io/backend-protocol`</a> | FCGI and AUTO_HTTP not supported.                                                                |
| <a id="opt-nginx-ingress-kubernetes-ioservice-upstream" href="#opt-nginx-ingress-kubernetes-ioservice-upstream" title="#opt-nginx-ingress-kubernetes-ioservice-upstream">`nginx.ingress.kubernetes.io/service-upstream`</a> |                                                                                                  |
| <a id="opt-nginx-ingress-kubernetes-ioupstream-vhost" href="#opt-nginx-ingress-kubernetes-ioupstream-vhost" title="#opt-nginx-ingress-kubernetes-ioupstream-vhost">`nginx.ingress.kubernetes.io/upstream-vhost`</a> |                                                                                                  |
| <a id="opt-nginx-ingress-kubernetes-iocustom-headers" href="#opt-nginx-ingress-kubernetes-iocustom-headers" title="#opt-nginx-ingress-kubernetes-iocustom-headers">`nginx.ingress.kubernetes.io/custom-headers`</a> | Header whitelisting, similar to `global-allowed-response-headers` NGINX config is not supported. |

### CORS

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-ioenable-cors" href="#opt-nginx-ingress-kubernetes-ioenable-cors" title="#opt-nginx-ingress-kubernetes-ioenable-cors">`nginx.ingress.kubernetes.io/enable-cors`</a> | Partial support.                                                                           |
| <a id="opt-nginx-ingress-kubernetes-iocors-allow-credentials" href="#opt-nginx-ingress-kubernetes-iocors-allow-credentials" title="#opt-nginx-ingress-kubernetes-iocors-allow-credentials">`nginx.ingress.kubernetes.io/cors-allow-credentials`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iocors-allow-headers" href="#opt-nginx-ingress-kubernetes-iocors-allow-headers" title="#opt-nginx-ingress-kubernetes-iocors-allow-headers">`nginx.ingress.kubernetes.io/cors-allow-headers`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iocors-allow-methods" href="#opt-nginx-ingress-kubernetes-iocors-allow-methods" title="#opt-nginx-ingress-kubernetes-iocors-allow-methods">`nginx.ingress.kubernetes.io/cors-allow-methods`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iocors-allow-origin" href="#opt-nginx-ingress-kubernetes-iocors-allow-origin" title="#opt-nginx-ingress-kubernetes-iocors-allow-origin">`nginx.ingress.kubernetes.io/cors-allow-origin`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iocors-expose-headers" href="#opt-nginx-ingress-kubernetes-iocors-expose-headers" title="#opt-nginx-ingress-kubernetes-iocors-expose-headers">`nginx.ingress.kubernetes.io/cors-expose-headers`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iocors-max-age" href="#opt-nginx-ingress-kubernetes-iocors-max-age" title="#opt-nginx-ingress-kubernetes-iocors-max-age">`nginx.ingress.kubernetes.io/cors-max-age`</a> |                                                                                            |

### Routing

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-ioapp-root" href="#opt-nginx-ingress-kubernetes-ioapp-root" title="#opt-nginx-ingress-kubernetes-ioapp-root">`nginx.ingress.kubernetes.io/app-root`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iouse-regex" href="#opt-nginx-ingress-kubernetes-iouse-regex" title="#opt-nginx-ingress-kubernetes-iouse-regex">`nginx.ingress.kubernetes.io/use-regex`</a> |                                                                                            |
| <a id="opt-nginx-ingress-kubernetes-iorewrite-target" href="#opt-nginx-ingress-kubernetes-iorewrite-target" title="#opt-nginx-ingress-kubernetes-iorewrite-target">`nginx.ingress.kubernetes.io/rewrite-target`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iopermanent-redirect" href="#opt-nginx-ingress-kubernetes-iopermanent-redirect" title="#opt-nginx-ingress-kubernetes-iopermanent-redirect">`nginx.ingress.kubernetes.io/permanent-redirect`</a> | Defaults to a 301 Moved Permanently status code.                                                     |
| <a id="opt-nginx-ingress-kubernetes-iopermanent-redirect-code" href="#opt-nginx-ingress-kubernetes-iopermanent-redirect-code" title="#opt-nginx-ingress-kubernetes-iopermanent-redirect-code">`nginx.ingress.kubernetes.io/permanent-redirect-code`</a> | Only valid 3XX HTTP Status Codes are accepted.                                                     |
| <a id="opt-nginx-ingress-kubernetes-iotemporal-redirect" href="#opt-nginx-ingress-kubernetes-iotemporal-redirect" title="#opt-nginx-ingress-kubernetes-iotemporal-redirect">`nginx.ingress.kubernetes.io/temporal-redirect`</a> | Takes precedence over the `permanent-redirect` annotation. Defaults to a 302 Found status code.                                                |
| <a id="opt-nginx-ingress-kubernetes-iotemporal-redirect-code" href="#opt-nginx-ingress-kubernetes-iotemporal-redirect-code" title="#opt-nginx-ingress-kubernetes-iotemporal-redirect-code">`nginx.ingress.kubernetes.io/temporal-redirect-code`</a> | Only valid 3XX HTTP Status Codes are accepted.                                                      |

### IP Whitelist

| Annotation                                            | Limitations / Notes                                                                        |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-iowhitelist-source-range" href="#opt-nginx-ingress-kubernetes-iowhitelist-source-range" title="#opt-nginx-ingress-kubernetes-iowhitelist-source-range">`nginx.ingress.kubernetes.io/whitelist-source-range`</a> |                                                      |


## Limitations

### Caveats and Key Behavioral Differences

- **Authentication**: Forward auth behaves differently and session caching is not supported. NGINX supports sub-request based auth, while Traefik forwards the original request.
- **Session Affinity**: Only persistent mode is supported.
- **Leader Election**: Not supported; no cluster mode with leader election.
- **Default Backend**: Only defaultBackend in Ingress spec is supported; the annotation is ignored.
- **Load Balancing**: Only round_robin is supported; EWMA and IP hash are not supported.
- **CORS**: NGINX responds with all configured headers unconditionally; Traefik handles headers differently between pre-flight and regular requests.
- **TLS/Backend Protocols**: AUTO_HTTP, FCGI and some TLS options are not supported in Traefik.
- **Path Handling**: Traefik preserves trailing slashes by default; NGINX removes them unless configured otherwise

### Unsupported Annotations

!!! question "Want to Add Support for More Annotations?"

    You can help extend support in two ways:

    - [**Open a PR**](../../../contributing/submitting-pull-requests.md) with the new annotation support.
    - **Reach out** to the [Traefik Labs support team](https://info.traefik.io/request-commercial-support?cta=doc).

    All contributions and suggestions are welcome — let's build this together!

| Annotation                                                                  | Notes                                                |
|-----------------------------------------------------------------------------|------------------------------------------------------|
| <a id="opt-nginx-ingress-kubernetes-ioaffinity-canary-behavior" href="#opt-nginx-ingress-kubernetes-ioaffinity-canary-behavior" title="#opt-nginx-ingress-kubernetes-ioaffinity-canary-behavior">`nginx.ingress.kubernetes.io/affinity-canary-behavior`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-signin" href="#opt-nginx-ingress-kubernetes-ioauth-signin" title="#opt-nginx-ingress-kubernetes-ioauth-signin">`nginx.ingress.kubernetes.io/auth-signin`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-tls-secret" href="#opt-nginx-ingress-kubernetes-ioauth-tls-secret" title="#opt-nginx-ingress-kubernetes-ioauth-tls-secret">`nginx.ingress.kubernetes.io/auth-tls-secret`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-tls-verify-depth" href="#opt-nginx-ingress-kubernetes-ioauth-tls-verify-depth" title="#opt-nginx-ingress-kubernetes-ioauth-tls-verify-depth">`nginx.ingress.kubernetes.io/auth-tls-verify-depth`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-tls-verify-client" href="#opt-nginx-ingress-kubernetes-ioauth-tls-verify-client" title="#opt-nginx-ingress-kubernetes-ioauth-tls-verify-client">`nginx.ingress.kubernetes.io/auth-tls-verify-client`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-tls-error-page" href="#opt-nginx-ingress-kubernetes-ioauth-tls-error-page" title="#opt-nginx-ingress-kubernetes-ioauth-tls-error-page">`nginx.ingress.kubernetes.io/auth-tls-error-page`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-tls-pass-certificate-to-upstream" href="#opt-nginx-ingress-kubernetes-ioauth-tls-pass-certificate-to-upstream" title="#opt-nginx-ingress-kubernetes-ioauth-tls-pass-certificate-to-upstream">`nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-tls-match-cn" href="#opt-nginx-ingress-kubernetes-ioauth-tls-match-cn" title="#opt-nginx-ingress-kubernetes-ioauth-tls-match-cn">`nginx.ingress.kubernetes.io/auth-tls-match-cn`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-cache-key" href="#opt-nginx-ingress-kubernetes-ioauth-cache-key" title="#opt-nginx-ingress-kubernetes-ioauth-cache-key">`nginx.ingress.kubernetes.io/auth-cache-key`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-cache-duration" href="#opt-nginx-ingress-kubernetes-ioauth-cache-duration" title="#opt-nginx-ingress-kubernetes-ioauth-cache-duration">`nginx.ingress.kubernetes.io/auth-cache-duration`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-keepalive" href="#opt-nginx-ingress-kubernetes-ioauth-keepalive" title="#opt-nginx-ingress-kubernetes-ioauth-keepalive">`nginx.ingress.kubernetes.io/auth-keepalive`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-keepalive-share-vars" href="#opt-nginx-ingress-kubernetes-ioauth-keepalive-share-vars" title="#opt-nginx-ingress-kubernetes-ioauth-keepalive-share-vars">`nginx.ingress.kubernetes.io/auth-keepalive-share-vars`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-keepalive-requests" href="#opt-nginx-ingress-kubernetes-ioauth-keepalive-requests" title="#opt-nginx-ingress-kubernetes-ioauth-keepalive-requests">`nginx.ingress.kubernetes.io/auth-keepalive-requests`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-keepalive-timeout" href="#opt-nginx-ingress-kubernetes-ioauth-keepalive-timeout" title="#opt-nginx-ingress-kubernetes-ioauth-keepalive-timeout">`nginx.ingress.kubernetes.io/auth-keepalive-timeout`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-proxy-set-headers" href="#opt-nginx-ingress-kubernetes-ioauth-proxy-set-headers" title="#opt-nginx-ingress-kubernetes-ioauth-proxy-set-headers">`nginx.ingress.kubernetes.io/auth-proxy-set-headers`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioauth-snippet" href="#opt-nginx-ingress-kubernetes-ioauth-snippet" title="#opt-nginx-ingress-kubernetes-ioauth-snippet">`nginx.ingress.kubernetes.io/auth-snippet`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-global-auth" href="#opt-nginx-ingress-kubernetes-ioenable-global-auth" title="#opt-nginx-ingress-kubernetes-ioenable-global-auth">`nginx.ingress.kubernetes.io/enable-global-auth`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary" href="#opt-nginx-ingress-kubernetes-iocanary" title="#opt-nginx-ingress-kubernetes-iocanary">`nginx.ingress.kubernetes.io/canary`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary-by-header" href="#opt-nginx-ingress-kubernetes-iocanary-by-header" title="#opt-nginx-ingress-kubernetes-iocanary-by-header">`nginx.ingress.kubernetes.io/canary-by-header`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary-by-header-value" href="#opt-nginx-ingress-kubernetes-iocanary-by-header-value" title="#opt-nginx-ingress-kubernetes-iocanary-by-header-value">`nginx.ingress.kubernetes.io/canary-by-header-value`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary-by-header-pattern" href="#opt-nginx-ingress-kubernetes-iocanary-by-header-pattern" title="#opt-nginx-ingress-kubernetes-iocanary-by-header-pattern">`nginx.ingress.kubernetes.io/canary-by-header-pattern`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary-by-cookie" href="#opt-nginx-ingress-kubernetes-iocanary-by-cookie" title="#opt-nginx-ingress-kubernetes-iocanary-by-cookie">`nginx.ingress.kubernetes.io/canary-by-cookie`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary-weight" href="#opt-nginx-ingress-kubernetes-iocanary-weight" title="#opt-nginx-ingress-kubernetes-iocanary-weight">`nginx.ingress.kubernetes.io/canary-weight`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocanary-weight-total" href="#opt-nginx-ingress-kubernetes-iocanary-weight-total" title="#opt-nginx-ingress-kubernetes-iocanary-weight-total">`nginx.ingress.kubernetes.io/canary-weight-total`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioclient-body-buffer-size" href="#opt-nginx-ingress-kubernetes-ioclient-body-buffer-size" title="#opt-nginx-ingress-kubernetes-ioclient-body-buffer-size">`nginx.ingress.kubernetes.io/client-body-buffer-size`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioconfiguration-snippet" href="#opt-nginx-ingress-kubernetes-ioconfiguration-snippet" title="#opt-nginx-ingress-kubernetes-ioconfiguration-snippet">`nginx.ingress.kubernetes.io/configuration-snippet`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iocustom-http-errors" href="#opt-nginx-ingress-kubernetes-iocustom-http-errors" title="#opt-nginx-ingress-kubernetes-iocustom-http-errors">`nginx.ingress.kubernetes.io/custom-http-errors`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iodisable-proxy-intercept-errors" href="#opt-nginx-ingress-kubernetes-iodisable-proxy-intercept-errors" title="#opt-nginx-ingress-kubernetes-iodisable-proxy-intercept-errors">`nginx.ingress.kubernetes.io/disable-proxy-intercept-errors`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iodefault-backend" href="#opt-nginx-ingress-kubernetes-iodefault-backend" title="#opt-nginx-ingress-kubernetes-iodefault-backend">`nginx.ingress.kubernetes.io/default-backend`</a> | Use `defaultBackend` in Ingress spec. |
| <a id="opt-nginx-ingress-kubernetes-iolimit-rate-after" href="#opt-nginx-ingress-kubernetes-iolimit-rate-after" title="#opt-nginx-ingress-kubernetes-iolimit-rate-after">`nginx.ingress.kubernetes.io/limit-rate-after`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iolimit-rate" href="#opt-nginx-ingress-kubernetes-iolimit-rate" title="#opt-nginx-ingress-kubernetes-iolimit-rate">`nginx.ingress.kubernetes.io/limit-rate`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iolimit-whitelist" href="#opt-nginx-ingress-kubernetes-iolimit-whitelist" title="#opt-nginx-ingress-kubernetes-iolimit-whitelist">`nginx.ingress.kubernetes.io/limit-whitelist`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iolimit-rps" href="#opt-nginx-ingress-kubernetes-iolimit-rps" title="#opt-nginx-ingress-kubernetes-iolimit-rps">`nginx.ingress.kubernetes.io/limit-rps`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iolimit-rpm" href="#opt-nginx-ingress-kubernetes-iolimit-rpm" title="#opt-nginx-ingress-kubernetes-iolimit-rpm">`nginx.ingress.kubernetes.io/limit-rpm`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iolimit-burst-multiplier" href="#opt-nginx-ingress-kubernetes-iolimit-burst-multiplier" title="#opt-nginx-ingress-kubernetes-iolimit-burst-multiplier">`nginx.ingress.kubernetes.io/limit-burst-multiplier`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iolimit-connections" href="#opt-nginx-ingress-kubernetes-iolimit-connections" title="#opt-nginx-ingress-kubernetes-iolimit-connections">`nginx.ingress.kubernetes.io/limit-connections`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioglobal-rate-limit" href="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit" title="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit">`nginx.ingress.kubernetes.io/global-rate-limit`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioglobal-rate-limit-window" href="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit-window" title="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit-window">`nginx.ingress.kubernetes.io/global-rate-limit-window`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioglobal-rate-limit-key" href="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit-key" title="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit-key">`nginx.ingress.kubernetes.io/global-rate-limit-key`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioglobal-rate-limit-ignored-cidrs" href="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit-ignored-cidrs" title="#opt-nginx-ingress-kubernetes-ioglobal-rate-limit-ignored-cidrs">`nginx.ingress.kubernetes.io/global-rate-limit-ignored-cidrs`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iopreserve-trailing-slash" href="#opt-nginx-ingress-kubernetes-iopreserve-trailing-slash" title="#opt-nginx-ingress-kubernetes-iopreserve-trailing-slash">`nginx.ingress.kubernetes.io/preserve-trailing-slash`</a> | Traefik preserves trailing slash by default.         |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-cookie-domain" href="#opt-nginx-ingress-kubernetes-ioproxy-cookie-domain" title="#opt-nginx-ingress-kubernetes-ioproxy-cookie-domain">`nginx.ingress.kubernetes.io/proxy-cookie-domain`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-cookie-path" href="#opt-nginx-ingress-kubernetes-ioproxy-cookie-path" title="#opt-nginx-ingress-kubernetes-ioproxy-cookie-path">`nginx.ingress.kubernetes.io/proxy-cookie-path`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-connect-timeout" href="#opt-nginx-ingress-kubernetes-ioproxy-connect-timeout" title="#opt-nginx-ingress-kubernetes-ioproxy-connect-timeout">`nginx.ingress.kubernetes.io/proxy-connect-timeout`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-send-timeout" href="#opt-nginx-ingress-kubernetes-ioproxy-send-timeout" title="#opt-nginx-ingress-kubernetes-ioproxy-send-timeout">`nginx.ingress.kubernetes.io/proxy-send-timeout`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-read-timeout" href="#opt-nginx-ingress-kubernetes-ioproxy-read-timeout" title="#opt-nginx-ingress-kubernetes-ioproxy-read-timeout">`nginx.ingress.kubernetes.io/proxy-read-timeout`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-next-upstream" href="#opt-nginx-ingress-kubernetes-ioproxy-next-upstream" title="#opt-nginx-ingress-kubernetes-ioproxy-next-upstream">`nginx.ingress.kubernetes.io/proxy-next-upstream`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-next-upstream-timeout" href="#opt-nginx-ingress-kubernetes-ioproxy-next-upstream-timeout" title="#opt-nginx-ingress-kubernetes-ioproxy-next-upstream-timeout">`nginx.ingress.kubernetes.io/proxy-next-upstream-timeout`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-next-upstream-tries" href="#opt-nginx-ingress-kubernetes-ioproxy-next-upstream-tries" title="#opt-nginx-ingress-kubernetes-ioproxy-next-upstream-tries">`nginx.ingress.kubernetes.io/proxy-next-upstream-tries`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-request-buffering" href="#opt-nginx-ingress-kubernetes-ioproxy-request-buffering" title="#opt-nginx-ingress-kubernetes-ioproxy-request-buffering">`nginx.ingress.kubernetes.io/proxy-request-buffering`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-redirect-from" href="#opt-nginx-ingress-kubernetes-ioproxy-redirect-from" title="#opt-nginx-ingress-kubernetes-ioproxy-redirect-from">`nginx.ingress.kubernetes.io/proxy-redirect-from`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-redirect-to" href="#opt-nginx-ingress-kubernetes-ioproxy-redirect-to" title="#opt-nginx-ingress-kubernetes-ioproxy-redirect-to">`nginx.ingress.kubernetes.io/proxy-redirect-to`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-http-version" href="#opt-nginx-ingress-kubernetes-ioproxy-http-version" title="#opt-nginx-ingress-kubernetes-ioproxy-http-version">`nginx.ingress.kubernetes.io/proxy-http-version`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-ciphers" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-ciphers" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-ciphers">`nginx.ingress.kubernetes.io/proxy-ssl-ciphers`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-verify-depth" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-verify-depth" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-verify-depth">`nginx.ingress.kubernetes.io/proxy-ssl-verify-depth`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-ssl-protocols" href="#opt-nginx-ingress-kubernetes-ioproxy-ssl-protocols" title="#opt-nginx-ingress-kubernetes-ioproxy-ssl-protocols">`nginx.ingress.kubernetes.io/proxy-ssl-protocols`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-rewrite-log" href="#opt-nginx-ingress-kubernetes-ioenable-rewrite-log" title="#opt-nginx-ingress-kubernetes-ioenable-rewrite-log">`nginx.ingress.kubernetes.io/enable-rewrite-log`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iosatisfy" href="#opt-nginx-ingress-kubernetes-iosatisfy" title="#opt-nginx-ingress-kubernetes-iosatisfy">`nginx.ingress.kubernetes.io/satisfy`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioserver-alias" href="#opt-nginx-ingress-kubernetes-ioserver-alias" title="#opt-nginx-ingress-kubernetes-ioserver-alias">`nginx.ingress.kubernetes.io/server-alias`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioserver-snippet" href="#opt-nginx-ingress-kubernetes-ioserver-snippet" title="#opt-nginx-ingress-kubernetes-ioserver-snippet">`nginx.ingress.kubernetes.io/server-snippet`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-conditional-samesite-none" href="#opt-nginx-ingress-kubernetes-iosession-cookie-conditional-samesite-none" title="#opt-nginx-ingress-kubernetes-iosession-cookie-conditional-samesite-none">`nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iosession-cookie-change-on-failure" href="#opt-nginx-ingress-kubernetes-iosession-cookie-change-on-failure" title="#opt-nginx-ingress-kubernetes-iosession-cookie-change-on-failure">`nginx.ingress.kubernetes.io/session-cookie-change-on-failure`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iossl-ciphers" href="#opt-nginx-ingress-kubernetes-iossl-ciphers" title="#opt-nginx-ingress-kubernetes-iossl-ciphers">`nginx.ingress.kubernetes.io/ssl-ciphers`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iossl-prefer-server-ciphers" href="#opt-nginx-ingress-kubernetes-iossl-prefer-server-ciphers" title="#opt-nginx-ingress-kubernetes-iossl-prefer-server-ciphers">`nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioconnection-proxy-header" href="#opt-nginx-ingress-kubernetes-ioconnection-proxy-header" title="#opt-nginx-ingress-kubernetes-ioconnection-proxy-header">`nginx.ingress.kubernetes.io/connection-proxy-header`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-access-log" href="#opt-nginx-ingress-kubernetes-ioenable-access-log" title="#opt-nginx-ingress-kubernetes-ioenable-access-log">`nginx.ingress.kubernetes.io/enable-access-log`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-opentracing" href="#opt-nginx-ingress-kubernetes-ioenable-opentracing" title="#opt-nginx-ingress-kubernetes-ioenable-opentracing">`nginx.ingress.kubernetes.io/enable-opentracing`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioopentracing-trust-incoming-span" href="#opt-nginx-ingress-kubernetes-ioopentracing-trust-incoming-span" title="#opt-nginx-ingress-kubernetes-ioopentracing-trust-incoming-span">`nginx.ingress.kubernetes.io/opentracing-trust-incoming-span`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-opentelemetry" href="#opt-nginx-ingress-kubernetes-ioenable-opentelemetry" title="#opt-nginx-ingress-kubernetes-ioenable-opentelemetry">`nginx.ingress.kubernetes.io/enable-opentelemetry`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioopentelemetry-trust-incoming-span" href="#opt-nginx-ingress-kubernetes-ioopentelemetry-trust-incoming-span" title="#opt-nginx-ingress-kubernetes-ioopentelemetry-trust-incoming-span">`nginx.ingress.kubernetes.io/opentelemetry-trust-incoming-span`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-modsecurity" href="#opt-nginx-ingress-kubernetes-ioenable-modsecurity" title="#opt-nginx-ingress-kubernetes-ioenable-modsecurity">`nginx.ingress.kubernetes.io/enable-modsecurity`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioenable-owasp-core-rules" href="#opt-nginx-ingress-kubernetes-ioenable-owasp-core-rules" title="#opt-nginx-ingress-kubernetes-ioenable-owasp-core-rules">`nginx.ingress.kubernetes.io/enable-owasp-core-rules`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iomodsecurity-transaction-id" href="#opt-nginx-ingress-kubernetes-iomodsecurity-transaction-id" title="#opt-nginx-ingress-kubernetes-iomodsecurity-transaction-id">`nginx.ingress.kubernetes.io/modsecurity-transaction-id`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iomodsecurity-snippet" href="#opt-nginx-ingress-kubernetes-iomodsecurity-snippet" title="#opt-nginx-ingress-kubernetes-iomodsecurity-snippet">`nginx.ingress.kubernetes.io/modsecurity-snippet`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iomirror-request-body" href="#opt-nginx-ingress-kubernetes-iomirror-request-body" title="#opt-nginx-ingress-kubernetes-iomirror-request-body">`nginx.ingress.kubernetes.io/mirror-request-body`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iomirror-target" href="#opt-nginx-ingress-kubernetes-iomirror-target" title="#opt-nginx-ingress-kubernetes-iomirror-target">`nginx.ingress.kubernetes.io/mirror-target`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iomirror-host" href="#opt-nginx-ingress-kubernetes-iomirror-host" title="#opt-nginx-ingress-kubernetes-iomirror-host">`nginx.ingress.kubernetes.io/mirror-host`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iox-forwarded-prefix" href="#opt-nginx-ingress-kubernetes-iox-forwarded-prefix" title="#opt-nginx-ingress-kubernetes-iox-forwarded-prefix">`nginx.ingress.kubernetes.io/x-forwarded-prefix`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioupstream-hash-by" href="#opt-nginx-ingress-kubernetes-ioupstream-hash-by" title="#opt-nginx-ingress-kubernetes-ioupstream-hash-by">`nginx.ingress.kubernetes.io/upstream-hash-by`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iodenylist-source-range" href="#opt-nginx-ingress-kubernetes-iodenylist-source-range" title="#opt-nginx-ingress-kubernetes-iodenylist-source-range">`nginx.ingress.kubernetes.io/denylist-source-range`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-buffering" href="#opt-nginx-ingress-kubernetes-ioproxy-buffering" title="#opt-nginx-ingress-kubernetes-ioproxy-buffering">`nginx.ingress.kubernetes.io/proxy-buffering`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-buffers-number" href="#opt-nginx-ingress-kubernetes-ioproxy-buffers-number" title="#opt-nginx-ingress-kubernetes-ioproxy-buffers-number">`nginx.ingress.kubernetes.io/proxy-buffers-number`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-buffer-size" href="#opt-nginx-ingress-kubernetes-ioproxy-buffer-size" title="#opt-nginx-ingress-kubernetes-ioproxy-buffer-size">`nginx.ingress.kubernetes.io/proxy-buffer-size`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-ioproxy-max-temp-file-size" href="#opt-nginx-ingress-kubernetes-ioproxy-max-temp-file-size" title="#opt-nginx-ingress-kubernetes-ioproxy-max-temp-file-size">`nginx.ingress.kubernetes.io/proxy-max-temp-file-size`</a> |                                                      |
| <a id="opt-nginx-ingress-kubernetes-iostream-snippet" href="#opt-nginx-ingress-kubernetes-iostream-snippet" title="#opt-nginx-ingress-kubernetes-iostream-snippet">`nginx.ingress.kubernetes.io/stream-snippet`</a> |                                                      |

### Global Configuration

Traefik does not expose all global configuration options to control default behaviors for Ingresses in the same way NGINX does.

Some behaviors that are globally configurable in NGINX (such as default SSL redirect, rate limiting, or affinity) are currently not supported and cannot be overridden per-Ingress as in NGINX. These limitations are noted in the annotation tables below where applicable.
