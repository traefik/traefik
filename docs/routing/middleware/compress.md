# Compress

Compressing the Response before Sending it to the Client
{: .subtitle }

![Compress](../../img/middleware/compress.png)

The Compress middleware enables the gzip compression. 

## Configuration Examples

??? example "File -- enable gzip compression"

    ```toml
    [Middlewares]
      [Middlewares.test-compress.Compress]
    ```
    
??? example "Docker -- enable gzip compression"

    ```yml
    a-container:
          image: a-container-image 
            labels:
              - "traefik.middlewares.test-compress.compress=true",
    ```

## Notes

Responses are compressed when:

* The response body is larger than `512` bytes.
* The `Accept-Encoding` request header contains `gzip`.
* The response is not already compressed, i.e. the `Content-Encoding` response header is not already set.
