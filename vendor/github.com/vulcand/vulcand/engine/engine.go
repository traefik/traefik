package engine

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/vulcand/plugin"
)

type NewEngineFn func() (Engine, error)

// Engine is an interface for storage and configuration engine, e.g. Etcd.
// Simple in memory implementation is available at engine/memng package
// Engines should pass the following acceptance suite to be compatible:
// engine/test/suite.go, see engine/etcdng/etcd_test.go and engine/memng/mem_test.go for details
type Engine interface {
	// GetHosts returns list of hosts registered in the storage engine
	// Returns empty list in case if there are no hosts.
	GetHosts() ([]Host, error)
	// GetHost returns host by given key, or engine.NotFoundError if it's not found
	GetHost(HostKey) (*Host, error)
	// UpsertHost updates or inserts the host, make sure to supply valid hostname
	UpsertHost(Host) error
	// DeleteHost deletes host by given key or returns engine.NotFoundError if it's not found
	DeleteHost(HostKey) error

	// GetListeners returns list of listeners registered in the storage engine
	// Returns empty list in case if there are no listeners
	GetListeners() ([]Listener, error)
	// GetListener returns a listener by key or engine.NotFoundError if it's not found
	GetListener(ListenerKey) (*Listener, error)
	// Updates or inserts a new listener, Listener.Id should not be empty
	UpsertListener(Listener) error
	// DeleteListener deletes a listener by key, returns engine.NotFoundError if it's not found
	DeleteListener(ListenerKey) error

	// GetFrontends returns a list of frontends registered in Vulcand
	// Returns empty list in case if there are no frontends
	GetFrontends() ([]Frontend, error)
	// GetFrontend returns a frontend by given key, or engine.NotFoundError if it's not found
	GetFrontend(FrontendKey) (*Frontend, error)
	// UpsertFrontend updates or inserts the frontend. Frontend.Id should not be empty. The second field specifies TTL, will be set to 0
	// in case if the frontend should not expire.
	UpsertFrontend(Frontend, time.Duration) error
	// DeleteFrontend deletes a frontend by a given key, returns engine.NotFoundError if it's not found
	DeleteFrontend(FrontendKey) error

	// GetMiddlewares returns middlewares registered for a given frontend
	// Returns empty list if there are no registered middlewares
	GetMiddlewares(FrontendKey) ([]Middleware, error)
	// GetMiddleware returns middleware by a given key, returns engine.NotFoundError if it's not there
	GetMiddleware(MiddlewareKey) (*Middleware, error)
	// UpsertMiddleware updates or inserts a middleware for a frontend. FrontendKey.Id and Middleware.Id should not be empty
	UpsertMiddleware(FrontendKey, Middleware, time.Duration) error
	// Delete middleware by given key, returns engine.NotFoundError if it's not found
	DeleteMiddleware(MiddlewareKey) error

	// GetBackends returns list of registered backends. Returns empty list if there are no backends
	GetBackends() ([]Backend, error)
	// GetBackend returns backend by given key, returns engine.NotFoundError if its not found
	GetBackend(BackendKey) (*Backend, error)
	// UpsertBackend updates or inserts a new backend. Backend.Id should not be empty
	UpsertBackend(Backend) error
	// DeleteBackend deletes backend by it's key. BackendKey.Id should not be empty. In case if backend is being used by frontends
	// this method should fail to preserve integrity, otherwise it will leave frontends in broken state
	DeleteBackend(BackendKey) error

	// GetServers returns servers assigned to the backend. BackendKey.Id should not be empty
	// Returns empty list if there are not assigned servers. Returns engine.NotFoundError if Backend does not exist
	GetServers(BackendKey) ([]Server, error)
	// GetServer returns server by given key or engine.NotFoundError if server is not found
	GetServer(ServerKey) (*Server, error)
	// UpsertServer updates or inserts a server. BackendKey.Id and Server.Id should not be empty.
	// TTL provides time to expire, in case if it's 0 server is permanent.
	UpsertServer(BackendKey, Server, time.Duration) error
	// DeleteServer deletes a server by given key. ServerKey.Id should not be empty.
	// Returns engine.NotFoundError if server not found
	DeleteServer(ServerKey) error

	// Subscribe is an entry point for getting the configuration changes as well as the initial configuration.
	// It should be a blocking function generating events from change.go to the changes channel.
	// Each change should be an instance of the struct provided in events.go
	// In  case if cancel channel is closed, the subscribe events should no longer be generated.
	Subscribe(events chan interface{}, cancel chan bool) error

	// GetRegistry returns registry with the supported plugins. It should be stored by Engine instance.
	GetRegistry() *plugin.Registry

	// GetLogSeverity returns the current logging severity level
	GetLogSeverity() log.Level
	// SetLogSeverity updates the logging severity level
	SetLogSeverity(log.Level)

	// Close should close all underlying resources such as connections, files, etc.
	Close()
}
