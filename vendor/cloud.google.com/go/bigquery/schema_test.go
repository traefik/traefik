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
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/internal/pretty"

	bq "google.golang.org/api/bigquery/v2"
)

func (fs *FieldSchema) GoString() string {
	if fs == nil {
		return "<nil>"
	}

	return fmt.Sprintf("{Name:%s Description:%s Repeated:%t Required:%t Type:%s Schema:%s}",
		fs.Name,
		fs.Description,
		fs.Repeated,
		fs.Required,
		fs.Type,
		fmt.Sprintf("%#v", fs.Schema),
	)
}

func bqTableFieldSchema(desc, name, typ, mode string) *bq.TableFieldSchema {
	return &bq.TableFieldSchema{
		Description: desc,
		Name:        name,
		Mode:        mode,
		Type:        typ,
	}
}

func fieldSchema(desc, name, typ string, repeated, required bool) *FieldSchema {
	return &FieldSchema{
		Description: desc,
		Name:        name,
		Repeated:    repeated,
		Required:    required,
		Type:        FieldType(typ),
	}
}

func TestSchemaConversion(t *testing.T) {
	testCases := []struct {
		schema   Schema
		bqSchema *bq.TableSchema
	}{
		{
			// required
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "STRING", "REQUIRED"),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "STRING", false, true),
			},
		},
		{
			// repeated
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "STRING", "REPEATED"),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "STRING", true, false),
			},
		},
		{
			// nullable, string
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "STRING", ""),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "STRING", false, false),
			},
		},
		{
			// integer
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "INTEGER", ""),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "INTEGER", false, false),
			},
		},
		{
			// float
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "FLOAT", ""),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "FLOAT", false, false),
			},
		},
		{
			// boolean
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "BOOLEAN", ""),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "BOOLEAN", false, false),
			},
		},
		{
			// timestamp
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "name", "TIMESTAMP", ""),
				},
			},
			schema: Schema{
				fieldSchema("desc", "name", "TIMESTAMP", false, false),
			},
		},
		{
			// civil times
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					bqTableFieldSchema("desc", "f1", "TIME", ""),
					bqTableFieldSchema("desc", "f2", "DATE", ""),
					bqTableFieldSchema("desc", "f3", "DATETIME", ""),
				},
			},
			schema: Schema{
				fieldSchema("desc", "f1", "TIME", false, false),
				fieldSchema("desc", "f2", "DATE", false, false),
				fieldSchema("desc", "f3", "DATETIME", false, false),
			},
		},
		{
			// nested
			bqSchema: &bq.TableSchema{
				Fields: []*bq.TableFieldSchema{
					{
						Description: "An outer schema wrapping a nested schema",
						Name:        "outer",
						Mode:        "REQUIRED",
						Type:        "RECORD",
						Fields: []*bq.TableFieldSchema{
							bqTableFieldSchema("inner field", "inner", "STRING", ""),
						},
					},
				},
			},
			schema: Schema{
				&FieldSchema{
					Description: "An outer schema wrapping a nested schema",
					Name:        "outer",
					Required:    true,
					Type:        "RECORD",
					Schema: []*FieldSchema{
						{
							Description: "inner field",
							Name:        "inner",
							Type:        "STRING",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		bqSchema := tc.schema.asTableSchema()
		if !reflect.DeepEqual(bqSchema, tc.bqSchema) {
			t.Errorf("converting to TableSchema: got:\n%v\nwant:\n%v",
				pretty.Value(bqSchema), pretty.Value(tc.bqSchema))
		}
		schema := convertTableSchema(tc.bqSchema)
		if !reflect.DeepEqual(schema, tc.schema) {
			t.Errorf("converting to Schema: got:\n%v\nwant:\n%v", schema, tc.schema)
		}
	}
}

type allStrings struct {
	String    string
	ByteSlice []byte
}

type allSignedIntegers struct {
	Int64 int64
	Int32 int32
	Int16 int16
	Int8  int8
	Int   int
}

type allUnsignedIntegers struct {
	Uint32 uint32
	Uint16 uint16
	Uint8  uint8
}

type allFloat struct {
	Float64 float64
	Float32 float32
	// NOTE: Complex32 and Complex64 are unsupported by BigQuery
}

type allBoolean struct {
	Bool bool
}

type allTime struct {
	Timestamp time.Time
	Time      civil.Time
	Date      civil.Date
	DateTime  civil.DateTime
}

func reqField(name, typ string) *FieldSchema {
	return &FieldSchema{
		Name:     name,
		Type:     FieldType(typ),
		Required: true,
	}
}

func TestSimpleInference(t *testing.T) {
	testCases := []struct {
		in   interface{}
		want Schema
	}{
		{
			in: allSignedIntegers{},
			want: Schema{
				reqField("Int64", "INTEGER"),
				reqField("Int32", "INTEGER"),
				reqField("Int16", "INTEGER"),
				reqField("Int8", "INTEGER"),
				reqField("Int", "INTEGER"),
			},
		},
		{
			in: allUnsignedIntegers{},
			want: Schema{
				reqField("Uint32", "INTEGER"),
				reqField("Uint16", "INTEGER"),
				reqField("Uint8", "INTEGER"),
			},
		},
		{
			in: allFloat{},
			want: Schema{
				reqField("Float64", "FLOAT"),
				reqField("Float32", "FLOAT"),
			},
		},
		{
			in: allBoolean{},
			want: Schema{
				reqField("Bool", "BOOLEAN"),
			},
		},
		{
			in: &allBoolean{},
			want: Schema{
				reqField("Bool", "BOOLEAN"),
			},
		},
		{
			in: allTime{},
			want: Schema{
				reqField("Timestamp", "TIMESTAMP"),
				reqField("Time", "TIME"),
				reqField("Date", "DATE"),
				reqField("DateTime", "DATETIME"),
			},
		},
		{
			in: allStrings{},
			want: Schema{
				reqField("String", "STRING"),
				reqField("ByteSlice", "BYTES"),
			},
		},
	}
	for _, tc := range testCases {
		got, err := InferSchema(tc.in)
		if err != nil {
			t.Fatalf("%T: error inferring TableSchema: %v", tc.in, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%T: inferring TableSchema: got:\n%#v\nwant:\n%#v", tc.in,
				pretty.Value(got), pretty.Value(tc.want))
		}
	}
}

type containsNested struct {
	hidden    string
	NotNested int
	Nested    struct {
		Inside int
	}
}

type containsDoubleNested struct {
	NotNested int
	Nested    struct {
		InsideNested struct {
			Inside int
		}
	}
}

type ptrNested struct {
	Ptr *struct{ Inside int }
}

type dup struct { // more than one field of the same struct type
	A, B allBoolean
}

func TestNestedInference(t *testing.T) {
	testCases := []struct {
		in   interface{}
		want Schema
	}{
		{
			in: containsNested{},
			want: Schema{
				reqField("NotNested", "INTEGER"),
				&FieldSchema{
					Name:     "Nested",
					Required: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("Inside", "INTEGER")},
				},
			},
		},
		{
			in: containsDoubleNested{},
			want: Schema{
				reqField("NotNested", "INTEGER"),
				&FieldSchema{
					Name:     "Nested",
					Required: true,
					Type:     "RECORD",
					Schema: Schema{
						{
							Name:     "InsideNested",
							Required: true,
							Type:     "RECORD",
							Schema:   Schema{reqField("Inside", "INTEGER")},
						},
					},
				},
			},
		},
		{
			in: ptrNested{},
			want: Schema{
				&FieldSchema{
					Name:     "Ptr",
					Required: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("Inside", "INTEGER")},
				},
			},
		},
		{
			in: dup{},
			want: Schema{
				&FieldSchema{
					Name:     "A",
					Required: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("Bool", "BOOLEAN")},
				},
				&FieldSchema{
					Name:     "B",
					Required: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("Bool", "BOOLEAN")},
				},
			},
		},
	}

	for _, tc := range testCases {
		got, err := InferSchema(tc.in)
		if err != nil {
			t.Fatalf("%T: error inferring TableSchema: %v", tc.in, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%T: inferring TableSchema: got:\n%#v\nwant:\n%#v", tc.in,
				pretty.Value(got), pretty.Value(tc.want))
		}
	}
}

type repeated struct {
	NotRepeated       []byte
	RepeatedByteSlice [][]byte
	Slice             []int
	Array             [5]bool
}

type nestedRepeated struct {
	NotRepeated int
	Repeated    []struct {
		Inside int
	}
	RepeatedPtr []*struct{ Inside int }
}

func repField(name, typ string) *FieldSchema {
	return &FieldSchema{
		Name:     name,
		Type:     FieldType(typ),
		Repeated: true,
	}
}

func TestRepeatedInference(t *testing.T) {
	testCases := []struct {
		in   interface{}
		want Schema
	}{
		{
			in: repeated{},
			want: Schema{
				reqField("NotRepeated", "BYTES"),
				repField("RepeatedByteSlice", "BYTES"),
				repField("Slice", "INTEGER"),
				repField("Array", "BOOLEAN"),
			},
		},
		{
			in: nestedRepeated{},
			want: Schema{
				reqField("NotRepeated", "INTEGER"),
				{
					Name:     "Repeated",
					Repeated: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("Inside", "INTEGER")},
				},
				{
					Name:     "RepeatedPtr",
					Repeated: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("Inside", "INTEGER")},
				},
			},
		},
	}

	for i, tc := range testCases {
		got, err := InferSchema(tc.in)
		if err != nil {
			t.Fatalf("%d: error inferring TableSchema: %v", i, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%d: inferring TableSchema: got:\n%#v\nwant:\n%#v", i,
				pretty.Value(got), pretty.Value(tc.want))
		}
	}
}

type Embedded struct {
	Embedded int
}

type embedded struct {
	Embedded2 int
}

type nestedEmbedded struct {
	Embedded
	embedded
}

func TestEmbeddedInference(t *testing.T) {
	got, err := InferSchema(nestedEmbedded{})
	if err != nil {
		t.Fatal(err)
	}
	want := Schema{
		reqField("Embedded", "INTEGER"),
		reqField("Embedded2", "INTEGER"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", pretty.Value(got), pretty.Value(want))
	}
}

func TestRecursiveInference(t *testing.T) {
	type List struct {
		Val  int
		Next *List
	}

	_, err := InferSchema(List{})
	if err == nil {
		t.Fatal("got nil, want error")
	}
}

type withTags struct {
	NoTag         int
	ExcludeTag    int `bigquery:"-"`
	SimpleTag     int `bigquery:"simple_tag"`
	UnderscoreTag int `bigquery:"_id"`
	MixedCase     int `bigquery:"MIXEDcase"`
}

type withTagsNested struct {
	Nested          withTags `bigquery:"nested"`
	NestedAnonymous struct {
		ExcludeTag int `bigquery:"-"`
		Inside     int `bigquery:"inside"`
	} `bigquery:"anon"`
}

type withTagsRepeated struct {
	Repeated          []withTags `bigquery:"repeated"`
	RepeatedAnonymous []struct {
		ExcludeTag int `bigquery:"-"`
		Inside     int `bigquery:"inside"`
	} `bigquery:"anon"`
}

type withTagsEmbedded struct {
	withTags
}

var withTagsSchema = Schema{
	reqField("NoTag", "INTEGER"),
	reqField("simple_tag", "INTEGER"),
	reqField("_id", "INTEGER"),
	reqField("MIXEDcase", "INTEGER"),
}

func TestTagInference(t *testing.T) {
	testCases := []struct {
		in   interface{}
		want Schema
	}{
		{
			in:   withTags{},
			want: withTagsSchema,
		},
		{
			in: withTagsNested{},
			want: Schema{
				&FieldSchema{
					Name:     "nested",
					Required: true,
					Type:     "RECORD",
					Schema:   withTagsSchema,
				},
				&FieldSchema{
					Name:     "anon",
					Required: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("inside", "INTEGER")},
				},
			},
		},
		{
			in: withTagsRepeated{},
			want: Schema{
				&FieldSchema{
					Name:     "repeated",
					Repeated: true,
					Type:     "RECORD",
					Schema:   withTagsSchema,
				},
				&FieldSchema{
					Name:     "anon",
					Repeated: true,
					Type:     "RECORD",
					Schema:   Schema{reqField("inside", "INTEGER")},
				},
			},
		},
		{
			in:   withTagsEmbedded{},
			want: withTagsSchema,
		},
	}
	for i, tc := range testCases {
		got, err := InferSchema(tc.in)
		if err != nil {
			t.Fatalf("%d: error inferring TableSchema: %v", i, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%d: inferring TableSchema: got:\n%#v\nwant:\n%#v", i,
				pretty.Value(got), pretty.Value(tc.want))
		}
	}
}

func TestTagInferenceErrors(t *testing.T) {
	testCases := []struct {
		in  interface{}
		err error
	}{
		{
			in: struct {
				LongTag int `bigquery:"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxy"`
			}{},
			err: errInvalidFieldName,
		},
		{
			in: struct {
				UnsupporedStartChar int `bigquery:"øab"`
			}{},
			err: errInvalidFieldName,
		},
		{
			in: struct {
				UnsupportedEndChar int `bigquery:"abø"`
			}{},
			err: errInvalidFieldName,
		},
		{
			in: struct {
				UnsupportedMiddleChar int `bigquery:"aøb"`
			}{},
			err: errInvalidFieldName,
		},
		{
			in: struct {
				StartInt int `bigquery:"1abc"`
			}{},
			err: errInvalidFieldName,
		},
		{
			in: struct {
				Hyphens int `bigquery:"a-b"`
			}{},
			err: errInvalidFieldName,
		},
		{
			in: struct {
				OmitEmpty int `bigquery:"abc,omitempty"`
			}{},
			err: errInvalidFieldName,
		},
	}
	for i, tc := range testCases {
		want := tc.err
		_, got := InferSchema(tc.in)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: inferring TableSchema: got:\n%#v\nwant:\n%#v", i, got, want)
		}
	}
}

func TestSchemaErrors(t *testing.T) {
	testCases := []struct {
		in  interface{}
		err error
	}{
		{
			in:  []byte{},
			err: errNoStruct,
		},
		{
			in:  new(int),
			err: errNoStruct,
		},
		{
			in:  struct{ Uint uint }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ Uint64 uint64 }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ Uintptr uintptr }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ Complex complex64 }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ Map map[string]int }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ Chan chan bool }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ Ptr *int }{},
			err: errNoStruct,
		},
		{
			in:  struct{ Interface interface{} }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ MultiDimensional [][]int }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ MultiDimensional [][][]byte }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ ChanSlice []chan bool }{},
			err: errUnsupportedFieldType,
		},
		{
			in:  struct{ NestedChan struct{ Chan []chan bool } }{},
			err: errUnsupportedFieldType,
		},
	}
	for _, tc := range testCases {
		want := tc.err
		_, got := InferSchema(tc.in)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%#v: got:\n%#v\nwant:\n%#v", tc.in, got, want)
		}
	}
}

func TestHasRecursiveType(t *testing.T) {
	type (
		nonStruct int
		nonRec    struct{ A string }
		dup       struct{ A, B nonRec }
		rec       struct {
			A int
			B *rec
		}
		recUnexported struct {
			A int
			b *rec
		}
		hasRec struct {
			A int
			R *rec
		}
	)
	for _, test := range []struct {
		in   interface{}
		want bool
	}{
		{nonStruct(0), false},
		{nonRec{}, false},
		{dup{}, false},
		{rec{}, true},
		{recUnexported{}, false},
		{hasRec{}, true},
	} {
		got, err := hasRecursiveType(reflect.TypeOf(test.in), nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("%T: got %t, want %t", test.in, got, test.want)
		}
	}
}
