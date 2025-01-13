---
title: "Traefik Routing Configuration with KV stores"
description: "Read the technical documentation to learn the Traefik Routing Configuration with KV stores."
---

# Traefik & KV Stores

## Routing Configuration

!!! info "Keys"

    Keys are case-insensitive.

### Routers

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

??? info "`traefik/http/routers/<router_name>/rule`"

    See [rule](../http/router/rules-and-priority.md#rules) for more information.
    
    | Key (Path)                           | Value                      |
    |--------------------------------------|----------------------------|
    | `traefik/http/routers/myrouter/rule` | ```Host(`example.com`)```  |

??? info "`traefik/http/routers/<router_name>/entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    | Key (Path)                                    | Value       |
    |-----------------------------------------------|-------------|
    | `traefik/http/routers/myrouter/entrypoints/0` | `web`       |
    | `traefik/http/routers/myrouter/entrypoints/1` | `websecure` |

??? info "`traefik/http/routers/<router_name>/middlewares`"

    See [middlewares overview](../http/middlewares/overview.md) for more information.

    | Key (Path)                                    | Value       |
    |-----------------------------------------------|-------------|
    | `traefik/http/routers/myrouter/middlewares/0` | `auth`      |
    | `traefik/http/routers/myrouter/middlewares/1` | `prefix`    |
    | `traefik/http/routers/myrouter/middlewares/2` | `cb`        |

??? info "`traefik/http/routers/<router_name>/service`"

    See [service](../http/load-balancing/service.md) for more information.

    | Key (Path)                              | Value       |
    |-----------------------------------------|-------------|
    | `traefik/http/routers/myrouter/service` | `myservice` |

??? info "`traefik/http/routers/<router_name>/tls`"

    See [tls](../http/tls/overview.md) for more information.

    | Key (Path)                          | Value  |
    |-------------------------------------|--------|
    | `traefik/http/routers/myrouter/tls` | `true` |
    
??? info "`traefik/http/routers/<router_name>/tls/certresolver`"

    See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information.

    | Key (Path)                                       | Value        |
    |--------------------------------------------------|--------------|
    | `traefik/http/routers/myrouter/tls/certresolver` | `myresolver` |    

??? info "`traefik/http/routers/<router_name>/tls/domains/<n>/main`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    | Key (Path)                                         | Value         |
    |----------------------------------------------------|---------------|
    | `traefik/http/routers/myrouter/tls/domains/0/main` | `example.org` |
    
??? info "`traefik/http/routers/<router_name>/tls/domains/<n>/sans/<n>`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    | Key (Path)                                           | Value              |
    |------------------------------------------------------|--------------------|
    | `traefik/http/routers/myrouter/tls/domains/0/sans/0` | `test.example.org` |
    | `traefik/http/routers/myrouter/tls/domains/0/sans/1` | `dev.example.org`  |
    
??? info "`traefik/http/routers/<router_name>/tls/options`"

    See [TLS](../http/tls/overview.md) for more information.

    | Key (Path)                                  | Value    |
    |---------------------------------------------|----------|
    | `traefik/http/routers/myrouter/tls/options` | `foobar` |

??? info "`traefik/http/routers/<router_name>/priority`"

    See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information.

    | Key (Path)                               | Value |
    |------------------------------------------|-------|
    | `traefik/http/routers/myrouter/priority` | `42`  |

### Services

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

??? info "`traefik/http/services/<service_name>/loadbalancer/servers/<n>/url`"

    See [servers](../http/load-balancing/service.md#servers) for more information.

    | Key (Path)                                                      | Value                                   |
    |-----------------------------------------------------------------|-----------------------------------------|
    | `traefik/http/services/myservice/loadbalancer/servers/0/url`    | `http://<ip-server-1>:<port-server-1>/` |

??? info "`traefik/http/services/<service_name>/loadbalancer/serverstransport`"

    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../http/load-balancing/serverstransport.md) for more information.

    | Key (Path)                                                      | Value         |
    |-----------------------------------------------------------------|---------------|
    | `traefik/http/services/myservice/loadbalancer/serverstransport` | `foobar@file` |

??? info "`traefik/http/services/<service_name>/loadbalancer/passhostheader`"

    | Key (Path)                                                      | Value  |
    |-----------------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/passhostheader`   | `true` |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/headers/<header_name>`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                               | Value    |
    |--------------------------------------------------------------------------|----------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/headers/X-Foo` | `foobar` |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/hostname`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                          | Value         |
    |---------------------------------------------------------------------|---------------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/hostname` | `example.org` |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/interval`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                          | Value |
    |---------------------------------------------------------------------|-------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/interval` | `10`  |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/path`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                      | Value  |
    |-----------------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/path` | `/foo` |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/method`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                        | Value    |
    |-------------------------------------------------------------------|----------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/method` | `foobar` |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/status`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                        | Value |
    |-------------------------------------------------------------------|-------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/status` | `42`  |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/port`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                      | Value |
    |-----------------------------------------------------------------|-------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/port` | `42`  |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/scheme`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                        | Value  |
    |-------------------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/scheme` | `http` |

??? info "`traefik/http/services/<service_name>/loadbalancer/healthcheck/timeout`"

    See [health check](../http/load-balancing/service.md#health-check) for more information.

    | Key (Path)                                                         | Value |
    |--------------------------------------------------------------------|-------|
    | `traefik/http/services/myservice/loadbalancer/healthcheck/timeout` | `10`  |

??? info "`traefik/http/services/<service_name>/loadbalancer/sticky`"

    | Key (Path)                                            | Value  |
    |-------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/sticky` | `true` |

??? info "`traefik/http/services/<service_name>/loadbalancer/sticky/cookie/httponly`"

    | Key (Path)                                                            | Value  |
    |-----------------------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/sticky/cookie/httponly` | `true` |

??? info "`traefik/http/services/<service_name>/loadbalancer/sticky/cookie/name`"

    | Key (Path)                                                        | Value    |
    |-------------------------------------------------------------------|----------|
    | `traefik/http/services/myservice/loadbalancer/sticky/cookie/name` | `foobar` |

??? info "`traefik/http/services/<service_name>/loadbalancer/sticky/cookie/secure`"

    | Key (Path)                                                          | Value  |
    |---------------------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/sticky/cookie/secure` | `true` |

??? info "`traefik/http/services/<service_name>/loadbalancer/sticky/cookie/samesite`"

    | Key (Path)                                                            | Value  |
    |-----------------------------------------------------------------------|--------|
    | `traefik/http/services/myservice/loadbalancer/sticky/cookie/samesite` | `none` |

??? info "`traefik/http/services/<service_name>/loadbalancer/sticky/cookie/maxage`"

    | Key (Path)                                                          | Value |
    |---------------------------------------------------------------------|-------|
    | `traefik/http/services/myservice/loadbalancer/sticky/cookie/maxage` | `42`  |

??? info "`traefik/http/services/<service_name>/loadbalancer/responseforwarding/flushinterval`"

    | Key (Path)                                                                      | Value |
    |---------------------------------------------------------------------------------|-------|
    | `traefik/http/services/myservice/loadbalancer/responseforwarding/flushinterval` | `10`  |

??? info "`traefik/http/services/<service_name>/mirroring/service`"

    | Key (Path)                                               | Value    |
    |----------------------------------------------------------|----------|
    | `traefik/http/services/<service_name>/mirroring/service` | `foobar` |

??? info "`traefik/http/services/<service_name>/mirroring/mirrors/<n>/name`"

    | Key (Path)                                                        | Value    |
    |-------------------------------------------------------------------|----------|
    | `traefik/http/services/<service_name>/mirroring/mirrors/<n>/name` | `foobar` |

??? info "`traefik/http/services/<service_name>/mirroring/mirrors/<n>/percent`"

    | Key (Path)                                                           | Value |
    |----------------------------------------------------------------------|-------|
    | `traefik/http/services/<service_name>/mirroring/mirrors/<n>/percent` | `42`  |

??? info "`traefik/http/services/<service_name>/weighted/services/<n>/name`"

    | Key (Path)                                                        | Value    |
    |-------------------------------------------------------------------|----------|
    | `traefik/http/services/<service_name>/weighted/services/<n>/name` | `foobar` |

??? info "`traefik/http/services/<service_name>/weighted/services/<n>/weight`"

    | Key (Path)                                                          | Value |
    |---------------------------------------------------------------------|-------|
    | `traefik/http/services/<service_name>/weighted/services/<n>/weight` | `42`  |

??? info "`traefik/http/services/<service_name>/weighted/sticky/cookie/name`"

    | Key (Path)                                                         | Value    |
    |--------------------------------------------------------------------|----------|
    | `traefik/http/services/<service_name>/weighted/sticky/cookie/name` | `foobar` |

??? info "`traefik/http/services/<service_name>/weighted/sticky/cookie/secure`"

    | Key (Path)                                                           | Value  |
    |----------------------------------------------------------------------|--------|
    | `traefik/http/services/<service_name>/weighted/sticky/cookie/secure` | `true` |

??? info "`traefik/http/services/<service_name>/weighted/sticky/cookie/samesite`"

    | Key (Path)                                                             | Value  |
    |------------------------------------------------------------------------|--------|
    | `traefik/http/services/<service_name>/weighted/sticky/cookie/samesite` | `none` |

??? info "`traefik/http/services/<service_name>/weighted/sticky/cookie/httpOnly`"

    | Key (Path)                                                             | Value  |
    |------------------------------------------------------------------------|--------|
    | `traefik/http/services/<service_name>/weighted/sticky/cookie/httpOnly` | `true` |

??? info "`traefik/http/services/<service_name>/weighted/sticky/cookie/maxage`"

    | Key (Path)                                                           | Value |
    |----------------------------------------------------------------------|-------|
    | `traefik/http/services/<service_name>/weighted/sticky/cookie/maxage` | `42`  |

### Middleware

More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md).

!!! warning "The character `@` is not authorized in the middleware name."

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### TCP

You can declare TCP Routers and/or Services using KV.

#### TCP Routers

??? info "`traefik/tcp/routers/<router_name>/entrypoints`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    | Key (Path)                                      | Value |
    |-------------------------------------------------|-------|
    | `traefik/tcp/routers/mytcprouter/entrypoints/0` | `ep1` |
    | `traefik/tcp/routers/mytcprouter/entrypoints/1` | `ep2` |
    
??? info "`traefik/tcp/routers/<router_name>/rule`"

    See [entry points](../../install-configuration/entrypoints.md) for more information.

    | Key (Path)                           | Value                        |
    |--------------------------------------|------------------------------|
    | `traefik/tcp/routers/my-router/rule` | ```HostSNI(`example.com`)``` |  

??? info "`traefik/tcp/routers/<router_name>/service`"

    See [service](../tcp/service.md) for more information.
    
    | Key (Path)                                | Value       |
    |-------------------------------------------|-------------|
    | `traefik/tcp/routers/mytcprouter/service` | `myservice` |

??? info "`traefik/tcp/routers/<router_name>/tls`"

    See [TLS](../tcp/tls.md) for more information.

    | Key (Path)                            | Value  |
    |---------------------------------------|--------|
    | `traefik/tcp/routers/mytcprouter/tls` | `true` |

??? info "`traefik/tcp/routers/<router_name>/tls/certresolver`"

    See [certResolver](../tcp/tls.md#configuration-options) for more information.

    | Key (Path)                                         | Value        |
    |----------------------------------------------------|--------------|
    | `traefik/tcp/routers/mytcprouter/tls/certresolver` | `myresolver` |

??? info "`traefik/tcp/routers/<router_name>/tls/domains/<n>/main`"

    See [TLS](../tcp/tls.md) for more information.

    | Key (Path)                                           | Value         |
    |------------------------------------------------------|---------------|
    | `traefik/tcp/routers/mytcprouter/tls/domains/0/main` | `example.org` |
        
??? info "`traefik/tcp/routers/<router_name>/tls/domains/<n>/sans`"

    See [TLS](../tcp/tls.md) for more information.

    | Key (Path)                                             | Value              |
    |--------------------------------------------------------|--------------------|
    | `traefik/tcp/routers/mytcprouter/tls/domains/0/sans/0` | `test.example.org` |
    | `traefik/tcp/routers/mytcprouter/tls/domains/0/sans/1` | `dev.example.org`  |
    
??? info "`traefik/tcp/routers/<router_name>/tls/options`"

    See [TLS](../tcp/tls.md) for more information.

    | Key (Path)                                    | Value    |
    |-----------------------------------------------|----------|
    | `traefik/tcp/routers/mytcprouter/tls/options` | `foobar` |

??? info "`traefik/tcp/routers/<router_name>/tls/passthrough`"

    See [TLS](../tcp/tls.md) for more information.

    | Key (Path)                                        | Value  |
    |---------------------------------------------------|--------|
    | `traefik/tcp/routers/mytcprouter/tls/passthrough` | `true` |

??? info "`traefik/tcp/routers/<router_name>/priority`"

    See [priority](../tcp/router/rules-and-priority.md#priority) for more information.

    | Key (Path)                               | Value |
    |------------------------------------------|-------|
    | `traefik/tcp/routers/myrouter/priority`  | `42`  |

#### TCP Services

??? info "`traefik/tcp/services/<service_name>/loadbalancer/servers/<n>/address`"

    See [servers](../tcp/service.md#servers-load-balancer) for more information.

    | Key (Path)                                                         | Value            |
    |--------------------------------------------------------------------|------------------|
    | `traefik/tcp/services/mytcpservice/loadbalancer/servers/0/address` | `xx.xx.xx.xx:xx` |
    
??? info "`traefik/tcp/services/<service_name>/loadbalancer/proxyprotocol/version`"

    See [PROXY protocol](../tcp/service.md#proxy-protocol) for more information.

    | Key (Path)                                                             | Value |
    |------------------------------------------------------------------------|-------|
    | `traefik/tcp/services/mytcpservice/loadbalancer/proxyprotocol/version` | `1`   |

??? info "`traefik/tcp/services/<service_name>/loadbalancer/serverstransport`"

    Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.
    See [serverstransport](../tcp/serverstransport.md) for more information.

    | Key (Path)                                                      | Value         |
    |-----------------------------------------------------------------|---------------|
    | `traefik/tcp/services/myservice/loadbalancer/serverstransport` | `foobar@file` |

??? info "`traefik/tcp/services/<service_name>/weighted/services/<n>/name`"

    | Key (Path)                                                          | Value    |
    |---------------------------------------------------------------------|----------|
    | `traefik/tcp/services/<service_name>/weighted/services/0/name`      | `foobar` |

??? info "`traefik/tcp/services/<service_name>/weighted/services/<n>/weight`"

    | Key (Path)                                                       | Value |
    |------------------------------------------------------------------|-------|
    | `traefik/tcp/services/<service_name>/weighted/services/0/weight` | `42`  |
