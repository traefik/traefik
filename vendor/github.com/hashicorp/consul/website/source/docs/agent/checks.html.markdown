---
layout: "docs"
page_title: "Check Definition"
sidebar_current: "docs-agent-checks"
description: |-
  One of the primary roles of the agent is management of system- and application-level health checks. A health check is considered to be application-level if it is associated with a service. A check is defined in a configuration file or added at runtime over the HTTP interface.
---

# Checks

One of the primary roles of the agent is management of system-level and application-level health
checks. A health check is considered to be application-level if it is associated with a
service. If not associated with a service, the check monitors the health of the entire node.

A check is defined in a configuration file or added at runtime over the HTTP interface. Checks
created via the HTTP interface persist with that node.

There are five different kinds of checks:

* Script + Interval - These checks depend on invoking an external application
  that performs the health check, exits with an appropriate exit code, and potentially
  generates some output. A script is paired with an invocation interval (e.g.
  every 30 seconds). This is similar to the Nagios plugin system. The output of
  a script check is limited to 4K. Output larger than this will be truncated.
  By default, Script checks will be configured with a timeout equal to 30 seconds.
  It is possible to configure a custom Script check timeout value by specifying the
  `timeout` field in the check definition.

* HTTP + Interval - These checks make an HTTP `GET` request every Interval (e.g.
  every 30 seconds) to the specified URL. The status of the service depends on the HTTP response code:
  any `2xx` code is considered passing, a `429 Too Many Requests` is a warning, and anything else is a failure.
  This type of check should be preferred over a script that uses `curl` or another external process
  to check a simple HTTP operation. By default, HTTP checks will be configured
  with a request timeout equal to the check interval, with a max of 10 seconds.
  It is possible to configure a custom HTTP check timeout value by specifying
  the `timeout` field in the check definition. The output of the check is
  limited to roughly 4K. Responses larger than this will be truncated. HTTP checks
  also support SSL. By default, a valid SSL certificate is expected. Certificate
  verification can be turned off by setting the `tls_skip_verify` field to `true`
  in the check definition.

* TCP + Interval - These checks make an TCP connection attempt every Interval
  (e.g. every 30 seconds) to the specified IP/hostname and port. If no hostname
  is specified, it defaults to "localhost". The status of the service depends on
  whether the connection attempt is successful (ie - the port is currently
  accepting connections). If the connection is accepted, the status is
  `success`, otherwise the status is `critical`. In the case of a hostname that
  resolves to both IPv4 and IPv6 addresses, an attempt will be made to both
  addresses, and the first successful connection attempt will result in a
  successful check. This type of check should be preferred over a script that
  uses `netcat` or another external process to check a simple socket operation.
  By default, TCP checks will be configured with a request timeout equal to the
  check interval, with a max of 10 seconds. It is possible to configure a custom
  TCP check timeout value by specifying the `timeout` field in the check
  definition.

* <a name="TTL"></a>Time to Live (TTL) - These checks retain their last known state for a given TTL.
  The state of the check must be updated periodically over the HTTP interface. If an
  external system fails to update the status within a given TTL, the check is
  set to the failed state. This mechanism, conceptually similar to a dead man's switch,
  relies on the application to directly report its health. For example, a healthy app
  can periodically `PUT` a status update to the HTTP endpoint; if the app fails, the TTL will
  expire and the health check enters a critical state. The endpoints used to
  update health information for a given check are the
  [pass endpoint](https://www.consul.io/api/agent.html#agent_check_pass)
  and the [fail endpoint](https://www.consul.io/api/agent.html#agent_check_fail).
  TTL checks also persist
  their last known status to disk. This allows the Consul agent to restore the
  last known status of the check across restarts. Persisted check status is
  valid through the end of the TTL from the time of the last check.

* Docker + Interval - These checks depend on invoking an external application which
is packaged within a Docker Container. The application is triggered within the running
container via the Docker Exec API. We expect that the Consul agent user has access
to either the Docker HTTP API or the unix socket. Consul uses ```$DOCKER_HOST``` to
determine the Docker API endpoint. The application is expected to run, perform a health
check of the service running inside the container, and exit with an appropriate exit code.
The check should be paired with an invocation interval. The shell on which the check
has to be performed is configurable which makes it possible to run containers which
have different shells on the same host. Check output for Docker is limited to
4K. Any output larger than this will be truncated.

## Check Definition

A script check:

```javascript
{
  "check": {
    "id": "mem-util",
    "name": "Memory utilization",
    "script": "/usr/local/bin/check_mem.py",
    "interval": "10s",
    "timeout": "1s"
  }
}
```

A HTTP check:

```javascript
{
  "check": {
    "id": "api",
    "name": "HTTP API on port 5000",
    "http": "http://localhost:5000/health",
    "interval": "10s",
    "timeout": "1s"
  }
}
```

A TCP check:

```javascript
{
  "check": {
    "id": "ssh",
    "name": "SSH TCP on port 22",
    "tcp": "localhost:22",
    "interval": "10s",
    "timeout": "1s"
  }
}
```

A TTL check:

```javascript
{
  "check": {
    "id": "web-app",
    "name": "Web App Status",
    "notes": "Web app does a curl internally every 10 seconds",
    "ttl": "30s"
  }
}
```

A Docker check:

```javascript
{
"check": {
    "id": "mem-util",
    "name": "Memory utilization",
    "docker_container_id": "f972c95ebf0e",
    "shell": "/bin/bash",
    "script": "/usr/local/bin/check_mem.py",
    "interval": "10s"
  }
}
```

Each type of definition must include a `name` and may optionally
provide an `id` and `notes` field. The `id` is set to the `name` if not
provided. It is required that all checks have a unique ID per node: if names
might conflict, unique IDs should be provided.

The `notes` field is opaque to Consul but can be used to provide a human-readable
description of the current state of the check. With a script check, the field is
set to any output generated by the script. Similarly, an external process updating
a TTL check via the HTTP interface can set the `notes` value.

Checks may also contain a `token` field to provide an ACL token. This token is
used for any interaction with the catalog for the check, including
[anti-entropy syncs](/docs/internals/anti-entropy.html) and deregistration.

Script, TCP, Docker and HTTP checks must include an `interval` field. This field is
parsed by Go's `time` package, and has the following
[formatting specification](https://golang.org/pkg/time/#ParseDuration):
> A duration string is a possibly signed sequence of decimal numbers, each with
> optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
> Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

In Consul 0.7 and later, checks that are associated with a service may also contain
an optional `deregister_critical_service_after` field, which is a timeout in the
same Go time format as `interval` and `ttl`. If a check is in the critical state
for more than this configured value, then its associated service (and all of its
associated checks) will automatically be deregistered. The minimum timeout is 1
minute, and the process that reaps critical services runs every 30 seconds, so it
may take slightly longer than the configured timeout to trigger the deregistration.
This should generally be configured with a timeout that's much, much longer than
any expected recoverable outage for the given service.

To configure a check, either provide it as a `-config-file` option to the
agent or place it inside the `-config-dir` of the agent. The file must
end in the ".json" extension to be loaded by Consul. Check definitions can
also be updated by sending a `SIGHUP` to the agent. Alternatively, the
check can be registered dynamically using the [HTTP API](/api/index.html).

## Check Scripts

A check script is generally free to do anything to determine the status
of the check. The only limitations placed are that the exit codes must obey
this convention:

 * Exit code 0 - Check is passing
 * Exit code 1 - Check is warning
 * Any other code - Check is failing

This is the only convention that Consul depends on. Any output of the script
will be captured and stored in the `notes` field so that it can be viewed
by human operators.

## Initial Health Check Status

By default, when checks are registered against a Consul agent, the state is set
immediately to "critical". This is useful to prevent services from being
registered as "passing" and entering the service pool before they are confirmed
to be healthy. In certain cases, it may be desirable to specify the initial
state of a health check. This can be done by specifying the `status` field in a
health check definition, like so:

```javascript
{
  "check": {
    "id": "mem",
    "script": "/bin/check_mem",
    "interval": "10s",
    "status": "passing"
  }
}
```

The above service definition would cause the new "mem" check to be
registered with its initial state set to "passing".

## Service-bound checks

Health checks may optionally be bound to a specific service. This ensures
that the status of the health check will only affect the health status of the
given service instead of the entire node. Service-bound health checks may be
provided by adding a `service_id` field to a check configuration:

```javascript
{
  "check": {
    "id": "web-app",
    "name": "Web App Status",
    "service_id": "web-app",
    "ttl": "30s"
  }
}
```

In the above configuration, if the web-app health check begins failing, it will
only affect the availability of the web-app service. All other services
provided by the node will remain unchanged.

## Multiple Check Definitions

Multiple check definitions can be defined using the `checks` (plural)
key in your configuration file.

```javascript
{
  "checks": [
    {
      "id": "chk1",
      "name": "mem",
      "script": "/bin/check_mem",
      "interval": "5s"
    },
    {
      "id": "chk2",
      "name": "/health",
      "http": "http://localhost:5000/health",
      "interval": "15s"
    },
    {
      "id": "chk3",
      "name": "cpu",
      "script": "/bin/check_cpu",
      "interval": "10s"
    },
    ...
  ]
}
```
