// mgo - MongoDB driver for Go
//
// Copyright (c) 2010-2012 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mgo

import (
	"errors"
	"net"
	"sort"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// ---------------------------------------------------------------------------
// Mongo server encapsulation.

type mongoServer struct {
	sync.RWMutex
	Addr          string
	ResolvedAddr  string
	tcpaddr       *net.TCPAddr
	unusedSockets []*mongoSocket
	liveSockets   []*mongoSocket
	closed        bool
	abended       bool
	sync          chan bool
	dial          dialer
	pingValue     time.Duration
	pingIndex     int
	pingCount     uint32
	pingWindow    [6]time.Duration
	info          *mongoServerInfo
}

type dialer struct {
	old func(addr net.Addr) (net.Conn, error)
	new func(addr *ServerAddr) (net.Conn, error)
}

func (dial dialer) isSet() bool {
	return dial.old != nil || dial.new != nil
}

type mongoServerInfo struct {
	Master         bool
	Mongos         bool
	Tags           bson.D
	MaxWireVersion int
	SetName        string
}

var defaultServerInfo mongoServerInfo

func newServer(addr string, tcpaddr *net.TCPAddr, sync chan bool, dial dialer) *mongoServer {
	server := &mongoServer{
		Addr:         addr,
		ResolvedAddr: tcpaddr.String(),
		tcpaddr:      tcpaddr,
		sync:         sync,
		dial:         dial,
		info:         &defaultServerInfo,
		pingValue:    time.Hour, // Push it back before an actual ping.
	}
	go server.pinger(true)
	return server
}

var errPoolLimit = errors.New("per-server connection limit reached")
var errServerClosed = errors.New("server was closed")

// AcquireSocket returns a socket for communicating with the server.
// This will attempt to reuse an old connection, if one is available. Otherwise,
// it will establish a new one. The returned socket is owned by the call site,
// and will return to the cache when the socket has its Release method called
// the same number of times as AcquireSocket + Acquire were called for it.
// If the poolLimit argument is greater than zero and the number of sockets in
// use in this server is greater than the provided limit, errPoolLimit is
// returned.
func (server *mongoServer) AcquireSocket(poolLimit int, timeout time.Duration) (socket *mongoSocket, abended bool, err error) {
	for {
		server.Lock()
		abended = server.abended
		if server.closed {
			server.Unlock()
			return nil, abended, errServerClosed
		}
		n := len(server.unusedSockets)
		if poolLimit > 0 && len(server.liveSockets)-n >= poolLimit {
			server.Unlock()
			return nil, false, errPoolLimit
		}
		if n > 0 {
			socket = server.unusedSockets[n-1]
			server.unusedSockets[n-1] = nil // Help GC.
			server.unusedSockets = server.unusedSockets[:n-1]
			info := server.info
			server.Unlock()
			err = socket.InitialAcquire(info, timeout)
			if err != nil {
				continue
			}
		} else {
			server.Unlock()
			socket, err = server.Connect(timeout)
			if err == nil {
				server.Lock()
				// We've waited for the Connect, see if we got
				// closed in the meantime
				if server.closed {
					server.Unlock()
					socket.Release()
					socket.Close()
					return nil, abended, errServerClosed
				}
				server.liveSockets = append(server.liveSockets, socket)
				server.Unlock()
			}
		}
		return
	}
	panic("unreachable")
}

// Connect establishes a new connection to the server. This should
// generally be done through server.AcquireSocket().
func (server *mongoServer) Connect(timeout time.Duration) (*mongoSocket, error) {
	server.RLock()
	master := server.info.Master
	dial := server.dial
	server.RUnlock()

	logf("Establishing new connection to %s (timeout=%s)...", server.Addr, timeout)
	var conn net.Conn
	var err error
	switch {
	case !dial.isSet():
		// Cannot do this because it lacks timeout support. :-(
		//conn, err = net.DialTCP("tcp", nil, server.tcpaddr)
		conn, err = net.DialTimeout("tcp", server.ResolvedAddr, timeout)
		if tcpconn, ok := conn.(*net.TCPConn); ok {
			tcpconn.SetKeepAlive(true)
		} else if err == nil {
			panic("internal error: obtained TCP connection is not a *net.TCPConn!?")
		}
	case dial.old != nil:
		conn, err = dial.old(server.tcpaddr)
	case dial.new != nil:
		conn, err = dial.new(&ServerAddr{server.Addr, server.tcpaddr})
	default:
		panic("dialer is set, but both dial.old and dial.new are nil")
	}
	if err != nil {
		logf("Connection to %s failed: %v", server.Addr, err.Error())
		return nil, err
	}
	logf("Connection to %s established.", server.Addr)

	stats.conn(+1, master)
	return newSocket(server, conn, timeout), nil
}

// Close forces closing all sockets that are alive, whether
// they're currently in use or not.
func (server *mongoServer) Close() {
	server.Lock()
	server.closed = true
	liveSockets := server.liveSockets
	unusedSockets := server.unusedSockets
	server.liveSockets = nil
	server.unusedSockets = nil
	server.Unlock()
	logf("Connections to %s closing (%d live sockets).", server.Addr, len(liveSockets))
	for i, s := range liveSockets {
		s.Close()
		liveSockets[i] = nil
	}
	for i := range unusedSockets {
		unusedSockets[i] = nil
	}
}

// RecycleSocket puts socket back into the unused cache.
func (server *mongoServer) RecycleSocket(socket *mongoSocket) {
	server.Lock()
	if !server.closed {
		server.unusedSockets = append(server.unusedSockets, socket)
	}
	server.Unlock()
}

func removeSocket(sockets []*mongoSocket, socket *mongoSocket) []*mongoSocket {
	for i, s := range sockets {
		if s == socket {
			copy(sockets[i:], sockets[i+1:])
			n := len(sockets) - 1
			sockets[n] = nil
			sockets = sockets[:n]
			break
		}
	}
	return sockets
}

// AbendSocket notifies the server that the given socket has terminated
// abnormally, and thus should be discarded rather than cached.
func (server *mongoServer) AbendSocket(socket *mongoSocket) {
	server.Lock()
	server.abended = true
	if server.closed {
		server.Unlock()
		return
	}
	server.liveSockets = removeSocket(server.liveSockets, socket)
	server.unusedSockets = removeSocket(server.unusedSockets, socket)
	server.Unlock()
	// Maybe just a timeout, but suggest a cluster sync up just in case.
	select {
	case server.sync <- true:
	default:
	}
}

func (server *mongoServer) SetInfo(info *mongoServerInfo) {
	server.Lock()
	server.info = info
	server.Unlock()
}

func (server *mongoServer) Info() *mongoServerInfo {
	server.Lock()
	info := server.info
	server.Unlock()
	return info
}

func (server *mongoServer) hasTags(serverTags []bson.D) bool {
NextTagSet:
	for _, tags := range serverTags {
	NextReqTag:
		for _, req := range tags {
			for _, has := range server.info.Tags {
				if req.Name == has.Name {
					if req.Value == has.Value {
						continue NextReqTag
					}
					continue NextTagSet
				}
			}
			continue NextTagSet
		}
		return true
	}
	return false
}

var pingDelay = 15 * time.Second

func (server *mongoServer) pinger(loop bool) {
	var delay time.Duration
	if raceDetector {
		// This variable is only ever touched by tests.
		globalMutex.Lock()
		delay = pingDelay
		globalMutex.Unlock()
	} else {
		delay = pingDelay
	}
	op := queryOp{
		collection: "admin.$cmd",
		query:      bson.D{{"ping", 1}},
		flags:      flagSlaveOk,
		limit:      -1,
	}
	for {
		if loop {
			time.Sleep(delay)
		}
		op := op
		socket, _, err := server.AcquireSocket(0, delay)
		if err == nil {
			start := time.Now()
			_, _ = socket.SimpleQuery(&op)
			delay := time.Now().Sub(start)

			server.pingWindow[server.pingIndex] = delay
			server.pingIndex = (server.pingIndex + 1) % len(server.pingWindow)
			server.pingCount++
			var max time.Duration
			for i := 0; i < len(server.pingWindow) && uint32(i) < server.pingCount; i++ {
				if server.pingWindow[i] > max {
					max = server.pingWindow[i]
				}
			}
			socket.Release()
			server.Lock()
			if server.closed {
				loop = false
			}
			server.pingValue = max
			server.Unlock()
			logf("Ping for %s is %d ms", server.Addr, max/time.Millisecond)
		} else if err == errServerClosed {
			return
		}
		if !loop {
			return
		}
	}
}

type mongoServerSlice []*mongoServer

func (s mongoServerSlice) Len() int {
	return len(s)
}

func (s mongoServerSlice) Less(i, j int) bool {
	return s[i].ResolvedAddr < s[j].ResolvedAddr
}

func (s mongoServerSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s mongoServerSlice) Sort() {
	sort.Sort(s)
}

func (s mongoServerSlice) Search(resolvedAddr string) (i int, ok bool) {
	n := len(s)
	i = sort.Search(n, func(i int) bool {
		return s[i].ResolvedAddr >= resolvedAddr
	})
	return i, i != n && s[i].ResolvedAddr == resolvedAddr
}

type mongoServers struct {
	slice mongoServerSlice
}

func (servers *mongoServers) Search(resolvedAddr string) (server *mongoServer) {
	if i, ok := servers.slice.Search(resolvedAddr); ok {
		return servers.slice[i]
	}
	return nil
}

func (servers *mongoServers) Add(server *mongoServer) {
	servers.slice = append(servers.slice, server)
	servers.slice.Sort()
}

func (servers *mongoServers) Remove(other *mongoServer) (server *mongoServer) {
	if i, found := servers.slice.Search(other.ResolvedAddr); found {
		server = servers.slice[i]
		copy(servers.slice[i:], servers.slice[i+1:])
		n := len(servers.slice) - 1
		servers.slice[n] = nil // Help GC.
		servers.slice = servers.slice[:n]
	}
	return
}

func (servers *mongoServers) Slice() []*mongoServer {
	return ([]*mongoServer)(servers.slice)
}

func (servers *mongoServers) Get(i int) *mongoServer {
	return servers.slice[i]
}

func (servers *mongoServers) Len() int {
	return len(servers.slice)
}

func (servers *mongoServers) Empty() bool {
	return len(servers.slice) == 0
}

func (servers *mongoServers) HasMongos() bool {
	for _, s := range servers.slice {
		if s.Info().Mongos {
			return true
		}
	}
	return false
}

// BestFit returns the best guess of what would be the most interesting
// server to perform operations on at this point in time.
func (servers *mongoServers) BestFit(mode Mode, serverTags []bson.D) *mongoServer {
	var best *mongoServer
	for _, next := range servers.slice {
		if best == nil {
			best = next
			best.RLock()
			if serverTags != nil && !next.info.Mongos && !best.hasTags(serverTags) {
				best.RUnlock()
				best = nil
			}
			continue
		}
		next.RLock()
		swap := false
		switch {
		case serverTags != nil && !next.info.Mongos && !next.hasTags(serverTags):
			// Must have requested tags.
		case mode == Secondary && next.info.Master && !next.info.Mongos:
			// Must be a secondary or mongos.
		case next.info.Master != best.info.Master && mode != Nearest:
			// Prefer slaves, unless the mode is PrimaryPreferred.
			swap = (mode == PrimaryPreferred) != best.info.Master
		case absDuration(next.pingValue-best.pingValue) > 15*time.Millisecond:
			// Prefer nearest server.
			swap = next.pingValue < best.pingValue
		case len(next.liveSockets)-len(next.unusedSockets) < len(best.liveSockets)-len(best.unusedSockets):
			// Prefer servers with less connections.
			swap = true
		}
		if swap {
			best.RUnlock()
			best = next
		} else {
			next.RUnlock()
		}
	}
	if best != nil {
		best.RUnlock()
	}
	return best
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
