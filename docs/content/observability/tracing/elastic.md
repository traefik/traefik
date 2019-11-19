# Elastic

To enable the Elastic:

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
```

```yaml tab="File (YAML)"
tracing:
  elastic: {}
```

```bash tab="CLI"
--tracing.elastic=true
```

#### `apmServerURL`

_Optional, Default="http://localhost:8200"_

APM ServerURL is the URL of the Elastic APM server.

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    apmServerURL = "http://apm:8200"
```

```yaml tab="File (YAML)"
tracing:
  elastic:
    apmServerURL: "http://apm:8200"
```

```bash tab="CLI"
--tracing.elastic.apmServerURL="http://apm:8200"
```

#### `apmSecretToken`

_Optional, Default=""_

APM Secret Token is the token used to connect to Elastic APM Server.

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    apmSecretToken = "mytoken"
```

```yaml tab="File (YAML)"
tracing:
  elastic:
    apmSecretToken: "mytoken"
```

```bash tab="CLI"
--tracing.elastic.apmSecretToken="mytoken"
```

#### `apmServiceEnvironment`

_Optional, Default=""_

APM Service Environment is the name of the environment Traefik is deployed in, e.g. `production` or `staging`.

```toml tab="File (TOML)"
[tracing]
  [tracing.elastic]
    apmServiceEnvironment = "production"
```

```yaml tab="File (YAML)"
tracing:
  elastic:
    apmServiceEnvironment: "production"
```

```bash tab="CLI"
--tracing.elastic.apmServiceEnvironment="production"
```

### Further

Additional configuration of Elastic APM Go agent can be done using environment variables.
See [APM Go agent reference](https://www.elastic.co/guide/en/apm/agent/go/current/configuration.html).
