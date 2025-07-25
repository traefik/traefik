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
| `traefik/http/routers/<router_name>/rule` | See [rule](../http/router/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)```  |
| `traefik/http/routers/<router_name>/ruleSyntax` | See [rule](../http/router/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3`  |
| `traefik/http/routers/<router_name>/entrypoints/0` | See [entry points](../../install-configuration/entrypoints.md) for more information. | `web`       |
| `traefik/http/routers/<router_name>/entrypoints/1` | See [entry points](../../install-configuration/entrypoints.md) for more information. | `websecure` |
| `traefik/http/routers/<router_name>/middlewares/0` | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth`      |
| `traefik/http/routers/<router_name>/middlewares/1` |  | `prefix`    |
| `traefik/http/routers/<router_name>/service` | See [service](../http/load-balancing/service.md) for more information. | `myservice` |
| `traefik/http/routers/<router_name>/tls` | See [tls](../http/tls/overview.md) for more information. | `true` |
| `traefik/http/routers/<router_name>/tls/certresolver` | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| `traefik/http/routers/<router_name>/tls/domains/0/main` | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| `traefik/http/routers/<router_name>/tls/domains/0/sans/0` | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `test.example.org` |
| `traefik/http/routers/<router_name>/tls/domains/0/sans/1` | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `dev.example.org`  |
| `traefik/http/routers/<router_name>/tls/options` | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` |
| `traefik/http/routers/<router_name>/observability/accesslogs` | The accessLogs option controls whether the router will produce access-logs. | `true` |
| `traefik/http/routers/<router_name>/observability/metrics` | The metrics option controls whether the router will produce metrics. | `true` |
| `traefik/http/routers/<router_name>/observability/tracing` | The tracing option controls whether the router will produce traces. | `true` |
| `traefik/http/routers/<router_name>/priority` | See [priority](../http/router/rules-and-priority.md#priority-calculation) for more information. | `42`  |

#### Services

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| `traefik/http/services/myservice/loadbalancer/servers/0/url`    | See [servers](../http/load-balancing/service.md#servers) for more information. | `http://<ip-server-1>:<port-server-1>/` |
| `traefik/http/services/myservice/loadbalancer/servers/0/preservePath`    | See [servers](../http/load-balancing/service.md#servers) for more information. | `true` |
| `traefik/http/services/myservice/loadbalancer/servers/0/weight`    | See [servers](../http/load-balancing/service.md#servers) for more information. | `1` |
| `traefik/http/services/myservice/loadbalancer/serverstransport` | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/> See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| `traefik/http/services/myservice/loadbalancer/passhostheader`   | See [Service](../http/load-balancing/service.md) for more information. | `true` |
| `traefik/http/services/myservice/loadbalancer/healthcheck/headers/X-Foo` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| `traefik/http/services/myservice/loadbalancer/healthcheck/hostname` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| `traefik/http/services/myservice/loadbalancer/healthcheck/interval` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| `traefik/http/services/myservice/loadbalancer/healthcheck/path` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| `traefik/http/services/myservice/loadbalancer/healthcheck/method` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| `traefik/http/services/myservice/loadbalancer/healthcheck/status` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| `traefik/http/services/myservice/loadbalancer/healthcheck/port` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42`  |
| `traefik/http/services/myservice/loadbalancer/healthcheck/scheme` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| `traefik/http/services/myservice/loadbalancer/healthcheck/timeout` | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10`  |
| `traefik/http/services/myservice/loadbalancer/sticky` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| `traefik/http/services/myservice/loadbalancer/sticky/cookie/httponly` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| `traefik/http/services/myservice/loadbalancer/sticky/cookie/name` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `foobar` |
| `traefik/http/services/myservice/loadbalancer/sticky/cookie/path` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `/foobar` |
| `traefik/http/services/myservice/loadbalancer/sticky/cookie/secure` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `true` |
| `traefik/http/services/myservice/loadbalancer/sticky/cookie/samesite` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `none` |
| `traefik/http/services/myservice/loadbalancer/sticky/cookie/maxage` | See [Service](../http/load-balancing/service.md#sticky-sessions) for more information. | `42`  |
| `traefik/http/services/myservice/loadbalancer/responseforwarding/flushinterval` | See [Service](../http/load-balancing/service.md) for more information. | `10`  |
| `traefik/http/services/<service_name>/mirroring/service` | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| `traefik/http/services/<service_name>/mirroring/mirrors/<n>/name` | See [Service](../http/load-balancing/service.md#mirroring) for more information. | `foobar` |
| `traefik/http/services/<service_name>/mirroring/mirrors/<n>/percent` | See [Service](../http/load-balancing/service.md#mirroring)for more information. | `42`  |
| `traefik/http/services/<service_name>/weighted/services/<n>/name` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| `traefik/http/services/<service_name>/weighted/services/<n>/weight` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| `traefik/http/services/<service_name>/weighted/sticky/cookie/name` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `foobar` |
| `traefik/http/services/<service_name>/weighted/sticky/cookie/secure` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| `traefik/http/services/<service_name>/weighted/sticky/cookie/samesite` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `none` |
| `traefik/http/services/<service_name>/weighted/sticky/cookie/httpOnly` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `true` |
| `traefik/http/services/<service_name>/weighted/sticky/cookie/maxage` | See [Service](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `42`  |
| `traefik/http/services/<service_name>/failover/fallback` | See [Failover](../http/load-balancing/service.md#failover) for more information. | `backup`  |
| `traefik/http/services/<service_name>/failover/healthcheck` | See [Failover](../http/load-balancing/service.md#failover) for more information. | `{}`  |
| `traefik/http/services/<service_name>/failover/service` | See [Failover](../http/load-balancing/service.md#failover) for more information. | `main`  |

#### Middleware

##### Configuration Options

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| `traefik/http/middlewares/mymiddleware/middleware_type/middleware_option`    | With `middleware_type` the type of middleware (ex: `forwardAuth`, `headers`, etc)<br/>and `middleware_option` the middleware option to set (ex for the middleware `addPrefix`: `prefix`).<br/> More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md). | `foobar` |

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
| `traefik/http/serversTransports/<serversTransportName>/st_option`    | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../http/load-balancing/serverstransport.md). | ServerTransport Options |

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
| `traefik/tcp/routers/mytcprouter/entrypoints/0` | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1` |
| `traefik/tcp/routers/mytcprouter/entrypoints/1` | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep2` |
| `traefik/tcp/routers/my-router/rule` | See [entry points](../../install-configuration/entrypoints.md) for more information. | ```HostSNI(`example.com`)``` |
| `traefik/tcp/routers/mytcprouter/service` | See [service](../tcp/service.md) for more information. | `myservice` |
| `traefik/tcp/routers/mytcprouter/tls` | See [TLS](../tcp/tls.md) for more information. | `true` |
| `traefik/tcp/routers/mytcprouter/tls/certresolver` | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| `traefik/tcp/routers/mytcprouter/tls/domains/0/main` | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| `traefik/tcp/routers/mytcprouter/tls/domains/0/sans/0` | See [TLS](../tcp/tls.md) for more information. | `test.example.org` |
| `traefik/tcp/routers/mytcprouter/tls/domains/0/sans/1` | See [TLS](../tcp/tls.md) for more information. | `dev.example.org`  |
| `traefik/tcp/routers/mytcprouter/tls/options` | See [TLS](../tcp/tls.md) for more information. | `foobar` |
| `traefik/tcp/routers/mytcprouter/tls/passthrough` | See [TLS](../tcp/tls.md) for more information. | `true` |
| `traefik/tcp/routers/mytcprouter/priority`  | See [priority](../tcp/router/rules-and-priority.md#priority) for more information. | `42`  |

#### Services

| Key (Path)                                                         | Description                                                         | Value            |
|--------------------------------------------------------------------|--------------------------------------------------------------------|------------------|
| `traefik/tcp/services/mytcpservice/loadbalancer/servers/0/address` | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `xx.xx.xx.xx:xx` |
| `traefik/tcp/services/mytcpservice/loadbalancer/servers/0/tls` | See [servers](../tcp/service.md#servers-load-balancer) for more information. | `true` |
| `traefik/tcp/services/mytcpservice/loadbalancer/proxyprotocol/version` | See [PROXY protocol](../tcp/service.md#proxy-protocol) for more information. | `1`   |
| `traefik/tcp/services/myservice/loadbalancer/serverstransport` | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |
| `traefik/tcp/services/<service_name>/weighted/services/0/name`      | See [Service](../tcp/service.md#weighted-round-robin) for more information. | `foobar` |
| `traefik/tcp/services/<service_name>/weighted/services/0/weight` | See [Service](../tcp/service.md#weighted-round-robin-wrr) for more information. | `42`  |

#### Middleware

##### Configuration Options

You can declare pieces of middleware using tags starting with `traefik/tcp/middlewares/{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`InFlightConn`](../tcp/middlewares/inflightconn.md) named `test-inflightconn`, you'd write `traefik/tcp/middlewares/test-inflightconn/inflightconn/amount=10`.

More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md).

| Key (Path)                                                      | Description                                                      |  Value                                   |
|-----------------------------------------------------------------|-----------------------------------------------------------------|-----------------------------------------|
| `traefik/tcp/middlewares/mymiddleware/middleware_type/middleware_option`    | With `middleware_type` the type of middleware (ex: `inflightconn`)<br/>and `middleware_option` the middleware option to set (ex for the middleware `inflightconn`: `amount`).<br/> More information about available middlewares in the dedicated [middlewares section](../tcp/middlewares/overview.md). | `foobar` |

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
| `traefik/tcp/serversTransports/<serversTransportName>/st_option`    | With  `st_option` the ServerTransport option to set (ex `maxIdleConnsPerHost`).<br/> More information about available options in the dedicated [ServerTransport section](../tcp/serverstransport.md). | ServerTransport Options |

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
| `traefik/udp/routers/myudprouter/entrypoints/0` | See [UDP Router](../udp/router/rules-priority.md#entrypoints) for more information. | `foobar`  |
| `traefik/udp/routers/myudprouter/service` | See [UDP Router](../udp/router/rules-priority.md#configuration-example) for more information. | `foobar`  |

#### Services

| Key (Path)                                                       | Description                                                       | Value |
|------------------------------------------------------------------|------------------------------------------------------------------|-------|
| `traefik/udp/services/loadBalancer/servers/<n>/address` | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| `traefik/udp/services/weighted/services/0/name` | See [UDP Service](../udp/service.md) for more information. | `foobar`  |
| `traefik/udp/services/weighted/servers/0/weight` |See [UDP Service](../udp/service.md) for more information. | `42`  | 

## TLS

### TLS Options

With the KV provider, you configure some parameters of the TLS connection using the `tls/options` key.

For example, you can define a basic setup like this:

| Key (Path)                                           | Description                                           | Value    |
|------------------------------------------------------|------------------------------------------------------|----------|
| `traefik/tls/options/Options0/alpnProtocols/0`       | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| `traefik/tls/options/Options0/cipherSuites/0`        | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| `traefik/tls/options/Options0/clientAuth/caFiles/0`  | See [TLS Options](../http/tls/tls-options.md) for more information. | `foobar` | 
| `traefik/tls/options/Options0/disableSessiontickets` | See [TLS Options](../http/tls/tls-options.md) for more information. |  `true`   |

### TLS Default Generated Certificates

You can configure Traefik to use an ACME provider (like Let's Encrypt) to generate the default certificate.

The configuration to resolve the default certificate should be defined in a TLS store.

| Key (Path)                                                     | Description                                                     | Value    |
|----------------------------------------------------------------|----------------------------------------------------------------|----------|
| `traefik/tls/stores/Store0/defaultGeneratedCert/domain/main`   | See [TLS](../http/tls/tls-certificates/#certificates-stores) for more information. | `foobar` |
| `traefik/tls/stores/Store0/defaultGeneratedCert/domain/sans/0` | See [TLS](../http/tls/tls-certificates/#certificates-stores) for more information| `foobar` |
| `traefik/tls/stores/Store0/defaultGeneratedCert/domain/sans/1` | See [TLS](../http/tls/tls-certificates/#certificates-stores) for more information| `foobar` |
| `traefik/tls/stores/Store0/defaultGeneratedCert/resolver`      | See [TLS](../http/tls/tls-certificates/#certificates-stores) for more information| `foobar` |
