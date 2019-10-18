# The Dashboard

See What's Going On
{: .subtitle }

The dashboard is the central place that shows you the current active routes handled by Traefik.

<figure>
    <img src="../../assets/img/webui-dashboard.png" alt="Dashboard - Providers" />
    <figcaption>The dashboard in action</figcaption>
</figure>

The dashboard is available at the same location as the [API](./api.md) but on the path `/dashboard/` by default.

!!! warning "The trailing slash `/` in `/dashboard/` is mandatory"

There are 2 ways to configure and access the dashboard:

- [Secure mode (Recommended)](#secure-mode)
- [Insecure mode](#insecure-mode)

!!! note ""
    There is also a redirect of the path `/` to the path `/dashboard/`,
    but one should not rely on that property as it is bound to change,
    and it might make for confusing routing rules anyway.

## Secure Mode

This is the **recommended** method.

Start by enabling the dashboard by using the following option from [Traefik's API](./api.md)
on the [static configuration](../getting-started/configuration-overview.md#the-static-configuration):

```toml tab="File (TOML)"
[api]
  # Dashboard
  #
  # Optional
  # Default: true
  #
  dashboard = true
```

```yaml tab="File (YAML)"
api:
  # Dashboard
  #
  # Optional
  # Default: true
  #
  dashboard: true
```

```bash tab="CLI"
# Dashboard
#
# Optional
# Default: true
#
--api.dashboard=true
```

Then define a routing configuration on Traefik itself, 
with a router attached to the service `api@internal` in the 
[dynamic configuration](../getting-started/configuration-overview.md#the-dynamic-configuration),
to allow defining:

- One or more security features through [middlewares](../middlewares/overview.md)
  like authentication ([basicAuth](../middlewares/basicauth.md) , [digestAuth](../middlewares/digestauth.md),
  [forwardAuth](../middlewares/forwardauth.md)) or [whitelisting](../middlewares/ipwhitelist.md).

- A [router rule](#dashboard-router-rule) for accessing the dashboard,
  through Traefik itself (sometimes referred as "Traefik-ception").

??? example "Dashboard Dynamic Configuration Examples"

    ```yaml tab="Docker"
    # Dynamic Configuration
    labels:
      - "traefik.http.routers.api.rule=Host(`traefik.domain.com`)
      - "traefik.http.routers.api.service=api@internal"
      - "traefik.http.routers.api.middlewares=auth"
      - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
    ```

    ```json tab="Marathon"
    "labels": {
      "traefik.http.routers.api.rule": "Host(`traefik.domain.com`)",
      "traefik.http.routers.api.service": "api@internal",
      "traefik.http.routers.api.middlewares": "auth",
      "traefik.http.middlewares.auth.basicauth.users": "test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
    }
    ```

    ```yaml tab="Rancher"
    # Dynamic Configuration
    labels:
      - "traefik.http.routers.api.rule=Host(`traefik.domain.com`)
      - "traefik.http.routers.api.service=api@internal"
      - "traefik.http.routers.api.middlewares=auth"
      - "traefik.http.middlewares.auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
    ```

    ```toml tab="File (TOML)"
    # Dynamic Configuration
    [http.routers.my-api]
        rule="Host(`traefik.domain.com`)
        service="api@internal"
        middlewares=["auth"]

    [http.middlewares.auth.basicAuth]
        users = [
          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
          "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
        ]
    ```

    ```yaml tab="File (YAML)"
    # Dynamic Configuration
    http:
      routers:
        api:
          rule: Host(`traefik.domain.com`)
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

### Dashboard Router Rule

As underlined in the [documentation for the `api.dashboard` option](./api.md#dashboard),
the [router rule](../routing/routers/index.md#rule) defined for Traefik must matches 
the path prefixes `/api` and `/dashboard`.

We recommend to use a "Host Based rule" as ```Host(`traefik.domain.com`)``` to match everything on the host domain,
or to make sure that the defined rule captures both prefixes:

```bash tab="Host Rule"
# Matches http://traefik.domain.com/api or http://traefik.domain.com/dashboard
rule = "Host(`traefik.domain.com`)"
```

```bash tab="Path Prefix Rule"
# Matches http://traefik.domain.com/api , http://domain.com/api or http://traefik.domain.com/dashboard
# but does not match http://traefik.domain.com/hello
rule = "PathPrefix(`/api`) || PathPrefix(`/dashboard`)"
```

```bash tab="Combination of Rules"
# Matches http://traefik.domain.com/api or http://traefik.domain.com/dashboard
# but does not match http://traefik.domain.com/hello
rule = "Host(`traefik.domain.com`) && (PathPrefix(`/api`) || PathPrefix(`/dashboard`))"
```

## Insecure Mode

This mode is not recommended because it does not allow the use of security features.

To enable the "insecure mode", use the following options from [Traefik's API](./api.md#insecure):

```toml tab="File (TOML)"
[api]
  dashboard = true
  insecure = true
```

```yaml tab="File (YAML)"
api:
  dashboard: true
  insecure: true
```

```bash tab="CLI"
--api.dashboard=true --api.insecure=true
```

You can now access the dashboard on the port `8080` of the Traefik instance,
at the following URL: `http://<Traefik IP>:8080/dashboard/` (trailing slash is mandatory).
