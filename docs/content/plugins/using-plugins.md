# Using Plugins

Plugins are available to any instance of Traefik v2.3 or later that is [registered](overview.md#connecting-to-traefik-pilot) with Traefik Pilot.
Plugins are hosted on GitHub, but you can browse plugins to add to your registered Traefik instances from the Traefik Pilot UI.

!!! danger "Experimental Features"
    Plugins can potentially modify the behavior of Traefik in unforeseen ways.
    Exercise caution when adding new plugins to production Traefik instances.

## Add a Plugin

To add a new plugin to a Traefik instance, you must modify that instance's static configuration.
The code to be added is provided by the Traefik Pilot UI when you choose **Install the Plugin**.

In the example below, we add the [`blockpath`](http://github.com/traefik/plugin-blockpath) and [`rewritebody`](https://github.com/traefik/plugin-rewritebody) plugins:

```toml tab="File (TOML)"
[entryPoints]
  [entryPoints.web]
    address = ":80"

[pilot]
  token = "xxxxxxxxx"

[experimental.plugins]
  [experimental.plugins.block]
    modulename = "github.com/traefik/plugin-blockpath"
    version = "v0.2.0"
    
  [experimental.plugins.rewrite]
    modulename = "github.com/traefik/plugin-rewritebody"
    version = "v0.3.0"
```

```yaml tab="File (YAML)"
entryPoints:
  web:
    address: :80

pilot:
    token: xxxxxxxxx

experimental:
  plugins:
    block:
      modulename: github.com/traefik/plugin-blockpath
      version: v0.2.0
    rewrite:
      modulename: github.com/traefik/plugin-rewritebody
      version: v0.3.0
```

```bash tab="CLI"
--entryPoints.web.address=:80
--pilot.token=xxxxxxxxx
--experimental.plugins.block.modulename=github.com/traefik/plugin-blockpath
--experimental.plugins.block.version=v0.2.0
--experimental.plugins.rewrite.modulename=github.com/traefik/plugin-rewritebody
--experimental.plugins.rewrite.version=v0.3.0
```

## Configuring Plugins

Some plugins will need to be configured by adding a dynamic configuration.
For the `bodyrewrite` plugin, for example:

```yaml tab="Docker"
labels:
  - "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].regex=example"
  - "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].replacement=test"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: my-rewritebody
spec:
  plugin:
    rewrite:
      rewrites:
        - regex: example
          replacement: test
```

```yaml tab="Consul Catalog"
- "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].regex=example"
- "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].replacement=test"
```

```json tab="Marathon"
"labels": {
  "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].regex": "example",
  "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].replacement": "test"
}
```

```yaml tab="Rancher"
labels:
  - "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].regex=example"
  - "traefik.http.middlewares.my-rewritebody.plugin.rewrite.rewrites[0].replacement=test"
```

```toml tab="File (TOML)"
[http.middlewares]
  [http.middlewares.my-rewritebody.plugin.rewrite]
    lastModified = true
    [[http.middlewares.my-rewritebody.plugin.rewrite.rewrites]]
      regex = "example"
      replacement = "test"
```

```yaml tab="File (YAML)"
http:
  middlewares:
    my-rewritebody:
      plugin:
        rewrite:
          rewrites:
            - regex: example
              replacement: test
```
