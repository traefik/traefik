package proxy

import (
	"github.com/vulcand/vulcand/conntracker"
	"net"
	"net/http"
	"sync"
)

type connTracker struct {
	mtx *sync.Mutex

	new    map[string]int64
	active map[string]int64
	idle   map[string]int64
}

func newDefaultConnTracker() conntracker.ConnectionTracker {
	return &connTracker{
		mtx:    &sync.Mutex{},
		new:    make(map[string]int64),
		active: make(map[string]int64),
		idle:   make(map[string]int64),
	}
}

func (c *connTracker) RegisterStateChange(conn net.Conn, prev http.ConnState, cur http.ConnState) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if cur == http.StateNew || cur == http.StateIdle || cur == http.StateActive {
		c.inc(conn, cur, 1)
	}

	if cur != http.StateNew {
		c.inc(conn, prev, -1)
	}
}

func (c *connTracker) inc(conn net.Conn, state http.ConnState, v int64) {
	addr := conn.LocalAddr().String()
	var m map[string]int64

	switch state {
	case http.StateNew:
		m = c.new
	case http.StateActive:
		m = c.active
	case http.StateIdle:
		m = c.idle
	default:
		return
	}

	m[addr] += v
}

func (c *connTracker) Counts() conntracker.ConnectionStats {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return conntracker.ConnectionStats{
		http.StateNew:    c.copy(c.new),
		http.StateActive: c.copy(c.active),
		http.StateIdle:   c.copy(c.idle),
	}
}

func (c *connTracker) copy(s map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}
