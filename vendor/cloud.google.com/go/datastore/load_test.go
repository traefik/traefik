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

type Simple struct {
	I int64
}

type SimpleWithTag struct {
	I int64 `datastore:"II"`
}

type NestedSimpleWithTag struct {
	A SimpleWithTag `datastore:"AA"`
}

type NestedSliceOfSimple struct {
	A []Simple
}

type SimpleTwoFields struct {
	S  string
	SS string
}

type NestedSimpleAnonymous struct {
	Simple
	X string
}

type NestedSimple struct {
	A Simple
	I int
}

type NestedSimple1 struct {
	A Simple
	X string
}

type NestedSimple2X struct {
	AA NestedSimple
	A  SimpleTwoFields
	S  string
}

type BDotB struct {
	B string `datastore:"B.B"`
}

type ABDotB struct {
	A BDotB
}

type MultiAnonymous struct {
	Simple
	SimpleTwoFields
	X string
}

func TestLoadEntityNestedLegacy(t *testing.T) {
	testCases := []struct {
		desc string
		src  *pb.Entity
		want interface{}
	}{
		{
			"nested",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"X":   {ValueType: &pb.Value_StringValue{"two"}},
					"A.I": {ValueType: &pb.Value_IntegerValue{2}},
				},
			},
			&NestedSimple1{
				A: Simple{I: 2},
				X: "two",
			},
		},
		{
			"nested with tag",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"AA.II": {ValueType: &pb.Value_IntegerValue{2}},
				},
			},
			&NestedSimpleWithTag{
				A: SimpleWithTag{I: 2},
			},
		},
		{
			"nested with anonymous struct field",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"X": {ValueType: &pb.Value_StringValue{"two"}},
					"I": {ValueType: &pb.Value_IntegerValue{2}},
				},
			},
			&NestedSimpleAnonymous{
				Simple: Simple{I: 2},
				X:      "two",
			},
		},
		{
			"nested with dotted field tag",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"A.B.B": {ValueType: &pb.Value_StringValue{"bb"}},
				},
			},
			&ABDotB{
				A: BDotB{
					B: "bb",
				},
			},
		},
		{
			"nested with multiple anonymous fields",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"I":  {ValueType: &pb.Value_IntegerValue{3}},
					"S":  {ValueType: &pb.Value_StringValue{"S"}},
					"SS": {ValueType: &pb.Value_StringValue{"s"}},
					"X":  {ValueType: &pb.Value_StringValue{"s"}},
				},
			},
			&MultiAnonymous{
				Simple:          Simple{I: 3},
				SimpleTwoFields: SimpleTwoFields{S: "S", SS: "s"},
				X:               "s",
			},
		},
	}

	for _, tc := range testCases {
		dst := reflect.New(reflect.TypeOf(tc.want).Elem()).Interface()
		err := loadEntityProto(dst, tc.src)
		if err != nil {
			t.Errorf("loadEntityProto: %s: %v", tc.desc, err)
			continue
		}

		if !reflect.DeepEqual(tc.want, dst) {
			t.Errorf("%s: compare:\ngot:  %#v\nwant: %#v", tc.desc, dst, tc.want)
		}
	}
}

type WithKey struct {
	X string
	I int
	K *Key `datastore:"__key__"`
}

type NestedWithKey struct {
	Y string
	N WithKey
}

var (
	incompleteKey = newKey("", nil)
	invalidKey    = newKey("s", incompleteKey)
)

func TestLoadEntityNested(t *testing.T) {
	testCases := []struct {
		desc string
		src  *pb.Entity
		want interface{}
	}{
		{
			"nested basic",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"A": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"I": {ValueType: &pb.Value_IntegerValue{3}},
							},
						},
					}},
					"I": {ValueType: &pb.Value_IntegerValue{10}},
				},
			},
			&NestedSimple{
				A: Simple{I: 3},
				I: 10,
			},
		},
		{
			"nested with struct tags",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"AA": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"II": {ValueType: &pb.Value_IntegerValue{1}},
							},
						},
					}},
				},
			},
			&NestedSimpleWithTag{
				A: SimpleWithTag{I: 1},
			},
		},
		{
			"nested 2x",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"AA": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"A": {ValueType: &pb.Value_EntityValue{
									&pb.Entity{
										Properties: map[string]*pb.Value{
											"I": {ValueType: &pb.Value_IntegerValue{3}},
										},
									},
								}},
								"I": {ValueType: &pb.Value_IntegerValue{1}},
							},
						},
					}},
					"A": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"S":  {ValueType: &pb.Value_StringValue{"S"}},
								"SS": {ValueType: &pb.Value_StringValue{"s"}},
							},
						},
					}},
					"S": {ValueType: &pb.Value_StringValue{"SS"}},
				},
			},
			&NestedSimple2X{
				AA: NestedSimple{
					A: Simple{I: 3},
					I: 1,
				},
				A: SimpleTwoFields{S: "S", SS: "s"},
				S: "SS",
			},
		},
		{
			"nested anonymous",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"I": {ValueType: &pb.Value_IntegerValue{3}},
					"X": {ValueType: &pb.Value_StringValue{"SomeX"}},
				},
			},
			&NestedSimpleAnonymous{
				Simple: Simple{I: 3},
				X:      "SomeX",
			},
		},
		{
			"nested simple with slice",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"A": {ValueType: &pb.Value_ArrayValue{
						&pb.ArrayValue{
							[]*pb.Value{
								{ValueType: &pb.Value_EntityValue{
									&pb.Entity{
										Properties: map[string]*pb.Value{
											"I": {ValueType: &pb.Value_IntegerValue{3}},
										},
									},
								}},
								{ValueType: &pb.Value_EntityValue{
									&pb.Entity{
										Properties: map[string]*pb.Value{
											"I": {ValueType: &pb.Value_IntegerValue{4}},
										},
									},
								}},
							},
						},
					}},
				},
			},

			&NestedSliceOfSimple{
				A: []Simple{Simple{I: 3}, Simple{I: 4}},
			},
		},
		{
			"nested with multiple anonymous fields",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"I":  {ValueType: &pb.Value_IntegerValue{3}},
					"S":  {ValueType: &pb.Value_StringValue{"S"}},
					"SS": {ValueType: &pb.Value_StringValue{"s"}},
					"X":  {ValueType: &pb.Value_StringValue{"ss"}},
				},
			},
			&MultiAnonymous{
				Simple:          Simple{I: 3},
				SimpleTwoFields: SimpleTwoFields{S: "S", SS: "s"},
				X:               "ss",
			},
		},
		{
			"nested with dotted field tag",
			&pb.Entity{
				Properties: map[string]*pb.Value{
					"A": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"B.B": {ValueType: &pb.Value_StringValue{"bb"}},
							},
						},
					}},
				},
			},
			&ABDotB{
				A: BDotB{
					B: "bb",
				},
			},
		},
		{
			"nested entity with key",
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
			&NestedWithKey{
				Y: "yyy",
				N: WithKey{
					X: "two",
					I: 2,
					K: testKey1a,
				},
			},
		},
		{
			"nested entity with invalid key",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"Y": {ValueType: &pb.Value_StringValue{"yyy"}},
					"N": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Key: keyToProto(invalidKey),
							Properties: map[string]*pb.Value{
								"X": {ValueType: &pb.Value_StringValue{"two"}},
								"I": {ValueType: &pb.Value_IntegerValue{2}},
							},
						},
					}},
				},
			},
			&NestedWithKey{
				Y: "yyy",
				N: WithKey{
					X: "two",
					I: 2,
					K: invalidKey,
				},
			},
		},
	}

	for _, tc := range testCases {
		dst := reflect.New(reflect.TypeOf(tc.want).Elem()).Interface()
		err := loadEntityProto(dst, tc.src)
		if err != nil {
			t.Errorf("loadEntityProto: %s: %v", tc.desc, err)
			continue
		}

		if !reflect.DeepEqual(tc.want, dst) {
			t.Errorf("%s: compare:\ngot:  %#v\nwant: %#v", tc.desc, dst, tc.want)
		}
	}
}

type NestedStructPtrs struct {
	*SimpleTwoFields
	Nest      *SimpleTwoFields
	TwiceNest *NestedSimple2
	I         int
}

type NestedSimple2 struct {
	A *Simple
	I int
}

func TestAlreadyPopulatedDst(t *testing.T) {
	testCases := []struct {
		desc string
		src  *pb.Entity
		dst  interface{}
		want interface{}
	}{
		{
			"simple already populated, nil properties",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"I": {ValueType: &pb.Value_NullValue{}},
				},
			},
			&Simple{
				I: 12,
			},
			&Simple{},
		},
		{
			"nested structs already populated",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"SS": {ValueType: &pb.Value_StringValue{"world"}},
				},
			},
			&SimpleTwoFields{S: "hello" /* SS: "" */},
			&SimpleTwoFields{S: "hello", SS: "world"},
		},
		{
			"nested structs already populated, pValues nil",
			&pb.Entity{
				Key: keyToProto(testKey0),
				Properties: map[string]*pb.Value{
					"S":    {ValueType: &pb.Value_NullValue{}},
					"SS":   {ValueType: &pb.Value_StringValue{"ss hello"}},
					"Nest": {ValueType: &pb.Value_NullValue{}},
					"TwiceNest": {ValueType: &pb.Value_EntityValue{
						&pb.Entity{
							Properties: map[string]*pb.Value{
								"A": {ValueType: &pb.Value_NullValue{}},
								"I": {ValueType: &pb.Value_IntegerValue{2}},
							},
						},
					}},
					"I": {ValueType: &pb.Value_IntegerValue{5}},
				},
			},
			&NestedStructPtrs{
				&SimpleTwoFields{S: "hello" /* SS: "" */},
				&SimpleTwoFields{ /* S: "" */ SS: "twice hello"},
				&NestedSimple2{
					A: &Simple{I: 2},
					/* I: 0 */
				},
				0,
			},
			&NestedStructPtrs{
				&SimpleTwoFields{ /* S: "" */ SS: "ss hello"},
				nil,
				&NestedSimple2{
					/* A: nil, */
					I: 2,
				},
				5,
			},
		},
	}

	for _, tc := range testCases {
		err := loadEntityProto(tc.dst, tc.src)
		if err != nil {
			t.Errorf("loadEntityProto: %s: %v", tc.desc, err)
			continue
		}

		if !reflect.DeepEqual(tc.want, tc.dst) {
			t.Errorf("%s: compare:\ngot:  %#v\nwant: %#v", tc.desc, tc.dst, tc.want)
		}
	}
}
