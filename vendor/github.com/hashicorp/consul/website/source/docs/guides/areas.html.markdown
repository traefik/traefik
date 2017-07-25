---
layout: "docs"
page_title: "Multiple Datacenters - Advanced Federation with Network Areas"
sidebar_current: "docs-guides-areas"
description: |-
  One of the key features of Consul is its support for multiple datacenters. The architecture of Consul is designed to promote low coupling of datacenters so that connectivity issues or failure of any datacenter does not impact the availability of Consul in other datacenters. This means each datacenter runs independently, each having a dedicated group of servers and a private LAN gossip pool.
---

# Multiple Datacenters
## Advanced Federation with Network Areas

~> The network area functionality described here is available only in
   [Consul Enterprise](https://www.hashicorp.com/products/consul/) version 0.8.0 and later.

One of the key features of Consul is its support for multiple datacenters.
The [architecture](/docs/internals/architecture.html) of Consul is designed to
promote a low coupling of datacenters so that connectivity issues or
failure of any datacenter does not impact the availability of Consul in other
datacenters. This means each datacenter runs independently, each having a dedicated
group of servers and a private LAN [gossip pool](/docs/internals/gossip.html).

This guide covers the advanced form of federating Consul clusters using the new
network areas capability added in [Consul Enterprise](https://www.hashicorp.com/products/consul/)
version 0.8.0. For the basic form of federation available in the open source version
of Consul, please see the [Basic Federation Guide](/docs/guides/datacenters.html)
for more details.

## Network Areas

Consul's [Basic Federation](/docs/guides/datacenters.html) support relies on all
Consul servers in all datacenters having full mesh connectivity via server RPC
(8300/tcp) and Serf WAN (8302/tcp and 8302/udp). Securing this setup requires TLS
in combination with managing a gossip keyring. With massive Consul deployments, it
becomes tricky to support a full mesh with all Consul servers, and to manage the
keyring.

Consul Enterprise version 0.8.0 added support for a new federation model based on
operator-created network areas. Network areas specify a relationship between a
pair of Consul datacenters. Operators create reciprocal areas on each side of the
relationship and then join them together, so a given Consul datacenter can participate
in many areas, even when some of the peer areas cannot contact each other. This
allows for more flexible relationships between Consul datacenters, such as hub/spoke
or more general tree structures. Traffic between areas is all performed via server
RPC (8300/tcp) so it can be secured with just TLS.

Currently, Consul will only route RPC requests to datacenters it is immediately adjacent
to via an area (or via the WAN), but future versions of Consul may add routing support.

The following can be used to manage network areas:

* [Network Areas HTTP Endpoint](/api/operator.html#network-areas)
* [Network Areas Operator CLI](/docs/commands/operator/area.html)

## Network Areas and the WAN Gossip Pool

Networks areas can be used alongside the Consul's [Basic Federation](/docs/guides/datacenters.html)
model and the WAN gossip pool. This helps ease migration, and clusters like the
[ACL datacenter](/docs/agent/options.html#acl_datacenter) are more easily managed via
the WAN because they need to be available to all Consul datacenters.

A peer datacenter can connected via the WAN gossip pool and a network area at the
same time, and RPCs will be forwarded as long as servers are available in either.

## Getting Started

To get started, follow the [bootstrapping guide](/docs/guides/bootstrapping.html) to
start each datacenter. After bootstrapping, we should have two datacenters now which
we can refer to as `dc1` and `dc2`. Note that datacenter names are opaque to Consul;
they are simply labels that help human operators reason about the Consul clusters.

A compatible pair of areas must be created in each datacenter:

```text
(dc1) $ consul operator area create -peer-datacenter=dc2
Created area "cbd364ae-3710-1770-911b-7214e98016c0" with peer datacenter "dc2"!
```

```text
(dc2) $ consul operator area create -peer-datacenter=dc1
Created area "2aea3145-f1e3-cb1d-a775-67d15ddd89bf" with peer datacenter "dc1"!
```

Now you can query for the members of the area:

```text
(dc1) $ consul operator area members
Area                                  Node        Address         Status  Build         Protocol  DC   RTT
cbd364ae-3710-1770-911b-7214e98016c0  node-1.dc1  127.0.0.1:8300  alive   0.8.0_entrc1  2         dc1  0s
```

Consul will automatically make sure that all servers within the datacenter where
the area was created are joined to the area using the LAN information. We need to
join with at least one Consul server in the other datacenter to complete the area:

```text
(dc1) $ consul operator area join -peer-datacenter=dc2 127.0.0.2
Address    Joined  Error
127.0.0.2  true    (none)
```

With a successful join, we should now see the remote Consul servers as part of the
area's members:

```text
(dc1) $ consul operator area members
Area                                  Node        Address         Status  Build         Protocol  DC   RTT
cbd364ae-3710-1770-911b-7214e98016c0  node-1.dc1  127.0.0.1:8300  alive   0.8.0_entrc1  2         dc1  0s
cbd364ae-3710-1770-911b-7214e98016c0  node-2.dc2  127.0.0.2:8300  alive   0.8.0_entrc1  2         dc2  581.649µs
```

Now we can route RPC commands in both directions. Here's a sample command to set a KV
entry in dc2 from dc1:

```text
(dc1) $ consul kv put -datacenter=dc2 hello world
Success! Data written to: hello
```

The DNS interface supports federation as well:

```text
(dc1) $ dig @127.0.0.1 -p 8600 consul.service.dc2.consul

; <<>> DiG 9.8.3-P1 <<>> @127.0.0.1 -p 8600 consul.service.dc2.consul
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 49069
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;consul.service.dc2.consul.     IN      A

;; ANSWER SECTION:
consul.service.dc2.consul. 0    IN      A       127.0.0.2

;; Query time: 3 msec
;; SERVER: 127.0.0.1#8600(127.0.0.1)
;; WHEN: Wed Mar 29 11:27:35 2017
;; MSG SIZE  rcvd: 59
```

There are a few networking requirements that must be satisfied for this to
work. Of course, all server nodes must be able to talk to each other via their server
RPC ports (8300/tcp). If service discovery is to be used across datacenters, the
network must be able to route traffic between IP addresses across regions as well.
Usually, this means that all datacenters must be connected using a VPN or other
tunneling mechanism. Consul does not handle VPN or NAT traversal for you.

The [`translate_wan_addrs`](/docs/agent/options.html#translate_wan_addrs) configuration
provides a basic address rewriting capability.

