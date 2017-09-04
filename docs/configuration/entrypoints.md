# Entry Points Definition

```toml
# Entrypoints definition
#
# Default:
# [entryPoints]
#   [entryPoints.http]
#   address = ":80"
#
[entryPoints]
  [entryPoints.http]
  address = ":80"
```

## Redirect HTTP to HTTPS

To redirect an http entrypoint to an https entrypoint (with SNI support).

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
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"
```

## Rewriting URL

To redirect an entrypoint rewriting the URL.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    regex = "^http://localhost/(.*)"
    replacement = "http://mydomain/$1"
```

## TLS Mutual Authentication

Only accept clients that present a certificate signed by a specified Certificate Authority (CA).
`ClientCAFiles` can be configured with multiple `CA:s` in the same file or use multiple files containing one or several `CA:s`.
The `CA:s` has to be in PEM format.

All clients will be required to present a valid cert.
The requirement will apply to all server certs in the entrypoint.

In the example below both `snitest.com` and `snitest.org` will require client certs

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
  [entryPoints.https.tls]
  ClientCAFiles = ["tests/clientca1.crt", "tests/clientca2.crt"]
    [[entryPoints.https.tls.certificates]]
    CertFile = "integration/fixtures/https/snitest.com.cert"
    KeyFile = "integration/fixtures/https/snitest.com.key"
    [[entryPoints.https.tls.certificates]]
    CertFile = "integration/fixtures/https/snitest.org.cert"
    KeyFile = "integration/fixtures/https/snitest.org.key"
```


## Authentication

### Basic Authentication

Passwords can be encoded in MD5, SHA1 and BCrypt: you can use `htpasswd` to generate those ones.

Users can be specified directly in the toml file, or indirectly by referencing an external file;
 if both are provided, the two are merged, with external file contents having precedence.

```toml
# To enable basic auth on an entrypoint with 2 user/pass: test:test and test2:test2
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.basic]
  users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
  usersFile = "/path/to/.htpasswd"
```

### Digest Authentication

You can use `htdigest` to generate those ones.

Users can be specified directly in the toml file, or indirectly by referencing an external file;
 if both are provided, the two are merged, with external file contents having precedence

```toml
# To enable digest auth on an entrypoint with 2 user/realm/pass: test:traefik:test and test2:traefik:test2
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.http.auth.basic]
  users = ["test:traefik:a2688e031edb4be6a3797f3882655c05 ", "test2:traefik:518845800f9e2bfb1f1f740ec24f074e"]
  usersFile = "/path/to/.htdigest"
```

## Specify Minimum TLS Version

To specify an https entrypoint with a minimum TLS version, and specifying an array of cipher suites (from crypto/tls).

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
    MinVersion = "VersionTLS12"
    CipherSuites = ["TLS_RSA_WITH_AES_256_GCM_SHA384"]
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.com.cert"
      KeyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      CertFile = "integration/fixtures/https/snitest.org.cert"
      KeyFile = "integration/fixtures/https/snitest.org.key"
```

## Compression

To enable compression support using gzip format.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  compress = true
```

## Whitelisting

To enable IP whitelisting at the entrypoint level.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  whiteListSourceRange = ["127.0.0.1/32"]
```

## ProxyProtocol Support

To enable [ProxyProtocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt) support.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  proxyprotocol = true
```
