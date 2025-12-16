---
title: "Traefik Docker Swarm Routing Documentation"
description: "This guide will teach you how to attach labels to your containers, to route traffic and load balance with Traefik and Docker Swarm."
---

# Traefik & Docker Swarm

One of the best feature of Traefik is to delegate the routing configuration to the application level.
With Docker Swarm, Traefik can leverage labels attached to a service to generate routing rules.

!!! warning "Labels & sensitive data"

    We recommend to *not* use labels to store sensitive data (certificates, credentials, etc).
    Instead, we recommend to store sensitive data in a safer storage (secrets, file, etc).

## Configuration Examples

??? example "Configuring Docker Swarm & Deploying / Exposing one Service"

    Enabling the docker provider (Swarm Mode)

    ```yaml tab="Structured (YAML)"
    providers:
      swarm:
        # swarm classic (1.12-)
        # endpoint: "tcp://127.0.0.1:2375"
        # docker swarm mode (1.12+)
        endpoint: "tcp://127.0.0.1:2377"
    ```

    ```toml tab="Structured (TOML)"
    [providers.swarm]
      # swarm classic (1.12-)
      # endpoint = "tcp://127.0.0.1:2375"
      # docker swarm mode (1.12+)
      endpoint = "tcp://127.0.0.1:2377"
    ```

    ```bash tab="CLI"
    # swarm classic (1.12-)
    # --providers.swarm.endpoint=tcp://127.0.0.1:2375
    # docker swarm mode (1.12+)
    --providers.swarm.endpoint=tcp://127.0.0.1:2377
    ```

    Attach labels to services (not containers) while in Swarm mode (in your Docker compose file).
    When there is only one service, and the router does not specify a service,
    then that service is automatically assigned to the router.

    ```yaml
    services:
      my-container:
        deploy:
          labels:
            - traefik.http.routers.my-container.rule=Host(`example.com`)
            - traefik.http.services.my-container-service.loadbalancer.server.port=8080
    ```

    !!! important "Labels in Docker Swarm Mode"
        While in Swarm Mode, Traefik uses labels found on services, not on individual containers.
        Therefore, if you use a compose file with Swarm Mode, labels should be defined in the `deploy` part of your service.
        This behavior is only enabled for docker-compose version 3+ ([Compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/#labels-1)).

??? example "Specify a Custom Port for the Container"

    Forward requests for `http://example.com` to `http://<private IP of container>:12345`:

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - traefik.http.routers.my-container.rule=Host(`example.com`)
            - traefik.http.routers.my-container.service=my-service"
            # Tell Traefik to use the port 12345 to connect to `my-container`
            - traefik.http.services.my-service.loadbalancer.server.port=12345
    ```

    !!! important "Traefik Connecting to the Wrong Port: `HTTP/502 Gateway Error`"
        By default, Traefik uses the lowest exposed port of a container as detailed in
        [Port Detection](../../install-configuration/providers/swarm.md#port-detection) of the Swarm provider.

        Setting the label `traefik.http.services.xxx.loadbalancer.server.port`
        overrides this behavior.

??? example "Specifying more than one router and service per container"

    Forwarding requests to more than one port on a container requires referencing the service loadbalancer port definition using the service parameter on the router.

    In this example, requests are forwarded for `http://example-a.com` to `http://<private IP of container>:8000` in addition to `http://example-b.com` forwarding to `http://<private IP of container>:9000`:

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - traefik.http.routers.www-router.rule=Host(`example-a.com`)
            - traefik.http.routers.www-router.service=www-service
            - traefik.http.services.www-service.loadbalancer.server.port=8000
            - traefik.http.routers.admin-router.rule=Host(`example-b.com`)
            - traefik.http.routers.admin-router.service=admin-service
            - traefik.http.services.admin-service.loadbalancer.server.port=9000
    ```

## Configuration Options

!!! info "Labels"

     Labels are case-insensitive.

!!! tip "TLS Default Generated Certificates"

    To learn how to configure Traefik default generated certificate, refer to the [TLS Certificates](../http/tls/tls-certificates.md#acme-default-certificate) page.

### General

Traefik creates, for each container, a corresponding [service](../http/load-balancing/service.md) and [router](../http/routing/rules-and-priority.md).

The Service automatically gets a server per instance of the container,
and the router automatically gets a rule defined by `defaultRule` (if no rule for it was defined in labels).

#### Service definition

--8<-- "content/routing/providers/service-by-label.md"

??? example "Automatic service assignment with labels"

    With labels in a compose file

    ```yaml
    labels:
      - "traefik.http.routers.myproxy.rule=Host(`example.net`)"
      # service myservice gets automatically assigned to router myproxy
      - "traefik.http.services.myservice.loadbalancer.server.port=8080"
    ```

??? example "Automatic service creation and assignment with labels"

    With labels in a compose file

    ```yaml
    labels:
      # no service specified or defined and yet one gets automatically created
      # and assigned to router myproxy.
      - "traefik.http.routers.myproxy.rule=Host(`example.net`)"
    ```

### Routers

To update the configuration of the Router automatically attached to the container,
add labels starting with `traefik.http.routers.<name-of-your-choice>.` and followed by the option you want to change.

For example, to change the rule, you could add the label ```traefik.http.routers.my-container.rule=Host(`example.com`)```.

!!! warning "The character `@` is not authorized in the router name `<router_name>`."

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-http-routers-router-name-rule" href="#opt-traefik-http-routers-router-name-rule" title="#opt-traefik-http-routers-router-name-rule">`traefik.http.routers.<router_name>.rule`</a> | See [rule](../http/routing/rules-and-priority.md#rules) for more information. | ```Host(`example.com`)``` |
| <a id="opt-traefik-http-routers-router-name-ruleSyntax" href="#opt-traefik-http-routers-router-name-ruleSyntax" title="#opt-traefik-http-routers-router-name-ruleSyntax">`traefik.http.routers.<router_name>.ruleSyntax`</a> | See [ruleSyntax](../http/routing/rules-and-priority.md#rulesyntax) for more information.<br/>RuleSyntax option is deprecated and will be removed in the next major version.<br/>Please do not use this field and rewrite the router rules to use the v3 syntax. | `v3` |
| <a id="opt-traefik-http-routers-router-name-entrypoints" href="#opt-traefik-http-routers-router-name-entrypoints" title="#opt-traefik-http-routers-router-name-entrypoints">`traefik.http.routers.<router_name>.entrypoints`</a> | See [entry points](../../install-configuration/entrypoints.md) for more information. | `ep1,ep2` |
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

To update the configuration of the Service automatically attached to the container,
add labels starting with `traefik.http.services.<name-of-your-choice>.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the label `traefik.http.services.<name-of-your-choice>.loadbalancer.passhostheader=false`.

!!! warning "The character `@` is not authorized in the service name `<service_name>`."

#### Configuration Options

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-http-services-service-name-loadbalancer-server-port" href="#opt-traefik-http-services-service-name-loadbalancer-server-port" title="#opt-traefik-http-services-service-name-loadbalancer-server-port">`traefik.http.services.<service_name>.loadbalancer.server.port`</a> | Registers a port.<br/>Useful when the container exposes multiples ports.<br/>Mandatory for Docker Swarm (see the section ["Port Detection with Docker Swarm"](../../install-configuration/providers/swarm.md#port-detection)). | `8080` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-server-scheme" href="#opt-traefik-http-services-service-name-loadbalancer-server-scheme" title="#opt-traefik-http-services-service-name-loadbalancer-server-scheme">`traefik.http.services.<service_name>.loadbalancer.server.scheme`</a> | Overrides the default scheme. | `http` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-server-url" href="#opt-traefik-http-services-service-name-loadbalancer-server-url" title="#opt-traefik-http-services-service-name-loadbalancer-server-url">`traefik.http.services.<service_name>.loadbalancer.server.url`</a> | Defines the service URL.<br/>This option cannot be used in combination with `port` or `scheme` definition. | `http://foobar:8080` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-server-weight" href="#opt-traefik-http-services-service-name-loadbalancer-server-weight" title="#opt-traefik-http-services-service-name-loadbalancer-server-weight">`traefik.http.services.<service_name>.loadbalancer.server.weight`</a> | Overrides the default weight. | `42` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-serverstransport" href="#opt-traefik-http-services-service-name-loadbalancer-serverstransport" title="#opt-traefik-http-services-service-name-loadbalancer-serverstransport">`traefik.http.services.<service_name>.loadbalancer.serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../http/load-balancing/serverstransport.md) for more information. | `foobar@file` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-passhostheader" href="#opt-traefik-http-services-service-name-loadbalancer-passhostheader" title="#opt-traefik-http-services-service-name-loadbalancer-passhostheader">`traefik.http.services.<service_name>.loadbalancer.passhostheader`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-headers-header-name" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-headers-header-name" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-headers-header-name">`traefik.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-hostname" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-hostname" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-hostname">`traefik.http.services.<service_name>.loadbalancer.healthcheck.hostname`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `example.org` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-interval" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-interval" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-interval">`traefik.http.services.<service_name>.loadbalancer.healthcheck.interval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10s` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-unhealthyinterval" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-unhealthyinterval" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-unhealthyinterval">`traefik.http.services.<service_name>.loadbalancer.healthcheck.unhealthyinterval`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10s` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-path" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-path" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-path">`traefik.http.services.<service_name>.loadbalancer.healthcheck.path`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `/foo` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-method" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-method" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-method">`traefik.http.services.<service_name>.loadbalancer.healthcheck.method`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-status" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-status" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-status">`traefik.http.services.<service_name>.loadbalancer.healthcheck.status`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-port" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-port" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-port">`traefik.http.services.<service_name>.loadbalancer.healthcheck.port`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `42` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-scheme" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-scheme" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-scheme">`traefik.http.services.<service_name>.loadbalancer.healthcheck.scheme`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `http` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-timeout" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-timeout" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-timeout">`traefik.http.services.<service_name>.loadbalancer.healthcheck.timeout`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `10s` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-healthcheck-followredirects" href="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-followredirects" title="#opt-traefik-http-services-service-name-loadbalancer-healthcheck-followredirects">`traefik.http.services.<service_name>.loadbalancer.healthcheck.followredirects`</a> | See [health check](../http/load-balancing/service.md#health-check) for more information. | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-httponly" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-httponly" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-httponly">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-name" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-name" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-name">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.name`</a> |  | `foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-path" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-path" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-path">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.path`</a> |  | `/foobar` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-secure" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-secure" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-secure">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.secure`</a> |  | `true` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-samesite" href="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-samesite" title="#opt-traefik-http-services-service-name-loadbalancer-sticky-cookie-samesite">`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.samesite`</a> |  | `none` |
| <a id="opt-traefik-http-services-service-name-loadbalancer-responseforwarding-flushinterval" href="#opt-traefik-http-services-service-name-loadbalancer-responseforwarding-flushinterval" title="#opt-traefik-http-services-service-name-loadbalancer-responseforwarding-flushinterval">`traefik.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`</a> | See [response forwarding](../http/load-balancing/service.md#configuration-options) for more information. | `10` |

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.<name-of-your-choice>.`,
followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../http/middlewares/redirectscheme.md) named `my-redirect`,
you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme=https`.

More information about available middlewares in the dedicated [middlewares section](../http/middlewares/overview.md).

!!! warning "The character `@` is not authorized in the middleware name."

??? example "Declaring and Referencing a Middleware"

    ```yaml
    services:
      my-container:
        # ...
        deploy:
        labels:
          # Declaring a middleware
          - traefik.http.middlewares.my-redirect.redirectscheme.scheme=https
          # Referencing a middleware
          - traefik.http.routers.my-container.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### TCP

You can declare TCP Routers and/or Services using labels.

??? example "Declaring TCP Routers and Services"

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - "traefik.tcp.routers.my-router.rule=HostSNI(`example.com`)"
            - "traefik.tcp.routers.my-router.tls=true"
            - "traefik.tcp.services.my-service.loadbalancer.server.port=4123"
    ```

!!! warning "TCP and HTTP"

    If you declare a TCP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no TCP Router/Service is defined).
    You can declare both a TCP Router/Service and an HTTP Router/Service for the same container (but you have to do so manually).

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
| <a id="opt-traefik-tcp-services-service-name-loadbalancer-serverstransport" href="#opt-traefik-tcp-services-service-name-loadbalancer-serverstransport" title="#opt-traefik-tcp-services-service-name-loadbalancer-serverstransport">`traefik.tcp.services.<service_name>.loadbalancer.serverstransport`</a> | Allows to reference a ServersTransport resource that is defined either with the File provider or the Kubernetes CRD one.<br/>See [serverstransport](../tcp/serverstransport.md) for more information. | `foobar@file` |

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

You can declare UDP Routers and/or Services using labels.

??? example "Declaring UDP Routers and Services"

    ```yaml
    services:
      my-container:
        # ...
        deploy:
          labels:
            - "traefik.udp.routers.my-router.entrypoints=udp"
            - "traefik.udp.services.my-service.loadbalancer.server.port=4123"
    ```

!!! warning "UDP and HTTP"

    If you declare a UDP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (like it does by default if no UDP Router/Service is defined).
    You can declare both a UDP Router/Service and an HTTP Router/Service for the same container (but you have to do so manually).

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

| Label | Description | Value |
|------|-------------|-------|
| <a id="opt-traefik-enable" href="#opt-traefik-enable" title="#opt-traefik-enable">`traefik.enable`</a> | You can tell Traefik to consider (or not) the container by setting `traefik.enable` to true or false.<br/>This option overrides the value of `exposedByDefault`. | `true` |
| <a id="opt-traefik-swarm-network" href="#opt-traefik-swarm-network" title="#opt-traefik-swarm-network">`traefik.swarm.network`</a> | Overrides the default docker network to use for connections to the container.<br/>If a container is linked to several networks, be sure to set the proper network name (you can check this with `docker inspect <container_id>`), otherwise it will randomly pick one (depending on how docker is returning them).<br/><br/>When deploying a stack from a compose file `stack`, the networks defined are prefixed with `stack`. | `mynetwork` |
| <a id="opt-traefik-swarm-lbswarm" href="#opt-traefik-swarm-lbswarm" title="#opt-traefik-swarm-lbswarm">`traefik.swarm.lbswarm`</a> | Enables Swarm's inbuilt load balancer (only relevant in Swarm Mode).<br/>If you enable this option, Traefik will use the virtual IP provided by docker swarm instead of the containers IPs.<br/>Which means that Traefik will not perform any kind of load balancing and will delegate this task to swarm. | `true` |
