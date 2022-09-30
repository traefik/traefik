---
title: "Traefik AWS ECS Documentation"
description: "Configuration discovery in Traefik is achieved through Providers. Read the technical documentation for leveraging AWS ECS in Traefik."
---

# Traefik & AWS ECS

A Story of Labels & Elastic Containers
{: .subtitle }

Attach labels to your ECS containers and let Traefik do the rest!

## Configuration Examples

??? example "Configuring ECS provider"

    Enabling the ECS provider:

    ```yaml tab="File (YAML)"
    providers:
      ecs: {}
    ```

    ```toml tab="File (TOML)"
    [providers.ecs]
    ```

    ```bash tab="CLI"
    --providers.ecs=true
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
                "ec2:DescribeInstances",
                "ssm:DescribeInstanceInformation"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

!!! info "ECS Anywhere"

    Please note that the `ssm:DescribeInstanceInformation` action is required for ECS anywhere instances discovery.

## Provider Configuration

### `autoDiscoverClusters`

_Optional, Default=false_

Search for services in cluster list.

- If set to `true` service discovery is enabled for all clusters.
- If set to `false` service discovery is enabled on configured clusters only.

```yaml tab="File (YAML)"
providers:
  ecs:
    autoDiscoverClusters: true
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  autoDiscoverClusters = true
  # ...
```

```bash tab="CLI"
--providers.ecs.autoDiscoverClusters=true
# ...
```

### `ecsAnywhere`

_Optional, Default=false_

Enable ECS Anywhere support.

- If set to `true` service discovery is enabled for ECS Anywhere instances.
- If set to `false` service discovery is disabled for ECS Anywhere instances.

```yaml tab="File (YAML)"
providers:
  ecs:
    ecsAnywhere: true
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  ecsAnywhere = true
  # ...
```

```bash tab="CLI"
--providers.ecs.ecsAnywhere=true
# ...
```

### `clusters`

_Optional, Default=["default"]_

Search for services in cluster list.
This option is ignored if `autoDiscoverClusters` is set to `true`.

```yaml tab="File (YAML)"
providers:
  ecs:
    clusters:
      - default
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  clusters = ["default"]
  # ...
```

```bash tab="CLI"
--providers.ecs.clusters=default
# ...
```

### `exposedByDefault`

_Optional, Default=true_

Expose ECS services by default in Traefik.

If set to `false`, services that do not have a `traefik.enable=true` label are ignored from the resulting routing configuration.

```yaml tab="File (YAML)"
providers:
  ecs:
    exposedByDefault: false
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  exposedByDefault = false
  # ...
```

```bash tab="CLI"
--providers.ecs.exposedByDefault=false
# ...
```

### `constraints`

_Optional, Default=""_

The `constraints` option can be set to an expression that Traefik matches against the container labels (task),
to determine whether to create any route for that container. 
If none of the container labels match the expression, no route for that container is created. 
If the expression is empty, all detected containers are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegex("key", "value")` functions,
as well as the usual boolean logic, as shown in examples below.

??? example "Constraints Expression Examples"

    ```toml
    # Includes only containers having a label with key `a.label.name` and value `foo`
    constraints = "Label(`a.label.name`, `foo`)"
    ```

    ```toml
    # Excludes containers having any label with key `a.label.name` and value `foo`
    constraints = "!Label(`a.label.name`, `value`)"
    ```

    ```toml
    # With logical AND.
    constraints = "Label(`a.label.name`, `valueA`) && Label(`another.label.name`, `valueB`)"
    ```

    ```toml
    # With logical OR.
    constraints = "Label(`a.label.name`, `valueA`) || Label(`another.label.name`, `valueB`)"
    ```

    ```toml
    # With logical AND and OR, with precedence set by parentheses.
    constraints = "Label(`a.label.name`, `valueA`) && (Label(`another.label.name`, `valueB`) || Label(`yet.another.label.name`, `valueC`))"
    ```

    ```toml
    # Includes only containers having a label with key `a.label.name` and a value matching the `a.+` regular expression.
    constraints = "LabelRegex(`a.label.name`, `a.+`)"
    ```

For additional information, refer to [Restrict the Scope of Service Discovery](./overview.md#restrict-the-scope-of-service-discovery).

```yaml tab="File (YAML)"
providers:
  ecs:
    constraints: "Label(`a.label.name`,`foo`)"
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  constraints = "Label(`a.label.name`,`foo`)"
  # ...
```

```bash tab="CLI"
--providers.ecs.constraints=Label(`a.label.name`,`foo`)
# ...
```

### `defaultRule`

_Optional, Default=```Host(`{{ normalize .Name }}`)```_

The `defaultRule` option defines what routing rule to apply to a container if no rule is defined by a label.

It must be a valid [Go template](https://pkg.go.dev/text/template/), and can use
[sprig template functions](https://masterminds.github.io/sprig/).
The container service name can be accessed with the `Name` identifier,
and the template has access to all the labels defined on this container.

```yaml tab="File (YAML)"
providers:
  ecs:
    defaultRule: "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  defaultRule = "Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)"
  # ...
```

```bash tab="CLI"
--providers.ecs.defaultRule=Host(`{{ .Name }}.{{ index .Labels \"customLabel\"}}`)
# ...
```

### `refreshSeconds`

_Optional, Default=15_

Polling interval (in seconds).

```yaml tab="File (YAML)"
providers:
  ecs:
    refreshSeconds: 15
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  refreshSeconds = 15
  # ...
```

```bash tab="CLI"
--providers.ecs.refreshSeconds=15
# ...
```

### Credentials

_Optional_

If `region` is not provided, it is resolved from the EC2 metadata endpoint for EC2 tasks.
In a FARGATE context it is resolved from the `AWS_REGION` environment variable.

If `accessKeyID` and `secretAccessKey` are not provided, credentials are resolved in the following order:

- Using the environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN`.
- Using shared credentials, determined by `AWS_PROFILE` and `AWS_SHARED_CREDENTIALS_FILE`, defaults to `default` and `~/.aws/credentials`.
- Using EC2 instance role or ECS task role

```yaml tab="File (YAML)"
providers:
  ecs:
    region: us-east-1
    accessKeyID: "abc"
    secretAccessKey: "123"
    # ...
```

```toml tab="File (TOML)"
[providers.ecs]
  region = "us-east-1"
  accessKeyID = "abc"
  secretAccessKey = "123"
```

```bash tab="CLI"
--providers.ecs.region="us-east-1"
--providers.ecs.accessKeyID="abc"
--providers.ecs.secretAccessKey="123"
# ...
```
