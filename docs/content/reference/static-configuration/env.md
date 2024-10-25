---
title: "Traefik Environment Variables Documentation"
description: "Reference the environment variables for static configuration in Traefik Proxy. Read the technical documentation."
---

# Static Configuration: Environment variables

!!! caution "Environment Variable Naming"

    Traefik normalizes the environment variable key-value pairs by lowercasing them.
    This means that when you interpolate a string in an environment variable's name,
    that string will be treated as lowercase, regardless of its original casing.

    For example, assuming you have set environment variables as follows:

    ```bash
        export TRAEFIK_ENTRYPOINTS_web=true
        export TRAEFIK_ENTRYPOINTS_web_ADDRESS=:80

        export TRAEFIK_CERTIFICATESRESOLVERS_myResolver=true
        export TRAEFIK_CERTIFICATESRESOLVERS_myResolver_ACME_CASERVER=....
    ```
    Although the Certificate Resolver is named `myResolver`, referencing it as follows will not work:

    !!! failure "Wrong Usage"
        ```bash
            export TRAEFIK_ENTRYPOINTS_web_HTTP_TLS_CERTRESOLVER=myResolver
        ```
    You must specify the Certificate Resolver's name as a lowercase string:

    !!! success "Correct Usage"
        ```bash
            export TRAEFIK_ENTRYPOINTS_web_HTTP_TLS_CERTRESOLVER=myresolver
        ```

### Below are the supported Environment Variables

--8<-- "content/reference/static-configuration/env-ref.md"
