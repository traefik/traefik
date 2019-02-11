# ErrorPage

It Has Never Been Easier to Say That Something Went Wrong
{: .subtitle }

![ErrorPages](../assets/img/middleware/errorpages.png)

The ErrorPage middleware returns a custom page in lieu of the default, according to configured ranges of HTTP Status codes.

!!! important
    The error page itself is _not_ hosted by Traefik.

## Configuration Examples

??? example "File -- Custom Error Page for 5XX"

    ```toml
    [Routers]
      [Routers.router1]
        Service = "my-service"
        Rule = Host(`my-domain`)

    [Middlewares]
      [Middlewares.5XX-errors.Errors]
        status = ["500-599"]
        service = "error-handler-service"
        query = "/error.html"
                
    [Services]
      # ... definition of error-handler-service and my-service
    ```

??? example "Docker -- Dynamic Custom Error Page for 5XX Status Code"

    ```yaml
    a-container:
      image: a-container-image 
        labels:
          - "traefik.middlewares.test-errorpage.errors.status=500-599",
          - "traefik.middlewares.test-errorpage.errors.service=serviceError",
          - "traefik.middlewares.test-errorpage.errors.query=/{status}.html",
            		
    ```
    
    !!! note 
        In this example, the error page URL is based on the status code (`query=/{status}.html)`.

## Configuration Options

### status

The `status` that will trigger the error page.

The status code ranges are inclusive (`500-599` will trigger with every code between `500` and `599`, `500` and `599` included).
 
!!! Note

    You can define either a status code like `500` or ranges with a syntax like `500-599`.

### service

The service that will serve the new requested error page.

### query

The URL for the error page (hosted by `service`). You can use `{status}` in the query, that will be replaced by the received status code.
