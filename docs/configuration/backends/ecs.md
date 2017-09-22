# ECS Backend

Træfik can be configured to use Amazon ECS as a backend configuration.

## Configuration

```toml
################################################################
# ECS configuration backend
################################################################

# Enable ECS configuration backend.
[ecs]

# ECS Cluster Name.
#
# DEPRECATED - Please use `clusters`.
#
cluster = "default"

# ECS Clusters Name.
#
# Optional
# Default: ["default"]
#
clusters = ["default"]

# Enable watch ECS changes.
#
# Optional
# Default: true
#
watch = true

# Default domain used.
#
# Optional
# Default: ""
#
domain = "ecs.localhost"

# Enable auto discover ECS clusters.
#
# Optional
# Default: false
#
autoDiscoverClusters = false

# Polling interval (in seconds).
#
# Optional
# Default: 15
#
refreshSeconds = 15

# Expose ECS services by default in Traefik.
#
# Optional
# Default: true
#
exposedByDefault = false

# Region to use when connecting to AWS.
#
# Optional
#
region = "us-east-1"

# AccessKeyID to use when connecting to AWS.
#
# Optional
#
accessKeyID = "abc"

# SecretAccessKey to use when connecting to AWS.
#
# Optional
#
secretAccessKey = "123"

# Override default configuration template.
# For advanced users :)
#
# Optional
#
# filename = "ecs.tmpl"
```

If `AccessKeyID`/`SecretAccessKey` is not given credentials will be resolved in the following order:

- From environment variables; `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN`.
- Shared credentials, determined by `AWS_PROFILE` and `AWS_SHARED_CREDENTIALS_FILE`, defaults to `default` and `~/.aws/credentials`.
- EC2 instance role or ECS task role

## Policy

Træfik needs the following policy to read ECS information:

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

## Labels: overriding default behaviour

Labels can be used on task containers to override default behaviour:

| Label                                             | Description                                                                              |
|---------------------------------------------------|------------------------------------------------------------------------------------------|
| `traefik.protocol=https`                          | override the default `http` protocol                                                     |
| `traefik.weight=10`                               | assign this weight to the container                                                      |
| `traefik.enable=false`                            | disable this container in Træfik                                                         |
| `traefik.backend.loadbalancer.method=drr`         | override the default `wrr` load balancer algorithm                                       |
| `traefik.backend.loadbalancer.sticky=true`        | enable backend sticky sessions                                                           |
| `traefik.frontend.rule=Host:test.traefik.io`      | override the default frontend rule (Default: `Host:{containerName}.{domain}`).           |
| `traefik.frontend.passHostHeader=true`            | forward client `Host` header to the backend.                                             |
| `traefik.frontend.priority=10`                    | override default frontend priority                                                       |
| `traefik.frontend.entryPoints=http,https`         | assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`. |
| `traefik.frontend.auth.basic=EXPR`                | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`         |
