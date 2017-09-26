# HTTPS file provider

Træfik HTTPS certificates can be managed dynamically with a configuration file. 
This provider allows users to add/renew certificates which are not managed by [Let's Encrypt](https://letsencrypt.org).

## Configuration

### Træfik configuration file

Define a TLS Entrypoint and the file where the certificates configuration is stored.

```toml
[entryPoints]
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
    # Empty configuration

[httpsFile]
configurationFile = "certificates.toml"
watch = true
```

- `configurationFile` : Path to access to the configuration file (absolute path or relative to Træfik configuration file)
- `watch` : Add a watcher on the configuration file to detects the modifications.

### Certificates configuration file

```toml
[[Tls]]
EntryPoints = ["https"]
    [Tls.Certificate]
     CertFile = "integration/fixtures/https/snitest.com.cert"
     KeyFile = "integration/fixtures/https/snitest.com.key"
[[Tls]]
EntryPoints = ["https"]
    [Tls.Certificate]
     CertFile = """-----BEGIN CERTIFICATE-----
Cert_File_Content
-----END CERTIFICATE-----"""
     KeyFile = """-----BEGIN RSA PRIVATE KEY-----
Key_File_Content
-----END RSA PRIVATE KEY-----"""
```

- `EntryPoints` : Array which contains all the entrypoints which have to use the certificates.
- `Certificate` : Pair of certificate and key files. Can be the path of files or content.
