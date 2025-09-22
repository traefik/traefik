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

## Configuration Options

| Field      | Description        | Default | Required |
|:-----------|:------------------------------------------------------------|:--------|:---------|
| <a id="pem" href="#pem" title="#pem">`pem`</a> | Fills the `X-Forwarded-Tls-Client-Cert` header with the certificate information.<br /> More information [here](#pem). | false      | No      |
| <a id="info-serialNumber" href="#info-serialNumber" title="#info-serialNumber">`info.serialNumber`</a> | Add the `Serial Number` of the certificate.<br /> More information about `info` [here](#info). | false | No |
| <a id="info-notAfter" href="#info-notAfter" title="#info-notAfter">`info.notAfter`</a> | Add the `Not After` information from the `Validity` part. <br /> More information about `info` [here](#info). | false | No |
| <a id="info-notBefore" href="#info-notBefore" title="#info-notBefore">`info.notBefore`</a> | Add the `Not Before` information from the `Validity` part. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-sans" href="#info-sans" title="#info-sans">`info.sans`</a> | Add the `Subject Alternative Name` information from the `Subject Alternative Name` part. <br /> More information about `info` [here](#info). | false      | No      |
| <a id="info-subject" href="#info-subject" title="#info-subject">`info.subject`</a> | The `info.subject` selects the specific client certificate subject details you want to add to the `X-Forwarded-Tls-Client-Cert-Info` header. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-subject-country" href="#info-subject-country" title="#info-subject-country">`info.subject.country`</a> | Add the `country` information into the subject.<br /> The data is taken from the subject part with the `C` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-subject-province" href="#info-subject-province" title="#info-subject-province">`info.subject.province`</a> | Add the `province` information into the subject.<br /> The data is taken from the subject part with the `ST` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-subject-locality" href="#info-subject-locality" title="#info-subject-locality">`info.subject.locality`</a> | Add the `locality` information into the subject.<br /> The data is taken from the subject part with the `L` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-subject-organization" href="#info-subject-organization" title="#info-subject-organization">`info.subject.organization`</a> | Add the `organization` information into the subject.<br /> The data is taken from the subject part with the `O` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-subject-organizationalUnit" href="#info-subject-organizationalUnit" title="#info-subject-organizationalUnit">`info.subject.organizationalUnit`</a> | Add the `organizationalUnit` information into the subject.<br /> The data is taken from the subject part with the `OU` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-subject-commonName" href="#info-subject-commonName" title="#info-subject-commonName">`info.subject.commonName`</a> | Add the `commonName` information into the subject.<br /> The data is taken from the subject part with the `CN` key.| false      | No      |
| <a id="info-subject-serialNumber" href="#info-subject-serialNumber" title="#info-subject-serialNumber">`info.subject.serialNumber`</a> | Add the `serialNumber` information into the subject.<br /> The data is taken from the subject part with the `SN` key.| false      | No      |
| <a id="info-subject-domainComponent" href="#info-subject-domainComponent" title="#info-subject-domainComponent">`info.subject.domainComponent`</a> | Add the `domainComponent` information into the subject.<br />The data is taken from the subject part with the `DC` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer" href="#info-issuer" title="#info-issuer">`info.issuer`</a> | The `info.issuer` selects the specific client certificate issuer details you want to add to the `X-Forwarded-Tls-Client-Cert-Info` header. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-country" href="#info-issuer-country" title="#info-issuer-country">`info.issuer.country`</a> | Add the `country` information into the issuer.<br /> The data is taken from the issuer part with the `C` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-province" href="#info-issuer-province" title="#info-issuer-province">`info.issuer.province`</a> | Add the `province` information into the issuer.<br />The data is taken from the issuer part with the `ST` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-locality" href="#info-issuer-locality" title="#info-issuer-locality">`info.issuer.locality`</a> | Add the `locality` information into the issuer.<br /> The data is taken from the issuer part with the `L` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-organization" href="#info-issuer-organization" title="#info-issuer-organization">`info.issuer.organization`</a> | Add the `organization` information into the issuer.<br /> The data is taken from the issuer part with the `O` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-commonName" href="#info-issuer-commonName" title="#info-issuer-commonName">`info.issuer.commonName`</a> |Add the `commonName` information into the issuer.<br /> The data is taken from the issuer part with the `CN` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-serialNumber" href="#info-issuer-serialNumber" title="#info-issuer-serialNumber">`info.issuer.serialNumber`</a> |Add the `serialNumber` information into the issuer.<br /> The data is taken from the issuer part with the `SN` key. <br />More information about `info` [here](#info). | false      | No      |
| <a id="info-issuer-domainComponent" href="#info-issuer-domainComponent" title="#info-issuer-domainComponent">`info.issuer.domainComponent`</a> | Add the `domainComponent` information into the issuer.<br /> The data is taken from the issuer part with the `DC` key. <br />More information about `info` [here](#info). | false      | No      |

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
