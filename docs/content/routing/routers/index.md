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

If not specified, HTTP routers will accept requests from all EntryPoints in the [list of default EntryPoints](../entrypoints.md#asdefault).
If you want to limit the router scope to a set of entry points, set the `entryPoints` option.

??? example "Listens to Every EntryPoint"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          # By default, routers listen to every EntryPoints.
          rule: "Host(`example.com`)"
          service: "service-1"
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        # By default, routers listen to every EntryPoints.
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
    --entryPoints.web.address=:80
    --entryPoints.websecure.address=:443
    --entryPoints.other.address=:9090
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
    --entryPoints.web.address=:80
    --entryPoints.websecure.address=:443
    --entryPoints.other.address=:9090
    ```

### Rule

Rules are a set of matchers configured with values, that determine if a particular request matches specific criteria.
If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

The table below lists all the available matchers:

| Rule                                                            | Description                                                                    |
|-----------------------------------------------------------------|:-------------------------------------------------------------------------------|
| [```Header(`key`, `value`)```](#header-and-headerregexp)        | Matches requests containing a header named `key` set to `value`.               |
| [```HeaderRegexp(`key`, `regexp`)```](#header-and-headerregexp) | Matches requests containing a header named `key` matching `regexp`.            |
| [```Host(`domain`)```](#host-and-hostregexp)                    | Matches requests host set to `domain`.                                         |
| [```HostRegexp(`regexp`)```](#host-and-hostregexp)              | Matches requests host matching `regexp`.                                       |
| [```Method(`method`)```](#method)                               | Matches requests method set to `method`.                                       |
| [```Path(`path`)```](#path-pathprefix-and-pathregexp)           | Matches requests path set to `path`.                                           |
| [```PathPrefix(`prefix`)```](#path-pathprefix-and-pathregexp)   | Matches requests path prefix set to `prefix`.                                  |
| [```PathRegexp(`regexp`)```](#path-pathprefix-and-pathregexp)   | Matches request path using `regexp`.                                           |
| [```Query(`key`, `value`)```](#query-and-queryregexp)           | Matches requests query parameters named `key` set to `value`.                  |
| [```QueryRegexp(`key`, `regexp`)```](#query-and-queryregexp)    | Matches requests query parameters named `key` matching `regexp`.               |
| [```ClientIP(`ip`)```](#clientip)                               | Matches requests client IP using `ip`. It accepts IPv4, IPv6 and CIDR formats. |

!!! tip "Backticks or Quotes?"

    To set the value of a rule, use [backticks](https://en.wiktionary.org/wiki/backtick) ``` ` ``` or escaped double-quotes `\"`.

    Single quotes `'` are not accepted since the values are [Go's String Literals](https://golang.org/ref/spec#String_literals).

!!! important "Regexp Syntax"

    Matchers that accept a regexp as their value use a [Go](https://golang.org/pkg/regexp/) flavored syntax.

!!! info "Expressing Complex Rules Using Operators and Parenthesis"

    The usual AND (`&&`) and OR (`||`) logical operators can be used, with the expected precedence rules,
    as well as parentheses.

    One can invert a matcher by using the NOT (`!`) operator.

    The following rule matches requests where:

    - either host is `example.com` OR,
    - host is `example.org` AND path is NOT `/traefik`

    ```yaml
    Host(`example.com`) || (Host(`example.org`) && !Path(`/traefik`))
    ```

#### Header and HeaderRegexp

The `Header` and `HeaderRegexp` matchers allow to match requests that contain specific header.

!!! example "Examples"

    Match requests with a `Content-Type` header set to `application/yaml`:

    ```yaml
    Header(`Content-Type`, `application/yaml`)
    ```

    Match requests with a `Content-Type` header set to either `application/json` or `application/yaml`:

    ```yaml
    HeaderRegexp(`Content-Type`, `^application/(json|yaml)$`)
    ```

    To match headers [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity), use the `(?i)` option:

    ```yaml
    HeaderRegexp(`Content-Type`, `(?i)^application/(json|yaml)$`)
    ```

#### Host and HostRegexp

The `Host` and `HostRegexp` matchers allow to match requests that are targeted to a given host.

These matchers do not support non-ASCII characters, use punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)) to match such domains.

If no Host is set in the request URL (e.g., it's an IP address), these matchers will look at the `Host` header.

These matchers will match the request's host in lowercase.

!!! example "Examples"

    Match requests with `Host` set to `example.com`:

    ```yaml
    Host(`example.com`)
    ```

    Match requests sent to any subdomain of `example.com`:

    ```yaml
    HostRegexp(`^.+\.example\.com$`)
    ```

    Match requests with `Host` set to either `example.com` or `example.org`:

    ```yaml
    HostRegexp(`^example\.(com|org)$`)
    ```

    To match domains [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity), use the `(?i)` option:

    ```yaml
    HostRegexp(`(?i)^example\.(com|org)$`)
    ```

#### Method

The `Method` matchers allows to match requests sent with the given method.

!!! example "Example"

    Match `OPTIONS` requests:

    ```yaml
    Method(`OPTIONS`)
    ```

#### Path, PathPrefix, and PathRegexp

These matchers allow matching requests based on their URL path.

For exact matches, use `Path` and its prefixed alternative `PathPrefix`, for regexp matches, use `PathRegexp`.

Path are always starting with a `/`, except for `PathRegexp`.

!!! example "Examples"

    Match `/products` but neither `/products/shoes` nor `/products/`:

    ```yaml
    Path(`/products`)
    ```

    Match `/products` as well as everything under `/products`,
    such as `/products/shoes`, `/products/` but also `/products-for-sale`:

    ```yaml
    PathPrefix(`/products`)
    ```

    Match both `/products/shoes` and `/products/socks` with and ID like `/products/shoes/57`:

    ```yaml
    PathRegexp(`^/products/(shoes|socks)/[0-9]+$`)
    ```

    Match requests with a path ending in either `.jpeg`, `.jpg` or `.png`:

    ```yaml
    PathRegexp(`\.(jpeg|jpg|png)$`)
    ```

    Match `/products` as well as everything under `/products`,
    such as `/products/shoes`, `/products/` but also `/products-for-sale`,
    [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity):

    ```yaml
    PathRegexp(`(?i)^/products`)
    ```

#### Query and QueryRegexp

The `Query` and `QueryRegexp` matchers allow to match requests based on query parameters.

!!! example "Examples"

    Match requests with a `mobile` query parameter set to `true`, such as in `/search?mobile=true`:

    ```yaml
    Query(`mobile`, `true`)
    ```

    To match requests with a query parameter `mobile` that has no value, such as in `/search?mobile`, use:

    ```yaml
    Query(`mobile`)
    ```

    Match requests with a `mobile` query parameter set to either `true` or `yes`:

    ```yaml
    QueryRegexp(`mobile`, `^(true|yes)$`)
    ```

    Match requests with a `mobile` query parameter set to any value (including the empty value):

    ```yaml
    QueryRegexp(`mobile`, `^.*$`)
    ```

    To match query parameters [case-insensitively](https://en.wikipedia.org/wiki/Case_sensitivity), use the `(?i)` option:

    ```yaml
    QueryRegexp(`mobile`, `(?i)^(true|yes)$`)
    ```

#### ClientIP

The `ClientIP` matcher allows matching requests sent from the given client IP.

It only matches the request client IP and does not use the `X-Forwarded-For` header for matching.

!!! example "Examples"

    Match requests coming from a given IP:

    ```yaml tab="IPv4"
    ClientIP(`10.76.105.11`)
    ```

    ```yaml tab="IPv6"
    ClientIP(`::1`)
    ```

    Match requests coming from a given subnet:

    ```yaml tab="IPv4"
    ClientIP(`192.168.1.0/24`)
    ```

    ```yaml tab="IPv6"
    ClientIP(`fe80::/10`)
    ```

### Priority

To avoid path overlap, routes are sorted, by default, in descending order using rules length.
The priority is directly equal to the length of the rule, and so the longest length has the highest priority.

A value of `0` for the priority is ignored: `priority = 0` means that the default rules length sorting is used.

??? warning "Maximum Value"
  
    Traefik reserves a range of priorities for its internal routers,
    the maximum user-defined router priority value is:

      - `(MaxInt32 - 1000)` for 32-bit platforms,
      - `(MaxInt64 - 1000)` for 64-bit platforms.

??? info "How default priorities are computed"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          rule: "HostRegexp(`[a-z]+\.traefik\.com`)"
          # ...
        Router-2:
          rule: "Host(`foobar.traefik.com`)"
          # ...
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-1]
        rule = "HostRegexp(`[a-z]+\\.traefik\\.com`)"
        # ...
      [http.routers.Router-2]
        rule = "Host(`foobar.traefik.com`)"
        # ...
    ```

    In this case, all requests with host `foobar.traefik.com` will be routed through `Router-1` instead of `Router-2`.

    | Name     | Rule                                     | Priority |
    |----------|------------------------------------------|----------|
    | Router-1 | ```HostRegexp(`[a-z]+\.traefik\.com`)``` | 34       |
    | Router-2 | ```Host(`foobar.traefik.com`)```         | 26       |

    The previous table shows that `Router-1` has a higher priority than `Router-2`.

    To solve this issue, the priority must be set.

??? example "Set priorities -- using the [File Provider](../../providers/file.md)"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-1:
          rule: "HostRegexp(`[a-z]+\\.traefik\\.com`)"
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
        rule = "HostRegexp(`[a-z]+\\.traefik\\.com`)"
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

### RuleSyntax

_Optional, Default=""_

In Traefik v3 a new rule syntax has been introduced ([migration guide](../../migration/v2-to-v3.md#router-rule-matchers)).
`ruleSyntax` option allows to configure the rule syntax to be used for parsing the rule on a per-router basis.
This allows to have heterogeneous router configurations and ease migration.

The default value of the `ruleSyntax` option is inherited from the `defaultRuleSyntax` option in the static configuration.
By default, the `defaultRuleSyntax` static option is `v3`, meaning that the default rule syntax is also `v3`.

??? example "Set rule syntax -- using the [File Provider](../../providers/file.md)"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    http:
      routers:
        Router-v3:
          rule: HostRegexp(`[a-z]+\\.traefik\\.com`)
          ruleSyntax: v3
        Router-v2:
          rule: HostRegexp(`{subdomain:[a-z]+}.traefik.com`)
          ruleSyntax: v2
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [http.routers]
      [http.routers.Router-v3]
        rule = "HostRegexp(`[a-z]+\\.traefik\\.com`)"
        ruleSyntax = v3
      [http.routers.Router-v2]
        rule = "HostRegexp(`{subdomain:[a-z]+}.traefik.com`)"
        ruleSyntax = v2
    ```

    ```yaml tab="Kubernetes traefik.io/v1alpha1"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRoute
    metadata:
      name: test.route
      namespace: default
    
    spec:
      routes:
        # route v3
        - match: HostRegexp(`[a-z]+\\.traefik\\.com`)
          syntax: v3
          kind: Rule

        # route v2
        - match: HostRegexp(`{subdomain:[a-z]+}.traefik.com`)
          syntax: v2
          kind: Rule
    ```

    In this configuration, the ruleSyntax is configured to allow `Router-v2` to use v2 syntax,
    while for `Router-v3` it is configured to use v3 syntax.

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
See the specific [docker](../providers/docker.md#service-definition) documentation.

!!! warning "The character `@` is not authorized in the service name."

!!! important "HTTP routers can only target HTTP services (not TCP services)."

### TLS

#### General

When a TLS section is specified, it instructs Traefik that the current router is dedicated to HTTPS requests only
(and that the router should ignore HTTP (non TLS) requests).
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
    The rule ```Host(`test1.example.com`) || Host(`test2.example.com`)``` will request a certificate with the main domain `test1.example.com` and SAN `test2.example.com`.

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

If both HTTP routers and TCP routers listen to the same EntryPoint, the TCP routers will apply _before_ the HTTP routers.
If no matching route is found for the TCP routers, then the HTTP routers will take over.

### EntryPoints

If not specified, TCP routers will accept requests from all EntryPoints in the [list of default EntryPoints](../entrypoints.md#asdefault).
If you want to limit the router scope to a set of entry points, set the entry points option.

??? info "How to handle Server First protocols?"

    To correctly handle a request, Traefik needs to wait for the first few bytes to arrive before it can decide what to do with it.

    For protocols where the server is expected to send first, such as SMTP, if no specific setup is in place,
    we could end up in a situation where both sides are waiting for data and the connection appears to have hanged.

    The only way that Traefik can deal with such a case, is to make sure that on the concerned entry point,
    there is no TLS router whatsoever (neither TCP nor HTTP),
    and there is at least one non-TLS TCP router that leads to the server in question.

??? example "Listens to Every Entry Point"

    **Dynamic Configuration**

    ```yaml tab="File (YAML)"
    ## Dynamic configuration

    tcp:
      routers:
        Router-1:
          # By default, routers listen to every EntryPoints.
          rule: "HostSNI(`example.com`)"
          service: "service-1"
          # will route TLS requests (and ignore non tls requests)
          tls: {}
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration

    [tcp.routers]
      [tcp.routers.Router-1]
        # By default, routers listen to every EntryPoints.
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
    --entryPoints.web.address=:80
    --entryPoints.websecure.address=:443
    --entryPoints.other.address=:9090
    ```

??? example "Listens to Specific EntryPoints"

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
    --entryPoints.web.address=:80
    --entryPoints.websecure.address=:443
    --entryPoints.other.address=:9090
    ```

### Rule

Rules are a set of matchers configured with values, that determine if a particular connection matches specific criteria.
If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

The table below lists all the available matchers:

| Rule                                                        | Description                                                                                      |
|-------------------------------------------------------------|:-------------------------------------------------------------------------------------------------|
| [```HostSNI(`domain`)```](#hostsni-and-hostsniregexp)       | Checks if the connection's Server Name Indication is equal to `domain`.                          |
| [```HostSNIRegexp(`regexp`)```](#hostsni-and-hostsniregexp) | Checks if the connection's Server Name Indication matches `regexp`.                              |
| [```ClientIP(`ip`)```](#clientip_1)                         | Checks if the connection's client IP correspond to `ip`. It accepts IPv4, IPv6 and CIDR formats. |<!-- markdownlint-disable-line MD051 -->
| [```ALPN(`protocol`)```](#alpn)                             | Checks if the connection's ALPN protocol equals `protocol`.                                      |

!!! tip "Backticks or Quotes?"

     To set the value of a rule, use [backticks](https://en.wiktionary.org/wiki/backtick) ``` ` ``` or escaped double-quotes `\"`.

    Single quotes `'` are not accepted since the values are [Go's String Literals](https://golang.org/ref/spec#String_literals).

!!! important "Regexp Syntax"

    Matchers that accept a regexp as their value use a [Go](https://golang.org/pkg/regexp/) flavored syntax.

!!! info "Expressing Complex Rules Using Operators and Parenthesis"

    The usual AND (`&&`) and OR (`||`) logical operators can be used, with the expected precedence rules,
    as well as parentheses.

    One can invert a matcher by using the NOT (`!`) operator.

    The following rule matches connections where:

    - either Server Name Indication is `example.com` OR,
    - Server Name Indication is `example.org` AND ALPN protocol is NOT `h2`

    ```yaml
    HostSNI(`example.com`) || (HostSNI(`example.org`) && !ALPN(`h2`))
    ```

#### HostSNI and HostSNIRegexp

`HostSNI` and `HostSNIRegexp` matchers allow to match connections targeted to a given domain.

These matchers do not support non-ASCII characters, use punycode encoded values ([rfc 3492](https://tools.ietf.org/html/rfc3492)) to match such domains.

!!! important "HostSNI & TLS"

    It is important to note that the Server Name Indication is an extension of the TLS protocol.
    Hence, only TLS routers will be able to specify a domain name with that rule.
    However, there is one special use case for HostSNI with non-TLS routers:
    when one wants a non-TLS router that matches all (non-TLS) requests,
    one should use the specific ```HostSNI(`*`)``` syntax.

!!! example "Examples"

    Match all connections:

    ```yaml tab="HostSNI"
    HostSNI(`*`)
    ```

    ```yaml tab="HostSNIRegexp"
    HostSNIRegexp(`^.*$`)
    ```

    Match TCP connections sent to `example.com`:

    ```yaml
    HostSNI(`example.com`)
    ```

    Match TCP connections openned on any subdomain of `example.com`:

    ```yaml
    HostSNIRegexp(`^.+\.example\.com$`)
    ```

#### ClientIP

The `ClientIP` matcher allows matching connections opened by a client with the given IP.

!!! example "Examples"

    Match connections opened by a given IP:

    ```yaml tab="IPv4"
    ClientIP(`10.76.105.11`)
    ```

    ```yaml tab="IPv6"
    ClientIP(`::1`)
    ```

    Match connections coming from a given subnet:

    ```yaml tab="IPv4"
    ClientIP(`192.168.1.0/24`)
    ```

    ```yaml tab="IPv6"
    ClientIP(`fe80::/10`)
    ```

#### ALPN

The `ALPN` matcher allows matching connections the given protocol.

It would be a security issue to let a user-defined router catch the response to
an ACME TLS challenge previously initiated by Traefik.
For this reason, the `ALPN` matcher is not allowed to match the `ACME-TLS/1`
protocol, and Traefik returns an error if this is attempted.

!!! example "Example"

    Match connections using the ALPN protocol `h2`:

    ```yaml
    ALPN(`h2`)
    ```

### Priority

To avoid path overlap, routes are sorted, by default, in descending order using rules length.
The priority is directly equal to the length of the rule, and so the longest length has the highest priority.

A value of `0` for the priority is ignored: `priority = 0` means that the default rules length sorting is used.

??? warning "Maximum Value"

    Traefik reserves a range of priorities for its internal routers,
    the maximum user-defined router priority value is:

      - `(MaxInt32 - 1000)` for 32-bit platforms,
      - `(MaxInt64 - 1000)` for 64-bit platforms.

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
    To achieve this intention, a priority (greater than 26) should be set on Router-1.

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

### RuleSyntax

_Optional, Default=""_

In Traefik v3 a new rule syntax has been introduced ([migration guide](../../migration/v2-to-v3.md#router-rule-matchers)).
`ruleSyntax` option allows to configure the rule syntax to be used for parsing the rule on a per-router basis.
This allows to have heterogeneous router configurations and ease migration.

The default value of the `ruleSyntax` option is inherited from the `defaultRuleSyntax` option in the static configuration.
By default, the `defaultRuleSyntax` static option is `v3`, meaning that the default rule syntax is also `v3`.

??? example "Set rule syntax -- using the [File Provider](../../providers/file.md)"

    ```yaml tab="File (YAML)"
    ## Dynamic configuration
    tcp:
      routers:
        Router-v3:
          rule: ClientIP(`192.168.0.11`) || ClientIP(`192.168.0.12`)
          ruleSyntax: v3
        Router-v2:
          rule: ClientIP(`192.168.0.11`, `192.168.0.12`)
          ruleSyntax: v2
    ```

    ```toml tab="File (TOML)"
    ## Dynamic configuration
    [tcp.routers]
      [tcp.routers.Router-v3]
        rule = "ClientIP(`192.168.0.11`) || ClientIP(`192.168.0.12`)"
        ruleSyntax = v3
      [tcp.routers.Router-v2]
        rule = "ClientIP(`192.168.0.11`, `192.168.0.12`)"
        ruleSyntax = v2
    ```

    ```yaml tab="Kubernetes traefik.io/v1alpha1"
    apiVersion: traefik.io/v1alpha1
    kind: IngressRouteTCP
    metadata:
      name: test.route
      namespace: default
    
    spec:
      routes:
        # route v3
        - match: ClientIP(`192.168.0.11`) || ClientIP(`192.168.0.12`)
          syntax: v3
          kind: Rule

        # route v2
        - match: ClientIP(`192.168.0.11`, `192.168.0.12`)
          syntax: v2
          kind: Rule
    ```

    In this configuration, the ruleSyntax is configured to allow `Router-v2` to use v2 syntax,
    while for `Router-v3` it is configured to use v3 syntax.

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
        middlewares = ["ipallowlist"]
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
          - ipallowlist
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

??? info "Postgres STARTTLS"

    Traefik supports the Postgres STARTTLS protocol,
    which allows TLS routing for Postgres connections.

    To do so, Traefik reads the first bytes sent by a Postgres client,
    identifies if they correspond to the message of a STARTTLS negotiation,
    and, if so, acknowledges and signals the client that it can start the TLS handshake.

    Please note/remember that there are subtleties inherent to STARTTLS in whether the connection ends up being a TLS one or not.
    These subtleties depend on the `sslmode` value in the client configuration (and on the server authentication rules).
    Therefore, it is recommended to use the `require` value for the `sslmode`.

    Afterwards, the TLS handshake, and routing based on TLS, can proceed as expected.

    !!! warning "Postgres STARTTLS with TCP TLS PassThrough routers"

        As mentioned above, the `sslmode` configuration parameter does have an impact on whether a STARTTLS session will succeed.
        In particular in the context of TCP TLS PassThrough, some of the values (such as `allow`) do not even make sense.
        Which is why, once more it is recommended to use the `require` value.

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
	Timeout can be configured using the `entryPoints.name.udp.timeout` option as described under [EntryPoints](../entrypoints/#udp-options).

### EntryPoints

If not specified, UDP routers will accept packets from all defined (UDP) EntryPoints.
If one wants to limit the router scope to a set of EntryPoints, one should set the `entryPoints` option.

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
    --entryPoints.web.address=":80"
    --entryPoints.other.address=":9090/udp"
    --entryPoints.streaming.address=":9191/udp"
    ```

??? example "Listens to Specific EntryPoints"

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
    --entryPoints.web.address=":80"
    --entryPoints.other.address=":9090/udp"
    --entryPoints.streaming.address=":9191/udp"
    ```

### Services

There must be one (and only one) UDP [service](../services/index.md) referenced per UDP router.
Services are the target for the router.

!!! important "UDP routers can only target UDP services (and not HTTP or TCP services)."

{!traefik-for-business-applications.md!}
