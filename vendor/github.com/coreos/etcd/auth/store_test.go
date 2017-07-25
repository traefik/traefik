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

package auth

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/auth/authpb"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/backend"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func init() { BcryptCost = bcrypt.MinCost }

func dummyIndexWaiter(index uint64) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
	}()
	return ch
}

// TestNewAuthStoreRevision ensures newly auth store
// keeps the old revision when there are no changes.
func TestNewAuthStoreRevision(t *testing.T) {
	b, tPath := backend.NewDefaultTmpBackend()
	defer os.Remove(tPath)

	tp, err := NewTokenProvider("simple", dummyIndexWaiter)
	if err != nil {
		t.Fatal(err)
	}
	as := NewAuthStore(b, tp)
	err = enableAuthAndCreateRoot(as)
	if err != nil {
		t.Fatal(err)
	}
	old := as.Revision()
	b.Close()
	as.Close()

	// no changes to commit
	b2 := backend.NewDefaultBackend(tPath)
	as = NewAuthStore(b2, tp)
	new := as.Revision()
	b2.Close()
	as.Close()

	if old != new {
		t.Fatalf("expected revision %d, got %d", old, new)
	}
}

func setupAuthStore(t *testing.T) (store *authStore, teardownfunc func(t *testing.T)) {
	b, tPath := backend.NewDefaultTmpBackend()

	tp, err := NewTokenProvider("simple", dummyIndexWaiter)
	if err != nil {
		t.Fatal(err)
	}
	as := NewAuthStore(b, tp)
	err = enableAuthAndCreateRoot(as)
	if err != nil {
		t.Fatal(err)
	}

	// adds a new role
	_, err = as.RoleAdd(&pb.AuthRoleAddRequest{Name: "role-test"})
	if err != nil {
		t.Fatal(err)
	}

	ua := &pb.AuthUserAddRequest{Name: "foo", Password: "bar"}
	_, err = as.UserAdd(ua) // add a non-existing user
	if err != nil {
		t.Fatal(err)
	}

	tearDown := func(t *testing.T) {
		b.Close()
		os.Remove(tPath)
		as.Close()
	}
	return as, tearDown
}

func enableAuthAndCreateRoot(as *authStore) error {
	_, err := as.UserAdd(&pb.AuthUserAddRequest{Name: "root", Password: "root"})
	if err != nil {
		return err
	}

	_, err = as.RoleAdd(&pb.AuthRoleAddRequest{Name: "root"})
	if err != nil {
		return err
	}

	_, err = as.UserGrantRole(&pb.AuthUserGrantRoleRequest{User: "root", Role: "root"})
	if err != nil {
		return err
	}

	return as.AuthEnable()
}

func TestUserAdd(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	ua := &pb.AuthUserAddRequest{Name: "foo"}
	_, err := as.UserAdd(ua) // add an existing user
	if err == nil {
		t.Fatalf("expected %v, got %v", ErrUserAlreadyExist, err)
	}
	if err != ErrUserAlreadyExist {
		t.Fatalf("expected %v, got %v", ErrUserAlreadyExist, err)
	}

	ua = &pb.AuthUserAddRequest{Name: ""}
	_, err = as.UserAdd(ua) // add a user with empty name
	if err != ErrUserEmpty {
		t.Fatal(err)
	}
}

func TestCheckPassword(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	// auth a non-existing user
	_, err := as.CheckPassword("foo-test", "bar")
	if err == nil {
		t.Fatalf("expected %v, got %v", ErrAuthFailed, err)
	}
	if err != ErrAuthFailed {
		t.Fatalf("expected %v, got %v", ErrAuthFailed, err)
	}

	// auth an existing user with correct password
	_, err = as.CheckPassword("foo", "bar")
	if err != nil {
		t.Fatal(err)
	}

	// auth an existing user but with wrong password
	_, err = as.CheckPassword("foo", "")
	if err == nil {
		t.Fatalf("expected %v, got %v", ErrAuthFailed, err)
	}
	if err != ErrAuthFailed {
		t.Fatalf("expected %v, got %v", ErrAuthFailed, err)
	}
}

func TestUserDelete(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	// delete an existing user
	ud := &pb.AuthUserDeleteRequest{Name: "foo"}
	_, err := as.UserDelete(ud)
	if err != nil {
		t.Fatal(err)
	}

	// delete a non-existing user
	_, err = as.UserDelete(ud)
	if err == nil {
		t.Fatalf("expected %v, got %v", ErrUserNotFound, err)
	}
	if err != ErrUserNotFound {
		t.Fatalf("expected %v, got %v", ErrUserNotFound, err)
	}
}

func TestUserChangePassword(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	ctx1 := context.WithValue(context.WithValue(context.TODO(), "index", uint64(1)), "simpleToken", "dummy")
	_, err := as.Authenticate(ctx1, "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}

	_, err = as.UserChangePassword(&pb.AuthUserChangePasswordRequest{Name: "foo", Password: "baz"})
	if err != nil {
		t.Fatal(err)
	}

	ctx2 := context.WithValue(context.WithValue(context.TODO(), "index", uint64(2)), "simpleToken", "dummy")
	_, err = as.Authenticate(ctx2, "foo", "baz")
	if err != nil {
		t.Fatal(err)
	}

	// change a non-existing user
	_, err = as.UserChangePassword(&pb.AuthUserChangePasswordRequest{Name: "foo-test", Password: "bar"})
	if err == nil {
		t.Fatalf("expected %v, got %v", ErrUserNotFound, err)
	}
	if err != ErrUserNotFound {
		t.Fatalf("expected %v, got %v", ErrUserNotFound, err)
	}
}

func TestRoleAdd(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	// adds a new role
	_, err := as.RoleAdd(&pb.AuthRoleAddRequest{Name: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserGrant(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	// grants a role to the user
	_, err := as.UserGrantRole(&pb.AuthUserGrantRoleRequest{User: "foo", Role: "role-test"})
	if err != nil {
		t.Fatal(err)
	}

	// grants a role to a non-existing user
	_, err = as.UserGrantRole(&pb.AuthUserGrantRoleRequest{User: "foo-test", Role: "role-test"})
	if err == nil {
		t.Errorf("expected %v, got %v", ErrUserNotFound, err)
	}
	if err != ErrUserNotFound {
		t.Errorf("expected %v, got %v", ErrUserNotFound, err)
	}
}

func TestGetUser(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	_, err := as.UserGrantRole(&pb.AuthUserGrantRoleRequest{User: "foo", Role: "role-test"})
	if err != nil {
		t.Fatal(err)
	}

	u, err := as.UserGet(&pb.AuthUserGetRequest{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatal("expect user not nil, got nil")
	}
	expected := []string{"role-test"}
	if !reflect.DeepEqual(expected, u.Roles) {
		t.Errorf("expected %v, got %v", expected, u.Roles)
	}
}

func TestListUsers(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	ua := &pb.AuthUserAddRequest{Name: "user1", Password: "pwd1"}
	_, err := as.UserAdd(ua) // add a non-existing user
	if err != nil {
		t.Fatal(err)
	}

	ul, err := as.UserList(&pb.AuthUserListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(ul.Users, "root") {
		t.Errorf("expected %v in %v", "root", ul.Users)
	}
	if !contains(ul.Users, "user1") {
		t.Errorf("expected %v in %v", "user1", ul.Users)
	}
}

func TestRoleGrantPermission(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	_, err := as.RoleAdd(&pb.AuthRoleAddRequest{Name: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	perm := &authpb.Permission{
		PermType: authpb.WRITE,
		Key:      []byte("Keys"),
		RangeEnd: []byte("RangeEnd"),
	}
	_, err = as.RoleGrantPermission(&pb.AuthRoleGrantPermissionRequest{
		Name: "role-test-1",
		Perm: perm,
	})

	if err != nil {
		t.Error(err)
	}

	r, err := as.RoleGet(&pb.AuthRoleGetRequest{Role: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(perm, r.Perm[0]) {
		t.Errorf("expected %v, got %v", perm, r.Perm[0])
	}
}

func TestRoleRevokePermission(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	_, err := as.RoleAdd(&pb.AuthRoleAddRequest{Name: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	perm := &authpb.Permission{
		PermType: authpb.WRITE,
		Key:      []byte("Keys"),
		RangeEnd: []byte("RangeEnd"),
	}
	_, err = as.RoleGrantPermission(&pb.AuthRoleGrantPermissionRequest{
		Name: "role-test-1",
		Perm: perm,
	})

	if err != nil {
		t.Fatal(err)
	}

	_, err = as.RoleGet(&pb.AuthRoleGetRequest{Role: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = as.RoleRevokePermission(&pb.AuthRoleRevokePermissionRequest{
		Role:     "role-test-1",
		Key:      "Keys",
		RangeEnd: "RangeEnd",
	})
	if err != nil {
		t.Fatal(err)
	}

	var r *pb.AuthRoleGetResponse
	r, err = as.RoleGet(&pb.AuthRoleGetRequest{Role: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Perm) != 0 {
		t.Errorf("expected %v, got %v", 0, len(r.Perm))
	}
}

func TestUserRevokePermission(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	_, err := as.RoleAdd(&pb.AuthRoleAddRequest{Name: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = as.UserGrantRole(&pb.AuthUserGrantRoleRequest{User: "foo", Role: "role-test"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = as.UserGrantRole(&pb.AuthUserGrantRoleRequest{User: "foo", Role: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	u, err := as.UserGet(&pb.AuthUserGetRequest{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"role-test", "role-test-1"}
	if !reflect.DeepEqual(expected, u.Roles) {
		t.Fatalf("expected %v, got %v", expected, u.Roles)
	}

	_, err = as.UserRevokeRole(&pb.AuthUserRevokeRoleRequest{Name: "foo", Role: "role-test-1"})
	if err != nil {
		t.Fatal(err)
	}

	u, err = as.UserGet(&pb.AuthUserGetRequest{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}

	expected = []string{"role-test"}
	if !reflect.DeepEqual(expected, u.Roles) {
		t.Errorf("expected %v, got %v", expected, u.Roles)
	}
}

func TestRoleDelete(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	_, err := as.RoleDelete(&pb.AuthRoleDeleteRequest{Role: "role-test"})
	if err != nil {
		t.Fatal(err)
	}
	rl, err := as.RoleList(&pb.AuthRoleListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"root"}
	if !reflect.DeepEqual(expected, rl.Roles) {
		t.Errorf("expected %v, got %v", expected, rl.Roles)
	}
}

func TestAuthInfoFromCtx(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	ctx := context.Background()
	ai, err := as.AuthInfoFromCtx(ctx)
	if err != nil && ai != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", ai, err)
	}

	ctx = metadata.NewContext(context.Background(), metadata.New(map[string]string{"tokens": "dummy"}))
	ai, err = as.AuthInfoFromCtx(ctx)
	if err != nil && ai != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", ai, err)
	}

	ctx = context.WithValue(context.WithValue(context.TODO(), "index", uint64(1)), "simpleToken", "dummy")
	resp, err := as.Authenticate(ctx, "foo", "bar")
	if err != nil {
		t.Error(err)
	}

	ctx = metadata.NewContext(context.Background(), metadata.New(map[string]string{"token": "Invalid Token"}))
	_, err = as.AuthInfoFromCtx(ctx)
	if err != ErrInvalidAuthToken {
		t.Errorf("expected %v, got %v", ErrInvalidAuthToken, err)
	}

	ctx = metadata.NewContext(context.Background(), metadata.New(map[string]string{"token": "Invalid.Token"}))
	_, err = as.AuthInfoFromCtx(ctx)
	if err != ErrInvalidAuthToken {
		t.Errorf("expected %v, got %v", ErrInvalidAuthToken, err)
	}

	ctx = metadata.NewContext(context.Background(), metadata.New(map[string]string{"token": resp.Token}))
	ai, err = as.AuthInfoFromCtx(ctx)
	if err != nil {
		t.Error(err)
	}
	if ai.Username != "foo" {
		t.Errorf("expected %v, got %v", "foo", ai.Username)
	}
}

func TestAuthDisable(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	as.AuthDisable()
	ctx := context.WithValue(context.WithValue(context.TODO(), "index", uint64(2)), "simpleToken", "dummy")
	_, err := as.Authenticate(ctx, "foo", "bar")
	if err != ErrAuthNotEnabled {
		t.Errorf("expected %v, got %v", ErrAuthNotEnabled, err)
	}

	// Disabling disabled auth to make sure it can return safely if store is already disabled.
	as.AuthDisable()
	_, err = as.Authenticate(ctx, "foo", "bar")
	if err != ErrAuthNotEnabled {
		t.Errorf("expected %v, got %v", ErrAuthNotEnabled, err)
	}
}

// TestAuthRevisionRace ensures that access to authStore.revision is thread-safe.
func TestAuthInfoFromCtxRace(t *testing.T) {
	b, tPath := backend.NewDefaultTmpBackend()
	defer os.Remove(tPath)

	tp, err := NewTokenProvider("simple", dummyIndexWaiter)
	if err != nil {
		t.Fatal(err)
	}
	as := NewAuthStore(b, tp)
	defer as.Close()

	donec := make(chan struct{})
	go func() {
		defer close(donec)
		ctx := metadata.NewContext(context.Background(), metadata.New(map[string]string{"token": "test"}))
		as.AuthInfoFromCtx(ctx)
	}()
	as.UserAdd(&pb.AuthUserAddRequest{Name: "test"})
	<-donec
}

func TestIsAdminPermitted(t *testing.T) {
	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	err := as.IsAdminPermitted(&AuthInfo{Username: "root", Revision: 1})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	// invalid user
	err = as.IsAdminPermitted(&AuthInfo{Username: "rooti", Revision: 1})
	if err != ErrUserNotFound {
		t.Errorf("expected %v, got %v", ErrUserNotFound, err)
	}

	// non-admin user
	err = as.IsAdminPermitted(&AuthInfo{Username: "foo", Revision: 1})
	if err != ErrPermissionDenied {
		t.Errorf("expected %v, got %v", ErrPermissionDenied, err)
	}

	// disabled auth should return nil
	as.AuthDisable()
	err = as.IsAdminPermitted(&AuthInfo{Username: "root", Revision: 1})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRecoverFromSnapshot(t *testing.T) {
	as, _ := setupAuthStore(t)

	ua := &pb.AuthUserAddRequest{Name: "foo"}
	_, err := as.UserAdd(ua) // add an existing user
	if err == nil {
		t.Fatalf("expected %v, got %v", ErrUserAlreadyExist, err)
	}
	if err != ErrUserAlreadyExist {
		t.Fatalf("expected %v, got %v", ErrUserAlreadyExist, err)
	}

	ua = &pb.AuthUserAddRequest{Name: ""}
	_, err = as.UserAdd(ua) // add a user with empty name
	if err != ErrUserEmpty {
		t.Fatal(err)
	}

	as.Close()

	tp, err := NewTokenProvider("simple", dummyIndexWaiter)
	if err != nil {
		t.Fatal(err)
	}
	as2 := NewAuthStore(as.be, tp)
	defer func(a *authStore) {
		a.Close()
	}(as2)

	if !as2.isAuthEnabled() {
		t.Fatal("recovering authStore from existing backend failed")
	}

	ul, err := as.UserList(&pb.AuthUserListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(ul.Users, "root") {
		t.Errorf("expected %v in %v", "root", ul.Users)
	}
}

func contains(array []string, str string) bool {
	for _, s := range array {
		if s == str {
			return true
		}
	}
	return false
}

func TestHammerSimpleAuthenticate(t *testing.T) {
	// set TTL values low to try to trigger races
	oldTTL, oldTTLRes := simpleTokenTTL, simpleTokenTTLResolution
	defer func() {
		simpleTokenTTL = oldTTL
		simpleTokenTTLResolution = oldTTLRes
	}()
	simpleTokenTTL = 10 * time.Millisecond
	simpleTokenTTLResolution = simpleTokenTTL
	users := make(map[string]struct{})

	as, tearDown := setupAuthStore(t)
	defer tearDown(t)

	// create lots of users
	for i := 0; i < 50; i++ {
		u := fmt.Sprintf("user-%d", i)
		ua := &pb.AuthUserAddRequest{Name: u, Password: "123"}
		if _, err := as.UserAdd(ua); err != nil {
			t.Fatal(err)
		}
		users[u] = struct{}{}
	}

	// hammer on authenticate with lots of users
	for i := 0; i < 10; i++ {
		var wg sync.WaitGroup
		wg.Add(len(users))
		for u := range users {
			go func(user string) {
				defer wg.Done()
				token := fmt.Sprintf("%s(%d)", user, i)
				ctx := context.WithValue(context.WithValue(context.TODO(), "index", uint64(1)), "simpleToken", token)
				if _, err := as.Authenticate(ctx, user, "123"); err != nil {
					t.Fatal(err)
				}
				if _, err := as.AuthInfoFromCtx(ctx); err != nil {
					t.Fatal(err)
				}
			}(u)
		}
		time.Sleep(time.Millisecond)
		wg.Wait()
	}
}
