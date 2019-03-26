# HTTPS & TLS

Traefik supports HTTPS & TLS, and is able to accept new certificates / updates over time (without being restarted).
TLS is enabled at the [router](../routing/routers/index.md) level, but some options are configured in dedicated sections (`tlsOptions` & `tlsStores`) described in this section.

## Configuration Example

??? example "Configuring a Default Certificate"

    ```toml
    [tlsStores]
     [tlsStores.default]
       [tlsStores.default.defaultCertificate]
         certFile = "path/to/cert.crt"
         keyFile  = "path/to/cert.key"
    ```

??? example "Configuring a Minimum TLS Version"

    ```toml
    [tlsOptions]
        [tlsOptions.default]
        minVersion = "VersionTLS12"
    ```

??? example "Defining Certificates"

    ```toml
    [[tls]]
      [tls.certificate]
         certFile = "/path/to/domain.cert"
         keyFile = "/path/to/domain.key"

    [[tls]]
      [tls.certificate]
         certFile = "/path/to/other-domain.cert"
         keyFile = "/path/to/other-domain.key"
    ```

!!! important "File Provider Only"
    
    In the above example, we've used the [file provider](../providers/file.md) to handle the TLS configuration (tlsStores, tlsOptions, and TLS certificates).
    In its current alpha version, it is the only available method to configure these elements.
    Of course, these options are hot reloaded and can be updated at runtime (they belong to the [dynamic configuration](../getting-started/configuration-overview.md)).

## Configuration Options

### Dynamic Certificates

To add / remove TLS certificates while Traefik is running, the [file provider](../providers/file.md) supports Dynamic TLS certificates in its `[[tls]]` section.

!!! example "Defining Certificates"

    ```toml
    [[tls]]
      stores = ["default"]
      [tls.certificate]
         certFile = "/path/to/domain.cert"
         keyFile = "/path/to/domain.key"

    [[tls]]
      stores = ["default"]
      [tls.certificate]
         certFile = "/path/to/other-domain.cert"
         keyFile = "/path/to/other-domain.key"
    ```

    ??? note "Stores"

        During the alpha version, the stores option will be ignored and be automatically set to ["default"].

### Mutual Authentication

Traefik supports both optional and non optional (defaut value) mutual authentication.

- When `optional = false`, Traefik accepts connections only from client presenting a certificate signed by a CA listed in `ClientCA.files`.
- When `optional = true`, Traefik authorizes connections from client presenting a certificate signed by an unknown CA.

!!! example "Non Optional Mutual Authentication"

    In the following example, both `snitest.com` and `snitest.org` will require client certificates.

    ```toml
    [tlsOptions]
       [tlsOptions.default]
          [tlsOptions.default.ClientCA]
            files = ["tests/clientca1.crt", "tests/clientca2.crt"]
            optional = false
    ```

    ??? note "ClientCA.files"

        You can use a file per `CA:s`, or a single file containing multiple `CA:s` (in `PEM` format).

        `ClientCA.files` is not optional: every client will have to present a valid certificate. (This requirement will apply to every server certificate declared in the entrypoint.)

### Minimum TLS Version

!!! example "Min TLS version & [cipherSuites](https://godoc.org/crypto/tls#pkg-constants)"

    ```toml
    [tlsOptions]
      [tlsOptions.default]
          minVersion = "VersionTLS12"
          cipherSuites = [
            "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
            "TLS_RSA_WITH_AES_256_GCM_SHA384"
          ]
    ```

### Strict SNI Checking

With strict SNI checking, Traefik won't allow connections without a matching certificate.

!!! example "Strict SNI"

    ```toml
    [tlsOptions]
      [tlsOptions.default]
        sniStrict = true
    ```

### Default Certificate

Traefik can use a default certificate for connections without a SNI, or without a matching domain.

If no default certificate is provided, Traefik generates and uses a self-signed certificate.

!!! example "Setting a Default Certificate"

    ```toml
    [tlsStores]
     [tlsStores.default]
       [tlsStores.default.defaultCertificate]
         certFile = "path/to/cert.crt"
         keyFile  = "path/to/cert.key"
    ```

    ??? note "Only One Default Certificate"

        There can only be one `defaultCertificate` per tlsOptions.

    ??? note "Default TLS Store"

        During the alpha version, there is only one globally available TLS Store (`default`).
