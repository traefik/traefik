package plugin

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/route"
	"github.com/vulcand/vulcand/conntracker"
	"github.com/vulcand/vulcand/router"
	"net/http"
	"reflect"
)

// Middleware specification, used to construct new middlewares and plug them into CLI API and backends
type MiddlewareSpec struct {
	Type string
	// Reader function that returns a middleware from another middleware structure
	FromOther interface{}
	// Flags for CLI tool to generate interface
	CliFlags []cli.Flag
	// Function that construtcs a middleware from CLI parameters
	FromCli CliReader
}

func (ms *MiddlewareSpec) FromJSON(data []byte) (Middleware, error) {
	// Get a function's type
	fnType := reflect.TypeOf(ms.FromOther)

	// Create a pointer to the function's first argument
	ptr := reflect.New(fnType.In(0)).Interface()
	err := json.Unmarshal(data, &ptr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %T from JSON, error: %s", ptr, err)
	}
	// Now let's call the function to produce a middleware
	fnVal := reflect.ValueOf(ms.FromOther)
	results := fnVal.Call([]reflect.Value{reflect.ValueOf(ptr).Elem()})

	m, out := results[0].Interface(), results[1].Interface()
	if out != nil {
		return nil, out.(error)
	}
	return m.(Middleware), nil
}

type Middleware interface {
	NewHandler(http.Handler) (http.Handler, error)
}

// Reader constructs the middleware from the CLI interface
type CliReader func(c *cli.Context) (Middleware, error)

// Function that returns middleware spec by it's type
type SpecGetter func(string) *MiddlewareSpec

//Holds a bunch of Listeners a frontend might have.
//This allows callers to consolidate all their listeners in one convenient struct.
type FrontendListeners struct {
	ConnTck           forward.UrlForwardingStateListener
	RbRewriteListener roundrobin.RequestRewriteListener
	RrRewriteListener roundrobin.RequestRewriteListener
}

// Registry contains currently registered middlewares and used to support pluggable middlewares across all modules of the vulcand
type Registry struct {
	specs                     []*MiddlewareSpec
	notFound                  Middleware
	router                    router.Router
	incomingConnectionTracker conntracker.ConnectionTracker
	frontendListeners         FrontendListeners
}

func NewRegistry() *Registry {
	return &Registry{
		specs:  []*MiddlewareSpec{},
		router: route.NewMux(),
	}
}

func (r *Registry) AddSpec(s *MiddlewareSpec) error {
	if s == nil {
		return fmt.Errorf("spec can not be nil")
	}
	if r.GetSpec(s.Type) != nil {
		return fmt.Errorf("middleware of type %s already registered", s.Type)
	}
	if err := verifySignature(s.FromOther); err != nil {
		return err
	}
	r.specs = append(r.specs, s)
	return nil
}

func (r *Registry) GetSpec(middlewareType string) *MiddlewareSpec {
	for _, s := range r.specs {
		if s.Type == middlewareType {
			return s
		}
	}
	return nil
}

func (r *Registry) GetSpecs() []*MiddlewareSpec {
	return r.specs
}

func (r *Registry) AddNotFoundMiddleware(notFound Middleware) error {
	r.notFound = notFound
	return nil
}

func (r *Registry) GetNotFoundMiddleware() Middleware {
	return r.notFound
}

func (r *Registry) SetRouter(router router.Router) error {
	r.router = router
	return nil
}

func (r *Registry) GetRouter() router.Router {
	return r.router
}

func (r *Registry) SetIncomingConnectionTracker(connTracker conntracker.ConnectionTracker) error {
	r.incomingConnectionTracker = connTracker
	return nil
}

func (r *Registry) GetIncomingConnectionTracker() conntracker.ConnectionTracker {
	return r.incomingConnectionTracker
}

func (r *Registry) GetFrontendListeners() FrontendListeners {
	return r.frontendListeners
}

func (r *Registry) SetFrontendListeners(frontendListeners FrontendListeners) error {
	r.frontendListeners = frontendListeners
	return nil
}

func verifySignature(fn interface{}) error {
	t := reflect.TypeOf(fn)
	if t == nil || t.Kind() != reflect.Func {
		return fmt.Errorf("expected function, got %s", t)
	}
	if t.NumIn() != 1 {
		return fmt.Errorf("expected function with one input argument, got %d", t.NumIn())
	}
	if t.In(0).Kind() != reflect.Struct {
		return fmt.Errorf("function argument should be struct, got %s", t.In(0).Kind())
	}
	if t.NumOut() != 2 {
		return fmt.Errorf("function should return 2 values, got %d", t.NumOut())
	}
	if !t.Out(0).AssignableTo(reflect.TypeOf((*Middleware)(nil)).Elem()) {
		return fmt.Errorf("function first return value should be Middleware got, %s", t.Out(0))
	}
	if !t.Out(1).AssignableTo(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("function second return value should be error got, %s", t.Out(1))
	}
	return nil
}
