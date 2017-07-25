package state

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/go-memdb"
	uuid "github.com/hashicorp/go-uuid"
)

func makeRandomNodeID(t *testing.T) types.NodeID {
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return types.NodeID(id)
}

func TestStateStore_EnsureRegistration(t *testing.T) {
	s := testStateStore(t)

	// Start with just a node.
	req := &structs.RegisterRequest{
		ID:      makeRandomNodeID(t),
		Node:    "node1",
		Address: "1.2.3.4",
		TaggedAddresses: map[string]string{
			"hello": "world",
		},
		NodeMeta: map[string]string{
			"somekey": "somevalue",
		},
	}
	nodeID := req.ID
	if err := s.EnsureRegistration(1, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Retrieve the node and verify its contents.
	verifyNode := func() {
		_, out, err := s.GetNode("node1")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if out.ID != nodeID ||
			out.Node != "node1" || out.Address != "1.2.3.4" ||
			len(out.TaggedAddresses) != 1 ||
			out.TaggedAddresses["hello"] != "world" ||
			out.Meta["somekey"] != "somevalue" ||
			out.CreateIndex != 1 || out.ModifyIndex != 1 {
			t.Fatalf("bad node returned: %#v", out)
		}
		_, out2, err := s.GetNodeID(nodeID)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !reflect.DeepEqual(out, out2) {
			t.Fatalf("bad node returned: %#v -- %#v", out, out2)
		}
	}
	verifyNode()

	// Add in a service definition.
	req.Service = &structs.NodeService{
		ID:      "redis1",
		Service: "redis",
		Address: "1.1.1.1",
		Port:    8080,
	}
	if err := s.EnsureRegistration(2, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify that the service got registered.
	verifyService := func() {
		idx, out, err := s.NodeServices(nil, "node1")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 2 {
			t.Fatalf("bad index: %d", idx)
		}
		if len(out.Services) != 1 {
			t.Fatalf("bad: %#v", out.Services)
		}
		r := out.Services["redis1"]
		if r == nil || r.ID != "redis1" || r.Service != "redis" ||
			r.Address != "1.1.1.1" || r.Port != 8080 ||
			r.CreateIndex != 2 || r.ModifyIndex != 2 {
			t.Fatalf("bad service returned: %#v", r)
		}

		idx, r, err = s.NodeService("node1", "redis1")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 2 {
			t.Fatalf("bad index: %d", idx)
		}
		if r == nil || r.ID != "redis1" || r.Service != "redis" ||
			r.Address != "1.1.1.1" || r.Port != 8080 ||
			r.CreateIndex != 2 || r.ModifyIndex != 2 {
			t.Fatalf("bad service returned: %#v", r)
		}
	}
	verifyNode()
	verifyService()

	// Add in a top-level check.
	req.Check = &structs.HealthCheck{
		Node:    "node1",
		CheckID: "check1",
		Name:    "check",
	}
	if err := s.EnsureRegistration(3, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify that the check got registered.
	verifyCheck := func() {
		idx, out, err := s.NodeChecks(nil, "node1")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 3 {
			t.Fatalf("bad index: %d", idx)
		}
		if len(out) != 1 {
			t.Fatalf("bad: %#v", out)
		}
		c := out[0]
		if c.Node != "node1" || c.CheckID != "check1" || c.Name != "check" ||
			c.CreateIndex != 3 || c.ModifyIndex != 3 {
			t.Fatalf("bad check returned: %#v", c)
		}

		idx, c, err = s.NodeCheck("node1", "check1")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 3 {
			t.Fatalf("bad index: %d", idx)
		}
		if c.Node != "node1" || c.CheckID != "check1" || c.Name != "check" ||
			c.CreateIndex != 3 || c.ModifyIndex != 3 {
			t.Fatalf("bad check returned: %#v", c)
		}
	}
	verifyNode()
	verifyService()
	verifyCheck()

	// Add in another check via the slice.
	req.Checks = structs.HealthChecks{
		&structs.HealthCheck{
			Node:    "node1",
			CheckID: "check2",
			Name:    "check",
		},
	}
	if err := s.EnsureRegistration(4, req); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify that the additional check got registered.
	verifyNode()
	verifyService()
	verifyChecks := func() {
		idx, out, err := s.NodeChecks(nil, "node1")
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 4 {
			t.Fatalf("bad index: %d", idx)
		}
		if len(out) != 2 {
			t.Fatalf("bad: %#v", out)
		}
		c1 := out[0]
		if c1.Node != "node1" || c1.CheckID != "check1" || c1.Name != "check" ||
			c1.CreateIndex != 3 || c1.ModifyIndex != 4 {
			t.Fatalf("bad check returned: %#v", c1)
		}

		c2 := out[1]
		if c2.Node != "node1" || c2.CheckID != "check2" || c2.Name != "check" ||
			c2.CreateIndex != 4 || c2.ModifyIndex != 4 {
			t.Fatalf("bad check returned: %#v", c2)
		}
	}
	verifyChecks()

	// Try to register a check for some other node (top-level check).
	req.Check = &structs.HealthCheck{
		Node:    "nope",
		CheckID: "check1",
		Name:    "check",
	}
	err := s.EnsureRegistration(5, req)
	if err == nil || !strings.Contains(err.Error(), "does not match node") {
		t.Fatalf("err: %s", err)
	}
	verifyNode()
	verifyService()
	verifyChecks()

	// Try to register a check for some other node (checks array).
	req.Check = nil
	req.Checks = structs.HealthChecks{
		&structs.HealthCheck{
			Node:    "nope",
			CheckID: "check2",
			Name:    "check",
		},
	}
	err = s.EnsureRegistration(6, req)
	if err == nil || !strings.Contains(err.Error(), "does not match node") {
		t.Fatalf("err: %s", err)
	}
	verifyNode()
	verifyService()
	verifyChecks()
}

func TestStateStore_EnsureRegistration_Restore(t *testing.T) {
	s := testStateStore(t)

	// Start with just a node.
	req := &structs.RegisterRequest{
		ID:      makeRandomNodeID(t),
		Node:    "node1",
		Address: "1.2.3.4",
	}
	nodeID := string(req.ID)
	nodeName := string(req.Node)
	restore := s.Restore()
	if err := restore.Registration(1, req); err != nil {
		t.Fatalf("err: %s", err)
	}
	restore.Commit()

	// Retrieve the node and verify its contents.
	verifyNode := func(nodeLookup string) {
		_, out, err := s.GetNode(nodeLookup)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if out == nil {
			_, out, err = s.GetNodeID(types.NodeID(nodeLookup))
			if err != nil {
				t.Fatalf("err: %s", err)
			}
		}

		if out == nil || out.Address != "1.2.3.4" ||
			!(out.Node == nodeLookup || string(out.ID) == nodeLookup) ||
			out.CreateIndex != 1 || out.ModifyIndex != 1 {
			t.Fatalf("bad node returned: %#v", out)
		}
	}
	verifyNode(nodeID)
	verifyNode(nodeName)

	// Add in a service definition.
	req.Service = &structs.NodeService{
		ID:      "redis1",
		Service: "redis",
		Address: "1.1.1.1",
		Port:    8080,
	}
	restore = s.Restore()
	if err := restore.Registration(2, req); err != nil {
		t.Fatalf("err: %s", err)
	}
	restore.Commit()

	// Verify that the service got registered.
	verifyService := func(nodeLookup string) {
		idx, out, err := s.NodeServices(nil, nodeLookup)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 2 {
			t.Fatalf("bad index: %d", idx)
		}
		if len(out.Services) != 1 {
			t.Fatalf("bad: %#v", out.Services)
		}
		s := out.Services["redis1"]
		if s.ID != "redis1" || s.Service != "redis" ||
			s.Address != "1.1.1.1" || s.Port != 8080 ||
			s.CreateIndex != 2 || s.ModifyIndex != 2 {
			t.Fatalf("bad service returned: %#v", s)
		}
	}

	// Add in a top-level check.
	req.Check = &structs.HealthCheck{
		Node:    nodeName,
		CheckID: "check1",
		Name:    "check",
	}
	restore = s.Restore()
	if err := restore.Registration(3, req); err != nil {
		t.Fatalf("err: %s", err)
	}
	restore.Commit()

	// Verify that the check got registered.
	verifyCheck := func() {
		idx, out, err := s.NodeChecks(nil, nodeName)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 3 {
			t.Fatalf("bad index: %d", idx)
		}
		if len(out) != 1 {
			t.Fatalf("bad: %#v", out)
		}
		c := out[0]
		if c.Node != nodeName || c.CheckID != "check1" || c.Name != "check" ||
			c.CreateIndex != 3 || c.ModifyIndex != 3 {
			t.Fatalf("bad check returned: %#v", c)
		}
	}
	verifyNode(nodeID)
	verifyNode(nodeName)
	verifyService(nodeID)
	verifyService(nodeName)
	verifyCheck()

	// Add in another check via the slice.
	req.Checks = structs.HealthChecks{
		&structs.HealthCheck{
			Node:    nodeName,
			CheckID: "check2",
			Name:    "check",
		},
	}
	restore = s.Restore()
	if err := restore.Registration(4, req); err != nil {
		t.Fatalf("err: %s", err)
	}
	restore.Commit()

	// Verify that the additional check got registered.
	verifyNode(nodeID)
	verifyNode(nodeName)
	verifyService(nodeID)
	verifyService(nodeName)
	func() {
		idx, out, err := s.NodeChecks(nil, nodeName)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if idx != 4 {
			t.Fatalf("bad index: %d", idx)
		}
		if len(out) != 2 {
			t.Fatalf("bad: %#v", out)
		}
		c1 := out[0]
		if c1.Node != nodeName || c1.CheckID != "check1" || c1.Name != "check" ||
			c1.CreateIndex != 3 || c1.ModifyIndex != 4 {
			t.Fatalf("bad check returned: %#v", c1)
		}

		c2 := out[1]
		if c2.Node != nodeName || c2.CheckID != "check2" || c2.Name != "check" ||
			c2.CreateIndex != 4 || c2.ModifyIndex != 4 {
			t.Fatalf("bad check returned: %#v", c2)
		}
	}()
}

func TestStateStore_EnsureNode(t *testing.T) {
	s := testStateStore(t)

	// Fetching a non-existent node returns nil
	if _, node, err := s.GetNode("node1"); node != nil || err != nil {
		t.Fatalf("expected (nil, nil), got: (%#v, %#v)", node, err)
	}

	// Create a node registration request
	in := &structs.Node{
		Node:    "node1",
		Address: "1.1.1.1",
	}

	// Ensure the node is registered in the db
	if err := s.EnsureNode(1, in); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Retrieve the node again
	idx, out, err := s.GetNode("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Correct node was returned
	if out.Node != "node1" || out.Address != "1.1.1.1" {
		t.Fatalf("bad node returned: %#v", out)
	}

	// Indexes are set properly
	if out.CreateIndex != 1 || out.ModifyIndex != 1 {
		t.Fatalf("bad node index: %#v", out)
	}
	if idx != 1 {
		t.Fatalf("bad index: %d", idx)
	}

	// Update the node registration
	in.Address = "1.1.1.2"
	if err := s.EnsureNode(2, in); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Retrieve the node
	idx, out, err = s.GetNode("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Node and indexes were updated
	if out.CreateIndex != 1 || out.ModifyIndex != 2 || out.Address != "1.1.1.2" {
		t.Fatalf("bad: %#v", out)
	}
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// Node upsert preserves the create index
	if err := s.EnsureNode(3, in); err != nil {
		t.Fatalf("err: %s", err)
	}
	idx, out, err = s.GetNode("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if out.CreateIndex != 1 || out.ModifyIndex != 3 || out.Address != "1.1.1.2" {
		t.Fatalf("node was modified: %#v", out)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Add an ID to the node
	in.ID = types.NodeID("cda916bc-a357-4a19-b886-59419fcee50c")
	if err := s.EnsureNode(4, in); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Now try to add another node with the same ID
	in = &structs.Node{
		Node:    "nope",
		ID:      types.NodeID("cda916bc-a357-4a19-b886-59419fcee50c"),
		Address: "1.2.3.4",
	}
	err = s.EnsureNode(5, in)
	if err == nil || !strings.Contains(err.Error(), "aliases existing node") {
		t.Fatalf("err: %v", err)
	}
}

func TestStateStore_GetNodes(t *testing.T) {
	s := testStateStore(t)

	// Listing with no results returns nil.
	ws := memdb.NewWatchSet()
	idx, res, err := s.Nodes(ws)
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Create some nodes in the state store.
	testRegisterNode(t, s, 0, "node0")
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Retrieve the nodes.
	ws = memdb.NewWatchSet()
	idx, nodes, err := s.Nodes(ws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Highest index was returned.
	if idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}

	// All nodes were returned.
	if n := len(nodes); n != 3 {
		t.Fatalf("bad node count: %d", n)
	}

	// Make sure the nodes match.
	for i, node := range nodes {
		if node.CreateIndex != uint64(i) || node.ModifyIndex != uint64(i) {
			t.Fatalf("bad node index: %d, %d", node.CreateIndex, node.ModifyIndex)
		}
		name := fmt.Sprintf("node%d", i)
		if node.Node != name {
			t.Fatalf("bad: %#v", node)
		}
	}

	// Make sure a node delete fires the watch.
	if watchFired(ws) {
		t.Fatalf("bad")
	}
	if err := s.DeleteNode(3, "node1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func BenchmarkGetNodes(b *testing.B) {
	s, err := NewStateStore(nil)
	if err != nil {
		b.Fatalf("err: %s", err)
	}

	if err := s.EnsureNode(100, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		b.Fatalf("err: %v", err)
	}
	if err := s.EnsureNode(101, &structs.Node{Node: "bar", Address: "127.0.0.2"}); err != nil {
		b.Fatalf("err: %v", err)
	}

	ws := memdb.NewWatchSet()
	for i := 0; i < b.N; i++ {
		s.Nodes(ws)
	}
}

func TestStateStore_GetNodesByMeta(t *testing.T) {
	s := testStateStore(t)

	// Listing with no results returns nil
	ws := memdb.NewWatchSet()
	idx, res, err := s.NodesByMeta(ws, map[string]string{"somekey": "somevalue"})
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Create some nodes in the state store.
	testRegisterNodeWithMeta(t, s, 0, "node0", map[string]string{"role": "client"})
	testRegisterNodeWithMeta(t, s, 1, "node1", map[string]string{"role": "client", "common": "1"})
	testRegisterNodeWithMeta(t, s, 2, "node2", map[string]string{"role": "server", "common": "1"})
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	cases := []struct {
		filters map[string]string
		nodes   []string
	}{
		// Simple meta filter
		{
			filters: map[string]string{"role": "server"},
			nodes:   []string{"node2"},
		},
		// Common meta filter
		{
			filters: map[string]string{"common": "1"},
			nodes:   []string{"node1", "node2"},
		},
		// Invalid meta filter
		{
			filters: map[string]string{"invalid": "nope"},
			nodes:   []string{},
		},
		// Multiple meta filters
		{
			filters: map[string]string{"role": "client", "common": "1"},
			nodes:   []string{"node1"},
		},
	}

	for _, tc := range cases {
		_, result, err := s.NodesByMeta(nil, tc.filters)
		if err != nil {
			t.Fatalf("bad: %v", err)
		}

		if len(result) != len(tc.nodes) {
			t.Fatalf("bad: %v %v", result, tc.nodes)
		}

		for i, node := range result {
			if node.Node != tc.nodes[i] {
				t.Fatalf("bad: %v %v", node.Node, tc.nodes[i])
			}
		}
	}

	// Set up a watch.
	ws = memdb.NewWatchSet()
	_, _, err = s.NodesByMeta(ws, map[string]string{"role": "client"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make an unrelated modification and make sure the watch doesn't fire.
	testRegisterNodeWithMeta(t, s, 3, "node3", map[string]string{"foo": "bar"})
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Change a watched key and make sure it fires.
	testRegisterNodeWithMeta(t, s, 4, "node0", map[string]string{"role": "different"})
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_NodeServices(t *testing.T) {
	s := testStateStore(t)

	// Register some nodes with similar IDs.
	{
		req := &structs.RegisterRequest{
			ID:      types.NodeID("40e4a748-2192-161a-0510-aaaaaaaaaaaa"),
			Node:    "node1",
			Address: "1.2.3.4",
		}
		if err := s.EnsureRegistration(1, req); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		req := &structs.RegisterRequest{
			ID:      types.NodeID("40e4a748-2192-161a-0510-bbbbbbbbbbbb"),
			Node:    "node2",
			Address: "5.6.7.8",
		}
		if err := s.EnsureRegistration(2, req); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Look up by name.
	{
		_, ns, err := s.NodeServices(nil, "node1")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns == nil || ns.Node.Node != "node1" {
			t.Fatalf("bad: %#v", *ns)
		}
	}
	{
		_, ns, err := s.NodeServices(nil, "node2")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns == nil || ns.Node.Node != "node2" {
			t.Fatalf("bad: %#v", *ns)
		}
	}

	// Look up by UUID.
	{
		_, ns, err := s.NodeServices(nil, "40e4a748-2192-161a-0510-aaaaaaaaaaaa")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns == nil || ns.Node.Node != "node1" {
			t.Fatalf("bad: %#v", ns)
		}
	}
	{
		_, ns, err := s.NodeServices(nil, "40e4a748-2192-161a-0510-bbbbbbbbbbbb")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns == nil || ns.Node.Node != "node2" {
			t.Fatalf("bad: %#v", ns)
		}
	}

	// Ambiguous prefix.
	{
		_, ns, err := s.NodeServices(nil, "40e4a748-2192-161a-0510")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns != nil {
			t.Fatalf("bad: %#v", ns)
		}
	}

	// Bad node, and not a UUID (should not get a UUID error).
	{
		_, ns, err := s.NodeServices(nil, "nope")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns != nil {
			t.Fatalf("bad: %#v", ns)
		}
	}

	// Specific prefix.
	{
		_, ns, err := s.NodeServices(nil, "40e4a748-2192-161a-0510-bb")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if ns == nil || ns.Node.Node != "node2" {
			t.Fatalf("bad: %#v", ns)
		}
	}
}

func TestStateStore_DeleteNode(t *testing.T) {
	s := testStateStore(t)

	// Create a node and register a service and health check with it.
	testRegisterNode(t, s, 0, "node1")
	testRegisterService(t, s, 1, "node1", "service1")
	testRegisterCheck(t, s, 2, "node1", "", "check1", structs.HealthPassing)

	// Delete the node
	if err := s.DeleteNode(3, "node1"); err != nil {
		t.Fatalf("err: %s", err)
	}

	// The node was removed
	if idx, n, err := s.GetNode("node1"); err != nil || n != nil || idx != 3 {
		t.Fatalf("bad: %#v %d (err: %#v)", n, idx, err)
	}

	// Associated service was removed. Need to query this directly out of
	// the DB to make sure it is actually gone.
	tx := s.db.Txn(false)
	defer tx.Abort()
	services, err := tx.Get("services", "id", "node1", "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if service := services.Next(); service != nil {
		t.Fatalf("bad: %#v", service)
	}

	// Associated health check was removed.
	checks, err := tx.Get("checks", "id", "node1", "check1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if check := checks.Next(); check != nil {
		t.Fatalf("bad: %#v", check)
	}

	// Indexes were updated.
	for _, tbl := range []string{"nodes", "services", "checks"} {
		if idx := s.maxIndex(tbl); idx != 3 {
			t.Fatalf("bad index: %d (%s)", idx, tbl)
		}
	}

	// Deleting a nonexistent node should be idempotent and not return
	// an error
	if err := s.DeleteNode(4, "node1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx := s.maxIndex("nodes"); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_Node_Snapshot(t *testing.T) {
	s := testStateStore(t)

	// Create some nodes in the state store.
	testRegisterNode(t, s, 0, "node0")
	testRegisterNode(t, s, 1, "node1")
	testRegisterNode(t, s, 2, "node2")

	// Snapshot the nodes.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	testRegisterNode(t, s, 3, "node3")

	// Verify the snapshot.
	if idx := snap.LastIndex(); idx != 2 {
		t.Fatalf("bad index: %d", idx)
	}
	nodes, err := snap.Nodes()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	for i := 0; i < 3; i++ {
		node := nodes.Next().(*structs.Node)
		if node == nil {
			t.Fatalf("unexpected end of nodes")
		}

		if node.CreateIndex != uint64(i) || node.ModifyIndex != uint64(i) {
			t.Fatalf("bad node index: %d, %d", node.CreateIndex, node.ModifyIndex)
		}
		if node.Node != fmt.Sprintf("node%d", i) {
			t.Fatalf("bad: %#v", node)
		}
	}
	if nodes.Next() != nil {
		t.Fatalf("unexpected extra nodes")
	}
}

func TestStateStore_EnsureService(t *testing.T) {
	s := testStateStore(t)

	// Fetching services for a node with none returns nil.
	ws := memdb.NewWatchSet()
	idx, res, err := s.NodeServices(ws, "node1")
	if err != nil || res != nil || idx != 0 {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Create the service registration.
	ns1 := &structs.NodeService{
		ID:      "service1",
		Service: "redis",
		Tags:    []string{"prod"},
		Address: "1.1.1.1",
		Port:    1111,
	}

	// Creating a service without a node returns an error.
	if err := s.EnsureService(1, "node1", ns1); err != ErrMissingNode {
		t.Fatalf("expected %#v, got: %#v", ErrMissingNode, err)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Register the nodes.
	testRegisterNode(t, s, 0, "node1")
	testRegisterNode(t, s, 1, "node2")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Service successfully registers into the state store.
	ws = memdb.NewWatchSet()
	_, _, err = s.NodeServices(ws, "node1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if err = s.EnsureService(10, "node1", ns1); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Register a similar service against both nodes.
	ns2 := *ns1
	ns2.ID = "service2"
	for _, n := range []string{"node1", "node2"} {
		if err := s.EnsureService(20, n, &ns2); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Register a different service on the bad node.
	ws = memdb.NewWatchSet()
	_, _, err = s.NodeServices(ws, "node1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	ns3 := *ns1
	ns3.ID = "service3"
	if err := s.EnsureService(30, "node2", &ns3); err != nil {
		t.Fatalf("err: %s", err)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Retrieve the services.
	ws = memdb.NewWatchSet()
	idx, out, err := s.NodeServices(ws, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 30 {
		t.Fatalf("bad index: %d", idx)
	}

	// Only the services for the requested node are returned.
	if out == nil || len(out.Services) != 2 {
		t.Fatalf("bad services: %#v", out)
	}

	// Results match the inserted services and have the proper indexes set.
	expect1 := *ns1
	expect1.CreateIndex, expect1.ModifyIndex = 10, 10
	if svc := out.Services["service1"]; !reflect.DeepEqual(&expect1, svc) {
		t.Fatalf("bad: %#v", svc)
	}

	expect2 := ns2
	expect2.CreateIndex, expect2.ModifyIndex = 20, 20
	if svc := out.Services["service2"]; !reflect.DeepEqual(&expect2, svc) {
		t.Fatalf("bad: %#v %#v", ns2, svc)
	}

	// Index tables were updated.
	if idx := s.maxIndex("services"); idx != 30 {
		t.Fatalf("bad index: %d", idx)
	}

	// Update a service registration.
	ns1.Address = "1.1.1.2"
	if err := s.EnsureService(40, "node1", ns1); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Retrieve the service again and ensure it matches..
	idx, out, err = s.NodeServices(nil, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 40 {
		t.Fatalf("bad index: %d", idx)
	}
	if out == nil || len(out.Services) != 2 {
		t.Fatalf("bad: %#v", out)
	}
	expect1.Address = "1.1.1.2"
	expect1.ModifyIndex = 40
	if svc := out.Services["service1"]; !reflect.DeepEqual(&expect1, svc) {
		t.Fatalf("bad: %#v", svc)
	}

	// Index tables were updated.
	if idx := s.maxIndex("services"); idx != 40 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_Services(t *testing.T) {
	s := testStateStore(t)

	// Listing with no results returns an empty list.
	ws := memdb.NewWatchSet()
	idx, services, err := s.Services(ws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if len(services) != 0 {
		t.Fatalf("bad: %v", services)
	}

	// Register several nodes and services.
	testRegisterNode(t, s, 1, "node1")
	ns1 := &structs.NodeService{
		ID:      "service1",
		Service: "redis",
		Tags:    []string{"prod", "master"},
		Address: "1.1.1.1",
		Port:    1111,
	}
	if err := s.EnsureService(2, "node1", ns1); err != nil {
		t.Fatalf("err: %s", err)
	}
	testRegisterService(t, s, 3, "node1", "dogs")
	testRegisterNode(t, s, 4, "node2")
	ns2 := &structs.NodeService{
		ID:      "service3",
		Service: "redis",
		Tags:    []string{"prod", "slave"},
		Address: "1.1.1.1",
		Port:    1111,
	}
	if err := s.EnsureService(5, "node2", ns2); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Pull all the services.
	ws = memdb.NewWatchSet()
	idx, services, err = s.Services(ws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}

	// Verify the result. We sort the lists since the order is
	// non-deterministic (it's built using a map internally).
	expected := structs.Services{
		"redis": []string{"prod", "master", "slave"},
		"dogs":  []string{},
	}
	sort.Strings(expected["redis"])
	for _, tags := range services {
		sort.Strings(tags)
	}
	if !reflect.DeepEqual(expected, services) {
		t.Fatalf("bad: %#v", services)
	}

	// Deleting a node with a service should fire the watch.
	if err := s.DeleteNode(6, "node1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ServicesByNodeMeta(t *testing.T) {
	s := testStateStore(t)

	// Listing with no results returns nil.
	ws := memdb.NewWatchSet()
	idx, res, err := s.ServicesByNodeMeta(ws, map[string]string{"somekey": "somevalue"})
	if idx != 0 || len(res) != 0 || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Create some nodes and services in the state store.
	node0 := &structs.Node{Node: "node0", Address: "127.0.0.1", Meta: map[string]string{"role": "client", "common": "1"}}
	if err := s.EnsureNode(0, node0); err != nil {
		t.Fatalf("err: %v", err)
	}
	node1 := &structs.Node{Node: "node1", Address: "127.0.0.1", Meta: map[string]string{"role": "server", "common": "1"}}
	if err := s.EnsureNode(1, node1); err != nil {
		t.Fatalf("err: %v", err)
	}
	ns1 := &structs.NodeService{
		ID:      "service1",
		Service: "redis",
		Tags:    []string{"prod", "master"},
		Address: "1.1.1.1",
		Port:    1111,
	}
	if err := s.EnsureService(2, "node0", ns1); err != nil {
		t.Fatalf("err: %s", err)
	}
	ns2 := &structs.NodeService{
		ID:      "service1",
		Service: "redis",
		Tags:    []string{"prod", "slave"},
		Address: "1.1.1.1",
		Port:    1111,
	}
	if err := s.EnsureService(3, "node1", ns2); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Filter the services by the first node's meta value.
	ws = memdb.NewWatchSet()
	_, res, err = s.ServicesByNodeMeta(ws, map[string]string{"role": "client"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	expected := structs.Services{
		"redis": []string{"master", "prod"},
	}
	sort.Strings(res["redis"])
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("bad: %v %v", res, expected)
	}

	// Get all services using the common meta value
	_, res, err = s.ServicesByNodeMeta(ws, map[string]string{"common": "1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	expected = structs.Services{
		"redis": []string{"master", "prod", "slave"},
	}
	sort.Strings(res["redis"])
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("bad: %v %v", res, expected)
	}

	// Get an empty list for an invalid meta value
	_, res, err = s.ServicesByNodeMeta(ws, map[string]string{"invalid": "nope"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	expected = structs.Services{}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("bad: %v %v", res, expected)
	}

	// Get the first node's service instance using multiple meta filters
	_, res, err = s.ServicesByNodeMeta(ws, map[string]string{"role": "client", "common": "1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	expected = structs.Services{
		"redis": []string{"master", "prod"},
	}
	sort.Strings(res["redis"])
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("bad: %v %v", res, expected)
	}

	// Sanity check the watch before we proceed.
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Registering some unrelated node + service should not fire the watch.
	testRegisterNode(t, s, 4, "nope")
	testRegisterService(t, s, 5, "nope", "nope")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Overwhelm the service tracking.
	idx = 6
	for i := 0; i < 2*watchLimit; i++ {
		node := fmt.Sprintf("many%d", i)
		testRegisterNodeWithMeta(t, s, idx, node, map[string]string{"common": "1"})
		idx++
		testRegisterService(t, s, idx, node, "nope")
		idx++
	}

	// Now get a fresh watch, which will be forced to watch the whole
	// service table.
	ws = memdb.NewWatchSet()
	_, _, err = s.ServicesByNodeMeta(ws, map[string]string{"common": "1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Registering some unrelated node + service should not fire the watch.
	testRegisterService(t, s, idx, "nope", "more-nope")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ServiceNodes(t *testing.T) {
	s := testStateStore(t)

	// Listing with no results returns an empty list.
	ws := memdb.NewWatchSet()
	idx, nodes, err := s.ServiceNodes(ws, "db")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if len(nodes) != 0 {
		t.Fatalf("bad: %v", nodes)
	}

	// Create some nodes and services.
	if err := s.EnsureNode(10, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureNode(11, &structs.Node{Node: "bar", Address: "127.0.0.2"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(12, "foo", &structs.NodeService{ID: "api", Service: "api", Tags: nil, Address: "", Port: 5000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(13, "bar", &structs.NodeService{ID: "api", Service: "api", Tags: nil, Address: "", Port: 5000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(14, "foo", &structs.NodeService{ID: "db", Service: "db", Tags: []string{"master"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(15, "bar", &structs.NodeService{ID: "db", Service: "db", Tags: []string{"slave"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(16, "bar", &structs.NodeService{ID: "db2", Service: "db", Tags: []string{"slave"}, Address: "", Port: 8001}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Read everything back.
	ws = memdb.NewWatchSet()
	idx, nodes, err = s.ServiceNodes(ws, "db")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 16 {
		t.Fatalf("bad: %d", idx)
	}
	if len(nodes) != 3 {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Node != "bar" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Address != "127.0.0.2" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].ServiceID != "db" {
		t.Fatalf("bad: %v", nodes)
	}
	if !lib.StrContains(nodes[0].ServiceTags, "slave") {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].ServicePort != 8000 {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[1].Node != "bar" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[1].Address != "127.0.0.2" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[1].ServiceID != "db2" {
		t.Fatalf("bad: %v", nodes)
	}
	if !lib.StrContains(nodes[1].ServiceTags, "slave") {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[1].ServicePort != 8001 {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[2].Node != "foo" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[2].Address != "127.0.0.1" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[2].ServiceID != "db" {
		t.Fatalf("bad: %v", nodes)
	}
	if !lib.StrContains(nodes[2].ServiceTags, "master") {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[2].ServicePort != 8000 {
		t.Fatalf("bad: %v", nodes)
	}

	// Registering some unrelated node should not fire the watch.
	testRegisterNode(t, s, 17, "nope")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// But removing a node with the "db" service should fire the watch.
	if err := s.DeleteNode(18, "bar"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Overwhelm the node tracking.
	idx = 19
	for i := 0; i < 2*watchLimit; i++ {
		node := fmt.Sprintf("many%d", i)
		if err := s.EnsureNode(idx, &structs.Node{Node: node, Address: "127.0.0.1"}); err != nil {
			t.Fatalf("err: %v", err)
		}
		if err := s.EnsureService(idx, node, &structs.NodeService{ID: "db", Service: "db", Port: 8000}); err != nil {
			t.Fatalf("err: %v", err)
		}
		idx++
	}

	// Now get a fresh watch, which will be forced to watch the whole nodes
	// table.
	ws = memdb.NewWatchSet()
	_, _, err = s.ServiceNodes(ws, "db")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Registering some unrelated node should fire the watch now.
	testRegisterNode(t, s, idx, "more-nope")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ServiceTagNodes(t *testing.T) {
	s := testStateStore(t)

	// Listing with no results returns an empty list.
	ws := memdb.NewWatchSet()
	idx, nodes, err := s.ServiceTagNodes(ws, "db", "master")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if len(nodes) != 0 {
		t.Fatalf("bad: %v", nodes)
	}

	// Create some nodes and services.
	if err := s.EnsureNode(15, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureNode(16, &structs.Node{Node: "bar", Address: "127.0.0.2"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(17, "foo", &structs.NodeService{ID: "db", Service: "db", Tags: []string{"master"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(18, "foo", &structs.NodeService{ID: "db2", Service: "db", Tags: []string{"slave"}, Address: "", Port: 8001}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(19, "bar", &structs.NodeService{ID: "db", Service: "db", Tags: []string{"slave"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Read everything back.
	ws = memdb.NewWatchSet()
	idx, nodes, err = s.ServiceTagNodes(ws, "db", "master")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 19 {
		t.Fatalf("bad: %v", idx)
	}
	if len(nodes) != 1 {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Node != "foo" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Address != "127.0.0.1" {
		t.Fatalf("bad: %v", nodes)
	}
	if !lib.StrContains(nodes[0].ServiceTags, "master") {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].ServicePort != 8000 {
		t.Fatalf("bad: %v", nodes)
	}

	// Registering some unrelated node should not fire the watch.
	testRegisterNode(t, s, 20, "nope")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// But removing a node with the "db:master" service should fire the watch.
	if err := s.DeleteNode(21, "foo"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ServiceTagNodes_MultipleTags(t *testing.T) {
	s := testStateStore(t)

	if err := s.EnsureNode(15, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := s.EnsureNode(16, &structs.Node{Node: "bar", Address: "127.0.0.2"}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := s.EnsureService(17, "foo", &structs.NodeService{ID: "db", Service: "db", Tags: []string{"master", "v2"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := s.EnsureService(18, "foo", &structs.NodeService{ID: "db2", Service: "db", Tags: []string{"slave", "v2", "dev"}, Address: "", Port: 8001}); err != nil {
		t.Fatalf("err: %v", err)
	}

	if err := s.EnsureService(19, "bar", &structs.NodeService{ID: "db", Service: "db", Tags: []string{"slave", "v2"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}

	idx, nodes, err := s.ServiceTagNodes(nil, "db", "master")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 19 {
		t.Fatalf("bad: %v", idx)
	}
	if len(nodes) != 1 {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Node != "foo" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Address != "127.0.0.1" {
		t.Fatalf("bad: %v", nodes)
	}
	if !lib.StrContains(nodes[0].ServiceTags, "master") {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].ServicePort != 8000 {
		t.Fatalf("bad: %v", nodes)
	}

	idx, nodes, err = s.ServiceTagNodes(nil, "db", "v2")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 19 {
		t.Fatalf("bad: %v", idx)
	}
	if len(nodes) != 3 {
		t.Fatalf("bad: %v", nodes)
	}

	idx, nodes, err = s.ServiceTagNodes(nil, "db", "dev")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 19 {
		t.Fatalf("bad: %v", idx)
	}
	if len(nodes) != 1 {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Node != "foo" {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].Address != "127.0.0.1" {
		t.Fatalf("bad: %v", nodes)
	}
	if !lib.StrContains(nodes[0].ServiceTags, "dev") {
		t.Fatalf("bad: %v", nodes)
	}
	if nodes[0].ServicePort != 8001 {
		t.Fatalf("bad: %v", nodes)
	}
}

func TestStateStore_DeleteService(t *testing.T) {
	s := testStateStore(t)

	// Register a node with one service and a check.
	testRegisterNode(t, s, 1, "node1")
	testRegisterService(t, s, 2, "node1", "service1")
	testRegisterCheck(t, s, 3, "node1", "service1", "check1", structs.HealthPassing)

	// Delete the service.
	ws := memdb.NewWatchSet()
	_, _, err := s.NodeServices(ws, "node1")
	if err := s.DeleteService(4, "node1", "service1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Service doesn't exist.
	ws = memdb.NewWatchSet()
	_, ns, err := s.NodeServices(ws, "node1")
	if err != nil || ns == nil || len(ns.Services) != 0 {
		t.Fatalf("bad: %#v (err: %#v)", ns, err)
	}

	// Check doesn't exist. Check using the raw DB so we can test
	// that it actually is removed in the state store.
	tx := s.db.Txn(false)
	defer tx.Abort()
	check, err := tx.First("checks", "id", "node1", "check1")
	if err != nil || check != nil {
		t.Fatalf("bad: %#v (err: %s)", check, err)
	}

	// Index tables were updated.
	if idx := s.maxIndex("services"); idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
	if idx := s.maxIndex("checks"); idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}

	// Deleting a nonexistent service should be idempotent and not return an
	// error, nor fire a watch.
	if err := s.DeleteService(5, "node1", "service1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx := s.maxIndex("services"); idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_Service_Snapshot(t *testing.T) {
	s := testStateStore(t)

	// Register a node with two services.
	testRegisterNode(t, s, 0, "node1")
	ns := []*structs.NodeService{
		&structs.NodeService{
			ID:      "service1",
			Service: "redis",
			Tags:    []string{"prod"},
			Address: "1.1.1.1",
			Port:    1111,
		},
		&structs.NodeService{
			ID:      "service2",
			Service: "nomad",
			Tags:    []string{"dev"},
			Address: "1.1.1.2",
			Port:    1112,
		},
	}
	for i, svc := range ns {
		if err := s.EnsureService(uint64(i+1), "node1", svc); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Create a second node/service to make sure node filtering works. This
	// will affect the index but not the dump.
	testRegisterNode(t, s, 3, "node2")
	testRegisterService(t, s, 4, "node2", "service2")

	// Snapshot the service.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	testRegisterService(t, s, 5, "node2", "service3")

	// Verify the snapshot.
	if idx := snap.LastIndex(); idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
	services, err := snap.Services("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	for i := 0; i < len(ns); i++ {
		svc := services.Next().(*structs.ServiceNode)
		if svc == nil {
			t.Fatalf("unexpected end of services")
		}

		ns[i].CreateIndex, ns[i].ModifyIndex = uint64(i+1), uint64(i+1)
		if !reflect.DeepEqual(ns[i], svc.ToNodeService()) {
			t.Fatalf("bad: %#v != %#v", svc, ns[i])
		}
	}
	if services.Next() != nil {
		t.Fatalf("unexpected extra services")
	}
}

func TestStateStore_EnsureCheck(t *testing.T) {
	s := testStateStore(t)

	// Create a check associated with the node
	check := &structs.HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "redis check",
		Status:      structs.HealthPassing,
		Notes:       "test check",
		Output:      "aaa",
		ServiceID:   "service1",
		ServiceName: "redis",
	}

	// Creating a check without a node returns error
	if err := s.EnsureCheck(1, check); err != ErrMissingNode {
		t.Fatalf("expected %#v, got: %#v", ErrMissingNode, err)
	}

	// Register the node
	testRegisterNode(t, s, 1, "node1")

	// Creating a check with a bad services returns error
	if err := s.EnsureCheck(1, check); err != ErrMissingService {
		t.Fatalf("expected: %#v, got: %#v", ErrMissingService, err)
	}

	// Register the service
	testRegisterService(t, s, 2, "node1", "service1")

	// Inserting the check with the prerequisites succeeds
	if err := s.EnsureCheck(3, check); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Retrieve the check and make sure it matches
	idx, checks, err := s.NodeChecks(nil, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
	if len(checks) != 1 {
		t.Fatalf("wrong number of checks: %d", len(checks))
	}
	if !reflect.DeepEqual(checks[0], check) {
		t.Fatalf("bad: %#v", checks[0])
	}

	// Modify the health check
	check.Output = "bbb"
	if err := s.EnsureCheck(4, check); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check that we successfully updated
	idx, checks, err = s.NodeChecks(nil, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
	if len(checks) != 1 {
		t.Fatalf("wrong number of checks: %d", len(checks))
	}
	if checks[0].Output != "bbb" {
		t.Fatalf("wrong check output: %#v", checks[0])
	}
	if checks[0].CreateIndex != 3 || checks[0].ModifyIndex != 4 {
		t.Fatalf("bad index: %#v", checks[0])
	}

	// Index tables were updated
	if idx := s.maxIndex("checks"); idx != 4 {
		t.Fatalf("bad index: %d", idx)
	}
}

func TestStateStore_EnsureCheck_defaultStatus(t *testing.T) {
	s := testStateStore(t)

	// Register a node
	testRegisterNode(t, s, 1, "node1")

	// Create and register a check with no health status
	check := &structs.HealthCheck{
		Node:    "node1",
		CheckID: "check1",
		Status:  "",
	}
	if err := s.EnsureCheck(2, check); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get the check again
	_, result, err := s.NodeChecks(nil, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check that the status was set to the proper default
	if len(result) != 1 || result[0].Status != structs.HealthCritical {
		t.Fatalf("bad: %#v", result)
	}
}

func TestStateStore_NodeChecks(t *testing.T) {
	s := testStateStore(t)

	// Do an initial query for a node that doesn't exist.
	ws := memdb.NewWatchSet()
	idx, checks, err := s.NodeChecks(ws, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if len(checks) != 0 {
		t.Fatalf("bad: %#v", checks)
	}

	// Create some nodes and checks.
	testRegisterNode(t, s, 0, "node1")
	testRegisterService(t, s, 1, "node1", "service1")
	testRegisterCheck(t, s, 2, "node1", "service1", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 3, "node1", "service1", "check2", structs.HealthPassing)
	testRegisterNode(t, s, 4, "node2")
	testRegisterService(t, s, 5, "node2", "service2")
	testRegisterCheck(t, s, 6, "node2", "service2", "check3", structs.HealthPassing)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Try querying for all checks associated with node1
	ws = memdb.NewWatchSet()
	idx, checks, err = s.NodeChecks(ws, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}
	if len(checks) != 2 || checks[0].CheckID != "check1" || checks[1].CheckID != "check2" {
		t.Fatalf("bad checks: %#v", checks)
	}

	// Creating some unrelated node should not fire the watch.
	testRegisterNode(t, s, 7, "node3")
	testRegisterCheck(t, s, 8, "node3", "", "check1", structs.HealthPassing)
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Try querying for all checks associated with node2
	ws = memdb.NewWatchSet()
	idx, checks, err = s.NodeChecks(ws, "node2")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 8 {
		t.Fatalf("bad index: %d", idx)
	}
	if len(checks) != 1 || checks[0].CheckID != "check3" {
		t.Fatalf("bad checks: %#v", checks)
	}

	// Changing node2 should fire the watch.
	testRegisterCheck(t, s, 9, "node2", "service2", "check3", structs.HealthCritical)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ServiceChecks(t *testing.T) {
	s := testStateStore(t)

	// Do an initial query for a service that doesn't exist.
	ws := memdb.NewWatchSet()
	idx, checks, err := s.ServiceChecks(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if len(checks) != 0 {
		t.Fatalf("bad: %#v", checks)
	}

	// Create some nodes and checks.
	testRegisterNode(t, s, 0, "node1")
	testRegisterService(t, s, 1, "node1", "service1")
	testRegisterCheck(t, s, 2, "node1", "service1", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 3, "node1", "service1", "check2", structs.HealthPassing)
	testRegisterNode(t, s, 4, "node2")
	testRegisterService(t, s, 5, "node2", "service2")
	testRegisterCheck(t, s, 6, "node2", "service2", "check3", structs.HealthPassing)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Try querying for all checks associated with service1.
	ws = memdb.NewWatchSet()
	idx, checks, err = s.ServiceChecks(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 6 {
		t.Fatalf("bad index: %d", idx)
	}
	if len(checks) != 2 || checks[0].CheckID != "check1" || checks[1].CheckID != "check2" {
		t.Fatalf("bad checks: %#v", checks)
	}

	// Adding some unrelated service + check should not fire the watch.
	testRegisterService(t, s, 7, "node1", "service3")
	testRegisterCheck(t, s, 8, "node1", "service3", "check3", structs.HealthPassing)
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Updating a related check should fire the watch.
	testRegisterCheck(t, s, 9, "node1", "service1", "check2", structs.HealthCritical)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ServiceChecksByNodeMeta(t *testing.T) {
	s := testStateStore(t)

	// Querying with no results returns nil.
	ws := memdb.NewWatchSet()
	idx, checks, err := s.ServiceChecksByNodeMeta(ws, "service1", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 0 {
		t.Fatalf("bad: %d", idx)
	}
	if len(checks) != 0 {
		t.Fatalf("bad: %#v", checks)
	}

	// Create some nodes and checks.
	testRegisterNodeWithMeta(t, s, 0, "node1", map[string]string{"somekey": "somevalue", "common": "1"})
	testRegisterService(t, s, 1, "node1", "service1")
	testRegisterCheck(t, s, 2, "node1", "service1", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 3, "node1", "service1", "check2", structs.HealthPassing)
	testRegisterNodeWithMeta(t, s, 4, "node2", map[string]string{"common": "1"})
	testRegisterService(t, s, 5, "node2", "service1")
	testRegisterCheck(t, s, 6, "node2", "service1", "check3", structs.HealthPassing)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	cases := []struct {
		filters map[string]string
		checks  []string
	}{
		// Basic meta filter
		{
			filters: map[string]string{"somekey": "somevalue"},
			checks:  []string{"check1", "check2"},
		},
		// Common meta field
		{
			filters: map[string]string{"common": "1"},
			checks:  []string{"check1", "check2", "check3"},
		},
		// Invalid meta filter
		{
			filters: map[string]string{"invalid": "nope"},
			checks:  []string{},
		},
		// Multiple filters
		{
			filters: map[string]string{"somekey": "somevalue", "common": "1"},
			checks:  []string{"check1", "check2"},
		},
	}

	// Try querying for all checks associated with service1.
	idx = 7
	for _, tc := range cases {
		ws = memdb.NewWatchSet()
		_, checks, err := s.ServiceChecksByNodeMeta(ws, "service1", tc.filters)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if len(checks) != len(tc.checks) {
			t.Fatalf("bad checks: %#v", checks)
		}
		for i, check := range checks {
			if check.CheckID != types.CheckID(tc.checks[i]) {
				t.Fatalf("bad checks: %#v", checks)
			}
		}

		// Registering some unrelated node should not fire the watch.
		testRegisterNode(t, s, idx, fmt.Sprintf("nope%d", idx))
		idx++
		if watchFired(ws) {
			t.Fatalf("bad")
		}
	}

	// Overwhelm the node tracking.
	for i := 0; i < 2*watchLimit; i++ {
		node := fmt.Sprintf("many%d", idx)
		testRegisterNodeWithMeta(t, s, idx, node, map[string]string{"common": "1"})
		idx++
		testRegisterService(t, s, idx, node, "service1")
		idx++
		testRegisterCheck(t, s, idx, node, "service1", "check1", structs.HealthPassing)
		idx++
	}

	// Now get a fresh watch, which will be forced to watch the whole
	// node table.
	ws = memdb.NewWatchSet()
	_, _, err = s.ServiceChecksByNodeMeta(ws, "service1",
		map[string]string{"common": "1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Registering some unrelated node should now fire the watch.
	testRegisterNode(t, s, idx, "nope")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ChecksInState(t *testing.T) {
	s := testStateStore(t)

	// Querying with no results returns nil
	ws := memdb.NewWatchSet()
	idx, res, err := s.ChecksInState(ws, structs.HealthPassing)
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Register a node with checks in varied states
	testRegisterNode(t, s, 0, "node1")
	testRegisterCheck(t, s, 1, "node1", "", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 2, "node1", "", "check2", structs.HealthCritical)
	testRegisterCheck(t, s, 3, "node1", "", "check3", structs.HealthPassing)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Query the state store for passing checks.
	ws = memdb.NewWatchSet()
	_, checks, err := s.ChecksInState(ws, structs.HealthPassing)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Make sure we only get the checks which match the state
	if n := len(checks); n != 2 {
		t.Fatalf("expected 2 checks, got: %d", n)
	}
	if checks[0].CheckID != "check1" || checks[1].CheckID != "check3" {
		t.Fatalf("bad: %#v", checks)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Changing the state of a check should fire the watch.
	testRegisterCheck(t, s, 4, "node1", "", "check1", structs.HealthCritical)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// HealthAny just returns everything.
	ws = memdb.NewWatchSet()
	_, checks, err = s.ChecksInState(ws, structs.HealthAny)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if n := len(checks); n != 3 {
		t.Fatalf("expected 3 checks, got: %d", n)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Adding a new check should fire the watch.
	testRegisterCheck(t, s, 5, "node1", "", "check4", structs.HealthCritical)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_ChecksInStateByNodeMeta(t *testing.T) {
	s := testStateStore(t)

	// Querying with no results returns nil.
	ws := memdb.NewWatchSet()
	idx, res, err := s.ChecksInStateByNodeMeta(ws, structs.HealthPassing, nil)
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Register a node with checks in varied states.
	testRegisterNodeWithMeta(t, s, 0, "node1", map[string]string{"somekey": "somevalue", "common": "1"})
	testRegisterCheck(t, s, 1, "node1", "", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 2, "node1", "", "check2", structs.HealthCritical)
	testRegisterNodeWithMeta(t, s, 3, "node2", map[string]string{"common": "1"})
	testRegisterCheck(t, s, 4, "node2", "", "check3", structs.HealthPassing)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	cases := []struct {
		filters map[string]string
		state   string
		checks  []string
	}{
		// Basic meta filter, any status
		{
			filters: map[string]string{"somekey": "somevalue"},
			state:   structs.HealthAny,
			checks:  []string{"check2", "check1"},
		},
		// Basic meta filter, only passing
		{
			filters: map[string]string{"somekey": "somevalue"},
			state:   structs.HealthPassing,
			checks:  []string{"check1"},
		},
		// Common meta filter, any status
		{
			filters: map[string]string{"common": "1"},
			state:   structs.HealthAny,
			checks:  []string{"check2", "check1", "check3"},
		},
		// Common meta filter, only passing
		{
			filters: map[string]string{"common": "1"},
			state:   structs.HealthPassing,
			checks:  []string{"check1", "check3"},
		},
		// Invalid meta filter
		{
			filters: map[string]string{"invalid": "nope"},
			checks:  []string{},
		},
		// Multiple filters, any status
		{
			filters: map[string]string{"somekey": "somevalue", "common": "1"},
			state:   structs.HealthAny,
			checks:  []string{"check2", "check1"},
		},
		// Multiple filters, only passing
		{
			filters: map[string]string{"somekey": "somevalue", "common": "1"},
			state:   structs.HealthPassing,
			checks:  []string{"check1"},
		},
	}

	// Try querying for all checks associated with service1.
	idx = 5
	for _, tc := range cases {
		ws = memdb.NewWatchSet()
		_, checks, err := s.ChecksInStateByNodeMeta(ws, tc.state, tc.filters)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if len(checks) != len(tc.checks) {
			t.Fatalf("bad checks: %#v", checks)
		}
		for i, check := range checks {
			if check.CheckID != types.CheckID(tc.checks[i]) {
				t.Fatalf("bad checks: %#v, %v", checks, tc.checks)
			}
		}

		// Registering some unrelated node should not fire the watch.
		testRegisterNode(t, s, idx, fmt.Sprintf("nope%d", idx))
		idx++
		if watchFired(ws) {
			t.Fatalf("bad")
		}
	}

	// Overwhelm the node tracking.
	for i := 0; i < 2*watchLimit; i++ {
		node := fmt.Sprintf("many%d", idx)
		testRegisterNodeWithMeta(t, s, idx, node, map[string]string{"common": "1"})
		idx++
		testRegisterService(t, s, idx, node, "service1")
		idx++
		testRegisterCheck(t, s, idx, node, "service1", "check1", structs.HealthPassing)
		idx++
	}

	// Now get a fresh watch, which will be forced to watch the whole
	// node table.
	ws = memdb.NewWatchSet()
	_, _, err = s.ChecksInStateByNodeMeta(ws, structs.HealthPassing,
		map[string]string{"common": "1"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Registering some unrelated node should now fire the watch.
	testRegisterNode(t, s, idx, "nope")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_DeleteCheck(t *testing.T) {
	s := testStateStore(t)

	// Register a node and a node-level health check.
	testRegisterNode(t, s, 1, "node1")
	testRegisterCheck(t, s, 2, "node1", "", "check1", structs.HealthPassing)

	// Make sure the check is there.
	ws := memdb.NewWatchSet()
	_, checks, err := s.NodeChecks(ws, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(checks) != 1 {
		t.Fatalf("bad: %#v", checks)
	}

	// Delete the check.
	if err := s.DeleteCheck(3, "node1", "check1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Check is gone
	ws = memdb.NewWatchSet()
	_, checks, err = s.NodeChecks(ws, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(checks) != 0 {
		t.Fatalf("bad: %#v", checks)
	}

	// Index tables were updated.
	if idx := s.maxIndex("checks"); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}

	// Deleting a nonexistent check should be idempotent and not return an
	// error.
	if err := s.DeleteCheck(4, "node1", "check1"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx := s.maxIndex("checks"); idx != 3 {
		t.Fatalf("bad index: %d", idx)
	}
	if watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_CheckServiceNodes(t *testing.T) {
	s := testStateStore(t)

	// Querying with no matches gives an empty response
	ws := memdb.NewWatchSet()
	idx, res, err := s.CheckServiceNodes(ws, "service1")
	if idx != 0 || res != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, res, err)
	}

	// Register some nodes.
	testRegisterNode(t, s, 0, "node1")
	testRegisterNode(t, s, 1, "node2")

	// Register node-level checks. These should be the final result.
	testRegisterCheck(t, s, 2, "node1", "", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 3, "node2", "", "check2", structs.HealthPassing)

	// Register a service against the nodes.
	testRegisterService(t, s, 4, "node1", "service1")
	testRegisterService(t, s, 5, "node2", "service2")

	// Register checks against the services.
	testRegisterCheck(t, s, 6, "node1", "service1", "check3", structs.HealthPassing)
	testRegisterCheck(t, s, 7, "node2", "service2", "check4", structs.HealthPassing)

	// At this point all the changes should have fired the watch.
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Query the state store for nodes and checks which have been registered
	// with a specific service.
	ws = memdb.NewWatchSet()
	idx, results, err := s.CheckServiceNodes(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 7 {
		t.Fatalf("bad index: %d", idx)
	}

	// Make sure we get the expected result (service check + node check).
	if n := len(results); n != 1 {
		t.Fatalf("expected 1 result, got: %d", n)
	}
	csn := results[0]
	if csn.Node == nil || csn.Service == nil || len(csn.Checks) != 2 ||
		csn.Checks[0].ServiceID != "" || csn.Checks[0].CheckID != "check1" ||
		csn.Checks[1].ServiceID != "service1" || csn.Checks[1].CheckID != "check3" {
		t.Fatalf("bad output: %#v", csn)
	}

	// Node updates alter the returned index and fire the watch.
	testRegisterNode(t, s, 8, "node1")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
	ws = memdb.NewWatchSet()
	idx, results, err = s.CheckServiceNodes(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 8 {
		t.Fatalf("bad index: %d", idx)
	}

	// Service updates alter the returned index and fire the watch.
	testRegisterService(t, s, 9, "node1", "service1")
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
	ws = memdb.NewWatchSet()
	idx, results, err = s.CheckServiceNodes(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check updates alter the returned index and fire the watch.
	testRegisterCheck(t, s, 10, "node1", "service1", "check1", structs.HealthCritical)
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
	ws = memdb.NewWatchSet()
	idx, results, err = s.CheckServiceNodes(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 10 {
		t.Fatalf("bad index: %d", idx)
	}

	// Registering some unrelated node + service should not fire the watch.
	testRegisterNode(t, s, 11, "nope")
	testRegisterService(t, s, 12, "nope", "nope")
	if watchFired(ws) {
		t.Fatalf("bad")
	}

	// Overwhelm node and check tracking.
	idx = 13
	for i := 0; i < 2*watchLimit; i++ {
		node := fmt.Sprintf("many%d", i)
		testRegisterNode(t, s, idx, node)
		idx++
		testRegisterCheck(t, s, idx, node, "", "check1", structs.HealthPassing)
		idx++
		testRegisterService(t, s, idx, node, "service1")
		idx++
		testRegisterCheck(t, s, idx, node, "service1", "check2", structs.HealthPassing)
		idx++
	}

	// Now registering an unrelated node will fire the watch.
	ws = memdb.NewWatchSet()
	idx, results, err = s.CheckServiceNodes(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	testRegisterNode(t, s, idx, "more-nope")
	idx++
	if !watchFired(ws) {
		t.Fatalf("bad")
	}

	// Also, registering an unrelated check will fire the watch.
	ws = memdb.NewWatchSet()
	idx, results, err = s.CheckServiceNodes(ws, "service1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	testRegisterCheck(t, s, idx, "more-nope", "", "check1", structs.HealthPassing)
	idx++
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func BenchmarkCheckServiceNodes(b *testing.B) {
	s, err := NewStateStore(nil)
	if err != nil {
		b.Fatalf("err: %s", err)
	}

	if err := s.EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		b.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(2, "foo", &structs.NodeService{ID: "db1", Service: "db", Tags: []string{"master"}, Address: "", Port: 8000}); err != nil {
		b.Fatalf("err: %v", err)
	}
	check := &structs.HealthCheck{
		Node:      "foo",
		CheckID:   "db",
		Name:      "can connect",
		Status:    structs.HealthPassing,
		ServiceID: "db1",
	}
	if err := s.EnsureCheck(3, check); err != nil {
		b.Fatalf("err: %v", err)
	}
	check = &structs.HealthCheck{
		Node:    "foo",
		CheckID: "check1",
		Name:    "check1",
		Status:  structs.HealthPassing,
	}
	if err := s.EnsureCheck(4, check); err != nil {
		b.Fatalf("err: %v", err)
	}

	ws := memdb.NewWatchSet()
	for i := 0; i < b.N; i++ {
		s.CheckServiceNodes(ws, "db")
	}
}

func TestStateStore_CheckServiceTagNodes(t *testing.T) {
	s := testStateStore(t)

	if err := s.EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.1"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s.EnsureService(2, "foo", &structs.NodeService{ID: "db1", Service: "db", Tags: []string{"master"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	check := &structs.HealthCheck{
		Node:      "foo",
		CheckID:   "db",
		Name:      "can connect",
		Status:    structs.HealthPassing,
		ServiceID: "db1",
	}
	if err := s.EnsureCheck(3, check); err != nil {
		t.Fatalf("err: %v", err)
	}
	check = &structs.HealthCheck{
		Node:    "foo",
		CheckID: "check1",
		Name:    "another check",
		Status:  structs.HealthPassing,
	}
	if err := s.EnsureCheck(4, check); err != nil {
		t.Fatalf("err: %v", err)
	}

	ws := memdb.NewWatchSet()
	idx, nodes, err := s.CheckServiceTagNodes(ws, "db", "master")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 4 {
		t.Fatalf("bad: %v", idx)
	}
	if len(nodes) != 1 {
		t.Fatalf("Bad: %v", nodes)
	}
	if nodes[0].Node.Node != "foo" {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[0].Service.ID != "db1" {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if len(nodes[0].Checks) != 2 {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[0].Checks[0].CheckID != "check1" {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[0].Checks[1].CheckID != "db" {
		t.Fatalf("Bad: %v", nodes[0])
	}

	// Changing a tag should fire the watch.
	if err := s.EnsureService(4, "foo", &structs.NodeService{ID: "db1", Service: "db", Tags: []string{"nope"}, Address: "", Port: 8000}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !watchFired(ws) {
		t.Fatalf("bad")
	}
}

func TestStateStore_Check_Snapshot(t *testing.T) {
	s := testStateStore(t)

	// Create a node, a service, and a service check as well as a node check.
	testRegisterNode(t, s, 0, "node1")
	testRegisterService(t, s, 1, "node1", "service1")
	checks := structs.HealthChecks{
		&structs.HealthCheck{
			Node:    "node1",
			CheckID: "check1",
			Name:    "node check",
			Status:  structs.HealthPassing,
		},
		&structs.HealthCheck{
			Node:      "node1",
			CheckID:   "check2",
			Name:      "service check",
			Status:    structs.HealthCritical,
			ServiceID: "service1",
		},
	}
	for i, hc := range checks {
		if err := s.EnsureCheck(uint64(i+1), hc); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Create a second node/service to make sure node filtering works. This
	// will affect the index but not the dump.
	testRegisterNode(t, s, 3, "node2")
	testRegisterService(t, s, 4, "node2", "service2")
	testRegisterCheck(t, s, 5, "node2", "service2", "check3", structs.HealthPassing)

	// Snapshot the checks.
	snap := s.Snapshot()
	defer snap.Close()

	// Alter the real state store.
	testRegisterCheck(t, s, 6, "node2", "service2", "check4", structs.HealthPassing)

	// Verify the snapshot.
	if idx := snap.LastIndex(); idx != 5 {
		t.Fatalf("bad index: %d", idx)
	}
	iter, err := snap.Checks("node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	for i := 0; i < len(checks); i++ {
		check := iter.Next().(*structs.HealthCheck)
		if check == nil {
			t.Fatalf("unexpected end of checks")
		}

		checks[i].CreateIndex, checks[i].ModifyIndex = uint64(i+1), uint64(i+1)
		if !reflect.DeepEqual(check, checks[i]) {
			t.Fatalf("bad: %#v != %#v", check, checks[i])
		}
	}
	if iter.Next() != nil {
		t.Fatalf("unexpected extra checks")
	}
}

func TestStateStore_NodeInfo_NodeDump(t *testing.T) {
	s := testStateStore(t)

	// Generating a node dump that matches nothing returns empty
	wsInfo := memdb.NewWatchSet()
	idx, dump, err := s.NodeInfo(wsInfo, "node1")
	if idx != 0 || dump != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, dump, err)
	}
	wsDump := memdb.NewWatchSet()
	idx, dump, err = s.NodeDump(wsDump)
	if idx != 0 || dump != nil || err != nil {
		t.Fatalf("expected (0, nil, nil), got: (%d, %#v, %#v)", idx, dump, err)
	}

	// Register some nodes
	testRegisterNode(t, s, 0, "node1")
	testRegisterNode(t, s, 1, "node2")

	// Register services against them
	testRegisterService(t, s, 2, "node1", "service1")
	testRegisterService(t, s, 3, "node1", "service2")
	testRegisterService(t, s, 4, "node2", "service1")
	testRegisterService(t, s, 5, "node2", "service2")

	// Register service-level checks
	testRegisterCheck(t, s, 6, "node1", "service1", "check1", structs.HealthPassing)
	testRegisterCheck(t, s, 7, "node2", "service1", "check1", structs.HealthPassing)

	// Register node-level checks
	testRegisterCheck(t, s, 8, "node1", "", "check2", structs.HealthPassing)
	testRegisterCheck(t, s, 9, "node2", "", "check2", structs.HealthPassing)

	// Both watches should have fired due to the changes above.
	if !watchFired(wsInfo) {
		t.Fatalf("bad")
	}
	if !watchFired(wsDump) {
		t.Fatalf("bad")
	}

	// Check that our result matches what we expect.
	expect := structs.NodeDump{
		&structs.NodeInfo{
			Node: "node1",
			Checks: structs.HealthChecks{
				&structs.HealthCheck{
					Node:        "node1",
					CheckID:     "check1",
					ServiceID:   "service1",
					ServiceName: "service1",
					Status:      structs.HealthPassing,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 6,
						ModifyIndex: 6,
					},
				},
				&structs.HealthCheck{
					Node:        "node1",
					CheckID:     "check2",
					ServiceID:   "",
					ServiceName: "",
					Status:      structs.HealthPassing,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 8,
						ModifyIndex: 8,
					},
				},
			},
			Services: []*structs.NodeService{
				&structs.NodeService{
					ID:      "service1",
					Service: "service1",
					Address: "1.1.1.1",
					Port:    1111,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 2,
						ModifyIndex: 2,
					},
				},
				&structs.NodeService{
					ID:      "service2",
					Service: "service2",
					Address: "1.1.1.1",
					Port:    1111,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 3,
						ModifyIndex: 3,
					},
				},
			},
		},
		&structs.NodeInfo{
			Node: "node2",
			Checks: structs.HealthChecks{
				&structs.HealthCheck{
					Node:        "node2",
					CheckID:     "check1",
					ServiceID:   "service1",
					ServiceName: "service1",
					Status:      structs.HealthPassing,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 7,
						ModifyIndex: 7,
					},
				},
				&structs.HealthCheck{
					Node:        "node2",
					CheckID:     "check2",
					ServiceID:   "",
					ServiceName: "",
					Status:      structs.HealthPassing,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 9,
						ModifyIndex: 9,
					},
				},
			},
			Services: []*structs.NodeService{
				&structs.NodeService{
					ID:      "service1",
					Service: "service1",
					Address: "1.1.1.1",
					Port:    1111,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 4,
						ModifyIndex: 4,
					},
				},
				&structs.NodeService{
					ID:      "service2",
					Service: "service2",
					Address: "1.1.1.1",
					Port:    1111,
					RaftIndex: structs.RaftIndex{
						CreateIndex: 5,
						ModifyIndex: 5,
					},
				},
			},
		},
	}

	// Get a dump of just a single node
	ws := memdb.NewWatchSet()
	idx, dump, err = s.NodeInfo(ws, "node1")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", idx)
	}
	if len(dump) != 1 || !reflect.DeepEqual(dump[0], expect[0]) {
		t.Fatalf("bad: %#v", dump)
	}

	// Generate a dump of all the nodes
	idx, dump, err = s.NodeDump(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if idx != 9 {
		t.Fatalf("bad index: %d", 9)
	}
	if !reflect.DeepEqual(dump, expect) {
		t.Fatalf("bad: %#v", dump[0].Services[0])
	}

	// Registering some unrelated node + service + check should not fire the
	// watch.
	testRegisterNode(t, s, 10, "nope")
	testRegisterService(t, s, 11, "nope", "nope")
	if watchFired(ws) {
		t.Fatalf("bad")
	}
}
