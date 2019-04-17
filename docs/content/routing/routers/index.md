# Routers

Connecting Requests to Services
{: .subtitle }

![routers](../../assets/img/routers.png)

A router is in charge of connecting incoming requests to the services that can handle them.
In the process, routers may use pieces of [middleware](../../middlewares/overview.md) to update the request, or act before forwarding the request to the service.

## Configuration Example

??? example "Requests /foo are Handled by service-foo -- Using the [File Provider](../../providers/file.md)"

    ```toml
      [http.routers]
        [http.routers.my-router]
        rule = "Path(`/foo`)"
        service = "service-foo"
    ```

??? example "With a [middleware](../../middlewares/overview.md) -- using the [File Provider](../../providers/file.md)"

    ```toml
      [http.routers]
        [http.routers.my-router]
        rule = "Path(`/foo`)"
        middlewares = ["authentication"] # declared elsewhere
        service = "service-foo"
    ```

??? example "Forwarding all (non-tls) requests on port 3306 to a database service"

    ```toml
      [entryPoints]
        [entryPoints.mysql-default]
          address = ":80"
        [entryPoints.mysql-default]
          address = ":3306"
          
      [tcp]
        [tcp.routers]
          [tcp.routers.to-database]
            entryPoints = ["mysql-default"]
            rule = "HostSNI(`*`)" # Catch every request (only available rule for non-tls routers. See below.)
            service = "database"
    ```

## Configuring HTTP Routers

### EntryPoints

If not specified, HTTP routers will accept requests from all defined entrypoints.
If you want to limit the router scope to a set of entrypoint, set the entrypoints option.

??? example "Listens to Every EntryPoint"

    ```toml
    [entryPoints]
       [entryPoints.web]
          # ...
       [entryPoints.web-secure]
          # ...
       [entryPoints.other]
          # ...

    [http.routers]
       [http.routers.Router-1]
          # By default, routers listen to every entrypoints
          rule = "Host(`traefik.io`)"
          service = "service-1"
    ```

??? example "Listens to Specific EntryPoints"

    ```toml
    [entryPoints]
       [entryPoints.web]
          # ...
       [entryPoints.web-secure]
          # ...
       [entryPoints.other]
          # ...

    [http.routers]
       [http.routers.Router-1]
          entryPoints = ["web-secure", "other"] # won't listen to entrypoint web
          rule = "Host(`traefik.io`)"
          service = "service-1"
    ```

### Rule

Rules are a set of matchers that determine if a particular request matches specific criteria.
If the rule is verified, then the router becomes active and calls middlewares, then forward the request to the service.

??? example "Host is traefik.io"

    ```toml
       rule = "Host(`traefik.io`)"
    ```

??? example "Host is traefik.io OR Host is containo.us AND path is /traefik"

    ```toml
       rule = "Host(`traefik.io`) || (Host(`containo.us`) && Path(`/traefik`))"
    ```
The table below lists all the available matchers:

| Rule                                                               | Description                                                                                                    |
|--------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| ``Headers(`key`, `value`)``                                        | Check if there is a key `key`defined in the headers, with the value `value`                                    |
| ``HeadersRegexp(`key`, `regexp`)``                                 | Check if there is a key `key`defined in the headers, with a value that matches the regular expression `regexp` |
| ``Host(`domain-1`, ...)``                                          | Check if the request domain targets one of the given `domains`.                                                |
| ``HostRegexp(`traefik.io`, `{subdomain:[a-z]+}.traefik.io`, ...)`` | Check if the request domain matches the given `regexp`.                                                        |
| `Method(methods, ...)`                                             | Check if the request method is one of the given `methods` (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`)            |
| ``Path(`path`, `/articles/{category}/{id:[0-9]+}`, ...)``          | Match exact request path. It accepts a sequence of literal and regular expression paths.                       |
| ``PathPrefix(`/products/`, `/articles/{category}/{id:[0-9]+}`)``   | Match request prefix path. It accepts a sequence of literal and regular expression prefix paths.               |
| ``Query(`foo=bar`, `bar=baz`)``                                    | Match` Query String parameters. It accepts a sequence of key=value pairs.                                      |

!!! important "Regexp Syntax"

    In order to use regular expressions with `Host` and `Path` expressions,
    you must declare an arbitrarily named variable followed by the colon-separated regular expression, all enclosed in curly braces.
    Any pattern supported by [Go's regexp package](https://golang.org/pkg/regexp/) may be used (example: `/posts/{id:[0-9]+}`).

!!! tip "Combining Matchers Using Operators and Parenthesis"

    You can combine multiple matchers using the AND (`&&`) and OR (`||) operators. You can also use parenthesis.

!!! important "Rule, Middleware, and Services"

    The rule is evaluated "before" any middleware has the opportunity to work, and "before" the request is forwarded to the service.

!!! tip "Path Vs PathPrefix"

    Use `Path` if your service listens on the exact path only. For instance, `Path: /products` would match `/products` but not `/products/shoes`.

    Use a `*Prefix*` matcher if your service listens on a particular base path but also serves requests on sub-paths.
    For instance, `PathPrefix: /products` would match `/products` but also `/products/shoes` and `/products/shirts`.
    Since the path is forwarded as-is, your service is expected to listen on `/products`.

### Middlewares

You can attach a list of [middlewares](../../middlewares/overview.md) to each HTTP router.
The middlewares will take effect only if the rule matches, and before forwarding the request to the service.

### Service

You must attach a [service](../services/index.md) per router.
Services are the target for the router.

!!! note "HTTP Only"

    HTTP routers can only target HTTP services (not TCP services).

### TLS

When specifying a TLS section, you tell Traefik that the current router is dedicated to HTTPS requests only (and that the router should ignore HTTP (non tls) requests).
Traefik will terminate the SSL connections (meaning that it will send decrypted data to the services).

??? example "Configuring the router to accept HTTPS requests only"

    ```toml
    [http.routers]
       [http.routers.Router-1]
          rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
          service = "service-id"
          [http.routers.Router-1.tls] # will terminate the TLS request
    ```

!!! note "HTTPS & ACME"

    In the current version, with [ACME](../../https-tls/acme.md) enabled, automatic certificate generation will apply to every router declaring a TLS section.
    In the near future, options will be available to enable fine-grain control of the TLS parameters.

!!! note "Passthrough"

    On TCP routers, you can configure a passthrough option so that Traefik doesn't terminate the TLS connection.

!!! important "Routers for HTTP & HTTPS"

    If you need to define the same route for both HTTP and HTTPS requests, you will need to define two different routers: one with the tls section, one without.

    ??? example "HTTP & HTTPS routes"

        ```toml
        [http.routers]
           [http.routers.Router-1-https]
              rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
              service = "service-id"
              [http.routers.Router-1.tls] # will terminate the TLS request

           [http.routers.Router-1-http]
              rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
              service = "service-id"
        ```

## Configuring TCP Routers

### General

If both HTTP routers and TCP routers listen to the same entrypoints, the TCP routers will apply *before* the HTTP routers.
If no matching route is found for the TCP routers, then the HTTP routers will take over.

### EntryPoints

If not specified, TCP routers will accept requests from all defined entrypoints.
If you want to limit the router scope to a set of entrypoints, set the entrypoints option.

??? example "Listens to Every EntryPoint"

    ```toml
    [entryPoints]
       [entryPoints.web]
          # ...
       [entryPoints.web-secure]
          # ...
       [entryPoints.other]
          # ...

    [tcp.routers]
       [tcp.routers.Router-1]
          # By default, routers listen to every entrypoints
          rule = "HostSNI(`traefik.io`)"
          service = "service-1"
          [tcp.routers.Router-1.tls] # will route TLS requests (and ignore non tls requests)
    ```

??? example "Listens to Specific EntryPoints"

    ```toml
    [entryPoints]
       [entryPoints.web]
          # ...
       [entryPoints.web-secure]
          # ...
       [entryPoints.other]
          # ...

    [tcp.routers]
       [tcp.routers.Router-1]
          entryPoints = ["web-secure", "other"] # won't listen to entrypoint web
          rule = "HostSNI(`traefik.io`)"
          service = "service-1"
          [tcp.routers.Router-1.tls] # will route TLS requests (and ignore non tls requests)
    ```

### Rule

| Rule                         | Description                                                             |
|------------------------------|-------------------------------------------------------------------------|
| ``HostSNI(`domain-1`, ...)`` | Check if the Server Name Indication corresponds to the given `domains`. |

!!! important "HostSNI & TLS"

    It is important to note that the Server Name Indication is an extension of the TLS protocol.
    Hence, only TLS routers will be able to specify a domain name with that rule.
    However, non-TLS routers will have to explicitly use that rule with `*` (every domain) to state that every non-TLS request will be handled by the router.

### Services

You must attach a TCP [service](../services/index.md) per TCP router.
Services are the target for the router.

!!! note "TCP Only"

    TCP routers can only target TCP services (not HTTP services).

### TLS

When specifying a TLS section, you tell Traefik that the current router is dedicated to TLS requests only (and that the router should ignore non-tls requests).
By default, Traefik will terminate the SSL connections (meaning that it will send decrypted data to the services), but you can tell Traefik that the request should pass through (keeping the encrypted data) and be forwarded to the service "as is". 

??? example "Configuring TLS Termination"

    ```toml
    [tcp.routers]
       [tcp.routers.Router-1]
          rule = "Host(`foo-domain`)"
          service = "service-id"
          [tcp.routers.Router-1.tls] # will terminate the TLS request by default
    ```

??? example "Configuring passthrough"

    ```toml
    [tcp.routers]
       [tcp.routers.Router-1]
          rule = "Host(`foo-domain`)"
          service = "service-id"
          [tcp.routers.Router-1.tls]
             passthrough=true
    ```

!!! note "TLS & ACME"

    In the current version, with [ACME](../../https-tls/acme.md) enabled, automatic certificate generation will apply to every router declaring a TLS section.
    In the near future, options will be available to enable fine-grain control of the TLS parameters.
