# Instana

To enable the Instana:

```toml tab="File"
[tracing]
  [tracing.instana]
```

```bash tab="CLI"
--tracing
--tracing.instana
```

#### `localAgentHost`

_Require, Default="127.0.0.1"_

Local Agent Host instructs reporter to send spans to instana-agent at this address.

```toml tab="File"
[tracing]
  [tracing.instana]
    localAgentHost = "127.0.0.1"
```

```bash tab="CLI"
--tracing
--tracing.instana.localAgentHost="127.0.0.1"
```

#### `localAgentPort`

_Require, Default=42699_

Local Agent port instructs reporter to send spans to the instana-agent at this port.

```toml tab="File"
[tracing]
  [tracing.instana]
    localAgentPort = 42699
```

```bash tab="CLI"
--tracing
--tracing.instana.localAgentPort=42699
```

#### `logLevel`

_Require, Default="info"_

Set Instana tracer log level.

Valid values for logLevel field are:

- `error`
- `warn`
- `debug`
- `info`

```toml tab="File"
[tracing]
  [tracing.instana]
    logLevel = "info"
```

```bash tab="CLI"
--tracing
--tracing.instana.logLevel="info"
```
