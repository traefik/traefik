# Traefik & Rancher

A Story of Labels, Services & Containers
{: .subtitle }

![Rancher](../../assets/img/providers/rancher.png)

Attach labels to your services and let Traefik do the rest!

!!! important "This provider is specific to Rancher 1.x."
    
    Rancher 2.x requires Kubernetes and does not have a metadata endpoint of its own for Traefik to query.
    As such, Rancher 2.x users should utilize the [Kubernetes provider](./kubernetes-crd.md) directly.

## Routing Configuration

!!! info "Labels"
    
    - Labels are case insensitive.
    - The complete list of labels can be found [the reference page](../../reference/dynamic-configuration/rancher.md)

### General

Traefik creates, for each rancher service, a corresponding [service](../services/index.md) and [router](../routers/index.md).

The Service automatically gets a server per container in this rancher service, and the router gets a default rule attached to it, based on the service name.

### Routers

To update the configuration of the Router automatically attached to the container, add labels starting with `traefik.routers.{name-of-your-choice}.` and followed by the option you want to change.

For example, to change the rule, you could add the label ```traefik.http.routers.my-container.rule=Host(`mydomain.com`)```.

??? info "`traefik.http.routers.<router_name>.rule`"
    
    See [rule](../routers/index.md#rule) for more information. 
    
    ```yaml
    - "traefik.http.routers.myrouter.rule=Host(`mydomain.com`)"
    ```

??? info "`traefik.http.routers.<router_name>.entrypoints`"
    
    See [entry points](../routers/index.md#entrypoints) for more information. 
    
    ```yaml
    - "traefik.http.routers.myrouter.entrypoints=web,websecure"
    ```

??? info "`traefik.http.routers.<router_name>.middlewares`"
    
    See [middlewares](../routers/index.md#middlewares) and [middlewares overview](../../middlewares/overview.md) for more information. 
    
    ```yaml
    - "traefik.http.routers.myrouter.middlewares=auth,prefix,cb"
    ```

??? info "`traefik.http.routers.<router_name>.service`"
    
    See [rule](../routers/index.md#service) for more information. 
    
    ```yaml
    - "traefik.http.routers.myrouter.service=myservice"
    ```

??? info "`traefik.http.routers.<router_name>.tls`"
    
    See [tls](../routers/index.md#tls) for more information.
    
    ```yaml
    - "traefik.http.routers.myrouter>.tls=true"
    ```

??? info "`traefik.http.routers.<router_name>.tls.certresolver`"
    
    See [certResolver](../routers/index.md#certresolver) for more information.
    
    ```yaml
    - "traefik.http.routers.myrouter.tls.certresolver=myresolver"
    ```

??? info "`traefik.http.routers.<router_name>.tls.domains[n].main`"
    
    See [domains](../routers/index.md#domains) for more information.
    
    ```yaml
    - "traefik.http.routers.myrouter.tls.domains[0].main=foobar.com"
    ```

??? info "`traefik.http.routers.<router_name>.tls.domains[n].sans`"
    
    See [domains](../routers/index.md#domains) for more information.
    
    ```yaml
    - "traefik.http.routers.myrouter.tls.domains[0].sans=test.foobar.com,dev.foobar.com"
    ```

??? info "`traefik.http.routers.<router_name>.tls.options`"
    
    See [options](../routers/index.md#options) for more information.
    
    ```yaml
    - "traefik.http.routers.myrouter.tls.options=foobar"
    ```

??? info "`traefik.http.routers.<router_name>.priority`"
    <!-- TODO doc priority in routers page -->
    
    ```yaml
    - "traefik.http.routers.myrouter.priority=42"
    ```

### Services

To update the configuration of the Service automatically attached to the container,
add labels starting with `traefik.http.services.{name-of-your-choice}.`, followed by the option you want to change.

For example, to change the `passHostHeader` behavior,
you'd add the label `traefik.http.services.{name-of-your-choice}.loadbalancer.passhostheader=false`.

??? info "`traefik.http.services.<service_name>.loadbalancer.server.port`"
    
    Registers a port.
    Useful when the container exposes multiples ports.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.server.port=8080"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.server.scheme`"
    
    Overrides the default scheme.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.server.scheme=http"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.passhostheader`"
    <!-- TODO doc passHostHeader in services page -->
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.passhostheader=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.headers.<header_name>`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.headers.X-Foo=foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.hostname`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.hostname=foobar.com"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.interval`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.interval=10"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.path`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.path=/foo"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.port`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.port=42"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.scheme`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.scheme=http"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.healthcheck.timeout`"
    
    See [health check](../services/index.md#health-check) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.healthcheck.timeout=10"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky`"
    
    See [sticky sessions](../services/index.md#sticky-sessions) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.httponly`"
    
    See [sticky sessions](../services/index.md#sticky-sessions) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.httponly=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.name`"
    
    See [sticky sessions](../services/index.md#sticky-sessions) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.name=foobar"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.sticky.cookie.secure`"
    
    See [sticky sessions](../services/index.md#sticky-sessions) for more information.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.sticky.cookie.secure=true"
    ```

??? info "`traefik.http.services.<service_name>.loadbalancer.responseforwarding.flushinterval`"
    <!-- TODO doc responseforwarding in services page -->
    
    FlushInterval specifies the flush interval to flush to the client while copying the response body.
    
    ```yaml
    - "traefik.http.services.myservice.loadbalancer.responseforwarding.flushinterval=10"
    ```

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.{name-of-your-choice}.`, followed by the middleware type/options.

For example, to declare a middleware [`redirectscheme`](../../middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme: https`.

More information about available middlewares in the dedicated [middlewares section](../../middlewares/overview.md).

??? example "Declaring and Referencing a Middleware"
    
    ```yaml
    # ...
    labels:
      # Declaring a middleware
      - traefik.http.middlewares.my-redirect.redirectscheme.scheme=https
      # Referencing a middleware
      - traefik.http.routers.my-container.middlewares=my-redirect
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

### Specific Provider Options

#### `traefik.enable`

```yaml
- "traefik.enable=true"
```

You can tell Traefik to consider (or not) the container by setting `traefik.enable` to true or false.

This option overrides the value of `exposedByDefault`.

#### Port Lookup

Traefik is now capable of detecting the port to use, by following the default rancher flow.
That means, if you just expose lets say port `:1337` on the rancher ui, traefik will pick up this port and use it.
