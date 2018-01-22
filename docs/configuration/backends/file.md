# File Backends

Like any other reverse proxy, Træfik can be configured with a file.

You have three choices:

- [Simple](/configuration/backends/file/#simple)
- [Rules in a Separate File](/configuration/backends/file/#rules-in-a-separate-file)
- [Multiple `.toml` Files](/configuration/backends/file/#multiple-toml-files)

To enable the file backend, you must either pass the `--file` option to the Træfik binary or put the `[file]` section (with or without inner settings) in the configuration file.

The configuration file allows managing both backends/frontends and HTTPS certificates (which are not [Let's Encrypt](https://letsencrypt.org) certificates generated through Træfik).

## Simple

Add your configuration at the end of the global configuration file `traefik.toml`:

```toml
defaultEntryPoints = ["http", "https"]

[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.org.cert"
      keyFile = "integration/fixtures/https/snitest.org.key"

[file]

# rules
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
    [backends.backend2.maxconn]
    amount = 10
    extractorfunc = "request.host"
    [backends.backend2.LoadBalancer]
    method = "drr"
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1
    [backends.backend2.servers.server2]
    url = "http://172.17.0.5:80"
    weight = 2

[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host:test.localhost"

  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  priority = 10

  # restrict access to this frontend to the specified list of IPv4/IPv6 CIDR Nets
  # an unset or empty list allows all Source-IPs to access
  # if one of the Net-Specifications are invalid, the whole list is invalid
  # and allows all Source-IPs to access.
  whitelistSourceRange = ["10.42.0.0/16", "152.89.1.33/32", "afed:be44::/16"]

  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"

  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
  rule = "Path:/test"

# HTTPS certificate
[[tlsConfiguration]]
  entryPoints = ["https"]
  [tlsConfiguration.certificate]
    certFile = "path/to/my.cert"
    keyFile = "path/to/my.key"
    
[[tlsConfiguration]]
  entryPoints = ["https"]
  [tlsConfiguration.certificate]
    certFile = "path/to/my/other.cert"
    keyFile = "path/to/my/other.key"
```

!!! note
    adding certificates directly to the entrypoint is still maintained but certificates declared in this way cannot be managed dynamically.
    It's recommended to use the file provider to declare certificates.

## Rules in a Separate File

Put your rules in a separate file, for example `rules.toml`:

```toml
# traefik.toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
      entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]

[file]
filename = "rules.toml"
```

```toml
# rules.toml
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
    [backends.backend2.maxconn]
    amount = 10
    extractorfunc = "request.host"
    [backends.backend2.LoadBalancer]
    method = "drr"
    [backends.backend2.servers.server1]
    url = "http://172.17.0.4:80"
    weight = 1
    [backends.backend2.servers.server2]
    url = "http://172.17.0.5:80"
    weight = 2

[frontends]
  [frontends.frontend1]
  backend = "backend2"
    [frontends.frontend1.routes.test_1]
    rule = "Host:test.localhost"
  [frontends.frontend2]
  backend = "backend1"
  passHostHeader = true
  priority = 10
  entrypoints = ["https"] # overrides defaultEntryPoints
    [frontends.frontend2.routes.test_1]
    rule = "Host:{subdomain:[a-z]+}.localhost"
  [frontends.frontend3]
  entrypoints = ["http", "https"] # overrides defaultEntryPoints
  backend = "backend2"
  rule = "Path:/test"
  
# HTTPS certificate
[[tlsConfiguration]]
  entryPoints = ["https"]
  [tlsConfiguration.certificate]
    certFile = "path/to/my.cert"
    keyFile = "path/to/my.key"
    
[[tlsConfiguration]]
  entryPoints = ["https"]
  [tlsConfiguration.certificate]
    certFile = "path/to/my/other.cert"
    keyFile = "path/to/my/other.key"

## Multiple `.toml` Files

You could have multiple `.toml` files in a directory (and recursively in its sub-directories):

```toml
[file]
directory = "/path/to/config/"
```

If you want Træfik to watch file changes automatically, just add:

```toml
[file]
watch = true
```
