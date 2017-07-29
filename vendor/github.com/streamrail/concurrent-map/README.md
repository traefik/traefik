# concurrent map [![Circle CI](https://circleci.com/gh/streamrail/concurrent-map.png?style=badge)](https://circleci.com/gh/streamrail/concurrent-map)

As explained [here](http://golang.org/doc/faq#atomic_maps) and [here](http://blog.golang.org/go-maps-in-action), the `map` type in Go doesn't support concurrent reads and writes. `concurrent-map` provides a high-performance solution to this by sharding the map with minimal time spent waiting for locks.

## usage

Import the package:

```go
import (
	"github.com/streamrail/concurrent-map"
)

```

```bash
go get "github.com/streamrail/concurrent-map"
```

The package is now imported under the "cmap" namespace. 

## example

```go

	// Create a new map.
	map := cmap.New()
	
	// Sets item within map, sets "bar" under key "foo"
	map.Set("foo", "bar")

	// Retrieve item from map.
	if tmp, ok := map.Get("foo"); ok {
		bar := tmp.(string)
	}

	// Removes item under key "foo"
	map.Remove("foo")

```

For more examples have a look at concurrent_map_test.go.


Running tests:

```bash
go test "github.com/streamrail/concurrent-map"
```


## license 
MIT (see [LICENSE](https://github.com/streamrail/concurrent-map/blob/master/LICENSE) file)
