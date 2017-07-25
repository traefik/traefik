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
	"reflect"
	"testing"

	proto3 "github.com/golang/protobuf/ptypes/struct"
	sppb "google.golang.org/genproto/googleapis/spanner/v1"
)

// Test KeySet.proto().
func TestKeySetToProto(t *testing.T) {
	for _, test := range []struct {
		ks        KeySet
		wantProto *sppb.KeySet
	}{
		{
			KeySet{},
			&sppb.KeySet{
				Keys:   []*proto3.ListValue{},
				Ranges: []*sppb.KeyRange{},
			},
		},
		{
			KeySet{All: true},
			&sppb.KeySet{
				All:    true,
				Keys:   []*proto3.ListValue{},
				Ranges: []*sppb.KeyRange{},
			},
		},
		{
			KeySet{Keys: []Key{{1, 2}, {3, 4}}},
			&sppb.KeySet{
				Keys:   []*proto3.ListValue{listValueProto(intProto(1), intProto(2)), listValueProto(intProto(3), intProto(4))},
				Ranges: []*sppb.KeyRange{},
			},
		},
		{
			KeySet{Ranges: []KeyRange{{Key{1}, Key{2}, ClosedClosed}, {Key{3}, Key{10}, OpenClosed}}},
			&sppb.KeySet{
				Keys: []*proto3.ListValue{},
				Ranges: []*sppb.KeyRange{
					&sppb.KeyRange{
						&sppb.KeyRange_StartClosed{listValueProto(intProto(1))},
						&sppb.KeyRange_EndClosed{listValueProto(intProto(2))},
					},
					&sppb.KeyRange{
						&sppb.KeyRange_StartOpen{listValueProto(intProto(3))},
						&sppb.KeyRange_EndClosed{listValueProto(intProto(10))},
					},
				},
			},
		},
	} {
		gotProto, err := test.ks.proto()
		if err != nil {
			t.Errorf("%v.proto() returns error %v; want nil error", test.ks, err)
		}
		if !reflect.DeepEqual(gotProto, test.wantProto) {
			t.Errorf("%v.proto() = \n%v\nwant:\n%v", test.ks, gotProto.String(), test.wantProto.String())
		}
	}
}

// Test helpers that help to create KeySets.
func TestKeySetHelpers(t *testing.T) {
	// Test Keys with one key.
	k := Key{[]byte{1, 2, 3}}
	if got, want := Keys(k), (KeySet{Keys: []Key{k}}); !reflect.DeepEqual(got, want) {
		t.Errorf("Keys(%q) = %q, want %q", k, got, want)
	}
	// Test Keys with multiple keys.
	ks := []Key{Key{57}, Key{NullString{"value", false}}}
	if got, want := Keys(ks...), (KeySet{Keys: ks}); !reflect.DeepEqual(got, want) {
		t.Errorf("Keys(%v) = %v, want %v", ks, got, want)
	}
	// Test Range.
	kr := KeyRange{Key{1}, Key{10}, ClosedClosed}
	if got, want := Range(kr), (KeySet{Ranges: []KeyRange{kr}}); !reflect.DeepEqual(got, want) {
		t.Errorf("Range(%v) = %v, want %v", kr, got, want)
	}
	// Test PrefixRange.
	k = Key{2}
	kr = KeyRange{k, k, ClosedClosed}
	if got, want := PrefixRange(k), (KeySet{Ranges: []KeyRange{kr}}); !reflect.DeepEqual(got, want) {
		t.Errorf("PrefixRange(%v) = %v, want %v", k, got, want)
	}
	// Test UnionKeySets.
	sk1, sk2 := Keys(Key{2}), Keys(Key{3})
	r1, r2 := Range(KeyRange{Key{1}, Key{10}, ClosedClosed}), Range(KeyRange{Key{15}, Key{20}, OpenClosed})
	want := KeySet{
		Keys:   []Key{Key{2}, Key{3}},
		Ranges: []KeyRange{KeyRange{Key{1}, Key{10}, ClosedClosed}, KeyRange{Key{15}, Key{20}, OpenClosed}},
	}
	if got := UnionKeySets(sk1, sk2, r1, r2); !reflect.DeepEqual(got, want) {
		t.Errorf("UnionKeySets(%v, %v, %v, %v) = %v, want %v", sk1, sk2, r1, r2, got, want)
	}
	all := AllKeys()
	if got := UnionKeySets(sk1, sk2, r1, r2, all); !reflect.DeepEqual(got, all) {
		t.Errorf("UnionKeySets(%v, %v, %v, %v, %v) = %v, want %v", sk1, sk2, r1, r2, all, got, all)
	}
}
