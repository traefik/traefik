---
title: "Traefik Elastic Documentation"
description: "Traefik supports several tracing backends, including Elastic. Learn how to implement it for observability in Traefik Proxy. Read the technical documentation."
---

# Elastic

To enable the Elastic tracer:

```yaml tab="File (YAML)"
tracing:
  elastic: {}
```

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
```

```bash tab="CLI"
--tracing.elastic=true
```

#### `serverURL`

_Optional, Default="http://localhost:8200"_

URL of the Elastic APM server.

```yaml tab="File (YAML)"
tracing:
  elastic:
    serverURL: "http://apm:8200"
```

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    serverURL = "http://apm:8200"
```

```bash tab="CLI"
--tracing.elastic.serverurl="http://apm:8200"
```

#### `secretToken`

_Optional, Default=""_

Token used to connect to Elastic APM Server.

```yaml tab="File (YAML)"
tracing:
  elastic:
    secretToken: "mytoken"
```

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    secretToken = "mytoken"
```

```bash tab="CLI"
--tracing.elastic.secrettoken="mytoken"
```

#### `serviceEnvironment`

_Optional, Default=""_

Environment's name where Traefik is deployed in, e.g. `production` or `staging`.

```yaml tab="File (YAML)"
tracing:
  elastic:
    serviceEnvironment: "production"
```

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    serviceEnvironment = "production"
```

```bash tab="CLI"
--tracing.elastic.serviceenvironment="production"
```

#### `globalTags`

_Optional, Default=empty_

Applies a list of shared key:value tags on all spans.

```yaml tab="File (YAML)"
tracing:
  elastic:
    globalTags:
      tag1: foo
      tag2: bar
```

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    [tracing.elastic.globalTags]
      tag1 = "foo"
      tag2 = "bar"
```

```bash tab="CLI"
--tracing.elastic.globalTags.tag1=foo
--tracing.elastic.globalTags.tag2=bar
```

#### `sampleRate`

_Optional, Default=1.0_

The proportion of requests to trace, specified between 0.0 and 1.0.

```yaml tab="File (YAML)"
tracing:
  elastic:
    sampleRate: 0.2
```

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    sampleRate = 0.2
```

```bash tab="CLI"
--tracing.elastic.sampleRate=0.2
```

### Further

Additional configuration of Elastic APM Go agent can be done using environment variables.
See [APM Go agent reference](https://www.elastic.co/guide/en/apm/agent/go/current/configuration.html).
