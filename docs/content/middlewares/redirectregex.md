# TODO - RedirectRegex

Redirecting the Client to a Different Location
{: .subtitle }

`TODO: add schema`

RegexRedirect redirect a request from an url to another with regex matching and replacement.

## Configuration Examples

??? example "File -- Redirect with domain replacement"

    ```toml
    [http.middlewares]
      [http.middlewares.test-redirectregex.redirectregex]
        regex = "^http://localhost/(.*)"
        replacement = "http://mydomain/$1"
    ```

??? example "Docker -- Redirect with domain replacement"

    ```yml
     a-container:
        image: a-container-image 
            labels:
                - "traefik.http.middlewares.test-redirectregex.redirectregex.regex=^http://localhost/(.*)"
                - "traefik.http.middlewares.test-redirectregex.redirectregex.replacement=http://mydomain/$1"
    ```

## Configuration Options

### permanent

Set the `permanent` option to `true` to apply a permanent redirection.

### regex

The `Regex` option is the regular expression to match and capture elements form the request URL.

!!! warning

    Care should be taken when defining replacement expand variables: `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax.

!!! tip

    Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).
    
### replacement

The `replacement` option defines how to modify the URl to have the new target URL.
 