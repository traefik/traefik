# Migration Guide: From v1 to v2

How to Migrate from Traefik v1 to Traefik v2.
{: .subtitle }

The version 2 of Traefik introduces a number of breaking changes, 
which require one to update their configuration when they migrate from v1 to v2.
The goal of this page is to recapitulate all of these changes, and in particular to give examples, 
feature by feature, of how the configuration looked like in v1, and how it now looks like in v2.

!!! Note "Migration Helper"
    
    We created a tool to help during the migration: [traefik-migration-tool](https://github.com/containous/traefik-migration-tool)

    This tool allows to:

    - convert `Ingress` to Traefik `IngressRoute` resources.
    - convert `acme.json` file from v1 to v2 format.

## Frontends and Backends Are Dead... <br/>... Long Live Routers, Middlewares, and Services

During the transition from v1 to v2, a number of internal pieces and components of Traefik were rewritten and reorganized.
As such, the combination of core notions such as frontends and backends has been replaced with the combination of routers, services, and middlewares.

Typically, a router replaces a frontend, and a service assumes the role of a backend, with each router referring to a service.
However, even though a backend was in charge of applying any desired modification on the fly to the incoming request,
the router defers that responsibility to another component.
Instead, a dedicated middleware is now defined for each kind of such modification.
Then any router can refer to an instance of the wanted middleware.

!!! example "One frontend with basic auth and one backend, become one router, one service, and one basic auth middleware."

    ### v1
    
    ```yaml tab="Docker"
    labels:
      - "traefik.frontend.rule=Host:test.localhost;PathPrefix:/test"
      - "traefik.frontend.auth.basic.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
    ```

    ```yaml tab="K8s Ingress"
    apiVersion: extensions/v1beta1
    kind: Ingress
    metadata:
      name: traefik
      namespace: kube-system
      annotations:
        kubernetes.io/ingress.class: traefik
        traefik.ingress.kubernetes.io/rule-type: PathPrefix
    spec:
      rules:
      - host: test.locahost
        http:
          paths:
          - path: /test
            backend:
              serviceName: server0
              servicePort: 80
          - path: /test
            backend:
              serviceName: server1
              servicePort: 80
    ```

    ```toml tab="File (TOML)"
    [frontends]
      [frontends.frontend1]
        entryPoints = ["http"]
        backend = "backend1"
    
        [frontends.frontend1.routes]
          [frontends.frontend1.routes.route0]
            rule = "Host:test.localhost"
          [frontends.frontend1.routes.route0]
            rule = "PathPrefix:/test"
        
        [frontends.frontend1.auth]
          [frontends.frontend1.auth.basic]
            users = [
              "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
              "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
            ]
    
    [backends]
      [backends.backend1]
        [backends.backend1.servers.server0]
          url = "http://10.10.10.1:80"
        [backends.backend1.servers.server1]
          url = "http://10.10.10.2:80"
    
        [backends.backend1.loadBalancer]
          method = "wrr"
    ```
    
    ### v2
    
    ```yaml tab="Docker"
    labels:
      - "traefik.http.routers.router0.rule=Host(`bar.com`) && PathPrefix(`/test`)"
      - "traefik.http.routers.router0.middlewares=auth"
      - "traefik.http.middlewares.auth.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
    ```

    ```yaml tab="K8s IngressRoute"
    # The definitions below require the definitions for the Middleware and IngressRoute kinds.  
    # https://docs.traefik.io/v2.0/providers/kubernetes-crd/#traefik-ingressroute-definition
    apiVersion: traefik.containo.us/v1alpha1
    kind: Middleware
    metadata:
      name: basicauth
      namespace: foo
    
    spec:
      basicAuth:
        users:
          - test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/
          - test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0
    
    ---
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - http
      routes:
      - match: Host(`test.localhost`) && PathPrefix(`/test`)
        kind: Rule
        services:
        - name: server0
          port: 80
        - name: server1
          port: 80
        middlewares:
        - name: basicauth
          namespace: foo
    ```

    ```toml tab="File (TOML)"
    [http.routers]
      [http.routers.router0]
        rule = "Host(`test.localhost`) && PathPrefix(`/test`)"
        middlewares = ["auth"]
        service = "my-service"
    
    [http.services]
      [[http.services.my-service.loadBalancer.servers]]
        url = "http://10.10.10.1:80"
      [[http.services.my-service.loadBalancer.servers]]
        url = "http://10.10.10.2:80"
    
    [http.middlewares]
      [http.middlewares.auth.basicAuth]
        users = [
          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", 
          "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
        ]
    ```

    ```yaml tab="File (YAML)"
    http:
      routers:
        router0:
          rule: "Host(`test.localhost`) && PathPrefix(`/test`)"
          service: my-service
          middlewares:
          - auth
    
      services:
        my-service:
          loadBalancer:
            servers:
            - url: http://10.10.10.1:80
            - url: http://10.10.10.2:80
    
      middlewares:
        auth:
          basicAuth:
            users:
            - "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"
            - "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
    ```

## TLS configuration is now dynamic, per router.

TLS parameters used to be specified in the static configuration, as an entryPoint field.
With Traefik v2, a new dynamic TLS section at the root contains all the desired TLS configurations.
Then, a router's TLS field can refer to one of the TLS configurations defined at the root, hence defining the TLS configuration for that router.

!!! example "TLS on web-secure entryPoint becomes TLS option on Router-1"

    ### v1
    
    ```toml tab="File (TOML)"
    # static configuration
    [entryPoints]
      [entryPoints.web-secure]
        address = ":443"
    
        [entryPoints.web-secure.tls]
          minVersion = "VersionTLS12"
          cipherSuites = [
            "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
            "TLS_RSA_WITH_AES_256_GCM_SHA384"
           ]
          [[entryPoints.web-secure.tls.certificates]]
            certFile = "path/to/my.cert"
            keyFile = "path/to/my.key"
    ```
    
    ```bash tab="CLI"
    --entryPoints='Name:web-secure Address::443 TLS:path/to/my.cert,path/to/my.key TLS.MinVersion:VersionTLS12 TLS.CipherSuites:TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA384'
    ```
    
    ### v2
    
    ```toml tab="File (TOML)"
    # dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        rule = "Host(`bar.com`)"
        service = "service-id"
        # will terminate the TLS request
        [http.routers.Router-1.tls]
          options = "myTLSOptions"
    
    [[tls.certificates]]
      certFile = "/path/to/domain.cert"
      keyFile = "/path/to/domain.key"
    
    [tls.options]
      [tls.options.default]
        minVersion = "VersionTLS12"
    
      [tls.options.myTLSOptions]
        minVersion = "VersionTLS13"
        cipherSuites = [
              "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
              "TLS_RSA_WITH_AES_256_GCM_SHA384"
            ]
    ```
    
    ```yaml tab="File (YAML)"
    http:
      routers:
        Router-1:
          rule: "Host(`bar.com`)"
          service: service-id
          # will terminate the TLS request
          tls:
            options: myTLSOptions
    
    tls:
      certificates:
        - certFile: /path/to/domain.cert
          keyFile: /path/to/domain.key
      options:
        myTLSOptions:
          minVersion: VersionTLS13
          cipherSuites:
          - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
          - TLS_RSA_WITH_AES_256_GCM_SHA384
    ```
    
    ```yaml tab="K8s IngressRoute"
    # The definitions below require the definitions for the TLSOption and IngressRoute kinds.  
    # https://docs.traefik.io/v2.0/providers/kubernetes-crd/#traefik-ingressroute-definition
    apiVersion: traefik.containo.us/v1alpha1
    kind: TLSOption
    metadata:
      name: mytlsoption
      namespace: default
    
    spec:
      minVersion: VersionTLS13
      cipherSuites:
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        - TLS_RSA_WITH_AES_256_GCM_SHA384
    
    ---
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: ingressroutebar
    
    spec:
      entryPoints:
        - web
      routes:
      - match: Host(`bar.com`)
        kind: Rule
        services:
        - name: whoami
          port: 80
      tls:
        options: 
          name: mytlsoption
          namespace: default
    ```
    
    ```yaml tab="Docker"
    labels:
      # myTLSOptions must be defined by another provider, in this instance in the File Provider.
      # see the cross provider section
      - "traefik.http.routers.router0.tls.options=myTLSOptions@file"
    ```

## HTTP to HTTPS Redirection is now applied on Router

Previously on Traefik v1, the redirection was applied on an entrypoint or on a FrontEnd.
With Traefik v2 it is applied on a Router. 

To apply a redirection, one of the redirect middlewares, [RedirectRegex](../middlewares/redirectregex.md) or [RedirectScheme](../middlewares/redirectscheme.md), has to be configured and added to the router middlewares list.

!!! example "HTTP to HTTPS redirection thanks to the RedirectScheme Middleware"

    ### v1
    
    ```toml tab="File (TOML)"
    # static configuration
    defaultEntryPoints = ["http", "https"]
    
    [entryPoints]
      [entryPoints.http]
      address = ":80"
        [entryPoints.http.redirect]
        entryPoint = "https"
      [entryPoints.https]
      address = ":443"
        [entryPoints.https.tls]
          [[entryPoints.https.tls.certificates]]
          certFile = "examples/traefik.crt"
          keyFile = "examples/traefik.key"
    ```
    
    ```bash tab="CLI"
    --entrypoints=Name:web Address::80 Redirect.EntryPoint:web-secure
    --entryPoints='Name:web-secure Address::443 TLS:path/to/my.cert,path/to/my.key'
    ```
    
    ### v2
    
    ```yaml tab="Docker"
    labels:
    - traefik.http.routers.web.entrypoints=web
    - traefik.http.routers.web.middlewares=redirect@file
    - traefik.http.routers.web-secured.entrypoints=web-secure
    - traefik.http.routers.web-secured.tls=true
    ```

    ```yaml tab="K8s IngressRoute"
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: http-redirect-ingressRoute
    
    spec:
      entryPoints:
        - web
      routes:
        - match: Host(`foo`)
          kind: Rule
          services:
            - name: whoami
              port: 80
          middlewares:
            - name: redirect
    
    ---
    apiVersion: traefik.containo.us/v1alpha1
    kind: IngressRoute
    metadata:
      name: https-ingressRoute
    
    spec:
      entryPoints:
        - web-secure
      routes:
        - match: Host(`foo`)
          kind: Rule
          services:
            - name: whoami
              port: 80
      tls: {}
      
    ---
    apiVersion: traefik.containo.us/v1alpha1
    kind: Middleware
    metadata:
      name: redirect
    spec:
      redirectScheme:
        scheme: https
      
    ```

    ```toml tab="File (TOML)"
    # static configuration
    [entryPoints.web]
        Address = ":80"
    
    [entryPoints.web-secure]
        Address = ":443"
    
    # dynamic configuration
    [http.routers]
      [http.routers.router0]
        service = "my-service"
        entrypoints = "web"
        middlewares = ["redirect"]
    
    [http.routers.router1]
        service = "my-service"
        entrypoints = "web-secure"
        [http.routers.router1.tls]
        
    [http.services]
      [[http.services.my-service.loadBalancer.servers]]
        url = "http://10.10.10.1:80"
      [[http.services.my-service.loadBalancer.servers]]
        url = "http://10.10.10.2:80"
    
    [http.middlewares]
      [http.middlewares.redirect.redirectScheme]
        scheme = "https"
    
    [[tls.certificates]]
          certFile = "/path/to/domain.cert"
          keyFile = "/path/to/domain.key"    
    ```
    
    ```yaml tab="File (YAML)"
    http:
      routers:
        router0:
            entryPoints:
            - web
            middlewares:
            - redirect
            service: my-service
    
        router1:
            entryPoints:
                - web-secure
            service: my-service
            tls: {}
    
      services:
        my-service:
          loadBalancer:
            servers:
            - url: http://10.10.10.1:80
            - url: http://10.10.10.2:80
    
      middlewares:
        redirect:
          redirectScheme:
            scheme: https
    
    tls:
      certificate:
      - certFile: /app/certs/server/server.pem
        keyFile: /app/certs/server/server.pem
    ``` 

## ACME (let's encrypt)

[ACME](../https/acme.md) is now a certificate resolver (under a certificatesResolvers section) but remains in the static configuration.

!!! example ""

    ### v1
    
    ```toml tab="File (TOML)"
    # static configuration
    defaultEntryPoints = ["web-secure","web"]
    
    [entryPoints.web]
    address = ":80"
      [entryPoints.web.redirect]
      entryPoint = "webs"
    [entryPoints.web-secure]
      address = ":443"
      [entryPoints.https.tls]
    
    [acme]
      email = "your-email-here@my-awesome-app.org"
      storage = "acme.json"
      entryPoint = "web-secure"
      onHostRule = true
      [acme.httpChallenge]
        entryPoint = "web"
    ```
    
    ```bash tab="CLI"
    --defaultentrypoints=web-secure,web
    --entryPoints=Name:web Address::80 Redirect.EntryPoint:web-secure
    --entryPoints=Name:web-secure Address::443 TLS
    --acme.email=your-email-here@my-awesome-app.org
    --acme.storage=acme.json
    --acme.entryPoint=web-secure
    --acme.onHostRule=true
    --acme.httpchallenge.entrypoint=http
    ```
    
    ### v2
    
    ```toml tab="File (TOML)"
    # static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"
    
      [entryPoints.web-secure]
        address = ":443"
    
    [certificatesResolvers.sample.acme]
      email = "your-email@your-domain.org"
      storage = "acme.json"
      [acme.httpChallenge]
        # used during the challenge
        entryPoint = "web"
    ```
    
    ```yaml tab="File (YAML)"
    entryPoints:
      web:
        address: ":80"
    
      web-secure:
        address: ":443"
    
    certificatesResolvers:
      sample:
        acme:
          email: your-email@your-domain.org
          storage: acme.json
          httpChallenge:
            # used during the challenge
            entryPoint: web
    ``` 
    
    ```bash tab="CLI"
    --entryPoints.web.address=":80"
    --entryPoints.websecure.address=":443"
    --certificatesResolvers.sample.acme.email: your-email@your-domain.org
    --certificatesResolvers.sample.acme.storage: acme.json
    --certificatesResolvers.sample.acme.httpChallenge.entryPoint: web
    ```

## Traefik Logs

	TODO

## Tracing

	TODO

## Metrics

	TODO

## No more root level key/values

	TODO

## Providers

Supported providers, for now:

- [ ] Azure Service Fabric
- [ ] BoltDB
- [ ] Consul
- [ ] Consul Catalog
- [x] Docker
- [ ] DynamoDB
- [ ] ECS
- [ ] Etcd
- [ ] Eureka
- [x] File
- [x] Kubernetes Ingress (without annotations)
- [x] Kubernetes IngressRoute
- [x] Marathon
- [ ] Mesos
- [x] Rest
- [ ] Zookeeper
