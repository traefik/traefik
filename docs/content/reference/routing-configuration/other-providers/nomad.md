---
title: "Traefik Nomad Service Discovery Routing"
description: "Learn how to use Nomad Service Discovery as a provider for routing configurations in Traefik Proxy. Read the technical documentation."
---

# Traefik and Nomad Service Discovery

One of the best feature of Traefik is to delegate the routing configuration to the application level.
With Nomad, Traefik can leverage tags attached to a service to generate routing rules.

!!! warning "Tags & sensitive data"

    We recommend to *not* use tags to store sensitive data (certificates, credentials, etc).
    Instead, we recommend to store sensitive data in a safer storage (secrets, file, etc).

## Routing Configuration

!!! info "Tags"

    Tags are case-insensitive.

!!! tip "TLS Default Generated Certificates"

    To learn how to configure Traefik default generated certificate, refer to the [TLS Certificates](../http/tls/tls-certificates.md#acme-default-certificate) page.

### General

Traefik creates, for each Nomad service, a corresponding Traefik [service](../http/load-balancing/service.md) and [router](../http/router/rules-and-priority.md).

The Traefik service automatically gets a server per instance in this Nomad service, and the router gets a default rule attached to it, based on the Nomad service name.

### Routers

To update the configuration of the Router automatically attached to the service, add tags starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.

For example, to change the rule, you could add the tag ```traefik.http.routers.my-service.rule=Host(`example.com`)```.

??? info "`traefik.http.routers.<router_name>.rule`"

    See [rule](../http/router/rules-and-priority.md) for more information.

    ```yaml
    traefik.http.routers.myrouter.rule=Host(`example.com`)
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
    traefik.http.routers.myrouter.entrypoints=web,websecure
    ```

??? info "`traefik.http.routers.<router_name>.middlewares`"

    See [middlewares overview](../http/middlewares/overview.md) for more information.

    ```yaml
    traefik.http.routers.myrouter.middlewares=auth,prefix,cb
    ```

??? info "`traefik.http.routers.<router_name>.service`"

    See [service](../http/load-balancing/service.md) for more information.

    ```yaml
    traefik.http.routers.myrouter.service=myservice
    ```

??? info "`traefik.http.routers.<router_name>.tls`"

    See [tls](../http/tls/overview.md) for more information.

    ```yaml
    traefik.http.routers.myrouter.tls=true
    ```

??? info "`traefik.http.routers.<router_name>.tls.certresolver`"

    See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information.

    ```yaml
    traefik.http.routers.myrouter.tls.certresolver=myresolver
    ```

??? info "`traefik.http.routers.<router_name>.tls.domains[n].main`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    ```yaml
    traefik.http.routers.myrouter.tls.domains[0].main=example.org
    ```

??? info "`traefik.http.routers.<router_name>.tls.domains[n].sans`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    ```yaml
    traefik.http.routers.myrouter.tls.domains[0].sans=test.example.org,dev.example.org
    ```

??? info "`traefik.http.routers.<router_name>.tls.options`"

    ```yaml
    traefik.http.routers.myrouter.tls.options=foobar
    ```

??? info "`traefik.http.routers.<router_name>.priority`"

    See [priority](../http/router/rules-and-priority.md#priority-calculation) for more information.

    ```yaml
    traefik.http.routers.myrouter.priority=42
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
    
### Services

To update the configuration of the Service automatically attached to the service,
add tags starting with `traefik.http.services.{name-of-your-choice}.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the tag `traefik.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

??? info "`traefik.http.services.<service_name>.loadbalancer.server.port`"

    Registers a port.
    Useful when the service exposes multiples ports.

    ```yaml
    traefik.http.services.myservice.loadbalancer.server.port=8080
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.server.scheme`"

    Overrides the default scheme.

    ```yaml
    traefik.http.services.myservice.loadbalancer.server.scheme=http
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
    traefik.http.services.myservice.loadbalancer.serverstransport=foobar@file
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.passhostheader`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.passhostheader=true
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.headers.X-Foo=foobar
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.hostname`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.hostname=example.org
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.interval`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.interval=10
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.unhealthyinterval`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.unhealthyinterval=10
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.path`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.path=/foo
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.status`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.status=42
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.port`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.port=42
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.scheme`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.scheme=http
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.timeout`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.timeout=10
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.followredirects`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    ```yaml
    traefik.http.services.myservice.loadbalancer.healthcheck.followredirects=true
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.sticky.cookie=true
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.sticky.cookie.httponly=true
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.name`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.sticky.cookie.name=foobar
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.path`"

    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.path=/foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.secure`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.sticky.cookie.secure=true
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.samesite`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.sticky.cookie.samesite=none
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.maxage`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.sticky.cookie.maxage=42
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`"

    ```yaml
    traefik.http.services.myservice.loadbalancer.responseforwarding.flushinterval=10
    ```

### Middleware

You can declare pieces of middleware using tags starting with `traefik.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../http/middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme: https`.

More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md).

??? example "Declaring and Referencing a Middleware"

    ```yaml
    # ...
    # Declaring a middleware
    traefik.http.middlewares.my-redirect.redirectscheme.scheme=https
    # Referencing a middleware
    traefik.http.routers.my-service.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### TCP

You can declare TCP Routers and/or Services using tags.

??? example "Declaring TCP Routers and Services"

    ```yaml
    traefik.tcp.routers.my-router.rule=HostSNI(`example.com`)
    traefik.tcp.routers.my-router.tls=true
    traefik.tcp.services.my-service.loadbalancer.server.port=4123
    ```

!!! warning "TCP and HTTP"

    If you declare a TCP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no TCP Router/Service is defined).
    You can declare both a TCP Router/Service and an HTTP Router/Service for the same Nomad service (but you have to do so manually).

#### TCP Routers

??? info "`traefik.tcp.routers.<router_name>.entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.entrypoints=ep1,ep2
    ```

??? info "`traefik.tcp.routers.<router_name>.rule`"

    See [rule](../tcp/router/rules-and-priority.md#rules) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.rule=HostSNI(`example.com`)
    ```

??? info "`traefik.tcp.routers.<router_name>.ruleSyntax`"

    !!! warning

        RuleSyntax option is deprecated and will be removed in the next major version.
        Please do not use this field and rewrite the router rules to use the v3 syntax.

    configure the rule syntax to be used for parsing the rule on a per-router basis.
    
    ```yaml
    traefik.tcp.routers.mytcprouter.ruleSyntax=v3
    ```
    
??? info "`traefik.tcp.routers.<router_name>.priority`"

    See [priority](../tcp/router/rules-and-priority.md#priority) for more information.

    ```yaml
    traefik.tcp.routers.myrouter.priority=42
    ```

??? info "`traefik.tcp.routers.<router_name>.service`"

    See [service](../tcp/service.md) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.service=myservice
    ```

??? info "`traefik.tcp.routers.<router_name>.tls`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.tls=true
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.certresolver`"

    See [certResolver](../tcp/tls.md#configuration-options) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.tls.certresolver=myresolver
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.domains[n].main`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.tls.domains[0].main=example.org
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.domains[n].sans`"

    See [TLS](../tcp/tls.md) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.tls.domains[0].sans=test.example.org,dev.example.org
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.options`"

    See [TLS](../tcp/tls.md#configuration-options) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.tls.options=myoptions
    ```

??? info "`traefik.tcp.routers.<router_name>.tls.passthrough`"

    See [Passthrough](../tcp/tls.md#passthrough) for more information.

    ```yaml
    traefik.tcp.routers.mytcprouter.tls.passthrough=true
    ```

#### TCP Services

??? info "`traefik.tcp.services.<service_name>.loadbalancer.server.port`"

    Registers a port of the application.

    ```yaml
    traefik.tcp.services.mytcpservice.loadbalancer.server.port=423
    ```

??? info "`traefik.tcp.services.<service_name>.loadbalancer.server.tls`"

    Determines whether to use TLS when dialing with the backend.

    ```yaml
    traefik.tcp.services.mytcpservice.loadbalancer.server.tls=true
    ```

??? info "`traefik.tcp.services.<service_name>.loadbalancer.proxyprotocol.version`"

    See [PROXY protocol](../tcp/service.md#proxy-protocol) for more information.

    ```yaml
    traefik.tcp.services.mytcpservice.loadbalancer.proxyprotocol.version=1
    ```

??? info "`traefik.tcp.services.<service_name>.loadbalancer.serverstransport`"

    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../tcp/serverstransport.md) for more information.

    ```yaml
    traefik.tcp.services.myservice.loadbalancer.serverstransport=foobar@file
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

You can declare UDP Routers and/or Services using tags.

??? example "Declaring UDP Routers and Services"

    ```yaml
    traefik.udp.routers.my-router.entrypoints=udp
    traefik.udp.services.my-service.loadbalancer.server.port=4123
    ```

!!! warning "UDP and HTTP"

    If you declare a UDP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no UDP Router/Service is defined).
    You can declare both a UDP Router/Service and an HTTP Router/Service for the same Nomad service (but you have to do so manually).

#### UDP Routers

??? info "`traefik.udp.routers.<router_name>.entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    ```yaml
    traefik.udp.routers.myudprouter.entrypoints=ep1,ep2
    ```

??? info "`traefik.udp.routers.<router_name>.service`"

    See [service](../udp/service.md) for more information.

    ```yaml
    traefik.udp.routers.myudprouter.service=myservice
    ```

#### UDP Services

??? info "`traefik.udp.services.<service_name>.loadbalancer.server.port`"

    Registers a port of the application.

    ```yaml
    traefik.udp.services.myudpservice.loadbalancer.server.port=423
    ```

### Specific Provider Options

#### `traefik.enable`

```yaml
traefik.enable=true
```

You can tell Traefik to consider (or not) the service by setting `traefik.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### `traefik.nomad.canary`

```yaml
traefik.nomad.canary=true
```

When Nomad orchestrator is a provider (of service registration) for Traefik,
one might have the need to distinguish within Traefik between a [Canary](https://learn.hashicorp.com/tutorials/nomad/job-blue-green-and-canary-deployments#deploy-with-canaries) instance of a service, or a production one.
For example if one does not want them to be part of the same load-balancer.

Therefore, this option, which is meant to be provided as one of the values of the `canary_tags` field in the Nomad [service stanza](https://www.nomadproject.io/docs/job-specification/service#canary_tags),
allows Traefik to identify that the associated instance is a canary one.

#### Port Lookup

Traefik is capable of detecting the port to use, by following the default Nomad Service Discovery flow.
That means, if you just expose lets say port `:1337` on the Nomad job, traefik will pick up this port and use it.
