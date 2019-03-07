# TODO - RedirectScheme

Redirecting the Client to a Different Scheme/Port
{: .subtitle }

`TODO: add schema`

RegexRedirect redirect request from a scheme to another.

## Configuration Examples

??? example "File -- Redirect to https"

    ```toml
    [http.middlewares]
      [http.middlewares.test-redirectscheme.redirectscheme]
        scheme = "https"
    ```

??? example "Docker -- Redirect to https"

    ```yml
     a-container:
        image: a-container-image 
            labels:
                - "traefik.http.middlewares.test-redirectscheme.redirectscheme.scheme=https"
    ```

## Configuration Options

### permanent

Set the `permanent` option to `true` to apply a permanent redirection.

### scheme

The `scheme` option defines the scheme of the new url.

### port

The `port` option defines the port of the new url.
