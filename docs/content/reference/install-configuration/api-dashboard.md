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
| <a id="api" href="#api" title="#api">`api`</a> | Enable api/dashboard. When set to `true`, its sub option `api.dashboard` is also set to true.| false     | No      |
| <a id="api-dashboard" href="#api-dashboard" title="#api-dashboard">`api.dashboard`</a> | Enable dashboard. | false      | No      |
| <a id="api-debug" href="#api-debug" title="#api-debug">`api.debug`</a> | Enable additional endpoints for debugging and profiling. | false      | No      |
| <a id="api-disabledashboardad" href="#api-disabledashboardad" title="#api-disabledashboardad">`api.disabledashboardad`</a> | Disable the advertisement from the dashboard. | false      | No      |
| <a id="api-insecure" href="#api-insecure" title="#api-insecure">`api.insecure`</a> | Enable the API and the dashboard on the entryPoint named traefik.| false      | No      |

## Endpoints

All the following endpoints must be accessed with a `GET` HTTP request.

| Path                           | Description                                                                                 |
|--------------------------------|---------------------------------------------------------------------------------------------|
| <a id="apihttprouters" href="#apihttprouters" title="#apihttprouters">`/api/http/routers`</a> | Lists all the HTTP routers information.                                                     |
| <a id="apihttproutersname" href="#apihttproutersname" title="#apihttproutersname">`/api/http/routers/{name}`</a> | Returns the information of the HTTP router specified by `name`.                             |
| <a id="apihttpservices" href="#apihttpservices" title="#apihttpservices">`/api/http/services`</a> | Lists all the HTTP services information.                                                    |
| <a id="apihttpservicesname" href="#apihttpservicesname" title="#apihttpservicesname">`/api/http/services/{name}`</a> | Returns the information of the HTTP service specified by `name`.                            |
| <a id="apihttpmiddlewares" href="#apihttpmiddlewares" title="#apihttpmiddlewares">`/api/http/middlewares`</a> | Lists all the HTTP middlewares information.                                                 |
| <a id="apihttpmiddlewaresname" href="#apihttpmiddlewaresname" title="#apihttpmiddlewaresname">`/api/http/middlewares/{name}`</a> | Returns the information of the HTTP middleware specified by `name`.                         |
| <a id="apitcprouters" href="#apitcprouters" title="#apitcprouters">`/api/tcp/routers`</a> | Lists all the TCP routers information.                                                      |
| <a id="apitcproutersname" href="#apitcproutersname" title="#apitcproutersname">`/api/tcp/routers/{name}`</a> | Returns the information of the TCP router specified by `name`.                              |
| <a id="apitcpservices" href="#apitcpservices" title="#apitcpservices">`/api/tcp/services`</a> | Lists all the TCP services information.                                                     |
| <a id="apitcpservicesname" href="#apitcpservicesname" title="#apitcpservicesname">`/api/tcp/services/{name}`</a> | Returns the information of the TCP service specified by `name`.                             |
| <a id="apitcpmiddlewares" href="#apitcpmiddlewares" title="#apitcpmiddlewares">`/api/tcp/middlewares`</a> | Lists all the TCP middlewares information.                                                  |
| <a id="apitcpmiddlewaresname" href="#apitcpmiddlewaresname" title="#apitcpmiddlewaresname">`/api/tcp/middlewares/{name}`</a> | Returns the information of the TCP middleware specified by `name`.                          |
| <a id="apiudprouters" href="#apiudprouters" title="#apiudprouters">`/api/udp/routers`</a> | Lists all the UDP routers information.                                                      |
| <a id="apiudproutersname" href="#apiudproutersname" title="#apiudproutersname">`/api/udp/routers/{name}`</a> | Returns the information of the UDP router specified by `name`.                              |
| <a id="apiudpservices" href="#apiudpservices" title="#apiudpservices">`/api/udp/services`</a> | Lists all the UDP services information.                                                     |
| <a id="apiudpservicesname" href="#apiudpservicesname" title="#apiudpservicesname">`/api/udp/services/{name}`</a> | Returns the information of the UDP service specified by `name`.                             |
| <a id="apientrypoints" href="#apientrypoints" title="#apientrypoints">`/api/entrypoints`</a> | Lists all the entry points information.                                                     |
| <a id="apientrypointsname" href="#apientrypointsname" title="#apientrypointsname">`/api/entrypoints/{name}`</a> | Returns the information of the entry point specified by `name`.                             |
| <a id="apioverview" href="#apioverview" title="#apioverview">`/api/overview`</a> | Returns statistic information about HTTP, TCP and about enabled features and providers. |
| <a id="apirawdata" href="#apirawdata" title="#apirawdata">`/api/rawdata`</a> | Returns information about dynamic configurations, errors, status and dependency relations.  |
| <a id="apiversion" href="#apiversion" title="#apiversion">`/api/version`</a> | Returns information about Traefik version.                                                  |
| <a id="debugvars" href="#debugvars" title="#debugvars">`/debug/vars`</a> | See the [expvar](https://golang.org/pkg/expvar/) Go documentation.                          |
| <a id="debugpprof" href="#debugpprof" title="#debugpprof">`/debug/pprof/`</a> | See the [pprof Index](https://golang.org/pkg/net/http/pprof/#Index) Go documentation.       |
| <a id="debugpprofcmdline" href="#debugpprofcmdline" title="#debugpprofcmdline">`/debug/pprof/cmdline`</a> | See the [pprof Cmdline](https://golang.org/pkg/net/http/pprof/#Cmdline) Go documentation.   |
| <a id="debugpprofprofile" href="#debugpprofprofile" title="#debugpprofprofile">`/debug/pprof/profile`</a> | See the [pprof Profile](https://golang.org/pkg/net/http/pprof/#Profile) Go documentation.   |
| <a id="debugpprofsymbol" href="#debugpprofsymbol" title="#debugpprofsymbol">`/debug/pprof/symbol`</a> | See the [pprof Symbol](https://golang.org/pkg/net/http/pprof/#Symbol) Go documentation.     |
| <a id="debugpproftrace" href="#debugpproftrace" title="#debugpproftrace">`/debug/pprof/trace`</a> | See the [pprof Trace](https://golang.org/pkg/net/http/pprof/#Trace) Go documentation.       |

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
