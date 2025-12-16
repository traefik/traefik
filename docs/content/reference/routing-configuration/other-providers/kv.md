---
title: "Traefik Routing Configuration with KV stores"
description: "Read the technical documentation to learn the Traefik Routing Configuration with KV stores."
---

# Traefik & KV Stores

## Configuration Examples

??? example "Configuring KV Store & Deploying / Exposing one Service"

    Enabling a KV store provider (example: Consul)

    ```yaml tab="Structured (YAML)"
    providers:
      consul:
        endpoints:
          - "127.0.0.1:8500"
    ```

    ```toml tab="Structured (TOML)"
    [providers.consul]
      endpoints = ["127.0.0.1:8500"]
    ```

    ```bash tab="CLI"
    --providers.consul.endpoints=127.0.0.1:8500
    ```

    Setting keys in the KV store (example: Consul)

    ```bash
    consul kv put traefik/http/routers/my-router/rule "Host(`example.com`)"
    consul kv put traefik/http/routers/my-router/service "my-service"
    consul kv put traefik/http/services/my-service/loadbalancer/servers/0/url "http://127.0.0.1:8080"
    ```

??? example "Specify a Custom Port for the Service"

    Forward requests for `http://example.com` to `http://127.0.0.1:12345`:

    ```bash
    consul kv put traefik/http/routers/my-router/rule "Host(`example.com`)"
    consul kv put traefik/http/routers/my-router/service "my-service"
    consul kv put traefik/http/services/my-service/loadbalancer/servers/0/url "http://127.0.0.1:12345"
    ```

??? example "Specifying more than one router and service"

    Forwarding requests to more than one service requires defining multiple routers and services.

    In this example, requests are forwarded for `http://example-a.com` to `http://127.0.0.1:8000` in addition to `http://example-b.com` forwarding to `http://127.0.0.1:9000`:

    ```bash
    consul kv put traefik/http/routers/www-router/rule "Host(`example-a.com`)"
    consul kv put traefik/http/routers/www-router/service "www-service"
    consul kv put traefik/http/services/www-service/loadbalancer/servers/0/url "http://127.0.0.1:8000"
    
    consul kv put traefik/http/routers/admin-router/rule "Host(`example-b.com`)"
    consul kv put traefik/http/routers/admin-router/service "admin-service"
    consul kv put traefik/http/services/admin-service/loadbalancer/servers/0/url "http://127.0.0.1:9000"
    ```

## Configuration Options

!!! info "Keys"

    Keys are case-insensitive.

### HTTP

#### Routers

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

| Key (Path)                           | Description                          | Value                      |
|--------------------------------------|--------------------------------------|----------------------------|
| <a id="opt-traefikhttproutersrouter-namerule" href="#opt-traefikhttproutersrouter-namerule" title="#opt-traefikhttproutersrouter-namerule">`traefik/http/routers/<router_name>/rule`</a> | See [rule](../http/routing/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)```  |
| <a id="opt-traefikhttproutersrouter-nameruleSyntax" href="#opt-traefikhttproutersrouter-nameruleSyntax" title="#opt-traefikhttproutersrouter-nameruleSyntax">`traefik/http/routers/<router_name>/ruleSyntax`</a> | See [rule](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3`  |
| <a id="opt-traefikhttproutersrouter-nameentrypoints0" href="#opt-traefikhttproutersrouter-nameentrypoints0" title="#opt-traefikhttproutersrouter-nameentrypoints0">`traefik/http/routers/<router_name>/entrypoints/0`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `web`       |
| <a id="opt-traefikhttproutersrouter-nameentrypoints1" href="#opt-traefikhttproutersrouter-nameentrypoints1" title="#opt-traefikhttproutersrouter-nameentrypoints1">`traefik/http/routers/<router_name>/entrypoints/1`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `websecure` |
| <a id="opt-traefikhttproutersrouter-namemiddlewares0" href="#opt-traefikhttproutersrouter-namemiddlewares0" title="#opt-traefikhttproutersrouter-namemiddlewares0">`traefik/http/routers/<router_name>/middlewares/0`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth`      |
| <a id="opt-traefikhttproutersrouter-namemiddlewares1" href="#opt-traefikhttproutersrouter-namemiddlewares1" title="#opt-traefikhttproutersrouter-namemiddlewares1">`traefik/http/routers/<router_name>/middlewares/1`</a> |  | `prefix`    |
| <a id="opt-traefikhttproutersrouter-nameservice" href="#opt-traefikhttproutersrouter-nameservice" title="#opt-traefikhttproutersrouter-nameservice">`traefik/http/routers/<router_name>/service`</a> | See [service](../http/load-balancing/service.md) for more information. | `myservice` |
| <a id="opt-traefikhttproutersrouter-nametls" href="#opt-traefikhttproutersrouter-nametls" title="#opt-traefikhttproutersrouter-nametls">`traefik/http/routers/<router_name>/tls`</a> | See [tls](../http/tls/overview.md) for more information. | `true` |
| <a id="opt-traefikhttproutersrouter-nametlscertresolver" href="#opt-traefikhttproutersrouter-nametlscertresolver" title="#opt-traefikhttproutersrouter-nametlscertresolver">`traefik/http/routers/<router_name>/tls/certresolver`</a> | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| <a id="opt-traefikhttproutersrouter-nametlsdomains0main" href="#opt-traefikhttproutersrouter-nametlsdomains0main" title="#opt-traefikhttproutersrouter-nametlsdomains0main">`traefik/http/routers/<router_name>/tls/domains/0/main`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| <a id="opt-traefikhttproutersrouter-nametlsdomains0sans0" href="#opt-traefikhttproutersrouter-nametlsdomains0sans0" title="#opt-traefikhttproutersrouter-nametlsdomains0sans0">`traefik/http/routers/<router_name>/tls/domains/0/sans/0`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `test.example.org` |
| <a id="opt-traefikhttproutersrouter-nametlsdomains0sans1" href="#opt-traefikhttproutersrouter-nametlsdomains0sans1" title="#opt-traefikhttproutersrouter-nametlsdomains0sans1">`traefik/http/routers/<router_name>/tls/domains/0/sans/1`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `dev.example.org`  |
| <a id="opt-traefikhttproutersrouter-nametlsoptions" href="#opt-traefikhttproutersrouter-nametlsoptions" title="#opt-traefikhttproutersrouter-nametlsoptions">`traefik/http/routers/<router_name>/tls/options`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` |
| <a id="opt-traefikhttproutersrouter-nameobservabilityaccesslogs" href="#opt-traefikhttproutersrouter-nameobservabilityaccesslogs" title="#opt-traefikhttproutersrouter-nameobservabilityaccesslogs">`traefik/http/routers/<router_name>/observability/accesslogs`</a> | The accessLogs option controls whether the router will produce access-logs. | `true` |
| <a id="opt-traefikhttproutersrouter-nameobservabilitymetrics" href="#opt-traefikhttproutersrouter-nameobservabilitymetrics" title="#opt-traefikhttproutersrouter-nameobservabilitymetrics">`traefik/http/routers/<router_name>/observability/metrics`</a> | The metrics option controls whether the router will produce metrics. | `true` |
| <a id="opt-traefikhttproutersrouter-nameobservabilitytracing" href="#opt-traefikhttproutersrouter-nameobservabilitytracing" title="#opt-traefikhttproutersrouter-nameobservabilitytracing">`traefik/http/routers/<router_name>/observability/tracing`</a> | The tracing option controls whether the router will produce traces. | `true` |
| <a id="opt-traefikhttproutersrouter-namepriority" href="#opt-traefikhttproutersrouter-namepriority" title="#opt-traefikhttproutersrouter-namepriority">`traefik/http/routers/<router_name>/priority`</a> | See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information. | `42`  |

#### Services

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-traefikhttpservicesmyserviceloadbalancerservers0url" href="#opt-traefikhttpservicesmyserviceloadbalancerservers0url" title="#opt-traefikhttpservicesmyserviceloadbalancerservers0url">`traefik/http/services/myservice/loadbalancer/servers/0/url`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `http://<ip-server-1>:<port-server-1>/` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerservers0preservePath" href="#opt-traefikhttpservicesmyserviceloadbalancerservers0preservePath" title="#opt-traefikhttpservicesmyserviceloadbalancerservers0preservePath">`traefik/http/services/myservice/loadbalancer/servers/0/preservePath`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `true` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerservers0weight" href="#opt-traefikhttpservicesmyserviceloadbalancerservers0weight" title="#opt-traefikhttpservicesmyserviceloadbalancerservers0weight">`traefik/http/services/myservice/loadbalancer/servers/0/weight`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `1` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerserverstransport" href="#opt-traefikhttpservicesmyserviceloadbalancerserverstransport" title="#opt-traefikhttpservicesmyserviceloadbalancerserverstransport">`traefik/http/services/myservice/loadbalancer/serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/> See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerpasshostheader" href="#opt-traefikhttpservicesmyserviceloadbalancerpasshostheader" title="#opt-traefikhttpservicesmyserviceloadbalancerpasshostheader">`traefik/http/services/myservice/loadbalancer/passhostheader`</a> | See [Service](../http/load-balancing/service.md) for more information. | `true` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo">`traefik/http/services/myservice/loadbalancer/healthcheck/headers/X-Foo`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckhostname" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckhostname" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckhostname">`traefik/http/services/myservice/loadbalancer/healthcheck/hostname`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckinterval" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckinterval" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckinterval">`traefik/http/services/myservice/loadbalancer/healthcheck/interval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckpath" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckpath" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckpath">`traefik/http/services/myservice/loadbalancer/healthcheck/path`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckmethod" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckmethod" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckmethod">`traefik/http/services/myservice/loadbalancer/healthcheck/method`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckstatus" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckstatus" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckstatus">`traefik/http/services/myservice/loadbalancer/healthcheck/status`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckport" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckport" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckport">`traefik/http/services/myservice/loadbalancer/healthcheck/port`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthcheckscheme" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckscheme" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthcheckscheme">`traefik/http/services/myservice/loadbalancer/healthcheck/scheme`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerhealthchecktimeout" href="#opt-traefikhttpservicesmyserviceloadbalancerhealthchecktimeout" title="#opt-traefikhttpservicesmyserviceloadbalancerhealthchecktimeout">`traefik/http/services/myservice/loadbalancer/healthcheck/timeout`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| <a id="opt-traefikhttpservicesmyserviceloadbalancersticky" href="#opt-traefikhttpservicesmyserviceloadbalancersticky" title="#opt-traefikhttpservicesmyserviceloadbalancersticky">`traefik/http/services/myservice/loadbalancer/sticky`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerstickycookiehttponly" href="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiehttponly" title="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiehttponly">`traefik/http/services/myservice/loadbalancer/sticky/cookie/httponly`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerstickycookiename" href="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiename" title="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiename">`traefik/http/services/myservice/loadbalancer/sticky/cookie/name`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerstickycookiepath" href="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiepath" title="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiepath">`traefik/http/services/myservice/loadbalancer/sticky/cookie/path`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `/foobar` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerstickycookiesecure" href="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiesecure" title="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiesecure">`traefik/http/services/myservice/loadbalancer/sticky/cookie/secure`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerstickycookiesamesite" href="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiesamesite" title="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiesamesite">`traefik/http/services/myservice/loadbalancer/sticky/cookie/samesite`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `none` |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerstickycookiemaxage" href="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiemaxage" title="#opt-traefikhttpservicesmyserviceloadbalancerstickycookiemaxage">`traefik/http/services/myservice/loadbalancer/sticky/cookie/maxage`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `42`  |
| <a id="opt-traefikhttpservicesmyserviceloadbalancerresponseforwardingflushinterval" href="#opt-traefikhttpservicesmyserviceloadbalancerresponseforwardingflushinterval" title="#opt-traefikhttpservicesmyserviceloadbalancerresponseforwardingflushinterval">`traefik/http/services/myservice/loadbalancer/responseforwarding/flushinterval`</a> | See [Service](../http/load-balancing/service.md) for more information. | `10`  |
| <a id="opt-traefikhttpservicesservice-namemirroringservice" href="#opt-traefikhttpservicesservice-namemirroringservice" title="#opt-traefikhttpservicesservice-namemirroringservice">`traefik/http/services/<service_name>/mirroring/service`</a> | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesservice-namemirroringmirrorsnname" href="#opt-traefikhttpservicesservice-namemirroringmirrorsnname" title="#opt-traefikhttpservicesservice-namemirroringmirrorsnname">`traefik/http/services/<service_name>/mirroring/mirrors/<n>/name`</a> | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesservice-namemirroringmirrorsnpercent" href="#opt-traefikhttpservicesservice-namemirroringmirrorsnpercent" title="#opt-traefikhttpservicesservice-namemirroringmirrorsnpercent">`traefik/http/services/<service_name>/mirroring/mirrors/<n>/percent`</a> | See [Service](../http/load-balancing/service.md#mirroring)for more information. | `42`  |
| <a id="opt-traefikhttpservicesservice-nameweightedservicesnname" href="#opt-traefikhttpservicesservice-nameweightedservicesnname" title="#opt-traefikhttpservicesservice-nameweightedservicesnname">`traefik/http/services/<service_name>/weighted/services/<n>/name`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesservice-nameweightedservicesnweight" href="#opt-traefikhttpservicesservice-nameweightedservicesnweight" title="#opt-traefikhttpservicesservice-nameweightedservicesnweight">`traefik/http/services/<service_name>/weighted/services/<n>/weight`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| <a id="opt-traefikhttpservicesservice-nameweightedstickycookiename" href="#opt-traefikhttpservicesservice-nameweightedstickycookiename" title="#opt-traefikhttpservicesservice-nameweightedstickycookiename">`traefik/http/services/<service_name>/weighted/sticky/cookie/name`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| <a id="opt-traefikhttpservicesservice-nameweightedstickycookiesecure" href="#opt-traefikhttpservicesservice-nameweightedstickycookiesecure" title="#opt-traefikhttpservicesservice-nameweightedstickycookiesecure">`traefik/http/services/<service_name>/weighted/sticky/cookie/secure`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| <a id="opt-traefikhttpservicesservice-nameweightedstickycookiesamesite" href="#opt-traefikhttpservicesservice-nameweightedstickycookiesamesite" title="#opt-traefikhttpservicesservice-nameweightedstickycookiesamesite">`traefik/http/services/<service_name>/weighted/sticky/cookie/samesite`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `none` |
| <a id="opt-traefikhttpservicesservice-nameweightedstickycookiehttpOnly" href="#opt-traefikhttpservicesservice-nameweightedstickycookiehttpOnly" title="#opt-traefikhttpservicesservice-nameweightedstickycookiehttpOnly">`traefik/http/services/<service_name>/weighted/sticky/cookie/httpOnly`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| <a id="opt-traefikhttpservicesservice-nameweightedstickycookiemaxage" href="#opt-traefikhttpservicesservice-nameweightedstickycookiemaxage" title="#opt-traefikhttpservicesservice-nameweightedstickycookiemaxage">`traefik/http/services/<service_name>/weighted/sticky/cookie/maxage`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| <a id="opt-traefikhttpservicesservice-namefailoverfallback" href="#opt-traefikhttpservicesservice-namefailoverfallback" title="#opt-traefikhttpservicesservice-namefailoverfallback">`traefik/http/services/<service_name>/failover/fallback`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `backup`  |
| <a id="opt-traefikhttpservicesservice-namefailoverhealthcheck" href="#opt-traefikhttpservicesservice-namefailoverhealthcheck" title="#opt-traefikhttpservicesservice-namefailoverhealthcheck">`traefik/http/services/<service_name>/failover/healthcheck`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `{}`  |
| <a id="opt-traefikhttpservicesservice-namefailoverservice" href="#opt-traefikhttpservicesservice-namefailoverservice" title="#opt-traefikhttpservicesservice-namefailoverservice">`traefik/http/services/<service_name>/failover/service`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `main`  |

#### Middleware

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-traefikhttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" href="#opt-traefikhttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" title="#opt-traefikhttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option">`traefik/http/middlewares/mymiddleware/middleware_type/middleware_option`</a> | With `middleware_type` the type of middleware (ex: `forwardAuth`, `headers`, etc)<br/>and `middleware_option` the middleware option to set (ex for the middleware `addPrefix`: `prefix`).<br/> More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md). | `foobar` |

!!! warning "The character `@` is not authorized in the middleware name."

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

#### ServerTransport

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-traefikhttpserversTransportsserversTransportNamest-option" href="#opt-traefikhttpserversTransportsserversTransportNamest-option" title="#opt-traefikhttpserversTransportsserversTransportNamest-option">`traefik/http/serversTransports/<serversTransportName>/st_option`</a> | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../http/load-balancing/serverstransport.md). | ServerTransport Options |

### TCP

You can declare TCP Routers and/or Services using KV.

#### Routers

| Key (Path)                                      |  Description | Value |
|-------------------------------------------------|-------------------------------------------------|-------|
| <a id="opt-traefiktcproutersmytcprouterentrypoints0" href="#opt-traefiktcproutersmytcprouterentrypoints0" title="#opt-traefiktcproutersmytcprouterentrypoints0">`traefik/tcp/routers/mytcprouter/entrypoints/0`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1` |
| <a id="opt-traefiktcproutersmytcprouterentrypoints1" href="#opt-traefiktcproutersmytcprouterentrypoints1" title="#opt-traefiktcproutersmytcprouterentrypoints1">`traefik/tcp/routers/mytcprouter/entrypoints/1`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep2` |
| <a id="opt-traefiktcproutersmy-routerrule" href="#opt-traefiktcproutersmy-routerrule" title="#opt-traefiktcproutersmy-routerrule">`traefik/tcp/routers/my-router/rule`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | ```HostSNI(`example.com`)``` |
| <a id="opt-traefiktcproutersmytcprouterservice" href="#opt-traefiktcproutersmytcprouterservice" title="#opt-traefiktcproutersmytcprouterservice">`traefik/tcp/routers/mytcprouter/service`</a> | See [service](../tcp/service.md) for more information. | `myservice` |
| <a id="opt-traefiktcproutersmytcproutertls" href="#opt-traefiktcproutersmytcproutertls" title="#opt-traefiktcproutersmytcproutertls">`traefik/tcp/routers/mytcprouter/tls`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="opt-traefiktcproutersmytcproutertlscertresolver" href="#opt-traefiktcproutersmytcproutertlscertresolver" title="#opt-traefiktcproutersmytcproutertlscertresolver">`traefik/tcp/routers/mytcprouter/tls/certresolver`</a> | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| <a id="opt-traefiktcproutersmytcproutertlsdomains0main" href="#opt-traefiktcproutersmytcproutertlsdomains0main" title="#opt-traefiktcproutersmytcproutertlsdomains0main">`traefik/tcp/routers/mytcprouter/tls/domains/0/main`</a> | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| <a id="opt-traefiktcproutersmytcproutertlsdomains0sans0" href="#opt-traefiktcproutersmytcproutertlsdomains0sans0" title="#opt-traefiktcproutersmytcproutertlsdomains0sans0">`traefik/tcp/routers/mytcprouter/tls/domains/0/sans/0`</a> | See [TLS](../tcp/tls.md) for more information. | `test.example.org` |
| <a id="opt-traefiktcproutersmytcproutertlsdomains0sans1" href="#opt-traefiktcproutersmytcproutertlsdomains0sans1" title="#opt-traefiktcproutersmytcproutertlsdomains0sans1">`traefik/tcp/routers/mytcprouter/tls/domains/0/sans/1`</a> | See [TLS](../tcp/tls.md) for more information. | `dev.example.org`  |
| <a id="opt-traefiktcproutersmytcproutertlsoptions" href="#opt-traefiktcproutersmytcproutertlsoptions" title="#opt-traefiktcproutersmytcproutertlsoptions">`traefik/tcp/routers/mytcprouter/tls/options`</a> | See [TLS](../tcp/tls.md) for more information. | `foobar` |
| <a id="opt-traefiktcproutersmytcproutertlspassthrough" href="#opt-traefiktcproutersmytcproutertlspassthrough" title="#opt-traefiktcproutersmytcproutertlspassthrough">`traefik/tcp/routers/mytcprouter/tls/passthrough`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="opt-traefiktcproutersmytcprouterpriority" href="#opt-traefiktcproutersmytcprouterpriority" title="#opt-traefiktcproutersmytcprouterpriority">`traefik/tcp/routers/mytcprouter/priority`</a> | See [priority](../tcp/routing/rules-and-priority.md#priority-calculation) for more information. | `42`  |

#### Services

| Key (Path)                                                         | Description                                                         | Value            |
|--------------------------------------------------------------------|--------------------------------------------------------------------|------------------|
| <a id="opt-traefiktcpservicesmytcpserviceloadbalancerservers0address" href="#opt-traefiktcpservicesmytcpserviceloadbalancerservers0address" title="#opt-traefiktcpservicesmytcpserviceloadbalancerservers0address">`traefik/tcp/services/mytcpservice/loadbalancer/servers/0/address`</a> | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `xx.xx.xx.xx:xx` |
| <a id="opt-traefiktcpservicesmytcpserviceloadbalancerservers0tls" href="#opt-traefiktcpservicesmytcpserviceloadbalancerservers0tls" title="#opt-traefiktcpservicesmytcpserviceloadbalancerservers0tls">`traefik/tcp/services/mytcpservice/loadbalancer/servers/0/tls`</a> | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `true` |
| <a id="opt-traefiktcpservicesmyserviceloadbalancerserverstransport" href="#opt-traefiktcpservicesmyserviceloadbalancerserverstransport" title="#opt-traefiktcpservicesmyserviceloadbalancerserverstransport">`traefik/tcp/services/myservice/loadbalancer/serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-traefiktcpservicesservice-nameweightedservices0name" href="#opt-traefiktcpservicesservice-nameweightedservices0name" title="#opt-traefiktcpservicesservice-nameweightedservices0name">`traefik/tcp/services/<service_name>/weighted/services/0/name`</a> | See [Service](../tcp/service.md#weighted-round-robin) for more information. | `foobar` |
| <a id="opt-traefiktcpservicesservice-nameweightedservices0weight" href="#opt-traefiktcpservicesservice-nameweightedservices0weight" title="#opt-traefiktcpservicesservice-nameweightedservices0weight">`traefik/tcp/services/<service_name>/weighted/services/0/weight`</a> | See [Service](../tcp/service.md#weighted-round-robin) for more information. | `42`  |

#### Middleware

##### Configuration Options

You can declare pieces of middleware using tags starting with `traefik/tcp/middlewares/{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`InFlightConn`](../tcp/middlewares/inflightconn.md) named `test-inflightconn`, you'd write `traefik/tcp/middlewares/test-inflightconn/inflightconn/amount=10`.

More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md).

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-traefiktcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" href="#opt-traefiktcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" title="#opt-traefiktcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option">`traefik/tcp/middlewares/mymiddleware/middleware_type/middleware_option`</a> | With `middleware_type` the type of middleware (ex: `inflightconn`)<br/>and `middleware_option` the middleware option to set (ex for the middleware `inflightconn`: `amount`).<br/> More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md). | `foobar` |

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

#### ServerTransport

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-traefiktcpserversTransportsserversTransportNamest-option" href="#opt-traefiktcpserversTransportsserversTransportNamest-option" title="#opt-traefiktcpserversTransportsserversTransportNamest-option">`traefik/tcp/serversTransports/<serversTransportName>/st_option`</a> | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../tcp/serverstransport.md). | ServerTransport Options |

### UDP

You can declare UDP Routers and/or Services using KV.

#### Routers

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| <a id="opt-traefikudproutersmyudprouterentrypoints0" href="#opt-traefikudproutersmyudprouterentrypoints0" title="#opt-traefikudproutersmyudprouterentrypoints0">`traefik/udp/routers/myudprouter/entrypoints/0`</a> | See [UDP Router](../udp/routing/rules-priority.md#entrypoints) for more information. | `foobar`  |
| <a id="opt-traefikudproutersmyudprouterservice" href="#opt-traefikudproutersmyudprouterservice" title="#opt-traefikudproutersmyudprouterservice">`traefik/udp/routers/myudprouter/service`</a> | See [UDP Router](../udp/routing/rules-priority.md#configuration-example) for more information. | `foobar`  |

#### Services

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| <a id="opt-traefikudpservicesloadBalancerserversnaddress" href="#opt-traefikudpservicesloadBalancerserversnaddress" title="#opt-traefikudpservicesloadBalancerserversnaddress">`traefik/udp/services/loadBalancer/servers/<n>/address`</a> | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| <a id="opt-traefikudpservicesweightedservices0name" href="#opt-traefikudpservicesweightedservices0name" title="#opt-traefikudpservicesweightedservices0name">`traefik/udp/services/weighted/services/0/name`</a> | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| <a id="opt-traefikudpservicesweightedservers0weight" href="#opt-traefikudpservicesweightedservers0weight" title="#opt-traefikudpservicesweightedservers0weight">`traefik/udp/services/weighted/servers/0/weight`</a> |See [UDP Service](../udp/service.md) for more information. | `42`  | 

## TLS

### TLS Options

With the KV provider, you configure some parameters of the TLS connection using the `tls/options` key.

For example, you can define a basic setup like this:

| Key (Path)                                           | Description                                           | Value    |
|------------------------------------------------------|------------------------------------------------------|----------|
| <a id="opt-traefiktlsoptionsOptions0alpnProtocols0" href="#opt-traefiktlsoptionsOptions0alpnProtocols0" title="#opt-traefiktlsoptionsOptions0alpnProtocols0">`traefik/tls/options/Options0/alpnProtocols/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="opt-traefiktlsoptionsOptions0cipherSuites0" href="#opt-traefiktlsoptionsOptions0cipherSuites0" title="#opt-traefiktlsoptionsOptions0cipherSuites0">`traefik/tls/options/Options0/cipherSuites/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="opt-traefiktlsoptionsOptions0clientAuthcaFiles0" href="#opt-traefiktlsoptionsOptions0clientAuthcaFiles0" title="#opt-traefiktlsoptionsOptions0clientAuthcaFiles0">`traefik/tls/options/Options0/clientAuth/caFiles/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="opt-traefiktlsoptionsOptions0disableSessiontickets" href="#opt-traefiktlsoptionsOptions0disableSessiontickets" title="#opt-traefiktlsoptionsOptions0disableSessiontickets">`traefik/tls/options/Options0/disableSessiontickets`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. |  `true`   |

### TLS Default Generated Certificates

You can configure Traefik to use an ACME provider (like Let's Encrypt) to generate the default certificate.

The configuration to resolve the default certificate should be defined in a TLS store.

| Key (Path)                                                     | Description                                                     | Value    |
|----------------------------------------------------------------|----------------------------------------------------------------|----------|
| <a id="opt-traefiktlsstoresStore0defaultGeneratedCertdomainmain" href="#opt-traefiktlsstoresStore0defaultGeneratedCertdomainmain" title="#opt-traefiktlsstoresStore0defaultGeneratedCertdomainmain">`traefik/tls/stores/Store0/defaultGeneratedCert/domain/main`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information. | `foobar` |
| <a id="opt-traefiktlsstoresStore0defaultGeneratedCertdomainsans0" href="#opt-traefiktlsstoresStore0defaultGeneratedCertdomainsans0" title="#opt-traefiktlsstoresStore0defaultGeneratedCertdomainsans0">`traefik/tls/stores/Store0/defaultGeneratedCert/domain/sans/0`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
| <a id="opt-traefiktlsstoresStore0defaultGeneratedCertdomainsans1" href="#opt-traefiktlsstoresStore0defaultGeneratedCertdomainsans1" title="#opt-traefiktlsstoresStore0defaultGeneratedCertdomainsans1">`traefik/tls/stores/Store0/defaultGeneratedCert/domain/sans/1`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
| <a id="opt-traefiktlsstoresStore0defaultGeneratedCertresolver" href="#opt-traefiktlsstoresStore0defaultGeneratedCertresolver" title="#opt-traefiktlsstoresStore0defaultGeneratedCertresolver">`traefik/tls/stores/Store0/defaultGeneratedCert/resolver`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
