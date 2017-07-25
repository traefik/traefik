package consul

import (
	"fmt"
	"time"

	"github.com/armon/go-metrics"
	"github.com/hashicorp/consul/consul/state"
	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/go-uuid"
)

// Session endpoint is used to manipulate sessions for KV
type Session struct {
	srv *Server
}

// Apply is used to apply a modifying request to the data store. This should
// only be used for operations that modify the data
func (s *Session) Apply(args *structs.SessionRequest, reply *string) error {
	if done, err := s.srv.forward("Session.Apply", args, args, reply); done {
		return err
	}
	defer metrics.MeasureSince([]string{"consul", "session", "apply"}, time.Now())

	// Verify the args
	if args.Session.ID == "" && args.Op == structs.SessionDestroy {
		return fmt.Errorf("Must provide ID")
	}
	if args.Session.Node == "" && args.Op == structs.SessionCreate {
		return fmt.Errorf("Must provide Node")
	}

	// Fetch the ACL token, if any, and apply the policy.
	acl, err := s.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && s.srv.config.ACLEnforceVersion8 {
		switch args.Op {
		case structs.SessionDestroy:
			state := s.srv.fsm.State()
			_, existing, err := state.SessionGet(nil, args.Session.ID)
			if err != nil {
				return fmt.Errorf("Unknown session %q", args.Session.ID)
			}
			if !acl.SessionWrite(existing.Node) {
				return permissionDeniedErr
			}

		case structs.SessionCreate:
			if !acl.SessionWrite(args.Session.Node) {
				return permissionDeniedErr
			}

		default:
			return fmt.Errorf("Invalid session operation %q", args.Op)
		}
	}

	// Ensure that the specified behavior is allowed
	switch args.Session.Behavior {
	case "":
		// Default behavior to Release for backwards compatibility
		args.Session.Behavior = structs.SessionKeysRelease
	case structs.SessionKeysRelease:
	case structs.SessionKeysDelete:
	default:
		return fmt.Errorf("Invalid Behavior setting '%s'", args.Session.Behavior)
	}

	// Ensure the Session TTL is valid if provided
	if args.Session.TTL != "" {
		ttl, err := time.ParseDuration(args.Session.TTL)
		if err != nil {
			return fmt.Errorf("Session TTL '%s' invalid: %v", args.Session.TTL, err)
		}

		if ttl != 0 && (ttl < s.srv.config.SessionTTLMin || ttl > structs.SessionTTLMax) {
			return fmt.Errorf("Invalid Session TTL '%d', must be between [%v=%v]",
				ttl, s.srv.config.SessionTTLMin, structs.SessionTTLMax)
		}
	}

	// If this is a create, we must generate the Session ID. This must
	// be done prior to appending to the raft log, because the ID is not
	// deterministic. Once the entry is in the log, the state update MUST
	// be deterministic or the followers will not converge.
	if args.Op == structs.SessionCreate {
		// Generate a new session ID, verify uniqueness
		state := s.srv.fsm.State()
		for {
			var err error
			if args.Session.ID, err = uuid.GenerateUUID(); err != nil {
				s.srv.logger.Printf("[ERR] consul.session: UUID generation failed: %v", err)
				return err
			}
			_, sess, err := state.SessionGet(nil, args.Session.ID)
			if err != nil {
				s.srv.logger.Printf("[ERR] consul.session: Session lookup failed: %v", err)
				return err
			}
			if sess == nil {
				break
			}
		}
	}

	// Apply the update
	resp, err := s.srv.raftApply(structs.SessionRequestType, args)
	if err != nil {
		s.srv.logger.Printf("[ERR] consul.session: Apply failed: %v", err)
		return err
	}

	if args.Op == structs.SessionCreate && args.Session.TTL != "" {
		// If we created a session with a TTL, reset the expiration timer
		s.srv.resetSessionTimer(args.Session.ID, &args.Session)
	} else if args.Op == structs.SessionDestroy {
		// If we destroyed a session, it might potentially have a TTL,
		// and we need to clear the timer
		s.srv.clearSessionTimer(args.Session.ID)
	}

	if respErr, ok := resp.(error); ok {
		return respErr
	}

	// Check if the return type is a string
	if respString, ok := resp.(string); ok {
		*reply = respString
	}
	return nil
}

// Get is used to retrieve a single session
func (s *Session) Get(args *structs.SessionSpecificRequest,
	reply *structs.IndexedSessions) error {
	if done, err := s.srv.forward("Session.Get", args, args, reply); done {
		return err
	}

	return s.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			index, session, err := state.SessionGet(ws, args.Session)
			if err != nil {
				return err
			}

			reply.Index = index
			if session != nil {
				reply.Sessions = structs.Sessions{session}
			} else {
				reply.Sessions = nil
			}
			if err := s.srv.filterACL(args.Token, reply); err != nil {
				return err
			}
			return nil
		})
}

// List is used to list all the active sessions
func (s *Session) List(args *structs.DCSpecificRequest,
	reply *structs.IndexedSessions) error {
	if done, err := s.srv.forward("Session.List", args, args, reply); done {
		return err
	}

	return s.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			index, sessions, err := state.SessionList(ws)
			if err != nil {
				return err
			}

			reply.Index, reply.Sessions = index, sessions
			if err := s.srv.filterACL(args.Token, reply); err != nil {
				return err
			}
			return nil
		})
}

// NodeSessions is used to get all the sessions for a particular node
func (s *Session) NodeSessions(args *structs.NodeSpecificRequest,
	reply *structs.IndexedSessions) error {
	if done, err := s.srv.forward("Session.NodeSessions", args, args, reply); done {
		return err
	}

	return s.srv.blockingQuery(
		&args.QueryOptions,
		&reply.QueryMeta,
		func(ws memdb.WatchSet, state *state.StateStore) error {
			index, sessions, err := state.NodeSessions(ws, args.Node)
			if err != nil {
				return err
			}

			reply.Index, reply.Sessions = index, sessions
			if err := s.srv.filterACL(args.Token, reply); err != nil {
				return err
			}
			return nil
		})
}

// Renew is used to renew the TTL on a single session
func (s *Session) Renew(args *structs.SessionSpecificRequest,
	reply *structs.IndexedSessions) error {
	if done, err := s.srv.forward("Session.Renew", args, args, reply); done {
		return err
	}
	defer metrics.MeasureSince([]string{"consul", "session", "renew"}, time.Now())

	// Get the session, from local state.
	state := s.srv.fsm.State()
	index, session, err := state.SessionGet(nil, args.Session)
	if err != nil {
		return err
	}

	reply.Index = index
	if session == nil {
		return nil
	}

	// Fetch the ACL token, if any, and apply the policy.
	acl, err := s.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && s.srv.config.ACLEnforceVersion8 {
		if !acl.SessionWrite(session.Node) {
			return permissionDeniedErr
		}
	}

	// Reset the session TTL timer.
	reply.Sessions = structs.Sessions{session}
	if err := s.srv.resetSessionTimer(args.Session, session); err != nil {
		s.srv.logger.Printf("[ERR] consul.session: Session renew failed: %v", err)
		return err
	}

	return nil
}
