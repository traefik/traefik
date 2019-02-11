# ForwardAuth

Using an External Service to Ccheck for Credentials
{: .subtitle }

![AuthForward](../img/middleware/authforward.png)

The ForwardAuth middleware delegate the authentication to an external service.
If the service response code is 2XX, access is granted and the original request is performed.
Otherwise, the response from the authentication server is returned.

## Configuration Examples

??? example "File -- Forward authentication to authserver.com"

    ```toml
    [Middlewares]
      [Middlewares.test-auth.forwardauth]
        address = "https://authserver.com/auth"
        trustForwardHeader = true
        authResponseHeaders = ["X-Auth-User", "X-Secret"]

        [Middlewares.test-auth.forwardauth.tls]
          ca = "path/to/local.crt"
          caOptional = true
          cert = "path/to/foo.cert"
          key = "path/to/foo.key"      
    ```


??? example "Docker -- Forward authentication to authserver.com"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.Middlewares.test-auth.ForwardAuth.Address=https://authserver.com/auth"
              - "traefik.Middlewares.test-auth.ForwardAuth.AuthResponseHeaders=X-Auth-User, X-Secret"
              - "traefik.Middlewares.test-auth.ForwardAuth.TLS.CA=path/to/local.crt"
              - "traefik.Middlewares.test-auth.ForwardAuth.TLS.CAOptional=true"
              - "traefik.Middlewares.test-auth.ForwardAuth.TLS.Cert=path/to/foo.cert"
              - "traefik.Middlewares.test-auth.ForwardAuth.TLS.InsecureSkipVerify=true"
              - "traefik.Middlewares.test-auth.ForwardAuth.TLS.Key=path/to/foo.key"
              - "traefik.Middlewares.test-auth.ForwardAuth.TrustForwardHeader=true"
              		
    ```

## Configuration Options

### address

The `address` option defines the authentication server address.

### trustForwardHeader

Set the `trustForwardHeader` option to true to trust all the existing X-Forwarded-* headers.

### authResponseHeaders

The `authResponseHeaders` option is the list of the headers to copy from the authentication server to the request.

### tls

The `tls` option is the tls configuration from Traefik to the authentication server.
