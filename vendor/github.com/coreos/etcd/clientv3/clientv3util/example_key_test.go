// Copyright 2017 The etcd Authors
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

package clientv3util_test

import (
	"context"
	"log"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
)

func ExampleKeyExists_put() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	kvc := clientv3.NewKV(cli)

	// perform a put only if key is missing
	// It is useful to do the check (transactionally) to avoid overwriting
	// the existing key which would generate potentially unwanted events,
	// unless of course you wanted to do an overwrite no matter what.
	_, err = kvc.Txn(context.Background()).
		If(clientv3util.KeyMissing("purpleidea")).
		Then(clientv3.OpPut("purpleidea", "hello world")).
		Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleKeyExists_delete() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	kvc := clientv3.NewKV(cli)

	// perform a delete only if key already exists
	_, err = kvc.Txn(context.Background()).
		If(clientv3util.KeyExists("purpleidea")).
		Then(clientv3.OpDelete("purpleidea")).
		Commit()
	if err != nil {
		log.Fatal(err)
	}
}
