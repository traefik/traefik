package consul

import (
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/net-rpc-msgpackrpc"
)

// verifyNodeSort makes sure the order of the nodes in the slice is the same as
// the expected order, expressed as a comma-separated string.
func verifyNodeSort(t *testing.T, nodes structs.Nodes, expected string) {
	vec := make([]string, len(nodes))
	for i, node := range nodes {
		vec[i] = node.Node
	}
	actual := strings.Join(vec, ",")
	if actual != expected {
		t.Fatalf("bad sort: %s != %s", actual, expected)
	}
}

// verifyServiceNodeSort makes sure the order of the nodes in the slice is the
// same as the expected order, expressed as a comma-separated string.
func verifyServiceNodeSort(t *testing.T, nodes structs.ServiceNodes, expected string) {
	vec := make([]string, len(nodes))
	for i, node := range nodes {
		vec[i] = node.Node
	}
	actual := strings.Join(vec, ",")
	if actual != expected {
		t.Fatalf("bad sort: %s != %s", actual, expected)
	}
}

// verifyHealthCheckSort makes sure the order of the nodes in the slice is the
// same as the expected order, expressed as a comma-separated string.
func verifyHealthCheckSort(t *testing.T, checks structs.HealthChecks, expected string) {
	vec := make([]string, len(checks))
	for i, check := range checks {
		vec[i] = check.Node
	}
	actual := strings.Join(vec, ",")
	if actual != expected {
		t.Fatalf("bad sort: %s != %s", actual, expected)
	}
}

// verifyCheckServiceNodeSort makes sure the order of the nodes in the slice is
// the same as the expected order, expressed as a comma-separated string.
func verifyCheckServiceNodeSort(t *testing.T, nodes structs.CheckServiceNodes, expected string) {
	vec := make([]string, len(nodes))
	for i, node := range nodes {
		vec[i] = node.Node.Node
	}
	actual := strings.Join(vec, ",")
	if actual != expected {
		t.Fatalf("bad sort: %s != %s", actual, expected)
	}
}

// seedCoordinates uses the client to set up a set of nodes with a specific
// set of distances from the origin. We also include the server so that we
// can wait for the coordinates to get committed to the Raft log.
//
// Here's the layout of the nodes:
//
//       node3 node2 node5                         node4       node1
//   |     |     |     |     |     |     |     |     |     |     |
//   0     1     2     3     4     5     6     7     8     9     10  (ms)
//
func seedCoordinates(t *testing.T, codec rpc.ClientCodec, server *Server) {
	// Register some nodes.
	for i := 0; i < 5; i++ {
		req := structs.RegisterRequest{
			Datacenter: "dc1",
			Node:       fmt.Sprintf("node%d", i+1),
			Address:    "127.0.0.1",
		}
		var reply struct{}
		if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &req, &reply); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Seed the fixed setup of the nodes.
	updates := []structs.CoordinateUpdateRequest{
		structs.CoordinateUpdateRequest{
			Datacenter: "dc1",
			Node:       "node1",
			Coord:      lib.GenerateCoordinate(10 * time.Millisecond),
		},
		structs.CoordinateUpdateRequest{
			Datacenter: "dc1",
			Node:       "node2",
			Coord:      lib.GenerateCoordinate(2 * time.Millisecond),
		},
		structs.CoordinateUpdateRequest{
			Datacenter: "dc1",
			Node:       "node3",
			Coord:      lib.GenerateCoordinate(1 * time.Millisecond),
		},
		structs.CoordinateUpdateRequest{
			Datacenter: "dc1",
			Node:       "node4",
			Coord:      lib.GenerateCoordinate(8 * time.Millisecond),
		},
		structs.CoordinateUpdateRequest{
			Datacenter: "dc1",
			Node:       "node5",
			Coord:      lib.GenerateCoordinate(3 * time.Millisecond),
		},
	}

	// Apply the updates and wait a while for the batch to get committed to
	// the Raft log.
	for _, update := range updates {
		var out struct{}
		if err := msgpackrpc.CallWithCodec(codec, "Coordinate.Update", &update, &out); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
	time.Sleep(2 * server.config.CoordinateUpdatePeriod)
}

func TestRTT_sortNodesByDistanceFrom(t *testing.T) {
	dir, server := testServer(t)
	defer os.RemoveAll(dir)
	defer server.Shutdown()

	codec := rpcClient(t, server)
	defer codec.Close()
	testutil.WaitForLeader(t, server.RPC, "dc1")

	seedCoordinates(t, codec, server)
	nodes := structs.Nodes{
		&structs.Node{Node: "apple"},
		&structs.Node{Node: "node1"},
		&structs.Node{Node: "node2"},
		&structs.Node{Node: "node3"},
		&structs.Node{Node: "node4"},
		&structs.Node{Node: "node5"},
	}

	// The zero value for the source should not trigger any sorting.
	var source structs.QuerySource
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "apple,node1,node2,node3,node4,node5")

	// Same for a source in some other DC.
	source.Node = "node1"
	source.Datacenter = "dc2"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "apple,node1,node2,node3,node4,node5")

	// Same for a source node in our DC that we have no coordinate for.
	source.Node = "apple"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "apple,node1,node2,node3,node4,node5")

	// Now sort relative to node1, note that apple doesn't have any seeded
	// coordinate info so it should end up at the end, despite its lexical
	// hegemony.
	source.Node = "node1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "node1,node4,node5,node2,node3,apple")
}

func TestRTT_sortNodesByDistanceFrom_Nodes(t *testing.T) {
	dir, server := testServer(t)
	defer os.RemoveAll(dir)
	defer server.Shutdown()

	codec := rpcClient(t, server)
	defer codec.Close()
	testutil.WaitForLeader(t, server.RPC, "dc1")

	seedCoordinates(t, codec, server)
	nodes := structs.Nodes{
		&structs.Node{Node: "apple"},
		&structs.Node{Node: "node1"},
		&structs.Node{Node: "node2"},
		&structs.Node{Node: "node3"},
		&structs.Node{Node: "node4"},
		&structs.Node{Node: "node5"},
	}

	// Now sort relative to node1, note that apple doesn't have any
	// seeded coordinate info so it should end up at the end, despite
	// its lexical hegemony.
	var source structs.QuerySource
	source.Node = "node1"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "node1,node4,node5,node2,node3,apple")

	// Try another sort from node2. Note that node5 and node3 are the
	// same distance away so the stable sort should preserve the order
	// they were in from the previous sort.
	source.Node = "node2"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "node2,node5,node3,node4,node1,apple")

	// Let's exercise the stable sort explicitly to make sure we didn't
	// just get lucky.
	nodes[1], nodes[2] = nodes[2], nodes[1]
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyNodeSort(t, nodes, "node2,node3,node5,node4,node1,apple")
}

func TestRTT_sortNodesByDistanceFrom_ServiceNodes(t *testing.T) {
	dir, server := testServer(t)
	defer os.RemoveAll(dir)
	defer server.Shutdown()

	codec := rpcClient(t, server)
	defer codec.Close()
	testutil.WaitForLeader(t, server.RPC, "dc1")

	seedCoordinates(t, codec, server)
	nodes := structs.ServiceNodes{
		&structs.ServiceNode{Node: "apple"},
		&structs.ServiceNode{Node: "node1"},
		&structs.ServiceNode{Node: "node2"},
		&structs.ServiceNode{Node: "node3"},
		&structs.ServiceNode{Node: "node4"},
		&structs.ServiceNode{Node: "node5"},
	}

	// Now sort relative to node1, note that apple doesn't have any
	// seeded coordinate info so it should end up at the end, despite
	// its lexical hegemony.
	var source structs.QuerySource
	source.Node = "node1"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyServiceNodeSort(t, nodes, "node1,node4,node5,node2,node3,apple")

	// Try another sort from node2. Note that node5 and node3 are the
	// same distance away so the stable sort should preserve the order
	// they were in from the previous sort.
	source.Node = "node2"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyServiceNodeSort(t, nodes, "node2,node5,node3,node4,node1,apple")

	// Let's exercise the stable sort explicitly to make sure we didn't
	// just get lucky.
	nodes[1], nodes[2] = nodes[2], nodes[1]
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyServiceNodeSort(t, nodes, "node2,node3,node5,node4,node1,apple")
}

func TestRTT_sortNodesByDistanceFrom_HealthChecks(t *testing.T) {
	dir, server := testServer(t)
	defer os.RemoveAll(dir)
	defer server.Shutdown()

	codec := rpcClient(t, server)
	defer codec.Close()
	testutil.WaitForLeader(t, server.RPC, "dc1")

	seedCoordinates(t, codec, server)
	checks := structs.HealthChecks{
		&structs.HealthCheck{Node: "apple"},
		&structs.HealthCheck{Node: "node1"},
		&structs.HealthCheck{Node: "node2"},
		&structs.HealthCheck{Node: "node3"},
		&structs.HealthCheck{Node: "node4"},
		&structs.HealthCheck{Node: "node5"},
	}

	// Now sort relative to node1, note that apple doesn't have any
	// seeded coordinate info so it should end up at the end, despite
	// its lexical hegemony.
	var source structs.QuerySource
	source.Node = "node1"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, checks); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyHealthCheckSort(t, checks, "node1,node4,node5,node2,node3,apple")

	// Try another sort from node2. Note that node5 and node3 are the
	// same distance away so the stable sort should preserve the order
	// they were in from the previous sort.
	source.Node = "node2"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, checks); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyHealthCheckSort(t, checks, "node2,node5,node3,node4,node1,apple")

	// Let's exercise the stable sort explicitly to make sure we didn't
	// just get lucky.
	checks[1], checks[2] = checks[2], checks[1]
	if err := server.sortNodesByDistanceFrom(source, checks); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyHealthCheckSort(t, checks, "node2,node3,node5,node4,node1,apple")
}

func TestRTT_sortNodesByDistanceFrom_CheckServiceNodes(t *testing.T) {
	dir, server := testServer(t)
	defer os.RemoveAll(dir)
	defer server.Shutdown()

	codec := rpcClient(t, server)
	defer codec.Close()
	testutil.WaitForLeader(t, server.RPC, "dc1")

	seedCoordinates(t, codec, server)
	nodes := structs.CheckServiceNodes{
		structs.CheckServiceNode{Node: &structs.Node{Node: "apple"}},
		structs.CheckServiceNode{Node: &structs.Node{Node: "node1"}},
		structs.CheckServiceNode{Node: &structs.Node{Node: "node2"}},
		structs.CheckServiceNode{Node: &structs.Node{Node: "node3"}},
		structs.CheckServiceNode{Node: &structs.Node{Node: "node4"}},
		structs.CheckServiceNode{Node: &structs.Node{Node: "node5"}},
	}

	// Now sort relative to node1, note that apple doesn't have any
	// seeded coordinate info so it should end up at the end, despite
	// its lexical hegemony.
	var source structs.QuerySource
	source.Node = "node1"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyCheckServiceNodeSort(t, nodes, "node1,node4,node5,node2,node3,apple")

	// Try another sort from node2. Note that node5 and node3 are the
	// same distance away so the stable sort should preserve the order
	// they were in from the previous sort.
	source.Node = "node2"
	source.Datacenter = "dc1"
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyCheckServiceNodeSort(t, nodes, "node2,node5,node3,node4,node1,apple")

	// Let's exercise the stable sort explicitly to make sure we didn't
	// just get lucky.
	nodes[1], nodes[2] = nodes[2], nodes[1]
	if err := server.sortNodesByDistanceFrom(source, nodes); err != nil {
		t.Fatalf("err: %v", err)
	}
	verifyCheckServiceNodeSort(t, nodes, "node2,node3,node5,node4,node1,apple")
}
