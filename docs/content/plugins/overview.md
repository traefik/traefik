# Plugins and Traefik Pilot

Overview
{: .subtitle}

Traefik Pilot is a software-as-a-service (SaaS) platform that connects to Traefik to extend its capabilities.
It does this through *plugins*, which are dynamically loaded components that enable new features.

For example, Traefik plugins can add features to modify requests or headers, issue redirects, add authentication, and so on, providing similar functionality to Traefik [middlewares](https://doc.traefik.io/traefik/middlewares/overview/).

Traefik Pilot can also monitor connected Traefik instances and issue alerts when one is not responding, or when it is subject to security vulnerabilities.

!!! note "Availability"
    Plugins are available for Traefik v2.3.0-rc1 and later.
    
!!! danger "Experimental Features"
    Plugins can potentially modify the behavior of Traefik in unforeseen ways.
    Exercise caution when adding new plugins to production Traefik instances.

## Connecting to Traefik Pilot

Plugins are available when a Traefik instance is connected to Traefik Pilot.

To register a new instance and begin working with plugins, login or create an account at the [Traefik Pilot homepage](https://pilot.traefik.io) and choose **Register New Instance**.

To complete the connection, Traefik Pilot will issue a token that must be added to your Traefik static configuration by following the instructions provided.

!!! note "Enabling Alerts" 
    Health and security alerts for registered Traefik instances can be enabled from the Preferences in your [Traefik Pilot Profile](https://pilot.traefik.io/profile).

## Creating Plugins

Traefik users can create their own plugins and contribute them to the Traefik Pilot catalog to share them with the community.

Plugins are written in [Go](https://golang.org/) and their code is executed by an [embedded Go interpreter](https://github.com/traefik/yaegi).
There is no need to compile binaries and all plugins are 100% cross-platform.

To learn more and see code for example Traefik plugins, please see the [developer documentation](https://github.com/traefik/plugindemo).
