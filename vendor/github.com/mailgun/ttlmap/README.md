**This repo is deprecated, Renamed to TTLMap and moved to http://github.com/mailgun/holster**

[![Build Status](https://travis-ci.org/mailgun/ttlmap.png)](https://travis-ci.org/mailgun/ttlmap)

TtlMap
=======

Redis-like Map with expiry times and maximum capacity

```go

import "github.com/mailgun/ttlmap"

mh, _ := ttlmap.NewMap(20)
mh.Set("key1", "value", 20)
valI, exists := mh.Get("key2")
if exists {
   val := valI.(string)
}
```

The ttlmap is not thread safe by default. You can either create a thread safe
instance with `ttlmap.NewConcurrent` that is effectively using `sync.RWLock`,
or implement locking in you application. Beware though that at the application
level `sync.RWLock` cannot be used, because `ttlmap.Get` can occasionally
modifies the internal data structure.
