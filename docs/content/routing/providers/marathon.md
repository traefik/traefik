# Traefik & Marathon

Traefik can be configured to use Marathon as a provider.
{: .subtitle }

See also [Marathon user guide](../../user-guides/marathon.md).

## Routing Configuration

!!! info "Labels"
    
    - Labels are case insensitive.
    - The complete list of labels can be found [the reference page](../../reference/dynamic-configuration/marathon.md)

### General

Traefik creates, for each Marathon application, a corresponding [service](../services/index.md) and [router](../routers/index.md).

The Service automatically gets a server per instance of the application,
and the router automatically gets a rule defined by defaultRule (if no rule for it was defined in labels).

### Routers

To update the configuration of the Router automatically attached to the application,
add labels starting with `traefik.http.routers.{router-name-of-your-choice}.` and followed by the option you want to change.

For example, to change the routing rule, you could add the label ```traefik.http.routers.routername.rule=Host(`mydomain.com`)```.

Every [Router](../routers/index.md) parameter can be updated this way.

### Services

To update the configuration of the Service automatically attached to the container,
add labels starting with `traefik.http.services.{service-name-of-your-choice}.`, followed by the option you want to change.

For example, to change the passHostHeader behavior, you'd add the label `traefik.http.services.servicename.loadbalancer.passhostheader=false`.

Every [Service](../services/index.md) parameter can be updated this way.

### Middleware

You can declare pieces of middleware using labels starting with `traefik.http.middlewares.{middleware-name-of-your-choice}.`, followed by the middleware type/options.
For example, to declare a middleware [`redirectscheme`](../../middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.http.middlewares.my-redirect.redirectscheme.scheme: https`.

??? example "Declaring and Referencing a Middleware"

    ```json
	{
		...
		"labels": {
			"traefik.http.middlewares.my-redirect.redirectscheme.scheme": "https",
			"traefik.http.routers.my-container.middlewares": "my-redirect"
		}
	}
    ```

!!! warning "Conflicts in Declaration"

    If you declare multiple middleware with the same name but with different parameters, the middleware fails to be declared.

More information about available middlewares in the dedicated [middlewares section](../../middlewares/overview.md).

### TCP

You can declare TCP Routers and/or Services using labels.

??? example "Declaring TCP Routers and Services"

    ```json
	{
		...
		"labels": {
			"traefik.tcp.routers.my-router.rule": "HostSNI(`my-host.com`)",
			"traefik.tcp.routers.my-router.tls": "true",
			"traefik.tcp.services.my-service.loadbalancer.server.port": "4123"
		}
	}
    ```

!!! warning "TCP and HTTP"

    If you declare a TCP Router/Service, it will prevent Traefik from automatically creating an HTTP Router/Service (as it would by default if no TCP Router/Service is defined).
    Both a TCP Router/Service and an HTTP Router/Service can be created for the same application, but it has to be done explicitly in the config.

### Specific Provider Options

#### `traefik.enable`

Setting this option controls whether Traefik exposes the application.
It overrides the value of `exposedByDefault`.

#### `traefik.marathon.ipadressidx`

If a task has several IP addresses, this option specifies which one, in the list of available addresses, to select.
