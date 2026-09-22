---
title: "Traefik File Routing Configuration"
description: "This guide will provide you with the reference for file-based routing configuration in Traefik Proxy. Read the technical documentation."
---

# Traefik File Routing Configuration

The file provider lets you define routing configuration in YAML or TOML.
Use it to declare routers, services, middlewares, TCP and UDP routing, and TLS options that Traefik should load from a file or a directory.

To configure the file provider itself, see the [File provider install configuration](../../install-configuration/providers/others/file.md) page.

## Configuration Examples

??? example "Configuring the File Provider and Exposing One HTTP Service"

    Enabling the file provider:

    ```yaml tab="Structured (YAML)"
    providers:
      file:
        filename: /etc/traefik/dynamic.yml
    ```

    ```toml tab="Structured (TOML)"
    [providers.file]
      filename = "/etc/traefik/dynamic.toml"
    ```

    ```bash tab="CLI"
    --providers.file.filename=/etc/traefik/dynamic.yml
    ```

    Declaring the dynamic HTTP configuration:

    ```yaml tab="Structured (YAML)"
    http:
      routers:
        app:
          rule: Host(`example.com`)
          entryPoints:
            - websecure
          service: app
          tls: {}

      services:
        app:
          loadBalancer:
            servers:
              - url: http://127.0.0.1:8080
    ```

    ```toml tab="Structured (TOML)"
    [http.routers.app]
      rule = "Host(`example.com`)"
      entryPoints = ["websecure"]
      service = "app"

      [http.routers.app.tls]

    [http.services.app.loadBalancer]
      [[http.services.app.loadBalancer.servers]]
        url = "http://127.0.0.1:8080"
    ```

??? example "Specifying More Than One Router and Service"

    Define each router and explicitly attach it to the service that should handle matching requests.

    ```yaml tab="Structured (YAML)"
    http:
      routers:
        app:
          rule: Host(`example-a.com`)
          service: app
        admin:
          rule: Host(`example-b.com`)
          service: admin

      services:
        app:
          loadBalancer:
            servers:
              - url: http://127.0.0.1:8000
        admin:
          loadBalancer:
            servers:
              - url: http://127.0.0.1:9000
    ```

    ```toml tab="Structured (TOML)"
    [http.routers.app]
      rule = "Host(`example-a.com`)"
      service = "app"

    [http.routers.admin]
      rule = "Host(`example-b.com`)"
      service = "admin"

    [http.services.app.loadBalancer]
      [[http.services.app.loadBalancer.servers]]
        url = "http://127.0.0.1:8000"

    [http.services.admin.loadBalancer]
      [[http.services.admin.loadBalancer.servers]]
        url = "http://127.0.0.1:9000"
    ```

??? example "Declaring and Referencing Middlewares"

    Middlewares declared by the file provider can be used by routers from the file provider or by routers from other providers.
    When another provider references them, use the `@file` provider suffix.

    ```yaml tab="Structured (YAML)"
    http:
      routers:
        app:
          rule: Host(`secure.example.com`)
          entryPoints:
            - websecure
          middlewares:
            - secure-headers
          service: app
          tls:
            options: modern

      middlewares:
        secure-headers:
          headers:
            stsSeconds: 31536000
            forceSTSHeader: true

      services:
        app:
          loadBalancer:
            servers:
              - url: http://127.0.0.1:8080

    tls:
      options:
        modern:
          minVersion: VersionTLS12
          sniStrict: true
    ```

    ```toml tab="Structured (TOML)"
    [http.routers.app]
      rule = "Host(`secure.example.com`)"
      entryPoints = ["websecure"]
      middlewares = ["secure-headers"]
      service = "app"

      [http.routers.app.tls]
        options = "modern"

    [http.middlewares.secure-headers.headers]
      stsSeconds = 31536000
      forceSTSHeader = true

    [http.services.app.loadBalancer]
      [[http.services.app.loadBalancer.servers]]
        url = "http://127.0.0.1:8080"

    [tls.options.modern]
      minVersion = "VersionTLS12"
      sniStrict = true
    ```

??? example "Loading Multiple Dynamic Configuration Files"

    Configure the file provider with a directory when you want to split dynamic configuration across multiple files.

    ```yaml tab="Structured (YAML)"
    providers:
      file:
        directory: /etc/traefik/dynamic
        watch: true
    ```

    ```toml tab="Structured (TOML)"
    [providers.file]
      directory = "/etc/traefik/dynamic"
      watch = true
    ```

    ```bash tab="CLI"
    --providers.file.directory=/etc/traefik/dynamic
    --providers.file.watch=true
    ```

    Example `/etc/traefik/dynamic/http.yml`:

    ```yaml
    http:
      routers:
        app:
          rule: Host(`example.com`)
          service: app

      services:
        app:
          loadBalancer:
            servers:
              - url: http://127.0.0.1:8080
    ```

    Example `/etc/traefik/dynamic/tls.yml`:

    ```yaml
    tls:
      certificates:
        - certFile: /certs/example.crt
          keyFile: /certs/example.key
    ```

## Configuration Options

### General

The file provider does not discover services automatically.
Define every router, service, middleware, and TLS resource explicitly in the routing configuration file.

When another provider references a resource declared by the file provider, append the `@file` provider suffix.
For example, a Docker label can reference a file-provider middleware with `secure-headers@file`.

The examples below use YAML-style field paths.
In TOML, use the equivalent table and array syntax, such as `[http.routers.<router_name>]` and `[[http.services.<service_name>.loadBalancer.servers]]`.

### HTTP

#### Routers

Define HTTP routers under `http.routers.<router_name>`.

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-http-routers-router-name-rule" href="#opt-http-routers-router-name-rule" title="#opt-http-routers-router-name-rule">`http.routers.<router_name>.rule`</a> | See [rule](../http/routing/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)``` |
| <a id="opt-http-routers-router-name-ruleSyntax" href="#opt-http-routers-router-name-ruleSyntax" title="#opt-http-routers-router-name-ruleSyntax">`http.routers.<router_name>.ruleSyntax`</a> | See [ruleSyntax](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax is deprecated and will be removed in the next major version. | `v3` |
| <a id="opt-http-routers-router-name-entryPointsn" href="#opt-http-routers-router-name-entryPointsn" title="#opt-http-routers-router-name-entryPointsn">`http.routers.<router_name>.entryPoints[n]`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `websecure` |
| <a id="opt-http-routers-router-name-middlewaresn" href="#opt-http-routers-router-name-middlewaresn" title="#opt-http-routers-router-name-middlewaresn">`http.routers.<router_name>.middlewares[n]`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `secure-headers` |
| <a id="opt-http-routers-router-name-service" href="#opt-http-routers-router-name-service" title="#opt-http-routers-router-name-service">`http.routers.<router_name>.service`</a> | See [service](../http/load-balancing/service.md) for more information. | `app` |
| <a id="opt-http-routers-router-name-parentRefsn" href="#opt-http-routers-router-name-parentRefsn" title="#opt-http-routers-router-name-parentRefsn">`http.routers.<router_name>.parentRefs[n]`</a> | See [multi-layer routing](../http/routing/multi-layer-routing.md) for more information. | `parent-router@file` |
| <a id="opt-http-routers-router-name-tls" href="#opt-http-routers-router-name-tls" title="#opt-http-routers-router-name-tls">`http.routers.<router_name>.tls`</a> | See [TLS](../http/tls/overview.md) for more information. | `{}` |
| <a id="opt-http-routers-router-name-tls-certResolver" href="#opt-http-routers-router-name-tls-certResolver" title="#opt-http-routers-router-name-tls-certResolver">`http.routers.<router_name>.tls.certResolver`</a> | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| <a id="opt-http-routers-router-name-tls-domainsn-main" href="#opt-http-routers-router-name-tls-domainsn-main" title="#opt-http-routers-router-name-tls-domainsn-main">`http.routers.<router_name>.tls.domains[n].main`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| <a id="opt-http-routers-router-name-tls-domainsn-sansn" href="#opt-http-routers-router-name-tls-domainsn-sansn" title="#opt-http-routers-router-name-tls-domainsn-sansn">`http.routers.<router_name>.tls.domains[n].sans[n]`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `www.example.org` |
| <a id="opt-http-routers-router-name-tls-options" href="#opt-http-routers-router-name-tls-options" title="#opt-http-routers-router-name-tls-options">`http.routers.<router_name>.tls.options`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `modern` |
| <a id="opt-http-routers-router-name-observability-accessLogs" href="#opt-http-routers-router-name-observability-accessLogs" title="#opt-http-routers-router-name-observability-accessLogs">`http.routers.<router_name>.observability.accessLogs`</a> | Enables or disables access logs for the router. | `true` |
| <a id="opt-http-routers-router-name-observability-metrics" href="#opt-http-routers-router-name-observability-metrics" title="#opt-http-routers-router-name-observability-metrics">`http.routers.<router_name>.observability.metrics`</a> | Enables or disables metrics for the router. | `true` |
| <a id="opt-http-routers-router-name-observability-tracing" href="#opt-http-routers-router-name-observability-tracing" title="#opt-http-routers-router-name-observability-tracing">`http.routers.<router_name>.observability.tracing`</a> | Enables or disables tracing for the router. | `true` |
| <a id="opt-http-routers-router-name-observability-traceVerbosity" href="#opt-http-routers-router-name-observability-traceVerbosity" title="#opt-http-routers-router-name-observability-traceVerbosity">`http.routers.<router_name>.observability.traceVerbosity`</a> | See [trace verbosity](../http/routing/observability.md#opt-traceVerbosity) for more information. | `minimal` |
| <a id="opt-http-routers-router-name-priority" href="#opt-http-routers-router-name-priority" title="#opt-http-routers-router-name-priority">`http.routers.<router_name>.priority`</a> | See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information. | `42` |

#### Services

Define HTTP services under `http.services.<service_name>`.

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-http-services-service-name-loadBalancer-serversn-url" href="#opt-http-services-service-name-loadBalancer-serversn-url" title="#opt-http-services-service-name-loadBalancer-serversn-url">`http.services.<service_name>.loadBalancer.servers[n].url`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `http://127.0.0.1:8080` |
| <a id="opt-http-services-service-name-loadBalancer-serversn-weight" href="#opt-http-services-service-name-loadBalancer-serversn-weight" title="#opt-http-services-service-name-loadBalancer-serversn-weight">`http.services.<service_name>.loadBalancer.servers[n].weight`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `1` |
| <a id="opt-http-services-service-name-loadBalancer-serversn-preservePath" href="#opt-http-services-service-name-loadBalancer-serversn-preservePath" title="#opt-http-services-service-name-loadBalancer-serversn-preservePath">`http.services.<service_name>.loadBalancer.servers[n].preservePath`</a> | See [servers](../http/load-balancing/service.md#servers) for more information. | `true` |
| <a id="opt-http-services-service-name-loadBalancer-strategy" href="#opt-http-services-service-name-loadBalancer-strategy" title="#opt-http-services-service-name-loadBalancer-strategy">`http.services.<service_name>.loadBalancer.strategy`</a> | See [load balancing strategies](../http/load-balancing/service.md#load-balancing-strategies) for more information. | `wrr` |
| <a id="opt-http-services-service-name-loadBalancer-passHostHeader" href="#opt-http-services-service-name-loadBalancer-passHostHeader" title="#opt-http-services-service-name-loadBalancer-passHostHeader">`http.services.<service_name>.loadBalancer.passHostHeader`</a> | See [service load balancer](../http/load-balancing/service.md) for more information. | `true` |
| <a id="opt-http-services-service-name-loadBalancer-healthCheck" href="#opt-http-services-service-name-loadBalancer-healthCheck" title="#opt-http-services-service-name-loadBalancer-healthCheck">`http.services.<service_name>.loadBalancer.healthCheck.*`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `path: /health` |
| <a id="opt-http-services-service-name-loadBalancer-passiveHealthCheck" href="#opt-http-services-service-name-loadBalancer-passiveHealthCheck" title="#opt-http-services-service-name-loadBalancer-passiveHealthCheck">`http.services.<service_name>.loadBalancer.passiveHealthCheck.*`</a> | See [passive health check](../http/load-balancing/service.md#passive-health-check) for more information. | `maxFailedAttempts: 3` |
| <a id="opt-http-services-service-name-loadBalancer-sticky-cookie" href="#opt-http-services-service-name-loadBalancer-sticky-cookie" title="#opt-http-services-service-name-loadBalancer-sticky-cookie">`http.services.<service_name>.loadBalancer.sticky.cookie.*`</a> | See [sticky sessions](../http/load-balancing/service.md#sticky-sessions) for more information. | `name: app-cookie` |
| <a id="opt-http-services-service-name-loadBalancer-responseForwarding-flushInterval" href="#opt-http-services-service-name-loadBalancer-responseForwarding-flushInterval" title="#opt-http-services-service-name-loadBalancer-responseForwarding-flushInterval">`http.services.<service_name>.loadBalancer.responseForwarding.flushInterval`</a> | See [service load balancer](../http/load-balancing/service.md) for more information. | `100ms` |
| <a id="opt-http-services-service-name-loadBalancer-serversTransport" href="#opt-http-services-service-name-loadBalancer-serversTransport" title="#opt-http-services-service-name-loadBalancer-serversTransport">`http.services.<service_name>.loadBalancer.serversTransport`</a> | See [ServersTransport](../http/load-balancing/serverstransport.md) for more information. | `secure-transport` |
| <a id="opt-http-services-service-name-weighted-servicesn-name" href="#opt-http-services-service-name-weighted-servicesn-name" title="#opt-http-services-service-name-weighted-servicesn-name">`http.services.<service_name>.weighted.services[n].name`</a> | See [weighted round robin](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `app-v1` |
| <a id="opt-http-services-service-name-weighted-servicesn-weight" href="#opt-http-services-service-name-weighted-servicesn-weight" title="#opt-http-services-service-name-weighted-servicesn-weight">`http.services.<service_name>.weighted.services[n].weight`</a> | See [weighted round robin](../http/load-balancing/service.md#weighted-round-robin-wrr) for more information. | `3` |
| <a id="opt-http-services-service-name-weighted-sticky-cookie" href="#opt-http-services-service-name-weighted-sticky-cookie" title="#opt-http-services-service-name-weighted-sticky-cookie">`http.services.<service_name>.weighted.sticky.cookie.*`</a> | See [sticky sessions](../http/load-balancing/service.md#sticky-sessions) for more information. | `name: app-cookie` |
| <a id="opt-http-services-service-name-weighted-healthCheck" href="#opt-http-services-service-name-weighted-healthCheck" title="#opt-http-services-service-name-weighted-healthCheck">`http.services.<service_name>.weighted.healthCheck`</a> | See [weighted service health check](../http/load-balancing/service.md#health-check) for more information. | `{}` |
| <a id="opt-http-services-service-name-highestRandomWeight-servicesn-name" href="#opt-http-services-service-name-highestRandomWeight-servicesn-name" title="#opt-http-services-service-name-highestRandomWeight-servicesn-name">`http.services.<service_name>.highestRandomWeight.services[n].name`</a> | See [highest random weight](../http/load-balancing/service.md#highest-random-weight) for more information. | `app-v1` |
| <a id="opt-http-services-service-name-highestRandomWeight-servicesn-weight" href="#opt-http-services-service-name-highestRandomWeight-servicesn-weight" title="#opt-http-services-service-name-highestRandomWeight-servicesn-weight">`http.services.<service_name>.highestRandomWeight.services[n].weight`</a> | See [highest random weight](../http/load-balancing/service.md#highest-random-weight) for more information. | `3` |
| <a id="opt-http-services-service-name-highestRandomWeight-healthCheck" href="#opt-http-services-service-name-highestRandomWeight-healthCheck" title="#opt-http-services-service-name-highestRandomWeight-healthCheck">`http.services.<service_name>.highestRandomWeight.healthCheck`</a> | See [highest random weight](../http/load-balancing/service.md#highest-random-weight) for more information. | `{}` |
| <a id="opt-http-services-service-name-mirroring-service" href="#opt-http-services-service-name-mirroring-service" title="#opt-http-services-service-name-mirroring-service">`http.services.<service_name>.mirroring.service`</a> | See [mirroring](../http/load-balancing/service.md#mirroring) for more information. | `app-main` |
| <a id="opt-http-services-service-name-mirroring-mirrorBody" href="#opt-http-services-service-name-mirroring-mirrorBody" title="#opt-http-services-service-name-mirroring-mirrorBody">`http.services.<service_name>.mirroring.mirrorBody`</a> | See [mirroring](../http/load-balancing/service.md#mirroring) for more information. | `true` |
| <a id="opt-http-services-service-name-mirroring-maxBodySize" href="#opt-http-services-service-name-mirroring-maxBodySize" title="#opt-http-services-service-name-mirroring-maxBodySize">`http.services.<service_name>.mirroring.maxBodySize`</a> | See [mirroring](../http/load-balancing/service.md#mirroring) for more information. | `1048576` |
| <a id="opt-http-services-service-name-mirroring-mirrorsn-name" href="#opt-http-services-service-name-mirroring-mirrorsn-name" title="#opt-http-services-service-name-mirroring-mirrorsn-name">`http.services.<service_name>.mirroring.mirrors[n].name`</a> | See [mirroring](../http/load-balancing/service.md#mirroring) for more information. | `app-shadow` |
| <a id="opt-http-services-service-name-mirroring-mirrorsn-percent" href="#opt-http-services-service-name-mirroring-mirrorsn-percent" title="#opt-http-services-service-name-mirroring-mirrorsn-percent">`http.services.<service_name>.mirroring.mirrors[n].percent`</a> | See [mirroring](../http/load-balancing/service.md#mirroring) for more information. | `10` |
| <a id="opt-http-services-service-name-mirroring-healthCheck" href="#opt-http-services-service-name-mirroring-healthCheck" title="#opt-http-services-service-name-mirroring-healthCheck">`http.services.<service_name>.mirroring.healthCheck`</a> | See [mirroring](../http/load-balancing/service.md#mirroring) for more information. | `{}` |
| <a id="opt-http-services-service-name-failover-service" href="#opt-http-services-service-name-failover-service" title="#opt-http-services-service-name-failover-service">`http.services.<service_name>.failover.service`</a> | See [failover](../http/load-balancing/service.md#failover) for more information. | `app-main` |
| <a id="opt-http-services-service-name-failover-fallback" href="#opt-http-services-service-name-failover-fallback" title="#opt-http-services-service-name-failover-fallback">`http.services.<service_name>.failover.fallback`</a> | See [failover](../http/load-balancing/service.md#failover) for more information. | `app-backup` |
| <a id="opt-http-services-service-name-failover-healthCheck" href="#opt-http-services-service-name-failover-healthCheck" title="#opt-http-services-service-name-failover-healthCheck">`http.services.<service_name>.failover.healthCheck`</a> | See [failover](../http/load-balancing/service.md#failover) for more information. | `{}` |

#### Middlewares

Define HTTP middlewares under `http.middlewares.<middleware_name>`.

For example, to declare an [`AddPrefix`](../http/middlewares/addprefix.md) middleware named `add-api`, set `http.middlewares.add-api.addPrefix.prefix=/api`.

More information about available middlewares can be found in the dedicated [middlewares section](../http/middlewares/overview.md).

!!! warning "The character `@` is not authorized in the middleware name `<middleware_name>`."

!!! warning "Conflicts in Declaration"

    If you declare multiple middlewares with the same name but different parameters, the middleware fails to be declared.

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-http-middlewares-middleware-name-middleware-type-middleware-option" href="#opt-http-middlewares-middleware-name-middleware-type-middleware-option" title="#opt-http-middlewares-middleware-name-middleware-type-middleware-option">`http.middlewares.<middleware_name>.<middleware_type>.<middleware_option>`</a> | With `middleware_type` the middleware type, such as `addPrefix` or `headers`, and `middleware_option` the option to set. | `prefix: /api` |

#### ServersTransports

Define HTTP ServersTransports under `http.serversTransports.<servers_transport_name>`.

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-http-serversTransports-servers-transport-name" href="#opt-http-serversTransports-servers-transport-name" title="#opt-http-serversTransports-servers-transport-name">`http.serversTransports.<servers_transport_name>.*`</a> | See [ServersTransport](../http/load-balancing/serverstransport.md) for more information. | `serverName: example.org` |

### TCP

You can declare TCP routers, services, middlewares, and ServersTransports with the file provider.

#### TCP Routers

Define TCP routers under `tcp.routers.<router_name>`.

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tcp-routers-router-name-entryPointsn" href="#opt-tcp-routers-router-name-entryPointsn" title="#opt-tcp-routers-router-name-entryPointsn">`tcp.routers.<router_name>.entryPoints[n]`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `websecure` |
| <a id="opt-tcp-routers-router-name-rule" href="#opt-tcp-routers-router-name-rule" title="#opt-tcp-routers-router-name-rule">`tcp.routers.<router_name>.rule`</a> | See [rule](../tcp/routing/rules-and-priority.md#rules) for more information. | ```HostSNI(`example.com`)``` |
| <a id="opt-tcp-routers-router-name-ruleSyntax" href="#opt-tcp-routers-router-name-ruleSyntax" title="#opt-tcp-routers-router-name-ruleSyntax">`tcp.routers.<router_name>.ruleSyntax`</a> | Configures the rule syntax to use for parsing the rule on a per-router basis.<br/>RuleSyntax is deprecated and will be removed in the next major version. | `v3` |
| <a id="opt-tcp-routers-router-name-middlewaresn" href="#opt-tcp-routers-router-name-middlewaresn" title="#opt-tcp-routers-router-name-middlewaresn">`tcp.routers.<router_name>.middlewares[n]`</a> | See [TCP middlewares overview](../tcp/middlewares/overview.md) for more information. | `ip-allowlist` |
| <a id="opt-tcp-routers-router-name-service" href="#opt-tcp-routers-router-name-service" title="#opt-tcp-routers-router-name-service">`tcp.routers.<router_name>.service`</a> | See [service](../tcp/service.md) for more information. | `tcp-app` |
| <a id="opt-tcp-routers-router-name-tls" href="#opt-tcp-routers-router-name-tls" title="#opt-tcp-routers-router-name-tls">`tcp.routers.<router_name>.tls`</a> | See [TLS](../tcp/tls.md) for more information. | `{}` |
| <a id="opt-tcp-routers-router-name-tls-certResolver" href="#opt-tcp-routers-router-name-tls-certResolver" title="#opt-tcp-routers-router-name-tls-certResolver">`tcp.routers.<router_name>.tls.certResolver`</a> | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| <a id="opt-tcp-routers-router-name-tls-domainsn-main" href="#opt-tcp-routers-router-name-tls-domainsn-main" title="#opt-tcp-routers-router-name-tls-domainsn-main">`tcp.routers.<router_name>.tls.domains[n].main`</a> | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| <a id="opt-tcp-routers-router-name-tls-domainsn-sansn" href="#opt-tcp-routers-router-name-tls-domainsn-sansn" title="#opt-tcp-routers-router-name-tls-domainsn-sansn">`tcp.routers.<router_name>.tls.domains[n].sans[n]`</a> | See [TLS](../tcp/tls.md) for more information. | `www.example.org` |
| <a id="opt-tcp-routers-router-name-tls-options" href="#opt-tcp-routers-router-name-tls-options" title="#opt-tcp-routers-router-name-tls-options">`tcp.routers.<router_name>.tls.options`</a> | See [TLS](../tcp/tls.md) for more information. | `modern` |
| <a id="opt-tcp-routers-router-name-tls-passthrough" href="#opt-tcp-routers-router-name-tls-passthrough" title="#opt-tcp-routers-router-name-tls-passthrough">`tcp.routers.<router_name>.tls.passthrough`</a> | See [Passthrough](../tcp/tls.md#opt-passthrough) for more information. | `true` |
| <a id="opt-tcp-routers-router-name-priority" href="#opt-tcp-routers-router-name-priority" title="#opt-tcp-routers-router-name-priority">`tcp.routers.<router_name>.priority`</a> | See [priority](../tcp/routing/rules-and-priority.md#priority-calculation) for more information. | `42` |

#### TCP Services

Define TCP services under `tcp.services.<service_name>`.

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tcp-services-service-name-loadBalancer-serversn-address" href="#opt-tcp-services-service-name-loadBalancer-serversn-address" title="#opt-tcp-services-service-name-loadBalancer-serversn-address">`tcp.services.<service_name>.loadBalancer.servers[n].address`</a> | See [servers load balancer](../tcp/service.md#servers-load-balancer) for more information. | `127.0.0.1:9000` |
| <a id="opt-tcp-services-service-name-loadBalancer-serversn-tls" href="#opt-tcp-services-service-name-loadBalancer-serversn-tls" title="#opt-tcp-services-service-name-loadBalancer-serversn-tls">`tcp.services.<service_name>.loadBalancer.servers[n].tls`</a> | Determines whether to use TLS when dialing the backend server. | `true` |
| <a id="opt-tcp-services-service-name-loadBalancer-serversTransport" href="#opt-tcp-services-service-name-loadBalancer-serversTransport" title="#opt-tcp-services-service-name-loadBalancer-serversTransport">`tcp.services.<service_name>.loadBalancer.serversTransport`</a> | See [TCP ServersTransport](../tcp/serverstransport.md) for more information. | `secure-tcp` |
| <a id="opt-tcp-services-service-name-loadBalancer-proxyProtocol-version" href="#opt-tcp-services-service-name-loadBalancer-proxyProtocol-version" title="#opt-tcp-services-service-name-loadBalancer-proxyProtocol-version">`tcp.services.<service_name>.loadBalancer.proxyProtocol.version`</a> | Enables Proxy Protocol for backend connections. | `2` |
| <a id="opt-tcp-services-service-name-loadBalancer-terminationDelay" href="#opt-tcp-services-service-name-loadBalancer-terminationDelay" title="#opt-tcp-services-service-name-loadBalancer-terminationDelay">`tcp.services.<service_name>.loadBalancer.terminationDelay`</a> | Defines the delay before terminating connections. | `100` |
| <a id="opt-tcp-services-service-name-loadBalancer-healthCheck" href="#opt-tcp-services-service-name-loadBalancer-healthCheck" title="#opt-tcp-services-service-name-loadBalancer-healthCheck">`tcp.services.<service_name>.loadBalancer.healthCheck.*`</a> | See [TCP service health check](../tcp/service.md#health-check) for more information. | `interval: 10s` |
| <a id="opt-tcp-services-service-name-weighted-servicesn-name" href="#opt-tcp-services-service-name-weighted-servicesn-name" title="#opt-tcp-services-service-name-weighted-servicesn-name">`tcp.services.<service_name>.weighted.services[n].name`</a> | See [weighted round robin](../tcp/service.md#weighted-round-robin) for more information. | `tcp-v1` |
| <a id="opt-tcp-services-service-name-weighted-servicesn-weight" href="#opt-tcp-services-service-name-weighted-servicesn-weight" title="#opt-tcp-services-service-name-weighted-servicesn-weight">`tcp.services.<service_name>.weighted.services[n].weight`</a> | See [weighted round robin](../tcp/service.md#weighted-round-robin) for more information. | `3` |
| <a id="opt-tcp-services-service-name-weighted-healthCheck" href="#opt-tcp-services-service-name-weighted-healthCheck" title="#opt-tcp-services-service-name-weighted-healthCheck">`tcp.services.<service_name>.weighted.healthCheck`</a> | See [weighted round robin](../tcp/service.md#weighted-round-robin) for more information. | `{}` |

#### TCP Middlewares

Define TCP middlewares under `tcp.middlewares.<middleware_name>`.

For example, to declare an [`InFlightConn`](../tcp/middlewares/inflightconn.md) middleware named `limit`, set `tcp.middlewares.limit.inFlightConn.amount=10`.

More information about available middlewares is available in the dedicated [TCP middlewares section](../tcp/middlewares/overview.md).

!!! warning "The character `@` is not authorized in the middleware name `<middleware_name>`."

!!! warning "Conflicts in Declaration"

    If you declare multiple middlewares with the same name but different parameters, the middleware fails to be declared.

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tcp-middlewares-middleware-name-middleware-type-middleware-option" href="#opt-tcp-middlewares-middleware-name-middleware-type-middleware-option" title="#opt-tcp-middlewares-middleware-name-middleware-type-middleware-option">`tcp.middlewares.<middleware_name>.<middleware_type>.<middleware_option>`</a> | With `middleware_type` the middleware type, such as `inFlightConn`, and `middleware_option` the option to set. | `amount: 10` |

#### TCP ServersTransports

Define TCP ServersTransports under `tcp.serversTransports.<servers_transport_name>`.

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tcp-serversTransports-servers-transport-name" href="#opt-tcp-serversTransports-servers-transport-name" title="#opt-tcp-serversTransports-servers-transport-name">`tcp.serversTransports.<servers_transport_name>.*`</a> | See [TCP ServersTransport](../tcp/serverstransport.md) for more information. | `dialTimeout: 30s` |

### UDP

You can declare UDP routers and services with the file provider.

#### UDP Routers

Define UDP routers under `udp.routers.<router_name>`.

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-udp-routers-router-name-entryPointsn" href="#opt-udp-routers-router-name-entryPointsn" title="#opt-udp-routers-router-name-entryPointsn">`udp.routers.<router_name>.entryPoints[n]`</a> | See [UDP router entrypoints](../udp/routing/rules-priority.md#entrypoints) for more information. | `dns` |
| <a id="opt-udp-routers-router-name-service" href="#opt-udp-routers-router-name-service" title="#opt-udp-routers-router-name-service">`udp.routers.<router_name>.service`</a> | See [UDP router configuration](../udp/routing/rules-priority.md#configuration-example) for more information. | `dns-service` |

#### UDP Services

Define UDP services under `udp.services.<service_name>`.

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-udp-services-service-name-loadBalancer-serversn-address" href="#opt-udp-services-service-name-loadBalancer-serversn-address" title="#opt-udp-services-service-name-loadBalancer-serversn-address">`udp.services.<service_name>.loadBalancer.servers[n].address`</a> | See [UDP service](../udp/service.md) for more information. | `127.0.0.1:5353` |
| <a id="opt-udp-services-service-name-weighted-servicesn-name" href="#opt-udp-services-service-name-weighted-servicesn-name" title="#opt-udp-services-service-name-weighted-servicesn-name">`udp.services.<service_name>.weighted.services[n].name`</a> | See [UDP service](../udp/service.md) for more information. | `dns-v1` |
| <a id="opt-udp-services-service-name-weighted-servicesn-weight" href="#opt-udp-services-service-name-weighted-servicesn-weight" title="#opt-udp-services-service-name-weighted-servicesn-weight">`udp.services.<service_name>.weighted.services[n].weight`</a> | See [UDP service](../udp/service.md) for more information. | `3` |

### TLS

You can declare TLS certificates, options, and stores with the file provider.

#### Certificates

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tls-certificatesn-certFile" href="#opt-tls-certificatesn-certFile" title="#opt-tls-certificatesn-certFile">`tls.certificates[n].certFile`</a> | See [TLS certificates](../http/tls/tls-certificates.md) for more information. | `/certs/example.crt` |
| <a id="opt-tls-certificatesn-keyFile" href="#opt-tls-certificatesn-keyFile" title="#opt-tls-certificatesn-keyFile">`tls.certificates[n].keyFile`</a> | See [TLS certificates](../http/tls/tls-certificates.md) for more information. | `/certs/example.key` |
| <a id="opt-tls-certificatesn-storesn" href="#opt-tls-certificatesn-storesn" title="#opt-tls-certificatesn-storesn">`tls.certificates[n].stores[n]`</a> | See [certificate stores](../http/tls/tls-certificates.md#certificates-stores) for more information. | `default` |

#### TLS Options

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tls-options-options-name-minVersion" href="#opt-tls-options-options-name-minVersion" title="#opt-tls-options-options-name-minVersion">`tls.options.<options_name>.minVersion`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `VersionTLS12` |
| <a id="opt-tls-options-options-name-maxVersion" href="#opt-tls-options-options-name-maxVersion" title="#opt-tls-options-options-name-maxVersion">`tls.options.<options_name>.maxVersion`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `VersionTLS13` |
| <a id="opt-tls-options-options-name-cipherSuitesn" href="#opt-tls-options-options-name-cipherSuitesn" title="#opt-tls-options-options-name-cipherSuitesn">`tls.options.<options_name>.cipherSuites[n]`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256` |
| <a id="opt-tls-options-options-name-curvePreferencesn" href="#opt-tls-options-options-name-curvePreferencesn" title="#opt-tls-options-options-name-curvePreferencesn">`tls.options.<options_name>.curvePreferences[n]`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `CurveP256` |
| <a id="opt-tls-options-options-name-clientAuth-caFilesn" href="#opt-tls-options-options-name-clientAuth-caFilesn" title="#opt-tls-options-options-name-clientAuth-caFilesn">`tls.options.<options_name>.clientAuth.caFiles[n]`</a> | See [client authentication](../http/tls/tls-options.md#client-authentication-mtls) for more information. | `/certs/client-ca.crt` |
| <a id="opt-tls-options-options-name-clientAuth-clientAuthType" href="#opt-tls-options-options-name-clientAuth-clientAuthType" title="#opt-tls-options-options-name-clientAuth-clientAuthType">`tls.options.<options_name>.clientAuth.clientAuthType`</a> | See [client authentication](../http/tls/tls-options.md#client-authentication-mtls) for more information. | `RequireAndVerifyClientCert` |
| <a id="opt-tls-options-options-name-sniStrict" href="#opt-tls-options-options-name-sniStrict" title="#opt-tls-options-options-name-sniStrict">`tls.options.<options_name>.sniStrict`</a> | See [strict SNI checking](../http/tls/tls-options.md#strict-sni-checking) for more information. | `true` |
| <a id="opt-tls-options-options-name-alpnProtocolsn" href="#opt-tls-options-options-name-alpnProtocolsn" title="#opt-tls-options-options-name-alpnProtocolsn">`tls.options.<options_name>.alpnProtocols[n]`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `h2` |
| <a id="opt-tls-options-options-name-disableSessionTickets" href="#opt-tls-options-options-name-disableSessionTickets" title="#opt-tls-options-options-name-disableSessionTickets">`tls.options.<options_name>.disableSessionTickets`</a> | See [TLS options](../http/tls/tls-options.md) for more information. | `true` |
| <a id="opt-tls-options-options-name-preferServerCipherSuites" href="#opt-tls-options-options-name-preferServerCipherSuites" title="#opt-tls-options-options-name-preferServerCipherSuites">`tls.options.<options_name>.preferServerCipherSuites`</a> | **Deprecated:** This option is no longer effective and will be ignored by the Go TLS stack. See [TLS options](../http/tls/tls-options.md) for more information. | `true` |

#### TLS Stores

| Field | Description | Value |
|------|-------------|-------|
| <a id="opt-tls-stores-store-name-defaultCertificate-certFile" href="#opt-tls-stores-store-name-defaultCertificate-certFile" title="#opt-tls-stores-store-name-defaultCertificate-certFile">`tls.stores.<store_name>.defaultCertificate.certFile`</a> | See [default certificate](../http/tls/tls-certificates.md#default-certificate) for more information. | `/certs/default.crt` |
| <a id="opt-tls-stores-store-name-defaultCertificate-keyFile" href="#opt-tls-stores-store-name-defaultCertificate-keyFile" title="#opt-tls-stores-store-name-defaultCertificate-keyFile">`tls.stores.<store_name>.defaultCertificate.keyFile`</a> | See [default certificate](../http/tls/tls-certificates.md#default-certificate) for more information. | `/certs/default.key` |
| <a id="opt-tls-stores-store-name-defaultGeneratedCert-resolver" href="#opt-tls-stores-store-name-defaultGeneratedCert-resolver" title="#opt-tls-stores-store-name-defaultGeneratedCert-resolver">`tls.stores.<store_name>.defaultGeneratedCert.resolver`</a> | See [ACME default certificate](../http/tls/tls-certificates.md#acme-default-certificate) for more information. | `myresolver` |
| <a id="opt-tls-stores-store-name-defaultGeneratedCert-domain-main" href="#opt-tls-stores-store-name-defaultGeneratedCert-domain-main" title="#opt-tls-stores-store-name-defaultGeneratedCert-domain-main">`tls.stores.<store_name>.defaultGeneratedCert.domain.main`</a> | See [ACME default certificate](../http/tls/tls-certificates.md#acme-default-certificate) for more information. | `example.org` |
| <a id="opt-tls-stores-store-name-defaultGeneratedCert-domain-sansn" href="#opt-tls-stores-store-name-defaultGeneratedCert-domain-sansn" title="#opt-tls-stores-store-name-defaultGeneratedCert-domain-sansn">`tls.stores.<store_name>.defaultGeneratedCert.domain.sans[n]`</a> | See [ACME default certificate](../http/tls/tls-certificates.md#acme-default-certificate) for more information. | `www.example.org` |

## Go Templating

!!! warning

    Go Templating only works with dedicated dynamic configuration files.
    Templating does not work in the Traefik main static configuration file.

Traefik supports using Go templating to automatically generate repetitive sections of configuration files.
These sections must be a valid [Go template](https://pkg.go.dev/text/template/), and can use
[sprig template functions](https://masterminds.github.io/sprig/).

To illustrate, it is possible to easily define multiple routers, services, and TLS certificates as described in the following examples:

??? example "Configuring Using Templating"

    ```yaml tab="YAML"
    http:
      routers:
        {{range $i, $e := until 100 }}
        router{{ $e }}-{{ env "MY_ENV_VAR" }}:
          # ...
        {{end}}

      services:
        {{range $i, $e := until 100 }}
        application{{ $e }}:
          # ...
        {{end}}

    tcp:
      routers:
        {{range $i, $e := until 100 }}
        router{{ $e }}:
          # ...
        {{end}}

      services:
        {{range $i, $e := until 100 }}
        service{{ $e }}:
          # ...
        {{end}}

    tls:
      certificates:
      {{ range $i, $e := until 10 }}
      - certFile: "/etc/traefik/cert-{{ $e }}.pem"
        keyFile: "/etc/traefik/cert-{{ $e }}.key"
        stores:
        - "my-store-foo-{{ $e }}"
        - "my-store-bar-{{ $e }}"
      {{end}}
    ```

    ```toml tab="TOML"
    # template-rules.toml
    [http]

      [http.routers]
      {{ range $i, $e := until 100 }}
        [http.routers.router{{ $e }}-{{ env "MY_ENV_VAR" }}]
        # ...
      {{ end }}

      [http.services]
      {{ range $i, $e := until 100 }}
          [http.services.service{{ $e }}]
          # ...
      {{ end }}

    [tcp]

      [tcp.routers]
      {{ range $i, $e := until 100 }}
        [tcp.routers.router{{ $e }}]
        # ...
      {{ end }}

      [tcp.services]
      {{ range $i, $e := until 100 }}
          [tcp.services.service{{ $e }}]
          # ...
      {{ end }}

    {{ range $i, $e := until 10 }}
    [[tls.certificates]]
      certFile = "/etc/traefik/cert-{{ $e }}.pem"
      keyFile = "/etc/traefik/cert-{{ $e }}.key"
      stores = ["my-store-foo-{{ $e }}", "my-store-bar-{{ $e }}"]
    {{ end }}

    [tls.options]
    {{ range $i, $e := until 10 }}
      [tls.options.TLS{{ $e }}]
      # ...
    {{ end }}
    ```
