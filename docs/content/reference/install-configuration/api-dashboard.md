---
title: "Traefik API & Dashboard Documentation"
description: "Traefik Proxy exposes information through API handlers and showcase them on the Dashboard. Learn about the security, configuration, and endpoints of the APIs and Dashboard. Read the technical documentation."
---

The dashboard is the central place that shows you the current active routes handled by Traefik.

<figure>
    <img src="../../../assets/img/webui-dashboard.png" alt="Dashboard - Providers" />
    <figcaption>The dashboard in action</figcaption>
</figure>

## Configuration Example

Enable the dashboard:

```yaml tab="File(YAML)"
api: {}
```

```toml tab="File(TOML)"
[api]
```

```cli tab="CLI"
--api=true
```

Expose the dashboard:

```yaml tab="Kubernetes CRD"
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: traefik-dashboard
spec:
  routes:
  - match: Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))
    kind: Rule
    services:
    - name: api@internal
      kind: TraefikService
    middlewares:
      - name: auth
---
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: auth
spec:
  basicAuth:
    secret: secretName # Kubernetes secret named "secretName"
```

```yaml tab="Helm Chart Values (values.yaml)"
# Create an IngressRoute for the dashboard
ingressRoute:
  dashboard:
    enabled: true
    # Custom match rule with host domain
    matchRule: Host(`traefik.example.com`)
    entryPoints: ["websecure"]
    # Add custom middlewares : authentication and redirection
    middlewares:
      - name: traefik-dashboard-auth

# Create the custom middlewares used by the IngressRoute dashboard (can also be created in another way).
# /!\ Yes, you need to replace "changeme" password with a better one. /!\
extraObjects:
  - apiVersion: v1
    kind: Secret
    metadata:
      name: traefik-dashboard-auth-secret
    type: kubernetes.io/basic-auth
    stringData:
      username: admin
      password: changeme

  - apiVersion: traefik.io/v1alpha1
    kind: Middleware
    metadata:
      name: traefik-dashboard-auth
    spec:
      basicAuth:
        secret: traefik-dashboard-auth-secret
```

```yaml tab="Docker"
# Dynamic Configuration
labels:
  - "traefik.http.routers.dashboard.rule=Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))"
  - "traefik.http.routers.dashboard.service=api@internal"
  - "traefik.http.routers.dashboard.middlewares=auth"
  - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```yaml tab="Swarm"
# Dynamic Configuration
deploy:
  labels:
    - "traefik.http.routers.dashboard.rule=Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))"
    - "traefik.http.routers.dashboard.service=api@internal"
    - "traefik.http.routers.dashboard.middlewares=auth"
    - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
    # Dummy service for Swarm port detection. The port can be any valid integer value.
    - "traefik.http.services.dummy-svc.loadbalancer.server.port=9999"
```

```yaml tab="Consul Catalog"
# Dynamic Configuration
- "traefik.http.routers.dashboard.rule=Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))"
- "traefik.http.routers.dashboard.service=api@internal"
- "traefik.http.routers.dashboard.middlewares=auth"
- "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```yaml tab="File (YAML)"
# Dynamic Configuration
http:
  routers:
    dashboard:
      rule: Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))
      service: api@internal
      middlewares:
        - auth
  middlewares:
    auth:
      basicAuth:
        users:
          - "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"
          - "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
```

```toml tab="File (TOML)"
# Dynamic Configuration
[http.routers.my-api]
  rule = "Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))"
  service = "api@internal"
  middlewares = ["auth"]

[http.middlewares.auth.basicAuth]
  users = [
    "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
    "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
  ]
```

## Configuration Options

The API and the dashboard can be configured:

- In the Helm Chart: You can find the options to customize the Traefik installation
enabing the dashboard [here](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/values.yaml#L155).
- In the Traefik Static Configuration as described below.

| Field      | Description  | Default | Required |
|:-----------|:---------------------------------|:--------|:---------|
| `api` | Enable api/dashboard. When set to `true`, its sub option `api.dashboard` is also set to true.| false     | No      |
| `api.dashboard` | Enable dashboard. | false      | No      |
| `api.debug` | Enable additional endpoints for debugging and profiling. | false      | No      |
| `api.disabledashboardad` | Disable the advertisement from the dashboard. | false      | No      |
| `api.insecure` | Enable the API and the dashboard on the entryPoint named traefik.| false      | No      |

## Endpoints

All the following endpoints must be accessed with a `GET` HTTP request.

| Path                           | Description                                                                                 |
|--------------------------------|---------------------------------------------------------------------------------------------|
| `/api/http/routers`            | Lists all the HTTP routers information.                                                     |
| `/api/http/routers/{name}`     | Returns the information of the HTTP router specified by `name`.                             |
| `/api/http/services`           | Lists all the HTTP services information.                                                    |
| `/api/http/services/{name}`    | Returns the information of the HTTP service specified by `name`.                            |
| `/api/http/middlewares`        | Lists all the HTTP middlewares information.                                                 |
| `/api/http/middlewares/{name}` | Returns the information of the HTTP middleware specified by `name`.                         |
| `/api/tcp/routers`             | Lists all the TCP routers information.                                                      |
| `/api/tcp/routers/{name}`      | Returns the information of the TCP router specified by `name`.                              |
| `/api/tcp/services`            | Lists all the TCP services information.                                                     |
| `/api/tcp/services/{name}`     | Returns the information of the TCP service specified by `name`.                             |
| `/api/tcp/middlewares`         | Lists all the TCP middlewares information.                                                  |
| `/api/tcp/middlewares/{name}`  | Returns the information of the TCP middleware specified by `name`.                          |
| `/api/udp/routers`             | Lists all the UDP routers information.                                                      |
| `/api/udp/routers/{name}`      | Returns the information of the UDP router specified by `name`.                              |
| `/api/udp/services`            | Lists all the UDP services information.                                                     |
| `/api/udp/services/{name}`     | Returns the information of the UDP service specified by `name`.                             |
| `/api/entrypoints`             | Lists all the entry points information.                                                     |
| `/api/entrypoints/{name}`      | Returns the information of the entry point specified by `name`.                             |
| `/api/overview`                | Returns statistic information about HTTP, TCP and about enabled features and providers. |
| `/api/rawdata`                 | Returns information about dynamic configurations, errors, status and dependency relations.  |
| `/api/version`                 | Returns information about Traefik version.                                                  |
| `/debug/vars`                  | See the [expvar](https://golang.org/pkg/expvar/) Go documentation.                          |
| `/debug/pprof/`                | See the [pprof Index](https://golang.org/pkg/net/http/pprof/#Index) Go documentation.       |
| `/debug/pprof/cmdline`         | See the [pprof Cmdline](https://golang.org/pkg/net/http/pprof/#Cmdline) Go documentation.   |
| `/debug/pprof/profile`         | See the [pprof Profile](https://golang.org/pkg/net/http/pprof/#Profile) Go documentation.   |
| `/debug/pprof/symbol`          | See the [pprof Symbol](https://golang.org/pkg/net/http/pprof/#Symbol) Go documentation.     |
| `/debug/pprof/trace`           | See the [pprof Trace](https://golang.org/pkg/net/http/pprof/#Trace) Go documentation.       |

## Dashboard

The dashboard is available at the same location as the [API](../../operations/api.md), but by default on the path  `/dashboard/`.

!!! note

    - The trailing slash `/` in `/dashboard/` is mandatory. This limitation can be mitigated using the the [RedirectRegex Middleware](../../middlewares/http/redirectregex.md).
	  - There is also a redirect from the path `/` to `/dashboard/`, but you should not rely on this behavior, as it is subject to change and may complicate routing rules.

To securely access the dashboard, you need to define a routing configuration within Traefik. This involves setting up a router attached to the service `api@internal`, which allows you to:

- Implement security features using [middlewares](../../middlewares/overview.md), such as authentication ([basicAuth](../../middlewares/http/basicauth.md), [digestAuth](../../middlewares/http/digestauth.md),
  [forwardAuth](../../middlewares/http/forwardauth.md)) or [allowlisting](../../middlewares/http/ipallowlist.md).

- Define a [router rule](#dashboard-router-rule) for accessing the dashboard through Traefik.

### Dashboard Router Rule

To ensure proper access to the dashboard, the [router rule](../../routing/routers/index.md#rule) you define must match requests intended for the `/api` and `/dashboard` paths. 
We recommend using either a *Host-based rule* to match all requests on the desired domain or explicitly defining a rule that includes both path prefixes. 
Here are some examples:

```bash tab="Host Rule"
# The dashboard can be accessed on http://traefik.example.com/dashboard/
rule = "Host(`traefik.example.com`)"
```

```bash tab="Path Prefix Rule"
# The dashboard can be accessed on http://example.com/dashboard/ or http://traefik.example.com/dashboard/
rule = "PathPrefix(`/api`) || PathPrefix(`/dashboard`)"
```

```bash tab="Combination of Rules"
# The dashboard can be accessed on http://traefik.example.com/dashboard/
rule = "Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))"
```

{!traefik-for-business-applications.md!}
