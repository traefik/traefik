# Routers

Connecting Requests to Services
{: .subtitle }

![Routers](../img/routers.png)

A router is in charge of connecting incoming requests to the services that can handle them. In the process, routers may use pieces of [middleware](../middlewares/overview.md) to update the request, or act before forwarding the request to the service.

## Configuration Example

??? example "Requests /foo are Handled by service-foo -- Using the [File Provider](../providers/file.md)"

    ```toml
      [Routers]
        [Routers.my-router]
        rule = "Path(/foo)"
        service = "service-foo"
    ```

??? example "With a [Middleware](../middlewares/overview.md) -- using the [File Provider](../providers/file.md)"

    ```toml
      [Routers]
        [Routers.my-router]
        rule = "Path(/foo)"
        middlewares = ["authentication"] # declared elsewhere
        service = "service-foo"
    ```

## Configuration

### EntryPoints

If not specified, routers will accept requests from all defined entrypoints.
If you want to limit the router scope to a set of entrypoint, set the entrypoints option.

??? example "Listens to Every EntryPoint"

    ```toml
    [EntryPoints]
       [EntryPoint.http]
          # ...
       [EntryPoint.https]
          # ...   
       [EntryPoint.other]
          # ...
          
    [Routers]
       [Routers.Router-1]
          # By default, routers listen to every entrypoints
          rule = "Host(traefik.io)"
          service = "service-1"      
    ```
    
??? example "Listens to Specific EntryPoints"

    ```toml
    [EntryPoints]
       [EntryPoint.http]
          # ...
       [EntryPoint.https]
          # ...   
       [EntryPoint.other]
          # ...
          
    [Routers]
       [Routers.Router-1]
          entryPoints = ["https", "other"] # won't listen to entrypoint http
          rule = "Host(traefik.io)"
          service = "service-1"      
    ```

### Rule

Rules are a set of matchers that determine if a particular request matches specific criterias. 
If the rule is verified, then the router becomes active and calls middlewares, then forward the request to the service.

??? example "Host is traefik.io"

    ```
       rule = "Host(`traefik.io`)" 
    ```

??? example "Host is traefik.io OR Host is containo.us AND path is /traefik"

    ```
       rule = "Host(`traefik.io`) || (Host(`containo.us`) && Path(`/traefik`))"
    ```
The table below lists all the available matchers:

| Rule                                                    | Description                                                                                                                                                                                                                                                                             |
|------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ``Headers(`key`, `value`)``                  | Check if there is a key `key`defined in the headers, with the value `value`                                                                                                                                                                               |
| ``HeadersRegexp(`key`, `regexp`)``     | Check if there is a key `key`defined in the headers, with a value that matches the regular expression `regexp`                                                                                                                                  |
| ``Host(`domain-1`, ...)``                         | Check if the request domain targets one of the given `domains`.                                                                                                                                                                                                                              |
| ``HostRegexp(`traefik.io`, `{subdomain:[a-z]+}.traefik.io`, ...)``    | Check if the request domain matches the given `regexp`.                                                                                                                                                                                                      |
| `Method(methods, ...)`                                   | Check if the request method is one of the given `methods` (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`)                                                                                                                                                                                                                       |
| ``Path(`path`, `/articles/{category}/{id:[0-9]+}`, ...)``       | Match exact request path. It accepts a sequence of literal and regular expression paths.                                                                                                                                                                                                |
| ``PathStrip(`/path/`)``                                    | Match exact path and strip off the path prior to forwarding the request to the backend. It accepts a sequence of literal paths.                                                                                                                                                         |
| ``PathStripRegex(`/articles/{category}/{id:[0-9]+})``         | Match exact path and strip off the path prior to forwarding the request to the backend. It accepts a sequence of literal and regular expression paths.                                                                                                                                  |
| ``PathPrefix(`/products/`, `/articles/{category}/{id:[0-9]+}`)`` | Match request prefix path. It accepts a sequence of literal and regular expression prefix paths.                                                                                                                                                                                        |
| ``PathPrefixStrip(`/products/`)``                              | Match request prefix path and strip off the path prefix prior to forwarding the request to the service. It accepts a sequence of literal prefix paths. Starting with Traefik 1.3, the stripped prefix path will be available in the `X-Forwarded-Prefix` header.                        |
| ``PathPrefixStripRegex(`/articles/{category}/{id:[0-9]+}`)``   | Match request prefix path and strip off the path prefix prior to forwarding the request to the service. It accepts a sequence of literal and regular expression prefix paths. Starting with Traefik 1.3, the stripped prefix path will be available in the `X-Forwarded-Prefix` header. |
| ``Query(`foo=bar`, `bar=baz`)``                                  | Match` Query String parameters. It accepts a sequence of key=value pairs.                                                                                                                                                                                                                |

!!! important "Regexp Syntax"

    In order to use regular expressions with `Host` and `Path` expressions, you must declare an arbitrarily named variable followed by the colon-separated regular expression, all enclosed in curly braces. Any pattern supported by [Go's regexp package](https://golang.org/pkg/regexp/) may be used (example: `/posts/{id:[0-9]+}`).
    

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

You can attach a list of [middlewares](../middlewares/overview.md) to the routers.
The middlewares will take effect only if the rule matches, and before forwarding the request to the service. 

### Service

You must attach a [service](./services.md) per router.
Services are the target for the router.