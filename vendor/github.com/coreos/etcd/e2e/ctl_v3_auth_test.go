// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"fmt"
	"testing"

	"github.com/coreos/etcd/clientv3"
)

func TestCtlV3AuthEnable(t *testing.T)              { testCtl(t, authEnableTest) }
func TestCtlV3AuthDisable(t *testing.T)             { testCtl(t, authDisableTest) }
func TestCtlV3AuthWriteKey(t *testing.T)            { testCtl(t, authCredWriteKeyTest) }
func TestCtlV3AuthRoleUpdate(t *testing.T)          { testCtl(t, authRoleUpdateTest) }
func TestCtlV3AuthUserDeleteDuringOps(t *testing.T) { testCtl(t, authUserDeleteDuringOpsTest) }
func TestCtlV3AuthRoleRevokeDuringOps(t *testing.T) { testCtl(t, authRoleRevokeDuringOpsTest) }
func TestCtlV3AuthTxn(t *testing.T)                 { testCtl(t, authTestTxn) }
func TestCtlV3AuthPerfixPerm(t *testing.T)          { testCtl(t, authTestPrefixPerm) }
func TestCtlV3AuthMemberAdd(t *testing.T)           { testCtl(t, authTestMemberAdd) }
func TestCtlV3AuthMemberRemove(t *testing.T) {
	testCtl(t, authTestMemberRemove, withQuorum(), withNoStrictReconfig())
}
func TestCtlV3AuthMemberUpdate(t *testing.T)     { testCtl(t, authTestMemberUpdate) }
func TestCtlV3AuthCertCN(t *testing.T)           { testCtl(t, authTestCertCN, withCfg(configClientTLSCertAuth)) }
func TestCtlV3AuthRevokeWithDelete(t *testing.T) { testCtl(t, authTestRevokeWithDelete) }
func TestCtlV3AuthInvalidMgmt(t *testing.T)      { testCtl(t, authTestInvalidMgmt) }
func TestCtlV3AuthFromKeyPerm(t *testing.T)      { testCtl(t, authTestFromKeyPerm) }
func TestCtlV3AuthAndWatch(t *testing.T)         { testCtl(t, authTestWatch) }

func authEnableTest(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}
}

func authEnable(cx ctlCtx) error {
	// create root user with root role
	if err := ctlV3User(cx, []string{"add", "root", "--interactive=false"}, "User root created", []string{"root"}); err != nil {
		return fmt.Errorf("failed to create root user %v", err)
	}
	if err := ctlV3User(cx, []string{"grant-role", "root", "root"}, "Role root is granted to user root", nil); err != nil {
		return fmt.Errorf("failed to grant root user root role %v", err)
	}
	if err := ctlV3AuthEnable(cx); err != nil {
		return fmt.Errorf("authEnableTest ctlV3AuthEnable error (%v)", err)
	}
	return nil
}

func ctlV3AuthEnable(cx ctlCtx) error {
	cmdArgs := append(cx.PrefixArgs(), "auth", "enable")
	return spawnWithExpect(cmdArgs, "Authentication Enabled")
}

func authDisableTest(cx ctlCtx) {
	// a key that isn't granted to test-user
	if err := ctlV3Put(cx, "hoo", "a", ""); err != nil {
		cx.t.Fatal(err)
	}

	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// test-user doesn't have the permission, it must fail
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3PutFailPerm(cx, "hoo", "bar"); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	if err := ctlV3AuthDisable(cx); err != nil {
		cx.t.Fatalf("authDisableTest ctlV3AuthDisable error (%v)", err)
	}

	// now ErrAuthNotEnabled of Authenticate() is simply ignored
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "hoo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}

	// now the key can be accessed
	cx.user, cx.pass = "", ""
	if err := ctlV3Put(cx, "hoo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"hoo"}, []kv{{"hoo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}
}

func ctlV3AuthDisable(cx ctlCtx) error {
	cmdArgs := append(cx.PrefixArgs(), "auth", "disable")
	return spawnWithExpect(cmdArgs, "Authentication Disabled")
}

func authCredWriteKeyTest(cx ctlCtx) {
	// baseline key to check for failed puts
	if err := ctlV3Put(cx, "foo", "a", ""); err != nil {
		cx.t.Fatal(err)
	}

	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// confirm root role can access to all keys
	if err := ctlV3Put(cx, "foo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// try invalid user
	cx.user, cx.pass = "a", "b"
	if err := ctlV3PutFailAuth(cx, "foo", "bar"); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put failed
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// try good user
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "foo", "bar2", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar2"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// try bad password
	cx.user, cx.pass = "test-user", "badpass"
	if err := ctlV3PutFailAuth(cx, "foo", "baz"); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put failed
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar2"}}...); err != nil {
		cx.t.Fatal(err)
	}
}

func authRoleUpdateTest(cx ctlCtx) {
	if err := ctlV3Put(cx, "foo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}

	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// try put to not granted key
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3PutFailPerm(cx, "hoo", "bar"); err != nil {
		cx.t.Fatal(err)
	}

	// grant a new key
	cx.user, cx.pass = "root", "root"
	if err := ctlV3RoleGrantPermission(cx, "test-role", grantingPerm{true, true, "hoo", "", false}); err != nil {
		cx.t.Fatal(err)
	}

	// try a newly granted key
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "hoo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"hoo"}, []kv{{"hoo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// revoke the newly granted key
	cx.user, cx.pass = "root", "root"
	if err := ctlV3RoleRevokePermission(cx, "test-role", "hoo", "", false); err != nil {
		cx.t.Fatal(err)
	}

	// try put to the revoked key
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3PutFailPerm(cx, "hoo", "bar"); err != nil {
		cx.t.Fatal(err)
	}

	// confirm a key still granted can be accessed
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}
}

func authUserDeleteDuringOpsTest(cx ctlCtx) {
	if err := ctlV3Put(cx, "foo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}

	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// create a key
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "foo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// delete the user
	cx.user, cx.pass = "root", "root"
	err := ctlV3User(cx, []string{"delete", "test-user"}, "User test-user deleted", []string{})
	if err != nil {
		cx.t.Fatal(err)
	}

	// check the user is deleted
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3PutFailAuth(cx, "foo", "baz"); err != nil {
		cx.t.Fatal(err)
	}
}

func authRoleRevokeDuringOpsTest(cx ctlCtx) {
	if err := ctlV3Put(cx, "foo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}

	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// create a key
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "foo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"foo"}, []kv{{"foo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// create a new role
	cx.user, cx.pass = "root", "root"
	if err := ctlV3Role(cx, []string{"add", "test-role2"}, "Role test-role2 created"); err != nil {
		cx.t.Fatal(err)
	}
	// grant a new key to the new role
	if err := ctlV3RoleGrantPermission(cx, "test-role2", grantingPerm{true, true, "hoo", "", false}); err != nil {
		cx.t.Fatal(err)
	}
	// grant the new role to the user
	if err := ctlV3User(cx, []string{"grant-role", "test-user", "test-role2"}, "Role test-role2 is granted to user test-user", nil); err != nil {
		cx.t.Fatal(err)
	}

	// try a newly granted key
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "hoo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"hoo"}, []kv{{"hoo", "bar"}}...); err != nil {
		cx.t.Fatal(err)
	}

	// revoke a role from the user
	cx.user, cx.pass = "root", "root"
	err := ctlV3User(cx, []string{"revoke-role", "test-user", "test-role"}, "Role test-role is revoked from user test-user", []string{})
	if err != nil {
		cx.t.Fatal(err)
	}

	// check the role is revoked and permission is lost from the user
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3PutFailPerm(cx, "foo", "baz"); err != nil {
		cx.t.Fatal(err)
	}

	// try a key that can be accessed from the remaining role
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3Put(cx, "hoo", "bar2", ""); err != nil {
		cx.t.Fatal(err)
	}
	// confirm put succeeded
	if err := ctlV3Get(cx, []string{"hoo"}, []kv{{"hoo", "bar2"}}...); err != nil {
		cx.t.Fatal(err)
	}
}

func ctlV3PutFailAuth(cx ctlCtx, key, val string) error {
	return spawnWithExpect(append(cx.PrefixArgs(), "put", key, val), "authentication failed")
}

func ctlV3PutFailPerm(cx ctlCtx, key, val string) error {
	return spawnWithExpect(append(cx.PrefixArgs(), "put", key, val), "permission denied")
}

func authSetupTestUser(cx ctlCtx) {
	if err := ctlV3User(cx, []string{"add", "test-user", "--interactive=false"}, "User test-user created", []string{"pass"}); err != nil {
		cx.t.Fatal(err)
	}
	if err := spawnWithExpect(append(cx.PrefixArgs(), "role", "add", "test-role"), "Role test-role created"); err != nil {
		cx.t.Fatal(err)
	}
	if err := ctlV3User(cx, []string{"grant-role", "test-user", "test-role"}, "Role test-role is granted to user test-user", nil); err != nil {
		cx.t.Fatal(err)
	}
	cmd := append(cx.PrefixArgs(), "role", "grant-permission", "test-role", "readwrite", "foo")
	if err := spawnWithExpect(cmd, "Role test-role updated"); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestTxn(cx ctlCtx) {
	// keys with 1 suffix aren't granted to test-user
	// keys with 2 suffix are granted to test-user

	keys := []string{"c1", "s1", "f1"}
	grantedKeys := []string{"c2", "s2", "f2"}
	for _, key := range keys {
		if err := ctlV3Put(cx, key, "v", ""); err != nil {
			cx.t.Fatal(err)
		}
	}

	for _, key := range grantedKeys {
		if err := ctlV3Put(cx, key, "v", ""); err != nil {
			cx.t.Fatal(err)
		}
	}

	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// grant keys to test-user
	cx.user, cx.pass = "root", "root"
	for _, key := range grantedKeys {
		if err := ctlV3RoleGrantPermission(cx, "test-role", grantingPerm{true, true, key, "", false}); err != nil {
			cx.t.Fatal(err)
		}
	}

	// now test txn
	cx.interactive = true
	cx.user, cx.pass = "test-user", "pass"

	rqs := txnRequests{
		compare:  []string{`version("c2") = "1"`},
		ifSucess: []string{"get s2"},
		ifFail:   []string{"get f2"},
		results:  []string{"SUCCESS", "s2", "v"},
	}
	if err := ctlV3Txn(cx, rqs); err != nil {
		cx.t.Fatal(err)
	}

	// a key of compare case isn't granted
	rqs = txnRequests{
		compare:  []string{`version("c1") = "1"`},
		ifSucess: []string{"get s2"},
		ifFail:   []string{"get f2"},
		results:  []string{"Error:  etcdserver: permission denied"},
	}
	if err := ctlV3Txn(cx, rqs); err != nil {
		cx.t.Fatal(err)
	}

	// a key of success case isn't granted
	rqs = txnRequests{
		compare:  []string{`version("c2") = "1"`},
		ifSucess: []string{"get s1"},
		ifFail:   []string{"get f2"},
		results:  []string{"Error:  etcdserver: permission denied"},
	}
	if err := ctlV3Txn(cx, rqs); err != nil {
		cx.t.Fatal(err)
	}

	// a key of failure case isn't granted
	rqs = txnRequests{
		compare:  []string{`version("c2") = "1"`},
		ifSucess: []string{"get s2"},
		ifFail:   []string{"get f1"},
		results:  []string{"Error:  etcdserver: permission denied"},
	}
	if err := ctlV3Txn(cx, rqs); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestPrefixPerm(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	prefix := "/prefix/" // directory like prefix
	// grant keys to test-user
	cx.user, cx.pass = "root", "root"
	if err := ctlV3RoleGrantPermission(cx, "test-role", grantingPerm{true, true, prefix, "", true}); err != nil {
		cx.t.Fatal(err)
	}

	// try a prefix granted permission
	cx.user, cx.pass = "test-user", "pass"
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%s%d", prefix, i)
		if err := ctlV3Put(cx, key, "val", ""); err != nil {
			cx.t.Fatal(err)
		}
	}

	if err := ctlV3PutFailPerm(cx, clientv3.GetPrefixRangeEnd(prefix), "baz"); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestMemberAdd(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	peerURL := fmt.Sprintf("http://localhost:%d", etcdProcessBasePort+11)
	// ordinary user cannot add a new member
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3MemberAdd(cx, peerURL); err == nil {
		cx.t.Fatalf("ordinary user must not be allowed to add a member")
	}

	// root can add a new member
	cx.user, cx.pass = "root", "root"
	if err := ctlV3MemberAdd(cx, peerURL); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestMemberRemove(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	ep, memIDToRemove, clusterID := cx.memberToRemove()

	// ordinary user cannot remove a member
	cx.user, cx.pass = "test-user", "pass"
	if err := ctlV3MemberRemove(cx, ep, memIDToRemove, clusterID); err == nil {
		cx.t.Fatalf("ordinary user must not be allowed to remove a member")
	}

	// root can remove a member
	cx.user, cx.pass = "root", "root"
	if err := ctlV3MemberRemove(cx, ep, memIDToRemove, clusterID); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestMemberUpdate(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	mr, err := getMemberList(cx)
	if err != nil {
		cx.t.Fatal(err)
	}

	// ordinary user cannot update a member
	cx.user, cx.pass = "test-user", "pass"
	peerURL := fmt.Sprintf("http://localhost:%d", etcdProcessBasePort+11)
	memberID := fmt.Sprintf("%x", mr.Members[0].ID)
	if err = ctlV3MemberUpdate(cx, memberID, peerURL); err == nil {
		cx.t.Fatalf("ordinary user must not be allowed to update a member")
	}

	// root can update a member
	cx.user, cx.pass = "root", "root"
	if err = ctlV3MemberUpdate(cx, memberID, peerURL); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestCertCN(cx ctlCtx) {
	if err := ctlV3User(cx, []string{"add", "etcd", "--interactive=false"}, "User etcd created", []string{""}); err != nil {
		cx.t.Fatal(err)
	}
	if err := spawnWithExpect(append(cx.PrefixArgs(), "role", "add", "test-role"), "Role test-role created"); err != nil {
		cx.t.Fatal(err)
	}
	if err := ctlV3User(cx, []string{"grant-role", "etcd", "test-role"}, "Role test-role is granted to user etcd", nil); err != nil {
		cx.t.Fatal(err)
	}
	cmd := append(cx.PrefixArgs(), "role", "grant-permission", "test-role", "readwrite", "foo")
	if err := spawnWithExpect(cmd, "Role test-role updated"); err != nil {
		cx.t.Fatal(err)
	}

	// grant a new key
	if err := ctlV3RoleGrantPermission(cx, "test-role", grantingPerm{true, true, "hoo", "", false}); err != nil {
		cx.t.Fatal(err)
	}

	// try a granted key
	cx.user, cx.pass = "", ""
	if err := ctlV3Put(cx, "hoo", "bar", ""); err != nil {
		cx.t.Fatal(err)
	}

	// try a non granted key
	cx.user, cx.pass = "", ""
	if err := ctlV3PutFailPerm(cx, "baz", "bar"); err == nil {
		cx.t.Fatal(err)
	}
}

func authTestRevokeWithDelete(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// create a new role
	cx.user, cx.pass = "root", "root"
	if err := ctlV3Role(cx, []string{"add", "test-role2"}, "Role test-role2 created"); err != nil {
		cx.t.Fatal(err)
	}

	// grant the new role to the user
	if err := ctlV3User(cx, []string{"grant-role", "test-user", "test-role2"}, "Role test-role2 is granted to user test-user", nil); err != nil {
		cx.t.Fatal(err)
	}

	// check the result
	if err := ctlV3User(cx, []string{"get", "test-user"}, "Roles: test-role test-role2", nil); err != nil {
		cx.t.Fatal(err)
	}

	// delete the role, test-role2 must be revoked from test-user
	if err := ctlV3Role(cx, []string{"delete", "test-role2"}, "Role test-role2 deleted"); err != nil {
		cx.t.Fatal(err)
	}

	// check the result
	if err := ctlV3User(cx, []string{"get", "test-user"}, "Roles: test-role", nil); err != nil {
		cx.t.Fatal(err)
	}
}

func authTestInvalidMgmt(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	if err := ctlV3Role(cx, []string{"delete", "root"}, "Error:  etcdserver: invalid auth management"); err == nil {
		cx.t.Fatal("deleting the role root must not be allowed")
	}

	if err := ctlV3User(cx, []string{"revoke-role", "root", "root"}, "Error:  etcdserver: invalid auth management", []string{}); err == nil {
		cx.t.Fatal("revoking the role root from the user root must not be allowed")
	}
}

func authTestFromKeyPerm(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// grant keys after z to test-user
	cx.user, cx.pass = "root", "root"
	if err := ctlV3RoleGrantPermission(cx, "test-role", grantingPerm{true, true, "z", "\x00", false}); err != nil {
		cx.t.Fatal(err)
	}

	// try the granted open ended permission
	cx.user, cx.pass = "test-user", "pass"
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("z%d", i)
		if err := ctlV3Put(cx, key, "val", ""); err != nil {
			cx.t.Fatal(err)
		}
	}
	largeKey := ""
	for i := 0; i < 10; i++ {
		largeKey += "\xff"
		if err := ctlV3Put(cx, largeKey, "val", ""); err != nil {
			cx.t.Fatal(err)
		}
	}

	// try a non granted key
	if err := ctlV3PutFailPerm(cx, "x", "baz"); err != nil {
		cx.t.Fatal(err)
	}

	// revoke the open ended permission
	cx.user, cx.pass = "root", "root"
	if err := ctlV3RoleRevokePermission(cx, "test-role", "z", "", true); err != nil {
		cx.t.Fatal(err)
	}

	// try the revoked open ended permission
	cx.user, cx.pass = "test-user", "pass"
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("z%d", i)
		if err := ctlV3PutFailPerm(cx, key, "val"); err != nil {
			cx.t.Fatal(err)
		}
	}
}

func authTestWatch(cx ctlCtx) {
	if err := authEnable(cx); err != nil {
		cx.t.Fatal(err)
	}

	cx.user, cx.pass = "root", "root"
	authSetupTestUser(cx)

	// grant a key range
	if err := ctlV3RoleGrantPermission(cx, "test-role", grantingPerm{true, true, "key", "key4", false}); err != nil {
		cx.t.Fatal(err)
	}

	tests := []struct {
		puts []kv
		args []string

		wkv  []kv
		want bool
	}{
		{ // watch 1 key, should be successful
			[]kv{{"key", "value"}},
			[]string{"key", "--rev", "1"},
			[]kv{{"key", "value"}},
			true,
		},
		{ // watch 3 keys by range, should be successful
			[]kv{{"key1", "val1"}, {"key3", "val3"}, {"key2", "val2"}},
			[]string{"key", "key3", "--rev", "1"},
			[]kv{{"key1", "val1"}, {"key2", "val2"}},
			true,
		},

		{ // watch 1 key, should not be successful
			[]kv{},
			[]string{"key5", "--rev", "1"},
			[]kv{},
			false,
		},
		{ // watch 3 keys by range, should not be successful
			[]kv{},
			[]string{"key", "key6", "--rev", "1"},
			[]kv{},
			false,
		},
	}

	cx.user, cx.pass = "test-user", "pass"
	for i, tt := range tests {
		donec := make(chan struct{})
		go func(i int, puts []kv) {
			defer close(donec)
			for j := range puts {
				if err := ctlV3Put(cx, puts[j].key, puts[j].val, ""); err != nil {
					cx.t.Fatalf("watchTest #%d-%d: ctlV3Put error (%v)", i, j, err)
				}
			}
		}(i, tt.puts)

		var err error
		if tt.want {
			err = ctlV3Watch(cx, tt.args, tt.wkv...)
		} else {
			err = ctlV3WatchFailPerm(cx, tt.args)
		}

		if err != nil {
			if cx.dialTimeout > 0 && !isGRPCTimedout(err) {
				cx.t.Errorf("watchTest #%d: ctlV3Watch error (%v)", i, err)
			}
		}

		<-donec
	}

}
