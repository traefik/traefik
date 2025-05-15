---
title: "Traefik PassTLSClientCert Documentation"
description: "In Traefik Proxy's HTTP middleware, the PassTLSClientCert adds selected data from passed client TLS certificates to headers. Read the technical documentation."
---

The `passTLSClientCert` middleware adds the selected data from the passed client TLS certificate to a header.

## Configuration Examples

Pass the pem in the `X-Forwarded-Tls-Client-Cert` header:

```yaml tab="Structured (YAML)"
# Pass the pem in the `X-Forwarded-Tls-Client-Cert` header.
http:
  middlewares:
    test-passtlsclientcert:
      passTLSClientCert:
        pem: true
```

```toml tab="Structured (TOML)"
# Pass the pem in the `X-Forwarded-Tls-Client-Cert` header.
[http.middlewares]
  [http.middlewares.test-passtlsclientcert.passTLSClientCert]
    pem = true
```

```yaml tab="Labels"
# Pass the pem in the `X-Forwarded-Tls-Client-Cert` header.
labels:
  - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.pem=true"
```

```json tab="Tags"
// Pass the pem in the `X-Forwarded-Tls-Client-Cert` header
{
  "Tags": [
    "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.pem=true"
  ]
}
```

```yaml tab="Kubernetes"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-passtlsclientcert
spec:
  passTLSClientCert:
    pem: true
```

??? example "Pass the pem in the `X-Forwarded-Tls-Client-Cert` header"

    ```yaml tab="Structured (YAML)"
    # Pass all the available info in the `X-Forwarded-Tls-Client-Cert-Info` header
    http:
      middlewares:
        test-passtlsclientcert:
          passTLSClientCert:
            info:
              notAfter: true
              notBefore: true
              sans: true
              subject:
                country: true
                province: true
                locality: true
                organization: true
                organizationalUnit: true
                commonName: true
                serialNumber: true
                domainComponent: true
              issuer:
                country: true
                province: true
                locality: true
                organization: true
                commonName: true
                serialNumber: true
                domainComponent: true
    ```

    ```toml tab="Structured (TOML)"
    # Pass all the available info in the `X-Forwarded-Tls-Client-Cert-Info` header
    [http.middlewares]
      [http.middlewares.test-passtlsclientcert.passTLSClientCert]
        [http.middlewares.test-passtlsclientcert.passTLSClientCert.info]
          notAfter = true
          notBefore = true
          sans = true
          [http.middlewares.test-passtlsclientcert.passTLSClientCert.info.subject]
            country = true
            province = true
            locality = true
            organization = true
            organizationalUnit = true
            commonName = true
            serialNumber = true
            domainComponent = true
          [http.middlewares.test-passtlsclientcert.passTLSClientCert.info.issuer]
            country = true
            province = true
            locality = true
            organization = true
            commonName = true
            serialNumber = true
            domainComponent = true
    ```

    ```yaml tab="Labels"
    # Pass all the available info in the `X-Forwarded-Tls-Client-Cert-Info` header
    labels:
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.notafter=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.notbefore=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.sans=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.serialnumber=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.commonname=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.country=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.domaincomponent=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.locality=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.organization=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.organizationalunit=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.province=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.serialnumber=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.commonname=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.country=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.domaincomponent=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.locality=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.organization=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.province=true"
      - "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.serialnumber=true"
    ```

    ```json tab="Tags"
    // Pass all the available info in the `X-Forwarded-Tls-Client-Cert-Info` header
    {
      //...
      "Tags" : [
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.notafter=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.notbefore=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.sans=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.commonname=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.country=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.domaincomponent=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.locality=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.organization=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.organizationalunit=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.province=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.subject.serialnumber=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.commonname=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.country=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.domaincomponent=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.locality=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.organization=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.province=true",
        "traefik.http.middlewares.test-passtlsclientcert.passtlsclientcert.info.issuer.serialnumber=true"
      ]
    }
    ```

    ```yaml tab="Kubernetes"
    # Pass all the available info in the `X-Forwarded-Tls-Client-Cert-Info` header
    apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: test-passtlsclientcert
    spec:
      passTLSClientCert:
        info:
          notAfter: true
          notBefore: true
          sans: true
          subject:
            country: true
            province: true
            locality: true
            organization: true
            organizationalUnit: true
            commonName: true
            serialNumber: true
            domainComponent: true
          issuer:
            country: true
            province: true
            locality: true
            organization: true
            commonName: true
            serialNumber: true
            domainComponent: true
    ```

## General Information

`passTLSClientCert` can add two headers to the request:

- `X-Forwarded-Tls-Client-Cert` that contains the pem.
- `X-Forwarded-Tls-Client-Cert-Info` that contains all the selected certificate information in an escaped string.

!!! info

    - `X-Forwarded-Tls-Client-Cert-Info` header value is a string that has been escaped in order to be a valid URL query.
    - These options only work accordingly to the MutualTLS configuration. i.e, only the certificates that match the `clientAuth.clientAuthType` policy are passed.

??? example "Example of a complete certificate and explaining each of the middleware options"

    ```txt
    Certificate:
        Data:
            Version: 3 (0x2)
            Serial Number: 1 (0x1)
            Signature Algorithm: sha256WithRSAEncryption
            Issuer: DC=org, DC=cheese, O=Cheese, O=Cheese 2, OU=Simple Signing Section, OU=Simple Signing Section 2, CN=Simple Signing CA, CN=Simple Signing CA 2, C=FR, C=US, L=TOULOUSE, L=LYON, ST=Signing State, ST=Signing State 2, emailAddress=simple@signing.com, emailAddress=simple2@signing.com
            Validity
                Not Before: May 15 04:13:46 2025 GMT
                Not After : May 15 04:13:46 2026 GMT
            Subject: DC=org, DC=cheese, O=Cheese, O=Cheese 2, OU=Simple Signing Section, OU=Simple Signing Section 2, CN=*.example.org, CN=*.example.com, C=FR, C=US, L=TOULOUSE, L=LYON, ST=Cheese org state, ST=Cheese com state, emailAddress=cert@example.org, emailAddress=cert@example.com
            Subject Public Key Info:
                Public Key Algorithm: rsaEncryption
                    Public-Key: (2048 bit)
                    Modulus:
                        00:b3:97:21:03:dd:65:24:43:b2:a7:19:c6:63:42:
                        9b:24:98:70:6b:d3:fc:1e:e2:7e:23:c4:92:4f:62:
                        92:7f:c5:68:e4:78:f0:a5:de:b8:f7:37:dc:b4:72:
                        b0:08:56:62:17:d5:f7:69:dd:94:8d:42:e1:c6:81:
                        08:3f:7f:f3:40:47:e0:c6:b4:79:30:4b:a8:e7:00:
                        56:ef:a3:28:27:bb:16:fe:33:24:7b:3a:9f:fd:72:
                        be:46:46:fd:a7:99:b0:a2:8f:d6:9c:f8:8a:01:ba:
                        a7:5f:f6:5b:aa:71:34:e2:7a:3b:13:ee:97:48:c8:
                        02:16:fe:66:5b:3e:b0:47:2d:65:20:5f:6b:83:d1:
                        51:11:1d:f9:9f:10:38:63:0a:ad:1a:1e:84:fc:95:
                        f1:4f:2a:91:22:4e:5f:9f:46:47:73:5d:8b:19:3f:
                        e0:1c:db:1d:13:3b:28:bc:d3:4b:73:28:a1:ad:24:
                        6a:af:09:1a:f3:54:3c:f3:07:4e:ae:ba:03:89:2c:
                        55:a4:99:92:d0:8a:ee:c9:54:b6:17:94:b8:76:16:
                        89:02:97:83:09:79:4a:cc:60:0e:1e:b3:ec:d4:13:
                        2c:af:0a:44:a8:7b:33:a2:c0:2f:5f:6b:cd:ed:b8:
                        92:bb:6f:b6:00:bd:9d:13:23:5c:c1:6e:e0:6c:66:
                        dd:97
                    Exponent: 65537 (0x10001)
            X509v3 extensions:
                X509v3 Key Usage: critical
                    Digital Signature, Key Encipherment
                X509v3 Basic Constraints:
                    CA:FALSE
                X509v3 Extended Key Usage:
                    TLS Web Server Authentication, TLS Web Client Authentication
                X509v3 Subject Key Identifier:
                    92:11:9E:12:92:9D:0F:4E:72:7C:F2:35:32:C2:C7:27:0E:59:A1:90
                X509v3 Authority Key Identifier:
                    DirName:/DC=org/DC=cheese/O=Cheese/O=Cheese 2/OU=Simple Signing Section/OU=Simple Signing Section 2/CN=Simple Signing CA/CN=Simple Signing CA 2/C=FR/C=US/L=TOULOUSE/L=LYON/ST=Signing State/ST=Signing State 2/emailAddress=simple@signing.com
                    serial:08:4A:00:44:82:3F:89:0F:B9:70:57:A1:88:47:05:BD:AD:9B:44:DA
                X509v3 Subject Alternative Name:
                    DNS:*.example.org, DNS:*.example.net, DNS:*.example.com, IP Address:10.0.1.0, IP Address:10.0.1.2, email:test@example.org, email:test@example.net
        Signature Algorithm: sha256WithRSAEncryption
        Signature Value:
            7d:79:93:0b:0e:2c:d5:43:9a:e5:11:39:f4:fe:14:d5:b0:7f:
            85:bc:c1:d9:f6:3f:e6:91:44:09:31:c0:c7:c6:6e:9a:6e:c4:
            91:4f:02:6f:ee:d4:a1:8b:7c:76:16:f3:e0:65:1a:de:1c:6e:
            06:65:67:8b:9e:ca:e9:d8:0a:52:34:c6:f4:78:5d:b1:07:7a:
            d2:7d:c0:26:87:ad:2b:7e:cb:02:47:a3:7c:a9:10:b8:8a:6e:
            11:6f:a7:39:0d:26:ed:d7:65:4a:39:4b:98:5d:62:34:04:33:
            aa:1e:d5:c1:04:58:5a:a9:b6:0f:d5:34:da:e8:32:6f:db:39:
            d5:9c:6c:8f:72:4d:d8:77:a7:23:3a:5b:56:41:6c:8b:e7:92:
            cf:6d:72:1a:c1:12:e1:56:63:38:8a:97:9c:6e:74:d1:b5:29:
            16:0d:c5:4e:11:a4:e6:3e:14:5e:14:8b:95:e7:c1:37:8d:dd:
            83:2f:a4:f4:0f:0c:8a:57:d1:20:5e:61:c6:69:10:06:49:3d:
            dc:2b:ec:fa:98:a9:a4:3b:19:21:7c:01:44:71:87:4b:b5:2e:
            59:e1:9e:f8:fc:9f:35:e6:0c:f3:e8:22:ae:ab:d5:4b:43:a0:
            a1:7e:c9:c1:04:27:47:29:b5:ee:d4:34:d7:8d:4a:80:b8:43:
            48:32:1b:98
    -----BEGIN CERTIFICATE-----
    MIIH+jCCBuKgAwIBAgIBATANBgkqhkiG9w0BAQsFADCCAYQxEzARBgoJkiaJk/Is
    ZAEZFgNvcmcxFjAUBgoJkiaJk/IsZAEZFgZjaGVlc2UxDzANBgNVBAoMBkNoZWVz
    ZTERMA8GA1UECgwIQ2hlZXNlIDIxHzAdBgNVBAsMFlNpbXBsZSBTaWduaW5nIFNl
    Y3Rpb24xITAfBgNVBAsMGFNpbXBsZSBTaWduaW5nIFNlY3Rpb24gMjEaMBgGA1UE
    AwwRU2ltcGxlIFNpZ25pbmcgQ0ExHDAaBgNVBAMME1NpbXBsZSBTaWduaW5nIENB
    IDIxCzAJBgNVBAYTAkZSMQswCQYDVQQGEwJVUzERMA8GA1UEBwwIVE9VTE9VU0Ux
    DTALBgNVBAcMBExZT04xFjAUBgNVBAgMDVNpZ25pbmcgU3RhdGUxGDAWBgNVBAgM
    D1NpZ25pbmcgU3RhdGUgMjEhMB8GCSqGSIb3DQEJARYSc2ltcGxlQHNpZ25pbmcu
    Y29tMSIwIAYJKoZIhvcNAQkBFhNzaW1wbGUyQHNpZ25pbmcuY29tMB4XDTI1MDUx
    NTA0MTM0NloXDTI2MDUxNTA0MTM0NlowggF5MRMwEQYKCZImiZPyLGQBGRYDb3Jn
    MRYwFAYKCZImiZPyLGQBGRYGY2hlZXNlMQ8wDQYDVQQKDAZDaGVlc2UxETAPBgNV
    BAoMCENoZWVzZSAyMR8wHQYDVQQLDBZTaW1wbGUgU2lnbmluZyBTZWN0aW9uMSEw
    HwYDVQQLDBhTaW1wbGUgU2lnbmluZyBTZWN0aW9uIDIxFjAUBgNVBAMMDSouZXhh
    bXBsZS5vcmcxFjAUBgNVBAMMDSouZXhhbXBsZS5jb20xCzAJBgNVBAYTAkZSMQsw
    CQYDVQQGEwJVUzERMA8GA1UEBwwIVE9VTE9VU0UxDTALBgNVBAcMBExZT04xGTAX
    BgNVBAgMEENoZWVzZSBvcmcgc3RhdGUxGTAXBgNVBAgMEENoZWVzZSBjb20gc3Rh
    dGUxHzAdBgkqhkiG9w0BCQEWEGNlcnRAZXhhbXBsZS5vcmcxHzAdBgkqhkiG9w0B
    CQEWEGNlcnRAZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
    AoIBAQCzlyED3WUkQ7KnGcZjQpskmHBr0/we4n4jxJJPYpJ/xWjkePCl3rj3N9y0
    crAIVmIX1fdp3ZSNQuHGgQg/f/NAR+DGtHkwS6jnAFbvoygnuxb+MyR7Op/9cr5G
    Rv2nmbCij9ac+IoBuqdf9luqcTTiejsT7pdIyAIW/mZbPrBHLWUgX2uD0VERHfmf
    EDhjCq0aHoT8lfFPKpEiTl+fRkdzXYsZP+Ac2x0TOyi800tzKKGtJGqvCRrzVDzz
    B06uugOJLFWkmZLQiu7JVLYXlLh2FokCl4MJeUrMYA4es+zUEyyvCkSoezOiwC9f
    a83tuJK7b7YAvZ0TI1zBbuBsZt2XAgMBAAGjggJ8MIICeDAOBgNVHQ8BAf8EBAMC
    BaAwCQYDVR0TBAIwADAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYD
    VR0OBBYEFJIRnhKSnQ9OcnzyNTLCxycOWaGQMIIBswYDVR0jBIIBqjCCAaahggGM
    pIIBiDCCAYQxEzARBgoJkiaJk/IsZAEZFgNvcmcxFjAUBgoJkiaJk/IsZAEZFgZj
    aGVlc2UxDzANBgNVBAoMBkNoZWVzZTERMA8GA1UECgwIQ2hlZXNlIDIxHzAdBgNV
    BAsMFlNpbXBsZSBTaWduaW5nIFNlY3Rpb24xITAfBgNVBAsMGFNpbXBsZSBTaWdu
    aW5nIFNlY3Rpb24gMjEaMBgGA1UEAwwRU2ltcGxlIFNpZ25pbmcgQ0ExHDAaBgNV
    BAMME1NpbXBsZSBTaWduaW5nIENBIDIxCzAJBgNVBAYTAkZSMQswCQYDVQQGEwJV
    UzERMA8GA1UEBwwIVE9VTE9VU0UxDTALBgNVBAcMBExZT04xFjAUBgNVBAgMDVNp
    Z25pbmcgU3RhdGUxGDAWBgNVBAgMD1NpZ25pbmcgU3RhdGUgMjEhMB8GCSqGSIb3
    DQEJARYSc2ltcGxlQHNpZ25pbmcuY29tMSIwIAYJKoZIhvcNAQkBFhNzaW1wbGUy
    QHNpZ25pbmcuY29tghQISgBEgj+JD7lwV6GIRwW9rZtE2jBmBgNVHREEXzBdgg0q
    LmV4YW1wbGUub3Jngg0qLmV4YW1wbGUubmV0gg0qLmV4YW1wbGUuY29thwQKAAEA
    hwQKAAECgRB0ZXN0QGV4YW1wbGUub3JngRB0ZXN0QGV4YW1wbGUubmV0MA0GCSqG
    SIb3DQEBCwUAA4IBAQB9eZMLDizVQ5rlETn0/hTVsH+FvMHZ9j/mkUQJMcDHxm6a
    bsSRTwJv7tShi3x2FvPgZRreHG4GZWeLnsrp2ApSNMb0eF2xB3rSfcAmh60rfssC
    R6N8qRC4im4Rb6c5DSbt12VKOUuYXWI0BDOqHtXBBFhaqbYP1TTa6DJv2znVnGyP
    ck3Yd6cjOltWQWyL55LPbXIawRLhVmM4ipecbnTRtSkWDcVOEaTmPhReFIuV58E3
    jd2DL6T0DwyKV9EgXmHGaRAGST3cK+z6mKmkOxkhfAFEcYdLtS5Z4Z74/J815gzz
    6CKuq9VLQ6ChfsnBBCdHKbXu1DTXjUqAuENIMhuY
    -----END CERTIFICATE-----
    ```

## Configuration Options

| Field      | Description        | Default | Required |
|:-----------|:------------------------------------------------------------|:--------|:---------|
| `pem` | Fills the `X-Forwarded-Tls-Client-Cert` header with the certificate information.<br /> More information [here](#pem). | false      | No      |
| `info.serialNumber` | Add the `Serial Number` of the certificate.<br /> More information about `info` [here](#info). | false | No |
| `info.notAfter` | Add the `Not After` information from the `Validity` part. <br /> More information about `info` [here](#info). | false | No |
| `info.notBefore` | Add the `Not Before` information from the `Validity` part. <br />More information about `info` [here](#info). | false      | No      |
| `info.sans` | Add the `Subject Alternative Name` information from the `Subject Alternative Name` part. <br /> More information about `info` [here](#info). | false      | No      |
| `info.subject` | The `info.subject` selects the specific client certificate subject details you want to add to the `X-Forwarded-Tls-Client-Cert-Info` header. <br />More information about `info` [here](#info). | false      | No      |
| `info.subject.country` | Add the `country` information into the subject.<br /> The data is taken from the subject part with the `C` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.subject.province` | Add the `province` information into the subject.<br /> The data is taken from the subject part with the `ST` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.subject.locality` | Add the `locality` information into the subject.<br /> The data is taken from the subject part with the `L` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.subject.organization` | Add the `organization` information into the subject.<br /> The data is taken from the subject part with the `O` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.subject.organizationalUnit` | Add the `organizationalUnit` information into the subject.<br /> The data is taken from the subject part with the `OU` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.subject.commonName` | Add the `commonName` information into the subject.<br /> The data is taken from the subject part with the `CN` key.| false      | No      |
| `info.subject.serialNumber` | Add the `serialNumber` information into the subject.<br /> The data is taken from the subject part with the `SN` key.| false      | No      |
| `info.subject.domainComponent` | Add the `domainComponent` information into the subject.<br />The data is taken from the subject part with the `DC` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer` | The `info.issuer` selects the specific client certificate issuer details you want to add to the `X-Forwarded-Tls-Client-Cert-Info` header. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.country` | Add the `country` information into the issuer.<br /> The data is taken from the issuer part with the `C` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.province` | Add the `province` information into the issuer.<br />The data is taken from the issuer part with the `ST` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.locality` | Add the `locality` information into the issuer.<br /> The data is taken from the issuer part with the `L` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.organization` | Add the `organization` information into the issuer.<br /> The data is taken from the issuer part with the `O` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.commonName` |Add the `commonName` information into the issuer.<br /> The data is taken from the issuer part with the `CN` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.serialNumber` |Add the `serialNumber` information into the issuer.<br /> The data is taken from the issuer part with the `SN` key. <br />More information about `info` [here](#info). | false      | No      |
| `info.issuer.domainComponent` | Add the `domainComponent` information into the issuer.<br /> The data is taken from the issuer part with the `DC` key. <br />More information about `info` [here](#info). | false      | No      |

### pem

#### Data Format

The delimiters and `\n` will be removed.  
If there are more than one certificate, they are separated by a "`,`".

#### Header size

The `X-Forwarded-Tls-Client-Cert` header value could exceed the web server header size limit

The header size limit of web servers is commonly between 4kb and 8kb.  
If that turns out to be a problem, and if reconfiguring the server to allow larger headers is not an option,
one can alleviate the problem by selecting only the interesting parts of the cert,
through the use of the `info` options described below. (And by setting `pem` to false).

### info

The `info` option selects the specific client certificate details you want to add to the `X-Forwarded-Tls-Client-Cert-Info` header.

#### Data Format

The value of the header is an escaped concatenation of all the selected certificate details.
Unless specified otherwise, all the header values examples are shown unescaped, for readability.

If there are more than one certificate, they are separated by a `,`.

The following example shows such a concatenation, when all the available fields are selected:

```text
Subject="DC=org,DC=cheese,C=FR,C=US,ST=Cheese org state,ST=Cheese com state,L=TOULOUSE,L=LYON,O=Cheese,O=Cheese 2,CN=*.example.com";Issuer="DC=org,DC=cheese,C=FR,C=US,ST=Signing State,ST=Signing State 2,L=TOULOUSE,L=LYON,O=Cheese,O=Cheese 2,CN=Simple Signing CA 2";NB="1747282426";NA="1778818426"SAN="*.example.org,*.example.net,*.example.com,test@example.org,test@example.net,10.0.1.0,10.0.1.2"
```
