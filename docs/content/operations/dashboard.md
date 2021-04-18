# The Dashboard

See What's Going On
{: .subtitle }

The dashboard is the central place that shows you the current active routes handled by Traefik.

<figure>
    <img src="../../assets/img/webui-dashboard.png" alt="Dashboard - Providers" />
    <figcaption>The dashboard in action</figcaption>
</figure>

The dashboard is available at a special internal `service` named `dashboard@internal`.

!!! warning "In older version, the dashboard is avaliable at the `api@internal` service with `/dashboard/` path prefix. (The trailing slash `/` in `/dashboard/` is mandatory)"

There are 2 ways to configure and access the dashboard:

- [Secure mode (Recommended)](#secure-mode)
- [Insecure mode](#insecure-mode)

!!! note ""
    There is no need a redirect of the path `/` to the path `/dashboard/`,
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
with a router attached to the two services `api@internal` and `dashboard@internal` in the
[dynamic configuration](../getting-started/configuration-overview.md#the-dynamic-configuration),
to allow defining:

- One or more security features through [middlewares](../middlewares/overview.md)
  like authentication ([basicAuth](../middlewares/basicauth.md) , [digestAuth](../middlewares/digestauth.md),
  [forwardAuth](../middlewares/forwardauth.md)) or [whitelisting](../middlewares/ipwhitelist.md).

- A [router rule](#dashboard-router-rule) for accessing the dashboard,
  through Traefik itself (sometimes referred as "Traefik-ception").

### Dashboard Router Rule

The `api@internal` service should match the path prefix `/api`, and `dashboard@internal` should match the path prefix `/`.
If the dashboard is served at `http://traefik.example.com/foo/bar/`, you must use middleware to strip the `/foo/bar` in the path. 


We recommend to use a "Host Based rule" as ```Host(`traefik.example.com`)``` to match everything on the host domain,
or to make sure that the defined rule captures both prefixes:

```bash tab="Host Rule"
# The dashboard can be accessed on http://traefik.example.com/
rule = "Host(`traefik.example.com`)"
```

```bash tab="Path Prefix Rule"
# The dashboard can be accessed on http://example.com/ or http://traefik.example.com/
rule = "PathPrefix(`/api`) || PathPrefix(`/`)"
```

```bash tab="Combination of Rules"
# The dashboard can be accessed on http://traefik.example.com/dashboard/
rule = "Host(`traefik.example.com`) && (PathPrefix(`/api`) || PathPrefix(`/`))"
```

??? example "Dashboard Dynamic Configuration Examples"
    --8<-- "content/operations/include-dashboard-examples.md"

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
