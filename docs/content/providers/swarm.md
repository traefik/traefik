---
title: "Traefik Docker Swarm Documentation"
description: "Learn how to achieve configuration discovery in Traefik through Docker Swarm. Read the technical documentation."
---

# Traefik & Docker Swarm

A Story of Labels & Containers
{: .subtitle }

![Docker](../assets/img/providers/docker.png)

Attach labels to your containers and let Traefik do the rest!

This provider works with [Docker Swarm Mode](https://docs.docker.com/engine/swarm/).

!!! tip "The Quick Start Uses Docker"

    If you have not already read it, maybe you would like to go through the [quick start guide](../getting-started/quick-start.md) that uses the Docker provider.

## Configuration Examples

??? example "Configuring Docker Swarm & Deploying / Exposing one Service"

    Enabling the Swarm provider

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        # swarm classic (1.12-)
        # endpoint: "tcp://127.0.0.1:2375"
        # docker swarm mode (1.12+)
        endpoint: "tcp://127.0.0.1:2377"
    ```

    ```toml tab="File (TOML)"
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

    Attach labels to a single service (not containers) while in Swarm mode (in your Docker compose file).
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

## Routing Configuration

When using Docker as a [provider](./overview.md),
Traefik uses [container labels](https://docs.docker.com/engine/reference/commandline/run/#label) to retrieve its routing configuration.

See the list of labels in the dedicated [routing](../routing/providers/docker.md) section.

### Routing Configuration with Labels

By default, Traefik watches for [container level labels](https://docs.docker.com/config/labels-custom-metadata/) on a standalone Docker Engine.

When using Docker Compose, labels are specified by the directive
[`labels`](https://docs.docker.com/compose/compose-file/compose-file-v3/#labels) from the
["services" objects](https://docs.docker.com/compose/compose-file/compose-file-v3/#service-configuration-reference).

!!! tip "Not Only Docker"

    Please note that any tool like Nomad, Terraform, Ansible, etc.
    that is able to define a Docker container with labels can work
    with Traefik and the  Swarm provider.

While in Swarm Mode, Traefik uses labels found on services, not on individual containers.

Therefore, if you use a compose file with Swarm Mode, labels should be defined in the
[`deploy`](https://docs.docker.com/compose/compose-file/compose-file-v3/#labels-1) part of your service.

This behavior is only enabled for docker-compose version 3+ ([Compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/)).

### Port Detection

Traefik retrieves the private IP and port of containers from the Docker API.

Docker Swarm does not provide any port detection information to Traefik.

Therefore, you **must** specify the port to use for communication by using the label `traefik.http.services.<service_name>.loadbalancer.server.port`
(Check the reference for this label in the [routing section for Swarm](../routing/providers/swarm.md#services)).

### Host networking

When exposing containers that are configured with [host networking](https://docs.docker.com/network/host/),
the IP address of the host is resolved as follows:

<!-- TODO: verify and document the swarm mode case with container.Node.IPAddress coming from the API -->
- try a lookup of `host.docker.internal`
- if the lookup was unsuccessful, try a lookup of `host.containers.internal`, ([Podman](https://docs.podman.io/en/latest/) equivalent of `host.docker.internal`)
- if that lookup was also unsuccessful, fall back to `127.0.0.1`

On Linux, for versions of Docker older than 20.10.0, for `host.docker.internal` to be defined, it should be provided
as an `extra_host` to the Traefik container, using the `--add-host` flag. For example, to set it to the IP address of
the bridge interface (`docker0` by default): `--add-host=host.docker.internal:172.17.0.1`

### IPv4 && IPv6

When using a docker stack that uses IPv6,
Traefik will use the IPv4 container IP before its IPv6 counterpart.
Therefore, on an IPv6 Docker stack,
Traefik will use the IPv6 container IP.

### Docker API Access

Traefik requires access to the docker socket to get its dynamic configuration.

You can specify which Docker API Endpoint to use with the directive [`endpoint`](#endpoint).

!!! warning "Security Note"

    Accessing the Docker API without any restriction is a security concern:
    If Traefik is attacked, then the attacker might get access to the underlying host.
    {: #security-note }

    As explained in the [Docker Daemon Attack Surface documentation](https://docs.docker.com/engine/security/#docker-daemon-attack-surface):

    !!! quote

        [...] only **trusted** users should be allowed to control your Docker daemon [...]

    ??? success "Solutions"

        Expose the Docker socket over TCP or SSH, instead of the default Unix socket file.
        It allows different implementation levels of the [AAA (Authentication, Authorization, Accounting) concepts](https://en.wikipedia.org/wiki/AAA_(computer_security)), depending on your security assessment:

        - Authentication with Client Certificates as described in ["Protect the Docker daemon socket."](https://docs.docker.com/engine/security/protect-access/)
        - Authorize and filter requests to restrict possible actions with [the TecnativaDocker Socket Proxy](https://github.com/Tecnativa/docker-socket-proxy).
        - Authorization with the [Docker Authorization Plugin Mechanism](https://web.archive.org/web/20190920092526/https://docs.docker.com/engine/extend/plugins_authorization/)
        - Accounting at networking level, by exposing the socket only inside a Docker private network, only available for Traefik.
        - Accounting at container level, by exposing the socket on a another container than Traefik's.
          It allows scheduling of Traefik on worker nodes, with only the "socket exposer" container on the manager nodes.
        - Accounting at kernel level, by enforcing kernel calls with mechanisms like [SELinux](https://en.wikipedia.org/wiki/Security-Enhanced_Linux), to only allows an identified set of actions for Traefik's process (or the "socket exposer" process).
        - SSH public key authentication (SSH is supported with Docker > 18.09)
        - Authentication using HTTP Basic authentication through an HTTP proxy that exposes the Docker daemon socket.

    ??? info "More Resources and Examples"

        - ["Paranoid about mounting /var/run/docker.sock?"](https://medium.com/@containeroo/traefik-2-0-paranoid-about-mounting-var-run-docker-sock-22da9cb3e78c)
        - [Traefik and Docker: A Discussion with Docker Captain, Bret Fisher](https://blog.traefik.io/traefik-and-docker-a-discussion-with-docker-captain-bret-fisher-7f0b9a54ff88)
        - [KubeCon EU 2018 Keynote, Running with Scissors, from Liz Rice](https://www.youtube.com/watch?v=ltrV-Qmh3oY)
        - [Don't expose the Docker socket (not even to a container)](https://www.lvh.io/posts/dont-expose-the-docker-socket-not-even-to-a-container/)
        - [A thread on Stack Overflow about sharing the `/var/run/docker.sock` file](https://news.ycombinator.com/item?id=17983623)
        - [To DinD or not to DinD](https://blog.loof.fr/2018/01/to-dind-or-not-do-dind.html)
        - [Traefik issue GH-4174 about security with Docker socket](https://github.com/traefik/traefik/issues/4174)
        - [Inspecting Docker Activity with Socat](https://developers.redhat.com/blog/2015/02/25/inspecting-docker-activity-with-socat/)
        - [Letting Traefik run on Worker Nodes](https://blog.mikesir87.io/2018/07/letting-traefik-run-on-worker-nodes/)
        - [Docker Socket Proxy from Tecnativa](https://github.com/Tecnativa/docker-socket-proxy)

Since the Swarm API is only exposed on the [manager nodes](https://docs.docker.com/engine/swarm/how-swarm-mode-works/nodes/#manager-nodes),
these are the nodes that Traefik should be scheduled on by deploying Traefik with a constraint on the node "role":

```shell tab="With Docker CLI"
docker service create \
  --constraint=node.role==manager \
  #... \
```

```yml tab="With Docker Compose"
services:
  traefik:
    # ...
    deploy:
      placement:
        constraints:
          - node.role == manager
```

!!! tip "Scheduling Traefik on Worker Nodes"

    Following the guidelines given in the previous section ["Docker API Access"](#docker-api-access),
    if you expose the Docker API through TCP, then Traefik can be scheduled on any node if the TCP
    socket is reachable.

    Please consider the security implications by reading the [Security Note](#security-note).

    A good example can be found on [Bret Fisher's repository](https://github.com/BretFisher/dogvscat/blob/master/stack-proxy-global.yml#L124).

### `endpoint`

_Required, Default="unix:///var/run/docker.sock"_

See the [Docker Swarm API Access](#docker-api-access) section for more information.

??? example "Using the docker.sock"

    The docker-compose file shares the docker sock with the Traefik container

    ```yaml
    services:
      traefik:
         image: traefik:v3.4 # The official v3 Traefik docker image
         ports:
           - "80:80"
         volumes:
           - /var/run/docker.sock:/var/run/docker.sock
    ```

    We specify the docker.sock in traefik's configuration file.

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        endpoint: "unix:///var/run/docker.sock"
         # ...
    ```

    ```toml tab="File (TOML)"
    [providers.swarm]
      endpoint = "unix:///var/run/docker.sock"
      # ...
    ```

    ```bash tab="CLI"
    --providers.swarm.endpoint=unix:///var/run/docker.sock
    # ...
    ```

??? example "Using SSH"

    Using Docker 18.09+ you can connect Traefik to daemon using SSH
    We specify the SSH host and user in Traefik's configuration file.
    Note that is server requires public keys for authentication you must have those accessible for user who runs Traefik.

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        endpoint: "ssh://traefik@192.168.2.5:2022"
         # ...
    ```

    ```toml tab="File (TOML)"
    [providers.swarm]
      endpoint = "ssh://traefik@192.168.2.5:2022"
      # ...
    ```

    ```bash tab="CLI"
    --providers.swarm.endpoint=ssh://traefik@192.168.2.5:2022
    # ...
    ```

??? example "Using HTTP"

    Using Docker Engine API you can connect Traefik to remote daemon using HTTP.

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        endpoint: "http://127.0.0.1:2375"
         # ...
    ```

    ```toml tab="File (TOML)"
    [providers.swarm]
      swarm = "http://127.0.0.1:2375"
      # ...
    ```

    ```bash tab="CLI"
    --providers.swarm.endpoint=http://127.0.0.1:2375
    # ...
    ```

??? example "Using TCP"

    Using Docker Engine API you can connect Traefik to remote daemon using TCP.

    ```yaml tab="File (YAML)"
    providers:
      swarm:
        endpoint: "tcp://127.0.0.1:2375"
         # ...
    ```

    ```toml tab="File (TOML)"
    [providers.swarm]
      swarm = "tcp://127.0.0.1:2375"
      # ...
    ```

    ```bash tab="CLI"
    --providers.swarm.endpoint=tcp://127.0.0.1:2375
    # ...
    ```

```yaml tab="File (YAML)"
providers:
  swarm:
    endpoint: "unix:///var/run/docker.sock"
```

```toml tab="File (TOML)"
[providers.swarm]
  endpoint = "unix:///var/run/docker.sock"
```

```bash tab="CLI"
--providers.swarm.endpoint=unix:///var/run/docker.sock
```

### `username`

_Optional, Default=""_

Defines the username for Basic HTTP authentication.
This should be used when the Docker daemon socket is exposed through an HTTP proxy that requires Basic HTTP authentication.

```yaml tab="File (YAML)"
providers:
  swarm:
    username: foo
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  username = "foo"
  # ...
```

```bash tab="CLI"
--providers.swarm.username="foo"
# ...
```

### `password`

_Optional, Default=""_

Defines the password for Basic HTTP authentication.
This should be used when the Docker daemon socket is exposed through an HTTP proxy that requires Basic HTTP authentication.

```yaml tab="File (YAML)"
providers:
  swarm:
    password: foo
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  password = "foo"
  # ...
```

```bash tab="CLI"
--providers.swarm.password="foo"
# ...
```

### `useBindPortIP`

_Optional, Default=false_

Traefik routes requests to the IP/port of the matching container.
When setting `useBindPortIP=true`, you tell Traefik to use the IP/Port attached to the container's _binding_ instead of its inner network IP/Port.

When used in conjunction with the `traefik.http.services.<name>.loadbalancer.server.port` label (that tells Traefik to route requests to a specific port),
Traefik tries to find a binding on port `traefik.http.services.<name>.loadbalancer.server.port`.
If it cannot find such a binding, Traefik falls back on the internal network IP of the container,
but still uses the `traefik.http.services.<name>.loadbalancer.server.port` that is set in the label.

??? example "Examples of `usebindportip` in different situations."

    | port label         | Container's binding                                | Routes to      |
    |--------------------|----------------------------------------------------|----------------|
    |          -         |           -                                        | IntIP:IntPort  |
    |          -         | ExtPort:IntPort                                    | IntIP:IntPort  |
    |          -         | ExtIp:ExtPort:IntPort                              | ExtIp:ExtPort  |
    | LblPort            |           -                                        | IntIp:LblPort  |
    | LblPort            | ExtIp:ExtPort:LblPort                              | ExtIp:ExtPort  |
    | LblPort            | ExtIp:ExtPort:OtherPort                            | IntIp:LblPort  |
    | LblPort            | ExtIp1:ExtPort1:IntPort1 & ExtIp2:LblPort:IntPort2 | ExtIp2:LblPort |

    !!! info ""
        In the above table:

        - `ExtIp` stands for "external IP found in the binding"
        - `IntIp` stands for "internal network container's IP",
        - `ExtPort` stands for "external Port found in the binding"
        - `IntPort` stands for "internal network container's port."

```yaml tab="File (YAML)"
providers:
  swarm:
    useBindPortIP: true
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  useBindPortIP = true
  # ...
```

```bash tab="CLI"
--providers.swarm.useBindPortIP=true
# ...
```

### `exposedByDefault`

_Optional, Default=true_

Expose containers by default through Traefik.
If set to `false`, containers that do not have a `traefik.enable=true` label are ignored from the resulting routing configuration.

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  swarm:
    exposedByDefault: false
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  exposedByDefault = false
  # ...
```

```bash tab="CLI"
--providers.swarm.exposedByDefault=false
# ...
```

### `network`

_Optional, Default=""_

Defines a default docker network to use for connections to all containers.

This option can be overridden on a per-container basis with the `traefik.swarm.network` [routing label](../routing/providers/swarm.md#traefikswarmnetwork).

```yaml tab="File (YAML)"
providers:
  swarm:
    network: test
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  network = "test"
  # ...
```

```bash tab="CLI"
--providers.swarm.network=test
# ...
```

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

The `defaultRule` option defines what routing rule to apply to a container if no rule is defined by a label.

It must be a valid [Go template](https://pkg.go.dev/text/template/), and can use
[sprig template functions](https://masterminds.github.io/sprig/).
The container service name can be accessed with the `Name` identifier,
and the template has access to all the labels defined on this container.

```yaml tab="File (YAML)"
providers:
  swarm:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```bash tab="CLI"
--providers.swarm.defaultRule='Host(`{{ .Name }}.{{ index .Labels "customLabel"}}`)'
# ...
```

??? info "Default rule and Traefik service"

    The exposure of the Traefik container, combined with the default rule mechanism,
    can lead to create a router targeting itself in a loop.
    In this case, to prevent an infinite loop,
    Traefik adds an internal middleware to refuse the request if it comes from the same router.

### `refreshSeconds`

_Optional, Default=15_

Defines the polling interval (in seconds) for Swarm Mode.

```yaml tab="File (YAML)"
providers:
  swarm:
    refreshSeconds: 30
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  refreshSeconds = 30
  # ...
```

```bash tab="CLI"
--providers.swarm.refreshSeconds=30
# ...
```

### `httpClientTimeout`

_Optional, Default=0_

Defines the client timeout (in seconds) for HTTP connections. If its value is `0`, no timeout is set.

```yaml tab="File (YAML)"
providers:
  swarm:
    httpClientTimeout: 300
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  httpClientTimeout = 300
  # ...
```

```bash tab="CLI"
--providers.swarm.httpClientTimeout=300
# ...
```

### `watch`

_Optional, Default=true_

Watch Docker events.

```yaml tab="File (YAML)"
providers:
  swarm:
    watch: false
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  watch = false
  # ...
```

```bash tab="CLI"
--providers.swarm.watch=false
# ...
```

### `constraints`

_Optional, Default=""_

The `constraints` option can be set to an expression that Traefik matches against the container labels to determine whether
to create any route for that container. If none of the container labels match the expression, no route for that container is
created. If the expression is empty, all detected containers are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegex("key", "value")` functions,
as well as the usual boolean logic, as shown in examples below.

??? example "Constraints Expression Examples"

    ```toml
    # Includes only containers having a label with key `a.label.name` and value `foo`
    constraints = "Label(`a.label.name`, `foo`)"
    ```

    ```toml
    # Excludes containers having any label with key `a.label.name` and value `foo`
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
    # Includes only containers having a label with key `a.label.name` and a value matching the `a.+` regular expression.
    constraints = "LabelRegex(`a.label.name`, `a.+`)"
    ```

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  swarm:
    constraints: "Label(`a.label.name`,`foo`)"
    # ...
```

```toml tab="File (TOML)"
[providers.swarm]
  constraints = "Label(`a.label.name`,`foo`)"
  # ...
```

```bash tab="CLI"
--providers.swarm.constraints=Label(`a.label.name`,`foo`)
# ...
```

### `tls`

_Optional_

Defines the TLS configuration used for the secure connection to Docker.

#### `ca`

_Optional_

`ca` is the path to the certificate authority used for the secure connection to Docker,
it defaults to the system bundle.

```yaml tab="File (YAML)"
providers:
  swarm:
    tls:
      ca: path/to/ca.crt
```

```toml tab="File (TOML)"
[providers.swarm.tls]
  ca = "path/to/ca.crt"
```

```bash tab="CLI"
--providers.swarm.tls.ca=path/to/ca.crt
```

#### `cert`

`cert` is the path to the public certificate used for the secure connection to Docker.
When using this option, setting the `key` option is required.

```yaml tab="File (YAML)"
providers:
  swarm:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.swarm.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.swarm.tls.cert=path/to/foo.cert
--providers.swarm.tls.key=path/to/foo.key
```

#### `key`

_Optional_

`key` is the path to the private key used for the secure connection Docker.
When using this option, setting the `cert` option is required.

```yaml tab="File (YAML)"
providers:
  swarm:
    tls:
      cert: path/to/foo.cert
      key: path/to/foo.key
```

```toml tab="File (TOML)"
[providers.swarm.tls]
  cert = "path/to/foo.cert"
  key = "path/to/foo.key"
```

```bash tab="CLI"
--providers.swarm.tls.cert=path/to/foo.cert
--providers.swarm.tls.key=path/to/foo.key
```

#### `insecureSkipVerify`

_Optional, Default=false_

If `insecureSkipVerify` is `true`, the TLS connection to Docker accepts any certificate presented by the server regardless of the hostnames it covers.

```yaml tab="File (YAML)"
providers:
  swarm:
    tls:
      insecureSkipVerify: true
```

```toml tab="File (TOML)"
[providers.swarm.tls]
  insecureSkipVerify = true
```

```bash tab="CLI"
--providers.swarm.tls.insecureSkipVerify=true
```

### `allowEmptyServices`

_Optional, Default=false_

If the parameter is set to `true`,
any [servers load balancer](../routing/services/index.md#servers-load-balancer) defined for Docker containers is created 
regardless of the [healthiness](https://docs.docker.com/engine/reference/builder/#healthcheck) of the corresponding containers.
It also then stays alive and responsive even at times when it becomes empty,
i.e. when all its children containers become unhealthy.
This results in `503` HTTP responses instead of `404` ones,
in the above cases.

```yaml tab="File (YAML)"
providers:
  swarm:
    allowEmptyServices: true
```

```toml tab="File (TOML)"
[providers.swarm]
  allowEmptyServices = true
```

```bash tab="CLI"
--providers.swarm.allowEmptyServices=true
```

{!traefik-for-business-applications.md!}
