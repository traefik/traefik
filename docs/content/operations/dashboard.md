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

- Insecure mode: by using the [insecure mode of the API](./api.md#insecure),
  then you can use the port `:8080` of Traefik and reach the dashboard at the URL `http://<Traefik IP>:8080/dashboard/`.

!!! tip "Changing the default port `8080` for insecured mode"
    By default, the insecure mode uses the default entrypoint named
    It is possible to customize the dashboard endpoint.
    To learn how, refer to the [API documentation](./api.md)

- Secured mode: by defining a router for Traefik's dashboard, associated to the service `api@internal`,
  then you can access the router through Traefik itself with your own routing rule.
  Read more on this on the section ["API/Dashboard Security"](#apidashboard-security).

!!! note ""
    There is also a redirect of the path `/` to the path `/dashboard/`,
    but one should not rely on that property as it is bound to change,
    and it might make for confusing routing rules anyway.

## Enabling the Dashboard

To enable the dashboard, you need to enable [Traefik's API](./api.md).

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

## API/Dashboard Security

To secure your dashboard, the use of a `service` named `api@internal` is mandatory and requires the definition of a router using one or more security [middlewares](../middlewares/overview.md)
like authentication ([basicAuth](../middlewares/basicauth.md) , [digestAuth](../middlewares/digestauth.md), [forwardAuth](../middlewares/forwardauth.md)) or [whitelisting](../middlewares/ipwhitelist.md).
More information about `api@internal` can be found in the [API documentation](./api.md#configuration)

!!! info "Did You Know?"
    The API provides more features than the Dashboard.
    To learn more about it, refer to the [API documentation](./api.md)
