---
title: "Traefik Dashboard Documentation"
description: "The dashboard shows you the current active routes handled by Traefik Proxy in one central place. Read the technical documentation to learn its operations."
---

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

```yaml tab="File (YAML)"
api: {}
```

```toml tab="File (TOML)"
[api]
```

```bash tab="CLI"
--api=true
```

Then define a routing configuration on Traefik itself,
with a router attached to the service `api@internal` in the
[dynamic configuration](../getting-started/configuration-overview.md#the-dynamic-configuration),
to allow defining:

- One or more security features through [middlewares](../middlewares/overview.md)
  like authentication ([basicAuth](../middlewares/http/basicauth.md), [digestAuth](../middlewares/http/digestauth.md),
  [forwardAuth](../middlewares/http/forwardauth.md)) or [allowlisting](../middlewares/http/ipallowlist.md).

- A [router rule](#dashboard-router-rule) for accessing the dashboard,
  through Traefik itself (sometimes referred to as "Traefik-ception").

### Dashboard Router Rule

As underlined in the [documentation for the `api.dashboard` option](./api.md#dashboard),
the [router rule](../routing/routers/index.md#rule) defined for Traefik must match
the path prefixes `/api` and `/dashboard`.

We recommend using a "Host Based rule" as ```Host(`traefik.example.com`)``` to match everything on the host domain,
or to make sure that the defined rule captures both prefixes:

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

??? example "Dashboard Dynamic Configuration Examples"
    --8<-- "content/operations/include-dashboard-examples.md"

### Custom API Base Path

As shown above, by default Traefik exposes its API and Dashboard under the `/` base path,
which means that respectively the API is served under the `/api` path,
and the dashboard under the `/dashboard` path.

However, it is possible to configure this base path:

```yaml tab="File (YAML)"
api:
  # Customizes the base path:
  # - Serving API under `/traefik/api`
  # - Serving Dashboard under `/traefik/dashboard`
  basePath: /traefik
```

```toml tab="File (TOML)"
[api]
  # Customizes the base path:
  # - Serving API under `/traefik/api`
  # - Serving Dashboard under `/traefik/dashboard`
  basePath = "/traefik"
```

```bash tab="CLI"
# Customizes the base path:
# - Serving API under `/traefik/api`
# - Serving Dashboard under `/traefik/dashboard`
--api.basePath=/traefik
```

??? example "Dashboard Under Custom Path Dynamic Configuration Examples"
    --8<-- "content/operations/include-dashboard-custom-path-examples.md"

## Insecure Mode

!!! warning "Please note that this mode is incompatible with the [custom API base path option](#custom-api-base-path)."

When _insecure_ mode is enabled, one can access the dashboard on the `traefik` port (default: `8080`) of the Traefik instance,
at the following URL: `http://<Traefik IP>:8080/dashboard/` (trailing slash is mandatory).

This mode is **not** recommended because it does not allow security features.
For example, it is not possible to add an authentication middleware with this mode.

It should be used for testing purpose **only**.

To enable the _insecure_ mode, use the following options from [Traefik's API](./api.md#insecure):

```yaml tab="File (YAML)"
api:
  insecure: true
```

```toml tab="File (TOML)"
[api]
  insecure = true
```

```bash tab="CLI"
--api.insecure=true
```

## Disable The Dashboard

By default, the dashboard is enabled when the API is enabled.
If necessary, the dashboard can be disabled by using the following option.

```yaml tab="File (YAML)"
api:
  dashboard: false
```

```toml tab="File (TOML)"
[api]
  dashboard = false
```

```bash tab="CLI"
--api.dashboard=false
```

{!traefik-for-business-applications.md!}
