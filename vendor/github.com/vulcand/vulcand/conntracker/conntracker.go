package conntracker

import (
	"net"
	"net/http"
)

type ConnectionTracker interface {
	RegisterStateChange(conn net.Conn, prev http.ConnState, cur http.ConnState)
	Counts() ConnectionStats
}

type ConnectionStats map[http.ConnState]map[string]int64
