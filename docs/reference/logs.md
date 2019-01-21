# Logs - Reference

## TOML

```toml
logLevel = "INFO"

[traefikLog]
  filePath = "/path/to/traefik.log"
  format   = "json"

[accessLog]
  filePath = "/path/to/access.log"
  format = "json"

  [accessLog.filters]
    statusCodes = ["200", "300-302"]
    retryAttempts = true
    minDuration = "10ms"

  [accessLog.fields]
    defaultMode = "keep"
    [accessLog.fields.names]
      "ClientUsername" = "drop"
      # ...

    [accessLog.fields.headers]
      defaultMode = "keep"
      [accessLog.fields.headers.names]
        "User-Agent" = "redact"
        "Authorization" = "drop"
        "Content-Type" = "keep"
        # ...
```

## CLI

For more information about the CLI, see the documentation about [Traefik command](/basics/#traefik).

```shell
--logLevel="DEBUG"
--traefikLog.filePath="/path/to/traefik.log"
--traefikLog.format="json"
--accessLog.filePath="/path/to/access.log"
--accessLog.format="json"
--accessLog.filters.statusCodes="200,300-302"
--accessLog.filters.retryAttempts="true"
--accessLog.filters.minDuration="10ms"
--accessLog.fields.defaultMode="keep"
--accessLog.fields.names="Username=drop Hostname=drop"
--accessLog.fields.headers.defaultMode="keep"
--accessLog.fields.headers.names="User-Agent=redact Authorization=drop Content-Type=keep"
```

