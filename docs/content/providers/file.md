# Traefik & File

Good Old Configuration File
{: .subtitle } 

No longer afraid of using files for your configuration, the file provider is here for you !

## Configuration

### Overview

This provider apply the traefik configuration from a `toml` file.
It could be defined:

* At the end of the global configuration file `traefik.toml`
* In [one dedicated file](#filename-optional)
* In [multiple dedicated files](http://0.0.0.0:8000/providers/file/#directory-optional)

As it is a configuration reference, you will find it everywhere in the traefik documentation.

!!! tip
    As Traefik is multi-providers, you can define one or more middlewares in the file provider and use it in another provider (like docker).

### Configuration Examples

??? example "Configuring File & Deploying / Exposing Services"

    ``` toml
    # Enabling the file provider
    [providers.files]
    
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
    If you're in a hurry, maybe you'd rather go through the [File Reference](../reference/providers/file.md).

### filename (_Optional_)

Defines the path of a dedicated configuration file to use.

```toml
[providers]
  [providers.file]
    filename = "rules.toml"
```

### directory (_Optional_)

Defines the directory path of dedicated configuration files to use.

```toml
[providers]
  [providers.file]
    directory = "/path/to/config"
```

### watch (_Optional_)

Set the `watch` option to `true` to allow Traefik to watch file changes automatically.  
It works both for the `filename` and `directory` options.

```toml
[providers]
  [providers.file]
    filename = "rules.toml"
    watch = true
```

### TOML Templating

!!! warning
    TOML templating can only be used **if rules are defined in one or more separate files**. Templating will not work in the Traefik configuration file.

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
