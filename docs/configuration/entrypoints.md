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

!!! note
    Please note that `regex` and `replacement` do not have to be set in the `redirect` structure if an entrypoint is defined for the redirection (they will not be used in this case).

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

!!! note
    Please note that `regex` and `replacement` do not have to be set in the `redirect` structure if an entrypoint is defined for the redirection (they will not be used in this case).

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

### Forward Authentication

This configuration will first forward the request to `http://authserver.com/auth`.

If the response code is 2XX, access is granted and the original request is performed.
Otherwise, the response from the auth server is returned.

```toml
[entryPoints]
  [entrypoints.http]
    # ...
    # To enable forward auth on an entrypoint
    [entrypoints.http.auth.forward]
    address = "https://authserver.com/auth"
    
    # Trust existing X-Forwarded-* headers.
    # Useful with another reverse proxy in front of Traefik.
    #
    # Optional
    # Default: false
    #
    trustForwardHeader = true
    
    # Enable forward auth TLS connection.
    #
    # Optional
    #
    [entrypoints.http.auth.forward.tls]
    cert = "authserver.crt"
    key = "authserver.key"
```

## Specify Minimum TLS Version

To specify an https entry point with a minimum TLS version, and specifying an array of cipher suites (from crypto/tls).

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
    minVersion = "VersionTLS12"
    cipherSuites = ["TLS_RSA_WITH_AES_256_GCM_SHA384"]
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.com.cert"
      keyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.org.cert"
      keyFile = "integration/fixtures/https/snitest.org.key"
```

## Compression

To enable compression support using gzip format.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  compress = true
```

Responses are compressed when:

* The response body is larger than `512` bytes
* And the `Accept-Encoding` request header contains `gzip`
* And the response is not already compressed, i.e. the `Content-Encoding` response header is not already set.

## Whitelisting

To enable IP whitelisting at the entrypoint level.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
  whiteListSourceRange = ["127.0.0.1/32", "192.168.1.7"]
```

## ProxyProtocol

To enable [ProxyProtocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt) support.
Only IPs in `trustedIPs` will lead to remote client address replacement: you should declare your load-balancer IP or CIDR range here (in testing environment, you can trust everyone using `insecure = true`).

!!! danger
    When queuing Træfik behind another load-balancer, be sure to carefully configure Proxy Protocol on both sides.
    Otherwise, it could introduce a security risk in your system by forging requests. 

```toml
[entryPoints]
  [entryPoints.http]
    address = ":80"

    # Enable ProxyProtocol
    [entryPoints.http.proxyProtocol]
      # List of trusted IPs
      #
      # Required
      # Default: []
      #
      trustedIPs = ["127.0.0.1/32", "192.168.1.7"]

      # Insecure mode FOR TESTING ENVIRONNEMENT ONLY
      #
      # Optional
      # Default: false
      #
      # insecure = true
```

## Forwarded Header

Only IPs in `trustedIPs` will be authorize to trust the client forwarded headers (`X-Forwarded-*`).

```toml
[entryPoints]
  [entryPoints.http]
    address = ":80"

    # Enable Forwarded Headers
    [entryPoints.http.forwardedHeaders]
      # List of trusted IPs
      #
      # Required
      # Default: []
      #
      trustedIPs = ["127.0.0.1/32", "192.168.1.7"]
```
