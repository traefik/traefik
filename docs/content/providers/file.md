# Traefik & File

Good Old Configuration File
{: .subtitle } 

The file provider lets you define the [dynamic configuration](./overview.md) in a `toml` file.
You can write these configuration elements:

* At the end of the main Traefik configuration file (by default: `traefik.toml`).
* In [a dedicated file](#filename-optional)
* In [several dedicated files](#directory-optional)

!!! note
    The file provider is the default format used throughout the documentation to show samples of the configuration for many features. 

!!! tip
    The file provider can be a good location for common elements you'd like to re-use from other providers; e.g. declaring whitelist middlewares, basic authentication, ...

## Configuration Examples

??? example "Declaring Routers, Middlewares & Services"

    ``` toml
    # Enabling the file provider
    [providers.file]
    
    [http]
      # Add the router
      [http.routers]
        [http.routers.router0]
          entrypoints = ["web"]
          middlewares = ["my-basic-auth"]
          service = "service-foo"
          rule = "Path(`foo`)"
    
        # Add the middleware
        [http.middlewares]    
          [http.middlewares.my-basic-auth.BasicAuth]
            users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", 
                      "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
            usersFile = "etc/traefik/.htpasswd"
        
        # Add the service
        [http.services]
          [http.services.service-foo]
            [http.services.service-foo.LoadBalancer]
              method = "wrr"
              [[http.services.service-foo.LoadBalancer.Servers]]
                url = "http://foo/"
                weight = 30
              [[http.services.service-foo.LoadBalancer.Servers]]
                url = "http://bar/"
                weight = 70
    ```

## Provider Configuration Options

!!! tip "Browse the Reference"
    If you're in a hurry, maybe you'd rather go through the [static](../reference/static-configuration.md) and the [dynamic](../reference/providers/file.md) configuration references.
    
### filename (_Optional_)

Defines the path of the configuration file.

```toml
[providers]
  [providers.file]
    filename = "rules.toml"
```

### directory (_Optional_)

Defines the directory that contains the configuration files.

```toml
[providers]
  [providers.file]
    directory = "/path/to/config"
```

### watch (_Optional_)

Set the `watch` option to `true` to allow Traefik to automatically watch for file changes.  
It works with both the `filename` and the `directory` options.

```toml
[providers]
  [providers.file]
    filename = "rules.toml"
    watch = true
```

### TOML Templating

!!! warning
    TOML templating only works along with dedicated configuration files. Templating does not work in the Traefik main configuration file.

Traefik allows using TOML templating.  
Thus, it's possible to define easily lot of routers, services and TLS certificates as described in the file `template-rules.toml` :

??? example "Configuring Using Templating"

    ```toml
    # template-rules.toml
    [http]
    
      [http.routers]
      {{ range $i, $e := until 100 }}
        [http.routers.router{{ $e }}]
        # ...
      {{ end }}  
      
      
      [http.Services]
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
      
      
      [tcp.Services]
      {{ range $i, $e := until 100 }}
          [http.services.service{{ $e }}]
          # ...
      {{ end }}  
    
    {{ range $i, $e := until 10 }}
    [[TLS]]
      Store = ["my-store-foo-{{ $e }}", "my-store-bar-{{ $e }}"]
      [TLS.Certificate]
        CertFile = "/etc/traefik/cert-{{ $e }}.pem"
        KeyFile = "/etc/traefik/cert-{{ $e }}.key"
    {{ end }}
    
    [TLSConfig]
    {{ range $i, $e := until 10 }}
      [TLSConfig.TLS{{ $e }}]
      # ...
    {{ end }}
    
    ```
