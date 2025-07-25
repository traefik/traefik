---
title: "Traefik File Dynamic Configuration"
description: "This guide will provide you with the YAML and TOML files for dynamic configuration in Traefik Proxy. Read the technical documentation."
---


# Traefik and Configuration Files

!!! warning "Work In Progress"

    This page is still work in progress to provide a better documention of the routing options.

    It has been created to provide a centralized page with all the option in YAML and TOML format.

## Configuration Options

```yml  tab="YAML"
--8<-- "content/reference/routing-configuration/other-providers/file.yaml"
```

```toml  tab="TOML"
--8<-- "content/reference/routing-configuration/other-providers/file.toml"
```

## Go Templating

!!! warning

    Go Templating only works with dedicated dynamic configuration files.
    Templating does not work in the Traefik main static configuration file.

Traefik supports using Go templating to automatically generate repetitive sections of configuration files.
These sections must be a valid [Go template](https://pkg.go.dev/text/template/), and can use
[sprig template functions](https://masterminds.github.io/sprig/).

To illustrate, it is possible to easily define multiple routers, services, and TLS certificates as described in the following examples:

??? example "Configuring Using Templating"

    ```yaml tab="YAML"
    http:
      routers:
        {{range $i, $e := until 100 }}
        router{{ $e }}-{{ env "MY_ENV_VAR" }}:
          # ...
        {{end}}

      services:
        {{range $i, $e := until 100 }}
        application{{ $e }}:
          # ...
        {{end}}

    tcp:
      routers:
        {{range $i, $e := until 100 }}
        router{{ $e }}:
          # ...
        {{end}}

      services:
        {{range $i, $e := until 100 }}
        service{{ $e }}:
          # ...
        {{end}}

    tls:
      certificates:
      {{ range $i, $e := until 10 }}
      - certFile: "/etc/traefik/cert-{{ $e }}.pem"
        keyFile: "/etc/traefik/cert-{{ $e }}.key"
        store:
        - "my-store-foo-{{ $e }}"
        - "my-store-bar-{{ $e }}"
      {{end}}
    ```

    ```toml tab="TOML"
    # template-rules.toml
    [http]

      [http.routers]
      {{ range $i, $e := until 100 }}
        [http.routers.router{{ $e }}-{{ env "MY_ENV_VAR" }}]
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
      stores = ["my-store-foo-{{ $e }}", "my-store-bar-{{ $e }}"]
    {{ end }}

    [tls.config]
    {{ range $i, $e := until 10 }}
      [tls.config.TLS{{ $e }}]
      # ...
    {{ end }}
    ```
