package consul

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/consul/agent"
	"github.com/hashicorp/consul/consul/state"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/net-rpc-msgpackrpc"
	"github.com/hashicorp/yamux"
)

type RPCType byte

const (
	rpcConsul RPCType = iota
	rpcRaft
	rpcMultiplex // Old Muxado byte, no longer supported.
	rpcTLS
	rpcMultiplexV2
	rpcSnapshot
	rpcGossip
)

const (
	// maxQueryTime is used to bound the limit of a blocking query
	maxQueryTime = 600 * time.Second

	// defaultQueryTime is the amount of time we block waiting for a change
	// if no time is specified. Previously we would wait the maxQueryTime.
	defaultQueryTime = 300 * time.Second

	// jitterFraction is a the limit to the amount of jitter we apply
	// to a user specified MaxQueryTime. We divide the specified time by
	// the fraction. So 16 == 6.25% limit of jitter. This same fraction
	// is applied to the RPCHoldTimeout
	jitterFraction = 16

	// Warn if the Raft command is larger than this.
	// If it's over 1MB something is probably being abusive.
	raftWarnSize = 1024 * 1024

	// enqueueLimit caps how long we will wait to enqueue
	// a new Raft command. Something is probably wrong if this
	// value is ever reached. However, it prevents us from blocking
	// the requesting goroutine forever.
	enqueueLimit = 30 * time.Second
)

// listen is used to listen for incoming RPC connections
func (s *Server) listen() {
	for {
		// Accept a connection
		conn, err := s.rpcListener.Accept()
		if err != nil {
			if s.shutdown {
				return
			}
			s.logger.Printf("[ERR] consul.rpc: failed to accept RPC conn: %v", err)
			continue
		}

		go s.handleConn(conn, false)
		metrics.IncrCounter([]string{"consul", "rpc", "accept_conn"}, 1)
	}
}

// logConn is a wrapper around memberlist's LogConn so that we format references
// to "from" addresses in a consistent way. This is just a shorter name.
func logConn(conn net.Conn) string {
	return memberlist.LogConn(conn)
}

// handleConn is used to determine if this is a Raft or
// Consul type RPC connection and invoke the correct handler
func (s *Server) handleConn(conn net.Conn, isTLS bool) {
	// Read a single byte
	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err != nil {
		if err != io.EOF {
			s.logger.Printf("[ERR] consul.rpc: failed to read byte: %v %s", err, logConn(conn))
		}
		conn.Close()
		return
	}

	// Enforce TLS if VerifyIncoming is set
	if s.config.VerifyIncoming && !isTLS && RPCType(buf[0]) != rpcTLS {
		s.logger.Printf("[WARN] consul.rpc: Non-TLS connection attempted with VerifyIncoming set %s", logConn(conn))
		conn.Close()
		return
	}

	// Switch on the byte
	switch RPCType(buf[0]) {
	case rpcConsul:
		s.handleConsulConn(conn)

	case rpcRaft:
		metrics.IncrCounter([]string{"consul", "rpc", "raft_handoff"}, 1)
		s.raftLayer.Handoff(conn)

	case rpcTLS:
		if s.rpcTLS == nil {
			s.logger.Printf("[WARN] consul.rpc: TLS connection attempted, server not configured for TLS %s", logConn(conn))
			conn.Close()
			return
		}
		conn = tls.Server(conn, s.rpcTLS)
		s.handleConn(conn, true)

	case rpcMultiplexV2:
		s.handleMultiplexV2(conn)

	case rpcSnapshot:
		s.handleSnapshotConn(conn)

	default:
		s.logger.Printf("[ERR] consul.rpc: unrecognized RPC byte: %v %s", buf[0], logConn(conn))
		conn.Close()
		return
	}
}

// handleMultiplexV2 is used to multiplex a single incoming connection
// using the Yamux multiplexer
func (s *Server) handleMultiplexV2(conn net.Conn) {
	defer conn.Close()
	conf := yamux.DefaultConfig()
	conf.LogOutput = s.config.LogOutput
	server, _ := yamux.Server(conn, conf)
	for {
		sub, err := server.Accept()
		if err != nil {
			if err != io.EOF {
				s.logger.Printf("[ERR] consul.rpc: multiplex conn accept failed: %v %s", err, logConn(conn))
			}
			return
		}
		go s.handleConsulConn(sub)
	}
}

// handleConsulConn is used to service a single Consul RPC connection
func (s *Server) handleConsulConn(conn net.Conn) {
	defer conn.Close()
	rpcCodec := msgpackrpc.NewServerCodec(conn)
	for {
		select {
		case <-s.shutdownCh:
			return
		default:
		}

		if err := s.rpcServer.ServeRequest(rpcCodec); err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "closed") {
				s.logger.Printf("[ERR] consul.rpc: RPC error: %v %s", err, logConn(conn))
				metrics.IncrCounter([]string{"consul", "rpc", "request_error"}, 1)
			}
			return
		}
		metrics.IncrCounter([]string{"consul", "rpc", "request"}, 1)
	}
}

// handleSnapshotConn is used to dispatch snapshot saves and restores, which
// stream so don't use the normal RPC mechanism.
func (s *Server) handleSnapshotConn(conn net.Conn) {
	go func() {
		defer conn.Close()
		if err := s.handleSnapshotRequest(conn); err != nil {
			s.logger.Printf("[ERR] consul.rpc: Snapshot RPC error: %v %s", err, logConn(conn))
		}
	}()
}

// forward is used to forward to a remote DC or to forward to the local leader
// Returns a bool of if forwarding was performed, as well as any error
func (s *Server) forward(method string, info structs.RPCInfo, args interface{}, reply interface{}) (bool, error) {
	var firstCheck time.Time

	// Handle DC forwarding
	dc := info.RequestDatacenter()
	if dc != s.config.Datacenter {
		err := s.forwardDC(method, dc, args, reply)
		return true, err
	}

	// Check if we can allow a stale read
	if info.IsRead() && info.AllowStaleRead() {
		return false, nil
	}

CHECK_LEADER:
	// Find the leader
	isLeader, remoteServer := s.getLeader()

	// Handle the case we are the leader
	if isLeader {
		return false, nil
	}

	// Handle the case of a known leader
	if remoteServer != nil {
		err := s.forwardLeader(remoteServer, method, args, reply)
		return true, err
	}

	// Gate the request until there is a leader
	if firstCheck.IsZero() {
		firstCheck = time.Now()
	}
	if time.Now().Sub(firstCheck) < s.config.RPCHoldTimeout {
		jitter := lib.RandomStagger(s.config.RPCHoldTimeout / jitterFraction)
		select {
		case <-time.After(jitter):
			goto CHECK_LEADER
		case <-s.shutdownCh:
		}
	}

	// No leader found and hold time exceeded
	return true, structs.ErrNoLeader
}

// getLeader returns if the current node is the leader, and if not then it
// returns the leader which is potentially nil if the cluster has not yet
// elected a leader.
func (s *Server) getLeader() (bool, *agent.Server) {
	// Check if we are the leader
	if s.IsLeader() {
		return true, nil
	}

	// Get the leader
	leader := s.raft.Leader()
	if leader == "" {
		return false, nil
	}

	// Lookup the server
	s.localLock.RLock()
	server := s.localConsuls[leader]
	s.localLock.RUnlock()

	// Server could be nil
	return false, server
}

// forwardLeader is used to forward an RPC call to the leader, or fail if no leader
func (s *Server) forwardLeader(server *agent.Server, method string, args interface{}, reply interface{}) error {
	// Handle a missing server
	if server == nil {
		return structs.ErrNoLeader
	}
	return s.connPool.RPC(s.config.Datacenter, server.Addr, server.Version, method, args, reply)
}

// forwardDC is used to forward an RPC call to a remote DC, or fail if no servers
func (s *Server) forwardDC(method, dc string, args interface{}, reply interface{}) error {
	manager, server, ok := s.router.FindRoute(dc)
	if !ok {
		s.logger.Printf("[WARN] consul.rpc: RPC request for DC %q, no path found", dc)
		return structs.ErrNoDCPath
	}

	metrics.IncrCounter([]string{"consul", "rpc", "cross-dc", dc}, 1)
	if err := s.connPool.RPC(dc, server.Addr, server.Version, method, args, reply); err != nil {
		manager.NotifyFailedServer(server)
		s.logger.Printf("[ERR] consul: RPC failed to server %s in DC %q: %v", server.Addr, dc, err)
		return err
	}

	return nil
}

// globalRPC is used to forward an RPC request to one server in each datacenter.
// This will only error for RPC-related errors. Otherwise, application-level
// errors can be sent in the response objects.
func (s *Server) globalRPC(method string, args interface{},
	reply structs.CompoundResponse) error {

	errorCh := make(chan error)
	respCh := make(chan interface{})

	// Make a new request into each datacenter
	dcs := s.router.GetDatacenters()
	for _, dc := range dcs {
		go func(dc string) {
			rr := reply.New()
			if err := s.forwardDC(method, dc, args, &rr); err != nil {
				errorCh <- err
				return
			}
			respCh <- rr
		}(dc)
	}

	replies, total := 0, len(dcs)
	for replies < total {
		select {
		case err := <-errorCh:
			return err
		case rr := <-respCh:
			reply.Add(rr)
			replies++
		}
	}
	return nil
}

// raftApply is used to encode a message, run it through raft, and return
// the FSM response along with any errors
func (s *Server) raftApply(t structs.MessageType, msg interface{}) (interface{}, error) {
	buf, err := structs.Encode(t, msg)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode request: %v", err)
	}

	// Warn if the command is very large
	if n := len(buf); n > raftWarnSize {
		s.logger.Printf("[WARN] consul: Attempting to apply large raft entry (%d bytes)", n)
	}

	future := s.raft.Apply(buf, enqueueLimit)
	if err := future.Error(); err != nil {
		return nil, err
	}

	return future.Response(), nil
}

// queryFn is used to perform a query operation. If a re-query is needed, the
// passed-in watch set will be used to block for changes. The passed-in state
// store should be used (vs. calling fsm.State()) since the given state store
// will be correctly watched for changes if the state store is restored from
// a snapshot.
type queryFn func(memdb.WatchSet, *state.StateStore) error

// blockingQuery is used to process a potentially blocking query operation.
func (s *Server) blockingQuery(queryOpts *structs.QueryOptions, queryMeta *structs.QueryMeta,
	fn queryFn) error {
	var timeout *time.Timer

	// Fast path right to the non-blocking query.
	if queryOpts.MinQueryIndex == 0 {
		goto RUN_QUERY
	}

	// Restrict the max query time, and ensure there is always one.
	if queryOpts.MaxQueryTime > maxQueryTime {
		queryOpts.MaxQueryTime = maxQueryTime
	} else if queryOpts.MaxQueryTime <= 0 {
		queryOpts.MaxQueryTime = defaultQueryTime
	}

	// Apply a small amount of jitter to the request.
	queryOpts.MaxQueryTime += lib.RandomStagger(queryOpts.MaxQueryTime / jitterFraction)

	// Setup a query timeout.
	timeout = time.NewTimer(queryOpts.MaxQueryTime)
	defer timeout.Stop()

RUN_QUERY:
	// Update the query metadata.
	s.setQueryMeta(queryMeta)

	// If the read must be consistent we verify that we are still the leader.
	if queryOpts.RequireConsistent {
		if err := s.consistentRead(); err != nil {
			return err
		}
	}

	// Run the query.
	metrics.IncrCounter([]string{"consul", "rpc", "query"}, 1)

	// Operate on a consistent set of state. This makes sure that the
	// abandon channel goes with the state that the caller is using to
	// build watches.
	state := s.fsm.State()

	// We can skip all watch tracking if this isn't a blocking query.
	var ws memdb.WatchSet
	if queryOpts.MinQueryIndex > 0 {
		ws = memdb.NewWatchSet()

		// This channel will be closed if a snapshot is restored and the
		// whole state store is abandoned.
		ws.Add(state.AbandonCh())
	}

	// Block up to the timeout if we didn't see anything fresh.
	err := fn(ws, state)
	if err == nil && queryMeta.Index > 0 && queryMeta.Index <= queryOpts.MinQueryIndex {
		if expired := ws.Watch(timeout.C); !expired {
			// If a restore may have woken us up then bail out from
			// the query immediately. This is slightly race-ey since
			// this might have been interrupted for other reasons,
			// but it's OK to kick it back to the caller in either
			// case.
			select {
			case <-state.AbandonCh():
			default:
				goto RUN_QUERY
			}
		}
	}
	return err
}

// setQueryMeta is used to populate the QueryMeta data for an RPC call
func (s *Server) setQueryMeta(m *structs.QueryMeta) {
	if s.IsLeader() {
		m.LastContact = 0
		m.KnownLeader = true
	} else {
		m.LastContact = time.Now().Sub(s.raft.LastContact())
		m.KnownLeader = (s.raft.Leader() != "")
	}
}

// consistentRead is used to ensure we do not perform a stale
// read. This is done by verifying leadership before the read.
func (s *Server) consistentRead() error {
	defer metrics.MeasureSince([]string{"consul", "rpc", "consistentRead"}, time.Now())
	future := s.raft.VerifyLeader()
	return future.Error()
}
