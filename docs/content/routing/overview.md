# Overview

What's Happening to the Requests?
{: .subtitle }

Let's zoom on Traefik's architecture and talk about the components that enable the routes to be created.

First, when you start Traefik, you define [entrypoints](./entrypoints.md) (in their most basic forms, they are port numbers). Then, connected to these entrypoints, [routers](./routers.md) analyze the incoming requests to see if they match a set of [rules](../routers#rule). If they do, the router might transform the request using pieces of [middleware](../middlewares/overview.md) before forwarding them to your [services](./services.md).


![Architecture](../img/architecture-overview.png)

## Clear Responsibilities

 - [_Providers_](../providers/overview.md) discover the services that live on your infrastructure (their IP, health, ...) 
 - [_Entrypoints_](./entrypoints.md) listen for incomming traffic (ports, SSL, ...)
 - [_Routers_](./routers.md) analyse the requests (host, path, headers, ...)
 - [_Services_](./services.md) forward the request to your services (load balancing, ...)
 - [_Middlewares_](../middlewares/overview.md) may update the request or make decisions based on the request (authentication, rate limiting, headers, ...)
 
## Example with a File Provider

Below is an example of a full configuration file for the [file provider](../providers/file.md) that forwards `http://domain/whoami/` requests to a service reachable on `http://private/whoami-service/`. In the process, Traefik will make sure that the user is authenticated (using the [BasicAuth middleware](../middlewares/basicauth.md)).     

```toml
[EntryPoints]
   [EntryPoints.http]
      address = ":8081" # Listen on port 8081 for incoming requests

[Providers]
   # Enable the file provider to define routers / middlewares / services in a file
   [Providers.file] 

[Routers]
   [Routers.to-whoami] # Define a connection between requests and services
      rule = "Host(domain) && PathPrefix(/whoami/)"
      middlewares = ["test-user"] # If the rule matches, applies the middleware
      service = "whoami" # If the rule matches, forward to the whoami service (declared below)
      
[Middlewares]
   [Middlewares.test-user.basicauth] # Define an authentication mechanism
      users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]
      
[Services]
   [Services.whoami.loadbalancer] # Define how to reach an existing service on our infrastructure
      [[Services.whoami.loadbalancer.servers]]
         url = "http://private/whoami-service"
```

!!! note "The File Provider"

    In this example, we use the [file provider](../providers/file.md). Even if it is one of the least magical way of configuring Traefik, it explicitely describes every available notion.
