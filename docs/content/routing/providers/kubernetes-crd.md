# Traefik & Kubernetes

The Kubernetes Ingress Controller, The Custom Resource Way.
{: .subtitle }

## Configuration Examples

??? example "Configuring KubernetesCRD and Deploying/Exposing Services"

    ```yaml tab="Resource Definition"
    # All resources definition must be declared
    --8<-- "content/reference/dynamic-configuration/kubernetes-crd-definition.yml"
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
              image: traefik:v2.1
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
    apiVersion: traefik.containo.us/v1alpha1
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
    apiVersion: traefik.containo.us/v1alpha1
      kind: IngressRouteTCP
      metadata:
        name: ingressroute.tcp
      namespace: default
      
      spec:
        entryPoints:
          - tcpep
        routes:
          - match: HostSNI(`bar`)
            kind: Rule
            services:
              - name: whoamitcp
                port: 8080
        
    ---
    apiVersion: traefik.containo.us/v1alpha1
       kind: IngressRouteUDP
       metadata:
         name: ingressroute.udp
         namespace: default
         
       spec:
         entryPoints:
           - fooudp
         routes:
           - kind: Rule
             services:
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
        app: containous
        name: whoami
    
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: containous
          task: whoami
      template:
        metadata:
          labels:
            app: containous
            task: whoami
        spec:
          containers:
            - name: containouswhoami
              image: containous/whoami
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
        app: containous
        task: whoami
    
    ---
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: whoamitcp
      namespace: default
      labels:
        app: containous
        name: whoamitcp
    
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: containous
          task: whoamitcp
      template:
        metadata:
          labels:
            app: containous
            task: whoamitcp
        spec:
          containers:
            - name: containouswhoamitcp
              image: containous/whoamitcp
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
        app: containous
        task: whoamitcp
    
    ---
    kind: Deployment
    apiVersion: apps/v1
    metadata:
      name: whoamiudp
      namespace: default
      labels:
        app: containous
        name: whoamiudp
    
    spec:
      replicas: 2
      selector:
        matchLabels:
          app: containous
          task: whoamiudp
      template:
        metadata:
          labels:
            app: containous
            task: whoamiudp
        spec:
          containers:
            - name: containouswhoamiudp
              image: containous/whoamiudp:dev
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
        app: containous
        task: whoamiudp
    ```

## Routing Configuration

### Custom Resource Definition (CRD)

* You can find an exhaustive list, generated from Traefik's source code, of the custom resources and their attributes in [the reference page](../../reference/dynamic-configuration/kubernetes-crd.md).
* Validate that [the prerequisites](../../providers/kubernetes-crd.md#configuration-requirements) are fulfilled before using the Traefik custom resources.
* Traefik CRDs are building blocks that you can assemble according to your needs.
    
You can find an excerpt of the available custom resources in the table below:

| Kind                                     | Purpose                                                       | Concept Behind                                                 |
|------------------------------------------|---------------------------------------------------------------|----------------------------------------------------------------|
| [IngressRoute](#kind-ingressroute)       | HTTP Routing                                                  | [HTTP router](../routers/index.md#configuring-http-routers)    |
| [Middleware](#kind-middleware)           | Tweaks the HTTP requests before they are sent to your service | [HTTP Middlewares](../../middlewares/overview.md)              |
| [TraefikService](#kind-traefikservice)   | Abstraction for HTTP loadbalancing/mirroring                  | [HTTP service](../services/index.md#configuring-http-services) |
| [IngressRouteTCP](#kind-ingressroutetcp) | TCP Routing                                                   | [TCP router](../routers/index.md#configuring-tcp-routers)      |
| [IngressRouteUDP](#kind-ingressrouteudp) | UDP Routing                                                   | [UDP router](../routers/index.md#configuring-udp-routers)      |
| [TLSOptions](#kind-tlsoption)            | Allows to configure some parameters of the TLS connection     | [TLSOptions](../../https/tls.md#tls-options)                   |
| [TLSStores](#kind-tlsstore)              | Allows to configure the default TLS store                     | [TLSStores](../../https/tls.md#certificates-stores)            |

### Kind: `IngressRoute`

`IngressRoute` is the CRD implementation of a [Traefik HTTP router](../routers/index.md#configuring-http-routers).

Register the `IngressRoute` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `IngressRoute` objects.

!!! info "IngressRoute Attributes"

    ```yaml
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: foo
      namespace: bar
    spec:
      entryPoints:                      # [1]
        - foo
      routes:                           # [2]
      - kind: Rule
        match: Host(`test.domain.com`)  # [3]
        priority: 10                    # [4]
        middlewares:                    # [5]
        - name: middleware1             # [6]
          namespace: default            # [7]
        services:                       # [8]
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
      tls:                              # [9]
        secretName: supersecret         # [10]
        options:                        # [11]
          name: opt                     # [12]
          namespace: default            # [13]
        certResolver: foo               # [14]
        domains:                        # [15]
        - main: foo.com                 # [16]
          sans:                         # [17]
          - a.foo.com
          - b.foo.com
    ```

| Ref  | Attribute                  | Purpose                                                                                                                                                                     |
|------|----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `entryPoints`              | List of [entry points](../routers/index.md#entrypoints) names                                                                                                               |
| [2]  | `routes`                   | List of routes                                                                                                                                                              |
| [3]  | `routes[n].match`          | Defines the [rule](../routers/index.md#rule) corresponding to an underlying router.                                                                                         |
| [4]  | `routes[n].priority`       | [Disambiguate](../routers/index.md#priority) rules of the same length, for route matching                                                                                   |
| [5]  | `routes[n].middlewares`    | List of reference to [Middleware](#kind-middleware)                                                                                                                         |
| [6]  | `middlewares[n].name`      | Defines the [Middleware](#kind-middleware) name                                                                                                                             |
| [7]  | `middlewares[n].namespace` | Defines the [Middleware](#kind-middleware) namespace                                                                                                                        |
| [8]  | `routes[n].services`       | List of any combination of [TraefikService](#kind-traefikservice) and reference to a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) |
| [9]  | `tls`                      | Defines [TLS](../routers/index.md#tls) certificate configuration                                                                                                            |
| [10] | `tls.secretName`           | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace)                        |
| [11] | `tls.options`              | Defines the reference to a [TLSOption](#kind-tlsoption)                                                                                                                     |
| [12] | `options.name`             | Defines the [TLSOption](#kind-tlsoption) name                                                                                                                               |
| [13] | `options.namespace`        | Defines the [TLSOption](#kind-tlsoption) namespace                                                                                                                          |
| [14] | `tls.certResolver`          | Defines the reference to a [CertResolver](../routers/index.md#certresolver)                                                                                                 |
| [15] | `tls.domains`              | List of [domains](../routers/index.md#domains)                                                                                                                              |
| [16] | `domains[n].main`          | Defines the main domain name                                                                                                                                                |
| [17] | `domains[n].sans`          | List of SANs (alternative domains)                                                                                                                                          |

??? example "Declaring an IngressRoute"

    ```yaml tab="IngressRoute"
    # All resources definition must be declared
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: testName
      namespace: default
    spec:
      entryPoints:
        - web
      routes:
      - kind: Rule
        match: Host(`test.domain.com`)
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
        - main: foo.com
          sans:
          - a.foo.com
          - b.foo.com
        options:
          name: opt
          namespace: default
        secretName: supersecret
    ```

    ```yaml tab="Middlewares"
    # All resources definition must be declared
    # Prefixing with /foo
    apiVersion: traefik.containo.us/v1alpha1
    kind: Middleware
    metadata:
      name: middleware1
      namespace: default
    spec:
      addPrefix:
        prefix: /foo
    ```

    ```yaml tab="TLSOption"
    apiVersion: traefik.containo.us/v1alpha1
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

### Kind: `Middleware`

`Middleware` is the CRD implementation of a [Traefik middleware](../../middlewares/overview.md).

Register the `Middleware` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `Middleware` objects or referencing middlewares in the [`IngressRoute`](#kind-ingressroute) objects.

??? "Declaring and Referencing a Middleware"
    
    ```yaml tab="Middleware"
    apiVersion: traefik.containo.us/v1alpha1
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
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`) && PathPrefix(`/stripit`)
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
	(in the reference to the middleware) with the [provider namespace](../../middlewares/overview.md#provider-namespace),
	when the definition of the middleware comes from another provider.
	In this context, specifying a namespace when referring to the resource does not make any sense, and will be ignored.

More information about available middlewares in the dedicated [middlewares section](../../middlewares/overview.md).

### Kind: `TraefikService`

`TraefikService` is the CRD implementation of a ["Traefik Service"](../services/index.md).

Register the `TraefikService` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `TraefikService` objects,
referencing services in the [`IngressRoute`](#kind-ingressroute) objects, or recursively in others `TraefikService` objects.

!!! info "Disambiguate Traefik and Kubernetes Services "

    As the field `name` can reference different types of objects, use the field `kind` to avoid any ambiguity.
    
    The field `kind` allows the following values:
    
    * `Service` (default value): to reference a [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/)
    * `TraefikService`: to reference another [Traefik Service](../services/index.md)

`TraefikService` object allows to use any (valid) combinations of:

* servers [load balancing](#server-load-balancing).  
* services [Weighted Round Robin](#weighted-round-robin) load balancing.
* services [mirroring](#mirroring).


#### Server Load Balancing

More information in the dedicated server [load balancing](../services/index.md#load-balancing) section.

??? "Declaring and Using Server Load Balancing"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`) && PathPrefix(`/foo`)
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
        app: containous
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
        app: containous
        task: app2
    ```

#### Weighted Round Robin

More information in the dedicated [Weighted Round Robin](../services/index.md#weighted-round-robin-service) service load balancing section.

??? "Declaring and Using Weighted Round Robin"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`) && PathPrefix(`/foo`)
        kind: Rule
        services:
        - name: wrr1
          namespace: default
          kind: TraefikService
    ```
    
    ```yaml tab="Weighted Round Robin"
    apiVersion: traefik.containo.us/v1alpha1
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
    apiVersion: traefik.containo.us/v1alpha1
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
        app: containous
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
        app: containous
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
        app: containous
        task: app3
    ```

#### Mirroring

More information in the dedicated [mirroring](../services/index.md#mirroring-service) service section.

??? "Declaring and Using Mirroring"

    ```yaml tab="IngressRoute"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
      namespace: default
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`) && PathPrefix(`/foo`)
        kind: Rule
        services:
        - name: mirror1
          namespace: default
          kind: TraefikService
    ```
    
    ```yaml tab="Mirroring k8s Service"
    # Mirroring from a k8s Service
    apiVersion: traefik.containo.us/v1alpha1
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
    apiVersion: traefik.containo.us/v1alpha1
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
        app: containous
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
        app: containous
        task: app2
    ```

!!! important "References and namespaces"

    If the optional `namespace` attribute is not set, the configuration will be applied with the namespace of the current resource.
    
    Additionally, when the definition of the `TraefikService` is from another provider,
    the cross-provider syntax (`service@provider`) should be used to refer to the `TraefikService`, just as in the middleware case.
    
    Specifying a namespace attribute in this case would not make any sense, and will be ignored (except if the provider is `kubernetescrd`).

### Kind `IngressRouteTCP`

`IngressRouteTCP` is the CRD implementation of a [Traefik TCP router](../routers/index.md#configuring-tcp-routers).

Register the `IngressRouteTCP` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `IngressRouteTCP` objects.

!!! info "IngressRouteTCP Attributes"

    ```yaml
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRouteTCP
    metadata:
      name: ingressroutetcpfoo
    
    spec:
      entryPoints:                  # [1]
        - footcp
      routes:                       # [2]
      - match: HostSNI(`*`)         # [3]
        services:                   # [4]
        - name: foo                 # [5]
          port: 8080                # [6]
          weight: 10                # [7]
          TerminationDelay: 400     # [8]
      tls:                          # [9]
        secretName: supersecret     # [10]
        options:                    # [11]
          name: opt                 # [12]
          namespace: default        # [13]
        certResolver: foo           # [14]
        domains:                    # [15]
        - main: foo.com             # [16]
          sans:                     # [17]
          - a.foo.com
          - b.foo.com
        passthrough: false          # [18]
    ```

| Ref  | Attribute                      | Purpose                                                                                                                                                                                                                                                                                                                                                                                  |
|------|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `entryPoints`                  | List of [entrypoints](../routers/index.md#entrypoints_1) names                                                                                                                                                                                                                                                                                                                           |
| [2]  | `routes`                       | List of routes                                                                                                                                                                                                                                                                                                                                                                           |
| [3]  | `routes[n].match`              | Defines the [rule](../routers/index.md#rule_1) corresponding to an underlying router                                                                                                                                                                                                                                                                                                     |
| [4]  | `routes[n].services`           | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions                                                                                                                                                                                                                                                                               |
| [5]  | `services[n].name`             | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)                                                                                                                                                                                                                                                                             |
| [6]  | `services[n].port`             | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)                                                                                                                                                                                                                                                                             |
| [7]  | `services[n].weight`           | Defines the weight to apply to the server load balancing                                                                                                                                                                                                                                                                                                                                 |
| [8]  | `services[n].TerminationDelay` | corresponds to the deadline that the proxy sets, after one of its connected peers indicates it has closed the writing capability of its connection, to close the reading capability as well, hence fully terminating the connection.<br/>It is a duration in milliseconds, defaulting to 100. A negative value means an infinite deadline (i.e. the reading capability is never closed). |
| [9]  | `tls`                          | Defines [TLS](../routers/index.md#tls_1) certificate configuration                                                                                                                                                                                                                                                                                                                       |
| [10] | `tls.secretName`               | Defines the [secret](https://kubernetes.io/docs/concepts/configuration/secret/) name used to store the certificate (in the `IngressRoute` namespace)                                                                                                                                                                                                                                     |
| [11] | `tls.options`                  | Defines the reference to a [TLSOption](#kind-tlsoption)                                                                                                                                                                                                                                                                                                                                  |
| [12] | `options.name`                 | Defines the [TLSOption](#kind-tlsoption) name                                                                                                                                                                                                                                                                                                                                            |
| [13] | `options.namespace`            | Defines the [TLSOption](#kind-tlsoption) namespace                                                                                                                                                                                                                                                                                                                                       |
| [14] | `tls.certResolver`             | Defines the reference to a [CertResolver](../routers/index.md#certresolver_1)                                                                                                                                                                                                                                                                                                            |
| [15] | `tls.domains`                  | List of [domains](../routers/index.md#domains_1)                                                                                                                                                                                                                                                                                                                                         |
| [16] | `domains[n].main`              | Defines the main domain name                                                                                                                                                                                                                                                                                                                                                             |
| [17] | `domains[n].sans`              | List of SANs (alternative domains)                                                                                                                                                                                                                                                                                                                                                       |
| [18] | `tls.passthrough`              | If `true`, delegates the TLS termination to the backend                                                                                                                                                                                                                                                                                                                                  |

??? example "Declaring an IngressRouteTCP"

    ```yaml tab="IngressRouteTCP"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRouteTCP
    metadata:
      name: ingressroutetcpfoo
    
    spec:
      entryPoints:
        - footcp
      routes:
      # Match is the rule corresponding to an underlying router.
      - match: HostSNI(`*`)
        services:
        - name: foo
          port: 8080
          TerminationDelay: 400
          weight: 10
        - name: bar
          port: 8081
          TerminationDelay: 500
          weight: 10
      tls:
        certResolver: foo
        domains:
        - main: foo.com
          sans:
          - a.foo.com
          - b.foo.com
        options:
          name: opt
          namespace: default
        secretName: supersecret
        passthrough: false
    ```
    
    ```yaml tab="TLSOption"
    apiVersion: traefik.containo.us/v1alpha1
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

### Kind `IngressRouteUDP`

`IngressRouteUDP` is the CRD implementation of a [Traefik UDP router](../routers/index.md#configuring-udp-routers).

Register the `IngressRouteUDP` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `IngressRouteUDP` objects.

!!! info "IngressRouteUDP Attributes"

    ```yaml
    apiVersion: traefik.containo.us/v1alpha1
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
    ```

| Ref  | Attribute                      | Purpose                                                                                                                                                                                                                                                                                                                                                                                  |
|------|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1]  | `entryPoints`                  | List of [entrypoints](../routers/index.md#entrypoints_1) names                                                                                                                                                                                                                                                                                                                           |
| [2]  | `routes`                       | List of routes                                                                                                                                                                                                                                                                                                                                                                           |
| [3]  | `routes[n].services`           | List of [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) definitions                                                                                                                                                                                                                                                                               |
| [4]  | `services[n].name`             | Defines the name of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)                                                                                                                                                                                                                                                                             |
| [6]  | `services[n].port`             | Defines the port of a [Kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/)                                                                                                                                                                                                                                                                             |
| [7]  | `services[n].weight`           | Defines the weight to apply to the server load balancing                                                                                                                                                                                                                                                                                                                                 |

??? example "Declaring an IngressRouteUDP"

    ```yaml
    apiVersion: traefik.containo.us/v1alpha1
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
    
### Kind: `TLSOption`

`TLSOption` is the CRD implementation of a [Traefik "TLS Option"](../../https/tls.md#tls-options).

Register the `TLSOption` [kind](../../reference/dynamic-configuration/kubernetes-crd.md#definitions) in the Kubernetes cluster before creating `TLSOption` objects
or referencing TLS options in the [`IngressRoute`](#kind-ingressroute) / [`IngressRouteTCP`](#kind-ingressroutetcp) objects.

!!! info "TLSOption Attributes"
   
    ```yaml tab="TLSOption"
    apiVersion: traefik.containo.us/v1alpha1
    kind: TLSOption
    metadata:
      name: mytlsoption
      namespace: default
    
    spec:
      minVersion: VersionTLS12                      # [1]
      maxVersion: VersionTLS13                      # [1]
      curvePreferences:                             # [3]
        - CurveP521
        - CurveP384
      cipherSuites:                                 # [4]
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        - TLS_RSA_WITH_AES_256_GCM_SHA384
      clientAuth:                                   # [5]
        secretNames:                                # [6]
          - secretCA1
          - secretCA2
        clientAuthType: VerifyClientCertIfGiven     # [7]
      sniStrict: true                               # [8]
    ```

| Ref | Attribute                   | Purpose                                                                                                                                                                    |
|-----|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1] | `minVersion`                | Defines the [minimum TLS version](../../https/tls.md#minimum-tls-version) that is acceptable                                                                               |
| [2] | `maxVersion`                | Defines the [maximum TLS version](../../https/tls.md#maximum-tls-version) that is acceptable                                                                               |
| [3] | `cipherSuites`              | list of supported [cipher suites](../../https/tls.md#cipher-suites) for TLS versions up to TLS 1.2                                                                         |
| [4] | `curvePreferences`          | List of the [elliptic curves references](../../https/tls.md#curve-preferences) that will be used in an ECDHE handshake, in preference order                                |
| [5] | `clientAuth`                | determines the server's policy for TLS [Client Authentication](../../https/tls.md#client-authentication-mtls)                                                              |
| [6] | `clientAuth.secretNames`    | list of names of the referenced Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/) (in TLSOption namespace)                                   |
| [7] | `clientAuth.clientAuthType` | defines the client authentication type to apply. The available values are: `NoClientCert`, `RequestClientCert`, `VerifyClientCertIfGiven` and `RequireAndVerifyClientCert` |
| [8] | `sniStrict`                 | if `true`, Traefik won't allow connections from clients connections that do not specify a server_name extension                                                            |

??? example "Declaring and referencing a TLSOption"
   
    ```yaml tab="TLSOption"
    apiVersion: traefik.containo.us/v1alpha1
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
          - secretCA1
          - secretCA2
        clientAuthType: VerifyClientCertIfGiven
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`) && PathPrefix(`/stripit`)
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
      name: secretCA1
      namespace: default
    
    data:
      tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
    
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: secretCA2
      namespace: default
    
    data:
      tls.ca: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=
    ```
        
!!! important "References and namespaces"

    If the optional `namespace` attribute is not set, the configuration will be applied with the namespace of the IngressRoute.

	Additionally, when the definition of the TLS option is from another provider,
	the cross-provider syntax (`middlewarename@provider`) should be used to refer to the TLS option,
	just as in the [middleware case](../../middlewares/overview.md#provider-namespace).
	Specifying a namespace attribute in this case would not make any sense, and will be ignored.

### Kind: `TLSStore`

`TLSStore` is the CRD implementation of a [Traefik "TLS Store"](../../https/tls.md#certificates-stores).

Register the `TLSStore` kind in the Kubernetes cluster before creating `TLSStore` objects
or referencing TLS stores in the [`IngressRoute`](#kind-ingressroute) / [`IngressRouteTCP`](#kind-ingressroutetcp) objects.

!!! important "Default TLS Store"

    Traefik currently only uses the [TLS Store named "default"](../../https/tls.md#certificates-stores).
    This means that if you have two stores that are named default in different kubernetes namespaces,
    they may be randomly chosen.
    For the time being, please only configure one TLSSTore named default.

!!! info "TLSStore Attributes"
   
    ```yaml tab="TLSStore"
    apiVersion: traefik.containo.us/v1alpha1
    kind: TLSStore
    metadata:
      name: default
      namespace: default
    
    spec:
      defaultCertificate:
        secretName: mySecret                      # [1]
    ```

| Ref | Attribute                   | Purpose                                                                                                                                                                    |
|-----|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [1] | `secretName`                | The name of the referenced Kubernetes [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that holds the default certificate for the store.                                                                             |

??? example "Declaring and referencing a TLSStore"
   
    ```yaml tab="TLSStore"
    apiVersion: traefik.containo.us/v1alpha1
    kind: TLSStore
    metadata:
      name: default
      namespace: default
    
    spec:
      defaultCertificate:
        secretName:  supersecret
    ```
    
    ```yaml tab="IngressRoute"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`) && PathPrefix(`/stripit`)
        kind: Rule
        services:
        - name: whoami
          port: 80
      tls:
        store: 
          name: default
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
## Further

Also see the [full example](../../user-guides/crd-acme/index.md) with Let's Encrypt.
