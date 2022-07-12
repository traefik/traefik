---
title: "Traefik Plugins Documentation"
description: "Learn how to use Traefik Plugins. Read the technical documentation."
---

# Traefik Plugins and the Plugin Catalog

Plugins are a powerful feature for extending Traefik with custom features and behaviors.
The [Plugin Catalog](https://plugins.traefik.io/) is a software-as-a-service (SaaS) platform that provides an exhaustive list of the existing plugins.

??? note "Plugin Catalog Access"
    You can reach the [Plugin Catalog](https://plugins.traefik.io/) from the Traefik Dashboard using the `Plugins` menu entry.

To add a new plugin to a Traefik instance, you must change that instance's static configuration.
Each plugin's **Install** section provides a static configuration example.
Many plugins have their own section in the Traefik dynamic configuration.

To learn more about Traefik plugins, consult the [documentation](https://plugins.traefik.io/install).

!!! danger "Experimental Features"
    Plugins can change the behavior of Traefik in unforeseen ways.
    Exercise caution when adding new plugins to production Traefik instances.

## Build Your Own Plugins

Traefik users can create their own plugins and share them with the community using the Plugin Catalog.

Traefik will load plugins dynamically.
They need not be compiled, and no complex toolchain is necessary to build them. 
The experience of implementing a Traefik plugin is comparable to writing a web browser extension.

To learn more about Traefik plugin creation, please refer to the [developer documentation](https://plugins.traefik.io/create).
