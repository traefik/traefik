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

Traefik creates, for each Nomad service, a corresponding Traefik [service](../http/load-balancing/service.md) and [router](../http/routing/rules-and-priority.md).

The Traefik service automatically gets a server per instance in this Nomad service, and the router gets a default rule attached to it, based on the Nomad service name.

### Routers

To update the configuration of the Router automatically attached to the service, add tags starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.

For example, to change the rule, you could add the tag ```traefik.http.routers.my-service.rule=Host(`example.com`)```.

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefikhttproutersrouter-namerule" href="#opt-traefikhttproutersrouter-namerule" title="#opt-traefikhttproutersrouter-namerule">`traefik.http.routers.<router_name>.rule`</a> | See [rule](../http/routing/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)``` |
| <a id="opt-traefikhttproutersrouter-nameruleSyntax" href="#opt-traefikhttproutersrouter-nameruleSyntax" title="#opt-traefikhttproutersrouter-nameruleSyntax">`traefik.http.routers.<router_name>.ruleSyntax`</a> | See [ruleSyntax](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3` |
| <a id="opt-traefikhttproutersrouter-nameentrypoints" href="#opt-traefikhttproutersrouter-nameentrypoints" title="#opt-traefikhttproutersrouter-nameentrypoints">`traefik.http.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `web,websecure` |
| <a id="opt-traefikhttproutersrouter-namemiddlewares" href="#opt-traefikhttproutersrouter-namemiddlewares" title="#opt-traefikhttproutersrouter-namemiddlewares">`traefik.http.routers.<router_name>.middlewares`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth,prefix,cb` |
| <a id="opt-traefikhttproutersrouter-nameservice" href="#opt-traefikhttproutersrouter-nameservice" title="#opt-traefikhttproutersrouter-nameservice">`traefik.http.routers.<router_name>.service`</a> | See [service](../http/load-balancing/service.md) for more information. | `myservice` |
| <a id="opt-traefikhttproutersrouter-nametls" href="#opt-traefikhttproutersrouter-nametls" title="#opt-traefikhttproutersrouter-nametls">`traefik.http.routers.<router_name>.tls`</a> | See [tls](../http/tls/overview.md) for more information. | `true` |
| <a id="opt-traefikhttproutersrouter-nametlscertresolver" href="#opt-traefikhttproutersrouter-nametlscertresolver" title="#opt-traefikhttproutersrouter-nametlscertresolver">`traefik.http.routers.<router_name>.tls.certresolver`</a> | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| <a id="opt-traefikhttproutersrouter-nametlsdomainsnmain" href="#opt-traefikhttproutersrouter-nametlsdomainsnmain" title="#opt-traefikhttproutersrouter-nametlsdomainsnmain">`traefik.http.routers.<router_name>.tls.domains[n].main`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| <a id="opt-traefikhttproutersrouter-nametlsdomainsnsans" href="#opt-traefikhttproutersrouter-nametlsdomainsnsans" title="#opt-traefikhttproutersrouter-nametlsdomainsnsans">`traefik.http.routers.<router_name>.tls.domains[n].sans`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `test.example.org,dev.example.org` |
| <a id="opt-traefikhttproutersrouter-nametlsoptions" href="#opt-traefikhttproutersrouter-nametlsoptions" title="#opt-traefikhttproutersrouter-nametlsoptions">`traefik.http.routers.<router_name>.tls.options`</a> |  | `foobar` |
| <a id="opt-traefikhttproutersrouter-namepriority" href="#opt-traefikhttproutersrouter-namepriority" title="#opt-traefikhttproutersrouter-namepriority">`traefik.http.routers.<router_name>.priority`</a> | See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information. | `42` |
| <a id="opt-traefikhttproutersrouter-nameobservabilityaccesslogs" href="#opt-traefikhttproutersrouter-nameobservabilityaccesslogs" title="#opt-traefikhttproutersrouter-nameobservabilityaccesslogs">`traefik.http.routers.<router_name>.observability.accesslogs`</a> | The accessLogs option controls whether the router will produce access-logs. | `true` |
| <a id="opt-traefikhttproutersrouter-nameobservabilitymetrics" href="#opt-traefikhttproutersrouter-nameobservabilitymetrics" title="#opt-traefikhttproutersrouter-nameobservabilitymetrics">`traefik.http.routers.<router_name>.observability.metrics`</a> | The metrics option controls whether the router will produce metrics. | `true` |
| <a id="opt-traefikhttproutersrouter-nameobservabilitytracing" href="#opt-traefikhttproutersrouter-nameobservabilitytracing" title="#opt-traefikhttproutersrouter-nameobservabilitytracing">`traefik.http.routers.<router_name>.observability.tracing`</a> | The tracing option controls whether the router will produce traces. | `true` |
    
### Services

To update the configuration of the Service automatically attached to the service,
add tags starting with `traefik.http.services.{name-of-your-choice}.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the tag `traefik.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefikhttpservicesservice-nameloadbalancerserverport" href="#opt-traefikhttpservicesservice-nameloadbalancerserverport" title="#opt-traefikhttpservicesservice-nameloadbalancerserverport">`traefik.http.services.<service_name>.loadbalancer.server.port`</a> | Registers a port.<br/>Useful when the service exposes multiples ports. | `8080` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerserverscheme" href="#opt-traefikhttpservicesservice-nameloadbalancerserverscheme" title="#opt-traefikhttpservicesservice-nameloadbalancerserverscheme">`traefik.http.services.<service_name>.loadbalancer.server.scheme`</a> | Overrides the default scheme. | `http` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerserverweight" href="#opt-traefikhttpservicesservice-nameloadbalancerserverweight" title="#opt-traefikhttpservicesservice-nameloadbalancerserverweight">`traefik.http.services.<service_name>.loadbalancer.server.weight`</a> | Overrides the default weight. | `42` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerserverstransport" href="#opt-traefikhttpservicesservice-nameloadbalancerserverstransport" title="#opt-traefikhttpservicesservice-nameloadbalancerserverstransport">`traefik.http.services.<service_name>.loadbalancer.serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerpasshostheader" href="#opt-traefikhttpservicesservice-nameloadbalancerpasshostheader" title="#opt-traefikhttpservicesservice-nameloadbalancerpasshostheader">`traefik.http.services.<service_name>.loadbalancer.passhostheader`</a> |  | `true` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckheadersheader-name" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckheadersheader-name" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckheadersheader-name">`traefik.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckhostname" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckhostname" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckhostname">`traefik.http.services.<service_name>.loadbalancer.healthcheck.hostname`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckinterval" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckinterval" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckinterval">`traefik.http.services.<service_name>.loadbalancer.healthcheck.interval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckunhealthyinterval" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckunhealthyinterval" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckunhealthyinterval">`traefik.http.services.<service_name>.loadbalancer.healthcheck.unhealthyinterval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckpath" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckpath" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckpath">`traefik.http.services.<service_name>.loadbalancer.healthcheck.path`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckstatus" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckstatus" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckstatus">`traefik.http.services.<service_name>.loadbalancer.healthcheck.status`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckport" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckport" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckport">`traefik.http.services.<service_name>.loadbalancer.healthcheck.port`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckscheme" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckscheme" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckscheme">`traefik.http.services.<service_name>.loadbalancer.healthcheck.scheme`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthchecktimeout" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthchecktimeout" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthchecktimeout">`traefik.http.services.<service_name>.loadbalancer.healthcheck.timeout`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerhealthcheckfollowredirects" href="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckfollowredirects" title="#opt-traefikhttpservicesservice-nameloadbalancerhealthcheckfollowredirects">`traefik.http.services.<service_name>.loadbalancer.healthcheck.followredirects`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `true` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookie" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookie" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookie">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie`</a> |  | `true` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookiehttponly" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiehttponly" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiehttponly">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`</a> |  | `true` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookiename" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiename" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiename">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.name`</a> |  | `foobar` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookiepath" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiepath" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiepath">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.path`</a> |  | `/foobar` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookiesecure" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiesecure" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiesecure">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.secure`</a> |  | `true` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookiesamesite" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiesamesite" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiesamesite">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.samesite`</a> |  | `none` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerstickycookiemaxage" href="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiemaxage" title="#opt-traefikhttpservicesservice-nameloadbalancerstickycookiemaxage">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.maxage`</a> |  | `42` |
| <a id="opt-traefikhttpservicesservice-nameloadbalancerresponseforwardingflushinterval" href="#opt-traefikhttpservicesservice-nameloadbalancerresponseforwardingflushinterval" title="#opt-traefikhttpservicesservice-nameloadbalancerresponseforwardingflushinterval">`traefik.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`</a> |  | `10` |

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

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefiktcproutersrouter-nameentrypoints" href="#opt-traefiktcproutersrouter-nameentrypoints" title="#opt-traefiktcproutersrouter-nameentrypoints">`traefik.tcp.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1,ep2` |
| <a id="opt-traefiktcproutersrouter-namerule" href="#opt-traefiktcproutersrouter-namerule" title="#opt-traefiktcproutersrouter-namerule">`traefik.tcp.routers.<router_name>.rule`</a> | See [rule](../tcp/routing/rules-and-priority.md#rules) for more information. | ```HostSNI(`example.com`)``` |
| <a id="opt-traefiktcproutersrouter-nameruleSyntax" href="#opt-traefiktcproutersrouter-nameruleSyntax" title="#opt-traefiktcproutersrouter-nameruleSyntax">`traefik.tcp.routers.<router_name>.ruleSyntax`</a> | configure the rule syntax to be used for parsing the rule on a per-router basis.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3` |
| <a id="opt-traefiktcproutersrouter-namepriority" href="#opt-traefiktcproutersrouter-namepriority" title="#opt-traefiktcproutersrouter-namepriority">`traefik.tcp.routers.<router_name>.priority`</a> | See [priority](../tcp/routing/rules-and-priority.md#priority-calculation) for more information. | `42` |
| <a id="opt-traefiktcproutersrouter-nameservice" href="#opt-traefiktcproutersrouter-nameservice" title="#opt-traefiktcproutersrouter-nameservice">`traefik.tcp.routers.<router_name>.service`</a> | See [service](../tcp/service.md) for more information. | `myservice` |
| <a id="opt-traefiktcproutersrouter-nametls" href="#opt-traefiktcproutersrouter-nametls" title="#opt-traefiktcproutersrouter-nametls">`traefik.tcp.routers.<router_name>.tls`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="opt-traefiktcproutersrouter-nametlscertresolver" href="#opt-traefiktcproutersrouter-nametlscertresolver" title="#opt-traefiktcproutersrouter-nametlscertresolver">`traefik.tcp.routers.<router_name>.tls.certresolver`</a> | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| <a id="opt-traefiktcproutersrouter-nametlsdomainsnmain" href="#opt-traefiktcproutersrouter-nametlsdomainsnmain" title="#opt-traefiktcproutersrouter-nametlsdomainsnmain">`traefik.tcp.routers.<router_name>.tls.domains[n].main`</a> | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| <a id="opt-traefiktcproutersrouter-nametlsdomainsnsans" href="#opt-traefiktcproutersrouter-nametlsdomainsnsans" title="#opt-traefiktcproutersrouter-nametlsdomainsnsans">`traefik.tcp.routers.<router_name>.tls.domains[n].sans`</a> | See [TLS](../tcp/tls.md) for more information. | `test.example.org,dev.example.org` |
| <a id="opt-traefiktcproutersrouter-nametlsoptions" href="#opt-traefiktcproutersrouter-nametlsoptions" title="#opt-traefiktcproutersrouter-nametlsoptions">`traefik.tcp.routers.<router_name>.tls.options`</a> | See [TLS](../tcp/tls.md#configuration-options) for more information. | `myoptions` |
| <a id="opt-traefiktcproutersrouter-nametlspassthrough" href="#opt-traefiktcproutersrouter-nametlspassthrough" title="#opt-traefiktcproutersrouter-nametlspassthrough">`traefik.tcp.routers.<router_name>.tls.passthrough`</a> | See [Passthrough](../tcp/tls.md#opt-passthrough) for more information. | `true` |

#### TCP Services

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefiktcpservicesservice-nameloadbalancerserverport" href="#opt-traefiktcpservicesservice-nameloadbalancerserverport" title="#opt-traefiktcpservicesservice-nameloadbalancerserverport">`traefik.tcp.services.<service_name>.loadbalancer.server.port`</a> | Registers a port of the application. | `423` |
| <a id="opt-traefiktcpservicesservice-nameloadbalancerservertls" href="#opt-traefiktcpservicesservice-nameloadbalancerservertls" title="#opt-traefiktcpservicesservice-nameloadbalancerservertls">`traefik.tcp.services.<service_name>.loadbalancer.server.tls`</a> | Determines whether to use TLS when dialing with the backend. | `true` |
| <a id="opt-traefiktcpservicesservice-nameloadbalancerserverstransport" href="#opt-traefiktcpservicesservice-nameloadbalancerserverstransport" title="#opt-traefiktcpservicesservice-nameloadbalancerserverstransport">`traefik.tcp.services.<service_name>.loadbalancer.serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |

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

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefikudproutersrouter-nameentrypoints" href="#opt-traefikudproutersrouter-nameentrypoints" title="#opt-traefikudproutersrouter-nameentrypoints">`traefik.udp.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1,ep2` |
| <a id="opt-traefikudproutersrouter-nameservice" href="#opt-traefikudproutersrouter-nameservice" title="#opt-traefikudproutersrouter-nameservice">`traefik.udp.routers.<router_name>.service`</a> | See [service](../udp/service.md) for more information. | `myservice` |

#### UDP Services

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefikudpservicesservice-nameloadbalancerserverport" href="#opt-traefikudpservicesservice-nameloadbalancerserverport" title="#opt-traefikudpservicesservice-nameloadbalancerserverport">`traefik.udp.services.<service_name>.loadbalancer.server.port`</a> | Registers a port of the application. | `423` |

### Specific Provider Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefikenable" href="#opt-traefikenable" title="#opt-traefikenable">`traefik.enable`</a> | You can tell Traefik to consider (or not) the service by setting `traefik.enable` to true or false.<br/>This option overrides the value of `exposedByDefault`. | `true` |
| <a id="opt-traefiknomadcanary" href="#opt-traefiknomadcanary" title="#opt-traefiknomadcanary">`traefik.nomad.canary`</a> | When Nomad orchestrator is a provider (of service registration) for Traefik, one might have the need to distinguish within Traefik between a [Canary](https://learn.hashicorp.com/tutorials/nomad/job-blue-green-and-canary-deployments#deploy-with-canaries) instance of a service, or a production one.<br/>For example if one does not want them to be part of the same load-balancer.<br/><br/>Therefore, this option, which is meant to be provided as one of the values of the `canary_tags` field in the Nomad [service stanza](https://www.nomadproject.io/docs/job-specification/service#canary_tags), allows Traefik to identify that the associated instance is a canary one. | `true` |

#### Port Lookup

Traefik is capable of detecting the port to use, by following the default Nomad Service Discovery flow.
That means, if you just expose lets say port `:1337` on the Nomad job, traefik will pick up this port and use it.
