---
title: "Traefik Instana Documentation"
description: "Traefik supports several tracing backends, including Instana. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# Instana

To enable the Instana tracer:

```yaml tab="File (YAML)"
tracing:
  instana: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
```

```bash tab="CLI"
--tracing.instana=true
```

#### `localAgentHost`

_Required, Default="127.0.0.1"_

Local Agent Host instructs reporter to send spans to the Instana Agent at this address.

```yaml tab="File (YAML)"
tracing:
  instana:
    localAgentHost: 127.0.0.1
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    localAgentHost = "127.0.0.1"
```

```bash tab="CLI"
--tracing.instana.localAgentHost=127.0.0.1
```

#### `localAgentPort`

_Required, Default=42699_

Local Agent port instructs reporter to send spans to the Instana Agent listening on this port.

```yaml tab="File (YAML)"
tracing:
  instana:
    localAgentPort: 42699
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    localAgentPort = 42699
```

```bash tab="CLI"
--tracing.instana.localAgentPort=42699
```

#### `logLevel`

_Required, Default="info"_

Sets Instana tracer log level.

Valid values are:

- `error`
- `warn`
- `debug`
- `info`

```yaml tab="File (YAML)"
tracing:
  instana:
    logLevel: info
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    logLevel = "info"
```

```bash tab="CLI"
--tracing.instana.logLevel=info
```

#### `enableAutoProfile`

_Required, Default=false_

Enables [automatic profiling](https://www.ibm.com/docs/en/obi/current?topic=instana-profile-processes) for the Traefik process.

```yaml tab="File (YAML)"
tracing:
  instana:
    enableAutoProfile: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    enableAutoProfile = true
```

```bash tab="CLI"
--tracing.instana.enableAutoProfile=true
```
