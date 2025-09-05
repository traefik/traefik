---
title: 'Secure API Access with JWT'
description: 'Traefik Hub API Gateway - Learn how to configure the JWT Authentication middleware for Ingress management.'
---

# Secure API Access with JWT

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

JSON Web Token (JWT) (defined in the [RFC 7519](https://tools.ietf.org/html/rfc7519)) allows
Traefik Hub API Gateway to secure the API access using a token signed using either a private signing secret or a plublic/private key.

Traefik Hub API Gateway provides many kinds of sources to perform the token validation:

- Setting a secret value in the middleware configuration (option `signingSecret`).
- Setting a public key: In that case, users should sign their token using a private key, and the public key can be used to verify the signature (option `publicKey`).
- Setting a [JSON Web Key (JWK)](https://datatracker.ietf.org/doc/html/rfc7517) file to define a set of JWK to be used to verify the signature of the incoming JWT (option `jwksFile`).
- Setting a [JSON Web Key (JWK)](https://datatracker.ietf.org/doc/html/rfc7517) URL to define the URL of the host serving a JWK set (option `jwksUrl`).

!!! note "One single source"
    The JWT middleware does not allow you to set more than one way to validate the incoming tokens.
    When a Hub API Gateway receives a request that must be validated using the JWT middleware, it verifies the token using the source configured as described above.
    If the token is successfully checked, the request is accepted.

!!! note "Claim Usage"
    A JWT can contain metadata in the form of claims (key-value pairs).
    The claims contained in the JWT can be used for advanced use-cases such as adding an Authorization layer using the `claims`.

    More information in the [dedicated section](../reference/routing-configuration/http/middlewares/jwt.md#claims).

## Verify a JWT with a secret

To allow the Traefik Hub API Gateway to validate a JWT with a secret value stored in a Kubernetes Secret, apply the following configuration:

```yaml tab="Middleware JWT"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-jwt
  namespace: apps
spec:
  plugin:
    jwt:
      signingSecret: "urn:k8s:secret:jwt:signingSecret"
```

```yaml tab="Kubernetes Secret"
apiVersion: v1
kind: Secret
metadata:
  name: jwt
  namespace: apps
stringData:
  signingSecret: mysuperlongsecret
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-app
  namespace: apps
spec:
  entryPoints:
    - websecure
  routes:
  - match: Path(`/my-app`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: test-jwt
```

```yaml tab="Service & Deployment"
kind: Deployment
apiVersion: apps/v1
metadata:
  name: whoami
  namespace: apps
spec:
  replicas: 3
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - name: whoami
        image: traefik/whoami

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: apps
spec:
  ports:
  - port: 80
    name: whoami
  selector:
    app: whoami
```

## Verify a JWT using an Identity Provider

To allow the Traefik Hub API Gateway to validate a JWT using an Identity Provider, such as Keycloak and Azure AD in the examples below, apply the following configuration:

```yaml tab="JWKS with Keycloak URL"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-jwt
  namespace: apps
spec:
  plugin:
    jwt:
      # Replace KEYCLOAK_URL and REALM_NAME with your values
      jwksUrl: https://KEYCLOAK_URL/realms/REALM_NAME/protocol/openid-connect/certs
      # Forward the content of the claim grp in the header Group
      forwardHeaders:
        Group: grp
      # Check the value of the claim grp before sending the request to the backend
      claims: Equals(`grp`, `admin`)
```

```yaml tab="JWKS with Azure AD URL"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-jwt
  namespace: apps
spec:
  plugin:
    jwt:
      jwksUrl: https://login.microsoftonline.com/common/discovery/v2.0/keys
```

```yaml tab="IngressRoute"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: my-app
  namespace: apps
spec:
  entryPoints:
    - websecure
  routes:
  - match: Path(`/my-app`)
    kind: Rule
    services:
    - name: whoami
      port: 80
    middlewares:
    - name: test-jwt
``` 

```yaml tab="Service & Deployment"
kind: Deployment
apiVersion: apps/v1
metadata:
  name: whoami
  namespace: apps
spec:
  replicas: 3
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
      - name: whoami
        image: traefik/whoami

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: apps
spec:
  ports:
  - port: 80
    name: whoami
  selector:
    app: whoami
```

!!! note "Advanced Configuration"
    Advanced options are described in the [reference page](../reference/routing-configuration/http/middlewares/jwt.md).

    For example, the metadata recovered from the Identity Provider can be used to restrict the access to the applications.
    To do so, you can use the `claims` option, more information in the [dedicated section](../reference/routing-configuration/http/middlewares/jwt.md#claims).

{!traefik-for-business-applications.md!}
