---
layout: "docs"
page_title: "Security Model"
sidebar_current: "docs-internals-security"
description: |-
  Consul relies on both a lightweight gossip mechanism and an RPC system to provide various features. Both of the systems have different security mechanisms that stem from their designs. However, the security mechanisms of Consul have a common goal: to provide confidentiality, integrity, and authentication.
---

# Security Model

Consul relies on both a lightweight gossip mechanism and an RPC system
to provide various features. Both of the systems have different security
mechanisms that stem from their designs. However, the security mechanisms
of Consul have a common goal: to provide
[confidentiality, integrity, and authentication](https://en.wikipedia.org/wiki/Information_security).

The [gossip protocol](/docs/internals/gossip.html) is powered by [Serf](https://www.serf.io/),
which uses a symmetric key, or shared secret, cryptosystem. There are more
details on the security of [Serf here](https://www.serf.io/docs/internals/security.html).
For details on how to enable Serf's gossip encryption in Consul, see the
[encryption doc here](/docs/agent/encryption.html).

The RPC system supports using end-to-end TLS with optional client authentication.
[TLS](https://en.wikipedia.org/wiki/Transport_Layer_Security) is a widely deployed asymmetric
cryptosystem and is the foundation of security on the Web.

This means Consul communication is protected against eavesdropping, tampering,
and spoofing. This makes it possible to run Consul over untrusted networks such
as EC2 and other shared hosting providers.

~> **Advanced Topic!** This page covers the technical details of
the security model of Consul. You don't need to know these details to
operate and use Consul. These details are documented here for those who wish
to learn about them without having to go spelunking through the source code.

## Threat Model

The following are the various parts of our threat model:

* Non-members getting access to data
* Cluster state manipulation due to malicious messages
* Fake data generation due to malicious messages
* Tampering causing state corruption
* Denial of Service against a node

Additionally, we recognize that an attacker that can observe network
traffic for an extended period of time may infer the cluster members.
The gossip mechanism used by Consul relies on sending messages to random
members, so an attacker can record all destinations and determine all
members of the cluster.

When designing security into a system you design it to fit the threat model.
Our goal is not to protect top secret data but to provide a "reasonable"
level of security that would require an attacker to commit a considerable
amount of resources to defeat.

## Network Ports

For configuring network rules to support Consul, please see [Ports Used](/docs/agent/options.html#ports)
for a listing of network ports used by Consul and details about which features
they are used for.
