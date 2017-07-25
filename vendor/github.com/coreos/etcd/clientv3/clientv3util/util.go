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

// Package clientv3util contains utility functions derived from clientv3.
package clientv3util

import (
	"github.com/coreos/etcd/clientv3"
)

// KeyExists returns a comparison operation that evaluates to true iff the given
// key exists. It does this by checking if the key `Version` is greater than 0.
// It is a useful guard in transaction delete operations.
func KeyExists(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.Version(key), ">", 0)
}

// KeyMissing returns a comparison operation that evaluates to true iff the
// given key does not exist.
func KeyMissing(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.Version(key), "=", 0)
}
