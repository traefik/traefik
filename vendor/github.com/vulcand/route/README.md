Route
=====

```go
Host("localhost") && Method("POST") && Path("/v1")
Host("localhost") && Method("POST") && Path("/v1") && Header("Content-Type", "application/<string>")
```

HTTP request routing language and library.

Features:

* Trie based matching
* Regexp based matching
* Matches hosts, headers, methods and paths
* Flexible matching language

Documentation:

http://godoc.org/github.com/vulcand/route
