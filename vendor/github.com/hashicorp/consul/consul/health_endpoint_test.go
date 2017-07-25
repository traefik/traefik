package consul

import (
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/lib"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/net-rpc-msgpackrpc"
)

func TestHealth_ChecksInState(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Check: &structs.HealthCheck{
			Name:   "memory utilization",
			Status: structs.HealthPassing,
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	var out2 structs.IndexedHealthChecks
	inState := structs.ChecksInStateRequest{
		Datacenter: "dc1",
		State:      structs.HealthPassing,
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ChecksInState", &inState, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}

	checks := out2.HealthChecks
	if len(checks) != 2 {
		t.Fatalf("Bad: %v", checks)
	}

	// Serf check is automatically added
	if checks[0].Name != "memory utilization" {
		t.Fatalf("Bad: %v", checks[0])
	}
	if checks[1].CheckID != SerfCheckID {
		t.Fatalf("Bad: %v", checks[1])
	}
}

func TestHealth_ChecksInState_NodeMetaFilter(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		NodeMeta: map[string]string{
			"somekey": "somevalue",
			"common":  "1",
		},
		Check: &structs.HealthCheck{
			Name:   "memory utilization",
			Status: structs.HealthPassing,
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	arg = structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "bar",
		Address:    "127.0.0.2",
		NodeMeta: map[string]string{
			"common": "1",
		},
		Check: &structs.HealthCheck{
			Name:   "disk space",
			Status: structs.HealthPassing,
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	cases := []struct {
		filters    map[string]string
		checkNames []string
	}{
		// Get foo's check by its unique meta value
		{
			filters:    map[string]string{"somekey": "somevalue"},
			checkNames: []string{"memory utilization"},
		},
		// Get both foo/bar's checks by their common meta value
		{
			filters:    map[string]string{"common": "1"},
			checkNames: []string{"disk space", "memory utilization"},
		},
		// Use an invalid meta value, should get empty result
		{
			filters:    map[string]string{"invalid": "nope"},
			checkNames: []string{},
		},
		// Use multiple filters to get foo's check
		{
			filters: map[string]string{
				"somekey": "somevalue",
				"common":  "1",
			},
			checkNames: []string{"memory utilization"},
		},
	}

	for _, tc := range cases {
		var out structs.IndexedHealthChecks
		inState := structs.ChecksInStateRequest{
			Datacenter:      "dc1",
			NodeMetaFilters: tc.filters,
			State:           structs.HealthPassing,
		}
		if err := msgpackrpc.CallWithCodec(codec, "Health.ChecksInState", &inState, &out); err != nil {
			t.Fatalf("err: %v", err)
		}

		checks := out.HealthChecks
		if len(checks) != len(tc.checkNames) {
			t.Fatalf("Bad: %v, %v", checks, tc.checkNames)
		}

		for i, check := range checks {
			if tc.checkNames[i] != check.Name {
				t.Fatalf("Bad: %v %v", checks, tc.checkNames)
			}
		}
	}
}

func TestHealth_ChecksInState_DistanceSort(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")
	if err := s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.2"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s1.fsm.State().EnsureNode(2, &structs.Node{Node: "bar", Address: "127.0.0.3"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	updates := structs.Coordinates{
		{"foo", lib.GenerateCoordinate(1 * time.Millisecond)},
		{"bar", lib.GenerateCoordinate(2 * time.Millisecond)},
	}
	if err := s1.fsm.State().CoordinateBatchUpdate(3, updates); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Check: &structs.HealthCheck{
			Name:   "memory utilization",
			Status: structs.HealthPassing,
		},
	}

	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg.Node = "bar"
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Query relative to foo to make sure it shows up first in the list.
	var out2 structs.IndexedHealthChecks
	inState := structs.ChecksInStateRequest{
		Datacenter: "dc1",
		State:      structs.HealthPassing,
		Source: structs.QuerySource{
			Datacenter: "dc1",
			Node:       "foo",
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ChecksInState", &inState, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}
	checks := out2.HealthChecks
	if len(checks) != 3 {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[0].Node != "foo" {
		t.Fatalf("Bad: %v", checks[1])
	}

	// Now query relative to bar to make sure it shows up first.
	inState.Source.Node = "bar"
	if err := msgpackrpc.CallWithCodec(codec, "Health.ChecksInState", &inState, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}
	checks = out2.HealthChecks
	if len(checks) != 3 {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[0].Node != "bar" {
		t.Fatalf("Bad: %v", checks[1])
	}
}

func TestHealth_NodeChecks(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Check: &structs.HealthCheck{
			Name:   "memory utilization",
			Status: structs.HealthPassing,
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	var out2 structs.IndexedHealthChecks
	node := structs.NodeSpecificRequest{
		Datacenter: "dc1",
		Node:       "foo",
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.NodeChecks", &node, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}

	checks := out2.HealthChecks
	if len(checks) != 1 {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[0].Name != "memory utilization" {
		t.Fatalf("Bad: %v", checks)
	}
}

func TestHealth_ServiceChecks(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "db connect",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	var out2 structs.IndexedHealthChecks
	node := structs.ServiceSpecificRequest{
		Datacenter:  "dc1",
		ServiceName: "db",
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceChecks", &node, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}

	checks := out2.HealthChecks
	if len(checks) != 1 {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[0].Name != "db connect" {
		t.Fatalf("Bad: %v", checks)
	}
}

func TestHealth_ServiceChecks_NodeMetaFilter(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		NodeMeta: map[string]string{
			"somekey": "somevalue",
			"common":  "1",
		},
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "memory utilization",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}
	arg = structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "bar",
		Address:    "127.0.0.2",
		NodeMeta: map[string]string{
			"common": "1",
		},
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "disk space",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	cases := []struct {
		filters    map[string]string
		checkNames []string
	}{
		// Get foo's check by its unique meta value
		{
			filters:    map[string]string{"somekey": "somevalue"},
			checkNames: []string{"memory utilization"},
		},
		// Get both foo/bar's checks by their common meta value
		{
			filters:    map[string]string{"common": "1"},
			checkNames: []string{"disk space", "memory utilization"},
		},
		// Use an invalid meta value, should get empty result
		{
			filters:    map[string]string{"invalid": "nope"},
			checkNames: []string{},
		},
		// Use multiple filters to get foo's check
		{
			filters: map[string]string{
				"somekey": "somevalue",
				"common":  "1",
			},
			checkNames: []string{"memory utilization"},
		},
	}

	for _, tc := range cases {
		var out structs.IndexedHealthChecks
		args := structs.ServiceSpecificRequest{
			Datacenter:      "dc1",
			NodeMetaFilters: tc.filters,
			ServiceName:     "db",
		}
		if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceChecks", &args, &out); err != nil {
			t.Fatalf("err: %v", err)
		}

		checks := out.HealthChecks
		if len(checks) != len(tc.checkNames) {
			t.Fatalf("Bad: %v, %v", checks, tc.checkNames)
		}

		for i, check := range checks {
			if tc.checkNames[i] != check.Name {
				t.Fatalf("Bad: %v %v", checks, tc.checkNames)
			}
		}
	}
}

func TestHealth_ServiceChecks_DistanceSort(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")
	if err := s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.2"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s1.fsm.State().EnsureNode(2, &structs.Node{Node: "bar", Address: "127.0.0.3"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	updates := structs.Coordinates{
		{"foo", lib.GenerateCoordinate(1 * time.Millisecond)},
		{"bar", lib.GenerateCoordinate(2 * time.Millisecond)},
	}
	if err := s1.fsm.State().CoordinateBatchUpdate(3, updates); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "db connect",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}

	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg.Node = "bar"
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Query relative to foo to make sure it shows up first in the list.
	var out2 structs.IndexedHealthChecks
	node := structs.ServiceSpecificRequest{
		Datacenter:  "dc1",
		ServiceName: "db",
		Source: structs.QuerySource{
			Datacenter: "dc1",
			Node:       "foo",
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceChecks", &node, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}
	checks := out2.HealthChecks
	if len(checks) != 2 {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[0].Node != "foo" {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[1].Node != "bar" {
		t.Fatalf("Bad: %v", checks)
	}

	// Now query relative to bar to make sure it shows up first.
	node.Source.Node = "bar"
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceChecks", &node, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}
	checks = out2.HealthChecks
	if len(checks) != 2 {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[0].Node != "bar" {
		t.Fatalf("Bad: %v", checks)
	}
	if checks[1].Node != "foo" {
		t.Fatalf("Bad: %v", checks)
	}
}

func TestHealth_ServiceNodes(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
			Tags:    []string{"master"},
		},
		Check: &structs.HealthCheck{
			Name:      "db connect",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg = structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "bar",
		Address:    "127.0.0.2",
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
			Tags:    []string{"slave"},
		},
		Check: &structs.HealthCheck{
			Name:      "db connect",
			Status:    structs.HealthWarning,
			ServiceID: "db",
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	var out2 structs.IndexedCheckServiceNodes
	req := structs.ServiceSpecificRequest{
		Datacenter:  "dc1",
		ServiceName: "db",
		ServiceTag:  "master",
		TagFilter:   false,
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceNodes", &req, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}

	nodes := out2.Nodes
	if len(nodes) != 2 {
		t.Fatalf("Bad: %v", nodes)
	}
	if nodes[0].Node.Node != "bar" {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[1].Node.Node != "foo" {
		t.Fatalf("Bad: %v", nodes[1])
	}
	if !lib.StrContains(nodes[0].Service.Tags, "slave") {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if !lib.StrContains(nodes[1].Service.Tags, "master") {
		t.Fatalf("Bad: %v", nodes[1])
	}
	if nodes[0].Checks[0].Status != structs.HealthWarning {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[1].Checks[0].Status != structs.HealthPassing {
		t.Fatalf("Bad: %v", nodes[1])
	}
}

func TestHealth_ServiceNodes_NodeMetaFilter(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		NodeMeta: map[string]string{
			"somekey": "somevalue",
			"common":  "1",
		},
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "memory utilization",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}
	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg = structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "bar",
		Address:    "127.0.0.2",
		NodeMeta: map[string]string{
			"common": "1",
		},
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "disk space",
			Status:    structs.HealthWarning,
			ServiceID: "db",
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	cases := []struct {
		filters map[string]string
		nodes   structs.CheckServiceNodes
	}{
		// Get foo's check by its unique meta value
		{
			filters: map[string]string{"somekey": "somevalue"},
			nodes: structs.CheckServiceNodes{
				structs.CheckServiceNode{
					Node:   &structs.Node{Node: "foo"},
					Checks: structs.HealthChecks{&structs.HealthCheck{Name: "memory utilization"}},
				},
			},
		},
		// Get both foo/bar's checks by their common meta value
		{
			filters: map[string]string{"common": "1"},
			nodes: structs.CheckServiceNodes{
				structs.CheckServiceNode{
					Node:   &structs.Node{Node: "bar"},
					Checks: structs.HealthChecks{&structs.HealthCheck{Name: "disk space"}},
				},
				structs.CheckServiceNode{
					Node:   &structs.Node{Node: "foo"},
					Checks: structs.HealthChecks{&structs.HealthCheck{Name: "memory utilization"}},
				},
			},
		},
		// Use an invalid meta value, should get empty result
		{
			filters: map[string]string{"invalid": "nope"},
			nodes:   structs.CheckServiceNodes{},
		},
		// Use multiple filters to get foo's check
		{
			filters: map[string]string{
				"somekey": "somevalue",
				"common":  "1",
			},
			nodes: structs.CheckServiceNodes{
				structs.CheckServiceNode{
					Node:   &structs.Node{Node: "foo"},
					Checks: structs.HealthChecks{&structs.HealthCheck{Name: "memory utilization"}},
				},
			},
		},
	}

	for _, tc := range cases {
		var out structs.IndexedCheckServiceNodes
		req := structs.ServiceSpecificRequest{
			Datacenter:      "dc1",
			NodeMetaFilters: tc.filters,
			ServiceName:     "db",
		}
		if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceNodes", &req, &out); err != nil {
			t.Fatalf("err: %v", err)
		}

		if len(out.Nodes) != len(tc.nodes) {
			t.Fatalf("bad: %v, %v, filters: %v", out.Nodes, tc.nodes, tc.filters)
		}

		for i, node := range out.Nodes {
			checks := tc.nodes[i].Checks
			if len(node.Checks) != len(checks) {
				t.Fatalf("bad: %v, %v, filters: %v", node.Checks, checks, tc.filters)
			}
			for j, check := range node.Checks {
				if check.Name != checks[j].Name {
					t.Fatalf("bad: %v, %v, filters: %v", check, checks[j], tc.filters)
				}
			}
		}
	}
}

func TestHealth_ServiceNodes_DistanceSort(t *testing.T) {
	dir1, s1 := testServer(t)
	defer os.RemoveAll(dir1)
	defer s1.Shutdown()
	codec := rpcClient(t, s1)
	defer codec.Close()

	testutil.WaitForLeader(t, s1.RPC, "dc1")
	if err := s1.fsm.State().EnsureNode(1, &structs.Node{Node: "foo", Address: "127.0.0.2"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := s1.fsm.State().EnsureNode(2, &structs.Node{Node: "bar", Address: "127.0.0.3"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	updates := structs.Coordinates{
		{"foo", lib.GenerateCoordinate(1 * time.Millisecond)},
		{"bar", lib.GenerateCoordinate(2 * time.Millisecond)},
	}
	if err := s1.fsm.State().CoordinateBatchUpdate(3, updates); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg := structs.RegisterRequest{
		Datacenter: "dc1",
		Node:       "foo",
		Address:    "127.0.0.1",
		Service: &structs.NodeService{
			ID:      "db",
			Service: "db",
		},
		Check: &structs.HealthCheck{
			Name:      "db connect",
			Status:    structs.HealthPassing,
			ServiceID: "db",
		},
	}

	var out struct{}
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	arg.Node = "bar"
	if err := msgpackrpc.CallWithCodec(codec, "Catalog.Register", &arg, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Query relative to foo to make sure it shows up first in the list.
	var out2 structs.IndexedCheckServiceNodes
	req := structs.ServiceSpecificRequest{
		Datacenter:  "dc1",
		ServiceName: "db",
		Source: structs.QuerySource{
			Datacenter: "dc1",
			Node:       "foo",
		},
	}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceNodes", &req, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}
	nodes := out2.Nodes
	if len(nodes) != 2 {
		t.Fatalf("Bad: %v", nodes)
	}
	if nodes[0].Node.Node != "foo" {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[1].Node.Node != "bar" {
		t.Fatalf("Bad: %v", nodes[1])
	}

	// Now query relative to bar to make sure it shows up first.
	req.Source.Node = "bar"
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceNodes", &req, &out2); err != nil {
		t.Fatalf("err: %v", err)
	}
	nodes = out2.Nodes
	if len(nodes) != 2 {
		t.Fatalf("Bad: %v", nodes)
	}
	if nodes[0].Node.Node != "bar" {
		t.Fatalf("Bad: %v", nodes[0])
	}
	if nodes[1].Node.Node != "foo" {
		t.Fatalf("Bad: %v", nodes[1])
	}
}

func TestHealth_NodeChecks_FilterACL(t *testing.T) {
	dir, token, srv, codec := testACLFilterServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer codec.Close()

	opt := structs.NodeSpecificRequest{
		Datacenter:   "dc1",
		Node:         srv.config.NodeName,
		QueryOptions: structs.QueryOptions{Token: token},
	}
	reply := structs.IndexedHealthChecks{}
	if err := msgpackrpc.CallWithCodec(codec, "Health.NodeChecks", &opt, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}
	found := false
	for _, chk := range reply.HealthChecks {
		switch chk.ServiceName {
		case "foo":
			found = true
		case "bar":
			t.Fatalf("bad: %#v", reply.HealthChecks)
		}
	}
	if !found {
		t.Fatalf("bad: %#v", reply.HealthChecks)
	}

	// We've already proven that we call the ACL filtering function so we
	// test node filtering down in acl.go for node cases. This also proves
	// that we respect the version 8 ACL flag, since the test server sets
	// that to false (the regression value of *not* changing this is better
	// for now until we change the sense of the version 8 ACL flag).
}

func TestHealth_ServiceChecks_FilterACL(t *testing.T) {
	dir, token, srv, codec := testACLFilterServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer codec.Close()

	opt := structs.ServiceSpecificRequest{
		Datacenter:   "dc1",
		ServiceName:  "foo",
		QueryOptions: structs.QueryOptions{Token: token},
	}
	reply := structs.IndexedHealthChecks{}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceChecks", &opt, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}
	found := false
	for _, chk := range reply.HealthChecks {
		if chk.ServiceName == "foo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bad: %#v", reply.HealthChecks)
	}

	opt.ServiceName = "bar"
	reply = structs.IndexedHealthChecks{}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceChecks", &opt, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(reply.HealthChecks) != 0 {
		t.Fatalf("bad: %#v", reply.HealthChecks)
	}

	// We've already proven that we call the ACL filtering function so we
	// test node filtering down in acl.go for node cases. This also proves
	// that we respect the version 8 ACL flag, since the test server sets
	// that to false (the regression value of *not* changing this is better
	// for now until we change the sense of the version 8 ACL flag).
}

func TestHealth_ServiceNodes_FilterACL(t *testing.T) {
	dir, token, srv, codec := testACLFilterServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer codec.Close()

	opt := structs.ServiceSpecificRequest{
		Datacenter:   "dc1",
		ServiceName:  "foo",
		QueryOptions: structs.QueryOptions{Token: token},
	}
	reply := structs.IndexedCheckServiceNodes{}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceNodes", &opt, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(reply.Nodes) != 1 {
		t.Fatalf("bad: %#v", reply.Nodes)
	}

	opt.ServiceName = "bar"
	reply = structs.IndexedCheckServiceNodes{}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ServiceNodes", &opt, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(reply.Nodes) != 0 {
		t.Fatalf("bad: %#v", reply.Nodes)
	}

	// We've already proven that we call the ACL filtering function so we
	// test node filtering down in acl.go for node cases. This also proves
	// that we respect the version 8 ACL flag, since the test server sets
	// that to false (the regression value of *not* changing this is better
	// for now until we change the sense of the version 8 ACL flag).
}

func TestHealth_ChecksInState_FilterACL(t *testing.T) {
	dir, token, srv, codec := testACLFilterServer(t)
	defer os.RemoveAll(dir)
	defer srv.Shutdown()
	defer codec.Close()

	opt := structs.ChecksInStateRequest{
		Datacenter:   "dc1",
		State:        structs.HealthPassing,
		QueryOptions: structs.QueryOptions{Token: token},
	}
	reply := structs.IndexedHealthChecks{}
	if err := msgpackrpc.CallWithCodec(codec, "Health.ChecksInState", &opt, &reply); err != nil {
		t.Fatalf("err: %s", err)
	}

	found := false
	for _, chk := range reply.HealthChecks {
		switch chk.ServiceName {
		case "foo":
			found = true
		case "bar":
			t.Fatalf("bad service 'bar': %#v", reply.HealthChecks)
		}
	}
	if !found {
		t.Fatalf("missing service 'foo': %#v", reply.HealthChecks)
	}

	// We've already proven that we call the ACL filtering function so we
	// test node filtering down in acl.go for node cases. This also proves
	// that we respect the version 8 ACL flag, since the test server sets
	// that to false (the regression value of *not* changing this is better
	// for now until we change the sense of the version 8 ACL flag).
}
