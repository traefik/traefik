---
title: "Baqup Routing Configuration with KV stores"
description: "Read the technical documentation to learn the Baqup Routing Configuration with KV stores."
---

# Baqup & KV Stores

## Configuration Options

!!! info "Keys"

    Keys are case-insensitive.

### HTTP

#### Routers

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

| Key (Path)                           | Description                          | Value                      |
|--------------------------------------|--------------------------------------|----------------------------|
| <a id="opt-baquphttproutersrouter-namerule" href="#opt-baquphttproutersrouter-namerule" title="#opt-baquphttproutersrouter-namerule">`baqup/http/routers/<router_name>/rule`</a> | See [rule](../http/routing/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)```  |
| <a id="opt-baquphttproutersrouter-nameruleSyntax" href="#opt-baquphttproutersrouter-nameruleSyntax" title="#opt-baquphttproutersrouter-nameruleSyntax">`baqup/http/routers/<router_name>/ruleSyntax`</a> | See [rule](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3`  |
| <a id="opt-baquphttproutersrouter-nameentrypoints0" href="#opt-baquphttproutersrouter-nameentrypoints0" title="#opt-baquphttproutersrouter-nameentrypoints0">`baqup/http/routers/<router_name>/entrypoints/0`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `web`       |
| <a id="opt-baquphttproutersrouter-nameentrypoints1" href="#opt-baquphttproutersrouter-nameentrypoints1" title="#opt-baquphttproutersrouter-nameentrypoints1">`baqup/http/routers/<router_name>/entrypoints/1`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `websecure` |
| <a id="opt-baquphttproutersrouter-namemiddlewares0" href="#opt-baquphttproutersrouter-namemiddlewares0" title="#opt-baquphttproutersrouter-namemiddlewares0">`baqup/http/routers/<router_name>/middlewares/0`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth`      |
| <a id="opt-baquphttproutersrouter-namemiddlewares1" href="#opt-baquphttproutersrouter-namemiddlewares1" title="#opt-baquphttproutersrouter-namemiddlewares1">`baqup/http/routers/<router_name>/middlewares/1`</a> |  | `prefix`    |
| <a id="opt-baquphttproutersrouter-nameservice" href="#opt-baquphttproutersrouter-nameservice" title="#opt-baquphttproutersrouter-nameservice">`baqup/http/routers/<router_name>/service`</a> | See [service](../http/load-balancing/service.md) for more information. | `myservice` |
| <a id="opt-baquphttproutersrouter-nametls" href="#opt-baquphttproutersrouter-nametls" title="#opt-baquphttproutersrouter-nametls">`baqup/http/routers/<router_name>/tls`</a> | See [tls](../http/tls/overview.md) for more information. | `true` |
| <a id="opt-baquphttproutersrouter-nametlscertresolver" href="#opt-baquphttproutersrouter-nametlscertresolver" title="#opt-baquphttproutersrouter-nametlscertresolver">`baqup/http/routers/<router_name>/tls/certresolver`</a> | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| <a id="opt-baquphttproutersrouter-nametlsdomains0main" href="#opt-baquphttproutersrouter-nametlsdomains0main" title="#opt-baquphttproutersrouter-nametlsdomains0main">`baqup/http/routers/<router_name>/tls/domains/0/main`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| <a id="opt-baquphttproutersrouter-nametlsdomains0sans0" href="#opt-baquphttproutersrouter-nametlsdomains0sans0" title="#opt-baquphttproutersrouter-nametlsdomains0sans0">`baqup/http/routers/<router_name>/tls/domains/0/sans/0`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `test.example.org` |
| <a id="opt-baquphttproutersrouter-nametlsdomains0sans1" href="#opt-baquphttproutersrouter-nametlsdomains0sans1" title="#opt-baquphttproutersrouter-nametlsdomains0sans1">`baqup/http/routers/<router_name>/tls/domains/0/sans/1`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `dev.example.org`  |
| <a id="opt-baquphttproutersrouter-nametlsoptions" href="#opt-baquphttproutersrouter-nametlsoptions" title="#opt-baquphttproutersrouter-nametlsoptions">`baqup/http/routers/<router_name>/tls/options`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` |
| <a id="opt-baquphttproutersrouter-nameobservabilityaccesslogs" href="#opt-baquphttproutersrouter-nameobservabilityaccesslogs" title="#opt-baquphttproutersrouter-nameobservabilityaccesslogs">`baqup/http/routers/<router_name>/observability/accesslogs`</a> | The accessLogs option controls whether the router will produce access-logs. | `true` |
| <a id="opt-baquphttproutersrouter-nameobservabilitymetrics" href="#opt-baquphttproutersrouter-nameobservabilitymetrics" title="#opt-baquphttproutersrouter-nameobservabilitymetrics">`baqup/http/routers/<router_name>/observability/metrics`</a> | The metrics option controls whether the router will produce metrics. | `true` |
| <a id="opt-baquphttproutersrouter-nameobservabilitytracing" href="#opt-baquphttproutersrouter-nameobservabilitytracing" title="#opt-baquphttproutersrouter-nameobservabilitytracing">`baqup/http/routers/<router_name>/observability/tracing`</a> | The tracing option controls whether the router will produce traces. | `true` |
| <a id="opt-baquphttproutersrouter-namepriority" href="#opt-baquphttproutersrouter-namepriority" title="#opt-baquphttproutersrouter-namepriority">`baqup/http/routers/<router_name>/priority`</a> | See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information. | `42`  |

#### Services

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-baquphttpservicesmyserviceloadbalancerservers0url" href="#opt-baquphttpservicesmyserviceloadbalancerservers0url" title="#opt-baquphttpservicesmyserviceloadbalancerservers0url">`baqup/http/services/myservice/loadbalancer/servers/0/url`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `http://<ip-server-1>:<port-server-1>/` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerservers0preservePath" href="#opt-baquphttpservicesmyserviceloadbalancerservers0preservePath" title="#opt-baquphttpservicesmyserviceloadbalancerservers0preservePath">`baqup/http/services/myservice/loadbalancer/servers/0/preservePath`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `true` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerservers0weight" href="#opt-baquphttpservicesmyserviceloadbalancerservers0weight" title="#opt-baquphttpservicesmyserviceloadbalancerservers0weight">`baqup/http/services/myservice/loadbalancer/servers/0/weight`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `1` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerserverstransport" href="#opt-baquphttpservicesmyserviceloadbalancerserverstransport" title="#opt-baquphttpservicesmyserviceloadbalancerserverstransport">`baqup/http/services/myservice/loadbalancer/serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/> See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerpasshostheader" href="#opt-baquphttpservicesmyserviceloadbalancerpasshostheader" title="#opt-baquphttpservicesmyserviceloadbalancerpasshostheader">`baqup/http/services/myservice/loadbalancer/passhostheader`</a> | See [Service](../http/load-balancing/service.md) for more information. | `true` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckheadersX-Foo">`baqup/http/services/myservice/loadbalancer/healthcheck/headers/X-Foo`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckhostname" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckhostname" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckhostname">`baqup/http/services/myservice/loadbalancer/healthcheck/hostname`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckinterval" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckinterval" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckinterval">`baqup/http/services/myservice/loadbalancer/healthcheck/interval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckpath" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckpath" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckpath">`baqup/http/services/myservice/loadbalancer/healthcheck/path`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckmethod" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckmethod" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckmethod">`baqup/http/services/myservice/loadbalancer/healthcheck/method`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckstatus" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckstatus" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckstatus">`baqup/http/services/myservice/loadbalancer/healthcheck/status`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckport" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckport" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckport">`baqup/http/services/myservice/loadbalancer/healthcheck/port`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthcheckscheme" href="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckscheme" title="#opt-baquphttpservicesmyserviceloadbalancerhealthcheckscheme">`baqup/http/services/myservice/loadbalancer/healthcheck/scheme`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerhealthchecktimeout" href="#opt-baquphttpservicesmyserviceloadbalancerhealthchecktimeout" title="#opt-baquphttpservicesmyserviceloadbalancerhealthchecktimeout">`baqup/http/services/myservice/loadbalancer/healthcheck/timeout`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| <a id="opt-baquphttpservicesmyserviceloadbalancersticky" href="#opt-baquphttpservicesmyserviceloadbalancersticky" title="#opt-baquphttpservicesmyserviceloadbalancersticky">`baqup/http/services/myservice/loadbalancer/sticky`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerstickycookiehttponly" href="#opt-baquphttpservicesmyserviceloadbalancerstickycookiehttponly" title="#opt-baquphttpservicesmyserviceloadbalancerstickycookiehttponly">`baqup/http/services/myservice/loadbalancer/sticky/cookie/httponly`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerstickycookiename" href="#opt-baquphttpservicesmyserviceloadbalancerstickycookiename" title="#opt-baquphttpservicesmyserviceloadbalancerstickycookiename">`baqup/http/services/myservice/loadbalancer/sticky/cookie/name`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `foobar` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerstickycookiepath" href="#opt-baquphttpservicesmyserviceloadbalancerstickycookiepath" title="#opt-baquphttpservicesmyserviceloadbalancerstickycookiepath">`baqup/http/services/myservice/loadbalancer/sticky/cookie/path`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `/foobar` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerstickycookiesecure" href="#opt-baquphttpservicesmyserviceloadbalancerstickycookiesecure" title="#opt-baquphttpservicesmyserviceloadbalancerstickycookiesecure">`baqup/http/services/myservice/loadbalancer/sticky/cookie/secure`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerstickycookiesamesite" href="#opt-baquphttpservicesmyserviceloadbalancerstickycookiesamesite" title="#opt-baquphttpservicesmyserviceloadbalancerstickycookiesamesite">`baqup/http/services/myservice/loadbalancer/sticky/cookie/samesite`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `none` |
| <a id="opt-baquphttpservicesmyserviceloadbalancerstickycookiemaxage" href="#opt-baquphttpservicesmyserviceloadbalancerstickycookiemaxage" title="#opt-baquphttpservicesmyserviceloadbalancerstickycookiemaxage">`baqup/http/services/myservice/loadbalancer/sticky/cookie/maxage`</a> | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `42`  |
| <a id="opt-baquphttpservicesmyserviceloadbalancerresponseforwardingflushinterval" href="#opt-baquphttpservicesmyserviceloadbalancerresponseforwardingflushinterval" title="#opt-baquphttpservicesmyserviceloadbalancerresponseforwardingflushinterval">`baqup/http/services/myservice/loadbalancer/responseforwarding/flushinterval`</a> | See [Service](../http/load-balancing/service.md) for more information. | `10`  |
| <a id="opt-baquphttpservicesservice-namemirroringservice" href="#opt-baquphttpservicesservice-namemirroringservice" title="#opt-baquphttpservicesservice-namemirroringservice">`baqup/http/services/<service_name>/mirroring/service`</a> | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| <a id="opt-baquphttpservicesservice-namemirroringmirrorsnname" href="#opt-baquphttpservicesservice-namemirroringmirrorsnname" title="#opt-baquphttpservicesservice-namemirroringmirrorsnname">`baqup/http/services/<service_name>/mirroring/mirrors/<n>/name`</a> | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| <a id="opt-baquphttpservicesservice-namemirroringmirrorsnpercent" href="#opt-baquphttpservicesservice-namemirroringmirrorsnpercent" title="#opt-baquphttpservicesservice-namemirroringmirrorsnpercent">`baqup/http/services/<service_name>/mirroring/mirrors/<n>/percent`</a> | See [Service](../http/load-balancing/service.md#mirroring)for more information. | `42`  |
| <a id="opt-baquphttpservicesservice-nameweightedservicesnname" href="#opt-baquphttpservicesservice-nameweightedservicesnname" title="#opt-baquphttpservicesservice-nameweightedservicesnname">`baqup/http/services/<service_name>/weighted/services/<n>/name`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| <a id="opt-baquphttpservicesservice-nameweightedservicesnweight" href="#opt-baquphttpservicesservice-nameweightedservicesnweight" title="#opt-baquphttpservicesservice-nameweightedservicesnweight">`baqup/http/services/<service_name>/weighted/services/<n>/weight`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| <a id="opt-baquphttpservicesservice-nameweightedstickycookiename" href="#opt-baquphttpservicesservice-nameweightedstickycookiename" title="#opt-baquphttpservicesservice-nameweightedstickycookiename">`baqup/http/services/<service_name>/weighted/sticky/cookie/name`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| <a id="opt-baquphttpservicesservice-nameweightedstickycookiesecure" href="#opt-baquphttpservicesservice-nameweightedstickycookiesecure" title="#opt-baquphttpservicesservice-nameweightedstickycookiesecure">`baqup/http/services/<service_name>/weighted/sticky/cookie/secure`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| <a id="opt-baquphttpservicesservice-nameweightedstickycookiesamesite" href="#opt-baquphttpservicesservice-nameweightedstickycookiesamesite" title="#opt-baquphttpservicesservice-nameweightedstickycookiesamesite">`baqup/http/services/<service_name>/weighted/sticky/cookie/samesite`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `none` |
| <a id="opt-baquphttpservicesservice-nameweightedstickycookiehttpOnly" href="#opt-baquphttpservicesservice-nameweightedstickycookiehttpOnly" title="#opt-baquphttpservicesservice-nameweightedstickycookiehttpOnly">`baqup/http/services/<service_name>/weighted/sticky/cookie/httpOnly`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| <a id="opt-baquphttpservicesservice-nameweightedstickycookiemaxage" href="#opt-baquphttpservicesservice-nameweightedstickycookiemaxage" title="#opt-baquphttpservicesservice-nameweightedstickycookiemaxage">`baqup/http/services/<service_name>/weighted/sticky/cookie/maxage`</a> | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| <a id="opt-baquphttpservicesservice-namefailoverfallback" href="#opt-baquphttpservicesservice-namefailoverfallback" title="#opt-baquphttpservicesservice-namefailoverfallback">`baqup/http/services/<service_name>/failover/fallback`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `backup`  |
| <a id="opt-baquphttpservicesservice-namefailoverhealthcheck" href="#opt-baquphttpservicesservice-namefailoverhealthcheck" title="#opt-baquphttpservicesservice-namefailoverhealthcheck">`baqup/http/services/<service_name>/failover/healthcheck`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `{}`  |
| <a id="opt-baquphttpservicesservice-namefailoverservice" href="#opt-baquphttpservicesservice-namefailoverservice" title="#opt-baquphttpservicesservice-namefailoverservice">`baqup/http/services/<service_name>/failover/service`</a> | See [Failover](../http/load-balancing/service.md#failover) for more information. | `main`  |

#### Middleware

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-baquphttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" href="#opt-baquphttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" title="#opt-baquphttpmiddlewaresmymiddlewaremiddleware-typemiddleware-option">`baqup/http/middlewares/mymiddleware/middleware_type/middleware_option`</a> | With `middleware_type` the type of middleware (ex: `forwardAuth`, `headers`, etc)<br/>and `middleware_option` the middleware option to set (ex for the middleware `addPrefix`: `prefix`).<br/> More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md). | `foobar` |

!!! warning "The character `@` is not authorized in the middleware name."

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

##### Configuration Example
    
```bash
# Declaring a middleware
baqup/http/middlewares/myAddPrefix/addPrefix/prefix=/foobar
# Referencing a middleware
baqup/http/routers/<router_name>/middlewares/0=myAddPrefix
```

#### ServerTransport

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-baquphttpserversTransportsserversTransportNamest-option" href="#opt-baquphttpserversTransportsserversTransportNamest-option" title="#opt-baquphttpserversTransportsserversTransportNamest-option">`baqup/http/serversTransports/<serversTransportName>/st_option`</a> | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../http/load-balancing/serverstransport.md). | ServerTransport Options |

##### Configuration Example
    
```bash
# Declaring a ServerTransport
baqup/http/serversTransports/myServerTransport/maxIdleConnsPerHost=-1
baqup/http/serversTransports/myServerTransport/certificates/0/certFile=mypath/cert.pem
baqup/http/serversTransports/myServerTransport/certificates/0/keyFile=mypath/key.pem
# Referencing a middleware
baqup/http/services/myService/serversTransports/0=myServerTransport
```

### TCP

You can declare TCP Routers and/or Services using KV.

#### Routers

| Key (Path)                                      |  Description | Value |
|-------------------------------------------------|-------------------------------------------------|-------|
| <a id="opt-baquptcproutersmytcprouterentrypoints0" href="#opt-baquptcproutersmytcprouterentrypoints0" title="#opt-baquptcproutersmytcprouterentrypoints0">`baqup/tcp/routers/mytcprouter/entrypoints/0`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1` |
| <a id="opt-baquptcproutersmytcprouterentrypoints1" href="#opt-baquptcproutersmytcprouterentrypoints1" title="#opt-baquptcproutersmytcprouterentrypoints1">`baqup/tcp/routers/mytcprouter/entrypoints/1`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep2` |
| <a id="opt-baquptcproutersmy-routerrule" href="#opt-baquptcproutersmy-routerrule" title="#opt-baquptcproutersmy-routerrule">`baqup/tcp/routers/my-router/rule`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | ```HostSNI(`example.com`)``` |
| <a id="opt-baquptcproutersmytcprouterservice" href="#opt-baquptcproutersmytcprouterservice" title="#opt-baquptcproutersmytcprouterservice">`baqup/tcp/routers/mytcprouter/service`</a> | See [service](../tcp/service.md) for more information. | `myservice` |
| <a id="opt-baquptcproutersmytcproutertls" href="#opt-baquptcproutersmytcproutertls" title="#opt-baquptcproutersmytcproutertls">`baqup/tcp/routers/mytcprouter/tls`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="opt-baquptcproutersmytcproutertlscertresolver" href="#opt-baquptcproutersmytcproutertlscertresolver" title="#opt-baquptcproutersmytcproutertlscertresolver">`baqup/tcp/routers/mytcprouter/tls/certresolver`</a> | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| <a id="opt-baquptcproutersmytcproutertlsdomains0main" href="#opt-baquptcproutersmytcproutertlsdomains0main" title="#opt-baquptcproutersmytcproutertlsdomains0main">`baqup/tcp/routers/mytcprouter/tls/domains/0/main`</a> | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| <a id="opt-baquptcproutersmytcproutertlsdomains0sans0" href="#opt-baquptcproutersmytcproutertlsdomains0sans0" title="#opt-baquptcproutersmytcproutertlsdomains0sans0">`baqup/tcp/routers/mytcprouter/tls/domains/0/sans/0`</a> | See [TLS](../tcp/tls.md) for more information. | `test.example.org` |
| <a id="opt-baquptcproutersmytcproutertlsdomains0sans1" href="#opt-baquptcproutersmytcproutertlsdomains0sans1" title="#opt-baquptcproutersmytcproutertlsdomains0sans1">`baqup/tcp/routers/mytcprouter/tls/domains/0/sans/1`</a> | See [TLS](../tcp/tls.md) for more information. | `dev.example.org`  |
| <a id="opt-baquptcproutersmytcproutertlsoptions" href="#opt-baquptcproutersmytcproutertlsoptions" title="#opt-baquptcproutersmytcproutertlsoptions">`baqup/tcp/routers/mytcprouter/tls/options`</a> | See [TLS](../tcp/tls.md) for more information. | `foobar` |
| <a id="opt-baquptcproutersmytcproutertlspassthrough" href="#opt-baquptcproutersmytcproutertlspassthrough" title="#opt-baquptcproutersmytcproutertlspassthrough">`baqup/tcp/routers/mytcprouter/tls/passthrough`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="opt-baquptcproutersmytcprouterpriority" href="#opt-baquptcproutersmytcprouterpriority" title="#opt-baquptcproutersmytcprouterpriority">`baqup/tcp/routers/mytcprouter/priority`</a> | See [priority](../tcp/routing/rules-and-priority.md#priority-calculation) for more information. | `42`  |

#### Services

| Key (Path)                                                         | Description                                                         | Value            |
|--------------------------------------------------------------------|--------------------------------------------------------------------|------------------|
| <a id="opt-baquptcpservicesmytcpserviceloadbalancerservers0address" href="#opt-baquptcpservicesmytcpserviceloadbalancerservers0address" title="#opt-baquptcpservicesmytcpserviceloadbalancerservers0address">`baqup/tcp/services/mytcpservice/loadbalancer/servers/0/address`</a> | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `xx.xx.xx.xx:xx` |
| <a id="opt-baquptcpservicesmytcpserviceloadbalancerservers0tls" href="#opt-baquptcpservicesmytcpserviceloadbalancerservers0tls" title="#opt-baquptcpservicesmytcpserviceloadbalancerservers0tls">`baqup/tcp/services/mytcpservice/loadbalancer/servers/0/tls`</a> | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `true` |
| <a id="opt-baquptcpservicesmyserviceloadbalancerserverstransport" href="#opt-baquptcpservicesmyserviceloadbalancerserverstransport" title="#opt-baquptcpservicesmyserviceloadbalancerserverstransport">`baqup/tcp/services/myservice/loadbalancer/serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-baquptcpservicesservice-nameweightedservices0name" href="#opt-baquptcpservicesservice-nameweightedservices0name" title="#opt-baquptcpservicesservice-nameweightedservices0name">`baqup/tcp/services/<service_name>/weighted/services/0/name`</a> | See [Service](../tcp/service.md#weighted-round-robin) for more information. | `foobar` |
| <a id="opt-baquptcpservicesservice-nameweightedservices0weight" href="#opt-baquptcpservicesservice-nameweightedservices0weight" title="#opt-baquptcpservicesservice-nameweightedservices0weight">`baqup/tcp/services/<service_name>/weighted/services/0/weight`</a> | See [Service](../tcp/service.md#weighted-round-robin) for more information. | `42`  |

#### Middleware

##### Configuration Options

You can declare pieces of middleware using tags starting with `baqup/tcp/middlewares/{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`InFlightConn`](../tcp/middlewares/inflightconn.md) named `test-inflightconn`, you'd write `baqup/tcp/middlewares/test-inflightconn/inflightconn/amount=10`.

More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md).

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-baquptcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" href="#opt-baquptcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option" title="#opt-baquptcpmiddlewaresmymiddlewaremiddleware-typemiddleware-option">`baqup/tcp/middlewares/mymiddleware/middleware_type/middleware_option`</a> | With `middleware_type` the type of middleware (ex: `inflightconn`)<br/>and `middleware_option` the middleware option to set (ex for the middleware `inflightconn`: `amount`).<br/> More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md). | `foobar` |

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

##### Configuration Example
    
```bash
# Declaring a middleware
baqup/tcp/middlewares/test-inflightconn/amount=10
# Referencing a middleware
baqup/tcp/routers/<router_name>/middlewares/0=test-inflightconn
```

#### ServerTransport

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| <a id="opt-baquptcpserversTransportsserversTransportNamest-option" href="#opt-baquptcpserversTransportsserversTransportNamest-option" title="#opt-baquptcpserversTransportsserversTransportNamest-option">`baqup/tcp/serversTransports/<serversTransportName>/st_option`</a> | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../tcp/serverstransport.md). | ServerTransport Options |

##### Configuration Example
    
```bash
# Declaring a ServerTransport
baqup/tcp/serversTransports/myServerTransport/maxIdleConnsPerHost=-1
# Referencing a middleware
baqup/tcp/services/myService/serversTransports/0=myServerTransport
```

### UDP

You can declare UDP Routers and/or Services using KV.

#### Routers

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| <a id="opt-baqupudproutersmyudprouterentrypoints0" href="#opt-baqupudproutersmyudprouterentrypoints0" title="#opt-baqupudproutersmyudprouterentrypoints0">`baqup/udp/routers/myudprouter/entrypoints/0`</a> | See [UDP Router](../udp/routing/rules-priority.md#entrypoints) for more information. | `foobar`  |
| <a id="opt-baqupudproutersmyudprouterservice" href="#opt-baqupudproutersmyudprouterservice" title="#opt-baqupudproutersmyudprouterservice">`baqup/udp/routers/myudprouter/service`</a> | See [UDP Router](../udp/routing/rules-priority.md#configuration-example) for more information. | `foobar`  |

#### Services

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| <a id="opt-baqupudpservicesloadBalancerserversnaddress" href="#opt-baqupudpservicesloadBalancerserversnaddress" title="#opt-baqupudpservicesloadBalancerserversnaddress">`baqup/udp/services/loadBalancer/servers/<n>/address`</a> | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| <a id="opt-baqupudpservicesweightedservices0name" href="#opt-baqupudpservicesweightedservices0name" title="#opt-baqupudpservicesweightedservices0name">`baqup/udp/services/weighted/services/0/name`</a> | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| <a id="opt-baqupudpservicesweightedservers0weight" href="#opt-baqupudpservicesweightedservers0weight" title="#opt-baqupudpservicesweightedservers0weight">`baqup/udp/services/weighted/servers/0/weight`</a> |See [UDP Service](../udp/service.md) for more information. | `42`  | 

## TLS

### TLS Options

With the KV provider, you configure some parameters of the TLS connection using the `tls/options` key.

For example, you can define a basic setup like this:

| Key (Path)                                           | Description                                           | Value    |
|------------------------------------------------------|------------------------------------------------------|----------|
| <a id="opt-baquptlsoptionsOptions0alpnProtocols0" href="#opt-baquptlsoptionsOptions0alpnProtocols0" title="#opt-baquptlsoptionsOptions0alpnProtocols0">`baqup/tls/options/Options0/alpnProtocols/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="opt-baquptlsoptionsOptions0cipherSuites0" href="#opt-baquptlsoptionsOptions0cipherSuites0" title="#opt-baquptlsoptionsOptions0cipherSuites0">`baqup/tls/options/Options0/cipherSuites/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="opt-baquptlsoptionsOptions0clientAuthcaFiles0" href="#opt-baquptlsoptionsOptions0clientAuthcaFiles0" title="#opt-baquptlsoptionsOptions0clientAuthcaFiles0">`baqup/tls/options/Options0/clientAuth/caFiles/0`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| <a id="opt-baquptlsoptionsOptions0disableSessiontickets" href="#opt-baquptlsoptionsOptions0disableSessiontickets" title="#opt-baquptlsoptionsOptions0disableSessiontickets">`baqup/tls/options/Options0/disableSessiontickets`</a> | See [TLS Options](../http/tls/tls-options.md) for more information. |  `true`   |

### TLS Default Generated Certificates

You can configure Baqup to use an ACME provider (like Let's Encrypt) to generate the default certificate.

The configuration to resolve the default certificate should be defined in a TLS store.

| Key (Path)                                                     | Description                                                     | Value    |
|----------------------------------------------------------------|----------------------------------------------------------------|----------|
| <a id="opt-baquptlsstoresStore0defaultGeneratedCertdomainmain" href="#opt-baquptlsstoresStore0defaultGeneratedCertdomainmain" title="#opt-baquptlsstoresStore0defaultGeneratedCertdomainmain">`baqup/tls/stores/Store0/defaultGeneratedCert/domain/main`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information. | `foobar` |
| <a id="opt-baquptlsstoresStore0defaultGeneratedCertdomainsans0" href="#opt-baquptlsstoresStore0defaultGeneratedCertdomainsans0" title="#opt-baquptlsstoresStore0defaultGeneratedCertdomainsans0">`baqup/tls/stores/Store0/defaultGeneratedCert/domain/sans/0`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
| <a id="opt-baquptlsstoresStore0defaultGeneratedCertdomainsans1" href="#opt-baquptlsstoresStore0defaultGeneratedCertdomainsans1" title="#opt-baquptlsstoresStore0defaultGeneratedCertdomainsans1">`baqup/tls/stores/Store0/defaultGeneratedCert/domain/sans/1`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
| <a id="opt-baquptlsstoresStore0defaultGeneratedCertresolver" href="#opt-baquptlsstoresStore0defaultGeneratedCertresolver" title="#opt-baquptlsstoresStore0defaultGeneratedCertresolver">`baqup/tls/stores/Store0/defaultGeneratedCert/resolver`</a> | See [TLS](../http/tls/tls-certificates.md#certificates-stores) for more information| `foobar` |
