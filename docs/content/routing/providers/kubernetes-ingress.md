# Traefik & Kubernetes

The Kubernetes Ingress Controller.
{: .subtitle }

## Routing Configuration

The provider then watches for incoming ingresses events, such as the example below,
and derives the corresponding dynamic configuration from it,
which in turn will create the resulting routers, services, handlers, etc.

```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: foo
  namespace: production

spec:
  rules:
  - host: foo.com
    http:
      paths:
      - path: /bar
        backend:
          serviceName: service1
          servicePort: 80
      - path: /foo
        backend:
          serviceName: service1
          servicePort: 80

  tls:
  - secretName: mySecret
```

### Annotations

??? example
    
    ```yaml tab="Ingress"
    kind: Ingress
    apiVersion: networking.k8s.io/v1beta1
    metadata:
      name: foo
      namespace: production
      annotations:
        traefik.ingress.kubernetes.io/router.entrypoints: web
    
    spec:
      rules:
      - host: foo.com
        http:
          paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
          - path: /foo
            backend:
              serviceName: service1
              servicePort: 80
    ```
    
    ```yaml tab="Service"
    kind: Service
    apiVersion: v1
    metadata:
      name: service1
      namespace: testing
      annotations:
        traefik.ingress.kubernetes.io/service.passhostheader: "false"
    
    spec:
      ports:
      - port: 80
      clusterIp: 10.0.0.1
    ```

#### On Ingress

??? info "`traefik.ingress.kubernetes.io/router.entrypoints`"

    See [entry points](../routers/index.md#entrypoints) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.entrypoints: ep1,ep2
    ```

??? info "`traefik.ingress.kubernetes.io/router.middlewares`"

    See [middlewares](../routers/index.md#middlewares) and [middlewares overview](../../middlewares/overview.md) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.middlewares: auth@file,prefix@kuberntes-crd,cb@file
    ```

??? info "`traefik.ingress.kubernetes.io/router.priority`"

    See [priority](../routers/index.md#priority) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.priority: "42"
    ```

??? info "`traefik.ingress.kubernetes.io/router.pathmatcher`"

    Overrides the default router rule type used for a path.  
    Only path-related matcher name can be specified: `Path`, `PathPrefix`.
    
    Default `PathPrefix`

    ```yaml
    traefik.ingress.kubernetes.io/router.pathmatcher: Path
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls`"

    See [tls](../routers/index.md#tls) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.tls: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.certresolver`"

    See [certResolver](../routers/index.md#certresolver) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.certresolver: myresolver
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.domains.n.main`"

    See [domains](../routers/index.md#domains) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.domains.0.main: foobar.com
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.domains.n.sans`"

    See [domains](../routers/index.md#domains) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.domains.0.sans: test.foobar.com,dev.foobar.com
    ```

??? info "`traefik.ingress.kubernetes.io/router.tls.options`"

    See [options](../routers/index.md#options) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/router.tls.options: foobar
    ```

#### On Service

??? info "`traefik.ingress.kubernetes.io/service.serversscheme`"

    Overrides the default scheme.

    ```yaml
    traefik.ingress.kubernetes.io/service.serversscheme: h2c
    ```

??? info "`traefik.ingress.kubernetes.io/service.passhostheader`"

    See [pass Host header](../services/index.md#pass-host-header) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.passhostheader: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky`"

    See [sticky sessions](../services/index.md#sticky-sessions) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.httponly`"

    See [sticky sessions](../services/index.md#sticky-sessions) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.httponly: "true"
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.name`"

    See [sticky sessions](../services/index.md#sticky-sessions) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.name: foobar
    ```

??? info "`traefik.ingress.kubernetes.io/service.sticky.cookie.secure`"

    See [sticky sessions](../services/index.md#sticky-sessions) for more information.

    ```yaml
    traefik.ingress.kubernetes.io/service.sticky.cookie.secure: "true"
    ```

### TLS

#### Communication Between Traefik and Pods

Traefik automatically requests endpoint information based on the service provided in the ingress spec.
Although Traefik will connect directly to the endpoints (pods),
it still checks the service port to see if TLS communication is required.

There are 3 ways to configure Traefik to use https to communicate with pods:

1. If the service port defined in the ingress spec is `443` (note that you can still use `targetPort` to use a different port on your pod).
1. If the service port defined in the ingress spec has a name that starts with https (such as `https-api`, `https-web` or just `https`).
1. If the ingress spec includes the annotation `traefik.ingress.kubernetes.io/service.serversscheme: https`.

If either of those configuration options exist, then the backend communication protocol is assumed to be TLS,
and will connect via TLS automatically.

!!! info
    
    Please note that by enabling TLS communication between traefik and your pods,
    you will have to have trusted certificates that have the proper trust chain and IP subject name.
    If this is not an option, you may need to skip TLS certificate verification.
    See the [insecureSkipVerify](../../routing/overview.md#insecureskipverify) setting for more details.

#### Certificates Management

??? example "Using a secret"
    
    ```yaml tab="Ingress"
    kind: Ingress
    apiVersion: networking.k8s.io/v1beta1
    metadata:
      name: foo
      namespace: production
    
    spec:
      rules:
      - host: foo.com
        http:
          paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
    
      tls:
      - secretName: supersecret
    ```
      
    ```yaml tab="Secret"
    apiVersion: v1
    kind: Secret
    metadata:
      name: supersecret
    
    data:
      tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
      tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0=
    ```

TLS certificates can be managed in Secrets objects.

!!! info
    
    Only TLS certificates provided by users can be stored in Kubernetes Secrets.
    [Let's Encrypt](../../https/acme.md) certificates cannot be managed in Kubernetes Secrets yet.

## Global Default Backend Ingresses

Ingresses can be created that look like the following:

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
 name: cheese

spec:
 backend:
   serviceName: stilton
   servicePort: 80
```

This ingress follows the Global Default Backend property of ingresses.
This will allow users to create a "default router" that will match all unmatched requests.

!!! info
    
    Due to Traefik's use of priorities, you may have to set this ingress priority lower than other ingresses in your environment,
    to avoid this global ingress from satisfying requests that could match other ingresses.
    
    To do this, use the `traefik.ingress.kubernetes.io/router.priority` annotation (as seen in [Annotations on Ingress](#on-ingress)) on your ingresses accordingly.
