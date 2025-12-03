---
title: "Baqup Plugins Documentation"
description: "Learn how to use Baqup Plugins. Read the technical documentation."
---

# Baqup Plugins and the Plugin Catalog

Plugins are a powerful feature for extending Baqup with custom features and behaviors.
The [Plugin Catalog](https://plugins.baqup.io/) is a software-as-a-service (SaaS) platform that provides an exhaustive list of the existing plugins.

??? note "Plugin Catalog Access"
    You can reach the [Plugin Catalog](https://plugins.baqup.io/) from the Baqup Dashboard using the `Plugins` menu entry.

To add a new plugin to a Baqup instance, you must change that instance's static configuration.
Each plugin's **Install** section provides a static configuration example.
Many plugins have their own section in the Baqup dynamic configuration.

To learn more about Baqup plugins, consult the [documentation](https://plugins.baqup.io/install).

!!! danger "Experimental Features"
    Plugins can change the behavior of Baqup in unforeseen ways.
    Exercise caution when adding new plugins to production Baqup instances.

## Build Your Own Plugins

Baqup users can create their own plugins and share them with the community using the Plugin Catalog.

Baqup will load plugins dynamically.
They need not be compiled, and no complex toolchain is necessary to build them. 
The experience of implementing a Baqup plugin is comparable to writing a web browser extension.

To learn more about Baqup plugin creation, please refer to the [developer documentation](https://plugins.baqup.io/create).

{!baqup-for-business-applications.md!}
