---
title: "Traefik DigestAuth Documentation"
description: "Traefik Proxy's HTTP DigestAuth middleware restricts access to your services to known users. Read the technical documentation."
---

![DigestAuth](../../../../assets/img/middleware/digestauth.png)

The `DigestAuth` middleware grants access to services to authorized users only.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Declaring the user list
http:
  middlewares:
    test-auth:
      digestAuth:
        users:
          - "test:traefik:a2688e031edb4be6a3797f3882655c05"
          - "test2:traefik:518845800f9e2bfb1f1f740ec24f074e"
```

```toml tab="Structured (TOML)"
# Declaring the user list
[http.middlewares]
  [http.middlewares.test-auth.digestAuth]
    users = [
      "test:traefik:a2688e031edb4be6a3797f3882655c05",
      "test2:traefik:518845800f9e2bfb1f1f740ec24f074e",
    ]
```

```yaml tab="Labels"
# Declaring the user list
labels:
  - "traefik.http.middlewares.test-auth.digestauth.users=test:traefik:a2688e031edb4be6a3797f3882655c05,test2:traefik:518845800f9e2bfb1f1f740ec24f074e"
```

```json tab="Tags"
// Declaring the user list
{
  //...
  "Tags" : [
    "traefik.http.middlewares.test-auth.digestauth.users=test:traefik:a2688e031edb4be6a3797f3882655c05,test2:traefik:518845800f9e2bfb1f1f740ec24f074e"
  ]
}
```

```yaml tab="Kubernetes"
# Declaring the user list
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  digestAuth:
    secret: userssecret
```

## Configuration Options

| Field      | Description    | Default | Required |
|:-----------|:---------------------------------------------------------------------------------|:--------|:---------|
| `users` | Array of authorized users. Each user must be declared using the `name:realm:encoded-password` format.<br /> The option `users` supports Kubernetes secrets.<br />(More information [here](#users--usersfile))| []  | No      |
| `usersFile` | Path to an external file that contains the authorized users for the middleware. <br />The file content is a list of `name:realm:encoded-password`. (More information [here](#users--usersfile)) | ""      | No      |
| `realm` | Allow customizing the realm for the authentication.| "traefik"      | No      |
| `headerField` | Allow defining a header field to store the authenticated user.| ""      | No      |
| `removeHeader` | Allow removing the authorization header before forwarding the request to your service. | false      | No      |

### Passwords format

Passwords must be hashed using MD5, SHA1, or BCrypt.
Use `htpasswd` to generate the passwords.

### users & usersFile

- If both `users` and `usersFile` are provided, they are merged. The contents of `usersFile` have precedence over the values in users.
- Because referencing a file path isn’t feasible on Kubernetes, the `users` & `usersFile` field isn’t used in Kubernetes IngressRoute. Instead, use the `secret` field.

### Kubernetes Secrets

On Kubernetes, you don’t use the `users` or `usersFile` fields. Instead, you reference a Kubernetes secret using the `secret` field in your Middleware resource. This secret can be one of two types:

- `kubernetes.io/basic-auth secret`: This secret type contains two keys—`username` and `password`—but is generally suited for a smaller number of users. Please note that these keys are not hashed or encrypted in any way, and therefore is less secure than the other method.
- Opaque secret with a users field: Here, the secret contains a single string field (often called `users`) where each line represents a user. This approach allows you to store multiple users in one secret.

{!traefik-for-business-applications.md!}
