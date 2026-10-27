package server

import (
	"context"
	"net"
	"sync"
)

type connContextFunc func(context.Context, net.Conn) context.Context

type multipleConnContext struct {
	fns []connContextFunc
}

func (m *multipleConnContext) AddConnContextFunc(fn connContextFunc) {
	m.fns = append(m.fns, fn)
}

func (m *multipleConnContext) Build() connContextFunc {
	if len(m.fns) == 0 {
		return nil
	}

	return func(ctx context.Context, c net.Conn) context.Context {
		for _, contextFunc := range m.fns {
			ctx = contextFunc(ctx, c)
		}
		return ctx
	}
}

// connContexts keeps the context of each connection until it is closed, as the http.Server ConnState hook is only
// given the connection, while the resources to release on close are only reachable from its context.
type connContexts struct {
	mu   sync.Mutex
	ctxs map[net.Conn]context.Context
}

func (c *connContexts) add(ctx context.Context, conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ctxs == nil {
		c.ctxs = make(map[net.Conn]context.Context)
	}
	c.ctxs[conn] = ctx
}

// remove returns the context of the given connection and forgets it,
// so that the resources of a connection are released only once.
func (c *connContexts) remove(conn net.Conn) (context.Context, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, ok := c.ctxs[conn]
	delete(c.ctxs, conn)

	return ctx, ok
}
