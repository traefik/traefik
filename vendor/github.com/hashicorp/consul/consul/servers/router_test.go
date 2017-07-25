package servers

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/serf/coordinate"
	"github.com/hashicorp/serf/serf"
)

type mockCluster struct {
	self    string
	members []serf.Member
	coords  map[string]*coordinate.Coordinate
	addr    int
}

func newMockCluster(self string) *mockCluster {
	return &mockCluster{
		self:   self,
		coords: make(map[string]*coordinate.Coordinate),
		addr:   1,
	}
}

func (m *mockCluster) NumNodes() int {
	return len(m.members)
}

func (m *mockCluster) Members() []serf.Member {
	return m.members
}

func (m *mockCluster) GetCoordinate() (*coordinate.Coordinate, error) {
	return m.coords[m.self], nil
}

func (m *mockCluster) GetCachedCoordinate(name string) (*coordinate.Coordinate, bool) {
	coord, ok := m.coords[name]
	return coord, ok
}

func (m *mockCluster) AddMember(dc string, name string, coord *coordinate.Coordinate) {
	member := serf.Member{
		Name: fmt.Sprintf("%s.%s", name, dc),
		Addr: net.ParseIP(fmt.Sprintf("127.0.0.%d", m.addr)),
		Port: 8300,
		Tags: map[string]string{
			"dc":    dc,
			"role":  "consul",
			"port":  "8300",
			"build": "0.8.0",
			"vsn":   "3",
		},
	}
	m.members = append(m.members, member)
	if coord != nil {
		m.coords[member.Name] = coord
	}
	m.addr++
}

// testCluster is used to generate a single WAN-like area with a known set of
// member and RTT topology.
//
// Here's the layout of the nodes:
//
//            /----   dc1         ----\         /-  dc2  -\ /-  dc0  -\
//             node2 node1       node3             node1       node0
//   |     |     |     |     |     |     |     |     |     |     |
//   0     1     2     3     4     5     6     7     8     9     10  (ms)
//
// We also include a node4 in dc1 with no known coordinate, as well as a
// mysterious dcX with no nodes with known coordinates.
func testCluster(self string) *mockCluster {
	c := newMockCluster(self)
	c.AddMember("dc0", "node0", lib.GenerateCoordinate(10*time.Millisecond))
	c.AddMember("dc1", "node1", lib.GenerateCoordinate(3*time.Millisecond))
	c.AddMember("dc1", "node2", lib.GenerateCoordinate(2*time.Millisecond))
	c.AddMember("dc1", "node3", lib.GenerateCoordinate(5*time.Millisecond))
	c.AddMember("dc1", "node4", nil)
	c.AddMember("dc2", "node1", lib.GenerateCoordinate(8*time.Millisecond))
	c.AddMember("dcX", "node1", nil)
	return c
}

func testRouter(dc string) *Router {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	shutdownCh := make(chan struct{})
	return NewRouter(logger, shutdownCh, dc)
}

func TestRouter_Routing(t *testing.T) {
	r := testRouter("dc0")

	// Create a WAN-looking area.
	self := "node0.dc0"
	wan := testCluster(self)
	if err := r.AddArea(types.AreaWAN, wan, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Adding the area should enable all the routes right away.
	if _, _, ok := r.FindRoute("dc0"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc1"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc2"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dcX"); !ok {
		t.Fatalf("bad")
	}

	// This hasn't been added yet.
	if _, _, ok := r.FindRoute("dcY"); ok {
		t.Fatalf("bad")
	}

	// Add another area.
	otherID := types.AreaID("other")
	other := newMockCluster(self)
	other.AddMember("dc0", "node0", nil)
	other.AddMember("dc1", "node1", nil)
	other.AddMember("dcY", "node1", nil)
	if err := r.AddArea(otherID, other, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now we should have a route to every DC.
	if _, _, ok := r.FindRoute("dc0"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc1"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc2"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dcX"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dcY"); !ok {
		t.Fatalf("bad")
	}

	// Get the route for dcY and then fail the server. This will still
	// give the server back since we have no other choice.
	_, s, ok := r.FindRoute("dcY")
	if !ok {
		t.Fatalf("bad")
	}
	if err := r.FailServer(otherID, s); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, _, ok := r.FindRoute("dcY"); !ok {
		t.Fatalf("bad")
	}

	// But if we remove the server we won't get a route.
	if err := r.RemoveServer(otherID, s); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, _, ok := r.FindRoute("dcY"); ok {
		t.Fatalf("bad")
	}

	// Make sure the dcY manager also got removed from the area and from
	// the index we use for routing.
	func() {
		r.RLock()
		defer r.RUnlock()

		area, ok := r.areas[otherID]
		if !ok {
			t.Fatalf("bad")
		}

		if _, ok := area.managers["dcY"]; ok {
			t.Fatalf("bad")
		}

		if _, ok := r.managers["dcY"]; ok {
			t.Fatalf("bad")
		}
	}()

	// Do similar for dc0, which will take two removes because the dc0 is
	// reachable from two different areas.
	_, s, ok = r.FindRoute("dc0")
	if !ok {
		t.Fatalf("bad")
	}
	if err := r.RemoveServer(types.AreaWAN, s); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, _, ok = r.FindRoute("dc0"); !ok {
		t.Fatalf("bad")
	}
	if err := r.RemoveServer(otherID, s); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, _, ok = r.FindRoute("dc0"); ok {
		t.Fatalf("bad")
	}

	// Now delete some areas.
	if _, _, ok = r.FindRoute("dc1"); !ok {
		t.Fatalf("bad")
	}
	if err := r.RemoveArea(types.AreaWAN); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, _, ok = r.FindRoute("dc1"); !ok {
		t.Fatalf("bad")
	}
	if err := r.RemoveArea(otherID); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, _, ok = r.FindRoute("dc1"); ok {
		t.Fatalf("bad")
	}
}

func TestRouter_Routing_Offline(t *testing.T) {
	r := testRouter("dc0")

	// Create a WAN-looking area.
	self := "node0.dc0"
	wan := testCluster(self)
	if err := r.AddArea(types.AreaWAN, wan, &fauxConnPool{1.0}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Adding the area should enable all the routes right away.
	if _, _, ok := r.FindRoute("dc0"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc1"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc2"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dcX"); !ok {
		t.Fatalf("bad")
	}

	// Do a rebalance for dc1, which should knock it offline.
	func() {
		r.Lock()
		defer r.Unlock()

		area, ok := r.areas[types.AreaWAN]
		if !ok {
			t.Fatalf("bad")
		}

		info, ok := area.managers["dc1"]
		if !ok {
			t.Fatalf("bad")
		}
		info.manager.RebalanceServers()
	}()

	// Recheck all the routes.
	if _, _, ok := r.FindRoute("dc0"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc1"); ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc2"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dcX"); !ok {
		t.Fatalf("bad")
	}

	// Add another area with a route to dc1.
	otherID := types.AreaID("other")
	other := newMockCluster(self)
	other.AddMember("dc0", "node0", nil)
	other.AddMember("dc1", "node1", nil)
	if err := r.AddArea(otherID, other, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Recheck all the routes and make sure it finds the one that's
	// online.
	if _, _, ok := r.FindRoute("dc0"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc1"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dc2"); !ok {
		t.Fatalf("bad")
	}
	if _, _, ok := r.FindRoute("dcX"); !ok {
		t.Fatalf("bad")
	}
}

func TestRouter_GetDatacenters(t *testing.T) {
	r := testRouter("dc0")

	self := "node0.dc0"
	wan := testCluster(self)
	if err := r.AddArea(types.AreaWAN, wan, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	actual := r.GetDatacenters()
	expected := []string{"dc0", "dc1", "dc2", "dcX"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestRouter_distanceSorter(t *testing.T) {
	actual := &datacenterSorter{
		Names: []string{"foo", "bar", "baz", "zoo"},
		Vec:   []float64{3.0, 1.0, 1.0, 0.0},
	}
	sort.Stable(actual)
	expected := &datacenterSorter{
		Names: []string{"zoo", "bar", "baz", "foo"},
		Vec:   []float64{0.0, 1.0, 1.0, 3.0},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", *expected)
	}
}

func TestRouter_GetDatacentersByDistance(t *testing.T) {
	r := testRouter("dc0")

	// Start with just the WAN area described in the diagram above.
	self := "node0.dc0"
	wan := testCluster(self)
	if err := r.AddArea(types.AreaWAN, wan, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	actual, err := r.GetDatacentersByDistance()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected := []string{"dc0", "dc2", "dc1", "dcX"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}

	// Now add another area with a closer route for dc1.
	otherID := types.AreaID("other")
	other := newMockCluster(self)
	other.AddMember("dc0", "node0", lib.GenerateCoordinate(20*time.Millisecond))
	other.AddMember("dc1", "node1", lib.GenerateCoordinate(21*time.Millisecond))
	if err := r.AddArea(otherID, other, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	actual, err = r.GetDatacentersByDistance()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected = []string{"dc0", "dc1", "dc2", "dcX"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestRouter_GetDatacenterMaps(t *testing.T) {
	r := testRouter("dc0")

	self := "node0.dc0"
	wan := testCluster(self)
	if err := r.AddArea(types.AreaWAN, wan, &fauxConnPool{}); err != nil {
		t.Fatalf("err: %v", err)
	}

	actual, err := r.GetDatacenterMaps()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(actual) != 3 {
		t.Fatalf("bad: %#v", actual)
	}
	for _, entry := range actual {
		switch entry.Datacenter {
		case "dc0":
			if !reflect.DeepEqual(entry, structs.DatacenterMap{
				Datacenter: "dc0",
				AreaID:     types.AreaWAN,
				Coordinates: structs.Coordinates{
					&structs.Coordinate{
						Node:  "node0.dc0",
						Coord: lib.GenerateCoordinate(10 * time.Millisecond),
					},
				},
			}) {
				t.Fatalf("bad: %#v", entry)
			}
		case "dc1":
			if !reflect.DeepEqual(entry, structs.DatacenterMap{
				Datacenter: "dc1",
				AreaID:     types.AreaWAN,
				Coordinates: structs.Coordinates{
					&structs.Coordinate{
						Node:  "node1.dc1",
						Coord: lib.GenerateCoordinate(3 * time.Millisecond),
					},
					&structs.Coordinate{
						Node:  "node2.dc1",
						Coord: lib.GenerateCoordinate(2 * time.Millisecond),
					},
					&structs.Coordinate{
						Node:  "node3.dc1",
						Coord: lib.GenerateCoordinate(5 * time.Millisecond),
					},
				},
			}) {
				t.Fatalf("bad: %#v", entry)
			}
		case "dc2":
			if !reflect.DeepEqual(entry, structs.DatacenterMap{
				Datacenter: "dc2",
				AreaID:     types.AreaWAN,
				Coordinates: structs.Coordinates{
					&structs.Coordinate{
						Node:  "node1.dc2",
						Coord: lib.GenerateCoordinate(8 * time.Millisecond),
					},
				},
			}) {
				t.Fatalf("bad: %#v", entry)
			}
		default:
			t.Fatalf("bad: %#v", entry)
		}
	}
}
