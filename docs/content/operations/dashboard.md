# The Dashboard

See What's Going On
{: .subtitle }

The dashboard is the central place that shows you the current active routes handled by Traefik. 

!!! warning "Dashboard WIP"
    Currently, the dashboard is in a Work In Progress State while being reconstructed for v2. 
    Therefore, the dashboard is currently not working.

<figure>
    <img src="../../assets/img/dashboard-main.png" alt="Dashboard - Providers" />
    <figcaption>The dashboard in action with Traefik listening to 3 different providers</figcaption>
</figure>

<figure>
    <img src="../../assets/img/dashboard-health.png" alt="Dashboard - Health" />
    <figcaption>The dashboard shows the health of the system.</figcaption>
</figure>

By default, the dashboard is available on `/` on port `:8080`.

!!! tip "Did You Know?"
    It is possible to customize the dashboard endpoint. 
    To learn how, refer to the `Traefik's API documentation`(TODO: add doc and link).
    
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
--api.dashboard
```

{!more-on-command-line.md!}

{!more-on-configuration-file.md!}

!!! tip "Did You Know?"
    The API provides more features than the Dashboard. 
    To learn more about it, refer to the `Traefik's API documentation`(TODO: add doc and link).
