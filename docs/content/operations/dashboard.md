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

Start by enabling the dashboard by using the following option from [Traefik's API](./api.md):

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

Then specify a router associated to the service `api@internal` to allow:

- Defining one or more security features through [middlewares](../middlewares/overview.md)
  like authentication ([basicAuth](../middlewares/basicauth.md) , [digestAuth](../middlewares/digestauth.md),
  [forwardAuth](../middlewares/forwardauth.md)) or [whitelisting](../middlewares/ipwhitelist.md).

- Defining your own [HTTP routing rule](../../routing/routers/#rule) for accessing the dashboard,
  through Traefik itself (sometimes referred as "Traefik-ception").

Please visit the ["Configuration" section of the API documentation](./api.md#configuration)
to learn about configuring a router with the service `api@internal` and enabling the security features.

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
