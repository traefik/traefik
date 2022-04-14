---
title: "Traefik Plugins Documentation"
description: "Learn how to connect Traefik Proxy with Pilot, a SaaS platform that offers features for metrics, alerts, and plugins. Read the technical documentation."
---

# Plugins and Traefik Pilot

Traefik Pilot is a software-as-a-service (SaaS) platform that connects to Traefik to extend its capabilities.
It offers a number of features to enhance observability and control of Traefik through a global control plane and dashboard, including:

* Metrics for network activity of Traefik proxies and groups of proxies
* Alerts for service health issues and security vulnerabilities
* Plugins that extend the functionality of Traefik

!!! important "Learn More About Traefik Pilot"
    This section is intended only as a brief overview for Traefik users who are not familiar with Traefik Pilot. 
    To explore all that Traefik Pilot has to offer, please consult the [Traefik Pilot Documentation](https://doc.traefik.io/traefik-pilot/)

!!! Note "Prerequisites"
    Traefik Pilot is compatible with Traefik Proxy 2.3 or later.

## Connecting to Traefik Pilot

To connect your Traefik proxies to Traefik Pilot, login or create an account at the [Traefik Pilot homepage](https://pilot.traefik.io) and choose **Register New Traefik Instance**.

To complete the connection, Traefik Pilot will issue a token that must be added to your Traefik static configuration, according to the instructions provided by the Traefik Pilot dashboard.
For more information, consult the [Quick Start Guide](https://doc.traefik.io/traefik-pilot/connecting/)

Health and security alerts for registered Traefik instances can be enabled from the Preferences in your [Traefik Pilot Profile](https://pilot.traefik.io/profile).

## Plugins

Plugins are available to any Traefik proxies that are connected to Traefik Pilot.
They are a powerful feature for extending Traefik with custom features and behaviors.

You can browse community-contributed plugins from the catalog in the [Traefik Pilot Dashboard](https://pilot.traefik.io/plugins).

To add a new plugin to a Traefik instance, you must modify that instance's static configuration.
The code to be added is provided for you when you choose **Install the Plugin** from the Traefik Pilot dashboard.
To learn more about Traefik plugins, consult the [documentation](https://doc.traefik.io/traefik-pilot/plugins/overview/).

!!! danger "Experimental Features"
    Plugins can potentially modify the behavior of Traefik in unforeseen ways.
    Exercise caution when adding new plugins to production Traefik instances.

## Build Your Own Plugins

Traefik users can create their own plugins and contribute them to the Traefik Pilot catalog to share them with the community.

Traefik plugins are loaded dynamically. 
They need not be compiled, and no complex toolchain is necessary to build them. 
The experience of implementing a Traefik plugin is comparable to writing a web browser extension.

To learn more and see code for example Traefik plugins, please see the [developer documentation](https://doc.traefik.io/traefik-pilot/plugins/plugin-dev/).
