package state

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
	"github.com/hashicorp/serf/coordinate"
)

// generateRandomCoordinate creates a random coordinate. This mucks with the
// underlying structure directly, so it's not really useful for any particular
// position in the network, but it's a good payload to send through to make
// sure things come out the other side or get stored correctly.
func generateRandomCoordinate() *coordinate.Coordinate {
	config := coordinate.DefaultConfig()
	coord := coordinate.NewCoordinate(config)
	for i := range coord.Vec {
		coord.Vec[i] = rand.NormFloat64()
	}
	coord.Error = rand.NormFloat64()
	coord.Adjustment = rand.NormFloat64()
	return coord
}

func TestStateStore_Coordinate_Updates(t *testing.T) {
	s := testStateStore(t)

	// Make sure the coordinates list starts out empty, and that a query for
	// a raw coordinate for a nonexistent node doesn't do anything bad.
	ws := memdb.NewWatchSet()
	idx, coords, err := s.Coordinates(ws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad index: %d", idx)
	}
	if coords != nil {
		t.Fatalf("bad: %#v", coords)
	}
	coord, err := s.CoordinateGetRaw("nope")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if coord != nil {
		t.Fatalf("bad: %#v", coord)
	}

	// Make an update for nodes that don't exist and make sure they get
	// ignored.
	updates := structs.Coordinates{
		&structs.Coordinate{
			Node:  "node1",
			Coord: generateRandomCoordinate(),
		},
		&structs.Coordinate{
			Node:  "node2",
			Coord: generateRandomCoordinate(),
		},
	}
	if err := s.CoordinateBatchUpdate(1, updates); err != nil {
		t.Fatalf("err: %s", err)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Should still be empty, though applying an empty batch does bump
	// the table index.
	ws = memdb.NewWatchSet()
	idx, coords, err = s.Coordinates(ws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 1 {
		t.Fatalf("bad index: %d", idx)
	}
	if coords != nil {
		t.Fatalf("bad: %#v", coords)
	}

	// Register the nodes then do the update again.
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")
	if err := s.CoordinateBatchUpdate(3, updates); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Should go through now.
	ws = memdb.NewWatchSet()
	idx, coords, err = s.Coordinates(ws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
	if !reflect.DeepEqual(coords, updates) {
		t.Fatalf("bad: %#v", coords)
	}

	// Also verify the raw coordinate interface.
	for _, update := range updates {
		coord, err := s.CoordinateGetRaw(update.Node)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !reflect.DeepEqual(coord, update.Coord) {
			t.Fatalf("bad: %#v", coord)
		}
	}

	// Update the coordinate for one of the nodes.
	updates[1].Coord = generateRandomCoordinate()
	if err := s.CoordinateBatchUpdate(4, updates); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Verify it got applied.
	idx, coords, err = s.Coordinates(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
	if !reflect.DeepEqual(coords, updates) {
		t.Fatalf("bad: %#v", coords)
	}

	// And check the raw coordinate version of the same thing.
	for _, update := range updates {
		coord, err := s.CoordinateGetRaw(update.Node)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !reflect.DeepEqual(coord, update.Coord) {
			t.Fatalf("bad: %#v", coord)
		}
	}
}

func TestStateStore_Coordinate_Cleanup(t *testing.T) {
	s := testStateStore(t)

	// Register a node and update its coordinate.
	testRegisterNode(t, s, 1, "node1")
	updates := structs.Coordinates{
		&structs.Coordinate{
			Node:  "node1",
			Coord: generateRandomCoordinate(),
		},
	}
	if err := s.CoordinateBatchUpdate(2, updates); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Make sure it's in there.
	coord, err := s.CoordinateGetRaw("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if !reflect.DeepEqual(coord, updates[0].Coord) {
		t.Fatalf("bad: %#v", coord)
	}

	// Now delete the node.
	if err := s.DeleteNode(3, "node1"); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Make sure the coordinate is gone.
	coord, err = s.CoordinateGetRaw("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if coord != nil {
		t.Fatalf("bad: %#v", coord)
	}

	// Make sure the index got updated.
	idx, coords, err := s.Coordinates(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
	if coords != nil {
		t.Fatalf("bad: %#v", coords)
	}
}

func TestStateStore_Coordinate_Snapshot_Restore(t *testing.T) {
	s := testStateStore(t)

	// Register two nodes and update their coordinates.
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")
	updates := structs.Coordinates{
		&structs.Coordinate{
			Node:  "node1",
			Coord: generateRandomCoordinate(),
		},
		&structs.Coordinate{
			Node:  "node2",
			Coord: generateRandomCoordinate(),
		},
	}
	if err := s.CoordinateBatchUpdate(3, updates); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Snapshot the coordinates.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	trash := structs.Coordinates{
		&structs.Coordinate{
			Node:  "node1",
			Coord: generateRandomCoordinate(),
		},
		&structs.Coordinate{
			Node:  "node2",
			Coord: generateRandomCoordinate(),
		},
	}
	if err := s.CoordinateBatchUpdate(4, trash); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the snapshot.
	if idx := snap.LastIndex(); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
	iter, err := snap.Coordinates()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	var dump structs.Coordinates
	for coord := iter.Next(); coord != nil; coord = iter.Next() {
		dump = append(dump, coord.(*structs.Coordinate))
	}
	if !reflect.DeepEqual(dump, updates) {
		t.Fatalf("bad: %#v", dump)
	}

	// Restore the values into a new state store.
	func() {
		s := testStateStore(t)
		restore := s.Restore()
		if err := restore.Coordinates(5, dump); err != nil {
			t.Fatalf("err: %s", err)
		}
		restore.Commit()

		// Read the restored coordinates back out and verify that they match.
		idx, res, err := s.Coordinates(nil)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 5 {
			t.Fatalf("bad index: %d", idx)
		}
		if !reflect.DeepEqual(res, updates) {
			t.Fatalf("bad: %#v", res)
		}

		// Check that the index was updated (note that it got passed
		// in during the restore).
		if idx := s.maxIndex("coordinates"); idx != 5 {
			t.Fatalf("bad index: %d", idx)
		}
	}()

}
