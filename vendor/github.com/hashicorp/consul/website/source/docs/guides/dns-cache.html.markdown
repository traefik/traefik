---
layout: "docs"
page_title: "DNS Caching"
sidebar_current: "docs-guides-dns-cache"
description: |-
  One of the main interfaces to Consul is DNS. Using DNS is a simple way to integrate Consul into an existing infrastructure without any high-touch integration.
---

# DNS Caching

One of the main interfaces to Consul is DNS. Using DNS is a simple way to
integrate Consul into an existing infrastructure without any high-touch
integration.

By default, Consul serves all DNS results with a 0 TTL value. This prevents
any caching. The advantage is that each DNS lookup is always re-evaluated,
so the most timely information is served. However, this adds a latency hit
for each lookup and can potentially exhaust the query throughput of a cluster.

For this reason, Consul provides a number of tuning parameters that can
customize how DNS queries are handled.

<a name="stale"></a>
## Stale Reads

Stale reads can be used to reduce latency and increase the throughput
of DNS queries. By default, all reads are serviced by a
[single leader node](/docs/internals/consensus.html).
These reads are strongly consistent but are limited by the throughput
of a single node. Doing a stale read allows any Consul server to
service a query, but non-leader nodes may return data that is
out-of-date. By allowing data to be slightly stale, we get horizontal
read scalability. Now any Consul server can service the request, so we
increase throughput by the number of servers in a cluster.

The [settings](/docs/agent/options.html) used to control stale reads
are [`dns_config.allow_stale`](/docs/agent/options.html#allow_stale),
which must be set to enable stale reads, and
[`dns_config.max_stale`](/docs/agent/options.html#max_stale)
which limits how stale results are allowed to be.

Starting from Consul 0.7, [`allow_stale`](/docs/agent/options.html#allow_stale)
is enabled by default, using a [`max_stale`](/docs/agent/options.html#max_stale)
value that defaults to 5 seconds, meaning that we will use data from
any Consul server that is within 5 seconds of the leader. In Consul 0.7.1, the
default for `max_stale` was been increased from 5 seconds to a near-indefinite
threshold (10 years) to allow DNS queries to continue to be served in the event
of a long outage with no leader. A new telemetry counter has also been added at
`consul.dns.stale_queries` to track when agents serve DNS queries that are stale
by more than 5 seconds.

## Negative Response Caching

Some DNS clients cache negative responses - that is, Consul returning a "not
found" style response because a service exists but there are no healthy
endpoints. What this means in practice is that cached negative responses may
mean that services appear "down" for longer than they are actually unavailable
when using DNS for service discovery.

One common example is that Windows will default to caching negative responses
for 15 minutes. DNS forwarders may also cache negative responses, with the same
effect. To avoid this problem, check the negative response cache defaults for
your client operating system and any DNS forwarder on the path between the
client and Consul and set the cache values appropriately. In many cases
"appropriately" simply is turning negative response caching off to get the best
recovery time when a service becomes available again.

<a name="ttl"></a>
## TTL Values

TTL values can be set to allow DNS results to be cached downstream of Consul. Higher
TTL values reduce the number of lookups on the Consul servers and speed lookups for
clients, at the cost of increasingly stale results. By default, all TTLs are zero,
preventing any caching.

To enable caching of node lookups (e.g. "foo.node.consul"), we can set the
[`dns_config.node_ttl`](/docs/agent/options.html#node_ttl) value. This can be set to
"10s" for example, and all node lookups will serve results with a 10 second TTL.

Service TTLs can be specified in a more granular fashion. You can set TTLs
per-service, with a wildcard TTL as the default. This is specified using the
[`dns_config.service_ttl`](/docs/agent/options.html#service_ttl) map. The "*"
service is the wildcard service.

For example, a [`dns_config`](/docs/agent/options.html#dns_config) that provides
a wildcard TTL and a specific TTL for a service might look like this:

```javascript
{
  "dns_config": {
    "service_ttl": {
      "*": "5s",
      "web": "30s"
    }
  }
}
```

This sets all lookups to "web.service.consul" to use a 30 second TTL
while lookups to "db.service.consul" or "api.service.consul" will use the
5 second TTL from the wildcard.

[Prepared Queries](/api/query.html) provide an additional
level of control over TTL. They allow for the TTL to be defined along with
the query, and they can be changed on the fly by updating the query definition.
If a TTL is not configured for a prepared query, then it will fall back to the
service-specific configuration defined in the Consul agent as described above,
and ultimately to 0 if no TTL is configured for the service in the Consul agent.
