// Copyright 2015 Google Inc. All Rights Reserved.
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

package bigquery

import (
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/internal/pretty"

	bq "google.golang.org/api/bigquery/v2"
)

func TestConvertBasicValues(t *testing.T) {
	schema := []*FieldSchema{
		{Type: StringFieldType},
		{Type: IntegerFieldType},
		{Type: FloatFieldType},
		{Type: BooleanFieldType},
		{Type: BytesFieldType},
	}
	row := &bq.TableRow{
		F: []*bq.TableCell{
			{V: "a"},
			{V: "1"},
			{V: "1.2"},
			{V: "true"},
			{V: base64.StdEncoding.EncodeToString([]byte("foo"))},
		},
	}
	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	want := []Value{"a", int64(1), 1.2, true, []byte("foo")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("converting basic values: got:\n%v\nwant:\n%v", got, want)
	}
}

func TestConvertTime(t *testing.T) {
	// TODO(jba): add tests for civil time types.
	schema := []*FieldSchema{
		{Type: TimestampFieldType},
	}
	thyme := time.Date(1970, 1, 1, 10, 0, 0, 10, time.UTC)
	row := &bq.TableRow{
		F: []*bq.TableCell{
			{V: fmt.Sprintf("%.10f", float64(thyme.UnixNano())/1e9)},
		},
	}
	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	if !got[0].(time.Time).Equal(thyme) {
		t.Errorf("converting basic values: got:\n%v\nwant:\n%v", got, thyme)
	}
	if got[0].(time.Time).Location() != time.UTC {
		t.Errorf("expected time zone UTC: got:\n%v", got)
	}
}

func TestConvertNullValues(t *testing.T) {
	schema := []*FieldSchema{
		{Type: StringFieldType},
	}
	row := &bq.TableRow{
		F: []*bq.TableCell{
			{V: nil},
		},
	}
	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	want := []Value{nil}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("converting null values: got:\n%v\nwant:\n%v", got, want)
	}
}

func TestBasicRepetition(t *testing.T) {
	schema := []*FieldSchema{
		{Type: IntegerFieldType, Repeated: true},
	}
	row := &bq.TableRow{
		F: []*bq.TableCell{
			{
				V: []interface{}{
					map[string]interface{}{
						"v": "1",
					},
					map[string]interface{}{
						"v": "2",
					},
					map[string]interface{}{
						"v": "3",
					},
				},
			},
		},
	}
	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	want := []Value{[]Value{int64(1), int64(2), int64(3)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("converting basic repeated values: got:\n%v\nwant:\n%v", got, want)
	}
}

func TestNestedRecordContainingRepetition(t *testing.T) {
	schema := []*FieldSchema{
		{
			Type: RecordFieldType,
			Schema: Schema{
				{Type: IntegerFieldType, Repeated: true},
			},
		},
	}
	row := &bq.TableRow{
		F: []*bq.TableCell{
			{
				V: map[string]interface{}{
					"f": []interface{}{
						map[string]interface{}{
							"v": []interface{}{
								map[string]interface{}{"v": "1"},
								map[string]interface{}{"v": "2"},
								map[string]interface{}{"v": "3"},
							},
						},
					},
				},
			},
		},
	}

	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	want := []Value{[]Value{[]Value{int64(1), int64(2), int64(3)}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("converting basic repeated values: got:\n%v\nwant:\n%v", got, want)
	}
}

func TestRepeatedRecordContainingRepetition(t *testing.T) {
	schema := []*FieldSchema{
		{
			Type:     RecordFieldType,
			Repeated: true,
			Schema: Schema{
				{Type: IntegerFieldType, Repeated: true},
			},
		},
	}
	row := &bq.TableRow{F: []*bq.TableCell{
		{
			V: []interface{}{ // repeated records.
				map[string]interface{}{ // first record.
					"v": map[string]interface{}{ // pointless single-key-map wrapper.
						"f": []interface{}{ // list of record fields.
							map[string]interface{}{ // only record (repeated ints)
								"v": []interface{}{ // pointless wrapper.
									map[string]interface{}{
										"v": "1",
									},
									map[string]interface{}{
										"v": "2",
									},
									map[string]interface{}{
										"v": "3",
									},
								},
							},
						},
					},
				},
				map[string]interface{}{ // second record.
					"v": map[string]interface{}{
						"f": []interface{}{
							map[string]interface{}{
								"v": []interface{}{
									map[string]interface{}{
										"v": "4",
									},
									map[string]interface{}{
										"v": "5",
									},
									map[string]interface{}{
										"v": "6",
									},
								},
							},
						},
					},
				},
			},
		},
	}}

	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	want := []Value{ // the row is a list of length 1, containing an entry for the repeated record.
		[]Value{ // the repeated record is a list of length 2, containing an entry for each repetition.
			[]Value{ // the record is a list of length 1, containing an entry for the repeated integer field.
				[]Value{int64(1), int64(2), int64(3)}, // the repeated integer field is a list of length 3.
			},
			[]Value{ // second record
				[]Value{int64(4), int64(5), int64(6)},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("converting repeated records with repeated values: got:\n%v\nwant:\n%v", got, want)
	}
}

func TestRepeatedRecordContainingRecord(t *testing.T) {
	schema := []*FieldSchema{
		{
			Type:     RecordFieldType,
			Repeated: true,
			Schema: Schema{
				{
					Type: StringFieldType,
				},
				{
					Type: RecordFieldType,
					Schema: Schema{
						{Type: IntegerFieldType},
						{Type: StringFieldType},
					},
				},
			},
		},
	}
	row := &bq.TableRow{F: []*bq.TableCell{
		{
			V: []interface{}{ // repeated records.
				map[string]interface{}{ // first record.
					"v": map[string]interface{}{ // pointless single-key-map wrapper.
						"f": []interface{}{ // list of record fields.
							map[string]interface{}{ // first record field (name)
								"v": "first repeated record",
							},
							map[string]interface{}{ // second record field (nested record).
								"v": map[string]interface{}{ // pointless single-key-map wrapper.
									"f": []interface{}{ // nested record fields
										map[string]interface{}{
											"v": "1",
										},
										map[string]interface{}{
											"v": "two",
										},
									},
								},
							},
						},
					},
				},
				map[string]interface{}{ // second record.
					"v": map[string]interface{}{
						"f": []interface{}{
							map[string]interface{}{
								"v": "second repeated record",
							},
							map[string]interface{}{
								"v": map[string]interface{}{
									"f": []interface{}{
										map[string]interface{}{
											"v": "3",
										},
										map[string]interface{}{
											"v": "four",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}}

	got, err := convertRow(row, schema)
	if err != nil {
		t.Fatalf("error converting: %v", err)
	}
	// TODO: test with flattenresults.
	want := []Value{ // the row is a list of length 1, containing an entry for the repeated record.
		[]Value{ // the repeated record is a list of length 2, containing an entry for each repetition.
			[]Value{ // record contains a string followed by a nested record.
				"first repeated record",
				[]Value{
					int64(1),
					"two",
				},
			},
			[]Value{ // second record.
				"second repeated record",
				[]Value{
					int64(3),
					"four",
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("converting repeated records containing record : got:\n%v\nwant:\n%v", got, want)
	}
}

func TestValuesSaverConvertsToMap(t *testing.T) {
	testCases := []struct {
		vs   ValuesSaver
		want *insertionRow
	}{
		{
			vs: ValuesSaver{
				Schema: []*FieldSchema{
					{Name: "intField", Type: IntegerFieldType},
					{Name: "strField", Type: StringFieldType},
				},
				InsertID: "iid",
				Row:      []Value{1, "a"},
			},
			want: &insertionRow{
				InsertID: "iid",
				Row:      map[string]Value{"intField": 1, "strField": "a"},
			},
		},
		{
			vs: ValuesSaver{
				Schema: []*FieldSchema{
					{Name: "intField", Type: IntegerFieldType},
					{
						Name: "recordField",
						Type: RecordFieldType,
						Schema: []*FieldSchema{
							{Name: "nestedInt", Type: IntegerFieldType, Repeated: true},
						},
					},
				},
				InsertID: "iid",
				Row:      []Value{1, []Value{[]Value{2, 3}}},
			},
			want: &insertionRow{
				InsertID: "iid",
				Row: map[string]Value{
					"intField": 1,
					"recordField": map[string]Value{
						"nestedInt": []Value{2, 3},
					},
				},
			},
		},
		{ // repeated nested field
			vs: ValuesSaver{
				Schema: Schema{
					{
						Name: "records",
						Type: RecordFieldType,
						Schema: Schema{
							{Name: "x", Type: IntegerFieldType},
							{Name: "y", Type: IntegerFieldType},
						},
						Repeated: true,
					},
				},
				InsertID: "iid",
				Row: []Value{ // a row is a []Value
					[]Value{ // repeated field's value is a []Value
						[]Value{1, 2}, // first record of the repeated field
						[]Value{3, 4}, // second record
					},
				},
			},
			want: &insertionRow{
				InsertID: "iid",
				Row: map[string]Value{
					"records": []Value{
						map[string]Value{"x": 1, "y": 2},
						map[string]Value{"x": 3, "y": 4},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		data, insertID, err := tc.vs.Save()
		if err != nil {
			t.Errorf("Expected successful save; got: %v", err)
		}
		got := &insertionRow{insertID, data}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("saving ValuesSaver:\ngot:\n%+v\nwant:\n%+v", got, tc.want)
		}
	}
}

func TestStructSaver(t *testing.T) {
	schema := Schema{
		{Name: "s", Type: StringFieldType},
		{Name: "r", Type: IntegerFieldType, Repeated: true},
		{Name: "nested", Type: RecordFieldType, Schema: Schema{
			{Name: "b", Type: BooleanFieldType},
		}},
		{Name: "rnested", Type: RecordFieldType, Repeated: true, Schema: Schema{
			{Name: "b", Type: BooleanFieldType},
		}},
	}

	type (
		N struct{ B bool }
		T struct {
			S       string
			R       []int
			Nested  *N
			Rnested []*N
		}
	)

	check := func(msg string, in interface{}, want map[string]Value) {
		ss := StructSaver{
			Schema:   schema,
			InsertID: "iid",
			Struct:   in,
		}
		got, gotIID, err := ss.Save()
		if err != nil {
			t.Fatalf("%s: %v", msg, err)
		}
		if wantIID := "iid"; gotIID != wantIID {
			t.Errorf("%s: InsertID: got %q, want %q", msg, gotIID, wantIID)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s:\ngot\n%#v\nwant\n%#v", msg, got, want)
		}
	}

	in := T{
		S:       "x",
		R:       []int{1, 2},
		Nested:  &N{B: true},
		Rnested: []*N{{true}, {false}},
	}
	want := map[string]Value{
		"s":       "x",
		"r":       []int{1, 2},
		"nested":  map[string]Value{"b": true},
		"rnested": []Value{map[string]Value{"b": true}, map[string]Value{"b": false}},
	}
	check("all values", in, want)
	check("all values, ptr", &in, want)
	check("empty struct", T{}, map[string]Value{"s": ""})

	// Missing and extra fields ignored.
	type T2 struct {
		S string
		// missing R, Nested, RNested
		Extra int
	}
	check("missing and extra", T2{S: "x"}, map[string]Value{"s": "x"})

	check("nils in slice", T{Rnested: []*N{{true}, nil, {false}}},
		map[string]Value{
			"s":       "",
			"rnested": []Value{map[string]Value{"b": true}, map[string]Value(nil), map[string]Value{"b": false}},
		})
}

func TestConvertRows(t *testing.T) {
	schema := []*FieldSchema{
		{Type: StringFieldType},
		{Type: IntegerFieldType},
		{Type: FloatFieldType},
		{Type: BooleanFieldType},
	}
	rows := []*bq.TableRow{
		{F: []*bq.TableCell{
			{V: "a"},
			{V: "1"},
			{V: "1.2"},
			{V: "true"},
		}},
		{F: []*bq.TableCell{
			{V: "b"},
			{V: "2"},
			{V: "2.2"},
			{V: "false"},
		}},
	}
	want := [][]Value{
		{"a", int64(1), 1.2, true},
		{"b", int64(2), 2.2, false},
	}
	got, err := convertRows(rows, schema)
	if err != nil {
		t.Fatalf("got %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot  %v\nwant %v", got, want)
	}
}

func TestValueList(t *testing.T) {
	schema := Schema{
		{Name: "s", Type: StringFieldType},
		{Name: "i", Type: IntegerFieldType},
		{Name: "f", Type: FloatFieldType},
		{Name: "b", Type: BooleanFieldType},
	}
	want := []Value{"x", 7, 3.14, true}
	var got []Value
	vl := (*valueList)(&got)
	if err := vl.Load(want, schema); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}

	// Load truncates, not appends.
	// https://github.com/GoogleCloudPlatform/google-cloud-go/issues/437
	if err := vl.Load(want, schema); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestValueMap(t *testing.T) {
	ns := Schema{
		{Name: "x", Type: IntegerFieldType},
		{Name: "y", Type: IntegerFieldType},
	}
	schema := Schema{
		{Name: "s", Type: StringFieldType},
		{Name: "i", Type: IntegerFieldType},
		{Name: "f", Type: FloatFieldType},
		{Name: "b", Type: BooleanFieldType},
		{Name: "n", Type: RecordFieldType, Schema: ns},
		{Name: "rn", Type: RecordFieldType, Schema: ns, Repeated: true},
	}
	in := []Value{"x", 7, 3.14, true,
		[]Value{1, 2},
		[]Value{[]Value{3, 4}, []Value{5, 6}},
	}
	var vm valueMap
	if err := vm.Load(in, schema); err != nil {
		t.Fatal(err)
	}
	want := map[string]Value{
		"s": "x",
		"i": 7,
		"f": 3.14,
		"b": true,
		"n": map[string]Value{"x": 1, "y": 2},
		"rn": []Value{
			map[string]Value{"x": 3, "y": 4},
			map[string]Value{"x": 5, "y": 6},
		},
	}
	if !reflect.DeepEqual(vm, valueMap(want)) {
		t.Errorf("got\n%+v\nwant\n%+v", vm, want)
	}

}

var (
	// For testing StructLoader
	schema2 = Schema{
		{Name: "s", Type: StringFieldType},
		{Name: "s2", Type: StringFieldType},
		{Name: "by", Type: BytesFieldType},
		{Name: "I", Type: IntegerFieldType},
		{Name: "F", Type: FloatFieldType},
		{Name: "B", Type: BooleanFieldType},
		{Name: "TS", Type: TimestampFieldType},
		{Name: "D", Type: DateFieldType},
		{Name: "T", Type: TimeFieldType},
		{Name: "DT", Type: DateTimeFieldType},
		{Name: "nested", Type: RecordFieldType, Schema: Schema{
			{Name: "nestS", Type: StringFieldType},
			{Name: "nestI", Type: IntegerFieldType},
		}},
		{Name: "t", Type: StringFieldType},
	}

	testTimestamp = time.Date(2016, 11, 5, 7, 50, 22, 8, time.UTC)
	testDate      = civil.Date{2016, 11, 5}
	testTime      = civil.Time{7, 50, 22, 8}
	testDateTime  = civil.DateTime{testDate, testTime}

	testValues = []Value{"x", "y", []byte{1, 2, 3}, int64(7), 3.14, true,
		testTimestamp, testDate, testTime, testDateTime,
		[]Value{"nested", int64(17)}, "z"}
)

type testStruct1 struct {
	B bool
	I int
	times
	S      string
	S2     String
	By     []byte
	s      string
	F      float64
	Nested nested
	Tagged string `bigquery:"t"`
}

type String string

type nested struct {
	NestS string
	NestI int
}

type times struct {
	TS time.Time
	T  civil.Time
	D  civil.Date
	DT civil.DateTime
}

func TestStructLoader(t *testing.T) {
	var ts1 testStruct1
	if err := load(&ts1, schema2, testValues); err != nil {
		t.Fatal(err)
	}
	// Note: the schema field named "s" gets matched to the exported struct
	// field "S", not the unexported "s".
	want := &testStruct1{
		B:      true,
		I:      7,
		F:      3.14,
		times:  times{TS: testTimestamp, T: testTime, D: testDate, DT: testDateTime},
		S:      "x",
		S2:     "y",
		By:     []byte{1, 2, 3},
		Nested: nested{NestS: "nested", NestI: 17},
		Tagged: "z",
	}
	if !reflect.DeepEqual(&ts1, want) {
		t.Errorf("got %+v, want %+v", pretty.Value(ts1), pretty.Value(*want))
		d, _, err := pretty.Diff(*want, ts1)
		if err == nil {
			t.Logf("diff:\n%s", d)
		}
	}

	// Test pointers to nested structs.
	type nestedPtr struct{ Nested *nested }
	var np nestedPtr
	if err := load(&np, schema2, testValues); err != nil {
		t.Fatal(err)
	}
	want2 := &nestedPtr{Nested: &nested{NestS: "nested", NestI: 17}}
	if !reflect.DeepEqual(&np, want2) {
		t.Errorf("got %+v, want %+v", pretty.Value(np), pretty.Value(*want2))
	}

	// Existing values should be reused.
	nst := &nested{NestS: "x", NestI: -10}
	np = nestedPtr{Nested: nst}
	if err := load(&np, schema2, testValues); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&np, want2) {
		t.Errorf("got %+v, want %+v", pretty.Value(np), pretty.Value(*want2))
	}
	if np.Nested != nst {
		t.Error("nested struct pointers not equal")
	}
}

type repStruct struct {
	Nums      []int
	ShortNums [2]int // to test truncation
	LongNums  [5]int // to test padding with zeroes
	Nested    []*nested
}

var (
	repSchema = Schema{
		{Name: "nums", Type: IntegerFieldType, Repeated: true},
		{Name: "shortNums", Type: IntegerFieldType, Repeated: true},
		{Name: "longNums", Type: IntegerFieldType, Repeated: true},
		{Name: "nested", Type: RecordFieldType, Repeated: true, Schema: Schema{
			{Name: "nestS", Type: StringFieldType},
			{Name: "nestI", Type: IntegerFieldType},
		}},
	}
	v123      = []Value{int64(1), int64(2), int64(3)}
	repValues = []Value{v123, v123, v123,
		[]Value{
			[]Value{"x", int64(1)},
			[]Value{"y", int64(2)},
		},
	}
)

func TestStructLoaderRepeated(t *testing.T) {
	var r1 repStruct
	if err := load(&r1, repSchema, repValues); err != nil {
		t.Fatal(err)
	}
	want := repStruct{
		Nums:      []int{1, 2, 3},
		ShortNums: [...]int{1, 2}, // extra values discarded
		LongNums:  [...]int{1, 2, 3, 0, 0},
		Nested:    []*nested{{"x", 1}, {"y", 2}},
	}
	if !reflect.DeepEqual(r1, want) {
		t.Errorf("got %+v, want %+v", pretty.Value(r1), pretty.Value(want))
	}

	r2 := repStruct{
		Nums:     []int{-1, -2, -3, -4, -5},    // truncated to zero and appended to
		LongNums: [...]int{-1, -2, -3, -4, -5}, // unset elements are zeroed
	}
	if err := load(&r2, repSchema, repValues); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r2, want) {
		t.Errorf("got %+v, want %+v", pretty.Value(r2), pretty.Value(want))
	}
	if got, want := cap(r2.Nums), 5; got != want {
		t.Errorf("cap(r2.Nums) = %d, want %d", got, want)
	}

	// Short slice case.
	r3 := repStruct{Nums: []int{-1}}
	if err := load(&r3, repSchema, repValues); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r3, want) {
		t.Errorf("got %+v, want %+v", pretty.Value(r3), pretty.Value(want))
	}
	if got, want := cap(r3.Nums), 3; got != want {
		t.Errorf("cap(r3.Nums) = %d, want %d", got, want)
	}

}

func TestStructLoaderOverflow(t *testing.T) {
	type S struct {
		I int16
		F float32
	}
	schema := Schema{
		{Name: "I", Type: IntegerFieldType},
		{Name: "F", Type: FloatFieldType},
	}
	var s S
	if err := load(&s, schema, []Value{int64(math.MaxInt16 + 1), 0}); err == nil {
		t.Error("int: got nil, want error")
	}
	if err := load(&s, schema, []Value{int64(0), math.MaxFloat32 * 2}); err == nil {
		t.Error("float: got nil, want error")
	}
}

func TestStructLoaderFieldOverlap(t *testing.T) {
	// It's OK if the struct has fields that the schema does not, and vice versa.
	type S1 struct {
		I int
		X [][]int // not in the schema; does not even correspond to a valid BigQuery type
		// many schema fields missing
	}
	var s1 S1
	if err := load(&s1, schema2, testValues); err != nil {
		t.Fatal(err)
	}
	want1 := S1{I: 7}
	if !reflect.DeepEqual(s1, want1) {
		t.Errorf("got %+v, want %+v", pretty.Value(s1), pretty.Value(want1))
	}

	// It's even valid to have no overlapping fields at all.
	type S2 struct{ Z int }

	var s2 S2
	if err := load(&s2, schema2, testValues); err != nil {
		t.Fatal(err)
	}
	want2 := S2{}
	if !reflect.DeepEqual(s2, want2) {
		t.Errorf("got %+v, want %+v", pretty.Value(s2), pretty.Value(want2))
	}
}

func TestStructLoaderErrors(t *testing.T) {
	check := func(sp interface{}) {
		var sl structLoader
		err := sl.set(sp, schema2)
		if err == nil {
			t.Errorf("%T: got nil, want error", sp)
		}
	}

	type bad1 struct{ F int32 } // wrong type for FLOAT column
	check(&bad1{})

	type bad2 struct{ I uint } // unsupported integer type
	check(&bad2{})

	// Using more than one struct type with the same structLoader.
	type different struct {
		B bool
		I int
		times
		S    string
		s    string
		Nums []int
	}

	var sl structLoader
	if err := sl.set(&testStruct1{}, schema2); err != nil {
		t.Fatal(err)
	}
	err := sl.set(&different{}, schema2)
	if err == nil {
		t.Error("different struct types: got nil, want error")
	}
}

func load(pval interface{}, schema Schema, vals []Value) error {
	var sl structLoader
	if err := sl.set(pval, schema); err != nil {
		return err
	}
	return sl.Load(vals, nil)
}

func BenchmarkStructLoader_NoCompile(b *testing.B) {
	benchmarkStructLoader(b, false)
}

func BenchmarkStructLoader_Compile(b *testing.B) {
	benchmarkStructLoader(b, true)
}

func benchmarkStructLoader(b *testing.B, compile bool) {
	var ts1 testStruct1
	for i := 0; i < b.N; i++ {
		var sl structLoader
		for j := 0; j < 10; j++ {
			if err := load(&ts1, schema2, testValues); err != nil {
				b.Fatal(err)
			}
			if !compile {
				sl.typ = nil
			}
		}
	}
}
