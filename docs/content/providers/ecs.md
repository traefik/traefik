# Traefik & AWS ECS

A Story of Labels & Elastic Containers
{: .subtitle }

Attach labels to your ECS containers and let Traefik do the rest!

## Configuration Examples

??? example "Configuring ECS provider"

    Enabling the ECS provider:
    
    ```toml tab="File (TOML)"
    [providers.ecs]
      clusters = ["default"]
    ```
    
    ```yaml tab="File (YAML)"
    providers:
      ecs:
        clusters:
          - default
    ```
    
    ```bash tab="CLI"
    --providers.ecs.clusters=default
    ```

## Policy

Traefik needs the following policy to read ECS information:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "TraefikECSReadAccess",
            "Effect": "Allow",
            "Action": [
                "ecs:ListClusters",
                "ecs:DescribeClusters",
                "ecs:ListTasks",
                "ecs:DescribeTasks",
                "ecs:DescribeContainerInstances",
                "ecs:DescribeTaskDefinition",
                "ec2:DescribeInstances"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

## Provider configuration

### `autoDiscoverClusters`

_Optional, Default=false_

```toml tab="File (TOML)"
[providers.ecs]
  autoDiscoverClusters = true
  # ...
```

```yaml tab="File (YAML)"
providers:
  ecs:
    autoDiscoverClusters: true
    # ...
```

```bash tab="CLI"
--providers.ecs.autoDiscoverClusters=true
# ...
```

Search for services in clusters list.

- If set to `true` the configured clusters will be ignored and the clusters will be discovered.
- If set to `false` the services will be discovered only in configured clusters.

### `clusters`

_Optional, Default=["default"]_

```toml tab="File (TOML)"
[providers.ecs]
  cluster = ["default"]
  # ...
```

```yaml tab="File (YAML)"
providers:
  ecs:
    clusters:
      - default
    # ...
```

```bash tab="CLI"
--providers.ecs.clusters=default
# ...
```

Search for services in clusters list.

### `exposedByDefault`

_Optional, Default=true_

```toml tab="File (TOML)"
[providers.ecs]
  exposedByDefault = false
  # ...
```

```yaml tab="File (YAML)"
providers:
  ecs:
    exposedByDefault: false
    # ...
```

```bash tab="CLI"
--providers.ecs.exposedByDefault=false
# ...
```

Expose ECS services by default in Traefik.
If set to false, services that don't have a `traefik.enable=true` label will be ignored from the resulting routing configuration.

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

```toml tab="File (TOML)"
[providers.ecs]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```yaml tab="File (YAML)"
providers:
  ecs:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```bash tab="CLI"
--providers.ecs.defaultRule=Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)
# ...
```

For a given container if no routing rule was defined by a label, it is defined by this defaultRule instead.
It must be a valid [Go template](https://golang.org/pkg/text/template/),
augmented with the [sprig template functions](http://masterminds.github.io/sprig/).
The service name can be accessed as the `Name` identifier,
and the template has access to all the labels defined on this container.

### `refreshSeconds`

_Optional, Default=15_

```toml tab="File (TOML)"
[providers.ecs]
  refreshSeconds = 15
  # ...
```

```yaml tab="File (YAML)"
providers:
  ecs:
    refreshSeconds: 15
    # ...
```

```bash tab="CLI"
--providers.ecs.refreshSeconds=15
# ...
```

Polling interval (in seconds).

### Credentials

_Optional_

```toml tab="File (TOML)"
[providers.ecs]
  region = "us-east-1"
  accessKeyID = "abc"
  secretAccessKey = "123"
```

```yaml tab="File (YAML)"
providers:
  ecs:
    region: us-east-1
    accessKeyID: "abc"
    secretAccessKey: "123"
    # ...
```

```bash tab="CLI"
--providers.ecs.region="us-east-1"
--providers.ecs.accessKeyID="abc"
--providers.ecs.secretAccessKey="123"
# ...
```

If `region` is not provided, it will be resolved from the EC2 metadata endpoint for EC2 tasks. 
In a FARGATE context it will be resolved from the `AWS_REGION` env variable.

If `accessKeyID` / `secretAccessKey` are not provided, credentials will be resolved in the following order:

- From environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN`.
- Shared credentials, determined by `AWS_PROFILE` and `AWS_SHARED_CREDENTIALS_FILE`, defaults to default and `~/.aws/credentials`.
- EC2 instance role or ECS task role
