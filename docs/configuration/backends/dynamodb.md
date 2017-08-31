# DynamoDB Backend

Træfik can be configured to use Amazon DynamoDB as a backend configuration:

```toml
################################################################
# DynamoDB configuration backend
################################################################

# Enable DynamoDB configuration backend
[dynamodb]

# DyanmoDB Table Name
#
# Optional
#
TableName = "traefik"

# Enable watch DynamoDB changes
#
# Optional
#
Watch = true

# Polling interval (in seconds)
#
# Optional
#
RefreshSeconds = 15

# Region to use when connecting to AWS
#
# Required
#
Region = "us-west-1"

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

# Endpoint of local dynamodb instance for testing
#
# Optional
#
Endpoint = "http://localhost:8080"
```

Items in the `dynamodb` table must have three attributes: 

- `id` (string): The id is the primary key.
- `name`(string): The name is used as the name of the frontend or backend.
- `frontend` or `backend` (map): This attribute's structure matches exactly the structure of a Frontend or Backend type in traefik.
    See `types/types.go` for details.
    The presence or absence of this attribute determines its type.
    So an item should never have both a `frontend` and a `backend` attribute.
