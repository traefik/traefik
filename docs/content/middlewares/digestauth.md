# DigestAuth

Adding Digest Authentication
{: .subtitle } 

![BasicAuth](../assets/img/middleware/digestauth.png)

The DigestAuth middleware is a quick way to restrict access to your services to known users.

## Configuration Examples

```yaml tab="Docker"
labels:
- "traefik.http.middlewares.test-auth.digestauth.users=test:traefik:a2688e031edb4be6a3797f3882655c05,test2:traefik:518845800f9e2bfb1f1f740ec24f074e"
```

```yaml tab="Kubernetes"
# Declaring the user list
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: test-auth
spec:
  digestAuth:
    users:
    - test:traefik:a2688e031edb4be6a3797f3882655c05
    - test2:traefik:518845800f9e2bfb1f1f740ec24f074e
```

```yaml tab="Rancher"
labels:
- "traefik.http.middlewares.test-auth.digestauth.users=test:traefik:a2688e031edb4be6a3797f3882655c05,test2:traefik:518845800f9e2bfb1f1f740ec24f074e"
```

```toml tab="File"
[http.middlewares]
  [http.middlewares.test-auth.digestAuth]
    users = [
      "test:traefik:a2688e031edb4be6a3797f3882655c05",
      "test2:traefik:518845800f9e2bfb1f1f740ec24f074e",
    ]
```

!!! tip 
   
    Use `htdigest` to generate passwords.

## Configuration Options

### `Users`

The `users` option is an array of authorized users. Each user will be declared using the `name:realm:encoded-password` format.

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### `UsersFile`

The `usersFile` option is the path to an external file that contains the authorized users for the middleware.

The file content is a list of `name:realm:encoded-password`.

??? example "A file containing test/test and test2/test2"

    ```
    test:traefik:a2688e031edb4be6a3797f3882655c05
    test2:traefik:518845800f9e2bfb1f1f740ec24f074e
    ```

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### `Realm`

You can customize the realm for the authentication with the `realm` option. The default value is `traefik`. 

### `HeaderField`

You can customize the header field for the authenticated user using the `headerField`option.

Example "File -- Passing Authenticated User to Services Via Headers"

```yaml tab="Docker"
labels:
  - "traefik.http.middlewares.my-auth.digestauth.headerField=X-WebAuth-User"
```

```yaml tab="Kubernetes"
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: my-auth
spec:
  digestAuth:
    # ...
    headerField: X-WebAuth-User
```

```yaml tab="Rancher"
labels:
  - "traefik.http.middlewares.my-auth.digestauth.headerField=X-WebAuth-User"
```

```toml tab="File"
[http.middlewares.my-auth.digestAuth]
  # ...
  headerField = "X-WebAuth-User"
```

### `RemoveHeader`

Set the `removeHeader` option to `true` to remove the authorization header before forwarding the request to your service. (Default value is `false`.)
