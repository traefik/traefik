# TODO -- ReplacePath

Updating the Path Before Forwarding the Request
{: .subtitle }

`TODO: add schema`

Replace the path of the request url.

## Configuration Examples

??? example "File -- Replace the path by /foo"

    ```toml
    [http.middlewares]
      [http.middlewares.test-replacepath.ReplacePath]
         path = "/foo"
    ```

??? example "Docker --Replace the path by /foo"

    ```yaml
    a-container:
      image: a-container-image 
        labels:
          - "traefik.http.middlewares.test-replacepath.replacepath.path=/foo"
    ```
    
## Configuration Options

### General

The ReplacePath middleware will:

* replace the actual path by the specified one.
* store the original path in a `X-Replaced-Path` header.

### path

The `path` option defines the path to use as replacement in the request url.
