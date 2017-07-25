---
layout: "docs"
page_title: "Commands: Keygen"
sidebar_current: "docs-commands-keygen"
description: |-
  The `keygen` command generates an encryption key that can be used for Consul agent traffic encryption. The keygen command uses a cryptographically strong pseudo-random number generator to generate the key.

---

# Consul Keygen

Command: `consul keygen`

The `keygen` command generates an encryption key that can be used for
[Consul agent traffic encryption](/docs/agent/encryption.html).
The keygen command uses a cryptographically
strong pseudo-random number generator to generate the key.
