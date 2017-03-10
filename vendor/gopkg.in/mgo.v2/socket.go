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
	"fmt"
	"net"
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type replyFunc func(err error, reply *replyOp, docNum int, docData []byte)

type mongoSocket struct {
	sync.Mutex
	server        *mongoServer // nil when cached
	conn          net.Conn
	timeout       time.Duration
	addr          string // For debugging only.
	nextRequestId uint32
	replyFuncs    map[uint32]replyFunc
	references    int
	creds         []Credential
	logout        []Credential
	cachedNonce   string
	gotNonce      sync.Cond
	dead          error
	serverInfo    *mongoServerInfo
}

type queryOpFlags uint32

const (
	_ queryOpFlags = 1 << iota
	flagTailable
	flagSlaveOk
	flagLogReplay
	flagNoCursorTimeout
	flagAwaitData
)

type queryOp struct {
	collection string
	query      interface{}
	skip       int32
	limit      int32
	selector   interface{}
	flags      queryOpFlags
	replyFunc  replyFunc

	mode       Mode
	options    queryWrapper
	hasOptions bool
	serverTags []bson.D
}

type queryWrapper struct {
	Query          interface{} "$query"
	OrderBy        interface{} "$orderby,omitempty"
	Hint           interface{} "$hint,omitempty"
	Explain        bool        "$explain,omitempty"
	Snapshot       bool        "$snapshot,omitempty"
	ReadPreference bson.D      "$readPreference,omitempty"
	MaxScan        int         "$maxScan,omitempty"
	MaxTimeMS      int         "$maxTimeMS,omitempty"
	Comment        string      "$comment,omitempty"
}

func (op *queryOp) finalQuery(socket *mongoSocket) interface{} {
	if op.flags&flagSlaveOk != 0 && socket.ServerInfo().Mongos {
		var modeName string
		switch op.mode {
		case Strong:
			modeName = "primary"
		case Monotonic, Eventual:
			modeName = "secondaryPreferred"
		case PrimaryPreferred:
			modeName = "primaryPreferred"
		case Secondary:
			modeName = "secondary"
		case SecondaryPreferred:
			modeName = "secondaryPreferred"
		case Nearest:
			modeName = "nearest"
		default:
			panic(fmt.Sprintf("unsupported read mode: %d", op.mode))
		}
		op.hasOptions = true
		op.options.ReadPreference = make(bson.D, 0, 2)
		op.options.ReadPreference = append(op.options.ReadPreference, bson.DocElem{"mode", modeName})
		if len(op.serverTags) > 0 {
			op.options.ReadPreference = append(op.options.ReadPreference, bson.DocElem{"tags", op.serverTags})
		}
	}
	if op.hasOptions {
		if op.query == nil {
			var empty bson.D
			op.options.Query = empty
		} else {
			op.options.Query = op.query
		}
		debugf("final query is %#v\n", &op.options)
		return &op.options
	}
	return op.query
}

type getMoreOp struct {
	collection string
	limit      int32
	cursorId   int64
	replyFunc  replyFunc
}

type replyOp struct {
	flags     uint32
	cursorId  int64
	firstDoc  int32
	replyDocs int32
}

type insertOp struct {
	collection string        // "database.collection"
	documents  []interface{} // One or more documents to insert
	flags      uint32
}

type updateOp struct {
	Collection string      `bson:"-"` // "database.collection"
	Selector   interface{} `bson:"q"`
	Update     interface{} `bson:"u"`
	Flags      uint32      `bson:"-"`
	Multi      bool        `bson:"multi,omitempty"`
	Upsert     bool        `bson:"upsert,omitempty"`
}

type deleteOp struct {
	Collection string      `bson:"-"` // "database.collection"
	Selector   interface{} `bson:"q"`
	Flags      uint32      `bson:"-"`
	Limit      int         `bson:"limit"`
}

type killCursorsOp struct {
	cursorIds []int64
}

type requestInfo struct {
	bufferPos int
	replyFunc replyFunc
}

func newSocket(server *mongoServer, conn net.Conn, timeout time.Duration) *mongoSocket {
	socket := &mongoSocket{
		conn:       conn,
		addr:       server.Addr,
		server:     server,
		replyFuncs: make(map[uint32]replyFunc),
	}
	socket.gotNonce.L = &socket.Mutex
	if err := socket.InitialAcquire(server.Info(), timeout); err != nil {
		panic("newSocket: InitialAcquire returned error: " + err.Error())
	}
	stats.socketsAlive(+1)
	debugf("Socket %p to %s: initialized", socket, socket.addr)
	socket.resetNonce()
	go socket.readLoop()
	return socket
}

// Server returns the server that the socket is associated with.
// It returns nil while the socket is cached in its respective server.
func (socket *mongoSocket) Server() *mongoServer {
	socket.Lock()
	server := socket.server
	socket.Unlock()
	return server
}

// ServerInfo returns details for the server at the time the socket
// was initially acquired.
func (socket *mongoSocket) ServerInfo() *mongoServerInfo {
	socket.Lock()
	serverInfo := socket.serverInfo
	socket.Unlock()
	return serverInfo
}

// InitialAcquire obtains the first reference to the socket, either
// right after the connection is made or once a recycled socket is
// being put back in use.
func (socket *mongoSocket) InitialAcquire(serverInfo *mongoServerInfo, timeout time.Duration) error {
	socket.Lock()
	if socket.references > 0 {
		panic("Socket acquired out of cache with references")
	}
	if socket.dead != nil {
		dead := socket.dead
		socket.Unlock()
		return dead
	}
	socket.references++
	socket.serverInfo = serverInfo
	socket.timeout = timeout
	stats.socketsInUse(+1)
	stats.socketRefs(+1)
	socket.Unlock()
	return nil
}

// Acquire obtains an additional reference to the socket.
// The socket will only be recycled when it's released as many
// times as it's been acquired.
func (socket *mongoSocket) Acquire() (info *mongoServerInfo) {
	socket.Lock()
	if socket.references == 0 {
		panic("Socket got non-initial acquire with references == 0")
	}
	// We'll track references to dead sockets as well.
	// Caller is still supposed to release the socket.
	socket.references++
	stats.socketRefs(+1)
	serverInfo := socket.serverInfo
	socket.Unlock()
	return serverInfo
}

// Release decrements a socket reference. The socket will be
// recycled once its released as many times as it's been acquired.
func (socket *mongoSocket) Release() {
	socket.Lock()
	if socket.references == 0 {
		panic("socket.Release() with references == 0")
	}
	socket.references--
	stats.socketRefs(-1)
	if socket.references == 0 {
		stats.socketsInUse(-1)
		server := socket.server
		socket.Unlock()
		socket.LogoutAll()
		// If the socket is dead server is nil.
		if server != nil {
			server.RecycleSocket(socket)
		}
	} else {
		socket.Unlock()
	}
}

// SetTimeout changes the timeout used on socket operations.
func (socket *mongoSocket) SetTimeout(d time.Duration) {
	socket.Lock()
	socket.timeout = d
	socket.Unlock()
}

type deadlineType int

const (
	readDeadline  deadlineType = 1
	writeDeadline deadlineType = 2
)

func (socket *mongoSocket) updateDeadline(which deadlineType) {
	var when time.Time
	if socket.timeout > 0 {
		when = time.Now().Add(socket.timeout)
	}
	whichstr := ""
	switch which {
	case readDeadline | writeDeadline:
		whichstr = "read/write"
		socket.conn.SetDeadline(when)
	case readDeadline:
		whichstr = "read"
		socket.conn.SetReadDeadline(when)
	case writeDeadline:
		whichstr = "write"
		socket.conn.SetWriteDeadline(when)
	default:
		panic("invalid parameter to updateDeadline")
	}
	debugf("Socket %p to %s: updated %s deadline to %s ahead (%s)", socket, socket.addr, whichstr, socket.timeout, when)
}

// Close terminates the socket use.
func (socket *mongoSocket) Close() {
	socket.kill(errors.New("Closed explicitly"), false)
}

func (socket *mongoSocket) kill(err error, abend bool) {
	socket.Lock()
	if socket.dead != nil {
		debugf("Socket %p to %s: killed again: %s (previously: %s)", socket, socket.addr, err.Error(), socket.dead.Error())
		socket.Unlock()
		return
	}
	logf("Socket %p to %s: closing: %s (abend=%v)", socket, socket.addr, err.Error(), abend)
	socket.dead = err
	socket.conn.Close()
	stats.socketsAlive(-1)
	replyFuncs := socket.replyFuncs
	socket.replyFuncs = make(map[uint32]replyFunc)
	server := socket.server
	socket.server = nil
	socket.gotNonce.Broadcast()
	socket.Unlock()
	for _, replyFunc := range replyFuncs {
		logf("Socket %p to %s: notifying replyFunc of closed socket: %s", socket, socket.addr, err.Error())
		replyFunc(err, nil, -1, nil)
	}
	if abend {
		server.AbendSocket(socket)
	}
}

func (socket *mongoSocket) SimpleQuery(op *queryOp) (data []byte, err error) {
	var wait, change sync.Mutex
	var replyDone bool
	var replyData []byte
	var replyErr error
	wait.Lock()
	op.replyFunc = func(err error, reply *replyOp, docNum int, docData []byte) {
		change.Lock()
		if !replyDone {
			replyDone = true
			replyErr = err
			if err == nil {
				replyData = docData
			}
		}
		change.Unlock()
		wait.Unlock()
	}
	err = socket.Query(op)
	if err != nil {
		return nil, err
	}
	wait.Lock()
	change.Lock()
	data = replyData
	err = replyErr
	change.Unlock()
	return data, err
}

func (socket *mongoSocket) Query(ops ...interface{}) (err error) {

	if lops := socket.flushLogout(); len(lops) > 0 {
		ops = append(lops, ops...)
	}

	buf := make([]byte, 0, 256)

	// Serialize operations synchronously to avoid interrupting
	// other goroutines while we can't really be sending data.
	// Also, record id positions so that we can compute request
	// ids at once later with the lock already held.
	requests := make([]requestInfo, len(ops))
	requestCount := 0

	for _, op := range ops {
		debugf("Socket %p to %s: serializing op: %#v", socket, socket.addr, op)
		if qop, ok := op.(*queryOp); ok {
			if cmd, ok := qop.query.(*findCmd); ok {
				debugf("Socket %p to %s: find command: %#v", socket, socket.addr, cmd)
			}
		}
		start := len(buf)
		var replyFunc replyFunc
		switch op := op.(type) {

		case *updateOp:
			buf = addHeader(buf, 2001)
			buf = addInt32(buf, 0) // Reserved
			buf = addCString(buf, op.Collection)
			buf = addInt32(buf, int32(op.Flags))
			debugf("Socket %p to %s: serializing selector document: %#v", socket, socket.addr, op.Selector)
			buf, err = addBSON(buf, op.Selector)
			if err != nil {
				return err
			}
			debugf("Socket %p to %s: serializing update document: %#v", socket, socket.addr, op.Update)
			buf, err = addBSON(buf, op.Update)
			if err != nil {
				return err
			}

		case *insertOp:
			buf = addHeader(buf, 2002)
			buf = addInt32(buf, int32(op.flags))
			buf = addCString(buf, op.collection)
			for _, doc := range op.documents {
				debugf("Socket %p to %s: serializing document for insertion: %#v", socket, socket.addr, doc)
				buf, err = addBSON(buf, doc)
				if err != nil {
					return err
				}
			}

		case *queryOp:
			buf = addHeader(buf, 2004)
			buf = addInt32(buf, int32(op.flags))
			buf = addCString(buf, op.collection)
			buf = addInt32(buf, op.skip)
			buf = addInt32(buf, op.limit)
			buf, err = addBSON(buf, op.finalQuery(socket))
			if err != nil {
				return err
			}
			if op.selector != nil {
				buf, err = addBSON(buf, op.selector)
				if err != nil {
					return err
				}
			}
			replyFunc = op.replyFunc

		case *getMoreOp:
			buf = addHeader(buf, 2005)
			buf = addInt32(buf, 0) // Reserved
			buf = addCString(buf, op.collection)
			buf = addInt32(buf, op.limit)
			buf = addInt64(buf, op.cursorId)
			replyFunc = op.replyFunc

		case *deleteOp:
			buf = addHeader(buf, 2006)
			buf = addInt32(buf, 0) // Reserved
			buf = addCString(buf, op.Collection)
			buf = addInt32(buf, int32(op.Flags))
			debugf("Socket %p to %s: serializing selector document: %#v", socket, socket.addr, op.Selector)
			buf, err = addBSON(buf, op.Selector)
			if err != nil {
				return err
			}

		case *killCursorsOp:
			buf = addHeader(buf, 2007)
			buf = addInt32(buf, 0) // Reserved
			buf = addInt32(buf, int32(len(op.cursorIds)))
			for _, cursorId := range op.cursorIds {
				buf = addInt64(buf, cursorId)
			}

		default:
			panic("internal error: unknown operation type")
		}

		setInt32(buf, start, int32(len(buf)-start))

		if replyFunc != nil {
			request := &requests[requestCount]
			request.replyFunc = replyFunc
			request.bufferPos = start
			requestCount++
		}
	}

	// Buffer is ready for the pipe.  Lock, allocate ids, and enqueue.

	socket.Lock()
	if socket.dead != nil {
		dead := socket.dead
		socket.Unlock()
		debugf("Socket %p to %s: failing query, already closed: %s", socket, socket.addr, socket.dead.Error())
		// XXX This seems necessary in case the session is closed concurrently
		// with a query being performed, but it's not yet tested:
		for i := 0; i != requestCount; i++ {
			request := &requests[i]
			if request.replyFunc != nil {
				request.replyFunc(dead, nil, -1, nil)
			}
		}
		return dead
	}

	wasWaiting := len(socket.replyFuncs) > 0

	// Reserve id 0 for requests which should have no responses.
	requestId := socket.nextRequestId + 1
	if requestId == 0 {
		requestId++
	}
	socket.nextRequestId = requestId + uint32(requestCount)
	for i := 0; i != requestCount; i++ {
		request := &requests[i]
		setInt32(buf, request.bufferPos+4, int32(requestId))
		socket.replyFuncs[requestId] = request.replyFunc
		requestId++
	}

	debugf("Socket %p to %s: sending %d op(s) (%d bytes)", socket, socket.addr, len(ops), len(buf))
	stats.sentOps(len(ops))

	socket.updateDeadline(writeDeadline)
	_, err = socket.conn.Write(buf)
	if !wasWaiting && requestCount > 0 {
		socket.updateDeadline(readDeadline)
	}
	socket.Unlock()
	return err
}

func fill(r net.Conn, b []byte) error {
	l := len(b)
	n, err := r.Read(b)
	for n != l && err == nil {
		var ni int
		ni, err = r.Read(b[n:])
		n += ni
	}
	return err
}

// Estimated minimum cost per socket: 1 goroutine + memory for the largest
// document ever seen.
func (socket *mongoSocket) readLoop() {
	p := make([]byte, 36) // 16 from header + 20 from OP_REPLY fixed fields
	s := make([]byte, 4)
	conn := socket.conn // No locking, conn never changes.
	for {
		err := fill(conn, p)
		if err != nil {
			socket.kill(err, true)
			return
		}

		totalLen := getInt32(p, 0)
		responseTo := getInt32(p, 8)
		opCode := getInt32(p, 12)

		// Don't use socket.server.Addr here.  socket is not
		// locked and socket.server may go away.
		debugf("Socket %p to %s: got reply (%d bytes)", socket, socket.addr, totalLen)

		_ = totalLen

		if opCode != 1 {
			socket.kill(errors.New("opcode != 1, corrupted data?"), true)
			return
		}

		reply := replyOp{
			flags:     uint32(getInt32(p, 16)),
			cursorId:  getInt64(p, 20),
			firstDoc:  getInt32(p, 28),
			replyDocs: getInt32(p, 32),
		}

		stats.receivedOps(+1)
		stats.receivedDocs(int(reply.replyDocs))

		socket.Lock()
		replyFunc, ok := socket.replyFuncs[uint32(responseTo)]
		if ok {
			delete(socket.replyFuncs, uint32(responseTo))
		}
		socket.Unlock()

		if replyFunc != nil && reply.replyDocs == 0 {
			replyFunc(nil, &reply, -1, nil)
		} else {
			for i := 0; i != int(reply.replyDocs); i++ {
				err := fill(conn, s)
				if err != nil {
					if replyFunc != nil {
						replyFunc(err, nil, -1, nil)
					}
					socket.kill(err, true)
					return
				}

				b := make([]byte, int(getInt32(s, 0)))

				// copy(b, s) in an efficient way.
				b[0] = s[0]
				b[1] = s[1]
				b[2] = s[2]
				b[3] = s[3]

				err = fill(conn, b[4:])
				if err != nil {
					if replyFunc != nil {
						replyFunc(err, nil, -1, nil)
					}
					socket.kill(err, true)
					return
				}

				if globalDebug && globalLogger != nil {
					m := bson.M{}
					if err := bson.Unmarshal(b, m); err == nil {
						debugf("Socket %p to %s: received document: %#v", socket, socket.addr, m)
					}
				}

				if replyFunc != nil {
					replyFunc(nil, &reply, i, b)
				}

				// XXX Do bound checking against totalLen.
			}
		}

		socket.Lock()
		if len(socket.replyFuncs) == 0 {
			// Nothing else to read for now. Disable deadline.
			socket.conn.SetReadDeadline(time.Time{})
		} else {
			socket.updateDeadline(readDeadline)
		}
		socket.Unlock()

		// XXX Do bound checking against totalLen.
	}
}

var emptyHeader = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func addHeader(b []byte, opcode int) []byte {
	i := len(b)
	b = append(b, emptyHeader...)
	// Enough for current opcodes.
	b[i+12] = byte(opcode)
	b[i+13] = byte(opcode >> 8)
	return b
}

func addInt32(b []byte, i int32) []byte {
	return append(b, byte(i), byte(i>>8), byte(i>>16), byte(i>>24))
}

func addInt64(b []byte, i int64) []byte {
	return append(b, byte(i), byte(i>>8), byte(i>>16), byte(i>>24),
		byte(i>>32), byte(i>>40), byte(i>>48), byte(i>>56))
}

func addCString(b []byte, s string) []byte {
	b = append(b, []byte(s)...)
	b = append(b, 0)
	return b
}

func addBSON(b []byte, doc interface{}) ([]byte, error) {
	if doc == nil {
		return append(b, 5, 0, 0, 0, 0), nil
	}
	data, err := bson.Marshal(doc)
	if err != nil {
		return b, err
	}
	return append(b, data...), nil
}

func setInt32(b []byte, pos int, i int32) {
	b[pos] = byte(i)
	b[pos+1] = byte(i >> 8)
	b[pos+2] = byte(i >> 16)
	b[pos+3] = byte(i >> 24)
}

func getInt32(b []byte, pos int) int32 {
	return (int32(b[pos+0])) |
		(int32(b[pos+1]) << 8) |
		(int32(b[pos+2]) << 16) |
		(int32(b[pos+3]) << 24)
}

func getInt64(b []byte, pos int) int64 {
	return (int64(b[pos+0])) |
		(int64(b[pos+1]) << 8) |
		(int64(b[pos+2]) << 16) |
		(int64(b[pos+3]) << 24) |
		(int64(b[pos+4]) << 32) |
		(int64(b[pos+5]) << 40) |
		(int64(b[pos+6]) << 48) |
		(int64(b[pos+7]) << 56)
}
