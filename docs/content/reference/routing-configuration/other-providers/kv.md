---
title: "Traefik Routing Configuration with KV stores"
description: "Read the technical documentation to learn the Traefik Routing Configuration with KV stores."
---

# Traefik & KV Stores

## Configuration Options

!!! info "Keys"

    Keys are case-insensitive.

### HTTP

#### Routers

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

| Key (Path)                           | Description                          | Value                      |
|--------------------------------------|--------------------------------------|----------------------------|
| <a id="traefikhttproutersrouter-namerule" href="#traefikhttproutersrouter-namerule" title="#traefikhttproutersrouter-namerule">`traefik/http/routers/<router_name>/rule`</a> | See [rule](../http/router/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)```  |
| <a id="traefikhttproutersrouter-nameruleSyntax" href="#traefikhttproutersrouter-nameruleSyntax" title="#traefikhttproutersrouter-nameruleSyntax">`traefik/http/routers/<router_name>/ruleSyntax`</a> | See [rule](../http/router/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3`  |
| <a id="traefikhttproutersrouter-nameentrypoints0" href="#traefikhttproutersrouter-nameentrypoints0" title="#traefikhttproutersrouter-nameentrypoints0">`traefik/http/routers/<router_name>/entrypoints/0`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `web`       |
| <a id="traefikhttproutersrouter-nameentrypoints1" href="#traefikhttproutersrouter-nameentrypoints1" title="#traefikhttproutersrouter-nameentrypoints1">`traefik/http/routers/<router_name>/entrypoints/1`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `websecure` |
| <a id="traefikhttproutersrouter-namemiddlewares0" href="#traefikhttproutersrouter-namemiddlewares0" title="#traefikhttproutersrouter-namemiddlewares0">`traefik/http/routers/<router_name>/middlewares/0`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth`      |
| <a id="traefikhttproutersrouter-namemiddlewares1" href="#traefikhttproutersrouter-namemiddlewares1" title="#traefikhttproutersrouter-namemiddlewares1">`traefik/http/routers/<router_name>/middlewares/1`</a> |  | `prefix`    |
| <a id="traefikhttproutersrouter-nameservice" href="#traefikhttproutersrouter-nameservice" title="#traefikhttproutersrouter-nameservice">`traefik/http/routers/<router_name>/service`</a> | See [service](../http/load-balancing/service.md) for more information. | `myservice` |
| <a id="traefikhttproutersrouter-nametls" href="#traefikhttproutersrouter-nametls" title="#traefikhttproutersrouter-nametls">`traefik/http/routers/<router_name>/tls`</a> | See [tls](../http/tls/overview.md) for more information. | `true` |
| <a id="traefikhttproutersrouter-nametlscertresolver" href="#traefikhttproutersrouter-nametlscertresolver" title="#traefikhttproutersrouter-nametlscertresolver">`traefik/http/routers/<router_name>/tls/certresolver`</a> | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| <a id="traefikhttproutersrouter-nametlsdomains0main" href="#traefikhttproutersrouter-nametlsdomains0main" title="#traefikhttproutersrouter-nametlsdomains0main">`traefik/http/routers/<router_name>/tls/domains/0/main`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| <a id="traefikhttproutersrouter-nametlsdomains0sans0" href="#traefikhttproutersrouter-nametlsdomains0sans0" title="#traefikhttproutersrouter-nametlsdomains0sans0">`traefik/http/routers/<router_name>/tls/domains/0/sans/0`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `test.example.org` |
| <a id="traefikhttproutersrouter-nametlsdomains0sans1" href="#traefikhttproutersrouter-nametlsdomains0sans1" title="#traefikhttproutersrouter-nametlsdomains0sans1">`traefik/http/routers/<router_name>/tls/domains/0/sans/1`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `dev.example.org`  |
| <a id="traefikhttproutersrouter-nametlsoptions" href="#traefikhttproutersrouter-nametlsoptions" title="#traefikhttproutersrouter-nametlsoptions">`traefik/http/routers/<router_name>/tls/options`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` |
| <a id="traefikhttproutersrouter-nameobservabilityaccesslogs" href="#traefikhttproutersrouter-nameobservabilityaccesslogs" title="#traefikhttproutersrouter-nameobservabilityaccesslogs">`traefik/http/routers/<router_name>/observability/accesslogs`</a> | The accessLogs option controls whether the router will produce access-logs. | `true` |
| <a id="traefikhttproutersrouter-nameobservabilitymetrics" href="#traefikhttproutersrouter-nameobservabilitymetrics" title="#traefikhttproutersrouter-nameobservabilitymetrics">`traefik/http/routers/<router_name>/observability/metrics`</a> | The metrics option controls whether the router will produce metrics. | `true` |
| <a id="traefikhttproutersrouter-nameobservabilitytracing" href="#traefikhttproutersrouter-nameobservabilitytracing" title="#traefikhttproutersrouter-nameobservabilitytracing">`traefik/http/routers/<router_name>/observability/tracing`</a> | The tracing option controls whether the router will produce traces. | `true` |
| <a id="traefikhttproutersrouter-namepriority" href="#traefikhttproutersrouter-namepriority" title="#traefikhttproutersrouter-namepriority">`traefik/http/routers/<router_name>/priority`</a> | See [priority](../http/router/rules-and-priority.md#priority-calculation) for more information. | `42`  |

#### Services

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="traefikhttpservicesmyserviceloadbalancerservers0url" href="#traefikhttpservicesmyserviceloadbalancerservers0url" title="#traefikhttpservicesmyserviceloadbalancerservers0url">`traefik/http/services/myservice/loadbalancer/servers/0/url`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `http://<ip-server-1>:<port-server-1>/` |
| <a id="traefikhttpservicesmyserviceloadbalancerservers0preservePath" href="#traefikhttpservicesmyserviceloadbalancerservers0preservePath" title="#traefikhttpservicesmyserviceloadbalancerservers0preservePath">`traefik/http/services/myservice/loadbalancer/servers/0/preservePath`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `true` |
| <a id="traefikhttpservicesmyserviceloadbalancerservers0weight" href="#traefikhttpservicesmyserviceloadbalancerservers0weight" title="#traefikhttpservicesmyserviceloadbalancerservers0weight">`traefik/http/services/myservice/loadbalancer/servers/0/weight`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `1` |
| <a id="traefikhttpservicesmyserviceloadbalancerserverstransport" href="#traefikhttpservicesmyserviceloadbalancerserverstransport" title="#traefikhttpservicesmyserviceloadbalancerserverstransport">`traefik/http/services/myservice/loadbalancer/serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/> See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| <a id="traefikhttpservicesmyserviceloadbalancerpasshostheader" href="#traefikhttpservicesmyserviceloadbalancerpasshostheader" title="#traefikhttpservicesmyserviceloadbalancerpasshostheader">`traefik/http/services/myservice/loadbalancer/passhostheader`</a> | See [Service](../http/load-balancing/service.md) for more information. | `true` |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo">`traefik/http/services/myservice/loadbalancer/healthcheck/headers/X-Foo`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckhostname" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckhostname" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckhostname">`traefik/http/services/myservice/loadbalancer/healthcheck/hostname`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckinterval" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckinterval" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckinterval">`traefik/http/services/myservice/loadbalancer/healthcheck/interval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckpath" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckpath" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckpath">`traefik/http/services/myservice/loadbalancer/healthcheck/path`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckmethod" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckmethod" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckmethod">`traefik/http/services/myservice/loadbalancer/healthcheck/method`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckstatus" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckstatus" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckstatus">`traefik/http/services/myservice/loadbalancer/healthcheck/status`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckport" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckport" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckport">`traefik/http/services/myservice/loadbalancer/healthcheck/port`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthcheckscheme" href="#traefikhttpservicesmyserviceloadbalancerhealthcheckscheme" title="#traefikhttpservicesmyserviceloadbalancerhealthcheckscheme">`traefik/http/services/myservice/loadbalancer/healthcheck/scheme`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| <a id="traefikhttpservicesmyserviceloadbalancerhealthchecktimeout" href="#traefikhttpservicesmyserviceloadbalancerhealthchecktimeout" title="#traefikhttpservicesmyserviceloadbalancerhealthchecktimeout">`traefik/http/services/myservice/loadbalancer/healthcheck/timeout`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| <a id="traefikhttpservicesmyserviceloadbalancersticky" href="#traefikhttpservicesmyserviceloadbalancersticky" title="#traefikhttpservicesmyserviceloadbalancersticky">`traefik/http/services/myservice/loadbalancer/sticky`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="traefikhttpservicesmyserviceloadbalancerstickycookiehttponly" href="#traefikhttpservicesmyserviceloadbalancerstickycookiehttponly" title="#traefikhttpservicesmyserviceloadbalancerstickycookiehttponly">`traefik/http/services/myservice/loadbalancer/sticky/cookie/httponly`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="traefikhttpservicesmyserviceloadbalancerstickycookiename" href="#traefikhttpservicesmyserviceloadbalancerstickycookiename" title="#traefikhttpservicesmyserviceloadbalancerstickycookiename">`traefik/http/services/myservice/loadbalancer/sticky/cookie/name`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `foobar` |
| <a id="traefikhttpservicesmyserviceloadbalancerstickycookiepath" href="#traefikhttpservicesmyserviceloadbalancerstickycookiepath" title="#traefikhttpservicesmyserviceloadbalancerstickycookiepath">`traefik/http/services/myservice/loadbalancer/sticky/cookie/path`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `/foobar` |
| <a id="traefikhttpservicesmyserviceloadbalancerstickycookiesecure" href="#traefikhttpservicesmyserviceloadbalancerstickycookiesecure" title="#traefikhttpservicesmyserviceloadbalancerstickycookiesecure">`traefik/http/services/myservice/loadbalancer/sticky/cookie/secure`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="traefikhttpservicesmyserviceloadbalancerstickycookiesamesite" href="#traefikhttpservicesmyserviceloadbalancerstickycookiesamesite" title="#traefikhttpservicesmyserviceloadbalancerstickycookiesamesite">`traefik/http/services/myservice/loadbalancer/sticky/cookie/samesite`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `none` |
| <a id="traefikhttpservicesmyserviceloadbalancerstickycookiemaxage" href="#traefikhttpservicesmyserviceloadbalancerstickycookiemaxage" title="#traefikhttpservicesmyserviceloadbalancerstickycookiemaxage">`traefik/http/services/myservice/loadbalancer/sticky/cookie/maxage`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `42`  |
| <a id="traefikhttpservicesmyserviceloadbalancerresponseforwardingflushinterval" href="#traefikhttpservicesmyserviceloadbalancerresponseforwardingflushinterval" title="#traefikhttpservicesmyserviceloadbalancerresponseforwardingflushinterval">`traefik/http/services/myservice/loadbalancer/responseforwarding/flushinterval`</a> | See [Service](../http/load-balancing/service.md) for more information. | `10`  |
| <a id="traefikhttpservicesservice-namemirroringservice" href="#traefikhttpservicesservice-namemirroringservice" title="#traefikhttpservicesservice-namemirroringservice">`traefik/http/services/<service_name>/mirroring/service`</a> | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| <a id="traefikhttpservicesservice-namemirroringmirrorsnname" href="#traefikhttpservicesservice-namemirroringmirrorsnname" title="#traefikhttpservicesservice-namemirroringmirrorsnname">`traefik/http/services/<service_name>/mirroring/mirrors/<n>/name`</a> | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| <a id="traefikhttpservicesservice-namemirroringmirrorsnpercent" href="#traefikhttpservicesservice-namemirroringmirrorsnpercent" title="#traefikhttpservicesservice-namemirroringmirrorsnpercent">`traefik/http/services/<service_name>/mirroring/mirrors/<n>/percent`</a> | See [Service](../http/load-balancing/service.md#mirroring)for more information. | `42`  |
| <a id="traefikhttpservicesservice-nameweightedservicesnname" href="#traefikhttpservicesservice-nameweightedservicesnname" title="#traefikhttpservicesservice-nameweightedservicesnname">`traefik/http/services/<service_name>/weighted/services/<n>/name`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| <a id="traefikhttpservicesservice-nameweightedservicesnweight" href="#traefikhttpservicesservice-nameweightedservicesnweight" title="#traefikhttpservicesservice-nameweightedservicesnweight">`traefik/http/services/<service_name>/weighted/services/<n>/weight`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| <a id="traefikhttpservicesservice-nameweightedstickycookiename" href="#traefikhttpservicesservice-nameweightedstickycookiename" title="#traefikhttpservicesservice-nameweightedstickycookiename">`traefik/http/services/<service_name>/weighted/sticky/cookie/name`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| <a id="traefikhttpservicesservice-nameweightedstickycookiesecure" href="#traefikhttpservicesservice-nameweightedstickycookiesecure" title="#traefikhttpservicesservice-nameweightedstickycookiesecure">`traefik/http/services/<service_name>/weighted/sticky/cookie/secure`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| <a id="traefikhttpservicesservice-nameweightedstickycookiesamesite" href="#traefikhttpservicesservice-nameweightedstickycookiesamesite" title="#traefikhttpservicesservice-nameweightedstickycookiesamesite">`traefik/http/services/<service_name>/weighted/sticky/cookie/samesite`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `none` |
| <a id="traefikhttpservicesservice-nameweightedstickycookiehttpOnly" href="#traefikhttpservicesservice-nameweightedstickycookiehttpOnly" title="#traefikhttpservicesservice-nameweightedstickycookiehttpOnly">`traefik/http/services/<service_name>/weighted/sticky/cookie/httpOnly`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| <a id="traefikhttpservicesservice-nameweightedstickycookiemaxage" href="#traefikhttpservicesservice-nameweightedstickycookiemaxage" title="#traefikhttpservicesservice-nameweightedstickycookiemaxage">`traefik/http/services/<service_name>/weighted/sticky/cookie/maxage`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| <a id="traefikhttpservicesservice-namefailoverfallback" href="#traefikhttpservicesservice-namefailoverfallback" title="#traefikhttpservicesservice-namefailoverfallback">`traefik/http/services/<service_name>/failover/fallback`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `backup`  |
| <a id="traefikhttpservicesservice-namefailoverhealthcheck" href="#traefikhttpservicesservice-namefailoverhealthcheck" title="#traefikhttpservicesservice-namefailoverhealthcheck">`traefik/http/services/<service_name>/failover/healthcheck`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `{}`  |
| <a id="traefikhttpservicesservice-namefailoverservice" href="#traefikhttpservicesservice-namefailoverservice" title="#traefikhttpservicesservice-namefailoverservice">`traefik/http/services/<service_name>/failover/service`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `main`  |

#### Middleware

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="traefikhttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" href="#traefikhttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" title="#traefikhttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option">`traefik/http/middlewares/mymiddleware/middleware_type/middleware_option`</a> | With `middleware_type` the type of middleware (ex: `forwardAuth`, `headers`, etc)<br/>and `middleware_option` the middleware option to set (ex for the middleware `addPrefix`: `prefix`).<br/> More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md). | `foobar` |

!!! warning "The character `@` is not authorized in the middleware name."

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

##### Configuration Example
    
```bash
# Declaring a middleware
traefik/http/middlewares/myAddPrefix/addPrefix/prefix=/foobar
# Referencing a middleware
traefik/http/routers/<router_name>/middlewares/0=myAddPrefix
```

#### ServerTransport

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="traefikhttpserversTransportsserversTransportNamest-option" href="#traefikhttpserversTransportsserversTransportNamest-option" title="#traefikhttpserversTransportsserversTransportNamest-option">`traefik/http/serversTransports/<serversTransportName>/st_option`</a> | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../http/load-balancing/serverstransport.md). | ServerTransport Options |

##### Configuration Example
    
```bash
# Declaring a ServerTransport
traefik/http/serversTransports/myServerTransport/maxIdleConnsPerHost=-1
traefik/http/serversTransports/myServerTransport/certificates/0/certFile=mypath/cert.pem
traefik/http/serversTransports/myServerTransport/certificates/0/keyFile=mypath/key.pem
# Referencing a middleware
traefik/http/services/myService/serversTransports/0=myServerTransport
```

### TCP

You can declare TCP Routers and/or Services using KV.

#### Routers

| Key (Path)                                      |  Description | Value |
|-------------------------------------------------|-------------------------------------------------|-------|
| <a id="traefiktcproutersmytcprouterentrypoints0" href="#traefiktcproutersmytcprouterentrypoints0" title="#traefiktcproutersmytcprouterentrypoints0">`traefik/tcp/routers/mytcprouter/entrypoints/0`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1` |
| <a id="traefiktcproutersmytcprouterentrypoints1" href="#traefiktcproutersmytcprouterentrypoints1" title="#traefiktcproutersmytcprouterentrypoints1">`traefik/tcp/routers/mytcprouter/entrypoints/1`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep2` |
| <a id="traefiktcproutersmy-routerrule" href="#traefiktcproutersmy-routerrule" title="#traefiktcproutersmy-routerrule">`traefik/tcp/routers/my-router/rule`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | ```HostSNI(`example.com`)``` |
| <a id="traefiktcproutersmytcprouterservice" href="#traefiktcproutersmytcprouterservice" title="#traefiktcproutersmytcprouterservice">`traefik/tcp/routers/mytcprouter/service`</a> | See [service](../tcp/service.md) for more information. | `myservice` |
| <a id="traefiktcproutersmytcproutertls" href="#traefiktcproutersmytcproutertls" title="#traefiktcproutersmytcproutertls">`traefik/tcp/routers/mytcprouter/tls`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="traefiktcproutersmytcproutertlscertresolver" href="#traefiktcproutersmytcproutertlscertresolver" title="#traefiktcproutersmytcproutertlscertresolver">`traefik/tcp/routers/mytcprouter/tls/certresolver`</a> | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| <a id="traefiktcproutersmytcproutertlsdomains0main" href="#traefiktcproutersmytcproutertlsdomains0main" title="#traefiktcproutersmytcproutertlsdomains0main">`traefik/tcp/routers/mytcprouter/tls/domains/0/main`</a> | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| <a id="traefiktcproutersmytcproutertlsdomains0sans0" href="#traefiktcproutersmytcproutertlsdomains0sans0" title="#traefiktcproutersmytcproutertlsdomains0sans0">`traefik/tcp/routers/mytcprouter/tls/domains/0/sans/0`</a> | See [TLS](../tcp/tls.md) for more information. | `test.example.org` |
| <a id="traefiktcproutersmytcproutertlsdomains0sans1" href="#traefiktcproutersmytcproutertlsdomains0sans1" title="#traefiktcproutersmytcproutertlsdomains0sans1">`traefik/tcp/routers/mytcprouter/tls/domains/0/sans/1`</a> | See [TLS](../tcp/tls.md) for more information. | `dev.example.org`  |
| <a id="traefiktcproutersmytcproutertlsoptions" href="#traefiktcproutersmytcproutertlsoptions" title="#traefiktcproutersmytcproutertlsoptions">`traefik/tcp/routers/mytcprouter/tls/options`</a> | See [TLS](../tcp/tls.md) for more information. | `foobar` |
| <a id="traefiktcproutersmytcproutertlspassthrough" href="#traefiktcproutersmytcproutertlspassthrough" title="#traefiktcproutersmytcproutertlspassthrough">`traefik/tcp/routers/mytcprouter/tls/passthrough`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="traefiktcproutersmytcprouterpriority" href="#traefiktcproutersmytcprouterpriority" title="#traefiktcproutersmytcprouterpriority">`traefik/tcp/routers/mytcprouter/priority`</a> | See [priority](../tcp/router/rules-and-priority.md#priority) for more information. | `42`  |

#### Services

| Key (Path)                                                         | Description                                                         | Value            |
|--------------------------------------------------------------------|--------------------------------------------------------------------|------------------|
| <a id="traefiktcpservicesmytcpserviceloadbalancerservers0address" href="#traefiktcpservicesmytcpserviceloadbalancerservers0address" title="#traefiktcpservicesmytcpserviceloadbalancerservers0address">`traefik/tcp/services/mytcpservice/loadbalancer/servers/0/address`</a> | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `xx.xx.xx.xx:xx` |
| <a id="traefiktcpservicesmytcpserviceloadbalancerservers0tls" href="#traefiktcpservicesmytcpserviceloadbalancerservers0tls" title="#traefiktcpservicesmytcpserviceloadbalancerservers0tls">`traefik/tcp/services/mytcpservice/loadbalancer/servers/0/tls`</a> | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `true` |
| <a id="traefiktcpservicesmytcpserviceloadbalancerproxyprotocolversion" href="#traefiktcpservicesmytcpserviceloadbalancerproxyprotocolversion" title="#traefiktcpservicesmytcpserviceloadbalancerproxyprotocolversion">`traefik/tcp/services/mytcpservice/loadbalancer/proxyprotocol/version`</a> | See [PROXY protocol](../tcp/service.md#proxy-protocol) for more information. | `1`   |
| <a id="traefiktcpservicesmyserviceloadbalancerserverstransport" href="#traefiktcpservicesmyserviceloadbalancerserverstransport" title="#traefiktcpservicesmyserviceloadbalancerserverstransport">`traefik/tcp/services/myservice/loadbalancer/serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |
| <a id="traefiktcpservicesservice-nameweightedservices0name" href="#traefiktcpservicesservice-nameweightedservices0name" title="#traefiktcpservicesservice-nameweightedservices0name">`traefik/tcp/services/<service_name>/weighted/services/0/name`</a> | See [Service](../tcp/service.md#weighted-round-robin) for more information. | `foobar` |
| <a id="traefiktcpservicesservice-nameweightedservices0weight" href="#traefiktcpservicesservice-nameweightedservices0weight" title="#traefiktcpservicesservice-nameweightedservices0weight">`traefik/tcp/services/<service_name>/weighted/services/0/weight`</a> | See [Service](../tcp/service.md#weighted-round-robin-wrr) for more information. | `42`  |

#### Middleware

##### Configuration Options

You can declare pieces of middleware using tags starting with `traefik/tcp/middlewares/{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`InFlightConn`](../tcp/middlewares/inflightconn.md) named `test-inflightconn`, you'd write `traefik/tcp/middlewares/test-inflightconn/inflightconn/amount=10`.

More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md).

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="traefiktcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" href="#traefiktcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" title="#traefiktcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option">`traefik/tcp/middlewares/mymiddleware/middleware_type/middleware_option`</a> | With `middleware_type` the type of middleware (ex: `inflightconn`)<br/>and `middleware_option` the middleware option to set (ex for the middleware `inflightconn`: `amount`).<br/> More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md). | `foobar` |

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

##### Configuration Example
    
```bash
# Declaring a middleware
traefik/tcp/middlewares/test-inflightconn/amount=10
# Referencing a middleware
traefik/tcp/routers/<router_name>/middlewares/0=test-inflightconn
```

#### ServerTransport

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="traefiktcpserversTransportsserversTransportNamest-option" href="#traefiktcpserversTransportsserversTransportNamest-option" title="#traefiktcpserversTransportsserversTransportNamest-option">`traefik/tcp/serversTransports/<serversTransportName>/st_option`</a> | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../tcp/serverstransport.md). | ServerTransport Options |

##### Configuration Example
    
```bash
# Declaring a ServerTransport
traefik/tcp/serversTransports/myServerTransport/maxIdleConnsPerHost=-1
# Referencing a middleware
traefik/tcp/services/myService/serversTransports/0=myServerTransport
```

### UDP

You can declare UDP Routers and/or Services using KV.

#### Routers

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| <a id="traefikudproutersmyudprouterentrypoints0" href="#traefikudproutersmyudprouterentrypoints0" title="#traefikudproutersmyudprouterentrypoints0">`traefik/udp/routers/myudprouter/entrypoints/0`</a> | See [UDP Router](../udp/router/rules-priority.md#entrypoints) for more information. | `foobar`  |
| <a id="traefikudproutersmyudprouterservice" href="#traefikudproutersmyudprouterservice" title="#traefikudproutersmyudprouterservice">`traefik/udp/routers/myudprouter/service`</a> | See [UDP Router](../udp/router/rules-priority.md#configuration-example) for more information. | `foobar`  |

#### Services

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| <a id="traefikudpservicesloadBalancerserversnaddress" href="#traefikudpservicesloadBalancerserversnaddress" title="#traefikudpservicesloadBalancerserversnaddress">`traefik/udp/services/loadBalancer/servers/<n>/address`</a> | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| <a id="traefikudpservicesweightedservices0name" href="#traefikudpservicesweightedservices0name" title="#traefikudpservicesweightedservices0name">`traefik/udp/services/weighted/services/0/name`</a> | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| <a id="traefikudpservicesweightedservers0weight" href="#traefikudpservicesweightedservers0weight" title="#traefikudpservicesweightedservers0weight">`traefik/udp/services/weighted/servers/0/weight`</a> |See [UDP Service](../udp/service.md) for more information. | `42`  | 

## TLS

### TLS Options

With the KV provider, you configure some parameters of the TLS connection using the `tls/options` key.

For example, you can define a basic setup like this:

| Key (Path)                                           | Description                                           | Value    |
|------------------------------------------------------|------------------------------------------------------|----------|
| <a id="traefiktlsoptionsOptions0alpnProtocols0" href="#traefiktlsoptionsOptions0alpnProtocols0" title="#traefiktlsoptionsOptions0alpnProtocols0">`traefik/tls/options/Options0/alpnProtocols/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="traefiktlsoptionsOptions0cipherSuites0" href="#traefiktlsoptionsOptions0cipherSuites0" title="#traefiktlsoptionsOptions0cipherSuites0">`traefik/tls/options/Options0/cipherSuites/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="traefiktlsoptionsOptions0clientAuthcaFiles0" href="#traefiktlsoptionsOptions0clientAuthcaFiles0" title="#traefiktlsoptionsOptions0clientAuthcaFiles0">`traefik/tls/options/Options0/clientAuth/caFiles/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="traefiktlsoptionsOptions0disableSessiontickets" href="#traefiktlsoptionsOptions0disableSessiontickets" title="#traefiktlsoptionsOptions0disableSessiontickets">`traefik/tls/options/Options0/disableSessiontickets`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. |  `true`   |

### TLS Default Generated Certificates

You can configure Traefik to use an ACME provider (like Let's Encrypt) to generate the default certificate.

The configuration to resolve the default certificate should be defined in a TLS store.

| Key (Path)                                                     | Description                                                     | Value    |
|----------------------------------------------------------------|----------------------------------------------------------------|----------|
| <a id="traefiktlsstoresStore0defaultGeneratedCertdomainmain" href="#traefiktlsstoresStore0defaultGeneratedCertdomainmain" title="#traefiktlsstoresStore0defaultGeneratedCertdomainmain">`traefik/tls/stores/Store0/defaultGeneratedCert/domain/main`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information. | `foobar` |
| <a id="traefiktlsstoresStore0defaultGeneratedCertdomainsans0" href="#traefiktlsstoresStore0defaultGeneratedCertdomainsans0" title="#traefiktlsstoresStore0defaultGeneratedCertdomainsans0">`traefik/tls/stores/Store0/defaultGeneratedCert/domain/sans/0`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
| <a id="traefiktlsstoresStore0defaultGeneratedCertdomainsans1" href="#traefiktlsstoresStore0defaultGeneratedCertdomainsans1" title="#traefiktlsstoresStore0defaultGeneratedCertdomainsans1">`traefik/tls/stores/Store0/defaultGeneratedCert/domain/sans/1`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
| <a id="traefiktlsstoresStore0defaultGeneratedCertresolver" href="#traefiktlsstoresStore0defaultGeneratedCertresolver" title="#traefiktlsstoresStore0defaultGeneratedCertresolver">`traefik/tls/stores/Store0/defaultGeneratedCert/resolver`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
