# The Dashboard

See What's Going On
{: .subtitle }

The dashboard is the central place that shows you the current active routes handled by Traefik. 

<figure>
    <img src="../../assets/img/webui-dashboard.png" alt="Dashboard - Providers" />
    <figcaption>The dashboard in action</figcaption>
</figure>

By default, the dashboard is available on `/` on port `:8080`.

!!! tip "Did You Know?"
    It is possible to customize the dashboard endpoint. 
    To learn how, refer to the [API documentation](./api.md)
    
## Enabling the Dashboard

To enable the dashboard, you need to enable Traefik's API.

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

{!more-on-command-line.md!}

{!more-on-configuration-file.md!}

!!! tip "Did You Know?"
    The API provides more features than the Dashboard. 
    To learn more about it, refer to the [API documentation](./api.md)
