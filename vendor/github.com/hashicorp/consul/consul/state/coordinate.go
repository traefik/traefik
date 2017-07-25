package state

import (
	"fmt"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/serf/coordinate"
)

// Coordinates is used to pull all the coordinates from the snapshot.
func (s *StateSnapshot) Coordinates() (memdb.ResultIterator, error) {
	iter, err := s.tx.Get("coordinates", "id")
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// Coordinates is used when restoring from a snapshot. For general inserts, use
// CoordinateBatchUpdate. We do less vetting of the updates here because they
// already got checked on the way in during a batch update.
func (s *StateRestore) Coordinates(idx uint64, updates structs.Coordinates) error {
	for _, update := range updates {
		if err := s.tx.Insert("coordinates", update); err != nil {
			return fmt.Errorf("failed restoring coordinate: %s", err)
		}
	}

	if err := indexUpdateMaxTxn(s.tx, idx, "coordinates"); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	return nil
}

// CoordinateGetRaw queries for the coordinate of the given node. This is an
// unusual state store method because it just returns the raw coordinate or
// nil, none of the Raft or node information is returned. This hits the 90%
// internal-to-Consul use case for this data, and this isn't exposed via an
// endpoint, so it doesn't matter that the Raft info isn't available.
func (s *StateStore) CoordinateGetRaw(node string) (*coordinate.Coordinate, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Pull the full coordinate entry.
	coord, err := tx.First("coordinates", "id", node)
	if err != nil {
		return nil, fmt.Errorf("failed coordinate lookup: %s", err)
	}

	// Pick out just the raw coordinate.
	if coord != nil {
		return coord.(*structs.Coordinate).Coord, nil
	}
	return nil, nil
}

// Coordinates queries for all nodes with coordinates.
func (s *StateStore) Coordinates(ws memdb.WatchSet) (uint64, structs.Coordinates, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the table index.
	idx := maxIndexTxn(tx, "coordinates")

	// Pull all the coordinates.
	iter, err := tx.Get("coordinates", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed coordinate lookup: %s", err)
	}
	ws.Add(iter.WatchCh())

	var results structs.Coordinates
	for coord := iter.Next(); coord != nil; coord = iter.Next() {
		results = append(results, coord.(*structs.Coordinate))
	}
	return idx, results, nil
}

// CoordinateBatchUpdate processes a batch of coordinate updates and applies
// them in a single transaction.
func (s *StateStore) CoordinateBatchUpdate(idx uint64, updates structs.Coordinates) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Upsert the coordinates.
	for _, update := range updates {
		// Since the cleanup of coordinates is tied to deletion of
		// nodes, we silently drop any updates for nodes that we don't
		// know about. This might be possible during normal operation
		// if we happen to get a coordinate update for a node that
		// hasn't been able to add itself to the catalog yet. Since we
		// don't carefully sequence this, and since it will fix itself
		// on the next coordinate update from that node, we don't return
		// an error or log anything.
		node, err := tx.First("nodes", "id", update.Node)
		if err != nil {
			return fmt.Errorf("failed node lookup: %s", err)
		}
		if node == nil {
			continue
		}

		if err := tx.Insert("coordinates", update); err != nil {
			return fmt.Errorf("failed inserting coordinate: %s", err)
		}
	}

	// Update the index.
	if err := tx.Insert("index", &IndexEntry{"coordinates", idx}); err != nil {
		return fmt.Errorf("failed updating index: %s", err)
	}

	tx.Commit()
	return nil
}
