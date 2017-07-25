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
	"sort"
	"strings"
	"testing"

	proto3 "github.com/golang/protobuf/ptypes/struct"

	sppb "google.golang.org/genproto/googleapis/spanner/v1"
)

// keysetProto returns protobuf encoding of valid spanner.KeySet.
func keysetProto(t *testing.T, ks KeySet) *sppb.KeySet {
	k, err := ks.proto()
	if err != nil {
		t.Fatalf("cannot convert keyset %v to protobuf: %v", ks, err)
	}
	return k
}

// Test encoding from spanner.Mutation to protobuf.
func TestMutationToProto(t *testing.T) {
	for i, test := range []struct {
		m    *Mutation
		want *sppb.Mutation
	}{
		// Delete Mutation
		{
			&Mutation{opDelete, "t_foo", Keys(Key{"foo"}), nil, nil},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Delete_{
					Delete: &sppb.Mutation_Delete{
						Table:  "t_foo",
						KeySet: keysetProto(t, Keys(Key{"foo"})),
					},
				},
			},
		},
		// Insert Mutation
		{
			&Mutation{opInsert, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{int64(1), int64(2)}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Insert{
					Insert: &sppb.Mutation_Write{
						Table:   "t_foo",
						Columns: []string{"col1", "col2"},
						Values: []*proto3.ListValue{
							&proto3.ListValue{
								Values: []*proto3.Value{intProto(1), intProto(2)},
							},
						},
					},
				},
			},
		},
		// InsertOrUpdate Mutation
		{
			&Mutation{opInsertOrUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{1.0, 2.0}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_InsertOrUpdate{
					InsertOrUpdate: &sppb.Mutation_Write{
						Table:   "t_foo",
						Columns: []string{"col1", "col2"},
						Values: []*proto3.ListValue{
							&proto3.ListValue{
								Values: []*proto3.Value{floatProto(1.0), floatProto(2.0)},
							},
						},
					},
				},
			},
		},
		// Replace Mutation
		{
			&Mutation{opReplace, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", 2.0}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Replace{
					Replace: &sppb.Mutation_Write{
						Table:   "t_foo",
						Columns: []string{"col1", "col2"},
						Values: []*proto3.ListValue{
							&proto3.ListValue{
								Values: []*proto3.Value{stringProto("one"), floatProto(2.0)},
							},
						},
					},
				},
			},
		},
		// Update Mutation
		{
			&Mutation{opUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", []byte(nil)}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Update{
					Update: &sppb.Mutation_Write{
						Table:   "t_foo",
						Columns: []string{"col1", "col2"},
						Values: []*proto3.ListValue{
							&proto3.ListValue{
								Values: []*proto3.Value{stringProto("one"), nullProto()},
							},
						},
					},
				},
			},
		},
	} {
		if got, err := test.m.proto(); err != nil || !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: (%#v).proto() = (%v, %v), want (%v, nil)", i, test.m, got, err, test.want)
		}
	}
}

// mutationColumnSorter implements sort.Interface for sorting column-value pairs in a Mutation by column names.
type mutationColumnSorter struct {
	Mutation
}

// newMutationColumnSorter creates new instance of mutationColumnSorter by duplicating the input Mutation so that
// sorting won't change the input Mutation.
func newMutationColumnSorter(m *Mutation) *mutationColumnSorter {
	return &mutationColumnSorter{
		Mutation{
			m.op,
			m.table,
			m.keySet,
			append([]string(nil), m.columns...),
			append([]interface{}(nil), m.values...),
		},
	}
}

// Len implements sort.Interface.Len.
func (ms *mutationColumnSorter) Len() int {
	return len(ms.columns)
}

// Swap implements sort.Interface.Swap.
func (ms *mutationColumnSorter) Swap(i, j int) {
	ms.columns[i], ms.columns[j] = ms.columns[j], ms.columns[i]
	ms.values[i], ms.values[j] = ms.values[j], ms.values[i]
}

// Less implements sort.Interface.Less.
func (ms *mutationColumnSorter) Less(i, j int) bool {
	return strings.Compare(ms.columns[i], ms.columns[j]) < 0
}

// mutationEqual returns true if two mutations in question are equal
// to each other.
func mutationEqual(t *testing.T, m1, m2 Mutation) bool {
	// Two mutations are considered to be equal even if their column values have different
	// orders.
	ms1 := newMutationColumnSorter(&m1)
	ms2 := newMutationColumnSorter(&m2)
	sort.Sort(ms1)
	sort.Sort(ms2)
	return reflect.DeepEqual(ms1, ms2)
}

// Test helper functions which help to generate spanner.Mutation.
func TestMutationHelpers(t *testing.T) {
	for _, test := range []struct {
		m    string
		got  *Mutation
		want *Mutation
	}{
		{
			"Insert",
			Insert("t_foo", []string{"col1", "col2"}, []interface{}{int64(1), int64(2)}),
			&Mutation{opInsert, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{int64(1), int64(2)}},
		},
		{
			"InsertMap",
			InsertMap("t_foo", map[string]interface{}{"col1": int64(1), "col2": int64(2)}),
			&Mutation{opInsert, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{int64(1), int64(2)}},
		},
		{
			"InsertStruct",
			func() *Mutation {
				m, err := InsertStruct(
					"t_foo",
					struct {
						notCol bool
						Col1   int64 `spanner:"col1"`
						Col2   int64 `spanner:"col2"`
					}{false, int64(1), int64(2)},
				)
				if err != nil {
					t.Errorf("cannot convert struct into mutation: %v", err)
				}
				return m
			}(),
			&Mutation{opInsert, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{int64(1), int64(2)}},
		},
		{
			"Update",
			Update("t_foo", []string{"col1", "col2"}, []interface{}{"one", []byte(nil)}),
			&Mutation{opUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", []byte(nil)}},
		},
		{
			"UpdateMap",
			UpdateMap("t_foo", map[string]interface{}{"col1": "one", "col2": []byte(nil)}),
			&Mutation{opUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", []byte(nil)}},
		},
		{
			"UpdateStruct",
			func() *Mutation {
				m, err := UpdateStruct(
					"t_foo",
					struct {
						Col1   string `spanner:"col1"`
						notCol int
						Col2   []byte `spanner:"col2"`
					}{"one", 1, nil},
				)
				if err != nil {
					t.Errorf("cannot convert struct into mutation: %v", err)
				}
				return m
			}(),
			&Mutation{opUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", []byte(nil)}},
		},
		{
			"InsertOrUpdate",
			InsertOrUpdate("t_foo", []string{"col1", "col2"}, []interface{}{1.0, 2.0}),
			&Mutation{opInsertOrUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{1.0, 2.0}},
		},
		{
			"InsertOrUpdateMap",
			InsertOrUpdateMap("t_foo", map[string]interface{}{"col1": 1.0, "col2": 2.0}),
			&Mutation{opInsertOrUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{1.0, 2.0}},
		},
		{
			"InsertOrUpdateStruct",
			func() *Mutation {
				m, err := InsertOrUpdateStruct(
					"t_foo",
					struct {
						Col1   float64 `spanner:"col1"`
						Col2   float64 `spanner:"col2"`
						notCol float64
					}{1.0, 2.0, 3.0},
				)
				if err != nil {
					t.Errorf("cannot convert struct into mutation: %v", err)
				}
				return m
			}(),
			&Mutation{opInsertOrUpdate, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{1.0, 2.0}},
		},
		{
			"Replace",
			Replace("t_foo", []string{"col1", "col2"}, []interface{}{"one", 2.0}),
			&Mutation{opReplace, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", 2.0}},
		},
		{
			"ReplaceMap",
			ReplaceMap("t_foo", map[string]interface{}{"col1": "one", "col2": 2.0}),
			&Mutation{opReplace, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", 2.0}},
		},
		{
			"ReplaceStruct",
			func() *Mutation {
				m, err := ReplaceStruct(
					"t_foo",
					struct {
						Col1   string  `spanner:"col1"`
						Col2   float64 `spanner:"col2"`
						notCol string
					}{"one", 2.0, "foo"},
				)
				if err != nil {
					t.Errorf("cannot convert struct into mutation: %v", err)
				}
				return m
			}(),
			&Mutation{opReplace, "t_foo", KeySet{}, []string{"col1", "col2"}, []interface{}{"one", 2.0}},
		},
		{
			"Delete",
			Delete("t_foo", Key{"foo"}),
			&Mutation{opDelete, "t_foo", Keys(Key{"foo"}), nil, nil},
		},
		{
			"DeleteRange",
			DeleteKeyRange("t_foo", KeyRange{Key{"bar"}, Key{"foo"}, ClosedClosed}),
			&Mutation{opDelete, "t_foo", Range(KeyRange{Key{"bar"}, Key{"foo"}, ClosedClosed}), nil, nil},
		},
	} {
		if !mutationEqual(t, *test.got, *test.want) {
			t.Errorf("%v: got Mutation %v, want %v", test.m, test.got, test.want)
		}
	}
}

// Test encoding non-struct types by using *Struct helpers.
func TestBadStructs(t *testing.T) {
	val := "i_am_not_a_struct"
	wantErr := errNotStruct(val)
	if _, gotErr := InsertStruct("t_test", val); !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("InsertStruct(%q) returns error %v, want %v", val, gotErr, wantErr)
	}
	if _, gotErr := InsertOrUpdateStruct("t_test", val); !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("InsertOrUpdateStruct(%q) returns error %v, want %v", val, gotErr, wantErr)
	}
	if _, gotErr := UpdateStruct("t_test", val); !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("UpdateStruct(%q) returns error %v, want %v", val, gotErr, wantErr)
	}
	if _, gotErr := ReplaceStruct("t_test", val); !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("ReplaceStruct(%q) returns error %v, want %v", val, gotErr, wantErr)
	}
}

// Test encoding Mutation into proto.
func TestEncodeMutation(t *testing.T) {
	for _, test := range []struct {
		name      string
		mutation  Mutation
		wantProto *sppb.Mutation
		wantErr   error
	}{
		{
			"OpDelete",
			Mutation{opDelete, "t_test", Keys(Key{1}), nil, nil},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Delete_{
					Delete: &sppb.Mutation_Delete{
						Table: "t_test",
						KeySet: &sppb.KeySet{
							Keys:   []*proto3.ListValue{listValueProto(intProto(1))},
							Ranges: []*sppb.KeyRange{},
						},
					},
				},
			},
			nil,
		},
		{
			"OpDelete - Key error",
			Mutation{opDelete, "t_test", Keys(Key{struct{}{}}), nil, nil},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Delete_{
					Delete: &sppb.Mutation_Delete{
						Table:  "t_test",
						KeySet: &sppb.KeySet{},
					},
				},
			},
			errInvdKeyPartType(struct{}{}),
		},
		{
			"OpInsert",
			Mutation{opInsert, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{"foo", 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Insert{
					Insert: &sppb.Mutation_Write{
						Table:   "t_test",
						Columns: []string{"key", "val"},
						Values:  []*proto3.ListValue{listValueProto(stringProto("foo"), intProto(1))},
					},
				},
			},
			nil,
		},
		{
			"OpInsert - Value Type Error",
			Mutation{opInsert, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{struct{}{}, 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Insert{
					Insert: &sppb.Mutation_Write{},
				},
			},
			errEncoderUnsupportedType(struct{}{}),
		},
		{
			"OpInsertOrUpdate",
			Mutation{opInsertOrUpdate, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{"foo", 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_InsertOrUpdate{
					InsertOrUpdate: &sppb.Mutation_Write{
						Table:   "t_test",
						Columns: []string{"key", "val"},
						Values:  []*proto3.ListValue{listValueProto(stringProto("foo"), intProto(1))},
					},
				},
			},
			nil,
		},
		{
			"OpInsertOrUpdate - Value Type Error",
			Mutation{opInsertOrUpdate, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{struct{}{}, 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_InsertOrUpdate{
					InsertOrUpdate: &sppb.Mutation_Write{},
				},
			},
			errEncoderUnsupportedType(struct{}{}),
		},
		{
			"OpReplace",
			Mutation{opReplace, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{"foo", 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Replace{
					Replace: &sppb.Mutation_Write{
						Table:   "t_test",
						Columns: []string{"key", "val"},
						Values:  []*proto3.ListValue{listValueProto(stringProto("foo"), intProto(1))},
					},
				},
			},
			nil,
		},
		{
			"OpReplace - Value Type Error",
			Mutation{opReplace, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{struct{}{}, 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Replace{
					Replace: &sppb.Mutation_Write{},
				},
			},
			errEncoderUnsupportedType(struct{}{}),
		},
		{
			"OpUpdate",
			Mutation{opUpdate, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{"foo", 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Update{
					Update: &sppb.Mutation_Write{
						Table:   "t_test",
						Columns: []string{"key", "val"},
						Values:  []*proto3.ListValue{listValueProto(stringProto("foo"), intProto(1))},
					},
				},
			},
			nil,
		},
		{
			"OpUpdate - Value Type Error",
			Mutation{opUpdate, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{struct{}{}, 1}},
			&sppb.Mutation{
				Operation: &sppb.Mutation_Update{
					Update: &sppb.Mutation_Write{},
				},
			},
			errEncoderUnsupportedType(struct{}{}),
		},
		{
			"OpKnown - Unknown Mutation Operation Code",
			Mutation{op(100), "t_test", KeySet{}, nil, nil},
			&sppb.Mutation{},
			errInvdMutationOp(Mutation{op(100), "t_test", KeySet{}, nil, nil}),
		},
	} {
		gotProto, gotErr := test.mutation.proto()
		if gotErr != nil {
			if !reflect.DeepEqual(gotErr, test.wantErr) {
				t.Errorf("%v: %v.proto() returns error %v, want %v", test.name, test.mutation, gotErr, test.wantErr)
			}
			continue
		}
		if !reflect.DeepEqual(gotProto, test.wantProto) {
			t.Errorf("%v: %v.proto() = (%v, nil), want (%v, nil)", test.name, test.mutation, gotProto, test.wantProto)
		}
	}
}

// Test Encoding an array of mutations.
func TestEncodeMutationArray(t *testing.T) {
	for _, test := range []struct {
		name    string
		ms      []*Mutation
		want    []*sppb.Mutation
		wantErr error
	}{
		{
			"Multiple Mutations",
			[]*Mutation{
				&Mutation{opDelete, "t_test", Keys(Key{"bar"}), nil, nil},
				&Mutation{opInsertOrUpdate, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{"foo", 1}},
			},
			[]*sppb.Mutation{
				&sppb.Mutation{
					Operation: &sppb.Mutation_Delete_{
						Delete: &sppb.Mutation_Delete{
							Table: "t_test",
							KeySet: &sppb.KeySet{
								Keys:   []*proto3.ListValue{listValueProto(stringProto("bar"))},
								Ranges: []*sppb.KeyRange{},
							},
						},
					},
				},
				&sppb.Mutation{
					Operation: &sppb.Mutation_InsertOrUpdate{
						InsertOrUpdate: &sppb.Mutation_Write{
							Table:   "t_test",
							Columns: []string{"key", "val"},
							Values:  []*proto3.ListValue{listValueProto(stringProto("foo"), intProto(1))},
						},
					},
				},
			},
			nil,
		},
		{
			"Multiple Mutations - Bad Mutation",
			[]*Mutation{
				&Mutation{opDelete, "t_test", Keys(Key{"bar"}), nil, nil},
				&Mutation{opInsertOrUpdate, "t_test", KeySet{}, []string{"key", "val"}, []interface{}{"foo", struct{}{}}},
			},
			[]*sppb.Mutation{},
			errEncoderUnsupportedType(struct{}{}),
		},
	} {
		gotProto, gotErr := mutationsProto(test.ms)
		if gotErr != nil {
			if !reflect.DeepEqual(gotErr, test.wantErr) {
				t.Errorf("%v: mutationsProto(%v) returns error %v, want %v", test.name, test.ms, gotErr, test.wantErr)
			}
			continue
		}
		if !reflect.DeepEqual(gotProto, test.want) {
			t.Errorf("%v: mutationsProto(%v) = (%v, nil), want (%v, nil)", test.name, test.ms, gotProto, test.want)
		}
	}
}
