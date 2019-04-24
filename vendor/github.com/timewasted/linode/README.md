linode
===

`linode` is a library for interacting with the [Linode API](https://www.linode.com/api).

Support for the API is as follows:

- [DNS](https://www.linode.com/api/dns): Supported.
- Raw interaction with the API: Supported.
- Batch requests: Partially supported.

## Usage

This section definitely needs to be fleshed out more in the future, but until
that point, here is a simple example:

```go
import (
        "log"

        "github.com/timewasted/linode/dns"
)

linode := dns.New("<insert api key here>")
domain, err := linode.GetDomain("example.com")
if err != nil {
        log.Fatalln(err)
}
resource, err := linode.CreateDomainResourceTXT(domain.DomainID, "_acme-challenge", "super secret value", 60)
if err != nil {
        log.Fatalln(err)
}
// Do stuff with resource
```
