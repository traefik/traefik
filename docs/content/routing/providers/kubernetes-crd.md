---
title: "Routing Configuration for Traefik CRD"
description: "Understand the routing configuration for the Kubernetes IngressRoute & Traefik CRD. Read the technical documentation."
---

# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

## Configuration Examples

??? example "Configuring KubernetesCRD and Deploying/Exposing Services"

    ```yaml tab="Resource Definition"
    # All resources definition must be declared
    --8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml"
    ```
    
    ```yaml tab="RBAC"
    --8<-- "content/reference/dynamic-configuration/kubernetes-crd-rbac.yml"
    ```
    
    ```yaml tab="Traefik"
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: traefik-ingress-controller
    
    ---
    kind: Deployment
    apiVersion: apps/v1
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
              image: traefik:v2.10
              args:
                - --log.level=DEBUG
                - --api
                - --api.insecure
                - --entrypoints.web.address=:80
                - --entrypoints.tcpep.address=:8000
                - --entrypoints.udpep.address=:9000/udp
                - --providers.kubernetescrd
              ports:
                - name: web
                  containerPort: 80
                - name: admin
                  containerPort: 8080
                - name: tcpep
                  containerPort: 8000
                - name: udpep
                  containerPort: 9000
    
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
        - protocol: TCP
          port: 80
          name: web
          targetPort: 80
        - protocol: TCP
          port: 8080
          name: admin
          targetPort: 8080
        - protocol: TCP
          port: 8000
          name: tcpep
          targetPort: 8000
     
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: traefikudp
    spec:
      type: LoadBalancer
      selector:
        app: traefik
      ports:
        - protocol: UDP
          port: 9000
          name: udpep
          targetPort: 9000
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: myingressroute
      namespace: default
    
    spec:
      entryPoints:
        - web
    
      routes:
      - match: Host(`foo`) && PathPrefix(`/bar`)
        kind: Rule
        services:
        - name: whoami
          port: 80
    
    ---
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
    metadata:
      name: ingressroute.tcp
      namespace: default
    
    spec:
      entryPoints:
        - tcpep
      routes:
      - match: HostSNI(`bar`)
        services:
          - name: whoamitcp
            port: 8080
    
    ---
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteUDP
    metadata:
      name: ingressroute.udp
      namespace: default
    
    spec:
      entryPoints:
        - udpep
      routes:
      - services:
          - name: whoamiudp
            port: 8080
    ```
    
    ```yaml tab="Whoami"
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: whoami
      namespace: default
      labels:
        app: traefiklabs
        name: whoami
    
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: traefiklabs
          task: whoami
      template:
        metadata:
          labels:
            app: traefiklabs
            task: whoami
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
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: whoami
    
    ---
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: whoamitcp
      namespace: default
      labels:
        app: traefiklabs
        name: whoamitcp
    
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: traefiklabs
          task: whoamitcp
      template:
        metadata:
          labels:
            app: traefiklabs
            task: whoamitcp
        spec:
          containers:
            - name: whoamitcp
              image: traefik/whoamitcp
              ports:
                - containerPort: 8080
    
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: whoamitcp
      namespace: default
    
    spec:
      ports:
        - protocol: TCP
          port: 8080
      selector:
        app: traefiklabs
        task: whoamitcp
    
    ---
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: whoamiudp
      namespace: default
      labels:
        app: traefiklabs
        name: whoamiudp
    
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: traefiklabs
          task: whoamiudp
      template:
        metadata:
          labels:
            app: traefiklabs
            task: whoamiudp
        spec:
          containers:
            - name: whoamiudp
              image: traefik/whoamiudp:latest
              ports:
                - containerPort: 8080
    
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: whoamiudp
      namespace: default
    
    spec:
      ports:
        - port: 8080
      selector:
        app: traefiklabs
        task: whoamiudp
    ```

## Routing Configuration

### Custom Resource Definition (CRD)

* You can find an exhaustive list, generated from Traefik's source code, of the custom resources and their attributes in [the reference page](../../reference/dynamic-configuration/kubernetes-crd.md).
* Validate that [the prerequisites](../../providers/kubernetes-crd.md#configuration-requirements) are fulfilled before using the Traefik custom resources.
* Traefik CRDs are building blocks that you can assemble according to your needs.
    
You can find an excerpt of the available custom resources in the table below:

| Kind                                       | Purpose                                                            | Concept Behind                                                 |
|--------------------------------------------|--------------------------------------------------------------------|----------------------------------------------------------------|
| [IngressRoute](#kind-ingressroute)         | HTTP Routing                                                       | [HTTP router](../routers/index.md#configuring-http-routers)    |
| [Middleware](#kind-middleware)             | Tweaks the HTTP requests before they are sent to your service      | [HTTP Middlewares](../../middlewares/http/overview.md)         |
| [TraefikService](#kind-traefikservice)     | Abstraction for HTTP loadbalancing/mirroring                       | [HTTP service](../services/index.md#configuring-http-services) |
| [IngressRouteTCP](#kind-ingressroutetcp)   | TCP Routing                                                        | [TCP router](../routers/index.md#configuring-tcp-routers)      |
| [MiddlewareTCP](#kind-middlewaretcp)       | Tweaks the TCP requests before they are sent to your service       | [TCP Middlewares](../../middlewares/tcp/overview.md)           |
| [IngressRouteUDP](#kind-ingressrouteudp)   | UDP Routing                                                        | [UDP router](../routers/index.md#configuring-udp-routers)      |
| [TLSOptions](#kind-tlsoption)              | Allows to configure some parameters of the TLS connection          | [TLSOptions](../../https/tls.md#tls-options)                   |
| [TLSStores](#kind-tlsstore)                | Allows to configure the default TLS store                          | [TLSStores](../../https/tls.md#certificates-stores)            |
| [ServersTransport](#kind-serverstransport) | Allows to configure the transport between Traefik and the backends | [ServersTransport](../../services/#serverstransport_1)         |

### Kind: `IngressRoute`

`IngressRoute` is the CRD implementation of a [Traefik HTTP router](../routers/index.md#configuring-http-routers).

Register the `IngressRoute` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `IngressRoute` objects.

!!! info "IngressRoute Attributes"

    ```yaml
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: foo
      namespace: bar
    spec:
      entryPoints:                      # [1]
        - foo
      routes:                           # [2]
      - kind: Rule
        match: Host(`test.example.com`) # [3]
        priority: 10                    # [4]
        middlewares:                    # [5]
        - name: middleware1             # [6]
          namespace: default            # [7]
        services:                       # [8]
        - kind: Service
          name: foo
          namespace: default
          passHostHeader: true
          port: 80                      # [9]
          responseForwarding:
            flushInterval: 1ms
          scheme: https
          serversTransport: transport   # [10]
          sticky:
            cookie:
              httpOnly: true
              name: cookie
              secure: true
              sameSite: none
          strategy: RoundRobin
          weight: 10
          nativeLB: true                # [11]
      tls:                              # [12]
        secretName: supersecret         # [13]
        options:                        # [14]
          name: opt                     # [15]
          namespace: default            # [16]
        certResolver: foo               # [17]
        domains:                        # [18]
        - main: example.net             # [19]
          sans:                         # [20]
          - a.example.net
          - b.example.net
    ```

| Ref  | Attribute                      | Purpose                                                                                                                                                                                                                                                                                      |
|------|--------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `entryPoints`                  | List of [entry points](../routers/index.md#entrypoints) names                                                                                                                                                                                                                                |
| [2]  | `routes`                       | List of routes                                                                                                                                                                                                                                                                               |
| [3]  | `routes[n].match`              | Defines the [rule](../routers/index.md#rule) corresponding to an underlying router.                                                                                                                                                                                                          |
| [4]  | `routes[n].priority`           | Defines the [priority](../routers/index.md#priority) to disambiguate rules of the same length, for route matching                                                                                                                                                                            |
| [5]  | `routes[n].middlewares`        | List of reference to [Middleware](#kind-middleware)                                                                                                                                                                                                                                          |
| [6]  | `middlewares[n].name`          | Defines the [Middleware](#kind-middleware) name                                                                                                                                                                                                                                              |
| [7]  | `middlewares[n].namespace`     | Defines the [Middleware](#kind-middleware) namespace                                                                                                                                                                                                                                         |
| [8]  | `routes[n].services`           | List of any combination of [TraefikService](#kind-traefikservice) and reference to a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) (See below for `ExternalName Service` setup)                                                                     |
| [9]  | `services[n].port`             | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.                                                                                                                                       |
| [10] | `services[n].serversTransport` | Defines the reference to a [ServersTransport](#kind-serverstransport). The ServersTransport namespace is assumed to be the [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) namespace (see [ServersTransport reference](#serverstransport-reference)). |
| [11] | `services[n].nativeLB`         | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.                                                                                                                                     |
| [12] | `tls`                          | Defines [TLS](../routers/index.md#tls) certificate configuration                                                                                                                                                                                                                             |
| [13] | `tls.secretName`               | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace)                                                                                                                                         |
| [14] | `tls.options`                  | Defines the reference to a [TLSOption](#kind-tlsoption)                                                                                                                                                                                                                                      |
| [15] | `options.name`                 | Defines the [TLSOption](#kind-tlsoption) name                                                                                                                                                                                                                                                |
| [16] | `options.namespace`            | Defines the [TLSOption](#kind-tlsoption) namespace                                                                                                                                                                                                                                           |
| [17] | `tls.certResolver`             | Defines the reference to a [CertResolver](../routers/index.md#certresolver)                                                                                                                                                                                                                  |
| [18] | `tls.domains`                  | List of [domains](../routers/index.md#domains)                                                                                                                                                                                                                                               |
| [19] | `domains[n].main`              | Defines the main domain name                                                                                                                                                                                                                                                                 |
| [20] | `domains[n].sans`              | List of SANs (alternative domains)                                                                                                                                                                                                                                                           |

??? example "Declaring an IngressRoute"

    ```yaml tab="IngressRoute"
    # All resources definition must be declared
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test-name
      namespace: default
    spec:
      entryPoints:
        - web
      routes:
      - kind: Rule
        match: Host(`test.example.com`)
        middlewares:
        - name: middleware1
          namespace: default
        priority: 10
        services:
        - kind: Service
          name: foo
          namespace: default
          passHostHeader: true
          port: 80
          responseForwarding:
            flushInterval: 1ms
          scheme: https
          sticky:
            cookie:
              httpOnly: true
              name: cookie
              secure: true
          strategy: RoundRobin
          weight: 10
      tls:
        certResolver: foo
        domains:
        - main: example.net
          sans:
          - a.example.net
          - b.example.net
        options:
          name: opt
          namespace: default
        secretName: supersecret
    ```

    ```yaml tab="Middlewares"
    # All resources definition must be declared
    # Prefixing with /foo
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: middleware1
      namespace: default
    spec:
      addPrefix:
        prefix: /foo
    ```

    ```yaml tab="TLSOption"
    apiVersion: traefik.io/v1alpha1
    kind: TLSOption
    metadata:
      name: opt
      namespace: default
    
    spec:
      minVersion: VersionTLS12
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

!!! important "Configuring Backend Protocol"

    There are 3 ways to configure the backend protocol for communication between Traefik and your pods:
	
    - Setting the scheme explicitly (http/https/h2c)
    - Configuring the name of the kubernetes service port to start with https (https)
    - Setting the kubernetes service port to use port 443 (https)

    If you do not configure the above, Traefik will assume an http connection.
    

!!! important "Using Kubernetes ExternalName Service"

    Traefik backends creation needs a port to be set, however Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) could be defined without any port.
    Accordingly, Traefik supports defining a port in two ways:
    
    - only on `IngressRoute` service
    - on both sides, you'll be warned if the ports don't match, and the `IngressRoute` service port is used
    
    Thus, in case of two sides port definition, Traefik expects a match between ports.
    
    ??? example "Examples"
        
        ```yaml tab="IngressRoute"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRoute
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - match: Host(`example.net`)
            kind: Rule
            services:
            - name: external-svc
              port: 80
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
        ```
        
        ```yaml tab="ExternalName Service"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRoute
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - match: Host(`example.net`)
            kind: Rule
            services:
            - name: external-svc
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
          ports:
            - port: 80
        ```
        
        ```yaml tab="Both sides"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRoute
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - match: Host(`example.net`)
            kind: Rule
            services:
            - name: external-svc
              port: 80
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
          ports:
            - port: 80
        ```

#### Load Balancing

More information in the dedicated server [load balancing](../services/index.md#load-balancing) section.

!!! info "Declaring and using Kubernetes Service Load Balancing"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/foo`)
        kind: Rule
        services:
        - name: svc1
          namespace: default
        - name: svc2
          namespace: default
    ```
    
    ```yaml tab="K8s Service"
    apiVersion: v1
    kind: Service
    metadata:
      name: svc1
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app1
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: svc2
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app2
    ```

!!! important "Kubernetes Service Native Load-Balancing"

    To avoid creating the server load-balancer with the pods IPs and use Kubernetes Service clusterIP directly,
    one should set the service `NativeLB` option to true.
    Please note that, by default, Traefik reuses the established connections to the backends for performance purposes. This can prevent the requests load balancing between the replicas from behaving as one would expect when the option is set.
    By default, `NativeLB` is false.

    ??? example "Example"

        ```yaml
        ---
        apiVersion: traefik.containo.us/v1alpha1
        kind: IngressRoute
        metadata:
          name: test.route
          namespace: default

        spec:
          entryPoints:
            - foo

          routes:
          - match: Host(`example.net`)
            kind: Rule
            services:
            - name: svc
              port: 80
              # Here, nativeLB instructs to build the servers load balancer with the Kubernetes Service clusterIP only.
              nativeLB: true

        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: svc
          namespace: default
        spec:
          type: ClusterIP
          ...
        ```

### Kind: `Middleware`

`Middleware` is the CRD implementation of a [Traefik middleware](../../middlewares/http/overview.md).

Register the `Middleware` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `Middleware` objects or referencing middlewares in the [`IngressRoute`](#kind-ingressroute) objects.

??? "Declaring and Referencing a Middleware"
    
    ```yaml tab="Middleware"
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: stripprefix
      namespace: foo
    
    spec:
      stripPrefix:
        prefixes:
          - /stripit
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/stripit`)
        kind: Rule
        services:
        - name: whoami
          port: 80
        middlewares:
        - name: stripprefix
          namespace: foo
    ```

!!! important "Cross-provider namespace"

    As Kubernetes also has its own notion of namespace, one should not confuse the kubernetes namespace of a resource
    (in the reference to the middleware) with the [provider namespace](../../providers/overview.md#provider-namespace),
    when the definition of the middleware comes from another provider.
    In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored.
    Additionally, when you want to reference a Middleware from the CRD Provider,
    you have to append the namespace of the resource in the resource-name as Traefik appends the namespace internally automatically.

More information about available middlewares in the dedicated [middlewares section](../../middlewares/http/overview.md).

### Kind: `TraefikService`

`TraefikService` is the CRD implementation of a ["Traefik Service"](../services/index.md).

Register the `TraefikService` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `TraefikService` objects,
referencing services in the [`IngressRoute`](#kind-ingressroute) objects, or recursively in others `TraefikService` objects.

!!! info "Disambiguate Traefik and Kubernetes Services"

    As the field `name` can reference different types of objects, use the field `kind` to avoid any ambiguity.
    
    The field `kind` allows the following values:
    
    * `Service` (default value): to reference a [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/)
    * `TraefikService`: to reference another [Traefik Service](../services/index.md)

`TraefikService` object allows to use any (valid) combinations of:

* [Weighted Round Robin](#weighted-round-robin) load balancing.
* [Mirroring](#mirroring).

#### Weighted Round Robin

More information in the dedicated [Weighted Round Robin](../services/index.md#weighted-round-robin-service) service load balancing section.

??? "Declaring and Using Weighted Round Robin"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/foo`)
        kind: Rule
        services:
        - name: wrr1
          namespace: default
          kind: TraefikService
    ```
    
    ```yaml tab="Weighted Round Robin"
    apiVersion: traefik.io/v1alpha1
    kind: TraefikService
    metadata:
      name: wrr1
      namespace: default
    
    spec:
      weighted:
        services:
          - name: svc1
            port: 80
            weight: 1
          - name: wrr2
            kind: TraefikService
            weight: 1
          - name: mirror1
            kind: TraefikService
            weight: 1

    ---
    apiVersion: traefik.io/v1alpha1
    kind: TraefikService
    metadata:
      name: wrr2
      namespace: default
    
    spec:
      weighted:
        services:
          - name: svc2
            port: 80
            weight: 1
          - name: svc3
            port: 80
            weight: 1
    ```
    
    ```yaml tab="K8s Service"
    apiVersion: v1
    kind: Service
    metadata:
      name: svc1
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app1
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: svc2
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app2
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: svc3
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app3
    ```

#### Mirroring

More information in the dedicated [mirroring](../services/index.md#mirroring-service) service section.

??? "Declaring and Using Mirroring"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/foo`)
        kind: Rule
        services:
        - name: mirror1
          namespace: default
          kind: TraefikService
    ```
    
    ```yaml tab="Mirroring k8s Service"
    # Mirroring from a k8s Service
    apiVersion: traefik.io/v1alpha1
    kind: TraefikService
    metadata:
      name: mirror1
      namespace: default
    
    spec:
      mirroring:
        name: svc1
        port: 80
        mirrors:
          - name: svc2
            port: 80
            percent: 20
          - name: svc3
            kind: TraefikService
            percent: 20
    ```
    
    ```yaml tab="Mirroring Traefik Service"
    # Mirroring from a Traefik Service
    apiVersion: traefik.io/v1alpha1
    kind: TraefikService
    metadata:
      name: mirror1
      namespace: default
    
    spec:
      mirroring:
        name: wrr1
        kind: TraefikService
         mirrors:
           - name: svc2
             port: 80
             percent: 20
           - name: svc3
             kind: TraefikService
             percent: 20
    ```

    ```yaml tab="K8s Service"
    apiVersion: v1
    kind: Service
    metadata:
      name: svc1
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app1
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: svc2
      namespace: default
    
    spec:
      ports:
        - name: http
          port: 80
      selector:
        app: traefiklabs
        task: app2
    ```

!!! important "References and namespaces"

    If the optional `namespace` attribute is not set, the configuration will be applied with the namespace of the current resource.
    
    Additionally, when the definition of the `TraefikService` is from another provider,
    the cross-provider syntax (`service@provider`) should be used to refer to the `TraefikService`, just as in the middleware case.
    
    Specifying a namespace attribute in this case would not make any sense, and will be ignored (except if the provider is `kubernetescrd`).

#### Stickiness and load-balancing

As explained in the section about [Sticky sessions](../../services/#sticky-sessions), for stickiness to work all the way,
it must be specified at each load-balancing level.

For instance, in the example below, there is a first level of load-balancing because there is a (Weighted Round Robin) load-balancing of the two `whoami` services,
and there is a second level because each whoami service is a `replicaset` and is thus handled as a load-balancer of servers.

??? "Stickiness on two load-balancing levels"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default

    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/foo`)
        kind: Rule
        services:
        - name: wrr1
          namespace: default
          kind: TraefikService
    ```

    ```yaml tab="Weighted Round Robin"
    apiVersion: traefik.io/v1alpha1
    kind: TraefikService
    metadata:
      name: wrr1
      namespace: default

    spec:
      weighted:
        services:
          - name: whoami1
            kind: Service
            port: 80
            weight: 1
            sticky:
              cookie:
                name: lvl2
          - name: whoami2
            kind: Service
            weight: 1
            port: 80
            sticky:
              cookie:
                name: lvl2
        sticky:
          cookie:
            name: lvl1
    ```

    ```yaml tab="K8s Service"
    apiVersion: v1
    kind: Service
    metadata:
      name: whoami1

    spec:
      ports:
        - protocol: TCP
          name: web
          port: 80
      selector:
        app: whoami1

    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: whoami2

    spec:
      ports:
        - protocol: TCP
          name: web
          port: 80
      selector:
        app: whoami2
    ```

    ```yaml tab="Deployment (to illustrate replicas)"
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      namespace: default
      name: whoami1
      labels:
        app: whoami1

    spec:
      replicas: 2
      selector:
        matchLabels:
          app: whoami1
      template:
        metadata:
          labels:
            app: whoami1
        spec:
          containers:
            - name: whoami1
              image: traefik/whoami
              ports:
                - name: web
                  containerPort: 80

    ---
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      namespace: default
      name: whoami2
      labels:
        app: whoami2

    spec:
      replicas: 2
      selector:
        matchLabels:
          app: whoami2
      template:
        metadata:
          labels:
            app: whoami2
        spec:
          containers:
            - name: whoami2
              image: traefik/whoami
              ports:
                - name: web
                  containerPort: 80
    ```

    To keep a session open with the same server, the client would then need to specify the two levels within the cookie for each request, e.g. with curl:

    ```bash
    curl -H Host:example.com -b "lvl1=default-whoami1-80; lvl2=http://10.42.0.6:80" http://localhost:8000/foo
    ```

    assuming `10.42.0.6` is the IP address of one of the replicas (a pod then) of the `whoami1` service.

### Kind: `IngressRouteTCP`

`IngressRouteTCP` is the CRD implementation of a [Traefik TCP router](../routers/index.md#configuring-tcp-routers).

Register the `IngressRouteTCP` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `IngressRouteTCP` objects.

!!! info "IngressRouteTCP Attributes"

    ```yaml
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
    metadata:
      name: ingressroutetcpfoo
    
    spec:
      entryPoints:                  # [1]
        - footcp
      routes:                       # [2]
      - match: HostSNI(`*`)         # [3]
        priority: 10                # [4]
        middlewares:
        - name: middleware1         # [5]
          namespace: default        # [6]
        services:                   # [7]
        - name: foo                 # [8]
          port: 8080                # [9]
          weight: 10                # [10]
          terminationDelay: 400     # [11]
          proxyProtocol:            # [12]
            version: 1              # [13]
          nativeLB: true            # [14]
      tls:                          # [15]
        secretName: supersecret     # [16]
        options:                    # [17]
          name: opt                 # [18]
          namespace: default        # [19]
        certResolver: foo           # [20]
        domains:                    # [21]
        - main: example.net         # [22]
          sans:                     # [23]
          - a.example.net
          - b.example.net
        passthrough: false          # [24]
    ```

| Ref  | Attribute                           | Purpose                                                                                                                                                                                                                                                                                                                                                                              |
|------|-------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `entryPoints`                       | List of [entrypoints](../routers/index.md#entrypoints_1) names                                                                                                                                                                                                                                                                                                                       |
| [2]  | `routes`                            | List of routes                                                                                                                                                                                                                                                                                                                                                                       |
| [3]  | `routes[n].match`                   | Defines the [rule](../routers/index.md#rule_1) of the underlying router                                                                                                                                                                                                                                                                                                              |
| [4]  | `routes[n].priority`                | Defines the [priority](../routers/index.md#priority_1) to disambiguate rules of the same length, for route matching                                                                                                                                                                                                                                                                  |
| [5]  | `middlewares[n].name`               | Defines the [MiddlewareTCP](#kind-middlewaretcp) name                                                                                                                                                                                                                                                                                                                                |
| [6]  | `middlewares[n].namespace`          | Defines the [MiddlewareTCP](#kind-middlewaretcp) namespace                                                                                                                                                                                                                                                                                                                           |
| [7]  | `routes[n].services`                | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions  (See below for `ExternalName Service` setup)                                                                                                                                                                                                                             |
| [8]  | `services[n].name`                  | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)                                                                                                                                                                                                                                                                         |
| [9]  | `services[n].port`                  | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.                                                                                                                                                                                                                               |
| [10] | `services[n].weight`                | Defines the weight to apply to the server load balancing                                                                                                                                                                                                                                                                                                                             |
| [11] | `services[n].terminationDelay`      | corresponds to the deadline that the proxy sets, after one of its connected peers indicates it has closed the writing capability of its connection, to close the reading capability as well, hence fully terminating the connection. It is a duration in milliseconds, defaulting to 100. A negative value means an infinite deadline (i.e. the reading capability is never closed). |
| [12] | `services[n].proxyProtocol`         | Defines the [PROXY protocol](../services/index.md#proxy-protocol) configuration                                                                                                                                                                                                                                                                                                      |
| [13] | `services[n].proxyProtocol.version` | Defines the [PROXY protocol](../services/index.md#proxy-protocol) version                                                                                                                                                                                                                                                                                                            |
| [14] | `services[n].nativeLB`              | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP.                                                                                                                                                                                                                             |
| [15] | `tls`                               | Defines [TLS](../routers/index.md#tls_1) certificate configuration                                                                                                                                                                                                                                                                                                                   |
| [16] | `tls.secretName`                    | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace)                                                                                                                                                                                                                                 |
| [17] | `tls.options`                       | Defines the reference to a [TLSOption](#kind-tlsoption)                                                                                                                                                                                                                                                                                                                              |
| [18] | `options.name`                      | Defines the [TLSOption](#kind-tlsoption) name                                                                                                                                                                                                                                                                                                                                        |
| [19] | `options.namespace`                 | Defines the [TLSOption](#kind-tlsoption) namespace                                                                                                                                                                                                                                                                                                                                   |
| [20] | `tls.certResolver`                  | Defines the reference to a [CertResolver](../routers/index.md#certresolver_1)                                                                                                                                                                                                                                                                                                        |
| [21] | `tls.domains`                       | List of [domains](../routers/index.md#domains_1)                                                                                                                                                                                                                                                                                                                                     |
| [22] | `domains[n].main`                   | Defines the main domain name                                                                                                                                                                                                                                                                                                                                                         |
| [23] | `domains[n].sans`                   | List of SANs (alternative domains)                                                                                                                                                                                                                                                                                                                                                   |
| [24] | `tls.passthrough`                   | If `true`, delegates the TLS termination to the backend                                                                                                                                                                                                                                                                                                                              |

??? example "Declaring an IngressRouteTCP"

    ```yaml tab="IngressRouteTCP"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
    metadata:
      name: ingressroutetcpfoo
    
    spec:
      entryPoints:
        - footcp
      routes:
      # Match is the rule corresponding to an underlying router.
      - match: HostSNI(`*`)
        priority: 10
        services:
        - name: foo
          port: 8080
          terminationDelay: 400
          weight: 10
        - name: bar
          port: 8081
          terminationDelay: 500
          weight: 10
      tls:
        certResolver: foo
        domains:
        - main: example.net
          sans:
          - a.example.net
          - b.example.net
        options:
          name: opt
          namespace: default
        secretName: supersecret
        passthrough: false
    ```
    
    ```yaml tab="TLSOption"
    apiVersion: traefik.io/v1alpha1
    kind: TLSOption
    metadata:
      name: opt
      namespace: default
    
    spec:
      minVersion: VersionTLS12
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

!!! important "Using Kubernetes ExternalName Service"

    Traefik backends creation needs a port to be set, however Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) could be defined without any port.
    Accordingly, Traefik supports defining a port in two ways:
    
    - only on `IngressRouteTCP` service
    - on both sides, you'll be warned if the ports don't match, and the `IngressRouteTCP` service port is used
    
    Thus, in case of two sides port definition, Traefik expects a match between ports.
    
    ??? example "Examples"
        
        ```yaml tab="Only on IngressRouteTCP"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRouteTCP
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - match: HostSNI(`*`)
            services:
            - name: external-svc
              port: 80
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
        ```
        
        ```yaml tab="On both sides"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRouteTCP
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - match: HostSNI(`*`)
            services:
            - name: external-svc
              port: 80
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
          ports:
            - port: 80
        ```

!!! important "Kubernetes Service Native Load-Balancing"

    To avoid creating the server load-balancer with the pods IPs and use Kubernetes Service clusterIP directly,
    one should set the TCP service `NativeLB` option to true.
    By default, `NativeLB` is false.

    ??? example "Examples"

        ```yaml
        ---
        apiVersion: traefik.containo.us/v1alpha1
        kind: IngressRouteTCP
        metadata:
          name: test.route
          namespace: default

        spec:
          entryPoints:
            - foo

          routes:
          - match: HostSNI(`*`)
            services:
            - name: svc
              port: 80
              # Here, nativeLB instructs to build the servers load balancer with the Kubernetes Service clusterIP only.
              nativeLB: true

        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: svc
          namespace: default
        spec:
          type: ClusterIP
          ...
        ```

### Kind: `MiddlewareTCP`

`MiddlewareTCP` is the CRD implementation of a [Traefik TCP middleware](../../middlewares/tcp/overview.md).

Register the `MiddlewareTCP` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `MiddlewareTCP` objects or referencing TCP middlewares in the [`IngressRouteTCP`](#kind-ingressroutetcp) objects.

??? "Declaring and Referencing a MiddlewareTCP "

    ```yaml tab="Middleware"
    apiVersion: traefik.io/v1alpha1
    kind: MiddlewareTCP
    metadata:
      name: ipwhitelist
    spec:
      ipWhiteList:
        sourceRange:
          - 127.0.0.1/32
          - 192.168.1.7
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/whitelist`)
        kind: Rule
        services:
        - name: whoami
          port: 80
        middlewares:
        - name: ipwhitelist
          namespace: foo
    ```

!!! important "Cross-provider namespace"

    As Kubernetes also has its own notion of namespace, one should not confuse the kubernetes namespace of a resource
    (in the reference to the middleware) with the [provider namespace](../../providers/overview.md#provider-namespace),
    when the definition of the TCP middleware comes from another provider.
    In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored.
    Additionally, when you want to reference a MiddlewareTCP from the CRD Provider,
    you have to append the namespace of the resource in the resource-name as Traefik appends the namespace internally automatically.

More information about available TCP middlewares in the dedicated [middlewares section](../../middlewares/tcp/overview.md).

### Kind: `IngressRouteUDP`

`IngressRouteUDP` is the CRD implementation of a [Traefik UDP router](../routers/index.md#configuring-udp-routers).

Register the `IngressRouteUDP` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `IngressRouteUDP` objects.

!!! info "IngressRouteUDP Attributes"

    ```yaml
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteUDP
    metadata:
      name: ingressrouteudpfoo
    
    spec:
      entryPoints:                  # [1]
        - fooudp
      routes:                       # [2]
      - services:                   # [3]
        - name: foo                 # [4]
          port: 8080                # [5]
          weight: 10                # [6]
          nativeLB: true            # [7]
    ```

| Ref | Attribute                     | Purpose                                                                                                                                                  |
|-----|-------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1] | `entryPoints`                 | List of [entrypoints](../routers/index.md#entrypoints_1) names                                                                                           |
| [2] | `routes`                      | List of routes                                                                                                                                           |
| [3] | `routes[n].services`          | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions (See below for `ExternalName Service` setup)  |
| [4] | `services[n].name`            | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)                                             |
| [5] | `services[n].port`            | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/). This can be a reference to a named port.   |
| [6] | `services[n].weight`          | Defines the weight to apply to the server load balancing                                                                                                 |
| [7] | `services[n].nativeLB`        | Controls, when creating the load-balancer, whether the LB's children are directly the pods IPs or if the only child is the Kubernetes Service clusterIP. |

??? example "Declaring an IngressRouteUDP"

    ```yaml
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteUDP
    metadata:
      name: ingressrouteudpfoo
    
    spec:
      entryPoints:
        - fooudp
      routes:
      - services:
        - name: foo
          port: 8080
          weight: 10
        - name: bar
          port: 8081
          weight: 10
    ```

!!! important "Using Kubernetes ExternalName Service"

    Traefik backends creation needs a port to be set, however Kubernetes [ExternalName Service](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) could be defined without any port.
    Accordingly, Traefik supports defining a port in two ways:
    
    - only on `IngressRouteUDP` service
    - on both sides, you'll be warned if the ports don't match, and the `IngressRouteUDP` service port is used
    
    Thus, in case of two sides port definition, Traefik expects a match between ports.
    
    ??? example "Examples"
        
        ```yaml tab="IngressRouteUDP"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRouteUDP
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - services:
            - name: external-svc
              port: 80
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
        ```
        
        ```yaml tab="ExternalName Service"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRouteUDP
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - services:
            - name: external-svc
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
          ports:
            - port: 80
        ```
        
        ```yaml tab="Both sides"
        ---
        apiVersion: traefik.io/v1alpha1
        kind: IngressRouteUDP
        metadata:
          name: test.route
          namespace: default
        
        spec:
          entryPoints:
            - foo
        
          routes:
          - services:
            - name: external-svc
              port: 80
        
        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: external-svc
          namespace: default
        spec:
          externalName: external.domain
          type: ExternalName
          ports:
            - port: 80
        ```

!!! important "Kubernetes Service Native Load-Balancing"

    To avoid creating the server load-balancer with the pods IPs and use Kubernetes Service clusterIP directly,
    one should set the UDP service `NativeLB` option to true.
    By default, `NativeLB` is false.

    ??? example "Example"

        ```yaml
        ---
        apiVersion: traefik.containo.us/v1alpha1
        kind: IngressRouteUDP
        metadata:
          name: test.route
          namespace: default

        spec:
          entryPoints:
            - foo

          routes:
          - services:
            - name: svc
              port: 80
              # Here, nativeLB instructs to build the servers load balancer with the Kubernetes Service clusterIP only.
              nativeLB: true

        ---
        apiVersion: v1
        kind: Service
        metadata:
          name: svc
          namespace: default
        spec:
          type: ClusterIP
          ...
        ```

### Kind: `TLSOption`

`TLSOption` is the CRD implementation of a [Traefik "TLS Option"](../../https/tls.md#tls-options).

Register the `TLSOption` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `TLSOption` objects
or referencing TLS options in the [`IngressRoute`](#kind-ingressroute) / [`IngressRouteTCP`](#kind-ingressroutetcp) objects.

!!! info "TLSOption Attributes"
   
    ```yaml tab="TLSOption"
    apiVersion: traefik.io/v1alpha1
    kind: TLSOption
    metadata:
      name: mytlsoption                             # [1]
      namespace: default
    
    spec:
      minVersion: VersionTLS12                      # [2]
      maxVersion: VersionTLS13                      # [3]
      curvePreferences:                             # [4]
        - CurveP521
        - CurveP384
      cipherSuites:                                 # [5]
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        - TLS_RSA_WITH_AES_256_GCM_SHA384
      clientAuth:                                   # [6]
        secretNames:                                # [7]
          - secret-ca1
          - secret-ca2
        clientAuthType: VerifyClientCertIfGiven     # [8]
      sniStrict: true                               # [9]
      alpnProtocols:                                # [10]
        - foobar
    ```

| Ref  | Attribute                   | Purpose                                                                                                                                                                                                                    |
|------|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `name`                      | Defines the name of the TLSOption resource. One can use `default` as name to redefine the [default TLSOption](../../https/tls.md#tls-options).                                                                             |
| [2]  | `minVersion`                | Defines the [minimum TLS version](../../https/tls.md#minimum-tls-version) that is acceptable.                                                                                                                              |
| [3]  | `maxVersion`                | Defines the [maximum TLS version](../../https/tls.md#maximum-tls-version) that is acceptable.                                                                                                                              |
| [4]  | `cipherSuites`              | list of supported [cipher suites](../../https/tls.md#cipher-suites) for TLS versions up to TLS 1.2.                                                                                                                        |
| [5]  | `curvePreferences`          | List of the [elliptic curves references](../../https/tls.md#curve-preferences) that will be used in an ECDHE handshake, in preference order.                                                                               |
| [6]  | `clientAuth`                | determines the server's policy for TLS [Client Authentication](../../https/tls.md#client-authentication-mtls).                                                                                                             |
| [7]  | `clientAuth.secretNames`    | list of names of the referenced Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) (in TLSOption namespace). The secret must contain a certificate under either a `tls.ca` or a `ca.crt` key. |
| [8]  | `clientAuth.clientAuthType` | defines the client authentication type to apply. The available values are: `NoClientCert`, `RequestClientCert`, `VerifyClientCertIfGiven` and `RequireAndVerifyClientCert`.                                                |
| [9]  | `sniStrict`                 | if `true`, Traefik won't allow connections from clients connections that do not specify a server_name extension.                                                                                                           |
| [10] | `alpnProtocols`             | List of supported [application level protocols](../../https/tls.md#alpn-protocols) for the TLS handshake, in order of preference.                                                                                          |

!!! info "CA Secret"

    The CA secret must contain a base64 encoded certificate under either a `tls.ca` or a `ca.crt` key.

??? example "Declaring and referencing a TLSOption"
   
    ```yaml tab="TLSOption"
    apiVersion: traefik.io/v1alpha1
    kind: TLSOption
    metadata:
      name: mytlsoption
      namespace: default
    
    spec:
      minVersion: VersionTLS12
      sniStrict: true
      cipherSuites:
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        - TLS_RSA_WITH_AES_256_GCM_SHA384
      clientAuth:
        secretNames:
          - secret-ca1
          - secret-ca2
        clientAuthType: VerifyClientCertIfGiven
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`) && PathPrefix(`/stripit`)
        kind: Rule
        services:
        - name: whoami
          port: 80
      tls:
        options: 
          name: mytlsoption
          namespace: default
    ```

    ```yaml tab="Secrets"
    apiVersion: v1
    kind: Secret
    metadata:
      name: secret-ca1
      namespace: default
    
    data:
      # Must contain a certificate under either a `tls.ca` or a `ca.crt` key.
      tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
    
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: secret-ca2
      namespace: default
    
    data:
      # Must contain a certificate under either a `tls.ca` or a `ca.crt` key. 
      tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
    ```
        
!!! important "References and namespaces"

    If the optional `namespace` attribute is not set, the configuration will be applied with the namespace of the IngressRoute.

	Additionally, when the definition of the TLS option is from another provider,
	the cross-provider [syntax](../../providers/overview.md#provider-namespace) (`middlewarename@provider`) should be used to refer to the TLS option.
	Specifying a namespace attribute in this case would not make any sense, and will be ignored.

### Kind: `TLSStore`

`TLSStore` is the CRD implementation of a [Traefik "TLS Store"](../../https/tls.md#certificates-stores).

Register the `TLSStore` kind in the Kubernetes cluster before creating `TLSStore` objects.

!!! important "Default TLS Store"

    Traefik currently only uses the [TLS Store named "default"](../../https/tls.md#certificates-stores).
    This _default_ `TLSStore` should be in a namespace discoverable by Traefik. Since it is used by default on [`IngressRoute`](#kind-ingressroute) and [`IngressRouteTCP`](#kind-ingressroutetcp) objects, there never is a need to actually reference it.
    This means that you cannot have two stores that are named default in different Kubernetes namespaces.
    As a consequence, with respect to TLS stores, the only change that makes sense (and only if needed) is to configure the default TLSStore.

!!! info "TLSStore Attributes"
    ```yaml tab="TLSStore"
    apiVersion: traefik.io/v1alpha1
    kind: TLSStore
    metadata:
      name: default

    spec:
      certificates:                            # [1]
        - secretName: foo                      
        - secretName: bar
      defaultCertificate:                      # [2]
        secretName: secret                     
    ```

| Ref | Attribute            | Purpose                                                                                                                                                   |
|-----|----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1] | `certificates`       | List of Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/), each of them holding a key/certificate pair to add to the store. |
| [2] | `defaultCertificate` | Name of a Kubernetes [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that holds the default key/certificate pair for the store.       |

??? example "Declaring and referencing a TLSStore"
   
    ```yaml tab="TLSStore"
    apiVersion: traefik.io/v1alpha1
    kind: TLSStore
    metadata:
      name: default

    spec:
      defaultCertificate:
        secretName:  supersecret
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - websecure
      routes:
      - match: Host(`example.com`) && PathPrefix(`/stripit`)
        kind: Rule
        services:
        - name: whoami
          port: 80
      tls: {}
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

### Kind: `ServersTransport`

`ServersTransport` is the CRD implementation of a [ServersTransport](../services/index.md#serverstransport).

!!! important "Default serversTransport"
    If no `serversTransport` is specified, the `default@internal` will be used. 
    The `default@internal` serversTransport is created from the [static configuration](../overview.md#transport-configuration). 

!!! info "ServersTransport Attributes"
   
    ```yaml tab="ServersTransport"
    apiVersion: traefik.io/v1alpha1
    kind: ServersTransport
    metadata:
      name: mytransport
      namespace: default
    
    spec:
      serverName: foobar               # [1]
      insecureSkipVerify: true         # [2]
      rootCAsSecrets:                  # [3]
        - foobar
        - foobar
      certificatesSecrets:             # [4]
        - foobar
        - foobar
      maxIdleConnsPerHost: 1           # [5]
      forwardingTimeouts:              # [6]
        dialTimeout: 42s               # [7]
        responseHeaderTimeout: 42s     # [8]
        idleConnTimeout: 42s           # [9]
      peerCertURI: foobar              # [10]
      disableHTTP2: true               # [11]
    ```

| Ref  | Attribute               | Purpose                                                                                                                                                                 |
|------|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `serverName`            | ServerName used to contact the server.                                                                                                                                  |
| [2]  | `insecureSkipVerify`    | Controls whether the server's certificate chain and host name is verified.                                                                                              |
| [3]  | `rootCAsSecrets`        | Defines the set of root certificate authorities to use when verifying server certificates. The secret must contain a certificate under either a tls.ca or a ca.crt key. |
| [4]  | `certificatesSecrets`   | Certificates to present to the server for mTLS.                                                                                                                         |
| [5]  | `maxIdleConnsPerHost`   | Controls the maximum idle (keep-alive) connections to keep per-host. If zero, `defaultMaxIdleConnsPerHost` is used.                                                     |
| [6]  | `forwardingTimeouts`    | Timeouts for requests forwarded to the servers.                                                                                                                         |
| [7]  | `dialTimeout`           | The amount of time to wait until a connection to a server can be established. If zero, no timeout exists.                                                               |
| [8]  | `responseHeaderTimeout` | The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists.                    |
| [9]  | `idleConnTimeout`       | The maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. If zero, no timeout exists.                                          |
| [10] | `peerCertURI`           | URI used to match against SAN URIs during the server's certificate verification.                                                                                        |
| [11] | `disableHTTP2`          | Disables HTTP/2 for connections with servers.                                                                                                                           |

!!! info "CA Secret"

    The CA secret must contain a base64 encoded certificate under either a `tls.ca` or a `ca.crt` key.

??? example "Declaring and referencing a ServersTransport"
   
    ```yaml tab="ServersTransport"
    apiVersion: traefik.io/v1alpha1
    kind: ServersTransport
    metadata:
      name: mytransport
      namespace: default
    
    spec:
      serverName: example.org
      insecureSkipVerify: true
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: testroute
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`example.com`)
        kind: Rule
        services:
        - name: whoami
          port: 80
          serversTransport: mytransport
    ```

#### ServersTransport reference

By default, the referenced ServersTransport CRD must be defined in the same [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) namespace.

To reference a ServersTransport CRD from another namespace, 
the value must be of form `namespace-name@kubernetescrd`,
and the [cross-namespace](../../../providers/kubernetes-crd/#allowcrossnamespace) option must be enabled.

If the ServersTransport CRD is defined in another provider the cross-provider format `name@provider` should be used.

## Further

Also see the [full example](../../user-guides/crd-acme/index.md) with Let's Encrypt.

{!traefik-for-business-applications.md!}
