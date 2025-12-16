---
title: "Traefik ECS Documentation"
description: "AWS ECS is a provider for routing and load balancing in Traefik Proxy. Read the technical documentation to get started."
---

# Traefik & ECS

One of the best feature of Traefik is to delegate the routing configuration to the application level.
With ECS, Traefik can leverage labels attached to a container to generate routing rules.

!!! warning "Labels & sensitive data"

    We recommend to *not* use labels to store sensitive data (certificates, credentials, etc).
    Instead, we recommend to store sensitive data in a safer storage (secrets, file, etc).

## Configuration Examples

??? example "Configuring ECS & Deploying / Exposing one Service"

    Enabling the ECS provider

    ```yaml tab="Structured (YAML)"
    providers:
      ecs: {}
    ```

    ```toml tab="Structured (TOML)"
    [providers.ecs]
    ```

    ```bash tab="CLI"
    --providers.ecs=true
    ```

    Attaching labels to containers (in your ECS task definition)

    ```json
    {
      "family": "my-service",
      "containerDefinitions": [
        {
          "name": "my-container",
          "image": "my-image:latest",
          "labels": {
            "traefik.http.routers.my-container.rule": "Host(`example.com`)"
          }
        }
      ]
    }
    ```

??? example "Specify a Custom Port for the Container"

    Forward requests for `http://example.com` to `http://<private IP of container>:12345`:

    ```json
    {
      "family": "my-service",
      "containerDefinitions": [
        {
          "name": "my-container",
          "image": "my-image:latest",
          "labels": {
            "traefik.http.routers.my-container.rule": "Host(`example.com`)",
            "traefik.http.routers.my-container.service": "my-service",
            "traefik.http.services.my-service.loadbalancer.server.port": "12345"
          }
        }
      ]
    }
    ```

    !!! important "Traefik Connecting to the Wrong Port: `HTTP/502 Gateway Error`"
        By default, Traefik uses the first exposed port of a container.

        Setting the label `traefik.http.services.xxx.loadbalancer.server.port`
        overrides that behavior.

??? example "Specifying more than one router and service per container"

    Forwarding requests to more than one port on a container requires referencing the service loadbalancer port definition using the service parameter on the router.

    In this example, requests are forwarded for `http://example-a.com` to `http://<private IP of container>:8000` in addition to `http://example-b.com` forwarding to `http://<private IP of container>:9000`:

    ```json
    {
      "family": "my-service",
      "containerDefinitions": [
        {
          "name": "my-container",
          "image": "my-image:latest",
          "labels": {
            "traefik.http.routers.www-router.rule": "Host(`example-a.com`)",
            "traefik.http.routers.www-router.service": "www-service",
            "traefik.http.services.www-service.loadbalancer.server.port": "8000",
            "traefik.http.routers.admin-router.rule": "Host(`example-b.com`)",
            "traefik.http.routers.admin-router.service": "admin-service",
            "traefik.http.services.admin-service.loadbalancer.server.port": "9000"
          }
        }
      ]
    }
    ```

## Configuration Options

!!! info "labels"
    
    Labels are case-insensitive.

!!! tip "TLS Default Generated Certificates"

    To learn how to configure Traefik default generated certificate, refer to the [TLS Certificates](../http/tls/tls-certificates.md#acme-default-certificate) page.

### General

Traefik creates, for each elastic service, a corresponding [service](../http/load-balancing/service.md) and [router](../http/routing/rules-and-priority.md).

The Service automatically gets a server per elastic container, and the router gets a default rule attached to it, based on the service name.

### Routers

To update the configuration of the Router automatically attached to the service, add labels starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.

For example, to change the rule, you could add the label ```traefik.http.routers.my-service.rule=Host(`example.com`)```.

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-http-routers-router-name-rule" href="#opt-traefik-http-routers-router-name-rule" title="#opt-traefik-http-routers-router-name-rule">`traefik.http.routers.<router_name>.rule`</a> | See [rule](../http/routing/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)``` |
| <a id="opt-traefik-http-routers-router-name-ruleSyntax" href="#opt-traefik-http-routers-router-name-ruleSyntax" title="#opt-traefik-http-routers-router-name-ruleSyntax">`traefik.http.routers.<router_name>.ruleSyntax`</a> | See [ruleSyntax](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3` |
| <a id="opt-traefik-http-routers-router-name-entrypoints" href="#opt-traefik-http-routers-router-name-entrypoints" title="#opt-traefik-http-routers-router-name-entrypoints">`traefik.http.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `web,websecure` |
| <a id="opt-traefik-http-routers-router-name-middlewares" href="#opt-traefik-http-routers-router-name-middlewares" title="#opt-traefik-http-routers-router-name-middlewares">`traefik.http.routers.<router_name>.middlewares`</a> | See [middlewares overview](../http/middlewares/overview.md) for more information. | `auth,prefix,cb` |
| <a id="opt-traefik-http-routers-router-name-service" href="#opt-traefik-http-routers-router-name-service" title="#opt-traefik-http-routers-router-name-service">`traefik.http.routers.<router_name>.service`</a> | See [service](../http/load-balancing/service.md) for more information. | `myservice` |
| <a id="opt-traefik-http-routers-router-name-tls" href="#opt-traefik-http-routers-router-name-tls" title="#opt-traefik-http-routers-router-name-tls">`traefik.http.routers.<router_name>.tls`</a> | See [tls](../http/tls/overview.md) for more information. | `true` |
| <a id="opt-traefik-http-routers-router-name-tls-certresolver" href="#opt-traefik-http-routers-router-name-tls-certresolver" title="#opt-traefik-http-routers-router-name-tls-certresolver">`traefik.http.routers.<router_name>.tls.certresolver`</a> | See [certResolver](../../install-configuration/tls/certificate-resolvers/overview.md) for more information. | `myresolver` |
| <a id="opt-traefik-http-routers-router-name-tls-domainsn-main" href="#opt-traefik-http-routers-router-name-tls-domainsn-main" title="#opt-traefik-http-routers-router-name-tls-domainsn-main">`traefik.http.routers.<router_name>.tls.domains[n].main`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `example.org` |
| <a id="opt-traefik-http-routers-router-name-tls-domainsn-sans" href="#opt-traefik-http-routers-router-name-tls-domainsn-sans" title="#opt-traefik-http-routers-router-name-tls-domainsn-sans">`traefik.http.routers.<router_name>.tls.domains[n].sans`</a> | See [domains](../../install-configuration/tls/certificate-resolvers/acme.md#domain-definition) for more information. | `test.example.org,dev.example.org` |
| <a id="opt-traefik-http-routers-router-name-tls-options" href="#opt-traefik-http-routers-router-name-tls-options" title="#opt-traefik-http-routers-router-name-tls-options">`traefik.http.routers.<router_name>.tls.options`</a> |  | `foobar` |
| <a id="opt-traefik-http-routers-router-name-observability-accesslogs" href="#opt-traefik-http-routers-router-name-observability-accesslogs" title="#opt-traefik-http-routers-router-name-observability-accesslogs">`traefik.http.routers.<router_name>.observability.accesslogs`</a> | The accessLogs option controls whether the router will produce access-logs. | `true` |
| <a id="opt-traefik-http-routers-router-name-observability-metrics" href="#opt-traefik-http-routers-router-name-observability-metrics" title="#opt-traefik-http-routers-router-name-observability-metrics">`traefik.http.routers.<router_name>.observability.metrics`</a> | The metrics option controls whether the router will produce metrics. | `true` |
| <a id="opt-traefik-http-routers-router-name-observability-tracing" href="#opt-traefik-http-routers-router-name-observability-tracing" title="#opt-traefik-http-routers-router-name-observability-tracing">`traefik.http.routers.<router_name>.observability.tracing`</a> | The tracing option controls whether the router will produce traces. | `true` |
| <a id="opt-traefik-http-routers-router-name-priority" href="#opt-traefik-http-routers-router-name-priority" title="#opt-traefik-http-routers-router-name-priority">`traefik.http.routers.<router_name>.priority`</a> | See [priority](../http/routing/rules-and-priority.md#priority-calculation) for more information. | `42` |

### Services

To update the configuration of the Service automatically attached to the service,
add labels starting with `traefik.http.services.{name-of-your-choice}.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the label `traefik.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-http-services-service-name-loadbalancer-server-port" href="#opt-traefik-http-services-service-name-loadbalancer-server-port" title="#opt-traefik-http-services-service-name-loadbalancer-server-port">`traefik.http.services.<service_name>.loadbalancer.server.port`</a> | Registers a port.<br/>Useful when the service exposes multiples ports. | `8080` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-server-scheme" href="#opt-traefik-http-services-service-name-loadbalancer-server-scheme" title="#opt-traefik-http-services-service-name-loadbalancer-server-scheme">`traefik.http.services.<service_name>.loadbalancer.server.scheme`</a> | Overrides the default scheme. | `http` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-serverstransport" href="#opt-traefik-http-services-service-name-loadbalancer-serverstransport" title="#opt-traefik-http-services-service-name-loadbalancer-serverstransport">`traefik.http.services.<service_name>.loadbalancer.serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-passhostheader" href="#opt-traefik-http-services-service-name-loadbalancer-passhostheader" title="#opt-traefik-http-services-service-name-loadbalancer-passhostheader">`traefik.http.services.<service_name>.loadbalancer.passhostheader`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-headers-header-name" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-headers-header-name" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-headers-header-name">`traefik.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-hostname" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-hostname" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-hostname">`traefik.http.services.<service_name>.loadbalancer.healthcheck.hostname`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-interval" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-interval" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-interval">`traefik.http.services.<service_name>.loadbalancer.healthcheck.interval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-unhealthyinterval" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-unhealthyinterval" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-unhealthyinterval">`traefik.http.services.<service_name>.loadbalancer.healthcheck.unhealthyinterval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-path" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-path" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-path">`traefik.http.services.<service_name>.loadbalancer.healthcheck.path`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-method" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-method" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-method">`traefik.http.services.<service_name>.loadbalancer.healthcheck.method`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-status" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-status" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-status">`traefik.http.services.<service_name>.loadbalancer.healthcheck.status`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-port" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-port" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-port">`traefik.http.services.<service_name>.loadbalancer.healthcheck.port`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-scheme" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-scheme" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-scheme">`traefik.http.services.<service_name>.loadbalancer.healthcheck.scheme`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-timeout" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-timeout" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-timeout">`traefik.http.services.<service_name>.loadbalancer.healthcheck.timeout`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-followredirects" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-followredirects" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-followredirects">`traefik.http.services.<service_name>.loadbalancer.healthcheck.followredirects`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-httponly" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-httponly" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-httponly">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-name" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-name" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-name">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.name`</a> |  | `foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-path" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-path" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-path">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.path`</a> |  | `/foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-secure" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-secure" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-secure">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.secure`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-samesite" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-samesite" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-samesite">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.samesite`</a> |  | `none` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-maxage" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-maxage" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-maxage">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.maxage`</a> |  | `42` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-responseforwarding-flushinterval" href="#opt-traefik-http-services-service-name-loadbalancer-responseforwarding-flushinterval" title="#opt-traefik-http-services-service-name-loadbalancer-responseforwarding-flushinterval">`traefik.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`</a> | `FlushInterval` specifies the flush interval to flush to the client while copying the response body. | `10` |

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../http/middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme: https`.

More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md).

!!! warning "The character `@` is not authorized in the middleware name."

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

You can declare TCP Routers and/or Services using labels.

??? example "Declaring TCP Routers and Services"

    ```yaml
    traefik.tcp.routers.my-router.rule=HostSNI(`example.com`)
    traefik.tcp.routers.my-router.tls=true
    traefik.tcp.services.my-service.loadbalancer.server.port=4123
    ```

!!! warning "TCP and HTTP"

    If you declare a TCP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no TCP Router/Service is defined).
    You can declare both a TCP Router/Service and an HTTP Router/Service for the same elastic service (but you have to do so manually).

#### TCP Routers

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-tcp-routers-router-name-entrypoints" href="#opt-traefik-tcp-routers-router-name-entrypoints" title="#opt-traefik-tcp-routers-router-name-entrypoints">`traefik.tcp.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1,ep2` |
| <a id="opt-traefik-tcp-routers-router-name-rule" href="#opt-traefik-tcp-routers-router-name-rule" title="#opt-traefik-tcp-routers-router-name-rule">`traefik.tcp.routers.<router_name>.rule`</a> | See [rule](../tcp/routing/rules-and-priority.md#rules) for more information. | ```HostSNI(`example.com`)``` |
| <a id="opt-traefik-tcp-routers-router-name-ruleSyntax" href="#opt-traefik-tcp-routers-router-name-ruleSyntax" title="#opt-traefik-tcp-routers-router-name-ruleSyntax">`traefik.tcp.routers.<router_name>.ruleSyntax`</a> | configure the rule syntax to be used for parsing the rule on a per-router basis.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3` |
| <a id="opt-traefik-tcp-routers-router-name-service" href="#opt-traefik-tcp-routers-router-name-service" title="#opt-traefik-tcp-routers-router-name-service">`traefik.tcp.routers.<router_name>.service`</a> | See [service](../tcp/service.md) for more information. | `myservice` |
| <a id="opt-traefik-tcp-routers-router-name-tls" href="#opt-traefik-tcp-routers-router-name-tls" title="#opt-traefik-tcp-routers-router-name-tls">`traefik.tcp.routers.<router_name>.tls`</a> | See [TLS](../tcp/tls.md) for more information. | `true` |
| <a id="opt-traefik-tcp-routers-router-name-tls-certresolver" href="#opt-traefik-tcp-routers-router-name-tls-certresolver" title="#opt-traefik-tcp-routers-router-name-tls-certresolver">`traefik.tcp.routers.<router_name>.tls.certresolver`</a> | See [certResolver](../tcp/tls.md#configuration-options) for more information. | `myresolver` |
| <a id="opt-traefik-tcp-routers-router-name-tls-domainsn-main" href="#opt-traefik-tcp-routers-router-name-tls-domainsn-main" title="#opt-traefik-tcp-routers-router-name-tls-domainsn-main">`traefik.tcp.routers.<router_name>.tls.domains[n].main`</a> | See [TLS](../tcp/tls.md) for more information. | `example.org` |
| <a id="opt-traefik-tcp-routers-router-name-tls-domainsn-sans" href="#opt-traefik-tcp-routers-router-name-tls-domainsn-sans" title="#opt-traefik-tcp-routers-router-name-tls-domainsn-sans">`traefik.tcp.routers.<router_name>.tls.domains[n].sans`</a> | See [TLS](../tcp/tls.md) for more information. | `test.example.org,dev.example.org` |
| <a id="opt-traefik-tcp-routers-router-name-tls-options" href="#opt-traefik-tcp-routers-router-name-tls-options" title="#opt-traefik-tcp-routers-router-name-tls-options">`traefik.tcp.routers.<router_name>.tls.options`</a> | See [TLS](../tcp/tls.md) for more information. | `mysoptions` |
| <a id="opt-traefik-tcp-routers-router-name-tls-passthrough" href="#opt-traefik-tcp-routers-router-name-tls-passthrough" title="#opt-traefik-tcp-routers-router-name-tls-passthrough">`traefik.tcp.routers.<router_name>.tls.passthrough`</a> | See [Passthrough](../tcp/tls.md#opt-passthrough) for more information. | `true` |
| <a id="opt-traefik-tcp-routers-router-name-priority" href="#opt-traefik-tcp-routers-router-name-priority" title="#opt-traefik-tcp-routers-router-name-priority">`traefik.tcp.routers.<router_name>.priority`</a> | See [priority](../tcp/routing/rules-and-priority.md#priority-calculation) for more information. | `42` |

#### TCP Services

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-tcp-services-service-name-loadbalancer-server-port" href="#opt-traefik-tcp-services-service-name-loadbalancer-server-port" title="#opt-traefik-tcp-services-service-name-loadbalancer-server-port">`traefik.tcp.services.<service_name>.loadbalancer.server.port`</a> | Registers a port of the application. | `423` |
| <a id="opt-traefik-tcp-services-service-name-loadbalancer-server-tls" href="#opt-traefik-tcp-services-service-name-loadbalancer-server-tls" title="#opt-traefik-tcp-services-service-name-loadbalancer-server-tls">`traefik.tcp.services.<service_name>.loadbalancer.server.tls`</a> | Determines whether to use TLS when dialing with the backend. | `true` |
| <a id="opt-traefik-tcp-services-service-name-loadbalancer-server-weight" href="#opt-traefik-tcp-services-service-name-loadbalancer-server-weight" title="#opt-traefik-tcp-services-service-name-loadbalancer-server-weight">`traefik.tcp.services.<service_name>.loadbalancer.server.weight`</a> | Overrides the default weight. | `42` |
| <a id="opt-traefik-tcp-services-service-name-loadbalancer-serverstransport" href="#opt-traefik-tcp-services-service-name-loadbalancer-serverstransport" title="#opt-traefik-tcp-services-service-name-loadbalancer-serverstransport">`traefik.tcp.services.<service_name>.loadbalancer.serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |

### UDP

You can declare UDP Routers and/or Services using tags.

??? example "Declaring UDP Routers and Services"

    ```yaml
    traefik.udp.routers.my-router.entrypoints=udp
    traefik.udp.services.my-service.loadbalancer.server.port=4123
    ```

!!! warning "UDP and HTTP"

    If you declare a UDP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no UDP Router/Service is defined).
    You can declare both a UDP Router/Service and an HTTP Router/Service for the same elastic service (but you have to do so manually).

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

#### UDP Routers

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-udp-routers-router-name-entrypoints" href="#opt-traefik-udp-routers-router-name-entrypoints" title="#opt-traefik-udp-routers-router-name-entrypoints">`traefik.udp.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1,ep2` |
| <a id="opt-traefik-udp-routers-router-name-service" href="#opt-traefik-udp-routers-router-name-service" title="#opt-traefik-udp-routers-router-name-service">`traefik.udp.routers.<router_name>.service`</a> | See [service](../udp/service.md) for more information. | `myservice` |

#### UDP Services

##### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-udp-services-service-name-loadbalancer-server-port" href="#opt-traefik-udp-services-service-name-loadbalancer-server-port" title="#opt-traefik-udp-services-service-name-loadbalancer-server-port">`traefik.udp.services.<service_name>.loadbalancer.server.port`</a> | Registers a port of the application. | `423` |

### Specific Provider Options

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-enable" href="#opt-traefik-enable" title="#opt-traefik-enable">`traefik.enable`</a> | You can tell Traefik to consider (or not) the ECS service by setting `traefik.enable` to true or false.<br/>This option overrides the value of `exposedByDefault`. | `true` |
