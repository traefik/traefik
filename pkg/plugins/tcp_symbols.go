package plugins

import (
	"reflect"

	"github.com/traefik/traefik/v3/pkg/tcp"
)

// tcpSymbols provides symbols for TCP package to yaegi plugins.
func tcpSymbols() map[string]map[string]reflect.Value {
	return map[string]map[string]reflect.Value{
		"github.com/traefik/traefik/v3/pkg/tcp/tcp": {
			// Export the Handler interface type
			"Handler": reflect.ValueOf((*tcp.Handler)(nil)).Elem(),
			// Export the WriteCloser interface type
			"WriteCloser": reflect.ValueOf((*tcp.WriteCloser)(nil)).Elem(),
			// Export the HandlerFunc type constructor
			"HandlerFunc": reflect.ValueOf(func(f func(tcp.WriteCloser)) tcp.HandlerFunc {
				return tcp.HandlerFunc(f)
			}),
		},
	}
}
