# Routers

Connecting Requests to Services
{: .subtitle }

![routers](../../assets/img/routers.png)

A router is in charge of connecting incoming requests to the services that can handle them.
In the process, routers may use pieces of [middleware](../../middlewares/overview.md) to update the request, or act before forwarding the request to the service.

## Configuration Example

??? example "Requests /foo are Handled by service-foo -- Using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
      [http.routers]
        [http.routers.my-router]
          rule = "Path(`/foo`)"
          service = "service-foo"
    ```

    ```yaml tab="YAML"
      http:
        routers:
          my-router:
            rule: "Path(`/foo`)"
            service: service-foo
    ```

??? example "With a [middleware](../../middlewares/overview.md) -- using the [File Provider](../../providers/file.md)"

    ```toml tab="TOML"
      [http.routers]
        [http.routers.my-router]
          rule = "Path(`/foo`)"
          # declared elsewhere
          middlewares = ["authentication"]
          service = "service-foo"
    ```

    ```yaml tab="YAML"
      http:
        routers:
          my-router:
            rule: "Path(`/foo`)"
            # declared elsewhere
            middlewares:
            - authentication
            service: service-foo
    ```

??? example "Forwarding all (non-tls) requests on port 3306 to a database service"
    
    ```toml tab="TOML"
    ## Static configuration ##
    
    [entryPoints]
      [entryPoints.web]
        address = ":80"
      [entryPoints.mysql-default]
        address = ":3306"   
    
    ## Dynamic configuration ##
    
    [tcp]
      [tcp.routers]
        [tcp.routers.to-database]
          entryPoints = ["mysql-default"]
          # Catch every request (only available rule for non-tls routers. See below.)
          rule = "HostSNI(`*`)"
          service = "database"
    ```
    
    ```yaml tab="YAML"
    ## Static configuration ##
    
    entryPoints:
      web:
        address: ":80"
      mysql-default:
        address: ":3306"   
    
    ## Dynamic configuration ##
    
    tcp:
      routers:
        to-database:
          entryPoints:
          - "mysql-default"
          # Catch every request (only available rule for non-tls routers. See below.)
          rule: "HostSNI(`*`)"
          service: database
    ```

## Configuring HTTP Routers

### EntryPoints

If not specified, HTTP routers will accept requests from all defined entry points.
If you want to limit the router scope to a set of entry points, set the `entryPoints` option.

??? example "Listens to Every EntryPoint"
    
    ```toml tab="TOML"
    ## Static configuration ##
    
    [entryPoints]
      [entryPoints.web]
        # ...
      [entryPoints.web-secure]
        # ...
      [entryPoints.other]
        # ...
    
    
    ## Dynamic configuration ##
        
    [http.routers]
      [http.routers.Router-1]
        # By default, routers listen to every entry points
        rule = "Host(`traefik.io`)"
        service = "service-1"
    ```
    
    ```yaml tab="YAML"
    ## Static configuration ##
    
    entryPoints:
      web:
        # ...
      web-secure:
        # ...
      other:
        # ...
    
    ## Dynamic configuration ##
        
    http:
      routers:
        Router-1:
          # By default, routers listen to every entry points
          rule: "Host(`traefik.io`)"
          service: "service-1"
    ```

??? example "Listens to Specific EntryPoints"
    
    ```toml tab="TOML"
    ## Static configuration ##
    
    [entryPoints]
      [entryPoints.web]
        # ...
      [entryPoints.web-secure]
        # ...
      [entryPoints.other]
        # ...
        
    ## Dynamic configuration ##
    
    [http.routers]
      [http.routers.Router-1]
        # won't listen to entry point web
        entryPoints = ["web-secure", "other"]
        rule = "Host(`traefik.io`)"
        service = "service-1"
    ```
    
    ```yaml tab="YAML"
    ## Static configuration ##
    
    entryPoints:
      web:
        # ...
      web-secure:
        # ...
      other:
        # ...
        
    ## Dynamic configuration ##
    
    http:
      routers:
        Router-1:
          # won't listen to entry point web
          entryPoints:
          - "web-secure"
          - "other"
          rule: "Host(`traefik.io`)"
          service: "service-1"
    ```

### Rule

Rules are a set of matchers that determine if a particular request matches specific criteria.
If the rule is verified, the router becomes active, calls middlewares, and then forwards the request to the service.

??? example "Host is traefik.io"

    ```toml
    rule = "Host(`traefik.io`)"
    ```

??? example "Host is traefik.io OR Host is containo.us AND path is /traefik"

    ```toml
    rule = "Host(`traefik.io`) || (Host(`containo.us`) && Path(`/traefik`))"
    ```

The table below lists all the available matchers:

| Rule                                                                 | Description                                                                                                    |
|----------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|
| ```Headers(`key`, `value`)```                                        | Check if there is a key `key`defined in the headers, with the value `value`                                    |
| ```HeadersRegexp(`key`, `regexp`)```                                 | Check if there is a key `key`defined in the headers, with a value that matches the regular expression `regexp` |
| ```Host(`domain-1`, ...)```                                          | Check if the request domain targets one of the given `domains`.                                                |
| ```HostRegexp(`traefik.io`, `{subdomain:[a-z]+}.traefik.io`, ...)``` | Check if the request domain matches the given `regexp`.                                                        |
| `Method(`methods`, ...)`                                             | Check if the request method is one of the given `methods` (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`)            |
| ```Path(`path`, `/articles/{category}/{id:[0-9]+}`, ...)```          | Match exact request path. It accepts a sequence of literal and regular expression paths.                       |
| ```PathPrefix(`/products/`, `/articles/{category}/{id:[0-9]+}`)```   | Match request prefix path. It accepts a sequence of literal and regular expression prefix paths.               |
| ```Query(`foo=bar`, `bar=baz`)```                                    | Match` Query String parameters. It accepts a sequence of key=value pairs.                                      |

!!! important "Regexp Syntax"

    In order to use regular expressions with `Host` and `Path` expressions,
    you must declare an arbitrarily named variable followed by the colon-separated regular expression, all enclosed in curly braces.
    Any pattern supported by [Go's regexp package](https://golang.org/pkg/regexp/) may be used (example: `/posts/{id:[0-9]+}`).

!!! tip "Combining Matchers Using Operators and Parenthesis"

    You can combine multiple matchers using the AND (`&&`) and OR (`||`) operators. You can also use parenthesis.

!!! important "Rule, Middleware, and Services"

    The rule is evaluated "before" any middleware has the opportunity to work, and "before" the request is forwarded to the service.

!!! tip "Path Vs PathPrefix"

    Use `Path` if your service listens on the exact path only. For instance, `Path: /products` would match `/products` but not `/products/shoes`.

    Use a `*Prefix*` matcher if your service listens on a particular base path but also serves requests on sub-paths.
    For instance, `PathPrefix: /products` would match `/products` but also `/products/shoes` and `/products/shirts`.
    Since the path is forwarded as-is, your service is expected to listen on `/products`.

### Middlewares

You can attach a list of [middlewares](../../middlewares/overview.md) to each HTTP router.
The middlewares will take effect only if the rule matches, and before forwarding the request to the service.

### Service

You must attach a [service](../services/index.md) per router.
Services are the target for the router.

!!! note "HTTP Only"

    HTTP routers can only target HTTP services (not TCP services).

### TLS

#### General

 When a TLS section is specified, it instructs Traefik that the current router is dedicated to HTTPS requests only (and that the router should ignore HTTP (non TLS) requests).
Traefik will terminate the SSL connections (meaning that it will send decrypted data to the services).

??? example "Configuring the router to accept HTTPS requests only"

    ```toml tab="TOML"
    [http.routers]
      [http.routers.Router-1]
        rule = "Host(`foo-domain`) && Path(`/foo-path/`)"
        service = "service-id"
        # will terminate the TLS request
        [http.routers.Router-1.tls]
    ```
    
    ```yaml tab="YAML"
    http:
      routers:
        Router-1:
          rule: "Host(`foo-domain`) && Path(`/foo-path/`)"
          service: service-id
          # will terminate the TLS request
          tls: {}
    ```

!!! note "HTTPS & ACME"

    In the current version, with [ACME](../../https/acme.md) enabled, automatic certificate generation will apply to every router declaring a TLS section.
    
!!! note "Passthrough"

    On TCP routers, you can configure a passthrough option so that Traefik doesn't terminate the TLS connection.

!!! important "Routers for HTTP & HTTPS"

    If you need to define the same route for both HTTP and HTTPS requests, you will need to define two different routers: one with the tls section, one without.

    ??? example "HTTP & HTTPS routes"

        ```toml tab="TOML"
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

        ```yaml tab="YAML"
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

#### `options`

The `options` field enables fine-grained control of the TLS parameters.
It refers to a [TLS Options](../../https/tls.md#tls-options) and will be applied only if a `Host` rule is defined.

!!! note "Server Name Association"

    Even though one might get the impression that a TLS options reference is mapped to a router, or a router rule, one should realize that it is actually mapped only to the host name found in the `Host` part of the rule. Of course, there could also be several `Host` parts in a rule, in which case the TLS options reference would be mapped to as many host names.

    Another thing to keep in mind is: the TLS option is picked from the mapping mentioned above and based on the server name provided during the TLS handshake, and it all happens before routing actually occurs.

??? example "Configuring the TLS options"

    ```toml tab="TOML"
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
          "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
          "TLS_RSA_WITH_AES_256_GCM_SHA384"
        ]
    ```
    
    ```yaml tab="YAML"
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
          - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
          - TLS_RSA_WITH_AES_256_GCM_SHA384
    ```

!!! important "Conflicting TLS Options"

    Since a TLS options reference is mapped to a host name, if a configuration introduces a situation where the same host name (from a `Host` rule) gets matched with two TLS options references, a conflict occurs, such as in the example below:

    ```toml tab="TOML"
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

    ```yaml tab="YAML"
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

    If that happens, both mappings are discarded, and the host name (`snitest.com` in this case) for these routers gets associated with the default TLS options instead.

#### `certResolver`

If `certResolver` is defined, Traefik will try to generate certificates based on routers `Host` & `HostSNI` rules.

```toml tab="TOML"
[http.routers]
  [http.routers.routerfoo]
    rule = "Host(`snitest.com`) && Path(`/foo`)"
    [http.routers.routerfoo.tls]
      certResolver = "foo"
```

```yaml tab="YAML"
http:
  routers:
    routerfoo:
      rule: "Host(`snitest.com`) && Path(`/foo`)"
      tls:
        certResolver: foo
```

!!! note "Multiple Hosts in a Rule"
    The rule `Host(test1.traefik.io,test2.traefik.io)` will request a certificate with the main domain `test1.traefik.io` and SAN `test2.traefik.io`.

#### `domains`

You can set SANs (alternative domains) for each main domain.
Every domain must have A/AAAA records pointing to Traefik.
Each domain & SAN will lead to a certificate request.

```toml tab="TOML"
[http.routers]
  [http.routers.routerbar]
    rule = "Host(`snitest.com`) && Path(`/bar`)"
    [http.routers.routerbar.tls]
      certResolver = "bar"
      [[http.routers.routerbar.tls.domains]]
        main = "snitest.com"
        sans = "*.snitest.com"
```

```yaml tab="YAML"
http:
  routers:
    routerbar:
      rule: "Host(`snitest.com`) && Path(`/bar`)"
      tls:
        certResolver: "bar"
      domains:
      - main: "snitest.com"
        sans: "*.snitest.com"
```

[ACME v2](https://community.letsencrypt.org/t/acme-v2-and-wildcard-certificate-support-is-live/55579) supports wildcard certificates.
As described in [Let's Encrypt's post](https://community.letsencrypt.org/t/staging-endpoint-for-acme-v2/49605) wildcard certificates can only be generated through a [`DNS-01` challenge](./../../https/acme.md#dnschallenge).

Most likely the root domain should receive a certificate too, so it needs to be specified as SAN and 2 `DNS-01` challenges are executed.
In this case the generated DNS TXT record for both domains is the same.
Even though this behavior is [DNS RFC](https://community.letsencrypt.org/t/wildcard-issuance-two-txt-records-for-the-same-name/54528/2) compliant,
it can lead to problems as all DNS providers keep DNS records cached for a given time (TTL) and this TTL can be greater than the challenge timeout making the `DNS-01` challenge fail.

The Traefik ACME client library [LEGO](https://github.com/go-acme/lego) supports some but not all DNS providers to work around this issue.
The [Supported `provider` table](./../../https/acme.md#providers) indicates if they allow generating certificates for a wildcard domain and its root domain.

!!! note
    Wildcard certificates can only be verified through a `DNS-01` challenge.

!!! note "Double Wildcard Certificates"
    It is not possible to request a double wildcard certificate for a domain (for example `*.*.local.com`).

## Configuring TCP Routers

### General

If both HTTP routers and TCP routers listen to the same entry points, the TCP routers will apply *before* the HTTP routers.
If no matching route is found for the TCP routers, then the HTTP routers will take over.

### EntryPoints

If not specified, TCP routers will accept requests from all defined entry points.
If you want to limit the router scope to a set of entry points, set the entry points option.

??? example "Listens to Every Entry Point"

    ```toml tab="TOML"
    ## Static configuration ##
    
    [entryPoints]
      [entryPoints.web]
        # ...
      [entryPoints.web-secure]
        # ...
      [entryPoints.other]
        # ...
    
    ## Dynamic configuration ##
    
    [tcp.routers]
      [tcp.routers.Router-1]
        # By default, routers listen to every entrypoints
        rule = "HostSNI(`traefik.io`)"
        service = "service-1"
        # will route TLS requests (and ignore non tls requests)
        [tcp.routers.Router-1.tls]
    ```
    
    ```yaml tab="YAML"
    ## Static configuration ##
    
    entryPoints:
      web:
        # ...
      web-secure:
        # ...
      other:
        # ...
    
    ## Dynamic configuration ##
    
    tcp:
      routers:
        Router-1:
          # By default, routers listen to every entrypoints
          rule: "HostSNI(`traefik.io`)"
          service: "service-1"
          # will route TLS requests (and ignore non tls requests)
          tls: {}
    ```

??? example "Listens to Specific Entry Points"
    
    ```toml tab="TOML"
    ## Static configuration ##
    
    [entryPoints]
      [entryPoints.web]
        # ...
      [entryPoints.web-secure]
        # ...
      [entryPoints.other]
        # ...
        
    ## Dynamic configuration ##
    
    [tcp.routers]
      [tcp.routers.Router-1]
        # won't listen to entry point web
        entryPoints = ["web-secure", "other"]
        rule = "HostSNI(`traefik.io`)"
        service = "service-1"
        # will route TLS requests (and ignore non tls requests)
        [tcp.routers.Router-1.tls]
    ```
    
    ```yaml tab="YAML"
    ## Static configuration ##
    
    entryPoints:
      web:
        # ...
      web-secure:
        # ...
      other:
        # ...
        
    ## Dynamic configuration ##
    
    tcp:
      routers:
        Router-1:
          # won't listen to entry point web
          entryPoints:
          - "web-secure"
          - "other"
          rule: "HostSNI(`traefik.io`)"
          service: "service-1"
          # will route TLS requests (and ignore non tls requests)
          tls: {}
    ```

### Rule

| Rule                           | Description                                                             |
|--------------------------------|-------------------------------------------------------------------------|
| ```HostSNI(`domain-1`, ...)``` | Check if the Server Name Indication corresponds to the given `domains`. |

!!! important "HostSNI & TLS"

    It is important to note that the Server Name Indication is an extension of the TLS protocol.
    Hence, only TLS routers will be able to specify a domain name with that rule.
    However, non-TLS routers will have to explicitly use that rule with `*` (every domain) to state that every non-TLS request will be handled by the router.

### Services

You must attach a TCP [service](../services/index.md) per TCP router.
Services are the target for the router.

!!! note "TCP Only"

    TCP routers can only target TCP services (not HTTP services).

### TLS

#### General

 When a TLS section is specified, it instructs Traefik that the current router is dedicated to TLS requests only (and that the router should ignore non-TLS requests).
 By default, Traefik will terminate the SSL connections (meaning that it will send decrypted data to the services), but Traefik can be configured in order to let the requests pass through (keeping the data encrypted), and be forwarded to the service "as is". 

??? example "Configuring TLS Termination"

    ```toml tab="TOML"
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "HostSNI(`foo-domain`)"
        service = "service-id"
        # will terminate the TLS request by default
        [tcp.routers.Router-1.tls]
    ```

    ```yaml tab="YAML"
    tcp:
      routers:
        Router-1:
          rule: "HostSNI(`foo-domain`)"
          service: service-id
          # will terminate the TLS request by default
          tld: {}
    ```

??? example "Configuring passthrough"

    ```toml tab="TOML"
    [tcp.routers]
      [tcp.routers.Router-1]
        rule = "HostSNI(`foo-domain`)"
        service = "service-id"
        [tcp.routers.Router-1.tls]
          passthrough = true
    ```

    ```yaml tab="YAML"
    tcp:
      routers:
        Router-1:
          rule: "HostSNI(`foo-domain`)"
          service: service-id
          tls:
            passthrough: true
    ```

!!! note "TLS & ACME"

    In the current version, with [ACME](../../https/acme.md) enabled, automatic certificate generation will apply to every router declaring a TLS section.

#### `options`

The `options` field enables fine-grained control of the TLS parameters.  
It refers to a [TLS Options](../../https/tls.md#tls-options) and will be applied only if a `HostSNI` rule is defined.

??? example "Configuring the tls options"

    ```toml tab="TOML"
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
          "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
          "TLS_RSA_WITH_AES_256_GCM_SHA384"
        ]
    ```

    ```yaml tab="YAML"
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
          - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
          - "TLS_RSA_WITH_AES_256_GCM_SHA384"
    ```

#### `certResolver`

See [`certResolver` for HTTP router](./index.md#certresolver) for more information.

```toml tab="TOML"
[tcp.routers]
  [tcp.routers.routerfoo]
    rule = "HostSNI(`snitest.com`)"
    [tcp.routers.routerfoo.tls]
      certResolver = "foo"
```

```yaml tab="YAML"
tcp:
  routers:
    routerfoo:
      rule: "HostSNI(`snitest.com`)"
      tls:
        certResolver: foo
```

#### `domains`

See [`domains` for HTTP router](./index.md#domains) for more information.

```toml tab="TOML"
[tcp.routers]
  [tcp.routers.routerbar]
    rule = "HostSNI(`snitest.com`)"
    [tcp.routers.routerbar.tls]
      certResolver = "bar"
      [[tcp.routers.routerbar.tls.domains]]
        main = "snitest.com"
        sans = "*.snitest.com"
```

```yaml tab="YAML"
tcp:
  routers:
    routerbar:
      rule: "HostSNI(`snitest.com`)"
      tls:
        certResolver: "bar"
      domains:
      - main: "snitest.com"
        sans: "*.snitest.com"
```
