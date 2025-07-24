---
title: "Traefik AWS ECS Documentation"
description: "Configuration discovery in Traefik is achieved through Providers. Read the technical documentation for leveraging AWS ECS in Traefik."
---

# Traefik & AWS ECS

## Configuration Example

You can enable the ECS provider with as detailed below:

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

## Configuration Options

| Field | Description                                               | Default              | Required |
|:------|:----------------------------------------------------------|:---------------------|:---------|
| `providers.providersThrottleDuration` | Minimum amount of time to wait for, after a configuration reload, before taking into account any new configuration refresh event.<br />If multiple events occur within this time, only the most recent one is taken into account, and all others are discarded.<br />**This option cannot be set per provider, but the throttling algorithm applies to each of them independently.** | 2s  | No |
| `providers.ecs.autoDiscoverClusters` | Search for services in cluster list. If set to `true` service discovery is enabled for all clusters. |  false  | No   |
| `providers.ecs.ecsAnywhere` | Enable ECS Anywhere support. |  false    | No   |
| `providers.ecs.clusters` | Search for services in cluster list. This option is ignored if `autoDiscoverClusters` is set to `true`. |  `["default"]`  | No   |
| `providers.ecs.exposedByDefault` | Expose ECS services by default in Traefik. | true  | No   |
| `providers.ecs.constraints` |  Defines an expression that Traefik matches against the container labels to determine whether to create any route for that container. See [here](#constraints) for more information.  | true  | No   |
| `providers.ecs.healthyTasksOnly` |  Defines whether Traefik discovers only healthy tasks (`HEALTHY` healthStatus).  | false  | No   |
| `providers.ecs.defaultRule` | The Default Host rule for all services. See [here](#defaultrule) for more information. |   ```"Host(`{{ normalize .Name }}`)"```  | No   |
| `providers.ecs.refreshSeconds` | Defines the polling interval (in seconds).   | 15   | No |
| `providers.ecs.region` | Defines the region of the ECS instance. See [here](#credentials) for more information.  | ""   | No |
| `providers.ecs.accessKeyID` | Defines the Access Key ID for the ECS instance. See [here](#credentials) for more information.  | ""   | No |
| `providers.ecs.secretAccessKey` | Defines the Secret Access Key for the ECS instance. See [here](#credentials) for more information.  | ""   | No |

### `constraints`

The `constraints` option can be set to an expression that Traefik matches against the container labels (task),
to determine whether to create any route for that container. 
If none of the container labels match the expression, no route for that container is created. 
If the expression is empty, all detected containers are included.

The expression syntax is based on the `Label("key", "value")`, and `LabelRegex("key", "value")` functions,
as well as the usual boolean logic, as shown in examples below.

!!! tip "Constraints key limitations"

    Note that `traefik.*` is a reserved label namespace for configuration and can not be used as a key for custom constraints.

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
--providers.ecs.constraints="Label(`a.label.name`,`foo`)"
# ...
```

For additional information, refer to [Restrict the Scope of Service Discovery](../overview.md#restrict-the-scope-of-service-discovery).

### `defaultRule`

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
--providers.ecs.defaultRule='Host(`{{ .Name }}.{{ index .Labels "customLabel"}}`)'
# ...
```

??? info "Default rule and Traefik service"

    The exposure of the Traefik container, combined with the default rule mechanism,
    can lead to create a router targeting itself in a loop.
    In this case, to prevent an infinite loop,
    Traefik adds an internal middleware to refuse the request if it comes from the same router.

### Credentials

This defines the credentials for the ECS instance

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
