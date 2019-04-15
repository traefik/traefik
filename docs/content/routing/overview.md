# Overview

What's Happening to the Requests?
{: .subtitle }

Let's zoom on Traefik's architecture and talk about the components that enable the routes to be created.

First, when you start Traefik, you define [entrypoints](../entrypoints) (in their most basic forms, they are port numbers).
Then, connected to these entrypoints, [routers](../routers) analyze the incoming requests to see if they match a set of [rules](../routers#rule).
If they do, the router might transform the request using pieces of [middleware](../middlewares/overview.md) before forwarding them to your [services](./services/index.md).

![Architecture](../assets/img/architecture-overview.png)

## Clear Responsibilities

- [_Providers_](../providers/overview.md) discover the services that live on your infrastructure (their IP, health, ...)
- [_Entrypoints_](./entrypoints.md) listen for incomming traffic (ports, ...)
- [_Routers_](./routers/index.md) analyse the requests (host, path, headers, SSL, ...)
- [_Services_](./services/index.md) forward the request to your services (load balancing, ...)
- [_Middlewares_](../middlewares/overview.md) may update the request or make decisions based on the request (authentication, rate limiting, headers, ...)

## Example with a File Provider

Below is an example of a full configuration file for the [file provider](../providers/file.md) that forwards `http://domain/whoami/` requests to a service reachable on `http://private/whoami-service/`.
In the process, Traefik will make sure that the user is authenticated (using the [BasicAuth middleware](../middlewares/basicauth.md)).

```toml
[entryPoints]
   [entryPoints.web]
      address = ":8081" # Listen on port 8081 for incoming requests

[providers]
   [providers.file] # Enable the file provider to define routers / middlewares / services in a file

[http] # http routing section
    [http.routers]
       [http.routers.to-whoami] # Define a connection between requests and services
          rule = "Host(domain) && PathPrefix(/whoami/)"
          middlewares = ["test-user"] # If the rule matches, applies the middleware
          service = "whoami" # If the rule matches, forward to the whoami service (declared below)

    [http.middlewares]
       [http.middlewares.test-user.basicauth] # Define an authentication mechanism
          users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]

    [http.services]
       [http.services.whoami.loadbalancer] # Define how to reach an existing service on our infrastructure
          [[http.services.whoami.loadbalancer.servers]]
             url = "http://private/whoami-service"
```

!!! note "The File Provider"

    In this example, we use the [file provider](../providers/file.md).
    Even if it is one of the least magical way of configuring Traefik, it explicitly describes every available notion.

!!! note "HTTP / TCP"

    In this example, we've defined routing rules for http requests only.
    Traefik also supports TCP requests. To add [TCP routers](./routers/index.md) and [TCP services](./services/index.md), declare them in a TCP section like in the following.

    ??? example "Adding a TCP route for TLS requests on whoami.traefik.io"

        ```toml
        [entryPoints]
           [entryPoints.web]
              address = ":8081" # Listen on port 8081 for incoming requests

        [providers]
           [providers.file] # Enable the file provider to define routers / middlewares / services in a file

        [http] # http routing section
            [http.routers]
               [http.routers.to-whoami] # Define a connection between requests and services
                  rule = "Host(`domain`) && PathPrefix(/whoami/)"
                  middlewares = ["test-user"] # If the rule matches, applies the middleware
                  service = "whoami" # If the rule matches, forward to the whoami service (declared below)

            [http.middlewares]
               [http.middlewares.test-user.basicauth] # Define an authentication mechanism
                  users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]

            [http.services]
               [http.services.whoami.loadbalancer] # Define how to reach an existing service on our infrastructure
                  [[http.services.whoami.loadbalancer.servers]]
                     url = "http://private/whoami-service"

           [tcp]
              [tcp.routers]
                 [tcp.routers.to-whoami-tcp]
                     rule = "HostSNI(`whoami-tcp.traefik.io`)"
                     service = "whoami-tcp"
                     [tcp.routers.to-whoami-tcp.tls]

              [tcp.services]
                 [tcp.services.whoami-tcp.loadbalancer]
                    [[tcp.services.whoami-tcp.loadbalancer.servers]]
                       address = "xx.xx.xx.xx:xx"
        ```
