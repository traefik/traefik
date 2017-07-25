/*
Copyright 2017 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner

import (
	proto3 "github.com/golang/protobuf/ptypes/struct"
	sppb "google.golang.org/genproto/googleapis/spanner/v1"
)

// A KeySet defines a collection of Cloud Spanner keys and/or key
// ranges. All the keys are expected to be in the same table or index. The keys
// need not be sorted in any particular way.
//
// If the same key is specified multiple times in the set (for example if two
// ranges, two keys, or a key and a range overlap), the Cloud Spanner backend behaves
// as if the key were only specified once.
type KeySet struct {
	// If All == true, then the KeySet names all rows of a table or
	// under a index.
	All bool
	// Keys is a list of keys covered by KeySet, see also documentation of
	// Key for details.
	Keys []Key
	// Ranges is a list of key ranges covered by KeySet, see also documentation of
	// KeyRange for details.
	Ranges []KeyRange
}

// AllKeys returns a KeySet that represents all Keys of a table or a index.
func AllKeys() KeySet {
	return KeySet{All: true}
}

// Keys returns a KeySet for a set of keys.
func Keys(keys ...Key) KeySet {
	ks := KeySet{Keys: make([]Key, len(keys))}
	copy(ks.Keys, keys)
	return ks
}

// Range returns a KeySet for a range of keys.
func Range(r KeyRange) KeySet {
	return KeySet{Ranges: []KeyRange{r}}
}

// PrefixRange returns a KeySet for all keys with the given prefix, which is
// a key itself.
func PrefixRange(prefix Key) KeySet {
	return KeySet{Ranges: []KeyRange{
		{
			Start: prefix,
			End:   prefix,
			Kind:  ClosedClosed,
		},
	}}
}

// UnionKeySets unions multiple KeySets into a superset.
func UnionKeySets(keySets ...KeySet) KeySet {
	s := KeySet{}
	for _, ks := range keySets {
		if ks.All {
			return KeySet{All: true}
		}
		s.Keys = append(s.Keys, ks.Keys...)
		s.Ranges = append(s.Ranges, ks.Ranges...)
	}
	return s
}

// proto converts KeySet into sppb.KeySet, which is the protobuf
// representation of KeySet.
func (keys KeySet) proto() (*sppb.KeySet, error) {
	pb := &sppb.KeySet{
		Keys:   make([]*proto3.ListValue, 0, len(keys.Keys)),
		Ranges: make([]*sppb.KeyRange, 0, len(keys.Ranges)),
		All:    keys.All,
	}
	for _, key := range keys.Keys {
		keyProto, err := key.proto()
		if err != nil {
			return nil, err
		}
		pb.Keys = append(pb.Keys, keyProto)
	}
	for _, r := range keys.Ranges {
		rProto, err := r.proto()
		if err != nil {
			return nil, err
		}
		pb.Ranges = append(pb.Ranges, rProto)
	}
	return pb, nil
}
