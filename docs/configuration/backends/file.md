# File Provider

Traefik can be configured with a file.

## Reference

```toml
[file]

# Backends
[backends]

  [backends.backend1]

    [backends.backend1.servers]
      [backends.backend1.servers.server0]
        url = "http://10.10.10.1:80"
        weight = 1
      [backends.backend1.servers.server1]
        url = "http://10.10.10.2:80"
        weight = 2
      # ...

    [backends.backend1.circuitBreaker]
      expression = "NetworkErrorRatio() > 0.5"
      
    [backends.backend1.responseForwarding]
      flushInterval = "10ms"

    [backends.backend1.loadBalancer]
      method = "drr"
      [backends.backend1.loadBalancer.stickiness]
        cookieName = "foobar"

    [backends.backend1.maxConn]
      amount = 10
      extractorfunc = "request.host"

    [backends.backend1.healthCheck]
      path = "/health"
      port = 88
      interval = "30s"
      scheme = "http"
      hostname = "myhost.com"
      [backends.backend1.healthcheck.headers]
        My-Custom-Header = "foo"
        My-Header = "bar"

  [backends.backend2]
    # ...

# Frontends
[frontends]

  [frontends.frontend1]
    entryPoints = ["http", "https"]
    backend = "backend1"
    passHostHeader = true
    priority = 42

    # Use frontends.frontend1.auth.basic below instead
    basicAuth = [
      "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
      "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
    ]
    [frontends.frontend1.passTLSClientCert]
        pem = true
        [frontends.frontend1.passTLSClientCert.infos]
            notBefore = true
            notAfter = true
            [frontends.frontend1.passTLSClientCert.infos.subject]
                country = true
                domainComponent = true
                province = true
                locality = true
                organization = true
                commonName = true
                serialNumber = true
            [frontends.frontend1.passTLSClientCert.infos.issuer]
                country = true
                domainComponent = true
                province = true
                locality = true
                organization = true
                commonName = true
                serialNumber = true
    [frontends.frontend1.auth]
      headerField = "X-WebAuth-User"
      [frontends.frontend1.auth.basic]
        removeHeader = true
        users = [
          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
          "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
        ]
        usersFile = "/path/to/.htpasswd"
      [frontends.frontend1.auth.digest]
        removeHeader = true
        users = [
          "test:traefik:a2688e031edb4be6a3797f3882655c05",
          "test2:traefik:518845800f9e2bfb1f1f740ec24f074e",
        ]
        usersFile = "/path/to/.htdigest"
      [frontends.frontend1.auth.forward]
        address = "https://authserver.com/auth"
        trustForwardHeader = true
        authResponseHeaders = ["X-Auth-User"]
        [frontends.frontend1.auth.forward.tls]
          ca = "path/to/local.crt"
          caOptional = true
          cert = "path/to/foo.cert"
          key = "path/to/foo.key"
          insecureSkipVerify = true

    [frontends.frontend1.whiteList]
      sourceRange = ["10.42.0.0/16", "152.89.1.33/32", "afed:be44::/16"]
      useXForwardedFor = true

    [frontends.frontend1.routes]
      [frontends.frontend1.routes.route0]
        rule = "Host:test.localhost"
      [frontends.frontend1.routes.Route1]
        rule = "Method:GET"
      # ...

    [frontends.frontend1.headers]
      allowedHosts = ["foobar", "foobar"]
      hostsProxyHeaders = ["foobar", "foobar"]
      SSLRedirect = true
      SSLTemporaryRedirect = true
      SSLHost = "foobar"
      STSSeconds = 42
      STSIncludeSubdomains = true
      STSPreload = true
      forceSTSHeader = true
      frameDeny = true
      customFrameOptionsValue = "foobar"
      contentTypeNosniff = true
      browserXSSFilter = true
      contentSecurityPolicy = "foobar"
      publicKey = "foobar"
      referrerPolicy = "foobar"
      isDevelopment = true
      [frontends.frontend1.headers.customRequestHeaders]
        X-Foo-Bar-01 = "foobar"
        X-Foo-Bar-02 = "foobar"
        # ...
      [frontends.frontend1.headers.customResponseHeaders]
        X-Foo-Bar-03 = "foobar"
        X-Foo-Bar-04 = "foobar"
        # ...
      [frontends.frontend1.headers.SSLProxyHeaders]
        X-Foo-Bar-05 = "foobar"
        X-Foo-Bar-06 = "foobar"
        # ...

    [frontends.frontend1.errors]
      [frontends.frontend1.errors.errorPage0]
        status = ["500-599"]
        backend = "error"
        query = "/{status}.html"
      [frontends.frontend1.errors.errorPage1]
        status = ["404", "403"]
        backend = "error"
        query = "/{status}.html"
      # ...

    [frontends.frontend1.ratelimit]
      extractorfunc = "client.ip"
        [frontends.frontend1.ratelimit.rateset.rateset1]
          period = "10s"
          average = 100
          burst = 200
        [frontends.frontend1.ratelimit.rateset.rateset2]
          period = "3s"
          average = 5
          burst = 10
        # ...

    [frontends.frontend1.redirect]
      entryPoint = "https"
      regex = "^http://localhost/(.*)"
      replacement = "http://mydomain/$1"
      permanent = true

  [frontends.frontend2]
    # ...

# HTTPS certificates
[[tls]]
  entryPoints = ["https"]
  [tls.certificate]
    certFile = "path/to/my.cert"
    keyFile = "path/to/my.key"

[[tls]]
  # ...
```

## Configuration Mode

You have two choices:

- [Rules in Traefik configuration file](/configuration/backends/file/#rules-in-traefik-configuration-file)
- [Rules in dedicated files](/configuration/backends/file/#rules-in-dedicated-files)

To enable the file backend, you must either pass the `--file` option to the Traefik binary or put the `[file]` section (with or without inner settings) in the configuration file.

The configuration file allows managing both backends/frontends and HTTPS certificates (which are not [Let's Encrypt](https://letsencrypt.org) certificates generated through Traefik).

TOML templating can be used if rules are not defined in the Traefik configuration file.

### Rules in Traefik Configuration File

Add your configuration at the end of the global configuration file `traefik.toml`:

```toml
defaultEntryPoints = ["http", "https"]

[entryPoints]
  [entryPoints.http]
    # ...
  [entryPoints.https]
    # ...

[file]

# rules
[backends]
  [backends.backend1]
    # ...
  [backends.backend2]
    # ...

[frontends]
  [frontends.frontend1]
  # ...
  [frontends.frontend2]
  # ...
  [frontends.frontend3]
  # ...

# HTTPS certificate
[[tls]]
  # ...

[[tls]]
  # ...
```

!!! note
    If `tls.entryPoints` is not defined, the certificate is attached to all the `defaultEntryPoints` with a TLS configuration.

!!! note
    Adding certificates directly to the entryPoint is still maintained but certificates declared in this way cannot be managed dynamically.
    It's recommended to use the file provider to declare certificates.

!!! warning
    TOML templating cannot be used if rules are defined in the Traefik configuration file.

### Rules in Dedicated Files

Traefik allows defining rules in one or more separate files.

#### One Separate File

You have to specify the file path in the `file.filename` option.

```toml
# traefik.toml
defaultEntryPoints = ["http", "https"]

[entryPoints]
  [entryPoints.http]
    # ...
  [entryPoints.https]
    # ...

[file]
  filename = "rules.toml"
  watch = true
```

The option `file.watch` allows Traefik to watch file changes automatically.

#### Multiple Separated Files

You could have multiple `.toml` files in a directory (and recursively in its sub-directories):

```toml
[file]
  directory = "/path/to/config/"
  watch = true
```

The option `file.watch` allows Traefik to watch file changes automatically.

#### Separate Files Content

If you are defining rules in one or more separate files, you can use two formats.

##### Simple Format

Backends, Frontends and TLS certificates are defined one at time, as described in the file `rules.toml`:

```toml
# rules.toml
[backends]
  [backends.backend1]
    # ...
  [backends.backend2]
    # ...

[frontends]
  [frontends.frontend1]
  # ...
  [frontends.frontend2]
  # ...
  [frontends.frontend3]
  # ...

# HTTPS certificate
[[tls]]
  # ...

[[tls]]
  # ...
```

##### TOML Templating

!!! warning
    TOML templating can only be used **if rules are defined in one or more separate files**.
    Templating will not work in the Traefik configuration file.

Traefik allows using TOML templating.

Thus, it's possible to define easily lot of Backends, Frontends and TLS certificates as described in the file `template-rules.toml` :

```toml
# template-rules.toml
[backends]
{{ range $i, $e := until 100 }}
  [backends.backend{{ $e }}]
    #...
{{ end }}

[frontends]
{{ range $i, $e := until 100 }}
  [frontends.frontend{{ $e }}]
    #...
{{ end }}


# HTTPS certificate
{{ range $i, $e := until 100 }}
[[tls]]
    #...
{{ end }}
```
