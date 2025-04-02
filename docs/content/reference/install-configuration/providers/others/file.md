---
title: "Traefik File Documentation"
description: "The file provider in Traefik Proxy lets you define the dynamic configuration in a YAML or TOML file. Read the technical documentation."
---

# File

The file provider lets you define the [install configuration](../overview.md) in a YAML or TOML file.

It supports providing configuration through a single configuration file or multiple separate files.

!!! info

    The file provider is the default format used throughout the documentation to show samples of the configuration for many features.

!!! tip

    The file provider can be a good solution for reusing common elements from other providers (e.g. declaring allowlist middlewares, basic authentication, ...)

## Configuration Example

You can enable the file provider as detailed below:

```yaml tab="File (YAML)"
providers:
  file:
    directory: "/path/to/dynamic/conf"
```

```toml tab="File (TOML)"
[providers.file]
  directory = "/path/to/dynamic/conf"
```

```bash tab="CLI"
--providers.file.directory=/path/to/dynamic/conf
```

Declaring the Routers, Middlewares & Services:

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
      rule: Path(`/foo`)

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

```toml tab="TOML"
[http]
  # Add the router
  [http.routers]
    [http.routers.router0]
      entryPoints = ["web"]
      middlewares = ["my-basic-auth"]
      service = "service-foo"
      rule = "Path(`/foo`)"

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

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.file.filename` | Defines the path to the configuration file.  |  ""    | Yes   |
| `providers.file.directory` | Defines the path to the directory that contains the configuration files. The `filename` and `directory` options are mutually exclusive. It is recommended to use `directory`.  |  ""    | Yes   |
| `providers.file.watch` | Set the `watch` option to `true` to allow Traefik to automatically watch for file changes. It works with both the `filename` and the `directory` options. | true | No |

!!! warning "Limitations"

    With the file provider, Traefik listens for file system notifications to update the dynamic configuration.

    If you use a mounted/bound file system in your orchestrator (like docker or kubernetes), the way the files are linked may be a source of errors.
    If the link between the file systems is broken, when a source file/directory is changed/renamed, nothing will be reported to the linked file/directory, so the file system notifications will be neither triggered nor caught.

    For example, in Docker, if the host file is renamed, the link to the mounted file is broken and the container's file is no longer updated.
    To avoid this kind of issue, it is recommended to:

    * set the Traefik [**directory**](#directory) configuration with the parent directory
    * mount/bind the parent directory

    As it is very difficult to listen to all file system notifications, Traefik uses [fsnotify](https://github.com/fsnotify/fsnotify).
    If using a directory with a mounted directory does not fix your issue, please check your file system compatibility with fsnotify.

{!traefik-for-business-applications.md!}
