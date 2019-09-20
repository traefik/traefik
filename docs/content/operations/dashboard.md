# The Dashboard

See What's Going On
{: .subtitle }

The dashboard is the central place that shows you the current active routes handled by Traefik. 

<figure>
    <img src="../../assets/img/webui-dashboard.png" alt="Dashboard - Providers" />
    <figcaption>The dashboard in action</figcaption>
</figure>

By default, the dashboard is available on `/dashboard` on port `:8080`.
There is also a redirect of `/` to `/dashboard`, but one should not rely on that property as it is bound to change,
and it might make for confusing routing rules anyway.

!!! note "Did You Know?"
    It is possible to customize the dashboard endpoint. 
    To learn how, refer to the [API documentation](./api.md)
    
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

!!! important "API/Dashboard Security" 
    
    To secure your dashboard, the use of a `service` named `api@internal` is mandatory and requires the definition of a router using one or more security [middlewares](../middlewares/overview.md)
    like authentication ([basicAuth](../middlewares/basicauth.md) , [digestAuth](../middlewares/digestauth.md), [forwardAuth](../middlewares/forwardauth.md)) or [whitelisting](../middlewares/ipwhitelist.md).  
    More information about `api@internal` can be found in the [API documentation](./api.md#configuration)

!!! note "Did You Know?"
    The API provides more features than the Dashboard. 
    To learn more about it, refer to the [API documentation](./api.md)
