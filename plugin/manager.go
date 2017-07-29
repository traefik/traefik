package plugin

import (
	"fmt"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	goplugin "plugin"
)

// Manager is in charge of instantiating plugins and storing those in memory
type Manager struct {
	middlewares []*Middleware
}

// NewManager builds a new manager
func NewManager() *Manager {
	return &Manager{
		middlewares: []*Middleware{},
	}
}

// Load loads a plugin
func (m *Manager) Load(plugin Plugin) error {
	errChan := make(chan error)
	defer close(errChan)

	safe.GoWithRecover(func() {
		p, err := goplugin.Open(plugin.Path)
		if err != nil {
			errChan <- fmt.Errorf("error opening plugin: %s", err)
			return
		}
		loader, err := p.Lookup("Load")
		if err != nil {
			errChan <- fmt.Errorf("error in plugin Lookup: %s", err)
			return
		}
		load, ok := loader.(func() interface{})
		if !ok {
			errChan <- fmt.Errorf("plugin from %+v does not implement Load() interface{} function", plugin)
			return
		}
		instance := load()
		if instance == nil {
			errChan <- fmt.Errorf("plugin from %+v does not implement Load() interface{} function", plugin)
			return
		}
		if middleware, ok := instance.(Middleware); ok {
			// if Middleware, add to middleware list
			m.middlewares = append(m.middlewares, &middleware)
		} else {
			errChan <- fmt.Errorf("plugin from %+v does not implement any plugin interface", plugin)
			return
		}
		errChan <- nil
		return
	}, func(err interface{}) {
		log.Errorf("Error in plugin Go routine: %s", err)
	})

	if err, ok := <-errChan; ok {
		return err
	}
	return nil
}

// GetMiddlewares return a list of all Middleware plugins
func (m *Manager) GetMiddlewares() []*Middleware {
	return m.middlewares
}
