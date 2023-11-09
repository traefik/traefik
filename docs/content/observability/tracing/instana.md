---
title: "Traefik Instana Documentation"
description: "Traefik supports several tracing backends, including Instana. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# Instana

You need first add `INSTANA_ENDPOINT_URL` and `INSTANA_AGENT_KEY` env variables to Traefik executable.

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

#### `globalTags`

_Optional, Default=empty_

Applies a list of shared key:value tags on all spans.

```yaml tab="File (YAML)"
tracing:
  instana:
    globalTags:
      tag1: foo
      tag2: bar
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    [tracing.instana.globalTags]
      tag1 = "foo"
      tag2 = "bar"
```

```bash tab="CLI"
--tracing.instana.globalTags.tag1=foo
--tracing.instana.globalTags.tag2=bar
```

#### `sampleRate`

_Optional, Default=1.0_

The proportion of requests to trace, specified between 0.0 and 1.0.

```yaml tab="File (YAML)"
tracing:
  instana:
    sampleRate: 0.2
```

```toml tab="File (TOML)"
[tracing]
  [tracing.instana]
    sampleRate = 0.2
```

```bash tab="CLI"
--tracing.instana.sampleRate=0.2
```
