# DigestAuth

Adding Digest Authentication
{: .subtitle } 

![BasicAuth](../assets/img/middleware/digestauth.png)

The DigestAuth middleware is a quick way to restrict access to your services to known users.

## Configuration Examples

??? example "File -- Declaring the user list"

    ```toml
    [Middlewares]
      [Middlewares.test-auth.digestauth]
      users = ["test:traefik:a2688e031edb4be6a3797f3882655c05", "test2:traefik:518845800f9e2bfb1f1f740ec24f074e"]
    ```

??? example "Docker -- Using an external file for the authorized users"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.middlewares.declared-users-only.digestauth.usersFile=path-to-file.ext",
    ```

!!! tip 
   
    Use `htdigest` to generate passwords.

## Configuration Options

### Users

The `users` option is an array of authorized users. Each user will be declared using the `name:realm:encoded-password` format.

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### UsersFile

The `usersFile` option is the path to an external file that contains the authorized users for the middleware.

The file content is a list of `name:realm:encoded-password`.

??? example "A file containing test/test and test2/test2"

    ```
    test:traefik:a2688e031edb4be6a3797f3882655c05
    test2:traefik:518845800f9e2bfb1f1f740ec24f074e
    ```

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### Realm

You can customize the realm for the authentication with the `realm` option. The default value is `traefik`. 

### HeaderField

You can customize the header field for the authenticated user using the `headerField`option.

??? example "File -- Passing Authenticated Users to Services Via Headers"

    ```toml
      [Middlewares.my-auth.digestauth]
        usersFile = "path-to-file.ext"
        headerField = "X-WebAuth-User" # header for the authenticated user
    ```

### RemoveHeader

Set the `removeHeader` option to `true` to remove the authorization header before forwarding the request to your service. (Default value is `false`.)
