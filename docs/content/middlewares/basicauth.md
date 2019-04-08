# BasicAuth

Adding Basic Authentication
{: .subtitle }

![BasicAuth](../assets/img/middleware/basicauth.png)

The BasicAuth middleware is a quick way to restrict access to your services to known users.

## Configuration Examples

```yaml tab="Docker"
# Declaring the user list
labels:
  - "traefik.http.middlewares.test-auth.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
```

```yaml tab="Kubernetes"
# Declaring the user list
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  basicAuth:
    users:
    - test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/
    - test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0
```

```yaml tab="Rancher"
# Declaring the user list
labels:
  - "traefik.http.middlewares.test-auth.basicauth.users=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"
```

```toml tab="File"
# Declaring the user list
[http.middlewares]
  [http.middlewares.test-auth.basicauth]
  users = [
    "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", 
    "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
  ]
```

## Configuration Options

### General

Passwords must be encoded using MD5, SHA1, or BCrypt.

!!! tip 
   
    Use `htpasswd` to generate the passwords.

### `users`

The `users` option is an array of authorized users. Each user will be declared using the `name:encoded-password` format.

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### `usersFile`

The `usersFile` option is the path to an external file that contains the authorized users for the middleware.

The file content is a list of `name:encoded-password`.

??? example "A file containing test/test and test2/test2"

    ```
    test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/
    test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0
    ```

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### `realm`

You can customize the realm for the authentication with the `realm` option. The default value is `traefik`. 

### `headerField`

You can customize the header field for the authenticated user using the `headerField`option.

```yaml tab="Docker"
labels:
  - "traefik.http.middlewares.my-auth.basicauth.headerField=X-WebAuth-User"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: my-auth
spec:
  basicAuth:
    # ...
    headerField: X-WebAuth-User
```

```toml tab="File"
[http.middlewares.my-auth.basicauth]
  # ...
  headerField = "X-WebAuth-User"
```

### `removeHeader`

Set the `removeHeader` option to `true` to remove the authorization header before forwarding the request to your service. (Default value is `false`.)
