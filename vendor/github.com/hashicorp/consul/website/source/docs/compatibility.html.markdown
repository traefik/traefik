---
layout: "docs"
page_title: "Consul Protocol Compatibility Promise"
sidebar_current: "docs-upgrading-compatibility"
description: |-
  We expect Consul to run in large clusters of long-running agents. Because safely upgrading agents in this sort of environment relies heavily on backwards compatibility, we have a strong commitment to keeping different Consul versions protocol-compatible with each other.
---

# Protocol Compatibility Promise

We expect Consul to run in large clusters of long-running agents. Because
safely upgrading agents in this sort of environment relies heavily on backwards
compatibility, we have a strong commitment to keeping different Consul
versions protocol-compatible with each other.

We promise that every subsequent release of Consul will remain backwards
compatible with _at least_ one prior version. Concretely: version 0.5 can
speak to 0.4 (and vice versa) but may not be able to speak to 0.1.

Backwards compatibility is automatic unless otherwise noted. Consul agents by
default will speak the latest protocol but can understand earlier ones.

-> **Note:** If speaking an earlier protocol, _new features may not be available_.

The ability for an agent to speak an earlier protocol is to ensure that any agent
can be upgraded without cluster disruption. Consul agents can be updated one
at a time, one version at a time.

For more details on the specifics of upgrading, see the [upgrading page](/docs/upgrading.html).

## Protocol Compatibility Table

<table class="table table-bordered table-striped">
  <tr>
    <th>Version</th>
    <th>Protocol Compatibility</th>
  </tr>
  <tr>
    <td>0.1 - 0.3</td>
    <td>1</td>
  </tr>
    <td>0.4</td>
    <td>1, 2</td>
  </tr>
  <tr>
    <td>0.5</td>
    <td>1, 2. 0.5.X servers cannot be mixed with older servers.</td>
  </tr>
  <tr>
    <td>0.6</td>
    <td>1, 2, 3</td>
  </tr>
  <tr>
    <td>0.7 - 0.8</td>
    <td>2, 3. Will automatically use protocol > 2 when speaking to compatible agents</td>
  </tr>
</table>

-> **Note:** Raft Protocol is versioned separately, but maintains compatibility with at least one prior version. See [here](https://www.consul.io/docs/upgrade-specific.html#raft-protocol-version-compatibility) for details.

