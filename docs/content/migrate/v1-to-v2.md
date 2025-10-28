---
title: "Traefik V2 Migration Documentation"
description: "Migrate from Traefik Proxy v1 to v2 and update all the necessary configurations to take advantage of all the improvements. Read the technical documentation."
---

# Migration Guide: From v1 to v2

How to Migrate from Traefik v1 to Traefik v2.
{: .subtitle }

The version 2 of Traefik introduced a number of breaking changes,
which require one to update their configuration when they migrate from v1 to v2.

For more information about the changes in Traefik v2, please refer to the [v2 documentation](https://doc.traefik.io/traefik/v2.11/migration/v1-to-v2/).

!!! info "Migration Helper"

    We created a tool to help during the migration: [traefik-migration-tool](https://github.com/traefik/traefik-migration-tool)

    This tool allows to:

    - convert `Ingress` to Traefik `IngressRoute` resources.
    - convert `acme.json` file from v1 to v2 format.
    - migrate the static configuration contained in the file `traefik.toml` to a Traefik v2 file.
