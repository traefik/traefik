# Rancher Backend

Træfik can be configured to use Rancher as a backend configuration.

## Global Configuration

```toml
################################################################
# Rancher configuration backend
################################################################

# Enable Rancher configuration backend.
[rancher]

# Default domain used.
# Can be overridden by setting the "traefik.domain" label on an service.
#
# Required
#
domain = "rancher.localhost"

# Enable watch Rancher changes.
#
# Optional
# Default: true
#
watch = true

# Polling interval (in seconds).
#
# Optional
# Default: 15
#
refreshSeconds = 15

# Expose Rancher services by default in Traefik.
#
# Optional
# Default: true
#
exposedByDefault = false

# Filter services with unhealthy states and inactive states.
#
# Optional
# Default: false
#
enableServiceHealthFilter = true
```

To enable constraints see [backend-specific constraints section](/configuration/commons/#backend-specific).

## Rancher Metadata Service

```toml
# Enable Rancher metadata service configuration backend instead of the API
# configuration backend.
#
# Optional
# Default: false
#
[rancher.metadata]

# Poll the Rancher metadata service for changes every `rancher.RefreshSeconds`.
# NOTE: this is less accurate than the default long polling technique which
# will provide near instantaneous updates to Traefik
#
# Optional
# Default: false
#
intervalPoll = true

# Prefix used for accessing the Rancher metadata service.
#
# Optional
# Default: "/latest"
#
prefix = "/2016-07-29"
```

## Rancher API

```toml
# Enable Rancher API configuration backend.
#
# Optional
# Default: true
#
[rancher.api]

# Endpoint to use when connecting to the Rancher API.
#
# Required
endpoint = "http://rancherserver.example.com/v1"

# AccessKey to use when connecting to the Rancher API.
#
# Required
accessKey = "XXXXXXXXXXXXXXXXXXXX"

# SecretKey to use when connecting to the Rancher API.
#
# Required
secretKey = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

!!! note
    If Traefik needs access to the Rancher API, you need to set the `endpoint`, `accesskey` and `secretkey` parameters.

    To enable Traefik to fetch information about the Environment it's deployed in only, you need to create an `Environment API Key`.
    This can be found within the API Key advanced options.

## Labels: overriding default behaviour

Labels can be used on task containers to override default behaviour:

| Label                                        | Description                                                                              |
|----------------------------------------------|------------------------------------------------------------------------------------------|
| `traefik.protocol=https`                     | override the default `http` protocol                                                     |
| `traefik.weight=10`                          | assign this weight to the container                                                      |
| `traefik.enable=false`                       | disable this container in Træfik                                                         |
| `traefik.frontend.rule=Host:test.traefik.io` | override the default frontend rule (Default: `Host:{containerName}.{domain}`).           |
| `traefik.frontend.passHostHeader=true`       | forward client `Host` header to the backend.                                             |
| `traefik.frontend.priority=10`               | override default frontend priority                                                       |
| `traefik.frontend.entryPoints=http,https`    | assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`. |
| `traefik.frontend.auth.basic=EXPR`           | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`.        |
