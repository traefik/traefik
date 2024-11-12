---
title: "Traefik Environment Variables Documentation"
description: "Reference the environment variables for static configuration in Traefik Proxy. Read the technical documentation."
---

# Static Configuration: Environment variables

!!! warning "Environment Variable Casing"

    Traefik normalizes the environment variable key-value pairs by lowercasing them.
    This means that when you interpolate a string in an environment variable's name,
    that string will be treated as lowercase, regardless of its original casing.

    For example, assuming you have set environment variables as follows:

    ```bash
        export TRAEFIK_ENTRYPOINTS_WEB=true
        export TRAEFIK_ENTRYPOINTS_WEB_ADDRESS=:80

        export TRAEFIK_CERTIFICATESRESOLVERS_myResolver=true
        export TRAEFIK_CERTIFICATESRESOLVERS_myResolver_ACME_CASERVER=....
    ```
    
    Although the Entrypoint is named `WEB` and the Certificate Resolver is named `myResolver`, 
    they have to be referenced respectively as `web`, and `myresolver` in the configuration.

--8<-- "content/reference/static-configuration/env-ref.md"
