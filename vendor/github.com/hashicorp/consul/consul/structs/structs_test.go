package structs

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/consul/types"
)

func TestEncodeDecode(t *testing.T) {
	arg := &RegisterRequest{
		Datacenter: "foo",
		Node:       "bar",
		Address:    "baz",
		Service: &NodeService{
			Service: "test",
			Address: "127.0.0.2",
		},
	}
	buf, err := Encode(RegisterRequestType, arg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	var out RegisterRequest
	err = Decode(buf[1:], &out)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if !reflect.DeepEqual(arg.Service, out.Service) {
		t.Fatalf("bad: %#v %#v", arg.Service, out.Service)
	}
	if !reflect.DeepEqual(arg, &out) {
		t.Fatalf("bad: %#v %#v", arg, out)
	}
}

func TestStructs_Implements(t *testing.T) {
	var (
		_ RPCInfo          = &RegisterRequest{}
		_ RPCInfo          = &DeregisterRequest{}
		_ RPCInfo          = &DCSpecificRequest{}
		_ RPCInfo          = &ServiceSpecificRequest{}
		_ RPCInfo          = &NodeSpecificRequest{}
		_ RPCInfo          = &ChecksInStateRequest{}
		_ RPCInfo          = &KVSRequest{}
		_ RPCInfo          = &KeyRequest{}
		_ RPCInfo          = &KeyListRequest{}
		_ RPCInfo          = &SessionRequest{}
		_ RPCInfo          = &SessionSpecificRequest{}
		_ RPCInfo          = &EventFireRequest{}
		_ RPCInfo          = &ACLPolicyRequest{}
		_ RPCInfo          = &KeyringRequest{}
		_ CompoundResponse = &KeyringResponses{}
	)
}

func TestStructs_ACL_IsSame(t *testing.T) {
	acl := &ACL{
		ID:    "guid",
		Name:  "An ACL for testing",
		Type:  "client",
		Rules: "service \"\" { policy = \"read\" }",
	}
	if !acl.IsSame(acl) {
		t.Fatalf("should be equal to itself")
	}

	other := &ACL{
		ID:    "guid",
		Name:  "An ACL for testing",
		Type:  "client",
		Rules: "service \"\" { policy = \"read\" }",
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
	if !acl.IsSame(other) || !other.IsSame(acl) {
		t.Fatalf("should not care about Raft fields")
	}

	check := func(twiddle, restore func()) {
		if !acl.IsSame(other) || !other.IsSame(acl) {
			t.Fatalf("should be the same")
		}

		twiddle()
		if acl.IsSame(other) || other.IsSame(acl) {
			t.Fatalf("should not be the same")
		}

		restore()
		if !acl.IsSame(other) || !other.IsSame(acl) {
			t.Fatalf("should be the same")
		}
	}

	check(func() { other.ID = "nope" }, func() { other.ID = "guid" })
	check(func() { other.Name = "nope" }, func() { other.Name = "An ACL for testing" })
	check(func() { other.Type = "management" }, func() { other.Type = "client" })
	check(func() { other.Rules = "" }, func() { other.Rules = "service \"\" { policy = \"read\" }" })
}

func TestStructs_RegisterRequest_ChangesNode(t *testing.T) {
	req := &RegisterRequest{
		ID:              types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5"),
		Node:            "test",
		Address:         "127.0.0.1",
		TaggedAddresses: make(map[string]string),
		NodeMeta: map[string]string{
			"role": "server",
		},
	}

	node := &Node{
		ID:              types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5"),
		Node:            "test",
		Address:         "127.0.0.1",
		TaggedAddresses: make(map[string]string),
		Meta: map[string]string{
			"role": "server",
		},
	}

	check := func(twiddle, restore func()) {
		if req.ChangesNode(node) {
			t.Fatalf("should not change")
		}

		twiddle()
		if !req.ChangesNode(node) {
			t.Fatalf("should change")
		}

		req.SkipNodeUpdate = true
		if req.ChangesNode(node) {
			t.Fatalf("should skip")
		}

		req.SkipNodeUpdate = false
		if !req.ChangesNode(node) {
			t.Fatalf("should change")
		}

		restore()
		if req.ChangesNode(node) {
			t.Fatalf("should not change")
		}
	}

	check(func() { req.ID = "nope" }, func() { req.ID = types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5") })
	check(func() { req.Node = "nope" }, func() { req.Node = "test" })
	check(func() { req.Address = "127.0.0.2" }, func() { req.Address = "127.0.0.1" })
	check(func() { req.TaggedAddresses["wan"] = "nope" }, func() { delete(req.TaggedAddresses, "wan") })
	check(func() { req.NodeMeta["invalid"] = "nope" }, func() { delete(req.NodeMeta, "invalid") })

	if !req.ChangesNode(nil) {
		t.Fatalf("should change")
	}
}

// testServiceNode gives a fully filled out ServiceNode instance.
func testServiceNode() *ServiceNode {
	return &ServiceNode{
		ID:      types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5"),
		Node:    "node1",
		Address: "127.0.0.1",
		TaggedAddresses: map[string]string{
			"hello": "world",
		},
		NodeMeta: map[string]string{
			"tag": "value",
		},
		ServiceID:                "service1",
		ServiceName:              "dogs",
		ServiceTags:              []string{"prod", "v1"},
		ServiceAddress:           "127.0.0.2",
		ServicePort:              8080,
		ServiceEnableTagOverride: true,
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
}

func TestStructs_ServiceNode_PartialClone(t *testing.T) {
	sn := testServiceNode()

	clone := sn.PartialClone()

	// Make sure the parts that weren't supposed to be cloned didn't get
	// copied over, then zero-value them out so we can do a DeepEqual() on
	// the rest of the contents.
	if clone.ID != "" ||
		clone.Address != "" ||
		len(clone.TaggedAddresses) != 0 ||
		len(clone.NodeMeta) != 0 {
		t.Fatalf("bad: %v", clone)
	}

	sn.ID = ""
	sn.Address = ""
	sn.TaggedAddresses = nil
	sn.NodeMeta = nil
	if !reflect.DeepEqual(sn, clone) {
		t.Fatalf("bad: %v", clone)
	}

	sn.ServiceTags = append(sn.ServiceTags, "hello")
	if reflect.DeepEqual(sn, clone) {
		t.Fatalf("clone wasn't independent of the original")
	}
}

func TestStructs_ServiceNode_Conversions(t *testing.T) {
	sn := testServiceNode()

	sn2 := sn.ToNodeService().ToServiceNode("node1")

	// These two fields get lost in the conversion, so we have to zero-value
	// them out before we do the compare.
	sn.ID = ""
	sn.Address = ""
	sn.TaggedAddresses = nil
	sn.NodeMeta = nil
	if !reflect.DeepEqual(sn, sn2) {
		t.Fatalf("bad: %v", sn2)
	}
}

func TestStructs_NodeService_IsSame(t *testing.T) {
	ns := &NodeService{
		ID:                "node1",
		Service:           "theservice",
		Tags:              []string{"foo", "bar"},
		Address:           "127.0.0.1",
		Port:              1234,
		EnableTagOverride: true,
	}
	if !ns.IsSame(ns) {
		t.Fatalf("should be equal to itself")
	}

	other := &NodeService{
		ID:                "node1",
		Service:           "theservice",
		Tags:              []string{"foo", "bar"},
		Address:           "127.0.0.1",
		Port:              1234,
		EnableTagOverride: true,
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
	if !ns.IsSame(other) || !other.IsSame(ns) {
		t.Fatalf("should not care about Raft fields")
	}

	check := func(twiddle, restore func()) {
		if !ns.IsSame(other) || !other.IsSame(ns) {
			t.Fatalf("should be the same")
		}

		twiddle()
		if ns.IsSame(other) || other.IsSame(ns) {
			t.Fatalf("should not be the same")
		}

		restore()
		if !ns.IsSame(other) || !other.IsSame(ns) {
			t.Fatalf("should be the same")
		}
	}

	check(func() { other.ID = "XXX" }, func() { other.ID = "node1" })
	check(func() { other.Service = "XXX" }, func() { other.Service = "theservice" })
	check(func() { other.Tags = nil }, func() { other.Tags = []string{"foo", "bar"} })
	check(func() { other.Tags = []string{"foo"} }, func() { other.Tags = []string{"foo", "bar"} })
	check(func() { other.Address = "XXX" }, func() { other.Address = "127.0.0.1" })
	check(func() { other.Port = 9999 }, func() { other.Port = 1234 })
	check(func() { other.EnableTagOverride = false }, func() { other.EnableTagOverride = true })
}

func TestStructs_HealthCheck_IsSame(t *testing.T) {
	hc := &HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "thecheck",
		Status:      HealthPassing,
		Notes:       "it's all good",
		Output:      "lgtm",
		ServiceID:   "service1",
		ServiceName: "theservice",
	}
	if !hc.IsSame(hc) {
		t.Fatalf("should be equal to itself")
	}

	other := &HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "thecheck",
		Status:      HealthPassing,
		Notes:       "it's all good",
		Output:      "lgtm",
		ServiceID:   "service1",
		ServiceName: "theservice",
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
	if !hc.IsSame(other) || !other.IsSame(hc) {
		t.Fatalf("should not care about Raft fields")
	}

	checkCheckIDField := func(field *types.CheckID) {
		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}

		old := *field
		*field = "XXX"
		if hc.IsSame(other) || other.IsSame(hc) {
			t.Fatalf("should not be the same")
		}
		*field = old

		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}
	}

	checkStringField := func(field *string) {
		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}

		old := *field
		*field = "XXX"
		if hc.IsSame(other) || other.IsSame(hc) {
			t.Fatalf("should not be the same")
		}
		*field = old

		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}
	}

	checkStringField(&other.Node)
	checkCheckIDField(&other.CheckID)
	checkStringField(&other.Name)
	checkStringField(&other.Status)
	checkStringField(&other.Notes)
	checkStringField(&other.Output)
	checkStringField(&other.ServiceID)
	checkStringField(&other.ServiceName)
}

func TestStructs_HealthCheck_Clone(t *testing.T) {
	hc := &HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "thecheck",
		Status:      HealthPassing,
		Notes:       "it's all good",
		Output:      "lgtm",
		ServiceID:   "service1",
		ServiceName: "theservice",
	}
	clone := hc.Clone()
	if !hc.IsSame(clone) {
		t.Fatalf("should be equal to its clone")
	}

	clone.Output = "different"
	if hc.IsSame(clone) {
		t.Fatalf("should not longer be equal to its clone")
	}
}

func TestStructs_CheckServiceNodes_Shuffle(t *testing.T) {
	// Make a huge list of nodes.
	var nodes CheckServiceNodes
	for i := 0; i < 100; i++ {
		nodes = append(nodes, CheckServiceNode{
			Node: &Node{
				Node:    fmt.Sprintf("node%d", i),
				Address: fmt.Sprintf("127.0.0.%d", i+1),
			},
		})
	}

	// Keep track of how many unique shuffles we get.
	uniques := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		nodes.Shuffle()

		var names []string
		for _, node := range nodes {
			names = append(names, node.Node.Node)
		}
		key := strings.Join(names, "|")
		uniques[key] = struct{}{}
	}

	// We have to allow for the fact that there won't always be a unique
	// shuffle each pass, so we just look for smell here without the test
	// being flaky.
	if len(uniques) < 50 {
		t.Fatalf("unique shuffle ratio too low: %d/100", len(uniques))
	}
}

func TestStructs_CheckServiceNodes_Filter(t *testing.T) {
	nodes := CheckServiceNodes{
		CheckServiceNode{
			Node: &Node{
				Node:    "node1",
				Address: "127.0.0.1",
			},
			Checks: HealthChecks{
				&HealthCheck{
					Status: HealthWarning,
				},
			},
		},
		CheckServiceNode{
			Node: &Node{
				Node:    "node2",
				Address: "127.0.0.2",
			},
			Checks: HealthChecks{
				&HealthCheck{
					Status: HealthPassing,
				},
			},
		},
		CheckServiceNode{
			Node: &Node{
				Node:    "node3",
				Address: "127.0.0.3",
			},
			Checks: HealthChecks{
				&HealthCheck{
					Status: HealthCritical,
				},
			},
		},
	}

	// Test the case where warnings are allowed.
	{
		twiddle := make(CheckServiceNodes, len(nodes))
		if n := copy(twiddle, nodes); n != len(nodes) {
			t.Fatalf("bad: %d", n)
		}
		filtered := twiddle.Filter(false)
		expected := CheckServiceNodes{
			nodes[0],
			nodes[1],
		}
		if !reflect.DeepEqual(filtered, expected) {
			t.Fatalf("bad: %v", filtered)
		}
	}

	// Limit to only passing checks.
	{
		twiddle := make(CheckServiceNodes, len(nodes))
		if n := copy(twiddle, nodes); n != len(nodes) {
			t.Fatalf("bad: %d", n)
		}
		filtered := twiddle.Filter(true)
		expected := CheckServiceNodes{
			nodes[1],
		}
		if !reflect.DeepEqual(filtered, expected) {
			t.Fatalf("bad: %v", filtered)
		}
	}
}

func TestStructs_DirEntry_Clone(t *testing.T) {
	e := &DirEntry{
		LockIndex: 5,
		Key:       "hello",
		Flags:     23,
		Value:     []byte("this is a test"),
		Session:   "session1",
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}

	clone := e.Clone()
	if !reflect.DeepEqual(e, clone) {
		t.Fatalf("bad: %v", clone)
	}

	e.Value = []byte("a new value")
	if reflect.DeepEqual(e, clone) {
		t.Fatalf("clone wasn't independent of the original")
	}
}

func TestStructs_ValidateMetadata(t *testing.T) {
	// Load a valid set of key/value pairs
	meta := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	// Should succeed
	if err := ValidateMetadata(meta); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Should get error
	meta = map[string]string{
		"": "value1",
	}
	if err := ValidateMetadata(meta); !strings.Contains(err.Error(), "Couldn't load metadata pair") {
		t.Fatalf("should have failed")
	}

	// Should get error
	meta = make(map[string]string)
	for i := 0; i < metaMaxKeyPairs+1; i++ {
		meta[string(i)] = "value"
	}
	if err := ValidateMetadata(meta); !strings.Contains(err.Error(), "cannot contain more than") {
		t.Fatalf("should have failed")
	}
}

func TestStructs_validateMetaPair(t *testing.T) {
	longKey := strings.Repeat("a", metaKeyMaxLength+1)
	longValue := strings.Repeat("b", metaValueMaxLength+1)
	pairs := []struct {
		Key   string
		Value string
		Error string
	}{
		// valid pair
		{"key", "value", ""},
		// invalid, blank key
		{"", "value", "cannot be blank"},
		// allowed special chars in key name
		{"k_e-y", "value", ""},
		// disallowed special chars in key name
		{"(%key&)", "value", "invalid characters"},
		// key too long
		{longKey, "value", "Key is too long"},
		// reserved prefix
		{metaKeyReservedPrefix + "key", "value", "reserved for internal use"},
		// value too long
		{"key", longValue, "Value is too long"},
	}

	for _, pair := range pairs {
		err := validateMetaPair(pair.Key, pair.Value)
		if pair.Error == "" && err != nil {
			t.Fatalf("should have succeeded: %v, %v", pair, err)
		} else if pair.Error != "" && !strings.Contains(err.Error(), pair.Error) {
			t.Fatalf("should have failed: %v, %v", pair, err)
		}
	}
}
