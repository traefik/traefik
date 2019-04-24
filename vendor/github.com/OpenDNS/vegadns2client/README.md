# vegadns2client
vegadns2client is a go client for [VegaDNS-API](https://github.com/shupp/VegaDNS-API).  This is an incomplete client, initially intended to support [lego](https://github.com/xenolf/lego).

## Example Usage

### An example of looking up the auth zone for a hostname:

```go
package main

import (
    "fmt"
    "github.com/opendns/vegadns2client"
)

func main() {
    v := vegadns2client.NewVegaDNSClient("http://localhost:5000")
    v.APIKey = "mykey"
    v.APISecret = "mysecret"

    authZone, domainID, err := v.GetAuthZone("example.com")
    fmt.Println(authZone, domainID, err)
}
```

Which will output the following:

```
2018/02/22 16:11:48 tmpHostname for i = 1: example.com
2018/02/22 16:11:48 {ok 1 [{active example.com 2 0}]}
2018/02/22 16:11:48 Found zone: example.com
	Shortened to foobar.com
foobar.com <nil>
```

### An example of creating and deleting a TXT record

```go
package main

import (
    "fmt"
    "github.com/opendns/vegadns2client"
)

func main() {
    v := vegadns2client.NewVegaDNSClient("http://localhost:5000")
    v.APIKey = "mykey"
    v.APISecret = "mysecret"

    authZone, domainID, err := v.GetAuthZone("example.com")
    fmt.Println(authZone, domainID, err)

    result := v.CreateTXT(domainID, "_acme-challenge.example.com", "test challenge", 25)
    fmt.Println(result)

    recordID, err := v.GetRecordID(domainID, "_acme-challenge.example.com", "TXT")
    fmt.Println(recordID, err)

    err = v.DeleteRecord(recordID)
    fmt.Println(err)
}
```

Which will output the following:

```
2018/02/26 14:59:53 tmpHostname for i = 1: example.com
2018/02/26 14:59:53 {ok 1 [{active example.com 1 0}]}
2018/02/26 14:59:53 Found zone: example.com
	Shortened to example.com
example.com 1 <nil>
<nil>
3 <nil>
<nil>
```
