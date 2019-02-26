# DynamoDB Provider

Traefik can be configured to use Amazon DynamoDB as a provider.

## Configuration

```toml
################################################################
# DynamoDB Provider
################################################################

# Enable DynamoDB Provider.
[dynamodb]

# Region to use when connecting to AWS.
#
# Required
#
region = "us-west-1"

# DyanmoDB Table Name.
#
# Optional
# Default: "traefik"
#
tableName = "traefik"

# Enable watch DynamoDB changes.
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

# Access Key ID to use when connecting to AWS.
#
# Optional
#
accessKeyID = "abc"

# Secret Access Key to use when connecting to AWS.
#
# Optional
#
secretAccessKey = "123"

# Endpoint of local dynamodb instance for testing?
#
# Optional
#
endpoint = "http://localhost:8080"
```

## Table Items

Items in the `dynamodb` table must have three attributes:

- `id` (string): The id is the primary key.
- `name`(string): The name is used as the name of the frontend or backend.
- `frontend` or `backend` (map): This attribute's structure matches exactly the structure of a Frontend or Backend type in Traefik.  
    See `types/types.go` for details.  
    The presence or absence of this attribute determines its type.
    So an item should never have both a `frontend` and a `backend` attribute.
