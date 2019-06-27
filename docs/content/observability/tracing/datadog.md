# DataDog

To enable the DataDog:

```toml tab="File"
[tracing]
  [tracing.datadog]
```

```bash tab="CLI"
--tracing
--tracing.datadog
```

#### `localAgentHostPort`

_Required, Default="127.0.0.1:8126"_

Local Agent Host Port instructs reporter to send spans to datadog-tracing-agent at this address.

```toml tab="File"
[tracing]
  [tracing.datadog]
    localAgentHostPort = "127.0.0.1:8126"
```

```bash tab="CLI"
--tracing
--tracing.datadog.localAgentHostPort="127.0.0.1:8126"
```

#### `debug`

_Optional, Default=false_

Enable DataDog debug.

```toml tab="File"
[tracing]
  [tracing.datadog]
    debug = true
```

```bash tab="CLI"
--tracing
--tracing.datadog.debug=true
```

#### `globalTag`

_Optional, Default=empty_

Apply shared tag in a form of Key:Value to all the traces.

```toml tab="File"
[tracing]
  [tracing.datadog]
    globalTag = "sample"
```

```bash tab="CLI"
--tracing
--tracing.datadog.globalTag="sample"
```

#### `prioritySampling`

_Optional, Default=false_

Enable priority sampling. When using distributed tracing,
this option must be enabled in order to get all the parts of a distributed trace sampled.

```toml tab="File"
[tracing]
  [tracing.datadog]
    prioritySampling = true
```

```bash tab="CLI"
--tracing
--tracing.datadog.prioritySampling=true
```
