# Traefik & File

Good Old Configuration File
{: .subtitle } 

The file provider lets you define the [dynamic configuration](./overview.md) in a TOML or YAML file.
You can write these configuration elements:

* In [a dedicated file](#filename)
* In [several dedicated files](#directory)

!!! note
    The file provider is the default format used throughout the documentation to show samples of the configuration for many features. 

!!! tip
    The file provider can be a good location for common elements you'd like to re-use from other providers; e.g. declaring whitelist middlewares, basic authentication, ...

## Configuration Examples

??? example "Declaring Routers, Middlewares & Services"

    Enabling the file provider:
    
    ```toml tab="File (TOML)"
    [providers.file]
      filename = "/my/path/to/dynamic-conf.toml"
    ```
    
    ```yaml tab="File (YAML)"
    providers:
      file:
        filename: "/my/path/to/dynamic-conf.yml"
    ```
    
    ```bash tab="CLI"
    --providers.file.filename=/my/path/to/dynamic_conf.toml
    ```
    
    Declaring Routers, Middlewares & Services:
    
    ```toml tab="TOML"
    [http]
      # Add the router
      [http.routers]
        [http.routers.router0]
          entryPoints = ["web"]
          middlewares = ["my-basic-auth"]
          service = "service-foo"
          rule = "Path(`foo`)"
    
        # Add the middleware
        [http.middlewares]    
          [http.middlewares.my-basic-auth.basicAuth]
            users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", 
                      "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
            usersFile = "etc/traefik/.htpasswd"
        
        # Add the service
        [http.services]
          [http.services.service-foo]
            [http.services.service-foo.loadBalancer]
              [[http.services.service-foo.loadBalancer.servers]]
                url = "http://foo/"
              [[http.services.service-foo.loadBalancer.servers]]
                url = "http://bar/"
    ```
    
    ```yaml tab="YAML"
    http:
      # Add the router
      routers:
        router0:
          entryPoints:
          - web
          middlewares:
          - my-basic-auth
          service: service-foo
          rule: Path(`foo`)
      
      # Add the middleware
      middlewares:
        my-basic-auth:
          basicAuth:
            users:
            - test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/
            - test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0
            usersFile: etc/traefik/.htpasswd
      
      # Add the service
      services:
        service-foo:
          loadBalancer:
            servers:
            - url: http://foo/
            - url: http://bar/
            passHostHeader: false
    ```

## Provider Configuration Options

!!! tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the [static](../reference/static-configuration/overview.md) and the [dynamic](../reference/dynamic-configuration/file.md) configuration references.
    
### `filename`

_Optional_

Defines the path of the configuration file.

```toml tab="File (TOML)"
[providers]
  [providers.file]
    filename = "dynamic_conf.toml"
```

```yaml tab="File (YAML)"
providers:
  file:
    filename: dynamic_conf.yml
```

```bash tab="CLI"
--providers.file.filename=dynamic_conf.toml
```

### `directory`

_Optional_

Defines the directory that contains the configuration files.

```toml tab="File (TOML)"
[providers]
  [providers.file]
    directory = "/path/to/config"
```

```yaml tab="File (YAML)"
providers:
  file:
    directory: /path/to/config
```

```bash tab="CLI"
--providers.file.directory=/path/to/config
```

### `watch`

_Optional_

Set the `watch` option to `true` to allow Traefik to automatically watch for file changes.  
It works with both the `filename` and the `directory` options.

```toml tab="File (TOML)"
[providers]
  [providers.file]
    filename = "dynamic_conf.toml"
    watch = true
```

```yaml tab="File (YAML)"
providers:
  file:
    filename: dynamic_conf.yml
    watch: true
```

```bash tab="CLI"
--providers.file.filename=dynamic_conf.toml
--providers.file.watch=true
```

### Go Templating

!!! warning
    Go Templating only works along with dedicated configuration files.
    Templating does not work in the Traefik main configuration file.

Traefik allows using Go templating.  
Thus, it's possible to define easily lot of routers, services and TLS certificates as described in the file `template-rules.toml` :

??? example "Configuring Using Templating"
    
    ```toml tab="TOML"
    # template-rules.toml
    [http]
    
      [http.routers]
      {{ range $i, $e := until 100 }}
        [http.routers.router{{ $e }}]
        # ...
      {{ end }}  
      
      
      [http.services]
      {{ range $i, $e := until 100 }}
          [http.services.service{{ $e }}]
          # ...
      {{ end }}  
      
    [tcp]
    
      [tcp.routers]
      {{ range $i, $e := until 100 }}
        [tcp.routers.router{{ $e }}]
        # ...
      {{ end }}  
      
      
      [tcp.services]
      {{ range $i, $e := until 100 }}
          [http.services.service{{ $e }}]
          # ...
      {{ end }}  
    
    {{ range $i, $e := until 10 }}
    [[tls.certificates]]
      certFile = "/etc/traefik/cert-{{ $e }}.pem"
      keyFile = "/etc/traefik/cert-{{ $e }}.key"
      store = ["my-store-foo-{{ $e }}", "my-store-bar-{{ $e }}"]
    {{ end }}
    
    [tls.config]
    {{ range $i, $e := until 10 }}
      [tls.config.TLS{{ $e }}]
      # ...
    {{ end }}
    ```
    
    ```yaml tab="YAML"
    http:
    
    {{range $i, $e := until 100 }}
      routers:
        router{{ $e }:
          # ...
    {{end}}
    
    {{range $i, $e := until 100 }}
      services:
        application{{ $e }}:
          # ...
    {{end}}
    
    tcp:
    
    {{range $i, $e := until 100 }}
      routers:
        router{{ $e }:
          # ...
    {{end}}
    
    {{range $i, $e := until 100 }}
      services:
        service{{ $e }}:
          # ...
    {{end}}
    
    {{ range $i, $e := until 10 }}
    tls:
      certificates:
      - certFile: "/etc/traefik/cert-{{ $e }}.pem"
        keyFile: "/etc/traefik/cert-{{ $e }}.key"
        store:
        - "my-store-foo-{{ $e }}"
        - "my-store-bar-{{ $e }}"
    {{end}}
    ```
