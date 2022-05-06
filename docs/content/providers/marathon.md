---
title: "Traefik Configuration for Marathon"
description: "Traefik Proxy can be configured to use Marathon as a provider. Read the technical documentation to learn how."
---

# Traefik & Marathon

Traefik can be configured to use Marathon as a provider.
{: .subtitle }

For additional information, refer to [Marathon user guide](../user-guides/marathon.md).

## Configuration Examples

??? example "Configuring Marathon & Deploying / Exposing Applications"

    Enabling the Marathon provider

    ```yaml tab="File (YAML)"
    providers:
      marathon: {}
    ```

    ```toml tab="File (TOML)"
    [providers.marathon]
    ```

    ```bash tab="CLI"
    --providers.marathon=true
    ```

    Attaching labels to Marathon applications

    ```json
	{
		"id": "/whoami",
		"container": {
			"type": "DOCKER",
			"docker": {
				"image": "traefik/whoami",
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

## Routing Configuration

See the dedicated section in [routing](../routing/providers/marathon.md).

## Provider Configuration

### `basic`

_Optional_

Enables Marathon basic authentication.

```yaml tab="File (YAML)"
providers:
  marathon:
    basic:
      httpBasicAuthUser: foo
      httpBasicPassword: bar
```

```toml tab="File (TOML)"
[providers.marathon.basic]
  httpBasicAuthUser = "foo"
  httpBasicPassword = "bar"
```

```bash tab="CLI"
--providers.marathon.basic.httpbasicauthuser=foo
--providers.marathon.basic.httpbasicpassword=bar
```

### `dcosToken`

_Optional_

Datacenter Operating System (DCOS) Token for DCOS environment.

If set, it overrides the Authorization header.

```toml tab="File (YAML)"
providers:
  marathon:
    dcosToken: "xxxxxx"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  dcosToken = "xxxxxx"
  # ...
```

```bash tab="CLI"
--providers.marathon.dcosToken=xxxxxx
```

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

The default host rule for all services.

For a given application, if no routing rule was defined by a label, it is defined by this `defaultRule` instead.

It must be a valid [Go template](https://pkg.go.dev/text/template/),
and can include [sprig template functions](https://masterminds.github.io/sprig/).

The app ID can be accessed with the `Name` identifier,
and the template has access to all the labels defined on this Marathon application.

```yaml tab="File (YAML)"
providers:
  marathon:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```bash tab="CLI"
--providers.marathon.defaultRule=Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)
# ...
```

### `dialerTimeout`

_Optional, Default=5s_

Amount of time the Marathon provider should wait before timing out,
when trying to open a TCP connection to a Marathon master.

The value of `dialerTimeout` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  marathon:
    dialerTimeout: "10s"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  dialerTimeout = "10s"
  # ...
```

```bash tab="CLI"
--providers.marathon.dialerTimeout=10s
```

### `endpoint`

_Optional, Default=http://127.0.0.1:8080_

Marathon server endpoint.

You can optionally specify multiple endpoints.

```yaml tab="File (YAML)"
providers:
  marathon:
    endpoint: "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  endpoint = "http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080"
  # ...
```

```bash tab="CLI"
--providers.marathon.endpoint=http://10.241.1.71:8080,10.241.1.72:8080,10.241.1.73:8080
```

### `exposedByDefault`

_Optional, Default=true_

Exposes Marathon applications by default through Traefik.

If set to `false`, applications that do not have a `traefik.enable=true` label are ignored from the resulting routing configuration.

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  marathon:
    exposedByDefault: false
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  exposedByDefault = false
  # ...
```

```bash tab="CLI"
--providers.marathon.exposedByDefault=false
# ...
```

### `constraints`

_Optional, Default=""_

The `constraints` option can be set to an expression that Traefik matches against the application labels to determine whether
to create any route for that application. If none of the application labels match the expression, no route for that application is
created. In addition, the expression is also matched against the application constraints, such as described
in [Marathon constraints](https://mesosphere.github.io/marathon/docs/constraints.html).
If the expression is empty, all detected applications are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegex("key", "value")` functions, as well as the usual boolean logic.
In addition, to match against Marathon constraints, the function `MarathonConstraint("field:operator:value")` can be used, where the field, operator, and value parts are concatenated in a single string using the `:` separator.

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
    constraints = "LabelRegex(`a.label.name`, `a.+`)"
    ```

    ```toml
    # Includes only applications having a Marathon constraint with field `A`, operator `B`, and value `C`.
    constraints = "MarathonConstraint(`A:B:C`)"
    ```

    ```toml
    # Uses both Marathon constraint and application label with logical operator.
    constraints = "MarathonConstraint(`A:B:C`) && Label(`a.label.name`, `value`)"
    ```

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  marathon:
    constraints: "Label(`a.label.name`,`foo`)"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  constraints = "Label(`a.label.name`,`foo`)"
  # ...
```

```bash tab="CLI"
--providers.marathon.constraints=Label(`a.label.name`,`foo`)
# ...
```

### `forceTaskHostname`

_Optional, Default=false_

By default, the task IP address (as returned by the Marathon API) is used as backend server if an IP-per-task configuration can be found;
otherwise, the name of the host running the task is used.
The latter behavior can be enforced by setting this option to `true`.

```yaml tab="File (YAML)"
providers:
  marathon:
    forceTaskHostname: true
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  forceTaskHostname = true
  # ...
```

```bash tab="CLI"
--providers.marathon.forceTaskHostname=true
# ...
```

### `keepAlive`

_Optional, Default=10s_

Set the TCP Keep Alive duration for the Marathon HTTP Client.
The value of `keepAlive` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  marathon:
    keepAlive: "30s"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  keepAlive = "30s"
  # ...
```

```bash tab="CLI"
--providers.marathon.keepAlive=30s
# ...
```

### `respectReadinessChecks`

_Optional, Default=false_

Applications may define readiness checks which are probed by Marathon during deployments periodically, and these check results are exposed via the API.
Enabling `respectReadinessChecks` causes Traefik to filter out tasks whose readiness checks have not succeeded.
Note that the checks are only valid during deployments.

See the Marathon guide for details.

```yaml tab="File (YAML)"
providers:
  marathon:
    respectReadinessChecks: true
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  respectReadinessChecks = true
  # ...
```

```bash tab="CLI"
--providers.marathon.respectReadinessChecks=true
# ...
```

### `responseHeaderTimeout`

_Optional, Default=60s_

Amount of time the Marathon provider should wait before timing out when waiting for the first response header
from a Marathon master.

The value of `responseHeaderTimeout` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  marathon:
    responseHeaderTimeout: "66s"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  responseHeaderTimeout = "66s"
  # ...
```

```bash tab="CLI"
--providers.marathon.responseHeaderTimeout=66s
# ...
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to Marathon.

#### `ca`

`ca` is the path to the certificate authority used for the secure connection to Marathon,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  marathon:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.marathon.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.marathon.tls.ca=path/to/ca.crt
```

#### `cert`

_Optional_

`cert` is the path to the public certificate used for the secure connection to Marathon.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  marathon:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.marathon.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.marathon.tls.cert=path/to/foo.cert
--providers.marathon.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection to Marathon.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  marathon:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.marathon.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.marathon.tls.cert=path/to/foo.cert
--providers.marathon.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Marathon accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  marathon:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.marathon.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.marathon.tls.insecureSkipVerify=true
```

### `tlsHandshakeTimeout`

_Optional, Default=5s_

Amount of time the Marathon provider should wait before timing out,
when waiting for the TLS handshake to complete.

The value of `tlsHandshakeTimeout` should be provided in seconds or as a valid duration format,
see [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).

```yaml tab="File (YAML)"
providers:
  marathon:
    tlsHandshakeTimeout: "10s"
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  tlsHandshakeTimeout = "10s"
  # ...
```

```bash tab="CLI"
--providers.marathon.tlsHandshakeTimeout=10s
# ...
```

### `trace`

_Optional, Default=false_

Displays additional provider logs when available.

```yaml tab="File (YAML)"
providers:
  marathon:
    trace: true
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  trace = true
  # ...
```

```bash tab="CLI"
--providers.marathon.trace=true
# ...
```

### `watch`

_Optional, Default=true_

When set to `true`, watches for Marathon changes.

```yaml tab="File (YAML)"
providers:
  marathon:
    watch: false
    # ...
```

```toml tab="File (TOML)"
[providers.marathon]
  watch = false
  # ...
```

```bash tab="CLI"
--providers.marathon.watch=false
# ...
```
