---
title: 'API Key Authentication'
description: 'Traefik Hub API Gateway - The API Key authentication middleware allows you to secure an API by requiring a secret key, base64 encoded or not, to be given, via an HTTP header, a cookie or a query parameter.'
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

The API Key authentication middleware allows you to secure an API by requiring a secret key, base64 encoded or not, to be given, via an HTTP header, a cookie or a query parameter.

---

## Configuration Example

```yaml tab="Middleware API Key"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-apikey
  namespace: apps
spec:
  plugin:
    apiKey:
      keySource:
        headerAuthScheme: Bearer
        header: Authorization
      secretNonBase64Encoded: true
      secretValues:
        - "urn:k8s:secret:apikey:secret"
        - "urn:k8s:secret:apikey:othersecret" 
```

```yaml tab="Values Secret"
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: apikey
  namespace: whoami
stringData:
  secret: $2y$05$D4SPFxzfWKcx1OXfVhRbvOTH/QB0Lm6AXTk8.NOmU4rPLX2t6UUuW # htpasswd -nbB "" foo | cut -c 2-
  othersecret: $2y$05$HbLL.g5dUqJippH0RuAGL.RaM9wNS2cT7hp6.vbv5okdCmVBSDzzK # htpasswd -nbB "" bar | cut -c 2-
```

## Configuration Options

| Field                        | Description   | Default | Required |
|:-----------------------------|:------------------------------------------------|:--------|:---------|
| <a id="opt-keySource-header" href="#opt-keySource-header" title="#opt-keySource-header">`keySource.header`</a> | Defines the header name containing the secret sent by the client.<br /> Either `keySource.header` or `keySource.query` or `keySource.cookie` must be set.                                                 | ""      | No       |
| <a id="opt-keySource-headerAuthScheme" href="#opt-keySource-headerAuthScheme" title="#opt-keySource-headerAuthScheme">`keySource.headerAuthScheme`</a> | Defines the scheme when using `Authorization` as header name. <br /> Check out the `Authorization` header [documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization#syntax). | ""      | No       |
| <a id="opt-keySource-query" href="#opt-keySource-query" title="#opt-keySource-query">`keySource.query`</a> | Defines the query parameter name containing the secret sent by the client.<br /> Either `keySource.header` or `keySource.query` or `keySource.cookie` must be set.                                       | ""      | No       |
| <a id="opt-keySource-cookie" href="#opt-keySource-cookie" title="#opt-keySource-cookie">`keySource.cookie`</a> | Defines the cookie name containing the secret sent by the client.<br /> Either `keySource.header` or `keySource.query` or `keySource.cookie` must be set.                                                | ""      | No       |
| <a id="opt-secretNonBase64Encoded" href="#opt-secretNonBase64Encoded" title="#opt-secretNonBase64Encoded">`secretNonBase64Encoded`</a> | Defines whether the secret sent by the client is base64 encoded. | false   | No       |
| <a id="opt-secretValues" href="#opt-secretValues" title="#opt-secretValues">`secretValues`</a> | Contain the hash of the API keys. <br /> Supported hashing algorithms are Bcrypt, SHA1 and MD5. <br /> The hash should be generated using `htpasswd`.<br />Can reference a Kubernetes Secret using the URN format: `urn:k8s:secret:[name]:[valueKey]` | []      | Yes      |

{!traefik-for-business-applications.md!}
