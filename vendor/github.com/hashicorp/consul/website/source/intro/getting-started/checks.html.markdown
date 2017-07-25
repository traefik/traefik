---
layout: "intro"
page_title: "Registering Health Checks"
sidebar_current: "gettingstarted-checks"
description: |-
  We've now seen how simple it is to run Consul, add nodes and services, and query those nodes and services. In this step, we will continue our tour by adding health checks to both nodes and services. Health checks are a critical component of service discovery that prevent using services that are unhealthy.
---

# Health Checks

We've now seen how simple it is to run Consul, add nodes and services, and
query those nodes and services. In this section, we will continue our tour
by adding health checks to both nodes and services. Health checks are a
critical component of service discovery that prevent using services that
are unhealthy.

This step builds upon [the Consul cluster created previously](join.html).
At this point, you should have a two-node cluster running.

## Defining Checks

Similar to a service, a check can be registered either by providing a
[check definition](/docs/agent/checks.html) or by making the
appropriate calls to the [HTTP API](/api/health.html).

We will use the check definition approach because, just like with
services, definitions are the most common way to set up checks.

Create two definition files in the Consul configuration directory of
the second node:

```text
vagrant@n2:~$ echo '{"check": {"name": "ping",
  "script": "ping -c1 google.com >/dev/null", "interval": "30s"}}' \
  >/etc/consul.d/ping.json

vagrant@n2:~$ echo '{"service": {"name": "web", "tags": ["rails"], "port": 80,
  "check": {"script": "curl localhost >/dev/null 2>&1", "interval": "10s"}}}' \
  >/etc/consul.d/web.json
```

The first definition adds a host-level check named "ping". This check runs
on a 30 second interval, invoking `ping -c1 google.com`. On a `script`-based
health check, the check runs as the same user that started the Consul process.
If the command exits with a non-zero exit code, then the node will be flagged
unhealthy. This is the contract for any `script`-based health check.

The second command modifies the service named `web`, adding a check that sends a
request every 10 seconds via curl to verify that the web server is accessible.
As with the host-level health check, if the script exits with a non-zero exit code,
the service will be flagged unhealthy.

Now, restart the second agent, reload it with `consul reload`, or send it a `SIGHUP` signal. You should see the
following log lines:

```text
==> Starting Consul agent...
...
    [INFO] agent: Synced service 'web'
    [INFO] agent: Synced check 'service:web'
    [INFO] agent: Synced check 'ping'
    [WARN] Check 'service:web' is now critical
```

The first few lines indicate that the agent has synced the new
definitions. The last line indicates that the check we added for
the `web` service is critical. This is because we're not actually running
a web server, so the curl test is failing!

## Checking Health Status

Now that we've added some simple checks, we can use the HTTP API to inspect
them. First, we can look for any failing checks using this command (note, this
can be run on either node):

```text
vagrant@n1:~$ curl http://localhost:8500/v1/health/state/critical
[{"Node":"agent-two","CheckID":"service:web","Name":"Service 'web' check","Status":"critical","Notes":"","ServiceID":"web","ServiceName":"web"}]
```

We can see that there is only a single check, our `web` service check, in the
`critical` state.

Additionally, we can attempt to query the web service using DNS. Consul
will not return any results since the service is unhealthy:

```text
dig @127.0.0.1 -p 8600 web.service.consul
...

;; QUESTION SECTION:
;web.service.consul.		IN	A
```

## Next Steps

In this section, you learned how easy it is to add health checks. Check definitions
can be updated by changing configuration files and sending a `SIGHUP` to the agent.
Alternatively, the HTTP API can be used to add, remove, and modify checks dynamically.
The API also allows for a "dead man's switch", a
[TTL-based check](/docs/agent/checks.html#TTL). TTL checks can be used to integrate an
application more tightly with Consul, enabling business logic to be evaluated as part
of assessing the state of the check.

Next, we will explore [Consul's K/V store](kv.html).
