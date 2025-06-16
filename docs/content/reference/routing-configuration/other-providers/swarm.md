---
title: "Traefik Docker Swarm Routing Documentation"
description: "This guide will teach you how to attach labels to your containers, to route traffic and load balance with Traefik and Docker Swarm."
---

# Traefik & Docker Swarm

One of the best feature of Traefik is to delegate the routing configuration to the application level.
With Docker Swarm, Traefik can leverage labels attached to a service to generate routing rules.

!!! warning "Labels & sensitive data"

    We recommend to *not* use labels to store sensitive data (certificates, credentials, etc).
    Instead, we recommend to store sensitive data in a safer storage (secrets, file, etc).

## Configuration Examples

??? example "Configuring Docker Swarm & Deploying / Exposing one Service"

    Enabling the docker provider (Swarm Mode)

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        # swarm classic (1.12-)
        # endpoint: "tcp://127.0.0.1:2375"
        # docker swarm mode (1.12+)
        endpoint: "tcp://127.0.0.1:2377"
    ```

    ```toml tab="File (TOML)"
    [providers.swarm]
      # swarm classic (1.12-)
      # endpoint = "tcp://127.0.0.1:2375"
      # docker swarm mode (1.12+)
      endpoint = "tcp://127.0.0.1:2377"
    ```

    ```bash tab="CLI"
    # swarm classic (1.12-)
    # --providers.swarm.endpoint=tcp://127.0.0.1:2375
    # docker swarm mode (1.12+)
    --providers.swarm.endpoint=tcp://127.0.0.1:2377
    ```

    Attach labels to services (not containers) while in Swarm mode (in your Docker compose file).
    When there is only one service, and the router does not specify a service,
    then that service is automatically assigned to the router.

    ```yaml
    services:
      my-container:
        deploy:
          labels:
            - traefik.http.routers.my-container.rule=Host(`example.com`)
            - traefik.http.services.my-container-service.loadbalancer.server.port=8080
    ```

    !!! important "Labels in Docker Swarm Mode"
        While in Swarm Mode, Traefik uses labels found on services, not on individual containers.
        Therefore, if you use a compose file with Swarm Mode, labels should be defined in the `deploy` part of your service.
        This behavior is only enabled for docker-compose version 3+ ([Compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/#labels-1)).

??? example "Specify a Custom Port for the Container"

    Forward requests for `http://example.com` to `http://<private IP of container>:12345`:

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - traefik.http.routers.my-container.rule=Host(`example.com`)
            - traefik.http.routers.my-container.service=my-service"
            # Tell Traefik to use the port 12345 to connect to `my-container`
            - traefik.http.services.my-service.loadbalancer.server.port=12345
    ```

    !!! important "Traefik Connecting to the Wrong Port: `HTTP/502 Gateway Error`"
        By default, Traefik uses the lowest exposed port of a container as detailed in
        [Port Detection](../../install-configuration/providers/swarm.md#port-detection) of the Swarm provider.

        Setting the label `traefik.http.services.xxx.loadbalancer.server.port`
        overrides this behavior.

??? example "Specifying more than one router and service per container"

    Forwarding requests to more than one port on a container requires referencing the service loadbalancer port definition using the service parameter on the router.

    In this example, requests are forwarded for `http://example-a.com` to `http://<private IP of container>:8000` in addition to `http://example-b.com` forwarding to `http://<private IP of container>:9000`:

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - traefik.http.routers.www-router.rule=Host(`example-a.com`)
            - traefik.http.routers.www-router.service=www-service
            - traefik.http.services.www-service.loadbalancer.server.port=8000
            - traefik.http.routers.admin-router.rule=Host(`example-b.com`)
            - traefik.http.routers.admin-router.service=admin-service
            - traefik.http.services.admin-service.loadbalancer.server.port=9000
    ```

## Routing Configuration

!!! info "Labels"

     Labels are case-insensitive.

!!! tip "TLS Default Generated Certificates"

    To learn how to configure Traefik default generated certificate, refer to the [TLS Certificates](../http/tls/tls-certificates.md#acme-default-certificate) page.

### General

Traefik creates, for each container, a corresponding [service](../http/load-balancing/service.md) and [router](../http/router/rules-and-priority.md).

The Service automatically gets a server per instance of the container,
and the router automatically gets a rule defined by `defaultRule` (if no rule for it was defined in labels).

#### Service definition

--8<-- "content/routing/providers/service-by-label.md"

??? example "Automatic service assignment with labels"

    With labels in a compose file

    ```yaml
    labels:
      - "traefik.http.routers.myproxy.rule=Host(`example.net`)"
      # service myservice gets automatically assigned to router myproxy
      - "traefik.http.services.myservice.loadbalancer.server.port=8080"
    ```

??? example "Automatic service creation and assignment with labels"

    With labels in a compose file

    ```yaml
    labels:
      # no service specified or defined and yet one gets automatically created
      # and assigned to router myproxy.
      - "traefik.http.routers.myproxy.rule=Host(`example.net`)"
    ```

### Routers

To update the configuration of the Router automatically attached to the container,
add labels starting with `traefik.http.routers.<name-of-your-choice>.` and followed by the option you want to change.

For example, to change the rule, you could add the label ```traefik.http.routers.my-container.rule=Host(`example.com`)```.

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

??? info "`traefik.http.routers.<router_name>.rule`"

    See [rule](../http/router/rules-and-priority.md) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.rule=Host(`example.com`)"
    ```

??? info "`traefik.http.routers.<router_name>.ruleSyntax`"

    !!! warning

        RuleSyntax option is deprecated and will be removed in the next major version.
        Please do not use this field and rewrite the router rules to use the v3 syntax.

    See [ruleSyntax](../http/router/rules-and-priority.md#rulesyntax) for more information.
    
    ```yaml
    traefik.http.routers.myrouter.ruleSyntax=v3
    ```

??? info "`traefik.http.routers.<router_name>.entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.entrypoints=ep1,ep2"
    ```

??? info "`traefik.http.routers.<router_name>.middlewares`"

    See [middlewares overview](../http/middlewares/overview.md) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.middlewares=auth,prefix,cb"
    ```

??? info "`traefik.http.routers.<router_name>.service`"

    See [service](../http/load-balancing/service.md) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.service=myservice"
    ```

??? info "`traefik.http.routers.<router_name>.tls`"

    See [tls](../http/tls/overview.md) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.tls=true"
    ```

??? info "`traefik.http.routers.<router_name>.tls.certresolver`"

    See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.tls.certresolver=myresolver"
    ```

??? info "`traefik.http.routers.<router_name>.tls.domains[n].main`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.tls.domains[0].main=example.org"
    ```

??? info "`traefik.http.routers.<router_name>.tls.domains[n].sans`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.tls.domains[0].sans=test.example.org,dev.example.org"
    ```

??? info "`traefik.http.routers.<router_name>.tls.options`"

    ```yaml
    - "traefik.http.routers.myrouter.tls.options=foobar"
    ```

??? info "`traefik.http.routers.<router_name>.observability.accesslogs`"
    
    The accessLogs option controls whether the router will produce access-logs.
    
    ```yaml
     "traefik.http.routers.myrouter.observability.accesslogs=true"
    ```

??? info "`traefik.http.routers.<router_name>.observability.metrics`"
    
    The metrics option controls whether the router will produce metrics.

    ```yaml
     "traefik.http.routers.myrouter.observability.metrics=true"
    ```

??? info "`traefik.http.routers.<router_name>.observability.tracing`"
    
    The tracing option controls whether the router will produce traces.

    ```yaml
     "traefik.http.routers.myrouter.observability.tracing=true"
    ```
    
??? info "`traefik.http.routers.<router_name>.priority`"

    See [priority](../http/router/rules-and-priority.md#priority-calculation) for more information.

    ```yaml
    - "traefik.http.routers.myrouter.priority=42"
    ```

### Services

To update the configuration of the Service automatically attached to the container,
add labels starting with `traefik.http.services.<name-of-your-choice>.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the label `traefik.http.services.<name-of-your-choice>.loadbalancer.passhostheader=false`.

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

??? info "`traefik.http.services.<service_name>.loadbalancer.server.port`"

    Registers a port.
    Useful when the container exposes multiples ports.

    Mandatory for Docker Swarm (see the section ["Port Detection with Docker Swarm"](../../install-configuration/providers/swarm.md#port-detection)).
    {: #port }

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.server.port=8080"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.server.scheme`"

    Overrides the default scheme.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.server.scheme=http"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.server.weight`"

    Overrides the default weight.
    
    ```yaml
    traefik.http.services.myservice.loadbalancer.server.weight=42
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.serverstransport`"

    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../http/load-balancing/serverstransport.md) for more information.

    ```yaml
    - "traefik.http.services.<service_name>.loadbalancer.serverstransport=foobar@file"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.passhostheader`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.passhostheader=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.headers.X-Foo=foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.hostname`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.hostname=example.org"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.interval`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.interval=10s"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.unhealthyinterval`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.unhealthyinterval=10s"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.path`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.path=/foo"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.method`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.method=foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.status`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.status=42"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.port`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.port=42"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.scheme`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.scheme=http"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.timeout`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.timeout=10s"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.followredirects`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.followredirects=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.httponly=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.name`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.name=foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.path`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.path=/foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.secure`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.secure=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.samesite`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.samesite=none"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`"

    See [response forwarding](../http/load-balancing/service.md#configuration-options) for more information.

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.responseforwarding.flushinterval=10"
    ```

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.<name-of-your-choice>.`,
followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../http/middlewares/redirectscheme.md) named `my-redirect`,
you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme=https`.

More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md).

!!! warning "The character `@` is not authorized in the middleware name."

??? example "Declaring and Referencing a Middleware"

    ```yaml
    services:
      my-container:
        # ...
        deploy:
        labels:
          # Declaring a middleware
          - traefik.http.middlewares.my-redirect.redirectscheme.scheme=https
          # Referencing a middleware
          - traefik.http.routers.my-container.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### TCP

You can declare TCP Routers and/or Services using labels.

??? example "Declaring TCP Routers and Services"

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - "traefik.tcp.routers.my-router.rule=HostSNI(`example.com`)"
            - "traefik.tcp.routers.my-router.tls=true"
            - "traefik.tcp.services.my-service.loadbalancer.server.port=4123"
    ```

!!! warning "TCP and HTTP"

    If you declare a TCP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no TCP Router/Service is defined).
    You can declare both a TCP Router/Service and an HTTP Router/Service for the same container (but you have to do so manually).

#### TCP Routers

??? info "`traefik.tcp.routers.<router_name>.entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.entrypoints=ep1,ep2"
    ```

??? info "`traefik.tcp.routers.<router_name>.rule`"

    See [rule](../tcp/router/rules-and-priority.md#rules) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.rule=HostSNI(`example.com`)"
    ```

??? info "`traefik.tcp.routers.<router_name>.ruleSyntax`"

    !!! warning

        RuleSyntax option is deprecated and will be removed in the next major version.
        Please do not use this field and rewrite the router rules to use the v3 syntax.

    configure the rule syntax to be used for parsing the rule on a per-router basis.
    
    ```yaml
    traefik.tcp.routers.mytcprouter.ruleSyntax=v3
    ```
    
??? info "`traefik.tcp.routers.<router_name>.service`"

    See [service](../tcp/service.md) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.service=myservice"
    ```

??? info "`traefik.tcp.routers.<router_name>.tls`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.tls=true"
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.certresolver`"

    See [certResolver](../tcp/tls.md#configuration-options) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.tls.certresolver=myresolver"
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.domains[n].main`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.tls.domains[0].main=example.org"
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.domains[n].sans`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.tls.domains[0].sans=test.example.org,dev.example.org"
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.options`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.tls.options=mysoptions"
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.passthrough`"

    See [Passthrough](../tcp/tls.md#passthrough) for more information.

    ```yaml
    - "traefik.tcp.routers.mytcprouter.tls.passthrough=true"
    ```

??? info "`traefik.tcp.routers.<router_name>.priority`"

    See [priority](../tcp/router/rules-and-priority.md) for more information.

    ```yaml
    - "traefik.tcp.routers.myrouter.priority=42"
    ```

#### TCP Services

??? info "`traefik.tcp.services.<service_name>.loadbalancer.server.port`"

    Registers a port of the application.

    ```yaml
    - "traefik.tcp.services.mytcpservice.loadbalancer.server.port=423"
    ```

??? info "`traefik.tcp.services.<service_name>.loadbalancer.server.tls`"

    Determines whether to use TLS when dialing with the backend.

    ```yaml
    - "traefik.tcp.services.mytcpservice.loadbalancer.server.tls=true"
    ```

??? info "`traefik.tcp.services.<service_name>.loadbalancer.proxyprotocol.version`"

    See [PROXY protocol](../tcp/service.md#proxy-protocol) for more information.

    ```yaml
    - "traefik.tcp.services.mytcpservice.loadbalancer.proxyprotocol.version=1"
    ```

??? info "`traefik.tcp.services.<service_name>.loadbalancer.serverstransport`"

    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../tcp/serverstransport.md) for more information.

    ```yaml
    - "traefik.tcp.services.<service_name>.loadbalancer.serverstransport=foobar@file"
    ```

#### TCP Middleware

You can declare pieces of middleware using tags starting with `traefik.tcp.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`InFlightConn`](../tcp/middlewares/inflightconn.md) named `test-inflightconn`, you'd write `traefik.tcp.middlewares.test-inflightconn.inflightconn.amount=10`.

More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md).

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    # Declaring a middleware
    traefik.tcp.middlewares.test-inflightconn.amount=10
    # Referencing a middleware
    traefik.tcp.routers.my-service.middlewares=test-inflightconn
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### UDP

You can declare UDP Routers and/or Services using labels.

??? example "Declaring UDP Routers and Services"

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - "traefik.udp.routers.my-router.entrypoints=udp"
            - "traefik.udp.services.my-service.loadbalancer.server.port=4123"
    ```

!!! warning "UDP and HTTP"

    If you declare a UDP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no UDP Router/Service is defined).
    You can declare both a UDP Router/Service and an HTTP Router/Service for the same container (but you have to do so manually).

#### UDP Routers

??? info "`traefik.udp.routers.<router_name>.entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    ```yaml
    - "traefik.udp.routers.myudprouter.entrypoints=ep1,ep2"
    ```

??? info "`traefik.udp.routers.<router_name>.service`"

    See [service](../udp/service.md) for more information.

    ```yaml
    - "traefik.udp.routers.myudprouter.service=myservice"
    ```

#### UDP Services

??? info "`traefik.udp.services.<service_name>.loadbalancer.server.port`"

    Registers a port of the application.

    ```yaml
    - "traefik.udp.services.myudpservice.loadbalancer.server.port=423"
    ```

### Specific Provider Options

#### `traefik.enable`

```yaml
- "traefik.enable=true"
```

You can tell Traefik to consider (or not) the container by setting `traefik.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### `traefik.swarm.network`

```yaml
- "traefik.swarm.network=mynetwork"
```

Overrides the default docker network to use for connections to the container.

If a container is linked to several networks, be sure to set the proper network name (you can check this with `docker inspect <container_id>`),
otherwise it will randomly pick one (depending on how docker is returning them).

!!! warning
    When deploying a stack from a compose file `stack`, the networks defined are prefixed with `stack`.

#### `traefik.swarm.lbswarm`

```yaml
- "traefik.docker.lbswarm=true"
```

Enables Swarm's inbuilt load balancer (only relevant in Swarm Mode).

If you enable this option, Traefik will use the virtual IP provided by docker swarm instead of the containers IPs.
Which means that Traefik will not perform any kind of load balancing and will delegate this task to swarm.
