package state

import (
	"fmt"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
)

// ACLs is used to pull all the ACLs from the snapshot.
func (s *StateSnapshot) ACLs() (memdb.ResultIterator, error) {
	iter, err := s.tx.Get("acls", "id")
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// ACL is used when restoring from a snapshot. For general inserts, use ACLSet.
func (s *StateRestore) ACL(acl *structs.ACL) error {
	if err := s.tx.Insert("acls", acl); err != nil {
		return fmt.Errorf("failed restoring acl: %s", err)
	}

	if err := indexUpdateMaxTxn(s.tx, acl.ModifyIndex, "acls"); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// ACLSet is used to insert an ACL rule into the state store.
func (s *StateStore) ACLSet(idx uint64, acl *structs.ACL) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call set on the ACL
	if err := s.aclSetTxn(tx, idx, acl); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// aclSetTxn is the inner method used to insert an ACL rule with the
// proper indexes into the state store.
func (s *StateStore) aclSetTxn(tx *memdb.Txn, idx uint64, acl *structs.ACL) error {
	// Check that the ID is set
	if acl.ID == "" {
		return ErrMissingACLID
	}

	// Check for an existing ACL
	existing, err := tx.First("acls", "id", acl.ID)
	if err != nil {
		return fmt.Errorf("failed acl lookup: %s", err)
	}

	// Set the indexes
	if existing != nil {
		acl.CreateIndex = existing.(*structs.ACL).CreateIndex
		acl.ModifyIndex = idx
	} else {
		acl.CreateIndex = idx
		acl.ModifyIndex = idx
	}

	// Insert the ACL
	if err := tx.Insert("acls", acl); err != nil {
		return fmt.Errorf("failed inserting acl: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"acls", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// ACLGet is used to look up an existing ACL by ID.
func (s *StateStore) ACLGet(ws memdb.WatchSet, aclID string) (uint64, *structs.ACL, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "acls")

	// Query for the existing ACL
	watchCh, acl, err := tx.FirstWatch("acls", "id", aclID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed acl lookup: %s", err)
	}
	ws.Add(watchCh)

	if acl != nil {
		return idx, acl.(*structs.ACL), nil
	}
	return idx, nil, nil
}

// ACLList is used to list out all of the ACLs in the state store.
func (s *StateStore) ACLList(ws memdb.WatchSet) (uint64, structs.ACLs, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "acls")

	// Return the ACLs.
	acls, err := s.aclListTxn(tx, ws)
	if err != nil {
		return 0, nil, fmt.Errorf("failed acl lookup: %s", err)
	}
	return idx, acls, nil
}

// aclListTxn is used to list out all of the ACLs in the state store. This is a
// function vs. a method so it can be called from the snapshotter.
func (s *StateStore) aclListTxn(tx *memdb.Txn, ws memdb.WatchSet) (structs.ACLs, error) {
	// Query all of the ACLs in the state store
	iter, err := tx.Get("acls", "id")
	if err != nil {
		return nil, fmt.Errorf("failed acl lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	// Go over all of the ACLs and build the response
	var result structs.ACLs
	for acl := iter.Next(); acl != nil; acl = iter.Next() {
		a := acl.(*structs.ACL)
		result = append(result, a)
	}
	return result, nil
}

// ACLDelete is used to remove an existing ACL from the state store. If
// the ACL does not exist this is a no-op and no error is returned.
func (s *StateStore) ACLDelete(idx uint64, aclID string) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Call the ACL delete
	if err := s.aclDeleteTxn(tx, idx, aclID); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// aclDeleteTxn is used to delete an ACL from the state store within
// an existing transaction.
func (s *StateStore) aclDeleteTxn(tx *memdb.Txn, idx uint64, aclID string) error {
	// Look up the existing ACL
	acl, err := tx.First("acls", "id", aclID)
	if err != nil {
		return fmt.Errorf("failed acl lookup: %s", err)
	}
	if acl == nil {
		return nil
	}

	// Delete the ACL from the state store and update indexes
	if err := tx.Delete("acls", acl); err != nil {
		return fmt.Errorf("failed deleting acl: %s", err)
	}
	if err := tx.Insert("index", &IndexEntry{"acls", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}
