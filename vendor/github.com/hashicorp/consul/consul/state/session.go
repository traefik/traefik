package state

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
)

// Sessions is used to pull the full list of sessions for use during snapshots.
func (s *StateSnapshot) Sessions() (memdb.ResultIterator, error) {
	iter, err := s.tx.Get("sessions", "id")
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// Session is used when restoring from a snapshot. For general inserts, use
// SessionCreate.
func (s *StateRestore) Session(sess *structs.Session) error {
	// Insert the session.
	if err := s.tx.Insert("sessions", sess); err != nil {
		return fmt.Errorf("failed inserting session: %s", err)
	}

	// Insert the check mappings.
	for _, checkID := range sess.Checks {
		mapping := &sessionCheck{
			Node:    sess.Node,
			CheckID: checkID,
			Session: sess.ID,
		}
		if err := s.tx.Insert("session_checks", mapping); err != nil {
			return fmt.Errorf("failed inserting session check mapping: %s", err)
		}
	}

	// Update the index.
	if err := indexUpdateMaxTxn(s.tx, sess.ModifyIndex, "sessions"); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// SessionCreate is used to register a new session in the state store.
func (s *StateStore) SessionCreate(idx uint64, sess *structs.Session) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// This code is technically able to (incorrectly) update an existing
	// session but we never do that in practice. The upstream endpoint code
	// always adds a unique ID when doing a create operation so we never hit
	// an existing session again. It isn't worth the overhead to verify
	// that here, but it's worth noting that we should never do this in the
	// future.

	// Call the session creation
	if err := s.sessionCreateTxn(tx, idx, sess); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// sessionCreateTxn is the inner method used for creating session entries in
// an open transaction. Any health checks registered with the session will be
// checked for failing status. Returns any error encountered.
func (s *StateStore) sessionCreateTxn(tx *memdb.Txn, idx uint64, sess *structs.Session) error {
	// Check that we have a session ID
	if sess.ID == "" {
		return ErrMissingSessionID
	}

	// Verify the session behavior is valid
	switch sess.Behavior {
	case "":
		// Release by default to preserve backwards compatibility
		sess.Behavior = structs.SessionKeysRelease
	case structs.SessionKeysRelease:
	case structs.SessionKeysDelete:
	default:
		return fmt.Errorf("Invalid session behavior: %s", sess.Behavior)
	}

	// Assign the indexes. ModifyIndex likely will not be used but
	// we set it here anyways for sanity.
	sess.CreateIndex = idx
	sess.ModifyIndex = idx

	// Check that the node exists
	node, err := tx.First("nodes", "id", sess.Node)
	if err != nil {
		return fmt.Errorf("failed node lookup: %s", err)
	}
	if node == nil {
		return ErrMissingNode
	}

	// Go over the session checks and ensure they exist.
	for _, checkID := range sess.Checks {
		check, err := tx.First("checks", "id", sess.Node, string(checkID))
		if err != nil {
			return fmt.Errorf("failed check lookup: %s", err)
		}
		if check == nil {
			return fmt.Errorf("Missing check '%s' registration", checkID)
		}

		// Check that the check is not in critical state
		status := check.(*structs.HealthCheck).Status
		if status == structs.HealthCritical {
			return fmt.Errorf("Check '%s' is in %s state", checkID, status)
		}
	}

	// Insert the session
	if err := tx.Insert("sessions", sess); err != nil {
		return fmt.Errorf("failed inserting session: %s", err)
	}

	// Insert the check mappings
	for _, checkID := range sess.Checks {
		mapping := &sessionCheck{
			Node:    sess.Node,
			CheckID: checkID,
			Session: sess.ID,
		}
		if err := tx.Insert("session_checks", mapping); err != nil {
			return fmt.Errorf("failed inserting session check mapping: %s", err)
		}
	}

	// Update the index
	if err := tx.Insert("index", &IndexEntry{"sessions", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// SessionGet is used to retrieve an active session from the state store.
func (s *StateStore) SessionGet(ws memdb.WatchSet, sessionID string) (uint64, *structs.Session, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "sessions")

	// Look up the session by its ID
	watchCh, session, err := tx.FirstWatch("sessions", "id", sessionID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed session lookup: %s", err)
	}
	ws.Add(watchCh)
	if session != nil {
		return idx, session.(*structs.Session), nil
	}
	return idx, nil, nil
}

// SessionList returns a slice containing all of the active sessions.
func (s *StateStore) SessionList(ws memdb.WatchSet) (uint64, structs.Sessions, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "sessions")

	// Query all of the active sessions.
	sessions, err := tx.Get("sessions", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed session lookup: %s", err)
	}
	ws.Add(sessions.WatchCh())

	// Go over the sessions and create a slice of them.
	var result structs.Sessions
	for session := sessions.Next(); session != nil; session = sessions.Next() {
		result = append(result, session.(*structs.Session))
	}
	return idx, result, nil
}

// NodeSessions returns a set of active sessions associated
// with the given node ID. The returned index is the highest
// index seen from the result set.
func (s *StateStore) NodeSessions(ws memdb.WatchSet, nodeID string) (uint64, structs.Sessions, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "sessions")

	// Get all of the sessions which belong to the node
	sessions, err := tx.Get("sessions", "node", nodeID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed session lookup: %s", err)
	}
	ws.Add(sessions.WatchCh())

	// Go over all of the sessions and return them as a slice
	var result structs.Sessions
	for session := sessions.Next(); session != nil; session = sessions.Next() {
		result = append(result, session.(*structs.Session))
	}
	return idx, result, nil
}

// SessionDestroy is used to remove an active session. This will
// implicitly invalidate the session and invoke the specified
// session destroy behavior.
func (s *StateStore) SessionDestroy(idx uint64, sessionID string) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the session deletion.
	if err := s.deleteSessionTxn(tx, idx, sessionID); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// deleteSessionTxn is the inner method, which is used to do the actual
// session deletion and handle session invalidation, etc.
func (s *StateStore) deleteSessionTxn(tx *memdb.Txn, idx uint64, sessionID string) error {
	// Look up the session.
	sess, err := tx.First("sessions", "id", sessionID)
	if err != nil {
		return fmt.Errorf("failed session lookup: %s", err)
	}
	if sess == nil {
		return nil
	}

	// Delete the session and write the new index.
	if err := tx.Delete("sessions", sess); err != nil {
		return fmt.Errorf("failed deleting session: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"sessions", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	// Enforce the max lock delay.
	session := sess.(*structs.Session)
	delay := session.LockDelay
	if delay > structs.MaxLockDelay {
		delay = structs.MaxLockDelay
	}

	// Snag the current now time so that all the expirations get calculated
	// the same way.
	now := time.Now()

	// Get an iterator over all of the keys with the given session.
	entries, err := tx.Get("kvs", "session", sessionID)
	if err != nil {
		return fmt.Errorf("failed kvs lookup: %s", err)
	}
	var kvs []interface{}
	for entry := entries.Next(); entry != nil; entry = entries.Next() {
		kvs = append(kvs, entry)
	}

	// Invalidate any held locks.
	switch session.Behavior {
	case structs.SessionKeysRelease:
		for _, obj := range kvs {
			// Note that we clone here since we are modifying the
			// returned object and want to make sure our set op
			// respects the transaction we are in.
			e := obj.(*structs.DirEntry).Clone()
			e.Session = ""
			if err := s.kvsSetTxn(tx, idx, e, true); err != nil {
				return fmt.Errorf("failed kvs update: %s", err)
			}

			// Apply the lock delay if present.
			if delay > 0 {
				s.lockDelay.SetExpiration(e.Key, now, delay)
			}
		}
	case structs.SessionKeysDelete:
		for _, obj := range kvs {
			e := obj.(*structs.DirEntry)
			if err := s.kvsDeleteTxn(tx, idx, e.Key); err != nil {
				return fmt.Errorf("failed kvs delete: %s", err)
			}

			// Apply the lock delay if present.
			if delay > 0 {
				s.lockDelay.SetExpiration(e.Key, now, delay)
			}
		}
	default:
		return fmt.Errorf("unknown session behavior %#v", session.Behavior)
	}

	// Delete any check mappings.
	mappings, err := tx.Get("session_checks", "session", sessionID)
	if err != nil {
		return fmt.Errorf("failed session checks lookup: %s", err)
	}
	{
		var objs []interface{}
		for mapping := mappings.Next(); mapping != nil; mapping = mappings.Next() {
			objs = append(objs, mapping)
		}

		// Do the delete in a separate loop so we don't trash the iterator.
		for _, obj := range objs {
			if err := tx.Delete("session_checks", obj); err != nil {
				return fmt.Errorf("failed deleting session check: %s", err)
			}
		}
	}

	// Delete any prepared queries.
	queries, err := tx.Get("prepared-queries", "session", sessionID)
	if err != nil {
		return fmt.Errorf("failed prepared query lookup: %s", err)
	}
	{
		var ids []string
		for wrapped := queries.Next(); wrapped != nil; wrapped = queries.Next() {
			ids = append(ids, toPreparedQuery(wrapped).ID)
		}

		// Do the delete in a separate loop so we don't trash the iterator.
		for _, id := range ids {
			if err := s.preparedQueryDeleteTxn(tx, idx, id); err != nil {
				return fmt.Errorf("failed prepared query delete: %s", err)
			}
		}
	}

	return nil
}
