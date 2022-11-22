---
title: "Traefik Datadog Tracing Documentation"
description: "Traefik Proxy supports Datadog for tracing. Read the technical documentation to enable Datadog for observability."
---

# Datadog

To enable the Datadog tracer:

```yaml tab="File (YAML)"
tracing:
  datadog: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
```

```bash tab="CLI"
--tracing.datadog=true
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:8126"_

Local Agent Host Port instructs the reporter to send spans to the Datadog Agent at this address (host:port).

```yaml tab="File (YAML)"
tracing:
  datadog:
    localAgentHostPort: 127.0.0.1:8126
```

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    localAgentHostPort = "127.0.0.1:8126"
```

```bash tab="CLI"
--tracing.datadog.localAgentHostPort=127.0.0.1:8126
```

#### `debug`

_Optional, Default=false_

Enables Datadog debug.

```yaml tab="File (YAML)"
tracing:
  datadog:
    debug: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    debug = true
```

```bash tab="CLI"
--tracing.datadog.debug=true
```

#### `globalTags`

_Optional, Default=empty_

Applies a list of shared key:value tags on all spans.

```yaml tab="File (YAML)"
tracing:
  datadog:
    globalTags:
      tag1: foo
      tag2: bar
```

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    [tracing.datadog.globalTags]
      tag1 = "foo"
      tag2 = "bar"
```

```bash tab="CLI"
--tracing.datadog.globalTags.tag1=foo
--tracing.datadog.globalTags.tag2=bar
```

#### `prioritySampling`

_Optional, Default=false_

Enables priority sampling.
When using distributed tracing, 
this option must be enabled in order to get all the parts of a distributed trace sampled.

```yaml tab="File (YAML)"
tracing:
  datadog:
    prioritySampling: true
```

```toml tab="File (TOML)"
[tracing]
  [tracing.datadog]
    prioritySampling = true
```

```bash tab="CLI"
--tracing.datadog.prioritySampling=true
```
