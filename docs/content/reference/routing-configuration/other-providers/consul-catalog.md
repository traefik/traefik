---
title: "Baqup Consul Catalog Routing"
description: "Learn how to use Consul Catalog as a provider for routing configurations in Baqup Proxy. Read the technical documentation."
---

# Baqup & Consul Catalog

One of the best feature of Baqup is to delegate the routing configuration to the application level.
With Consul Catalog, Baqup can leverage tags attached to a service to generate routing rules.

!!! warning "Tags & sensitive data"

    We recommend to *not* use tags to store sensitive data (certificates, credentials, etc).
    Instead, we recommend to store sensitive data in a safer storage (secrets, file, etc).

## Routing Configuration

!!! info "tags"
    
    Tags are case-insensitive.

!!! tip "TLS Default Generated Certificates"

    To learn how to configure Baqup default generated certificate, refer to the [TLS Certificates](../http/tls/tls-certificates.md#acme-default-certificate) page.

### General

Baqup creates, for each consul Catalog service, a corresponding [service](../http/load-balancing/service.md) and [router](../http/routing/rules-and-priority.md).

The Service automatically gets a server per instance in this consul Catalog service, and the router gets a default rule attached to it, based on the service name.

### Routers

To update the configuration of the Router automatically attached to the service, add tags starting with `baqup.routers.{name-of-your-choice}.` and followed by the option you want to change.

For example, to change the rule, you could add the tag ```baqup.http.routers.my-service.rule=Host(`example.com`)```.

??? info "`baqup.http.routers.<router_name>.rule`"
    
    See [rule](../http/routing/rules-and-priority.md) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.rule=Host(`example.com`)
    ```

??? info "`baqup.http.routers.<router_name>.ruleSyntax`"

    !!! warning

        RuleSyntax option is deprecated and will be removed in the next major version.
        Please do not use this field and rewrite the router rules to use the v3 syntax.

    See [ruleSyntax](../http/routing/rules-and-priority.md#rulesyntax) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.ruleSyntax=v3
    ```

??? info "`baqup.http.routers.<router_name>.priority`"

    See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information.

    ```yaml
    - "baqup.tcp.routers.mytcprouter.priority=42"
    ```

??? info "`baqup.http.routers.<router_name>.entrypoints`"
    
    See [entry points](../../install-configuration/entrypoints.md) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.entrypoints=web,websecure
    ```

??? info "`baqup.http.routers.<router_name>.middlewares`"
    
    See [middlewares overview](../http/middlewares/overview.md) for more information.
  
    ```yaml
    baqup.http.routers.myrouter.middlewares=auth,prefix,cb
    ```

??? info "`baqup.http.routers.<router_name>.service`"
    
    See [service](../http/load-balancing/service.md) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.service=myservice
    ```

??? info "`baqup.http.routers.<router_name>.tls`"
    
    See [tls](../http/tls/overview.md) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.tls=true
    ```

??? info "`baqup.http.routers.<router_name>.tls.certresolver`"
    
    See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.tls.certresolver=myresolver
    ```

??? info "`baqup.http.routers.<router_name>.tls.domains[n].main`"
    
    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.tls.domains[0].main=example.org
    ```

??? info "`baqup.http.routers.<router_name>.tls.domains[n].sans`"
    
    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.
    
    ```yaml
    baqup.http.routers.myrouter.tls.domains[0].sans=test.example.org,dev.example.org
    ```

??? info "`baqup.http.routers.<router_name>.tls.options`"
    
    ```yaml
    baqup.http.routers.myrouter.tls.options=foobar
    ```

??? info "`baqup.http.routers.<router_name>.observability.accesslogs`"
    
    The accessLogs option controls whether the router will produce access-logs.
    
    ```yaml
     "baqup.http.routers.myrouter.observability.accesslogs=true"
    ```

??? info "`baqup.http.routers.<router_name>.observability.metrics`"
    
    The metrics option controls whether the router will produce metrics.

    ```yaml
     "baqup.http.routers.myrouter.observability.metrics=true"
    ```

??? info "`baqup.http.routers.<router_name>.observability.tracing`"
    
    The tracing option controls whether the router will produce traces.

    ```yaml
     "baqup.http.routers.myrouter.observability.tracing=true"
    ```

### Services

To update the configuration of the Service automatically attached to the service,
add tags starting with `baqup.http.services.{name-of-your-choice}.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the tag `baqup.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

??? info "`baqup.http.services.<service_name>.loadbalancer.server.port`"
    
    Registers a port.
    Useful when the service exposes multiples ports.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.server.port=8080
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.server.scheme`"
    
    Overrides the default scheme.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.server.scheme=http
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.server.weight`"

    Overrides the default weight.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.server.weight=42
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.serverstransport`"
    
    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../http/load-balancing/serverstransport.md) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.serverstransport=foobar@file
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.passhostheader`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.passhostheader=true
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.headers.X-Foo=foobar
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.hostname`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.hostname=example.org
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.interval`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.interval=10
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.unhealthyinterval`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.unhealthyinterval=10
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.path`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.path=/foo
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.method`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.method=foobar
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.status`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.status=42
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.port`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.port=42
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.scheme`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.scheme=http
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.timeout`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.timeout=10
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.healthcheck.followredirects`"
    
    See [health check](../http/load-balancing/service.md#health-check) for more information.
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.healthcheck.followredirects=true
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie=true
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie.httponly=true
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie.name`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie.name=foobar
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie.path`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie.path=/foobar
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie.secure`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie.secure=true
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie.samesite`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie.samesite=none
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.sticky.cookie.maxage`"
    
    ```yaml
    baqup.http.services.myservice.loadbalancer.sticky.cookie.maxage=42
    ```

??? info "`baqup.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`"

    ```yaml
    baqup.http.services.myservice.loadbalancer.responseforwarding.flushinterval=10
    ```

### Middleware

You can declare pieces of middleware using tags starting with `baqup.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../http/middlewares/redirectscheme.md) named `my-redirect`, you'd write `baqup.http.middlewares.my-redirect.redirectscheme.scheme: https`.

More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md).

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    # Declaring a middleware
    baqup.http.middlewares.my-redirect.redirectscheme.scheme=https
    # Referencing a middleware
    baqup.http.routers.my-service.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### TCP

You can declare TCP Routers, Middlewares and/or Services using tags.

??? example "Declaring TCP Routers and Services"

    ```yaml
    baqup.tcp.routers.my-router.rule=HostSNI(`example.com`)
    baqup.tcp.routers.my-router.tls=true
    baqup.tcp.services.my-service.loadbalancer.server.port=4123
    ```

!!! warning "TCP and HTTP"

    If you declare a TCP Router/Service, it will prevent Baqup from automatically creating an HTTP Router/Service (like it does by default if no TCP Router/Service is defined).
    You can declare both a TCP Router/Service and an HTTP Router/Service for the same consul service (but you have to do so manually).

#### TCP Routers

??? info "`baqup.tcp.routers.<router_name>.entrypoints`"
    
    See [entry points](../../install-configuration/entrypoints.md) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.entrypoints=ep1,ep2
    ```

??? info "`baqup.tcp.routers.<router_name>.rule`"
    
    See [rule](../tcp/routing/rules-and-priority.md#rules) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.rule=HostSNI(`example.com`)
    ```

??? info "`baqup.tcp.routers.<router_name>.ruleSyntax`"

    !!! warning

        RuleSyntax option is deprecated and will be removed in the next major version.
        Please do not use this field and rewrite the router rules to use the v3 syntax.

    configure the rule syntax to be used for parsing the rule on a per-router basis.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.ruleSyntax=v3
    ```

??? info "`baqup.tcp.routers.<router_name>.priority`"
    See [priority](../tcp/routing/rules-and-priority.md#priority-calculation) for more information.
    ```yaml
    - "baqup.tcp.routers.mytcprouter.priority=42"
    ```
    
??? info "`baqup.tcp.routers.<router_name>.service`"
    
    See [service](../tcp/service.md) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.service=myservice
    ```

??? info "`baqup.tcp.routers.<router_name>.tls`"
    
    See [TLS](../tcp/tls.md) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.tls=true
    ```

??? info "`baqup.tcp.routers.<router_name>.tls.certresolver`"
    
    See [certResolver](../tcp/tls.md#configuration-options) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.tls.certresolver=myresolver
    ```

??? info "`baqup.tcp.routers.<router_name>.tls.domains[n].main`"
    
    See [TLS](../tcp/tls.md) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.tls.domains[0].main=example.org
    ```

??? info "`baqup.tcp.routers.<router_name>.tls.domains[n].sans`"
    
    See [TLS](../tcp/tls.md) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.tls.domains[0].sans=test.example.org,dev.example.org
    ```

??? info "`baqup.tcp.routers.<router_name>.tls.options`"
    
    See [TLS](../tcp/tls.md) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.tls.options=mysoptions
    ```

??? info "`baqup.tcp.routers.<router_name>.tls.passthrough`"
    
    See [Passthrough](../tcp/tls.md#opt-passthrough) for more information.
    
    ```yaml
    baqup.tcp.routers.mytcprouter.tls.passthrough=true
    ```
    

#### TCP Services

??? info "`baqup.tcp.services.<service_name>.loadbalancer.server.port`"
    
    Registers a port of the application.
    
    ```yaml
    baqup.tcp.services.mytcpservice.loadbalancer.server.port=423
    ```

??? info "`baqup.tcp.services.<service_name>.loadbalancer.server.tls`"
    
    Determines whether to use TLS when dialing with the backend.
    
    ```yaml
    baqup.tcp.services.mytcpservice.loadbalancer.server.tls=true
    ```

??? info "`baqup.tcp.services.<service_name>.loadbalancer.serverstransport`"

    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../tcp/serverstransport.md) for more information.
    
    ```yaml
    baqup.tcp.services.mytcpservice.loadbalancer.serverstransport=foobar@file
    ```

#### TCP Middleware

You can declare pieces of middleware using tags starting with `baqup.tcp.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`InFlightConn`](../tcp/middlewares/inflightconn.md) named `test-inflightconn`, you'd write `baqup.tcp.middlewares.test-inflightconn.inflightconn.amount=10`.

More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md).

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    # Declaring a middleware
    baqup.tcp.middlewares.test-inflightconn.amount=10
    # Referencing a middleware
    baqup.tcp.routers.my-service.middlewares=test-inflightconn
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### UDP

You can declare UDP Routers and/or Services using tags.

??? example "Declaring UDP Routers and Services"

    ```yaml
    baqup.udp.routers.my-router.entrypoints=udp
    baqup.udp.services.my-service.loadbalancer.server.port=4123
    ```

!!! warning "UDP and HTTP"

    If you declare a UDP Router/Service, it will prevent Baqup from automatically creating an HTTP Router/Service (like it does by default if no UDP Router/Service is defined).
    You can declare both a UDP Router/Service and an HTTP Router/Service for the same consul service (but you have to do so manually).

#### UDP Routers

??? info "`baqup.udp.routers.<router_name>.entrypoints`"
    
    See [entry points](../../install-configuration/entrypoints.md) for more information.
    
    ```yaml
    baqup.udp.routers.myudprouter.entrypoints=ep1,ep2
    ```

??? info "`baqup.udp.routers.<router_name>.service`"
    
    See [service](../udp/service.md) for more information.
    
    ```yaml
    baqup.udp.routers.myudprouter.service=myservice
    ```

#### UDP Services

??? info "`baqup.udp.services.<service_name>.loadbalancer.server.port`"
    
    Registers a port of the application.
    
    ```yaml
    baqup.udp.services.myudpservice.loadbalancer.server.port=423
    ```

### Specific Provider Options

#### `baqup.enable`

```yaml
baqup.enable=true
```

You can tell Baqup to consider (or not) the service by setting `baqup.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### `baqup.consulcatalog.connect`

```yaml
baqup.consulcatalog.connect=true
```

You can tell Baqup to consider (or not) the service as a Connect capable one by setting `baqup.consulcatalog.connect` to true or false.

This option overrides the value of `connectByDefault`.

#### `baqup.consulcatalog.canary`

```yaml
baqup.consulcatalog.canary=true
```

When ConsulCatalog, in the context of a Nomad orchestrator,
is a provider (of service registration) for Baqup,
one might have the need to distinguish within Baqup between a [Canary](https://learn.hashicorp.com/tutorials/nomad/job-blue-green-and-canary-deployments#deploy-with-canaries) instance of a service, or a production one.
For example if one does not want them to be part of the same load-balancer.

Therefore, this option, which is meant to be provided as one of the values of the `canary_tags` field in the Nomad [service stanza](https://www.nomadproject.io/docs/job-specification/service#canary_tags),
allows Baqup to identify that the associated instance is a canary one.

#### Port Lookup

Baqup is capable of detecting the port to use, by following the default consul Catalog flow.
That means, if you just expose lets say port `:1337` on the consul Catalog ui, baqup will pick up this port and use it.
