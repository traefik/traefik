---
title: "Traefik QueryParameters Documentation"
description: "In Traefik Proxy, the QueryParameters middleware sets, adds and removes query parameters on the request URL before it is forwarded. Read the technical documentation."
---

The QueryParameters middleware sets, adds and removes query parameters on the request URL before the request is forwarded.

## Configuration Examples

```yaml tab="Docker & Swarm"
# Set c to 6, add d=4 and remove fbclid.
labels:
  - "traefik.http.middlewares.test-queryparameters.queryparameters.set.c=6"
  - "traefik.http.middlewares.test-queryparameters.queryparameters.add.d=4"
  - "traefik.http.middlewares.test-queryparameters.queryparameters.delete=fbclid"
```

```yaml tab="Kubernetes"
# Set c to 6, add d=4 and remove fbclid.
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-queryparameters
spec:
  queryParameters:
    set:
      c: "6"
    add:
      d: "4"
    delete:
      - fbclid
```

```yaml tab="Consul Catalog"
# Set c to 6, add d=4 and remove fbclid.
- "traefik.http.middlewares.test-queryparameters.queryparameters.set.c=6"
- "traefik.http.middlewares.test-queryparameters.queryparameters.add.d=4"
- "traefik.http.middlewares.test-queryparameters.queryparameters.delete=fbclid"
```

```yaml tab="File (YAML)"
# Set c to 6, add d=4 and remove fbclid.
http:
  middlewares:
    test-queryparameters:
      queryParameters:
        set:
          c: "6"
        add:
          d: "4"
        delete:
          - fbclid
```

```toml tab="File (TOML)"
# Set c to 6, add d=4 and remove fbclid.
[http.middlewares]
  [http.middlewares.test-queryparameters.queryParameters]
    delete = ["fbclid"]
    [http.middlewares.test-queryparameters.queryParameters.set]
      c = "6"
    [http.middlewares.test-queryparameters.queryParameters.add]
      d = "4"
```

With this configuration, a request to `/foo?a=1&c=3&fbclid=abc` is forwarded as `/foo?a=1&c=6&d=4`.

## Configuration Options

| Field    | Description                                                                                                                                                       | Default | Required |
|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|
| <a id="opt-set" href="#opt-set" title="#opt-set">`set`</a> | Query parameters to set. Every existing value of the parameter is replaced by the configured value, in the position of its first occurrence. A missing parameter is added at the end of the query. | `{}`    | No       |
| <a id="opt-add" href="#opt-add" title="#opt-add">`add`</a> | Query parameters to add at the end of the query. Existing values of the same parameter are kept, so `a=1` with `add: {a: "10"}` becomes `a=1&a=10`.                | `{}`    | No       |
| <a id="opt-delete" href="#opt-delete" title="#opt-delete">`delete`</a> | Names of the query parameters to remove, with every value they have.                                                                                              | `[]`    | No       |

The options are applied in that order: `delete`, then `set`, then `add`.
A parameter cannot be listed in `delete` and in `set` or `add` at the same time.
A parameter can be listed in both `set` and `add`, which gives it exactly those values: `set: {a: "5"}` with `add: {a: "10"}` turns `a=1&a=2` into `a=5&a=10`.

Parameter names are matched after URL decoding, so `delete: [fbclid]` also removes `f%62clid`.
The configured names and values are URL-encoded when they are written to the query.
The parameters the middleware does not change keep their original order and encoding.
