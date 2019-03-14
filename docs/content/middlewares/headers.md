# Headers 

Adding Headers to the Request / Response
{: .subtitle }

![Headers](../assets/img/middleware/headers.png)

The Headers middleware can manage the requests/responses headers.

## Configuration Examples

### Adding Headers to the Request and the Response

Add the `X-Script-Name` header to the proxied request and the `X-Custom-Response-Header` to the response

??? example "File"

    ```toml
    [http.middlewares]
      [http.middlewares.testHeader.headers]
        [http.middlewares.testHeader.headers.CustomRequestHeaders]
            X-Script-Name = "test"
        [http.middlewares.testHeader.headers.CustomResponseHeaders]
            X-Custom-Response-Header = "True"
    ```

??? example "Docker"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.http.middlewares.testHeader.Headers.CustomRequestHeaders.X-Script-Name=test",
              - "traefik.http.middlewares.testHeader.Headers.CustomResponseHeaders.X-Custom-Response-Header=True",
    ```

### Adding and Removing Headers

`X-Script-Name` header added to the proxied request, the `X-Custom-Request-Header` header removed from the request, and the `X-Custom-Response-Header` header removed from the response.

??? example "File"

    ```toml    
    [http.middlewares]
      [http.middlewares.testHeader.headers]
        [http.middlewares.testHeader.headers.CustomRequestHeaders]
            X-Script-Name = "test"
        [http.middlewares.testHeader.headers.CustomResponseHeaders]
            X-Custom-Response-Header = "True"
    ```

??? example "Docker"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.http.middlewares.testHeader.Headers.CustomRequestHeaders.X-Script-Name=test",
              - "traefik.http.middlewares.testHeader.Headers.CustomResponseHeaders.X-Custom-Response-Header=True",
    ```

### Using Security Headers

Security related headers (HSTS headers, SSL redirection, Browser XSS filter, etc) can be added and configured per frontend in a similar manner to the custom headers above.
This functionality allows for some easy security features to quickly be set.

??? example "File"

    ```toml    
    [http.middlewares]
      [http.middlewares.testHeader.headers]
        FrameDeny = true
        SSLRedirect = true
    ```

??? example "Docker"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.http.middlewares.testHeader.Headers.FrameDeny=true",
              - "traefik.http.middlewares.testHeader.Headers.SSLRedirect=true",
    ```
       
## Configuration Options

### General

!!! warning
    If the custom header name is the same as one header name of the request or response, it will be replaced.

!!! note
    The detailed documentation for the security headers can be found in [unrolled/secure](https://github.com/unrolled/secure#available-options).

### customRequestHeaders

The `customRequestHeaders` option lists the Header names and values to apply to the request.

### allowedHosts 

The `allowedHosts` option lists fully qualified domain names that are allowed.

### hostsProxyHeaders 

The `hostsProxyHeaders` option is a set of header keys that may hold a proxied hostname value for the request.

### sslRedirect 

The `sslRedirect` is set to true, then only allow https requests.

### sslTemporaryRedirect
                    
Set the `sslTemporaryRedirect` to `true` to force an SSL redirection using a 302 (instead of a 301).

### sslHost 

The `SSLHost` option is the host name that is used to redirect http requests to https.

### sslProxyHeaders 

The `sslProxyHeaders` option is set of header keys with associated values that would indicate a valid https request. Useful when using other proxies with header like: `"X-Forwarded-Proto": "https"`.

### sslForceHost 

Set `sslForceHost` to true and set SSLHost to forced requests to use `SSLHost` even the ones that are already using SSL.

### stsSeconds 

The `stsSeconds` is the max-age of the Strict-Transport-Security header. If set to 0, would NOT include the header.

### stsIncludeSubdomains 

The `stsIncludeSubdomains` is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header.

### stsPreload 

Set `STSPreload` to true to have the `preload` flag appended to the Strict-Transport-Security header.

### forceSTSHeader

Set `ForceSTSHeader` to true, to add the STS header even when the connection is HTTP.

### frameDeny 

Set `frameDeny` to true to add the `X-Frame-Options` header with the value of `DENY`.
 
### customFrameOptionsValue 

The `customFrameOptionsValue` allows the `X-Frame-Options` header value to be set with a custom value. This overrides the FrameDeny option.

### contentTypeNosniff

Set `contentTypeNosniff` to true to add the `X-Content-Type-Options` header with the value `nosniff`.

### browserXssFilter

Set `BrowserXssFilter` to true to add the `X-XSS-Protection` header with the value `1; mode=block`.

### customBrowserXSSValue

The `customBrowserXssValue` option allows the `X-XSS-Protection` header value to be set with a custom value. This overrides the BrowserXssFilter option.

### contentSecurityPolicy

The `contentSecurityPolicy` option allows the `Content-Security-Policy` header value to be set with a custom value.

### publicKey

The `publicKey` implements HPKP to prevent MITM attacks with forged certificates. 

### referrerPolicy

The `referrerPolicy` allows sites to control when browsers will pass the Referer header to other sites.

### isDevelopment

Set `isDevelopment` to true when developing. The AllowedHosts, SSL, and STS options can cause some unwanted effects. Usually testing happens on http, not https, and on localhost, not your production domain.  
If you would like your development environment to mimic production with complete Host blocking, SSL redirects, and STS headers, leave this as false.
