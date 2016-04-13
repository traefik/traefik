
# Concepts

Let's take our example from the [overview](https://docs.traefik.io/#overview) again:


> Imagine that you have deployed a bunch of microservices on your infrastructure. You probably used a service registry (like etcd or consul) and/or an orchestrator (swarm, Mesos/Marathon) to manage all these services.
> If you want your users to access some of your microservices from the Internet, you will have to use a reverse proxy and configure it using virtual hosts or prefix paths:

> - domain `api.domain.com` will point the microservice `api` in your private network
> - path `domain.com/web` will point the microservice `web` in your private network
> - domain `backoffice.domain.com` will point the microservices `backoffice` in your private network, load-balancing between your multiple instances

> ![Architecture](img/architecture.png)

Let's zoom on Træfɪk and have an overview of its internal architecture:


![Architecture](img/internal.png)

- Incoming requests end on [entrypoints](#entrypoints), as the name suggests, they are the network entry points into Træfɪk (listening port, SSL, traffic redirection...).
- Traffic is then forwared to a matching [frontend](#frontends). A frontend defines routes from [entrypoints](#entrypoints) to [backends](#backends).
Routes are created using requests fields (`Host`, `Path`, `Headers`...) and can match or not a request.
- The [frontend](#frontends) will then send the request to a [backend](#backends). A backend can be composed by one or more [servers](#servers), and by a load-balancing strategy.
- Finally, the [server](#servers) will forward the request to the corresponding microservice in the private network.

## Entrypoints

Entrypoints are the network entry points into Træfɪk.
They can be defined using:

- a port (80, 443...)
- SSL (Certificates. Keys...)
- redirection to another entrypoint (redirect `HTTP` to `HTTPS`)

Here is an example of entrypoints definition:

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "tests/traefik.crt"
      keyFile = "tests/traefik.key"
```

- Two entrypoints are defined `http` and `https`.
- `http` listens on port `80` et `https` on port `443`.
- We enable SSL en `https` by giving a certificate and a key.
- We also redirect all the traffic from entrypoint `http` to `https`.

## Frontends

A frontend is a set of rules that forwards the incoming traffic from an entrypoint to a backend.
Frontends can be defined using the following rules:

- `Headers: Content-Type, application/json`: Headers adds a matcher for request header values. It accepts a sequence of key/value pairs to be matched.
- `HeadersRegexp: Content-Type, application/(text|json)`: Regular expressions can be used with headers as well. It accepts a sequence of key/value pairs, where the value has regex support.
- `Host: traefik.io, www.traefik.io`: Match request host with given host list.
- `HostRegexp: traefik.io, {subdomain:[a-z]+}.traefik.io`: Adds a matcher for the URL hosts. It accepts templates with zero or more URL variables enclosed by `{}`. Variables can define an optional regexp pattern to be matched.
- `Method: GET, POST, PUT`: Method adds a matcher for HTTP methods. It accepts a sequence of one or more methods to be matched.
- `Path: /products/, /articles/{category}/{id:[0-9]+}`: Path adds a matcher for the URL paths. It accepts templates with zero or more URL variables enclosed by `{}`.
- `PathStrip`: Same as `Path` but strip the given prefix from the request URL's Path.
- `PathPrefix`: PathPrefix adds a matcher for the URL path prefixes. This matches if the given template is a prefix of the full URL path.
- `PathPrefixStrip`: Same as `PathPrefix` but strip the given prefix from the request URL's Path.

You can optionally enable `passHostHeader` to forward client `Host` header to the backend.

Here is an example of frontends definition:

```toml
[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host: test.localhost, test2.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host: localhost, {subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  backend = "backend2"
    rule = "Path:/test"
```

- Three frontends are defined: `frontend1`, `frontend2` and `frontend3`
- `frontend1` will forward the traffic to the `backend2` if the rule `Host: test.localhost, test2.localhost` is matched
- `frontend2` will forward the traffic to the `backend1` if the rule `Host: localhost, {subdomain:[a-z]+}.localhost` is matched (forwarding client `Host` header to the backend)
- `frontend3` will forward the traffic to the `backend2` if the rule `Path:/test` is matched

## Backends

A backend is responsible to load-balance the traffic coming from one or more frontends to a set of http servers.
Various methods of load-balancing is supported:

- `wrr`: Weighted Round Robin
- `drr`: Dynamic Round Robin: increases weights on servers that perform better than others. It also rolls back to original weights if the servers have changed.

A circuit breaker can also be applied to a backend, preventing high loads on failing servers.
Initial state is Standby. CB observes the statistics and does not modify the request.
In case if condition matches, CB enters Tripped state, where it responds with predefines code or redirects to another frontend.
Once Tripped timer expires, CB enters Recovering state and resets all stats.
In case if the condition does not match and recovery timer expires, CB enters Standby state.

It can be configured using:

- Methods: `LatencyAtQuantileMS`, `NetworkErrorRatio`, `ResponseCodeRatio`
- Operators:  `AND`, `OR`, `EQ`, `NEQ`, `LT`, `LE`, `GT`, `GE`

For example:

- `NetworkErrorRatio() > 0.5`: watch error ratio over 10 second sliding window for a frontend
- `LatencyAtQuantileMS(50.0) > 50`:  watch latency at quantile in milliseconds.
- `ResponseCodeRatio(500, 600, 0, 600) > 0.5`: ratio of response codes in range [500-600) to  [0-600)

To proactively prevent backends from being overwhelmed with high load, a maximum connection limit can
also be applied to each backend.

Maximum connections can be configured by specifying an integer value for `maxconn.amount` and
`maxconn.extractorfunc` which is a strategy used to determine how to categorize requests in order to
evaluate the maximum connections.

For example:
```toml
[backends]
  [backends.backend1]
    [backends.backend1.maxconn]
       amount = 10
       extractorfunc = "request.host"
```

- `backend1` will return `HTTP code 429 Too Many Requests` if there are already 10 requests in progress for the same Host header.
- Another possible value for `extractorfunc` is `client.ip` which will categorize requests based on client source ip.
- Lastly `extractorfunc` can take the value of `request.header.ANY_HEADER` which will categorize requests based on `ANY_HEADER` that you provide.

## Servers

Servers are simply defined using a `URL`. You can also apply a custom `weight` to each server (this will be used by load-balacning).

Here is an example of backends and servers definition:

```toml
[backends]
  [backends.backend1]
    [backends.backend1.circuitbreaker]
      expression = "NetworkErrorRatio() > 0.5"
    [backends.backend1.servers.server1]
    url = "http://172.17.0.2:80"
    weight = 10
    [backends.backend1.servers.server2]
    url = "http://172.17.0.3:80"
    weight = 1
  [backends.backend2]
    [backends.backend2.LoadBalancer]
      method = "drr"
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1
    [backends.backend2.servers.server2]
    url = "http://172.17.0.5:80"
    weight = 2
```

- Two backends are defined: `backend1` and `backend2`
- `backend1` will forward the traffic to two servers: `http://172.17.0.2:80"` with weight `10` and `http://172.17.0.3:80` with weight `1` using default `wrr` load-balancing strategy.
- `backend2` will forward the traffic to two servers: `http://172.17.0.4:80"` with weight `1` and `http://172.17.0.5:80` with weight `2` using `drr` load-balancing strategy.
- a circuit breaker is added on `backend1` using the expression `NetworkErrorRatio() > 0.5`: watch error ratio over 10 second sliding window

# Launch

Træfɪk can be configured using a TOML file configuration, arguments, or both.
By default, Træfɪk will try to find a `traefik.toml` in the following places:

- `/etc/traefik/`
- `$HOME/.traefik/`
- `.` *the working directory*

You can override this by setting a `configFile` argument:

```bash
$ traefik --configFile=foo/bar/myconfigfile.toml
```

Træfɪk uses the following precedence order. Each item takes precedence over the item below it:

- arguments
- configuration file
- default

It means that arguments overrides configuration file.
Each argument is described in the help section:

```bash
$ traefik --help
```
