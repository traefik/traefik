---
title: "Traefik Routers Documentation"
description: "In Traefik Proxy, a router is in charge of connecting incoming requests to the Services that can handle them. Read the technical documentation."
---

# Routers

Connecting Requests to Services
{: .subtitle }

![routers](../../assets/img/routers.png)

A router is in charge of connecting incoming requests to the services that can handle them.
In the process, routers may use pieces of [middleware](../../middlewares/overview.md) to update the request,
or act before forwarding the request to the service.

## Configuration Example

??? example "Requests /foo are Handled by service-foo -- Using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      routers:
        my-router:
          rule: "Path(`/foo`)"
          service: service-foo
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.routers]
      [http.routers.my-router]
        rule = "Path(`/foo`)"
        service = "service-foo"
    ```

??? example "Forwarding all (non-tls) requests on port 3306 to a database service"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        to-database:
          entryPoints:
            - "mysql"
          # Catch every request (only available rule for non-tls routers. See below.)
          rule: "HostSNI(`*`)"
          service: database
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp]
      [tcp.routers]
        [tcp.routers.to-database]
          entryPoints = ["mysql"]
          # Catch every request (only available rule for non-tls routers. See below.)
          rule = "HostSNI(`*`)"
          service = "database"
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
      mysql:
        address: ":3306"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.mysql]
        address = ":3306"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entryPoints.web.address=:80
    --entryPoints.mysql.address=:3306
    ```

## Configuring HTTP Routers

!!! warning "The character `@` is not authorized in the router name"

### EntryPoints

If not specified, HTTP routers will accept requests from all defined entry points.
If you want to limit the router scope to a set of entry points, set the `entryPoints` option.

??? example "Listens to Every EntryPoint"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          # By default, routers listen to every entry points
          rule: "Host(`example.com`)"
          service: "service-1"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        # By default, routers listen to every entry points
        rule = "Host(`example.com`)"
        service = "service-1"
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
      websecure:
        address: ":443"
      other:
        address: ":9090"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.websecure]
        address = ":443"
      [entryPoints.other]
        address = ":9090"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entrypoints.web.address=:80
    --entrypoints.websecure.address=:443
    --entrypoints.other.address=:9090
    ```

??? example "Listens to Specific EntryPoints"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          # won't listen to entry point web
          entryPoints:
            - "websecure"
            - "other"
          rule: "Host(`example.com`)"
          service: "service-1"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        # won't listen to entry point web
        entryPoints = ["websecure", "other"]
        rule = "Host(`example.com`)"
        service = "service-1"
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration
    entryPoints:
      web:
        address: ":80"
      websecure:
        address: ":443"
      other:
        address: ":9090"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration
    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.websecure]
        address = ":443"
      [entryPoints.other]
        address = ":9090"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entrypoints.web.address=:80
    --entrypoints.websecure.address=:443
    --entrypoints.other.address=:9090
    ```

### Rule

Rules are a set of matchers configured with values, that determine if a particular request matches specific criteria.
If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

??? tip "Backticks or Quotes?"
    To set the value of a rule, use [backticks](https://en.wiktionary.org/wiki/backtick) ``` ` ``` or escaped double-quotes `\"`.

    Single quotes `'` are not accepted since the values are [Golang's String Literals](https://golang.org/ref/spec#String_literals).

!!! example "Host is example.com"

    ```toml
    rule = "Host(`example.com`)"
    ```

!!! example "Host is example.com OR Host is example.org AND path is /traefik"

    ```toml
    rule = "Host(`example.com`) || (Host(`example.org`) && Path(`/traefik`))"
    ```

The table below lists all the available matchers:

| Rule                                                                                       | Description                                                                                                    |
|--------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| ```Headers(`key`, `value`)```                                                              | Check if there is a key `key`defined in the headers, with the value `value`                                    |
| ```HeadersRegexp(`key`, `regexp`)```                                                       | Check if there is a key `key`defined in the headers, with a value that matches the regular expression `regexp` |
| ```Host(`example.com`, ...)```                                                             | Check if the request domain (host header value) targets one of the given `domains`.                            |
| ```HostHeader(`example.com`, ...)```                                                       | Same as `Host`, only exists for historical reasons.                                                            |
| ```HostRegexp(`example.com`, `{subdomain:[a-z]+}.example.com`, ...)```                     | Match the request domain. See "Regexp Syntax" below.                                                           |
| ```Method(`GET`, ...)```                                                                   | Check if the request method is one of the given `methods` (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`)    |
| ```Path(`/path`, `/articles/{cat:[a-z]+}/{id:[0-9]+}`, ...)```                             | Match exact request path. See "Regexp Syntax" below.                                                           |
| ```PathPrefix(`/products/`, `/articles/{cat:[a-z]+}/{id:[0-9]+}`)```                       | Match request prefix path. See "Regexp Syntax" below.                                                          |
| ```Query(`foo=bar`, `bar=baz`)```                                                          | Match Query String parameters. It accepts a sequence of key=value pairs.                                       |
| ```ClientIP(`10.0.0.0/16`, `::1`)```                                                       | Match if the request client IP is one of the given IP/CIDR. It accepts IPv4, IPv6 and CIDR formats.            |

!!! important "Non-ASCII Domain Names"

    Non-ASCII characters are not supported in `Host` and `HostRegexp` expressions, and by doing so the associated router will be invalid.
    For the `Host` expression, domain names containing non-ASCII characters must be provided as punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)).
    As well, when using the `HostRegexp` expressions, in order to match domain names containing non-ASCII characters, the regular expression should match a punycode encoded domain name.

!!! important "Regexp Syntax"

    `HostRegexp`, `PathPrefix`, and `Path` accept an expression with zero or more groups enclosed by curly braces, which are called named regexps.
    Named regexps, of the form `{name:regexp}`, are the only expressions considered for regexp matching.
    The regexp name (`name` in the above example) is an arbitrary value, that exists only for historical reasons.

    Any `regexp` supported by [Go's regexp package](https://golang.org/pkg/regexp/) may be used.
    For example, here is a case insensitive path matcher syntax: ```Path(`/{path:(?i:Products)}`)```.

!!! info "Combining Matchers Using Operators and Parenthesis"

    The usual AND (`&&`) and OR (`||`) logical operators can be used, with the expected precedence rules,
    as well as parentheses.

!!! info "Inverting a matcher"

    One can invert a matcher by using the `!` operator.

!!! important "Rule, Middleware, and Services"

    The rule is evaluated "before" any middleware has the opportunity to work, and "before" the request is forwarded to the service.

!!! info "Path Vs PathPrefix"

    Use `Path` if your service listens on the exact path only. For instance, ```Path(`/products`)``` would match `/products` but not `/products/shoes`.

    Use a `*Prefix*` matcher if your service listens on a particular base path but also serves requests on sub-paths.
    For instance, ```PathPrefix(`/products`)``` would match `/products` and `/products/shoes`,
    as well as `/productsforsale`, and `/productsforsale/shoes`.
    Since the path is forwarded as-is, your service is expected to listen on `/products`.

!!! info "ClientIP matcher"

    The `ClientIP` matcher will only match the request client IP and does not use the `X-Forwarded-For` header for matching.

### Priority

To avoid path overlap, routes are sorted, by default, in descending order using rules length. The priority is directly equal to the length of the rule, and so the longest length has the highest priority.

A value of `0` for the priority is ignored: `priority = 0` means that the default rules length sorting is used.

??? info "How default priorities are computed"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          rule: "HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
          # ...
        Router-2:
          rule: "Host(`foobar.traefik.com`)"
          # ...
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        rule = "HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
        # ...
      [http.routers.Router-2]
        rule = "Host(`foobar.traefik.com`)"
        # ...
    ```

    In this case, all requests with host `foobar.traefik.com` will be routed through `Router-1` instead of `Router-2`.

    | Name     | Rule                                               | Priority |
    |----------|----------------------------------------------------|----------|
    | Router-1 | ```HostRegexp(`{subdomain:[a-z]+}.traefik.com`)``` | 44       |
    | Router-2 | ```Host(`foobar.traefik.com`)```                   | 26       |

    The previous table shows that `Router-1` has a higher priority than `Router-2`.

    To solve this issue, the priority must be set.

??? example "Set priorities -- using the [File Provider](../../providers/file.md)"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          rule: "HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
          entryPoints:
          - "web"
          service: service-1
          priority: 1
        Router-2:
          rule: "Host(`foobar.traefik.com`)"
          entryPoints:
          - "web"
          priority: 2
          service: service-2
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        rule = "HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
        entryPoints = ["web"]
        service = "service-1"
        priority = 1
      [http.routers.Router-2]
        rule = "Host(`foobar.traefik.com`)"
        entryPoints = ["web"]
        priority = 2
        service = "service-2"
    ```

    In this configuration, the priority is configured to allow `Router-2` to handle requests with the `foobar.traefik.com` host.

### Middlewares

You can attach a list of [middlewares](../../middlewares/overview.md) to each HTTP router.
The middlewares will take effect only if the rule matches, and before forwarding the request to the service.

!!! warning "The character `@` is not authorized in the middleware name."

!!! tip "Middlewares order"

    Middlewares are applied in the same order as their declaration in **router**.

??? example "With a [middleware](../../middlewares/overview.md) -- using the [File Provider](../../providers/file.md)"

    ```yaml tab="YAML"
    ## Dynamic configuration
    http:
      routers:
        my-router:
          rule: "Path(`/foo`)"
          # declared elsewhere
          middlewares:
          - authentication
          service: service-foo
    ```

    ```toml tab="TOML"
    ## Dynamic configuration
    [http.routers]
      [http.routers.my-router]
        rule = "Path(`/foo`)"
        # declared elsewhere
        middlewares = ["authentication"]
        service = "service-foo"
    ```

### Service

Each request must eventually be handled by a [service](../services/index.md),
which is why each router definition should include a service target,
which is basically where the request will be passed along to.

In general, a service assigned to a router should have been defined,
but there are exceptions for label-based providers.
See the specific [docker](../providers/docker.md#service-definition), [rancher](../providers/rancher.md#service-definition),
or [marathon](../providers/marathon.md#service-definition) documentation.

!!! warning "The character `@` is not authorized in the service name."

!!! important "HTTP routers can only target HTTP services (not TCP services)."

### TLS

#### General

 When a TLS section is specified, it instructs Traefik that the current router is dedicated to HTTPS requests only (and that the router should ignore HTTP (non TLS) requests).
Traefik will terminate the SSL connections (meaning that it will send decrypted data to the services).

??? example "Configuring the router to accept HTTPS requests only"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          rule: "Host(`foo-domain`) && Path(`/foo-path/`)"
          service: service-id
          # will terminate the TLS request
          tls: {}
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
        service = "service-id"
        # will terminate the TLS request
        [http.routers.Router-1.tls]
    ```

!!! important "Routers for HTTP & HTTPS"

    If you need to define the same route for both HTTP and HTTPS requests, you will need to define two different routers:
    one with the tls section, one without.

    ??? example "HTTP & HTTPS routes"

        ```yaml tab="File (YAML)"
        ## Dynamic configuration
        http:
          routers:
            my-https-router:
              rule: "Host(`foo-domain`) && Path(`/foo-path/`)"
              service: service-id
              # will terminate the TLS request
              tls: {}

            my-http-router:
              rule: "Host(`foo-domain`) && Path(`/foo-path/`)"
              service: service-id
        ```

        ```toml tab="File (TOML)"
        ## Dynamic configuration
        [http.routers]
          [http.routers.my-https-router]
            rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
            service = "service-id"
            # will terminate the TLS request
            [http.routers.my-https-router.tls]

          [http.routers.my-http-router]
            rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
            service = "service-id"
        ```

#### `options`

The `options` field enables fine-grained control of the TLS parameters.
It refers to a [TLS Options](../../https/tls.md#tls-options) and will be applied only if a `Host` rule is defined.

!!! info "Server Name Association"

    Even though one might get the impression that a TLS options reference is mapped to a router, or a router rule,
    one should realize that it is actually mapped only to the host name found in the `Host` part of the rule.
    Of course, there could also be several `Host` parts in a rule, in which case the TLS options reference would be mapped to as many host names.

    Another thing to keep in mind is:
    the TLS option is picked from the mapping mentioned above and based on the server name provided during the TLS handshake,
    and it all happens before routing actually occurs.

!!! info "Domain Fronting"

    In the case of domain fronting,
    if the TLS options associated with the Host Header and the SNI are different then Traefik will respond with a status code `421`.

??? example "Configuring the TLS options"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          rule: "Host(`foo-domain`) && Path(`/foo-path/`)"
          service: service-id
          # will terminate the TLS request
          tls:
            options: foo

    tls:
      options:
        foo:
          minVersion: VersionTLS12
          cipherSuites:
            - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
            - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
            - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
            - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
            - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
        service = "service-id"
        # will terminate the TLS request
        [http.routers.Router-1.tls]
          options = "foo"

    [tls.options]
      [tls.options.foo]
        minVersion = "VersionTLS12"
        cipherSuites = [
          "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
          "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
          "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
          "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
          "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
        ]
    ```

!!! important "Conflicting TLS Options"

    Since a TLS options reference is mapped to a host name,
    if a configuration introduces a situation where the same host name (from a `Host` rule) gets matched with two TLS options references,
    a conflict occurs, such as in the example below:

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        routerfoo:
          rule: "Host(`snitest.com`) && Path(`/foo`)"
          tls:
            options: foo

        routerbar:
          rule: "Host(`snitest.com`) && Path(`/bar`)"
          tls:
            options: bar
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.routerfoo]
        rule = "Host(`snitest.com`) && Path(`/foo`)"
        [http.routers.routerfoo.tls]
          options = "foo"

    [http.routers]
      [http.routers.routerbar]
        rule = "Host(`snitest.com`) && Path(`/bar`)"
        [http.routers.routerbar.tls]
          options = "bar"
    ```

    If that happens, both mappings are discarded, and the host name (`snitest.com` in this case) for these routers gets associated with the default TLS options instead.

#### `certResolver`

If `certResolver` is defined, Traefik will try to generate certificates based on routers `Host` & `HostSNI` rules.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  routers:
    routerfoo:
      rule: "Host(`snitest.com`) && Path(`/foo`)"
      tls:
        certResolver: foo
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.routerfoo]
    rule = "Host(`snitest.com`) && Path(`/foo`)"
    [http.routers.routerfoo.tls]
      certResolver = "foo"
```

!!! info "Multiple Hosts in a Rule"
    The rule ```Host(`test1.example.com`,`test2.example.com`)``` will request a certificate with the main domain `test1.example.com` and SAN `test2.example.com`.

#### `domains`

You can set SANs (alternative domains) for each main domain.
Every domain must have A/AAAA records pointing to Traefik.
Each domain & SAN will lead to a certificate request.

```yaml tab="File (YAML)"
## Dynamic configuration
http:
  routers:
    routerbar:
      rule: "Host(`snitest.com`) && Path(`/bar`)"
      tls:
        certResolver: "bar"
        domains:
          - main: "snitest.com"
            sans:
              - "*.snitest.com"
```

```toml tab="File (TOML)"
## Dynamic configuration
[http.routers]
  [http.routers.routerbar]
    rule = "Host(`snitest.com`) && Path(`/bar`)"
    [http.routers.routerbar.tls]
      certResolver = "bar"
      [[http.routers.routerbar.tls.domains]]
        main = "snitest.com"
        sans = ["*.snitest.com"]
```

[ACME v2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) supports wildcard certificates.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a [`DNS-01` challenge](../../https/acme.md#dnschallenge).

Most likely the root domain should receive a certificate too, so it needs to be specified as SAN and 2 `DNS-01` challenges are executed.
In this case the generated DNS TXT record for both domains is the same.
Even though this behavior is [DNS RFC](https://community.letsencrypt.org/t/wildcard-issuance-two-txt-records-for-the-same-name/54528/2) compliant,
it can lead to problems as all DNS providers keep DNS records cached for a given time (TTL) and this TTL can be greater than the challenge timeout making the `DNS-01` challenge fail.

The Traefik ACME client library [lego](https://github.com/go-acme/lego) supports some but not all DNS providers to work around this issue.
The [supported `provider` table](../../https/acme.md#providers) indicates if they allow generating certificates for a wildcard domain and its root domain.

!!! important "Wildcard certificates can only be verified through a [`DNS-01` challenge](../../https/acme.md#dnschallenge)."

!!! warning "Double Wildcard Certificates"
    It is not possible to request a double wildcard certificate for a domain (for example `*.*.local.com`).

## Configuring TCP Routers

!!! warning "The character `@` is not authorized in the router name"

### General

If both HTTP routers and TCP routers listen to the same entry points, the TCP routers will apply *before* the HTTP routers.
If no matching route is found for the TCP routers, then the HTTP routers will take over.

### EntryPoints

If not specified, TCP routers will accept requests from all defined entry points.
If you want to limit the router scope to a set of entry points, set the entry points option.

??? info "How to handle Server First protocols?"

    To correctly handle a request, Traefik needs to wait for the first
    few bytes to arrive before it can decide what to do with it.

    For protocols where the server is expected to send first, such
    as SMTP, if no specific setup is in place, we could end up in
    a situation where both sides are waiting for data and the
    connection appears to have hanged.

    The only way that Traefik can deal with such a case, is to make 
    sure that on the concerned entry point, there is no TLS router 
    whatsoever (neither TCP nor HTTP), and there is at least one 
    non-TLS TCP router that leads to the server in question.

??? example "Listens to Every Entry Point"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration

    tcp:
      routers:
        Router-1:
          # By default, routers listen to every entrypoints
          rule: "HostSNI(`example.com`)"
          service: "service-1"
          # will route TLS requests (and ignore non tls requests)
          tls: {}
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration

    [tcp.routers]
      [tcp.routers.Router-1]
        # By default, routers listen to every entrypoints
        rule = "HostSNI(`example.com`)"
        service = "service-1"
        # will route TLS requests (and ignore non tls requests)
        [tcp.routers.Router-1.tls]
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration

    entryPoints:
      web:
        address: ":80"
      websecure:
        address: ":443"
      other:
        address: ":9090"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration

    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.websecure]
        address = ":443"
      [entryPoints.other]
        address = ":9090"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entrypoints.web.address=:80
    --entrypoints.websecure.address=:443
    --entrypoints.other.address=:9090
    ```

??? example "Listens to Specific Entry Points"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          # won't listen to entry point web
          entryPoints:
            - "websecure"
            - "other"
          rule: "HostSNI(`example.com`)"
          service: "service-1"
          # will route TLS requests (and ignore non tls requests)
          tls: {}
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        # won't listen to entry point web
        entryPoints = ["websecure", "other"]
        rule = "HostSNI(`example.com`)"
        service = "service-1"
        # will route TLS requests (and ignore non tls requests)
        [tcp.routers.Router-1.tls]
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration

    entryPoints:
      web:
        address: ":80"
      websecure:
        address: ":443"
      other:
        address: ":9090"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration

    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.websecure]
        address = ":443"
      [entryPoints.other]
        address = ":9090"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entrypoints.web.address=:80
    --entrypoints.websecure.address=:443
    --entrypoints.other.address=:9090
    ```

### Rule

Rules are a set of matchers configured with values, that determine if a particular request matches specific criteria.
If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

??? tip "Backticks or Quotes?"

     To set the value of a rule, use [backticks](https://en.wiktionary.org/wiki/backtick) ``` ` ``` or escaped double-quotes `\"`.

    Single quotes `'` are not accepted since the values are [Golang's String Literals](https://golang.org/ref/spec#String_literals).

!!! example "HostSNI is example.com"

    ```toml
    rule = "HostSNI(`example.com`)"
    ```

!!! example "HostSNI is example.com OR HostSNI is example.org AND ClientIP is 0.0.0.0"

    ```toml
    rule = "HostSNI(`example.com`) || (HostSNI(`example.org`) && ClientIP(`0.0.0.0`))"
    ```

The table below lists all the available matchers:

| Rule                                                                      | Description                                                                                             |
|---------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
| ```HostSNI(`domain-1`, ...)```                                            | Checks if the Server Name Indication corresponds to the given `domains`.                                |
| ```HostSNIRegexp(`example.com`, `{subdomain:[a-z]+}.example.com`, ...)``` | Checks if the Server Name Indication matches the given regular expressions. See "Regexp Syntax" below.  |
| ```ClientIP(`10.0.0.0/16`, `::1`)```                                      | Checks if the connection client IP is one of the given IP/CIDR. It accepts IPv4, IPv6 and CIDR formats. |
| ```ALPN(`mqtt`, `h2c`)```                                                 | Checks if any of the connection ALPN protocols is one of the given protocols.                           |

!!! important "Non-ASCII Domain Names"

    Non-ASCII characters are not supported in the `HostSNI` and `HostSNIRegexp` expressions, and so using them would invalidate the associated TCP router.
    Domain names containing non-ASCII characters must be provided as punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)).

!!! important "Regexp Syntax"

    `HostSNIRegexp` accepts an expression with zero or more groups enclosed by curly braces, which are called named regexps.
    Named regexps, of the form `{name:regexp}`, are the only expressions considered for regexp matching.
    The regexp name (`name` in the above example) is an arbitrary value, that exists only for historical reasons.

    Any `regexp` supported by [Go's regexp package](https://golang.org/pkg/regexp/) may be used.

!!! important "HostSNI & TLS"

    It is important to note that the Server Name Indication is an extension of the TLS protocol.
    Hence, only TLS routers will be able to specify a domain name with that rule.
    However, there is one special use case for HostSNI with non-TLS routers:
    when one wants a non-TLS router that matches all (non-TLS) requests,
    one should use the specific ```HostSNI(`*`)``` syntax.

!!! info "Combining Matchers Using Operators and Parenthesis"

    The usual AND (`&&`) and OR (`||`) logical operators can be used, with the expected precedence rules,
    as well as parentheses.

!!! info "Inverting a matcher"

    One can invert a matcher by using the `!` operator.

!!! important "Rule, Middleware, and Services"

    The rule is evaluated "before" any middleware has the opportunity to work, and "before" the request is forwarded to the service.

!!! important "ALPN ACME-TLS/1"

    It would be a security issue to let a user-defined router catch the response to
    an ACME TLS challenge previously initiated by Traefik.
    For this reason, the `ALPN` matcher is not allowed to match the `ACME-TLS/1`
    protocol, and Traefik returns an error if this is attempted.

### Priority

To avoid path overlap, routes are sorted, by default, in descending order using rules length. 
The priority is directly equal to the length of the rule, and so the longest length has the highest priority.

A value of `0` for the priority is ignored: `priority = 0` means that the default rules length sorting is used.

??? info "How default priorities are computed"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          rule: "ClientIP(`192.168.0.12`)"
          # ...
        Router-2:
          rule: "ClientIP(`192.168.0.0/24`)"
          # ...
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "ClientIP(`192.168.0.12`)"
        # ...
      [tcp.routers.Router-2]
        rule = "ClientIP(`192.168.0.0/24`)"
        # ...
    ```

    The table below shows that `Router-2` has a higher computed priority than `Router-1`.

    | Name     | Rule                                                        | Priority |
    |----------|-------------------------------------------------------------|----------|
    | Router-1 | ```ClientIP(`192.168.0.12`)```                              | 24       |
    | Router-2 | ```ClientIP(`192.168.0.0/24`)```                            | 26       |

    Which means that requests from `192.168.0.12` would go to Router-2 even though Router-1 is intended to specifically handle them.
    To achieve this intention, a priority (higher than 26) should be set on Router-1.

??? example "Setting priorities -- using the [File Provider](../../providers/file.md)"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          rule: "ClientIP(`192.168.0.12`)"
          entryPoints:
          - "web"
          service: service-1
          priority: 2
        Router-2:
          rule: "ClientIP(`192.168.0.0/24`)"
          entryPoints:
          - "web"
          priority: 1
          service: service-2
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "ClientIP(`192.168.0.12`)"
        entryPoints = ["web"]
        service = "service-1"
        priority = 2
      [tcp.routers.Router-2]
        rule = "ClientIP(`192.168.0.0/24`)"
        entryPoints = ["web"]
        priority = 1
        service = "service-2"
    ```

    In this configuration, the priority is configured so that `Router-1` will handle requests from `192.168.0.12`.

### Middlewares

You can attach a list of [middlewares](../../middlewares/overview.md) to each TCP router.
The middlewares will take effect only if the rule matches, and before connecting to the service.

!!! warning "The character `@` is not allowed to be used in the middleware name."

!!! tip "Middlewares order"

    Middlewares are applied in the same order as their declaration in **router**.

??? example "With a [middleware](../../middlewares/tcp/overview.md) -- using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.my-router]
        rule = "HostSNI(`*`)"
        # declared elsewhere
        middlewares = ["ipwhitelist"]
        service = "service-foo"
    ```

    ```yaml tab="YAML"
    ## Dynamic configuration
    tcp:
      routers:
        my-router:
          rule: "HostSNI(`*`)"
          # declared elsewhere
          middlewares:
          - ipwhitelist
          service: service-foo
    ```

### Services

You must attach a TCP [service](../services/index.md) per TCP router.
Services are the target for the router.

!!! important "TCP routers can only target TCP services (not HTTP services)."

### TLS

#### General

When a TLS section is specified,
it instructs Traefik that the current router is dedicated to TLS requests only (and that the router should ignore non-TLS requests).

By default, a router with a TLS section will terminate the TLS connections, meaning that it will send decrypted data to the services.

??? example "Router for TLS requests"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          rule: "HostSNI(`foo-domain`)"
          service: service-id
          # will terminate the TLS request by default
          tls: {}
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "HostSNI(`foo-domain`)"
        service = "service-id"
        # will terminate the TLS request by default
        [tcp.routers.Router-1.tls]
    ```

#### `passthrough`

As seen above, a TLS router will terminate the TLS connection by default.
However, the `passthrough` option can be specified to set whether the requests should be forwarded "as is", keeping all data encrypted.

It defaults to `false`.

??? example "Configuring passthrough"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          rule: "HostSNI(`foo-domain`)"
          service: service-id
          tls:
            passthrough: true
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "HostSNI(`foo-domain`)"
        service = "service-id"
        [tcp.routers.Router-1.tls]
          passthrough = true
    ```

#### `options`

The `options` field enables fine-grained control of the TLS parameters.
It refers to a [TLS Options](../../https/tls.md#tls-options) and will be applied only if a `HostSNI` rule is defined.

!!! example "Configuring the tls options"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-1:
          rule: "HostSNI(`foo-domain`)"
          service: service-id
          # will terminate the TLS request
          tls:
            options: foo

    tls:
      options:
        foo:
          minVersion: VersionTLS12
          cipherSuites:
            - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
            - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
            - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
            - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
            - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "HostSNI(`foo-domain`)"
        service = "service-id"
        # will terminate the TLS request
        [tcp.routers.Router-1.tls]
          options = "foo"

    [tls.options]
      [tls.options.foo]
        minVersion = "VersionTLS12"
        cipherSuites = [
          "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
          "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
          "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
          "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
          "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
        ]
    ```

#### `certResolver`

See [`certResolver` for HTTP router](./index.md#certresolver) for more information.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  routers:
    routerfoo:
      rule: "HostSNI(`snitest.com`)"
      tls:
        certResolver: foo
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.routers]
  [tcp.routers.routerfoo]
    rule = "HostSNI(`snitest.com`)"
    [tcp.routers.routerfoo.tls]
      certResolver = "foo"
```

#### `domains`

See [`domains` for HTTP router](./index.md#domains) for more information.

```yaml tab="File (YAML)"
## Dynamic configuration
tcp:
  routers:
    routerbar:
      rule: "HostSNI(`snitest.com`)"
      tls:
        certResolver: "bar"
        domains:
          - main: "snitest.com"
            sans:
              - "*.snitest.com"
```

```toml tab="File (TOML)"
## Dynamic configuration
[tcp.routers]
  [tcp.routers.routerbar]
    rule = "HostSNI(`snitest.com`)"
    [tcp.routers.routerbar.tls]
      certResolver = "bar"
      [[tcp.routers.routerbar.tls.domains]]
        main = "snitest.com"
        sans = ["*.snitest.com"]
```

## Configuring UDP Routers

!!! warning "The character `@` is not allowed in the router name"

### General

Similarly to TCP, as UDP is the transport layer, there is no concept of a request,
so there is no notion of an URL path prefix to match an incoming UDP packet with.
Furthermore, as there is no good TLS support at the moment for multiple hosts,
there is no Host SNI notion to match against either.
Therefore, there is no criterion that could be used as a rule to match incoming packets in order to route them.
So UDP "routers" at this time are pretty much only load-balancers in one form or another.

!!! important "Sessions and timeout"

	Even though UDP is connectionless (and because of that),
	the implementation of an UDP router in Traefik relies on what we (and a couple of other implementations) call a `session`.
	It basically means that some state is kept about an ongoing communication between a client and a backend,
	notably so that the proxy knows where to forward a response packet from a backend.
	As expected, a `timeout` is associated to each of these sessions,
	so that they get cleaned out if they go through a period of inactivity longer than a given duration. 
	Timeout can be configured using the `entryPoints.name.udp.timeout` option as described 
	under [entry points](../entrypoints/#udp-options).

### EntryPoints

If not specified, UDP routers will accept packets from all defined (UDP) entry points.
If one wants to limit the router scope to a set of entry points, one should set the entry points option.

??? example "Listens to Every Entry Point"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration

    udp:
      routers:
        Router-1:
          # By default, routers listen to all UDP entrypoints
          # i.e. "other", and "streaming".
          service: "service-1"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration

    [udp.routers]
      [udp.routers.Router-1]
        # By default, routers listen to all UDP entrypoints,
        # i.e. "other", and "streaming".
        service = "service-1"
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration

    entryPoints:
      # not used by UDP routers
      web:
        address: ":80"
      # used by UDP routers
      other:
        address: ":9090/udp"
      streaming:
        address: ":9191/udp"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration

    [entryPoints]
      # not used by UDP routers
      [entryPoints.web]
        address = ":80"
      # used by UDP routers
      [entryPoints.other]
        address = ":9090/udp"
      [entryPoints.streaming]
        address = ":9191/udp"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entrypoints.web.address=":80"
    --entrypoints.other.address=":9090/udp"
    --entrypoints.streaming.address=":9191/udp"
    ```

??? example "Listens to Specific Entry Points"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    udp:
      routers:
        Router-1:
          # does not listen on "other" entry point
          entryPoints:
            - "streaming"
          service: "service-1"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [udp.routers]
      [udp.routers.Router-1]
        # does not listen on "other" entry point
        entryPoints = ["streaming"]
        service = "service-1"
    ```

    **Static Configuration**

    ```yaml tab="File (YAML)"
    ## Static configuration

    entryPoints:
      web:
        address: ":80"
      other:
        address: ":9090/udp"
      streaming:
        address: ":9191/udp"
    ```

    ```toml tab="File (TOML)"
    ## Static configuration

    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.other]
        address = ":9090/udp"
      [entryPoints.streaming]
        address = ":9191/udp"
    ```

    ```bash tab="CLI"
    ## Static configuration
    --entrypoints.web.address=":80"
    --entrypoints.other.address=":9090/udp"
    --entrypoints.streaming.address=":9191/udp"
    ```

### Services

There must be one (and only one) UDP [service](../services/index.md) referenced per UDP router.
Services are the target for the router.

!!! important "UDP routers can only target UDP services (and not HTTP or TCP services)."

{!traefik-for-business-applications.md!}
