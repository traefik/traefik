# BasicAuth

Adding Basic Authentication
{: .subtitle }

![BasicAuth](../assets/img/middleware/basicauth.png)

The BasicAuth middleware is a quick way to restrict access to your services to known users.

## Configuration Examples

??? example "File -- Declaring the user list"

    ```toml
    [Middlewares]
      [Middlewares.test-auth.basicauth]
      users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", 
      "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"]
    ```

??? example "Docker -- Using an external file for the authorized users"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.middlewares.declared-users-only.basicauth.usersFile=path-to-file.ext",
    ```

## Configuration Options

### General

Passwords must be encoded using MD5, SHA1, or BCrypt.

!!! tip 
   
    Use `htpasswd` to generate the passwords.

### users

The `users` option is an array of authorized users. Each user will be declared using the `name:encoded-password` format.

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### usersFile

The `usersFile` option is the path to an external file that contains the authorized users for the middleware.

The file content is a list of `name:encoded-password`.

??? example "A file containing test/test and test2/test2"

    ```
    test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/
    test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0
    ```

!!! Note
    
    If both `users` and `usersFile` are provided, the two are merged. The content of `usersFile` has precedence over `users`.

### realm

You can customize the realm for the authentication with the `realm` option. The default value is `traefik`. 

### headerField

You can customize the header field for the authenticated user using the `headerField`option.

??? example "File -- Passing Authenticated Users to Services Via Headers"

    ```toml
      [Middlewares.my-auth.basicauth]
        usersFile = "path-to-file.ext"
        headerField = "X-WebAuth-User" # header for the authenticated user
    ```

### removeHeader

Set the `removeHeader` option to `true` to remove the authorization header before forwarding the request to your service. (Default value is `false`.)
