---
title: 'Open Policy Agent'
description: 'Traefik Hub API Gateway - The Open Policy Agent (OPA) middleware that allows you to restrict access to your services.'
---

!!! info "Traefik Hub Feature"
    This middleware is available exclusively in [Traefik Hub](https://traefik.io/traefik-hub/). Learn more about [Traefik Hub's advanced features](https://doc.traefik.io/traefik-hub/api-gateway/intro).

Traefik Hub comes with an Open Policy Agent middleware that allows you to restrict access to your services. It also allows you to enrich request headers with data extracted from policies.
The OPA middleware works as an [OPA agent](https://www.openpolicyagent.org/).

!!! note "OPA Version"

    This middleware uses the [v1.3.0 of the OPA specification](https://www.openpolicyagent.org/docs).

## Configuration Example

```yaml tab="Allow requests with specific JWT claim"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: opa-allow-jwt-claim
  namespace: apps
spec:
  plugin:
    opa:
      policy: |
        package example.policies

        allow {
          [_, encoded] := split(input.headers.Authorization, " ")
          [header, payload, signature] = io.jwt.decode(encoded)
          payload["email"] == "bibi@example.com"
        }
      forwardHeaders:
        Group: data.package.grp
```

```yaml tab="Deny requests with JSON Accept Header"
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: opa-deny-json
  namespace: apps
spec:
  plugin:
    opa:
      policy: |
        package example.policies

        default allow = false

        json_content {
          input.headers["Accept"] == "application/json"
        }

        allow {
          not json_content
        }
      allow: data.example.policies.allow
```

## Configuration Options

| Field    | Description   | Default | Required        |
|:---------|-----------------------|:--------|:----------------------------|
| `policy` | Path or the content of a [policy file](https://www.openpolicyagent.org/docs/v0.66.0/kubernetes-primer/#writing-policies). | ""      | No (one of `policy` or `bundlePath` must be set) |
| `bundlePath` | The `bundlePath` option should contain the path to an OPA [bundle](https://www.openpolicyagent.org/docs/v0.66.0/management-bundles/). | ""      | No (one of `policy` or `bundlePath` must be set) |
| `allow` | The `allow` option sets the expression to evaluate that determines if the request should be authorized. | ""      | No (one of `allow` or `forwardHeaders` must be set) |
| `forwardHeaders` | The `forwardHeaders` option sets the HTTP headers to add to requests and populates them with the result of the given expression. | ""      | No (one of `allow` or `forwardHeaders` must be set) |   

{!traefik-for-business-applications.md!}
