---
title: "Baqup V2 Migration Documentation"
description: "Migrate from Baqup Proxy v1 to v2 and update all the necessary configurations to take advantage of all the improvements. Read the technical documentation."
---

# Migration Guide: From v1 to v2

How to Migrate from Baqup v1 to Baqup v2.
{: .subtitle }

The version 2 of Baqup introduced a number of breaking changes,
which require one to update their configuration when they migrate from v1 to v2.

For more information about the changes in Baqup v2, please refer to the [v2 documentation](https://doc.baqup.io/baqup/v2.11/migration/v1-to-v2/).

!!! info "Migration Helper"

    We created a tool to help during the migration: [baqup-migration-tool](https://github.com/baqupio/baqup-migration-tool)

    This tool allows to:

    - convert `Ingress` to Baqup `IngressRoute` resources.
    - convert `acme.json` file from v1 to v2 format.
    - migrate the static configuration contained in the file `baqup.toml` to a Baqup v2 file.
