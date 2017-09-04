# ECS Backend

Træfik can be configured to use Amazon ECS as a backend configuration:

```toml
################################################################
# ECS configuration backend
################################################################

# Enable ECS configuration backend
[ecs]

# ECS Cluster Name
#
# DEPRECATED - Please use Clusters
#
Cluster = "default"

# ECS Clusters Name
#
# Optional
# Default: ["default"]
#
Clusters = ["default"]

# Enable watch ECS changes
#
# Optional
# Default: true
#
Watch = true

# Enable auto discover ECS clusters
#
# Optional
# Default: false
#
AutoDiscoverClusters = false

# Polling interval (in seconds)
#
# Optional
# Default: 15
#
RefreshSeconds = 15

# Expose ECS services by default in traefik
#
# Optional
# Default: true
#
ExposedByDefault = false

# Region to use when connecting to AWS
#
# Optional
#
Region = "us-east-1"

# AccessKeyID to use when connecting to AWS
#
# Optional
#
AccessKeyID = "abc"

# SecretAccessKey to use when connecting to AWS
#
# Optional
#
SecretAccessKey = "123"

# Override default configuration template. For advanced users :)
#
# Optional
#
# filename = "ecs.tmpl"
```

Labels can be used on task containers to override default behaviour:

| Label                                        | Description                                                                              |
|----------------------------------------------|------------------------------------------------------------------------------------------|
| `traefik.protocol=https`                     | override the default `http` protocol                                                     |
| `traefik.weight=10`                          | assign this weight to the container                                                      |
| `traefik.enable=false`                       | disable this container in Træfik                                                         |
| `traefik.backend.loadbalancer.method=drr`    | override the default `wrr` load balancer algorithm                                       |
| `traefik.backend.loadbalancer.sticky=true`   | enable backend sticky sessions                                                           |
| `traefik.frontend.rule=Host:test.traefik.io` | override the default frontend rule (Default: `Host:{containerName}.{domain}`).           |
| `traefik.frontend.passHostHeader=true`       | forward client `Host` header to the backend.                                             |
| `traefik.frontend.priority=10`               | override default frontend priority                                                       |
| `traefik.frontend.entryPoints=http,https`    | assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`. |

If `AccessKeyID`/`SecretAccessKey` is not given credentials will be resolved in the following order:

- From environment variables; `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN`.
- Shared credentials, determined by `AWS_PROFILE` and `AWS_SHARED_CREDENTIALS_FILE`, defaults to `default` and `~/.aws/credentials`.
- EC2 instance role or ECS task role

Træfik needs the following policy to read ECS information:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "Traefik ECS read access",
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
