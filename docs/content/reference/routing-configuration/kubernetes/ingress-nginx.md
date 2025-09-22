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
| <a id="nginx-ingress-kubernetes-ioaffinity" href="#nginx-ingress-kubernetes-ioaffinity" title="#nginx-ingress-kubernetes-ioaffinity">`nginx.ingress.kubernetes.io/affinity`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioaffinity-mode" href="#nginx-ingress-kubernetes-ioaffinity-mode" title="#nginx-ingress-kubernetes-ioaffinity-mode">`nginx.ingress.kubernetes.io/affinity-mode`</a> | Only persistent mode supported; balanced/canary not supported.                             |
| <a id="nginx-ingress-kubernetes-ioauth-type" href="#nginx-ingress-kubernetes-ioauth-type" title="#nginx-ingress-kubernetes-ioauth-type">`nginx.ingress.kubernetes.io/auth-type`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioauth-secret" href="#nginx-ingress-kubernetes-ioauth-secret" title="#nginx-ingress-kubernetes-ioauth-secret">`nginx.ingress.kubernetes.io/auth-secret`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioauth-secret-type" href="#nginx-ingress-kubernetes-ioauth-secret-type" title="#nginx-ingress-kubernetes-ioauth-secret-type">`nginx.ingress.kubernetes.io/auth-secret-type`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioauth-realm" href="#nginx-ingress-kubernetes-ioauth-realm" title="#nginx-ingress-kubernetes-ioauth-realm">`nginx.ingress.kubernetes.io/auth-realm`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioauth-url" href="#nginx-ingress-kubernetes-ioauth-url" title="#nginx-ingress-kubernetes-ioauth-url">`nginx.ingress.kubernetes.io/auth-url`</a> | Only URL and response headers copy supported. Forward auth behaves differently than NGINX. |
| <a id="nginx-ingress-kubernetes-ioauth-method" href="#nginx-ingress-kubernetes-ioauth-method" title="#nginx-ingress-kubernetes-ioauth-method">`nginx.ingress.kubernetes.io/auth-method`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioauth-response-headers" href="#nginx-ingress-kubernetes-ioauth-response-headers" title="#nginx-ingress-kubernetes-ioauth-response-headers">`nginx.ingress.kubernetes.io/auth-response-headers`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iossl-redirect" href="#nginx-ingress-kubernetes-iossl-redirect" title="#nginx-ingress-kubernetes-iossl-redirect">`nginx.ingress.kubernetes.io/ssl-redirect`</a> | Cannot opt-out per route if enabled globally.                                              |
| <a id="nginx-ingress-kubernetes-ioforce-ssl-redirect" href="#nginx-ingress-kubernetes-ioforce-ssl-redirect" title="#nginx-ingress-kubernetes-ioforce-ssl-redirect">`nginx.ingress.kubernetes.io/force-ssl-redirect`</a> | Cannot opt-out per route if enabled globally.                                              |
| <a id="nginx-ingress-kubernetes-iossl-passthrough" href="#nginx-ingress-kubernetes-iossl-passthrough" title="#nginx-ingress-kubernetes-iossl-passthrough">`nginx.ingress.kubernetes.io/ssl-passthrough`</a> | Some differences in SNI/default backend handling.                                          |
| <a id="nginx-ingress-kubernetes-iouse-regex" href="#nginx-ingress-kubernetes-iouse-regex" title="#nginx-ingress-kubernetes-iouse-regex">`nginx.ingress.kubernetes.io/use-regex`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iosession-cookie-name" href="#nginx-ingress-kubernetes-iosession-cookie-name" title="#nginx-ingress-kubernetes-iosession-cookie-name">`nginx.ingress.kubernetes.io/session-cookie-name`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iosession-cookie-path" href="#nginx-ingress-kubernetes-iosession-cookie-path" title="#nginx-ingress-kubernetes-iosession-cookie-path">`nginx.ingress.kubernetes.io/session-cookie-path`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iosession-cookie-domain" href="#nginx-ingress-kubernetes-iosession-cookie-domain" title="#nginx-ingress-kubernetes-iosession-cookie-domain">`nginx.ingress.kubernetes.io/session-cookie-domain`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iosession-cookie-samesite" href="#nginx-ingress-kubernetes-iosession-cookie-samesite" title="#nginx-ingress-kubernetes-iosession-cookie-samesite">`nginx.ingress.kubernetes.io/session-cookie-samesite`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioload-balance" href="#nginx-ingress-kubernetes-ioload-balance" title="#nginx-ingress-kubernetes-ioload-balance">`nginx.ingress.kubernetes.io/load-balance`</a> | Only round_robin supported; ewma and IP hash not supported.                                |
| <a id="nginx-ingress-kubernetes-iobackend-protocol" href="#nginx-ingress-kubernetes-iobackend-protocol" title="#nginx-ingress-kubernetes-iobackend-protocol">`nginx.ingress.kubernetes.io/backend-protocol`</a> | FCGI and AUTO_HTTP not supported.                                                          |
| <a id="nginx-ingress-kubernetes-ioenable-cors" href="#nginx-ingress-kubernetes-ioenable-cors" title="#nginx-ingress-kubernetes-ioenable-cors">`nginx.ingress.kubernetes.io/enable-cors`</a> | Partial support.                                                                           |
| <a id="nginx-ingress-kubernetes-iocors-allow-credentials" href="#nginx-ingress-kubernetes-iocors-allow-credentials" title="#nginx-ingress-kubernetes-iocors-allow-credentials">`nginx.ingress.kubernetes.io/cors-allow-credentials`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iocors-allow-headers" href="#nginx-ingress-kubernetes-iocors-allow-headers" title="#nginx-ingress-kubernetes-iocors-allow-headers">`nginx.ingress.kubernetes.io/cors-allow-headers`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iocors-allow-methods" href="#nginx-ingress-kubernetes-iocors-allow-methods" title="#nginx-ingress-kubernetes-iocors-allow-methods">`nginx.ingress.kubernetes.io/cors-allow-methods`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iocors-allow-origin" href="#nginx-ingress-kubernetes-iocors-allow-origin" title="#nginx-ingress-kubernetes-iocors-allow-origin">`nginx.ingress.kubernetes.io/cors-allow-origin`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-iocors-max-age" href="#nginx-ingress-kubernetes-iocors-max-age" title="#nginx-ingress-kubernetes-iocors-max-age">`nginx.ingress.kubernetes.io/cors-max-age`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-server-name" href="#nginx-ingress-kubernetes-ioproxy-ssl-server-name" title="#nginx-ingress-kubernetes-ioproxy-ssl-server-name">`nginx.ingress.kubernetes.io/proxy-ssl-server-name`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-name" href="#nginx-ingress-kubernetes-ioproxy-ssl-name" title="#nginx-ingress-kubernetes-ioproxy-ssl-name">`nginx.ingress.kubernetes.io/proxy-ssl-name`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-verify" href="#nginx-ingress-kubernetes-ioproxy-ssl-verify" title="#nginx-ingress-kubernetes-ioproxy-ssl-verify">`nginx.ingress.kubernetes.io/proxy-ssl-verify`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-secret" href="#nginx-ingress-kubernetes-ioproxy-ssl-secret" title="#nginx-ingress-kubernetes-ioproxy-ssl-secret">`nginx.ingress.kubernetes.io/proxy-ssl-secret`</a> |                                                                                            |
| <a id="nginx-ingress-kubernetes-ioservice-upstream" href="#nginx-ingress-kubernetes-ioservice-upstream" title="#nginx-ingress-kubernetes-ioservice-upstream">`nginx.ingress.kubernetes.io/service-upstream`</a> |                                                                                            |

### Unsupported NGINX Annotations

!!! question "Want to Add Support for More Annotations?"

    You can help extend support in two ways:

    - [**Open a PR**](../../../contributing/submitting-pull-requests.md) with the new annotation support.
    - **Reach out** to the [Traefik Labs support team](https://info.traefik.io/request-commercial-support?cta=doc).

    All contributions and suggestions are welcome — let's build this together!


| Annotation                                                                  | Notes                                                |
|-----------------------------------------------------------------------------|------------------------------------------------------|
| <a id="nginx-ingress-kubernetes-ioapp-root" href="#nginx-ingress-kubernetes-ioapp-root" title="#nginx-ingress-kubernetes-ioapp-root">`nginx.ingress.kubernetes.io/app-root`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioaffinity-canary-behavior" href="#nginx-ingress-kubernetes-ioaffinity-canary-behavior" title="#nginx-ingress-kubernetes-ioaffinity-canary-behavior">`nginx.ingress.kubernetes.io/affinity-canary-behavior`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-tls-secret" href="#nginx-ingress-kubernetes-ioauth-tls-secret" title="#nginx-ingress-kubernetes-ioauth-tls-secret">`nginx.ingress.kubernetes.io/auth-tls-secret`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-tls-verify-depth" href="#nginx-ingress-kubernetes-ioauth-tls-verify-depth" title="#nginx-ingress-kubernetes-ioauth-tls-verify-depth">`nginx.ingress.kubernetes.io/auth-tls-verify-depth`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-tls-verify-client" href="#nginx-ingress-kubernetes-ioauth-tls-verify-client" title="#nginx-ingress-kubernetes-ioauth-tls-verify-client">`nginx.ingress.kubernetes.io/auth-tls-verify-client`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-tls-error-page" href="#nginx-ingress-kubernetes-ioauth-tls-error-page" title="#nginx-ingress-kubernetes-ioauth-tls-error-page">`nginx.ingress.kubernetes.io/auth-tls-error-page`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-tls-pass-certificate-to-upstream" href="#nginx-ingress-kubernetes-ioauth-tls-pass-certificate-to-upstream" title="#nginx-ingress-kubernetes-ioauth-tls-pass-certificate-to-upstream">`nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-tls-match-cn" href="#nginx-ingress-kubernetes-ioauth-tls-match-cn" title="#nginx-ingress-kubernetes-ioauth-tls-match-cn">`nginx.ingress.kubernetes.io/auth-tls-match-cn`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-cache-key" href="#nginx-ingress-kubernetes-ioauth-cache-key" title="#nginx-ingress-kubernetes-ioauth-cache-key">`nginx.ingress.kubernetes.io/auth-cache-key`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-cache-duration" href="#nginx-ingress-kubernetes-ioauth-cache-duration" title="#nginx-ingress-kubernetes-ioauth-cache-duration">`nginx.ingress.kubernetes.io/auth-cache-duration`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-keepalive" href="#nginx-ingress-kubernetes-ioauth-keepalive" title="#nginx-ingress-kubernetes-ioauth-keepalive">`nginx.ingress.kubernetes.io/auth-keepalive`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-keepalive-share-vars" href="#nginx-ingress-kubernetes-ioauth-keepalive-share-vars" title="#nginx-ingress-kubernetes-ioauth-keepalive-share-vars">`nginx.ingress.kubernetes.io/auth-keepalive-share-vars`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-keepalive-requests" href="#nginx-ingress-kubernetes-ioauth-keepalive-requests" title="#nginx-ingress-kubernetes-ioauth-keepalive-requests">`nginx.ingress.kubernetes.io/auth-keepalive-requests`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-keepalive-timeout" href="#nginx-ingress-kubernetes-ioauth-keepalive-timeout" title="#nginx-ingress-kubernetes-ioauth-keepalive-timeout">`nginx.ingress.kubernetes.io/auth-keepalive-timeout`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-proxy-set-headers" href="#nginx-ingress-kubernetes-ioauth-proxy-set-headers" title="#nginx-ingress-kubernetes-ioauth-proxy-set-headers">`nginx.ingress.kubernetes.io/auth-proxy-set-headers`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioauth-snippet" href="#nginx-ingress-kubernetes-ioauth-snippet" title="#nginx-ingress-kubernetes-ioauth-snippet">`nginx.ingress.kubernetes.io/auth-snippet`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-global-auth" href="#nginx-ingress-kubernetes-ioenable-global-auth" title="#nginx-ingress-kubernetes-ioenable-global-auth">`nginx.ingress.kubernetes.io/enable-global-auth`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary" href="#nginx-ingress-kubernetes-iocanary" title="#nginx-ingress-kubernetes-iocanary">`nginx.ingress.kubernetes.io/canary`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary-by-header" href="#nginx-ingress-kubernetes-iocanary-by-header" title="#nginx-ingress-kubernetes-iocanary-by-header">`nginx.ingress.kubernetes.io/canary-by-header`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary-by-header-value" href="#nginx-ingress-kubernetes-iocanary-by-header-value" title="#nginx-ingress-kubernetes-iocanary-by-header-value">`nginx.ingress.kubernetes.io/canary-by-header-value`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary-by-header-pattern" href="#nginx-ingress-kubernetes-iocanary-by-header-pattern" title="#nginx-ingress-kubernetes-iocanary-by-header-pattern">`nginx.ingress.kubernetes.io/canary-by-header-pattern`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary-by-cookie" href="#nginx-ingress-kubernetes-iocanary-by-cookie" title="#nginx-ingress-kubernetes-iocanary-by-cookie">`nginx.ingress.kubernetes.io/canary-by-cookie`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary-weight" href="#nginx-ingress-kubernetes-iocanary-weight" title="#nginx-ingress-kubernetes-iocanary-weight">`nginx.ingress.kubernetes.io/canary-weight`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocanary-weight-total" href="#nginx-ingress-kubernetes-iocanary-weight-total" title="#nginx-ingress-kubernetes-iocanary-weight-total">`nginx.ingress.kubernetes.io/canary-weight-total`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioclient-body-buffer-size" href="#nginx-ingress-kubernetes-ioclient-body-buffer-size" title="#nginx-ingress-kubernetes-ioclient-body-buffer-size">`nginx.ingress.kubernetes.io/client-body-buffer-size`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioconfiguration-snippet" href="#nginx-ingress-kubernetes-ioconfiguration-snippet" title="#nginx-ingress-kubernetes-ioconfiguration-snippet">`nginx.ingress.kubernetes.io/configuration-snippet`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iocustom-http-errors" href="#nginx-ingress-kubernetes-iocustom-http-errors" title="#nginx-ingress-kubernetes-iocustom-http-errors">`nginx.ingress.kubernetes.io/custom-http-errors`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iodisable-proxy-intercept-errors" href="#nginx-ingress-kubernetes-iodisable-proxy-intercept-errors" title="#nginx-ingress-kubernetes-iodisable-proxy-intercept-errors">`nginx.ingress.kubernetes.io/disable-proxy-intercept-errors`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iodefault-backend" href="#nginx-ingress-kubernetes-iodefault-backend" title="#nginx-ingress-kubernetes-iodefault-backend">`nginx.ingress.kubernetes.io/default-backend`</a> | Not supported yet; use `defaultBackend` in Ingress spec. |
| <a id="nginx-ingress-kubernetes-iolimit-rate-after" href="#nginx-ingress-kubernetes-iolimit-rate-after" title="#nginx-ingress-kubernetes-iolimit-rate-after">`nginx.ingress.kubernetes.io/limit-rate-after`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iolimit-rate" href="#nginx-ingress-kubernetes-iolimit-rate" title="#nginx-ingress-kubernetes-iolimit-rate">`nginx.ingress.kubernetes.io/limit-rate`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iolimit-whitelist" href="#nginx-ingress-kubernetes-iolimit-whitelist" title="#nginx-ingress-kubernetes-iolimit-whitelist">`nginx.ingress.kubernetes.io/limit-whitelist`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iolimit-rps" href="#nginx-ingress-kubernetes-iolimit-rps" title="#nginx-ingress-kubernetes-iolimit-rps">`nginx.ingress.kubernetes.io/limit-rps`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iolimit-rpm" href="#nginx-ingress-kubernetes-iolimit-rpm" title="#nginx-ingress-kubernetes-iolimit-rpm">`nginx.ingress.kubernetes.io/limit-rpm`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iolimit-burst-multiplier" href="#nginx-ingress-kubernetes-iolimit-burst-multiplier" title="#nginx-ingress-kubernetes-iolimit-burst-multiplier">`nginx.ingress.kubernetes.io/limit-burst-multiplier`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iolimit-connections" href="#nginx-ingress-kubernetes-iolimit-connections" title="#nginx-ingress-kubernetes-iolimit-connections">`nginx.ingress.kubernetes.io/limit-connections`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioglobal-rate-limit" href="#nginx-ingress-kubernetes-ioglobal-rate-limit" title="#nginx-ingress-kubernetes-ioglobal-rate-limit">`nginx.ingress.kubernetes.io/global-rate-limit`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioglobal-rate-limit-window" href="#nginx-ingress-kubernetes-ioglobal-rate-limit-window" title="#nginx-ingress-kubernetes-ioglobal-rate-limit-window">`nginx.ingress.kubernetes.io/global-rate-limit-window`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioglobal-rate-limit-key" href="#nginx-ingress-kubernetes-ioglobal-rate-limit-key" title="#nginx-ingress-kubernetes-ioglobal-rate-limit-key">`nginx.ingress.kubernetes.io/global-rate-limit-key`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioglobal-rate-limit-ignored-cidrs" href="#nginx-ingress-kubernetes-ioglobal-rate-limit-ignored-cidrs" title="#nginx-ingress-kubernetes-ioglobal-rate-limit-ignored-cidrs">`nginx.ingress.kubernetes.io/global-rate-limit-ignored-cidrs`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iopermanent-redirect" href="#nginx-ingress-kubernetes-iopermanent-redirect" title="#nginx-ingress-kubernetes-iopermanent-redirect">`nginx.ingress.kubernetes.io/permanent-redirect`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iopermanent-redirect-code" href="#nginx-ingress-kubernetes-iopermanent-redirect-code" title="#nginx-ingress-kubernetes-iopermanent-redirect-code">`nginx.ingress.kubernetes.io/permanent-redirect-code`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iotemporal-redirect" href="#nginx-ingress-kubernetes-iotemporal-redirect" title="#nginx-ingress-kubernetes-iotemporal-redirect">`nginx.ingress.kubernetes.io/temporal-redirect`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iopreserve-trailing-slash" href="#nginx-ingress-kubernetes-iopreserve-trailing-slash" title="#nginx-ingress-kubernetes-iopreserve-trailing-slash">`nginx.ingress.kubernetes.io/preserve-trailing-slash`</a> | Not supported yet; Traefik preserves by default.         |
| <a id="nginx-ingress-kubernetes-ioproxy-cookie-domain" href="#nginx-ingress-kubernetes-ioproxy-cookie-domain" title="#nginx-ingress-kubernetes-ioproxy-cookie-domain">`nginx.ingress.kubernetes.io/proxy-cookie-domain`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-cookie-path" href="#nginx-ingress-kubernetes-ioproxy-cookie-path" title="#nginx-ingress-kubernetes-ioproxy-cookie-path">`nginx.ingress.kubernetes.io/proxy-cookie-path`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-connect-timeout" href="#nginx-ingress-kubernetes-ioproxy-connect-timeout" title="#nginx-ingress-kubernetes-ioproxy-connect-timeout">`nginx.ingress.kubernetes.io/proxy-connect-timeout`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-send-timeout" href="#nginx-ingress-kubernetes-ioproxy-send-timeout" title="#nginx-ingress-kubernetes-ioproxy-send-timeout">`nginx.ingress.kubernetes.io/proxy-send-timeout`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-read-timeout" href="#nginx-ingress-kubernetes-ioproxy-read-timeout" title="#nginx-ingress-kubernetes-ioproxy-read-timeout">`nginx.ingress.kubernetes.io/proxy-read-timeout`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-next-upstream" href="#nginx-ingress-kubernetes-ioproxy-next-upstream" title="#nginx-ingress-kubernetes-ioproxy-next-upstream">`nginx.ingress.kubernetes.io/proxy-next-upstream`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-next-upstream-timeout" href="#nginx-ingress-kubernetes-ioproxy-next-upstream-timeout" title="#nginx-ingress-kubernetes-ioproxy-next-upstream-timeout">`nginx.ingress.kubernetes.io/proxy-next-upstream-timeout`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-next-upstream-tries" href="#nginx-ingress-kubernetes-ioproxy-next-upstream-tries" title="#nginx-ingress-kubernetes-ioproxy-next-upstream-tries">`nginx.ingress.kubernetes.io/proxy-next-upstream-tries`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-request-buffering" href="#nginx-ingress-kubernetes-ioproxy-request-buffering" title="#nginx-ingress-kubernetes-ioproxy-request-buffering">`nginx.ingress.kubernetes.io/proxy-request-buffering`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-redirect-from" href="#nginx-ingress-kubernetes-ioproxy-redirect-from" title="#nginx-ingress-kubernetes-ioproxy-redirect-from">`nginx.ingress.kubernetes.io/proxy-redirect-from`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-redirect-to" href="#nginx-ingress-kubernetes-ioproxy-redirect-to" title="#nginx-ingress-kubernetes-ioproxy-redirect-to">`nginx.ingress.kubernetes.io/proxy-redirect-to`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-http-version" href="#nginx-ingress-kubernetes-ioproxy-http-version" title="#nginx-ingress-kubernetes-ioproxy-http-version">`nginx.ingress.kubernetes.io/proxy-http-version`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-ciphers" href="#nginx-ingress-kubernetes-ioproxy-ssl-ciphers" title="#nginx-ingress-kubernetes-ioproxy-ssl-ciphers">`nginx.ingress.kubernetes.io/proxy-ssl-ciphers`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-verify-depth" href="#nginx-ingress-kubernetes-ioproxy-ssl-verify-depth" title="#nginx-ingress-kubernetes-ioproxy-ssl-verify-depth">`nginx.ingress.kubernetes.io/proxy-ssl-verify-depth`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-ssl-protocols" href="#nginx-ingress-kubernetes-ioproxy-ssl-protocols" title="#nginx-ingress-kubernetes-ioproxy-ssl-protocols">`nginx.ingress.kubernetes.io/proxy-ssl-protocols`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-rewrite-log" href="#nginx-ingress-kubernetes-ioenable-rewrite-log" title="#nginx-ingress-kubernetes-ioenable-rewrite-log">`nginx.ingress.kubernetes.io/enable-rewrite-log`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iorewrite-target" href="#nginx-ingress-kubernetes-iorewrite-target" title="#nginx-ingress-kubernetes-iorewrite-target">`nginx.ingress.kubernetes.io/rewrite-target`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iosatisfy" href="#nginx-ingress-kubernetes-iosatisfy" title="#nginx-ingress-kubernetes-iosatisfy">`nginx.ingress.kubernetes.io/satisfy`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioserver-alias" href="#nginx-ingress-kubernetes-ioserver-alias" title="#nginx-ingress-kubernetes-ioserver-alias">`nginx.ingress.kubernetes.io/server-alias`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioserver-snippet" href="#nginx-ingress-kubernetes-ioserver-snippet" title="#nginx-ingress-kubernetes-ioserver-snippet">`nginx.ingress.kubernetes.io/server-snippet`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iosession-cookie-conditional-samesite-none" href="#nginx-ingress-kubernetes-iosession-cookie-conditional-samesite-none" title="#nginx-ingress-kubernetes-iosession-cookie-conditional-samesite-none">`nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iosession-cookie-expires" href="#nginx-ingress-kubernetes-iosession-cookie-expires" title="#nginx-ingress-kubernetes-iosession-cookie-expires">`nginx.ingress.kubernetes.io/session-cookie-expires`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iosession-cookie-change-on-failure" href="#nginx-ingress-kubernetes-iosession-cookie-change-on-failure" title="#nginx-ingress-kubernetes-iosession-cookie-change-on-failure">`nginx.ingress.kubernetes.io/session-cookie-change-on-failure`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iossl-ciphers" href="#nginx-ingress-kubernetes-iossl-ciphers" title="#nginx-ingress-kubernetes-iossl-ciphers">`nginx.ingress.kubernetes.io/ssl-ciphers`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iossl-prefer-server-ciphers" href="#nginx-ingress-kubernetes-iossl-prefer-server-ciphers" title="#nginx-ingress-kubernetes-iossl-prefer-server-ciphers">`nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioconnection-proxy-header" href="#nginx-ingress-kubernetes-ioconnection-proxy-header" title="#nginx-ingress-kubernetes-ioconnection-proxy-header">`nginx.ingress.kubernetes.io/connection-proxy-header`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-access-log" href="#nginx-ingress-kubernetes-ioenable-access-log" title="#nginx-ingress-kubernetes-ioenable-access-log">`nginx.ingress.kubernetes.io/enable-access-log`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-opentracing" href="#nginx-ingress-kubernetes-ioenable-opentracing" title="#nginx-ingress-kubernetes-ioenable-opentracing">`nginx.ingress.kubernetes.io/enable-opentracing`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioopentracing-trust-incoming-span" href="#nginx-ingress-kubernetes-ioopentracing-trust-incoming-span" title="#nginx-ingress-kubernetes-ioopentracing-trust-incoming-span">`nginx.ingress.kubernetes.io/opentracing-trust-incoming-span`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-opentelemetry" href="#nginx-ingress-kubernetes-ioenable-opentelemetry" title="#nginx-ingress-kubernetes-ioenable-opentelemetry">`nginx.ingress.kubernetes.io/enable-opentelemetry`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioopentelemetry-trust-incoming-span" href="#nginx-ingress-kubernetes-ioopentelemetry-trust-incoming-span" title="#nginx-ingress-kubernetes-ioopentelemetry-trust-incoming-span">`nginx.ingress.kubernetes.io/opentelemetry-trust-incoming-span`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-modsecurity" href="#nginx-ingress-kubernetes-ioenable-modsecurity" title="#nginx-ingress-kubernetes-ioenable-modsecurity">`nginx.ingress.kubernetes.io/enable-modsecurity`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioenable-owasp-core-rules" href="#nginx-ingress-kubernetes-ioenable-owasp-core-rules" title="#nginx-ingress-kubernetes-ioenable-owasp-core-rules">`nginx.ingress.kubernetes.io/enable-owasp-core-rules`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iomodsecurity-transaction-id" href="#nginx-ingress-kubernetes-iomodsecurity-transaction-id" title="#nginx-ingress-kubernetes-iomodsecurity-transaction-id">`nginx.ingress.kubernetes.io/modsecurity-transaction-id`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iomodsecurity-snippet" href="#nginx-ingress-kubernetes-iomodsecurity-snippet" title="#nginx-ingress-kubernetes-iomodsecurity-snippet">`nginx.ingress.kubernetes.io/modsecurity-snippet`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iomirror-request-body" href="#nginx-ingress-kubernetes-iomirror-request-body" title="#nginx-ingress-kubernetes-iomirror-request-body">`nginx.ingress.kubernetes.io/mirror-request-body`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iomirror-target" href="#nginx-ingress-kubernetes-iomirror-target" title="#nginx-ingress-kubernetes-iomirror-target">`nginx.ingress.kubernetes.io/mirror-target`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iomirror-host" href="#nginx-ingress-kubernetes-iomirror-host" title="#nginx-ingress-kubernetes-iomirror-host">`nginx.ingress.kubernetes.io/mirror-host`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iox-forwarded-prefix" href="#nginx-ingress-kubernetes-iox-forwarded-prefix" title="#nginx-ingress-kubernetes-iox-forwarded-prefix">`nginx.ingress.kubernetes.io/x-forwarded-prefix`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioupstream-hash-by" href="#nginx-ingress-kubernetes-ioupstream-hash-by" title="#nginx-ingress-kubernetes-ioupstream-hash-by">`nginx.ingress.kubernetes.io/upstream-hash-by`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioupstream-vhost" href="#nginx-ingress-kubernetes-ioupstream-vhost" title="#nginx-ingress-kubernetes-ioupstream-vhost">`nginx.ingress.kubernetes.io/upstream-vhost`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iodenylist-source-range" href="#nginx-ingress-kubernetes-iodenylist-source-range" title="#nginx-ingress-kubernetes-iodenylist-source-range">`nginx.ingress.kubernetes.io/denylist-source-range`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iowhitelist-source-range" href="#nginx-ingress-kubernetes-iowhitelist-source-range" title="#nginx-ingress-kubernetes-iowhitelist-source-range">`nginx.ingress.kubernetes.io/whitelist-source-range`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-buffering" href="#nginx-ingress-kubernetes-ioproxy-buffering" title="#nginx-ingress-kubernetes-ioproxy-buffering">`nginx.ingress.kubernetes.io/proxy-buffering`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-buffers-number" href="#nginx-ingress-kubernetes-ioproxy-buffers-number" title="#nginx-ingress-kubernetes-ioproxy-buffers-number">`nginx.ingress.kubernetes.io/proxy-buffers-number`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-buffer-size" href="#nginx-ingress-kubernetes-ioproxy-buffer-size" title="#nginx-ingress-kubernetes-ioproxy-buffer-size">`nginx.ingress.kubernetes.io/proxy-buffer-size`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-ioproxy-max-temp-file-size" href="#nginx-ingress-kubernetes-ioproxy-max-temp-file-size" title="#nginx-ingress-kubernetes-ioproxy-max-temp-file-size">`nginx.ingress.kubernetes.io/proxy-max-temp-file-size`</a> | Not supported yet.                                    |
| <a id="nginx-ingress-kubernetes-iostream-snippet" href="#nginx-ingress-kubernetes-iostream-snippet" title="#nginx-ingress-kubernetes-iostream-snippet">`nginx.ingress.kubernetes.io/stream-snippet`</a> | Not supported yet.                                    |
