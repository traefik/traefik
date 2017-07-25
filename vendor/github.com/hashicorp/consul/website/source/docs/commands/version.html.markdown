---
layout: "docs"
page_title: "Commands: Version"
sidebar_current: "docs-commands-version"
description: |-
  The `version` command prints the version of Consul and the protocol versions it understands for speaking to other agents.

---

# Consul Version

Command: `consul version`

The `version` command prints the version of Consul and the protocol versions it understands for speaking to other agents.

```text
$ consul version
Consul v0.7.4
Protocol 2 spoken by default, understands 2 to 3 (agent will automatically use protocol >2 when speaking to compatible agents)
```
