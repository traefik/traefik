// The snapshot endpoint is a special non-RPC endpoint that supports streaming
// for taking and restoring snapshots for disaster recovery. This gets wired
// directly into Consul's stream handler, and a new TCP connection is made for
// each request.
//
// This also includes a SnapshotRPC() function, which acts as a lightweight
// client that knows the details of the stream protocol.
package consul

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/snapshot"
	"github.com/hashicorp/go-msgpack/codec"
)

// dispatchSnapshotRequest takes an incoming request structure with possibly some
// streaming data (for a restore) and returns possibly some streaming data (for
// a snapshot save). We can't use the normal RPC mechanism in a streaming manner
// like this, so we have to dispatch these by hand.
func (s *Server) dispatchSnapshotRequest(args *structs.SnapshotRequest, in io.Reader,
	reply *structs.SnapshotResponse) (io.ReadCloser, error) {

	// Perform DC forwarding.
	if dc := args.Datacenter; dc != s.config.Datacenter {
		manager, server, ok := s.router.FindRoute(dc)
		if !ok {
			return nil, structs.ErrNoDCPath
		}

		snap, err := SnapshotRPC(s.connPool, dc, server.Addr, args, in, reply)
		if err != nil {
			manager.NotifyFailedServer(server)
			return nil, err
		}

		return snap, nil
	}

	// Perform leader forwarding if required.
	if !args.AllowStale {
		if isLeader, server := s.getLeader(); !isLeader {
			if server == nil {
				return nil, structs.ErrNoLeader
			}
			return SnapshotRPC(s.connPool, args.Datacenter, server.Addr, args, in, reply)
		}
	}

	// Verify token is allowed to operate on snapshots. There's only a
	// single ACL sense here (not read and write) since reading gets you
	// all the ACLs and you could escalate from there.
	if acl, err := s.resolveToken(args.Token); err != nil {
		return nil, err
	} else if acl != nil && !acl.Snapshot() {
		return nil, permissionDeniedErr
	}

	// Dispatch the operation.
	switch args.Op {
	case structs.SnapshotSave:
		if !args.AllowStale {
			if err := s.consistentRead(); err != nil {
				return nil, err
			}
		}

		// Set the metadata here before we do anything; this should always be
		// pessimistic if we get more data while the snapshot is being taken.
		s.setQueryMeta(&reply.QueryMeta)

		// Take the snapshot and capture the index.
		snap, err := snapshot.New(s.logger, s.raft)
		reply.Index = snap.Index()
		return snap, err

	case structs.SnapshotRestore:
		if args.AllowStale {
			return nil, fmt.Errorf("stale not allowed for restore")
		}

		// Restore the snapshot.
		if err := snapshot.Restore(s.logger, in, s.raft); err != nil {
			return nil, err
		}

		// Run a barrier so we are sure that our FSM is caught up with
		// any snapshot restore details (it's also part of Raft's restore
		// process but we don't want to depend on that detail for this to
		// be correct). Once that works, we can redo the leader actions
		// so our leader-maintained state will be up to date.
		barrier := s.raft.Barrier(0)
		if err := barrier.Error(); err != nil {
			return nil, err
		}
		if err := s.revokeLeadership(); err != nil {
			return nil, err
		}
		if err := s.establishLeadership(); err != nil {
			return nil, err
		}

		// Give the caller back an empty reader since there's nothing to
		// stream back.
		return ioutil.NopCloser(bytes.NewReader([]byte(""))), nil

	default:
		return nil, fmt.Errorf("unrecognized snapshot op %q", args.Op)
	}
}

// handleSnapshotRequest reads the request from the conn and dispatches it. This
// will be called from a goroutine after an incoming stream is determined to be
// a snapshot request.
func (s *Server) handleSnapshotRequest(conn net.Conn) error {
	var args structs.SnapshotRequest
	dec := codec.NewDecoder(conn, &codec.MsgpackHandle{})
	if err := dec.Decode(&args); err != nil {
		return fmt.Errorf("failed to decode request: %v", err)
	}

	var reply structs.SnapshotResponse
	snap, err := s.dispatchSnapshotRequest(&args, conn, &reply)
	if err != nil {
		reply.Error = err.Error()
		goto RESPOND
	}
	defer func() {
		if err := snap.Close(); err != nil {
			s.logger.Printf("[ERR] consul: Failed to close snapshot: %v", err)
		}
	}()

RESPOND:
	enc := codec.NewEncoder(conn, &codec.MsgpackHandle{})
	if err := enc.Encode(&reply); err != nil {
		return fmt.Errorf("failed to encode response: %v", err)
	}
	if snap != nil {
		if _, err := io.Copy(conn, snap); err != nil {
			return fmt.Errorf("failed to stream snapshot: %v", err)
		}
	}

	return nil
}

// SnapshotRPC is a streaming client function for performing a snapshot RPC
// request to a remote server. It will create a fresh connection for each
// request, send the request header, and then stream in any data from the
// reader (for a restore). It will then parse the received response header, and
// if there's no error will return an io.ReadCloser (that you must close) with
// the streaming output (for a snapshot). If the reply contains an error, this
// will always return an error as well, so you don't need to check the error
// inside the filled-in reply.
func SnapshotRPC(pool *ConnPool, dc string, addr net.Addr,
	args *structs.SnapshotRequest, in io.Reader, reply *structs.SnapshotResponse) (io.ReadCloser, error) {

	conn, hc, err := pool.DialTimeout(dc, addr, 10*time.Second)
	if err != nil {
		return nil, err
	}

	// keep will disarm the defer on success if we are returning the caller
	// our connection to stream the output.
	var keep bool
	defer func() {
		if !keep {
			conn.Close()
		}
	}()

	// Write the snapshot RPC byte to set the mode, then perform the
	// request.
	if _, err := conn.Write([]byte{byte(rpcSnapshot)}); err != nil {
		return nil, fmt.Errorf("failed to write stream type: %v", err)
	}

	// Push the header encoded as msgpack, then stream the input.
	enc := codec.NewEncoder(conn, &codec.MsgpackHandle{})
	if err := enc.Encode(&args); err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}
	if _, err := io.Copy(conn, in); err != nil {
		return nil, fmt.Errorf("failed to copy snapshot in: %v", err)
	}

	// Our RPC protocol requires support for a half-close in order to signal
	// the other side that they are done reading the stream, since we don't
	// know the size in advance. This saves us from having to buffer just to
	// calculate the size.
	if hc != nil {
		if err := hc.CloseWrite(); err != nil {
			return nil, fmt.Errorf("failed to half close snapshot connection: %v", err)
		}
	} else {
		return nil, fmt.Errorf("snapshot connection requires half-close support")
	}

	// Pull the header decoded as msgpack. The caller can continue to read
	// the conn to stream the remaining data.
	dec := codec.NewDecoder(conn, &codec.MsgpackHandle{})
	if err := dec.Decode(reply); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	keep = true
	return conn, nil
}
