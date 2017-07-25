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

package integration

import (
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/integration"
	"github.com/coreos/etcd/pkg/testutil"
	"golang.org/x/net/context"
)

func TestUserError(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	authapi := clus.RandClient()

	_, err := authapi.UserAdd(context.TODO(), "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}

	_, err = authapi.UserAdd(context.TODO(), "foo", "bar")
	if err != rpctypes.ErrUserAlreadyExist {
		t.Fatalf("expected %v, got %v", rpctypes.ErrUserAlreadyExist, err)
	}

	_, err = authapi.UserDelete(context.TODO(), "not-exist-user")
	if err != rpctypes.ErrUserNotFound {
		t.Fatalf("expected %v, got %v", rpctypes.ErrUserNotFound, err)
	}

	_, err = authapi.UserGrantRole(context.TODO(), "foo", "test-role-does-not-exist")
	if err != rpctypes.ErrRoleNotFound {
		t.Fatalf("expected %v, got %v", rpctypes.ErrRoleNotFound, err)
	}
}

func TestUserErrorAuth(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	authapi := clus.RandClient()
	authSetupRoot(t, authapi.Auth)

	// un-authenticated client
	if _, err := authapi.UserAdd(context.TODO(), "foo", "bar"); err != rpctypes.ErrUserNotFound {
		t.Fatalf("expected %v, got %v", rpctypes.ErrUserNotFound, err)
	}

	// wrong id or password
	cfg := clientv3.Config{Endpoints: authapi.Endpoints()}
	cfg.Username, cfg.Password = "wrong-id", "123"
	if _, err := clientv3.New(cfg); err != rpctypes.ErrAuthFailed {
		t.Fatalf("expected %v, got %v", rpctypes.ErrAuthFailed, err)
	}
	cfg.Username, cfg.Password = "root", "wrong-pass"
	if _, err := clientv3.New(cfg); err != rpctypes.ErrAuthFailed {
		t.Fatalf("expected %v, got %v", rpctypes.ErrAuthFailed, err)
	}

	cfg.Username, cfg.Password = "root", "123"
	authed, err := clientv3.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer authed.Close()

	if _, err := authed.UserList(context.TODO()); err != nil {
		t.Fatal(err)
	}
}

func authSetupRoot(t *testing.T, auth clientv3.Auth) {
	if _, err := auth.UserAdd(context.TODO(), "root", "123"); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.RoleAdd(context.TODO(), "root"); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.UserGrantRole(context.TODO(), "root", "root"); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.AuthEnable(context.TODO()); err != nil {
		t.Fatal(err)
	}
}
