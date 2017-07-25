package acl

import (
	"testing"
)

func TestRootACL(t *testing.T) {
	if RootACL("allow") != AllowAll() {
		t.Fatalf("Bad root")
	}
	if RootACL("deny") != DenyAll() {
		t.Fatalf("Bad root")
	}
	if RootACL("manage") != ManageAll() {
		t.Fatalf("Bad root")
	}
	if RootACL("foo") != nil {
		t.Fatalf("bad root")
	}
}

func TestStaticACL(t *testing.T) {
	all := AllowAll()
	if _, ok := all.(*StaticACL); !ok {
		t.Fatalf("expected static")
	}

	none := DenyAll()
	if _, ok := none.(*StaticACL); !ok {
		t.Fatalf("expected static")
	}

	manage := ManageAll()
	if _, ok := none.(*StaticACL); !ok {
		t.Fatalf("expected static")
	}

	if all.ACLList() {
		t.Fatalf("should not allow")
	}
	if all.ACLModify() {
		t.Fatalf("should not allow")
	}
	if !all.AgentRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.AgentWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !all.EventRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.EventWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !all.KeyRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.KeyWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !all.KeyringRead() {
		t.Fatalf("should allow")
	}
	if !all.KeyringWrite() {
		t.Fatalf("should allow")
	}
	if !all.NodeRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.NodeWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !all.OperatorRead() {
		t.Fatalf("should allow")
	}
	if !all.OperatorWrite() {
		t.Fatalf("should allow")
	}
	if !all.PreparedQueryRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.PreparedQueryWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !all.ServiceRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.ServiceWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !all.SessionRead("foobar") {
		t.Fatalf("should allow")
	}
	if !all.SessionWrite("foobar") {
		t.Fatalf("should allow")
	}
	if all.Snapshot() {
		t.Fatalf("should not allow")
	}

	if none.ACLList() {
		t.Fatalf("should not allow")
	}
	if none.ACLModify() {
		t.Fatalf("should not allow")
	}
	if none.AgentRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.AgentWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.EventRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.EventRead("") {
		t.Fatalf("should not allow")
	}
	if none.EventWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.EventWrite("") {
		t.Fatalf("should not allow")
	}
	if none.KeyRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.KeyWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.KeyringRead() {
		t.Fatalf("should now allow")
	}
	if none.KeyringWrite() {
		t.Fatalf("should not allow")
	}
	if none.NodeRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.NodeWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.OperatorRead() {
		t.Fatalf("should now allow")
	}
	if none.OperatorWrite() {
		t.Fatalf("should not allow")
	}
	if none.PreparedQueryRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.PreparedQueryWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.ServiceRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.ServiceWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.SessionRead("foobar") {
		t.Fatalf("should not allow")
	}
	if none.SessionWrite("foobar") {
		t.Fatalf("should not allow")
	}
	if none.Snapshot() {
		t.Fatalf("should not allow")
	}

	if !manage.ACLList() {
		t.Fatalf("should allow")
	}
	if !manage.ACLModify() {
		t.Fatalf("should allow")
	}
	if !manage.AgentRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.AgentWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.EventRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.EventWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.KeyRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.KeyWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.KeyringRead() {
		t.Fatalf("should allow")
	}
	if !manage.KeyringWrite() {
		t.Fatalf("should allow")
	}
	if !manage.NodeRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.NodeWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.OperatorRead() {
		t.Fatalf("should allow")
	}
	if !manage.OperatorWrite() {
		t.Fatalf("should allow")
	}
	if !manage.PreparedQueryRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.PreparedQueryWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.ServiceRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.ServiceWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.SessionRead("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.SessionWrite("foobar") {
		t.Fatalf("should allow")
	}
	if !manage.Snapshot() {
		t.Fatalf("should allow")
	}
}

func TestPolicyACL(t *testing.T) {
	all := AllowAll()
	policy := &Policy{
		Events: []*EventPolicy{
			&EventPolicy{
				Event:  "",
				Policy: PolicyRead,
			},
			&EventPolicy{
				Event:  "foo",
				Policy: PolicyWrite,
			},
			&EventPolicy{
				Event:  "bar",
				Policy: PolicyDeny,
			},
		},
		Keys: []*KeyPolicy{
			&KeyPolicy{
				Prefix: "foo/",
				Policy: PolicyWrite,
			},
			&KeyPolicy{
				Prefix: "foo/priv/",
				Policy: PolicyDeny,
			},
			&KeyPolicy{
				Prefix: "bar/",
				Policy: PolicyDeny,
			},
			&KeyPolicy{
				Prefix: "zip/",
				Policy: PolicyRead,
			},
		},
		PreparedQueries: []*PreparedQueryPolicy{
			&PreparedQueryPolicy{
				Prefix: "",
				Policy: PolicyRead,
			},
			&PreparedQueryPolicy{
				Prefix: "foo",
				Policy: PolicyWrite,
			},
			&PreparedQueryPolicy{
				Prefix: "bar",
				Policy: PolicyDeny,
			},
			&PreparedQueryPolicy{
				Prefix: "zoo",
				Policy: PolicyWrite,
			},
		},
		Services: []*ServicePolicy{
			&ServicePolicy{
				Name:   "",
				Policy: PolicyWrite,
			},
			&ServicePolicy{
				Name:   "foo",
				Policy: PolicyRead,
			},
			&ServicePolicy{
				Name:   "bar",
				Policy: PolicyDeny,
			},
			&ServicePolicy{
				Name:   "barfoo",
				Policy: PolicyWrite,
			},
		},
	}
	acl, err := New(all, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	type keycase struct {
		inp         string
		read        bool
		write       bool
		writePrefix bool
	}
	cases := []keycase{
		{"other", true, true, true},
		{"foo/test", true, true, true},
		{"foo/priv/test", false, false, false},
		{"bar/any", false, false, false},
		{"zip/test", true, false, false},
		{"foo/", true, true, false},
		{"", true, true, false},
	}
	for _, c := range cases {
		if c.read != acl.KeyRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.KeyWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
		if c.writePrefix != acl.KeyWritePrefix(c.inp) {
			t.Fatalf("Write prefix fail: %#v", c)
		}
	}

	// Test the services
	type servicecase struct {
		inp   string
		read  bool
		write bool
	}
	scases := []servicecase{
		{"other", true, true},
		{"foo", true, false},
		{"bar", false, false},
		{"foobar", true, false},
		{"barfo", false, false},
		{"barfoo", true, true},
		{"barfoo2", true, true},
	}
	for _, c := range scases {
		if c.read != acl.ServiceRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.ServiceWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
	}

	// Test the events
	type eventcase struct {
		inp   string
		read  bool
		write bool
	}
	eventcases := []eventcase{
		{"foo", true, true},
		{"foobar", true, true},
		{"bar", false, false},
		{"barbaz", false, false},
		{"baz", true, false},
	}
	for _, c := range eventcases {
		if c.read != acl.EventRead(c.inp) {
			t.Fatalf("Event fail: %#v", c)
		}
		if c.write != acl.EventWrite(c.inp) {
			t.Fatalf("Event fail: %#v", c)
		}
	}

	// Test prepared queries
	type querycase struct {
		inp   string
		read  bool
		write bool
	}
	querycases := []querycase{
		{"foo", true, true},
		{"foobar", true, true},
		{"bar", false, false},
		{"barbaz", false, false},
		{"baz", true, false},
		{"nope", true, false},
		{"zoo", true, true},
		{"zookeeper", true, true},
	}
	for _, c := range querycases {
		if c.read != acl.PreparedQueryRead(c.inp) {
			t.Fatalf("Prepared query fail: %#v", c)
		}
		if c.write != acl.PreparedQueryWrite(c.inp) {
			t.Fatalf("Prepared query fail: %#v", c)
		}
	}
}

func TestPolicyACL_Parent(t *testing.T) {
	deny := DenyAll()
	policyRoot := &Policy{
		Keys: []*KeyPolicy{
			&KeyPolicy{
				Prefix: "foo/",
				Policy: PolicyWrite,
			},
			&KeyPolicy{
				Prefix: "bar/",
				Policy: PolicyRead,
			},
		},
		PreparedQueries: []*PreparedQueryPolicy{
			&PreparedQueryPolicy{
				Prefix: "other",
				Policy: PolicyWrite,
			},
			&PreparedQueryPolicy{
				Prefix: "foo",
				Policy: PolicyRead,
			},
		},
		Services: []*ServicePolicy{
			&ServicePolicy{
				Name:   "other",
				Policy: PolicyWrite,
			},
			&ServicePolicy{
				Name:   "foo",
				Policy: PolicyRead,
			},
		},
	}
	root, err := New(deny, policyRoot)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	policy := &Policy{
		Keys: []*KeyPolicy{
			&KeyPolicy{
				Prefix: "foo/priv/",
				Policy: PolicyRead,
			},
			&KeyPolicy{
				Prefix: "bar/",
				Policy: PolicyDeny,
			},
			&KeyPolicy{
				Prefix: "zip/",
				Policy: PolicyRead,
			},
		},
		PreparedQueries: []*PreparedQueryPolicy{
			&PreparedQueryPolicy{
				Prefix: "bar",
				Policy: PolicyDeny,
			},
		},
		Services: []*ServicePolicy{
			&ServicePolicy{
				Name:   "bar",
				Policy: PolicyDeny,
			},
		},
	}
	acl, err := New(root, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	type keycase struct {
		inp         string
		read        bool
		write       bool
		writePrefix bool
	}
	cases := []keycase{
		{"other", false, false, false},
		{"foo/test", true, true, true},
		{"foo/priv/test", true, false, false},
		{"bar/any", false, false, false},
		{"zip/test", true, false, false},
	}
	for _, c := range cases {
		if c.read != acl.KeyRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.KeyWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
		if c.writePrefix != acl.KeyWritePrefix(c.inp) {
			t.Fatalf("Write prefix fail: %#v", c)
		}
	}

	// Test the services
	type servicecase struct {
		inp   string
		read  bool
		write bool
	}
	scases := []servicecase{
		{"fail", false, false},
		{"other", true, true},
		{"foo", true, false},
		{"bar", false, false},
	}
	for _, c := range scases {
		if c.read != acl.ServiceRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.ServiceWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
	}

	// Test prepared queries
	type querycase struct {
		inp   string
		read  bool
		write bool
	}
	querycases := []querycase{
		{"foo", true, false},
		{"foobar", true, false},
		{"bar", false, false},
		{"barbaz", false, false},
		{"baz", false, false},
		{"nope", false, false},
	}
	for _, c := range querycases {
		if c.read != acl.PreparedQueryRead(c.inp) {
			t.Fatalf("Prepared query fail: %#v", c)
		}
		if c.write != acl.PreparedQueryWrite(c.inp) {
			t.Fatalf("Prepared query fail: %#v", c)
		}
	}

	// Check some management functions that chain up
	if acl.ACLList() {
		t.Fatalf("should not allow")
	}
	if acl.ACLModify() {
		t.Fatalf("should not allow")
	}
	if acl.Snapshot() {
		t.Fatalf("should not allow")
	}
}

func TestPolicyACL_Agent(t *testing.T) {
	deny := DenyAll()
	policyRoot := &Policy{
		Agents: []*AgentPolicy{
			&AgentPolicy{
				Node:   "root-nope",
				Policy: PolicyDeny,
			},
			&AgentPolicy{
				Node:   "root-ro",
				Policy: PolicyRead,
			},
			&AgentPolicy{
				Node:   "root-rw",
				Policy: PolicyWrite,
			},
			&AgentPolicy{
				Node:   "override",
				Policy: PolicyDeny,
			},
		},
	}
	root, err := New(deny, policyRoot)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	policy := &Policy{
		Agents: []*AgentPolicy{
			&AgentPolicy{
				Node:   "child-nope",
				Policy: PolicyDeny,
			},
			&AgentPolicy{
				Node:   "child-ro",
				Policy: PolicyRead,
			},
			&AgentPolicy{
				Node:   "child-rw",
				Policy: PolicyWrite,
			},
			&AgentPolicy{
				Node:   "override",
				Policy: PolicyWrite,
			},
		},
	}
	acl, err := New(root, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	type agentcase struct {
		inp   string
		read  bool
		write bool
	}
	cases := []agentcase{
		{"nope", false, false},
		{"root-nope", false, false},
		{"root-ro", true, false},
		{"root-rw", true, true},
		{"root-nope-prefix", false, false},
		{"root-ro-prefix", true, false},
		{"root-rw-prefix", true, true},
		{"child-nope", false, false},
		{"child-ro", true, false},
		{"child-rw", true, true},
		{"child-nope-prefix", false, false},
		{"child-ro-prefix", true, false},
		{"child-rw-prefix", true, true},
		{"override", true, true},
	}
	for _, c := range cases {
		if c.read != acl.AgentRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.AgentWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
	}
}

func TestPolicyACL_Keyring(t *testing.T) {
	type keyringcase struct {
		inp   string
		read  bool
		write bool
	}
	cases := []keyringcase{
		{"", false, false},
		{PolicyRead, true, false},
		{PolicyWrite, true, true},
		{PolicyDeny, false, false},
	}
	for _, c := range cases {
		acl, err := New(DenyAll(), &Policy{Keyring: c.inp})
		if err != nil {
			t.Fatalf("bad: %s", err)
		}
		if acl.KeyringRead() != c.read {
			t.Fatalf("bad: %#v", c)
		}
		if acl.KeyringWrite() != c.write {
			t.Fatalf("bad: %#v", c)
		}
	}
}

func TestPolicyACL_Operator(t *testing.T) {
	type operatorcase struct {
		inp   string
		read  bool
		write bool
	}
	cases := []operatorcase{
		{"", false, false},
		{PolicyRead, true, false},
		{PolicyWrite, true, true},
		{PolicyDeny, false, false},
	}
	for _, c := range cases {
		acl, err := New(DenyAll(), &Policy{Operator: c.inp})
		if err != nil {
			t.Fatalf("bad: %s", err)
		}
		if acl.OperatorRead() != c.read {
			t.Fatalf("bad: %#v", c)
		}
		if acl.OperatorWrite() != c.write {
			t.Fatalf("bad: %#v", c)
		}
	}
}

func TestPolicyACL_Node(t *testing.T) {
	deny := DenyAll()
	policyRoot := &Policy{
		Nodes: []*NodePolicy{
			&NodePolicy{
				Name:   "root-nope",
				Policy: PolicyDeny,
			},
			&NodePolicy{
				Name:   "root-ro",
				Policy: PolicyRead,
			},
			&NodePolicy{
				Name:   "root-rw",
				Policy: PolicyWrite,
			},
			&NodePolicy{
				Name:   "override",
				Policy: PolicyDeny,
			},
		},
	}
	root, err := New(deny, policyRoot)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	policy := &Policy{
		Nodes: []*NodePolicy{
			&NodePolicy{
				Name:   "child-nope",
				Policy: PolicyDeny,
			},
			&NodePolicy{
				Name:   "child-ro",
				Policy: PolicyRead,
			},
			&NodePolicy{
				Name:   "child-rw",
				Policy: PolicyWrite,
			},
			&NodePolicy{
				Name:   "override",
				Policy: PolicyWrite,
			},
		},
	}
	acl, err := New(root, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	type nodecase struct {
		inp   string
		read  bool
		write bool
	}
	cases := []nodecase{
		{"nope", false, false},
		{"root-nope", false, false},
		{"root-ro", true, false},
		{"root-rw", true, true},
		{"root-nope-prefix", false, false},
		{"root-ro-prefix", true, false},
		{"root-rw-prefix", true, true},
		{"child-nope", false, false},
		{"child-ro", true, false},
		{"child-rw", true, true},
		{"child-nope-prefix", false, false},
		{"child-ro-prefix", true, false},
		{"child-rw-prefix", true, true},
		{"override", true, true},
	}
	for _, c := range cases {
		if c.read != acl.NodeRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.NodeWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
	}
}

func TestPolicyACL_Session(t *testing.T) {
	deny := DenyAll()
	policyRoot := &Policy{
		Sessions: []*SessionPolicy{
			&SessionPolicy{
				Node:   "root-nope",
				Policy: PolicyDeny,
			},
			&SessionPolicy{
				Node:   "root-ro",
				Policy: PolicyRead,
			},
			&SessionPolicy{
				Node:   "root-rw",
				Policy: PolicyWrite,
			},
			&SessionPolicy{
				Node:   "override",
				Policy: PolicyDeny,
			},
		},
	}
	root, err := New(deny, policyRoot)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	policy := &Policy{
		Sessions: []*SessionPolicy{
			&SessionPolicy{
				Node:   "child-nope",
				Policy: PolicyDeny,
			},
			&SessionPolicy{
				Node:   "child-ro",
				Policy: PolicyRead,
			},
			&SessionPolicy{
				Node:   "child-rw",
				Policy: PolicyWrite,
			},
			&SessionPolicy{
				Node:   "override",
				Policy: PolicyWrite,
			},
		},
	}
	acl, err := New(root, policy)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	type sessioncase struct {
		inp   string
		read  bool
		write bool
	}
	cases := []sessioncase{
		{"nope", false, false},
		{"root-nope", false, false},
		{"root-ro", true, false},
		{"root-rw", true, true},
		{"root-nope-prefix", false, false},
		{"root-ro-prefix", true, false},
		{"root-rw-prefix", true, true},
		{"child-nope", false, false},
		{"child-ro", true, false},
		{"child-rw", true, true},
		{"child-nope-prefix", false, false},
		{"child-ro-prefix", true, false},
		{"child-rw-prefix", true, true},
		{"override", true, true},
	}
	for _, c := range cases {
		if c.read != acl.SessionRead(c.inp) {
			t.Fatalf("Read fail: %#v", c)
		}
		if c.write != acl.SessionWrite(c.inp) {
			t.Fatalf("Write fail: %#v", c)
		}
	}
}
