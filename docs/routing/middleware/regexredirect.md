# RegexRedirect

Redirecting the Client to a Different Location
{: .subtitle }

# Old Content

## Redirect HTTP to HTTPS

To redirect an http entrypoint to an https entrypoint (with SNI support).

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    entryPoint = "https"
  [entryPoints.https]
  address = ":443"
    [entryPoints.https.tls]
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.com.cert"
      keyFile = "integration/fixtures/https/snitest.com.key"
      [[entryPoints.https.tls.certificates]]
      certFile = "integration/fixtures/https/snitest.org.cert"
      keyFile = "integration/fixtures/https/snitest.org.key"
```

!!! note
    Please note that `regex` and `replacement` do not have to be set in the `redirect` structure if an entrypoint is defined for the redirection (they will not be used in this case).

## Rewriting URL

To redirect an entrypoint rewriting the URL.

```toml
[entryPoints]
  [entryPoints.http]
  address = ":80"
    [entryPoints.http.redirect]
    regex = "^http://localhost/(.*)"
    replacement = "http://mydomain/$1"
```

!!! note
    Please note that `regex` and `replacement` do not have to be set in the `redirect` structure if an `entrypoint` is defined for the redirection (they will not be used in this case).

Care should be taken when defining replacement expand variables: `$1x` is equivalent to `${1x}`, not `${1}x` (see [Regexp.Expand](https://golang.org/pkg/regexp/#Regexp.Expand)), so use `${1}` syntax.

Regular expressions and replacements can be tested using online tools such as [Go Playground](https://play.golang.org/p/mWU9p-wk2ru) or the [Regex101](https://regex101.com/r/58sIgx/2).