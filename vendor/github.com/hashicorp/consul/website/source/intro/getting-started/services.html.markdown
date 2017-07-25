---
layout: "intro"
page_title: "Registering Services"
sidebar_current: "gettingstarted-services"
description: >
  A service can be registered either by providing a service definition or by
  making the appropriate calls to the HTTP API. A configuration file is the
  most common, so we will use this approach to register a service, and then
  query that service using the REST API and DNS interfaces.
---

# Registering Services

In the previous step, we ran our first agent, saw the cluster members (well,
our cluster _member_), and queried that node. In this guide, we'll register
our first service and query that service.

## Defining a Service

A service can be registered either by providing a
[service definition](/docs/agent/services.html) or by making the appropriate
calls to the [HTTP API](/api/index.html).

A service definition is the most common way to register services, so we'll
use that approach for this step. We'll be building on the agent configuration
we covered in the [previous step](/intro/getting-started/agent.html).

First, create a directory for Consul configuration. Consul loads all
configuration files in the configuration directory, so a common convention
on Unix systems is to name the directory something like `/etc/consul.d`
(the `.d` suffix implies "this directory contains a set of configuration
files").

```text
$ sudo mkdir /etc/consul.d
```

Next, we'll write a service definition configuration file. Let's
pretend we have a service named "web" running on port 80. Additionally,
we'll give it a tag we can use as an additional way to query the service:

```text
$ echo '{"service": {"name": "web", "tags": ["rails"], "port": 80}}' \
    | sudo tee /etc/consul.d/web.json
```

Now, restart the agent, providing the configuration directory:

```text
$ consul agent -dev -config-dir=/etc/consul.d
==> Starting Consul agent...
...
    [INFO] agent: Synced service 'web'
...
```

You'll notice in the output that it "synced" the web service. This means
that the agent loaded the service definition from the configuration file,
and has successfully registered it in the service catalog.

If you wanted to register multiple services, you could create multiple
service definition files in the Consul configuration directory.

## Querying Services

Once the agent is started and the service is synced, we can query the
service using either the DNS or HTTP API.

### DNS API

Let's first query our service using the DNS API. For the DNS API, the
DNS name for services is `NAME.service.consul`. By default, all DNS names
are always in the `consul` namespace, though
[this is configurable](/docs/agent/options.html#domain). The `service`
subdomain tells Consul we're querying services, and the `NAME` is the name
of the service.

For the web service we registered, these conventions and settings yield a
fully-qualified domain name of `web.service.consul`:

```text
$ dig @127.0.0.1 -p 8600 web.service.consul
...

;; QUESTION SECTION:
;web.service.consul.		IN	A

;; ANSWER SECTION:
web.service.consul.	0	IN	A	172.20.20.11
```

As you can see, an `A` record was returned with the IP address of the node on
which the service is available. `A` records can only hold IP addresses.

You can also use the DNS API to retrieve the entire address/port pair as a
`SRV` record:

```text
$ dig @127.0.0.1 -p 8600 web.service.consul SRV
...

;; QUESTION SECTION:
;web.service.consul.		IN	SRV

;; ANSWER SECTION:
web.service.consul.	0	IN	SRV	1 1 80 Armons-MacBook-Air.node.dc1.consul.

;; ADDITIONAL SECTION:
Armons-MacBook-Air.node.dc1.consul. 0 IN A	172.20.20.11
```

The `SRV` record says that the web service is running on port 80 and exists on
the node `Armons-MacBook-Air.node.dc1.consul.`. An additional section is returned by the
DNS with the `A` record for that node.

Finally, we can also use the DNS API to filter services by tags. The
format for tag-based service queries is `TAG.NAME.service.consul`. In
the example below, we ask Consul for all web services with the "rails"
tag. We get a successful response since we registered our service with
that tag:

```text
$ dig @127.0.0.1 -p 8600 rails.web.service.consul
...

;; QUESTION SECTION:
;rails.web.service.consul.		IN	A

;; ANSWER SECTION:
rails.web.service.consul.	0	IN	A	172.20.20.11
```

### HTTP API

In addition to the DNS API, the HTTP API can be used to query services:

```text
$ curl http://localhost:8500/v1/catalog/service/web
[{"Node":"Armons-MacBook-Air","Address":"172.20.20.11","ServiceID":"web", \
	"ServiceName":"web","ServiceTags":["rails"],"ServicePort":80}]
```

The catalog API gives all nodes hosting a given service. As we will see later
with [health checks](/intro/getting-started/checks.html) you'll typically want
to query just for healthy instances where the checks are passing. This is what
DNS is doing under the hood. Here's a query to look for only healthy instances:

```text
$ curl 'http://localhost:8500/v1/health/service/web?passing'
[{"Node":"Armons-MacBook-Air","Address":"172.20.20.11","Service":{ \
	"ID":"web", "Service":"web", "Tags":["rails"],"Port":80}, "Checks": ...}]
```

## Updating Services

Service definitions can be updated by changing configuration files and
sending a `SIGHUP` to the agent. This lets you update services without
any downtime or unavailability to service queries.

Alternatively, the HTTP API can be used to add, remove, and modify services
dynamically.

## Next Steps

We've now configured a single agent and registered a service. This is good
progress, but let's explore the full value of Consul by [setting up our
first cluster](/intro/getting-started/join.html)!
