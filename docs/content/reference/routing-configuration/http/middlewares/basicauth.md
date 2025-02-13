---
title: "Traefik BasicAuth Documentation"
description: "The HTTP basic authentication (BasicAuth) middleware in Traefik Proxy restricts access to your Services to known users. Read the technical documentation."
---

![BasicAuth](../../../../assets/img/middleware/basicauth.png)

The `basicAuth` middleware grants access to services to authorized users only.

## Configuration Examples

```yaml tab="Structured (YAML)"
# Declaring the user list
http:
  middlewares:
    test-auth:
      basicAuth:
        users:
          - "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"
          - "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
```

```toml tab="Structured (TOML)"
# Declaring the user list
[http.middlewares]
  [http.middlewares.test-auth.basicAuth]
  users = [
    "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
    "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
  ]
```

```yaml tab="Labels"
# Declaring the user list
#
# Note: when used in docker-compose.yml all dollar signs in the hash need to be doubled for escaping.
# To create user:password pair, it's possible to use this command:
# echo $(htpasswd -nB user) | sed -e s/\\$/\\$\\$/g
#
# Also, note that dollar signs should NOT be doubled when not evaluated (e.g. Ansible docker_container module).
labels:
  - "traefik.http.middlewares.test-auth.basicauth.users=test:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/,test2:$$apr1$$d9hr9HBB$$4HxwgUir3HP4EsggP/QNo0"
```

```json tab="Tags"
{
  // ...
  "Tags": [
    "traefik.http.middlewares.test-auth.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
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
  basicAuth:
    secret: secretName
```

## Configuration Options

| Field      | Description                                                                                                                                                                                 | Default | Required |
|:-----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------|:---------|
| `users` | Array of authorized users. Each user must be declared using the `name:hashed-password` format. (More information [here](#users))| ""      | No      |
| `usersFile` | Path to an external file that contains the authorized users for the middleware. <br />The file content is a list of `name:hashed-password`. (More information [here](#usersfile)) | ""      | No      |
| `realm` | Allow customizing the realm for the authentication.| "traefik"      | No      |
| `headerField` | Allow defining a header field to store the authenticated user.| ""      | No      |
| `removeHeader` | Allow removing the authorization header before forwarding the request to your service. | false      | No      |

### Passwords format

Passwords must be hashed using MD5, SHA1, or BCrypt.
Use `htpasswd` to generate the passwords.

### users & usersFile

- If both `users` and `usersFile` are provided, they are merged. The contents of `usersFile` have precedence over the values in users.
- Because referencing a file path isn’t feasible on Kubernetes, the `users` & `usersFile` field isn’t used in Kubernetes IngressRoute. Instead, use the `secret` field.

#### Kubernetes Secrets

The option `users` supports Kubernetes secrets.

!!! note "Kubernetes `kubernetes.io/basic-auth` secret type"

    Kubernetes supports a special `kubernetes.io/basic-auth` secret type.
    This secret must contain two keys: `username` and `password`.

    Please note that these keys are not hashed or encrypted in any way, and therefore is less secure than other methods.
    You can find more information on the [Kubernetes Basic Authentication Secret Documentation](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret)

{!traefik-for-business-applications.md!}
