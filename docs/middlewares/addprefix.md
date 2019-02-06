# Add Prefix

Prefixing the Path 
{: .subtitle }

![AddPrefix](../img/middleware/addprefix.png) 

The AddPrefix middleware updates the URL Path of the request before forwarding it.

## Configuration Examples

??? example "File -- Prefixing with /foo"

    ```toml
    [Middlewares]
      [Middlewares.add-foo.AddPrefix]
         prefix = "/foo"
    ```

??? example "Docker -- Prefixing with /bar"

    ```yaml
    a-container:
      image: a-container-image 
        labels:
          - "traefik.middlewares.add-bar.addprefix.prefix=/bar"
    ```
## Configuration Options

### prefix

`prefix` is the string to add before the current path in the requested URL. It should include the leading slash (`/`).
