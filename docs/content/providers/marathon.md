# Traefik & Marathon

Traefik can be configured to use Marathon as a provider.
{: .subtitle }

See also [Marathon user guide](../user-guides/marathon.md).

## Configuration Examples

??? example "Configuring Marathon & Deploying / Exposing Applications"

    Enabling the marathon provider

    ```toml tab="File"
    [providers.marathon]
    endpoint = "http://127.0.0.1:8080"
    ```
    
    ```txt tab="CLI"
    --providers.marathon
    --providers.marathon.endpoint="http://127.0.0.1:8080"
    ```

    Attaching labels to marathon applications

    ```json
	{
		"id": "/whoami",
		"container": {
			"type": "DOCKER",
			"docker": {
				"image": "containous/whoami",
				"network": "BRIDGE",
				"portMappings": [
					{
						"containerPort": 80,
						"hostPort": 0,
						"protocol": "tcp"
					}
				]
			}
		},
		"labels": {
			"traefik.http.Routers.app.Rule": "PathPrefix(`/app`)"
		}
	}
    ```

## Provider Configuration Options

!!! tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the [static](../reference/static-configuration/overview.md) and the [dynamic](../reference/dynamic-configuration/marathon.md) configuration references.

### `basic`

_Optional_

Enables Marathon basic authentication.

```toml tab="File"
[marathon.basic]
httpBasicAuthUser = "foo"
httpBasicPassword = "bar"
```

```txt tab="CLI"
--providers.marathon
--providers.marathon.basic.httpbasicauthuser="foo"
--providers.marathon.basic.httpbasicpassword="bar"
```

### `dcosToken`

_Optional_

DCOSToken for DCOS environment.

If set, it overrides the Authorization header.

```toml tab="File"
[providers.marathon]
dcosToken = "xxxxxx"
# ...
```

```txt tab="CLI"
--providers.marathon
--providers.marathon.dcosToken="xxxxxx"
```

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

For a given application if no routing rule was defined by a label, it is defined by this defaultRule instead.

It must be a valid [Go template](https://golang.org/pkg/text/template/),
augmented with the [sprig template functions](http://masterminds.github.io/sprig/).

The app ID can be accessed as the Name identifier,
and the template has access to all the labels defined on this Marathon application.

```toml tab="File"
[providers.marathon]
defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
# ...
```

```txt tab="CLI"
--providers.marathon
--providers.marathon.defaultRule="Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
```

### `dialerTimeout`

_Optional, Default=5s_

Overrides DialerTimeout.

Amount of time the Marathon provider should wait before timing out,
when trying to open a TCP connection to a Marathon master.

Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration),
or directly as a number of seconds.

### `endpoint`

_Optional, Default=http://127.0.0.1:8080_

Marathon server endpoint.

You can optionally specify multiple endpoints:

```toml tab="File"
[providers.marathon]
endpoint = "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
# ...
```

```txt tab="CLI"
--providers.marathon
--providers.marathon.endpoint="http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
```

### `exposedByDefault`

_Optional, Default=true_

Exposes Marathon applications by default through Traefik.

If set to false, applications that don't have a `traefik.enable=true` label will be ignored from the resulting routing configuration.

### `constraints`

_Optional, Default=""_

Constraints is an expression that Traefik matches against the application's labels to determine whether to create any route for that application.
That is to say, if none of the application's labels match the expression, no route for the application is created.
In addition, the expression also matched against the application's constraints, such as described in [Marathon constraints](https://mesosphere.github.io/marathon/docs/constraints.html).
If the expression is empty, all detected applications are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegexp("key", "value")`, as well as the usual boolean logic.
In addition, to match against marathon constraints, the function `MarathonConstraint("field:operator:value")` can be used, where the field, operator, and value parts are joined together in a single string with the `:` separator.

??? example "Constraints Expression Examples"

    ```toml
    # Includes only applications having a label with key `a.label.name` and value `foo`
    constraints = "Label(`a.label.name`, `foo`)"
    ```
    
    ```toml
    # Excludes applications having any label with key `a.label.name` and value `foo`
    constraints = "!Label(`a.label.name`, `value`)"
    ```
    
    ```toml
    # With logical AND.
    constraints = "Label(`a.label.name`, `valueA`) && Label(`another.label.name`, `valueB`)"
    ```
    
    ```toml
    # With logical OR.
    constraints = "Label(`a.label.name`, `valueA`) || Label(`another.label.name`, `valueB`)"
    ```
    
    ```toml
    # With logical AND and OR, with precedence set by parentheses.
    constraints = "Label(`a.label.name`, `valueA`) && (Label(`another.label.name`, `valueB`) || Label(`yet.another.label.name`, `valueC`))"
    ```
    
    ```toml
    # Includes only applications having a label with key `a.label.name` and a value matching the `a.+` regular expression.
    constraints = "LabelRegexp(`a.label.name`, `a.+`)"
    ```

    ```toml
    # Includes only applications having a Marathon constraint with field `A`, operator `B`, and value `C`.
    constraints = "MarathonConstraint(`A:B:C`)"
    ```

    ```toml
    # Uses both Marathon constraint and application label with logical operator.
    constraints = "MarathonConstraint(`A:B:C`) && Label(`a.label.name`, `value`)"
    ```

### `forceTaskHostname`

_Optional, Default=false_

By default, a task's IP address (as returned by the Marathon API) is used as backend server if an IP-per-task configuration can be found;
otherwise, the name of the host running the task is used.
The latter behavior can be enforced by enabling this switch.

### `keepAlive`

_Optional, Default=10s_

Set the TCP Keep Alive interval for the Marathon HTTP Client.
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration),
or directly as a number of seconds.

### `respectReadinessChecks`

_Optional, Default=false_

Applications may define readiness checks which are probed by Marathon during deployments periodically, and these check results are exposed via the API.
Enabling respectReadinessChecks causes Traefik to filter out tasks whose readiness checks have not succeeded.
Note that the checks are only valid at deployment times.

See the Marathon guide for details.

### `responseHeaderTimeout`

_Optional, Default=60s_

Overrides ResponseHeaderTimeout.
Amount of time the Marathon provider should wait before timing out,
when waiting for the first response header from a Marathon master.

Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration), or directly as a number of seconds.

### `TLS`

_Optional_

TLS client configuration. [tls/#Config](https://golang.org/pkg/crypto/tls/#Config).

```toml tab="File"
[marathon.TLS]
CA = "/etc/ssl/ca.crt"
Cert = "/etc/ssl/marathon.cert"
Key = "/etc/ssl/marathon.key"
insecureSkipVerify = true
```

```txt tab="CLI"
--providers.marathon.tls
--providers.marathon.tls.ca="/etc/ssl/ca.crt"
--providers.marathon.tls.cert="/etc/ssl/marathon.cert"
--providers.marathon.tls.key="/etc/ssl/marathon.key"
--providers.marathon.tls.insecureskipverify=true
```

### `TLSHandshakeTimeout`

_Optional, Default=5s_

Overrides TLSHandshakeTimeout.
Amount of time the Marathon provider should wait before timing out,
when waiting for the TLS handshake to complete.
Can be provided in a format supported by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration),
or directly as a number of seconds.

### `trace`

_Optional, Default=false_

Displays additional provider logs (if available).

### `watch`

_Optional, Default=true_

Enables watching for Marathon changes.

## Routing Configuration Options

### General

Traefik creates, for each Marathon application, a corresponding [service](../routing/services/index.md) and [router](../routing/routers/index.md).

The Service automatically gets a server per instance of the application,
and the router automatically gets a rule defined by defaultRule (if no rule for it was defined in labels).

### Routers

To update the configuration of the Router automatically attached to the application,
add labels starting with `traefik.HTTP.Routers.{router-name-of-your-choice}.` and followed by the option you want to change.
For example, to change the routing rule, you could add the label ```traefik.HTTP.Routers.Routername.Rule=Host(`my-domain`)```.

Every [Router](../routing/routers/index.md) parameter can be updated this way.

### Services

To update the configuration of the Service automatically attached to the container,
add labels starting with `traefik.HTTP.Services.{service-name-of-your-choice}.`, followed by the option you want to change.
For example, to change the passhostheader behavior, you'd add the label `traefik.HTTP.Services.Servicename.LoadBalancer.PassHostHeader=false`.

Every [Service](../routing/services/index.md) parameter can be updated this way.

### Middleware

You can declare pieces of middleware using labels starting with `traefik.HTTP.Middlewares.{middleware-name-of-your-choice}.`, followed by the middleware type/options.
For example, to declare a middleware [`redirectscheme`](../middlewares/redirectscheme.md) named `my-redirect`, you'd write `traefik.HTTP.Middlewares.my-redirect.RedirectScheme.Scheme: https`.

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

More information about available middlewares in the dedicated [middlewares section](../middlewares/overview.md).

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

### Specific Options

#### `traefik.enable`

Setting this option controls whether Traefik exposes the application.
It overrides the value of `exposedByDefault`.

#### `traefik.marathon.ipadressidx`

If a task has several IP addresses, this option specifies which one, in the list of available addresses, to select.
