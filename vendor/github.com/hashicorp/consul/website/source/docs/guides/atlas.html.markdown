---
layout: "docs"
page_title: "Atlas Integration"
sidebar_current: "docs-guides-atlas"
description: |-
  This guide covers how to integrate Atlas with Consul to provide features like an infrastructure dashboard and automatic cluster joining.
---

# Atlas Integration

~> **Notice:** The hosted version of Consul Enterprise will be deprecated on
March 7th, 2017. For details, see https://atlas.hashicorp.com/help/consul/alternatives

[Atlas](https://atlas.hashicorp.com?utm_source=oss&utm_medium=guide-atlas&utm_campaign=consul) is a service provided by HashiCorp to deploy applications and manage infrastructure.
Starting with Consul 0.5, it is possible to integrate Consul with Atlas. Atlas is able to display the state of the Consul cluster in its dashboard and set up alerts based on health checks. Additionally, nodes can use Atlas to auto-join a Consul cluster without hardcoding any configuration.

Atlas is able to securely retrieve data from nodes as Consul maintains a long-running connection to the
[SCADA](http://scada.hashicorp.com) service.

## Enabling Atlas Integration

To enable Atlas integration, you must specify the name of the Atlas infrastructure and the Atlas authentication
token in your Consul configuration. The Atlas infrastructure name can be set either with the [`-atlas` CLI flag](/docs/agent/options.html#_atlas) or with the [`atlas_infrastructure` configuration option](/docs/agent/options.html#atlas_infrastructure). The Atlas token is set with the [`-atlas-token` CLI flag](/docs/agent/options.html#_atlas_token),
[`atlas_token` configuration option](/docs/agent/options.html#atlas_token), or `ATLAS_TOKEN` environment variable.

To get an Atlas username and token, [create an account here](https://atlas.hashicorp.com/account/new?utm_source=oss&utm_medium=guide-atlas&utm_campaign=consul) and replace the respective values in your Consul configuration with your credentials.

To verify the integration, either run the agent with `debug`-level logging or use `consul monitor -log-level=debug`
and look for a line like:

    [DEBUG] scada-client: assigned session '406ca55d-1801-f964-2942-45f5f9df3995'

This shows that the Consul agent was successfully able to register with the SCADA service.

## Using Auto-Join

Once integrated with Atlas, the auto-join feature can be used to have nodes automatically join other
peers in their datacenter. Server nodes will automatically join peer LAN nodes and other WAN nodes.
Client nodes will only join other LAN nodes in their datacenter.

Auto-join is enabled with the [`-atlas-join` CLI flag](/docs/agent/options.html#_atlas_join) or the
[`atlas_join` configuration option](/docs/agent/options.html#atlas_join).

## Securing Atlas

The connection to Atlas does not have elevated privileges. API requests made by Atlas
are served in the same way any other HTTP request is handled. If ACLs are enabled, it is possible, via
the [`atlas_acl_token`](/docs/agent/options.html#atlas_acl_token) configuration option, to force an
Atlas ACL token to be used instead of the agent's default token.

The resolution order for ACL tokens is:

1. Request-specific token provided by `?token=`. These tokens are set in the Atlas UI.
2. The agent's [`atlas_acl_token`](/docs/agent/options.html#atlas_acl_token), if configured.
3. The agent's [`acl_token`](/docs/agent/options.html#acl_token), if configured.
4. The `anonymous` token.

Because the [`acl_token`](/docs/agent/options.html#acl_token) typically has elevated permissions
compared to the `anonymous` token, the [`atlas_acl_token`](/docs/agent/options.html#atlas_acl_token)
can be set to `anonymous` to drop privileges that would otherwise be inherited from the agent.
