// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"reflect"
	"testing"

	pb "google.golang.org/genproto/googleapis/datastore/v1"
)

func TestInterfaceToProtoNilKey(t *testing.T) {
	var iv *Key
	pv, err := interfaceToProto(iv, false)
	if err != nil {
		t.Fatalf("nil key: interfaceToProto: %v", err)
	}

	_, ok := pv.ValueType.(*pb.Value_NullValue)
	if !ok {
		t.Errorf("nil key: type:\ngot: %T\nwant: %T", pv.ValueType, &pb.Value_NullValue{})
	}
}

func TestSaveEntityNested(t *testing.T) {
	type WithKey struct {
		X string
		I int
		K *Key `datastore:"__key__"`
	}

	type NestedWithKey struct {
		Y string
		N WithKey
	}

	type WithoutKey struct {
		X string
		I int
	}

	type NestedWithoutKey struct {
		Y string
		N WithoutKey
	}

	type a struct {
		S string
	}

	type UnexpAnonym struct {
		a
	}

	testCases := []struct {
		desc string
		src  interface{}
		key  *Key
		want *pb.Entity
	}{
		{
			"nested entity with key",
			&NestedWithKey{
				Y: "yyy",
				N: WithKey{
					X: "two",
					I: 2,
					K: testKey1a,
				},
			},
			testKey0,
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"Y": {ValueType: &pb.Value_StringValue{"yyy"}},
					"N": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Key: keyToProto(testKey1a),
							Properties: map[string]*pb.Value{
								"X": {ValueType: &pb.Value_StringValue{"two"}},
								"I": {ValueType: &pb.Value_IntegerValue{2}},
							},
						},
					}},
				},
			},
		},
		{
			"nested entity with incomplete key",
			&NestedWithKey{
				Y: "yyy",
				N: WithKey{
					X: "two",
					I: 2,
					K: incompleteKey,
				},
			},
			testKey0,
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"Y": {ValueType: &pb.Value_StringValue{"yyy"}},
					"N": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Key: keyToProto(incompleteKey),
							Properties: map[string]*pb.Value{
								"X": {ValueType: &pb.Value_StringValue{"two"}},
								"I": {ValueType: &pb.Value_IntegerValue{2}},
							},
						},
					}},
				},
			},
		},
		{
			"nested entity without key",
			&NestedWithoutKey{
				Y: "yyy",
				N: WithoutKey{
					X: "two",
					I: 2,
				},
			},
			testKey0,
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"Y": {ValueType: &pb.Value_StringValue{"yyy"}},
					"N": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"X": {ValueType: &pb.Value_StringValue{"two"}},
								"I": {ValueType: &pb.Value_IntegerValue{2}},
							},
						},
					}},
				},
			},
		},
		{
			"key at top level",
			&WithKey{
				X: "three",
				I: 3,
				K: testKey0,
			},
			testKey0,
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"X": {ValueType: &pb.Value_StringValue{"three"}},
					"I": {ValueType: &pb.Value_IntegerValue{3}},
				},
			},
		},
		{
			"nested unexported anonymous struct field",
			&UnexpAnonym{
				a{S: "hello"},
			},
			testKey0,
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"S": {ValueType: &pb.Value_StringValue{"hello"}},
				},
			},
		},
	}

	for _, tc := range testCases {
		got, err := saveEntity(tc.key, tc.src)
		if err != nil {
			t.Errorf("saveEntity: %s: %v", tc.desc, err)
			continue
		}

		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("%s: compare:\ngot:  %#v\nwant: %#v", tc.desc, got, tc.want)
		}
	}
}
