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

package bigquery

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"golang.org/x/net/context"
	bq "google.golang.org/api/bigquery/v2"
)

var scalarTests = []struct {
	val  interface{}
	want string
}{
	{int64(0), "0"},
	{3.14, "3.14"},
	{3.14159e-87, "3.14159e-87"},
	{true, "true"},
	{"string", "string"},
	{"\u65e5\u672c\u8a9e\n", "\u65e5\u672c\u8a9e\n"},
	{math.NaN(), "NaN"},
	{[]byte("foo"), "Zm9v"}, // base64 encoding of "foo"
	{time.Date(2016, 3, 20, 4, 22, 9, 5000, time.FixedZone("neg1-2", -3720)),
		"2016-03-20 04:22:09.000005-01:02"},
	{civil.Date{2016, 3, 20}, "2016-03-20"},
	{civil.Time{4, 5, 6, 789000000}, "04:05:06.789000"},
	{civil.DateTime{civil.Date{2016, 3, 20}, civil.Time{4, 5, 6, 789000000}}, "2016-03-20 04:05:06.789000"},
}

type S1 struct {
	A int
	B *S2
	C bool
}

type S2 struct {
	D string
	e int
}

var s1 = S1{
	A: 1,
	B: &S2{D: "s"},
	C: true,
}

func sval(s string) bq.QueryParameterValue {
	return bq.QueryParameterValue{Value: s}
}

func TestParamValueScalar(t *testing.T) {
	for _, test := range scalarTests {
		got, err := paramValue(reflect.ValueOf(test.val))
		if err != nil {
			t.Errorf("%v: got %v, want nil", test.val, err)
			continue
		}
		want := sval(test.want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%v:\ngot  %+v\nwant %+v", test.val, got, want)
		}
	}
}

func TestParamValueArray(t *testing.T) {
	qpv := bq.QueryParameterValue{ArrayValues: []*bq.QueryParameterValue{
		{Value: "1"},
		{Value: "2"},
	},
	}
	for _, test := range []struct {
		val  interface{}
		want bq.QueryParameterValue
	}{
		{[]int(nil), bq.QueryParameterValue{}},
		{[]int{}, bq.QueryParameterValue{}},
		{[]int{1, 2}, qpv},
		{[2]int{1, 2}, qpv},
	} {
		got, err := paramValue(reflect.ValueOf(test.val))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%#v:\ngot  %+v\nwant %+v", test.val, got, test.want)
		}
	}
}

func TestParamValueStruct(t *testing.T) {
	got, err := paramValue(reflect.ValueOf(s1))
	if err != nil {
		t.Fatal(err)
	}
	want := bq.QueryParameterValue{
		StructValues: map[string]bq.QueryParameterValue{
			"A": sval("1"),
			"B": bq.QueryParameterValue{
				StructValues: map[string]bq.QueryParameterValue{
					"D": sval("s"),
				},
			},
			"C": sval("true"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func TestParamValueErrors(t *testing.T) {
	// paramValue lets a few invalid types through, but paramType catches them.
	// Since we never call one without the other that's fine.
	for _, val := range []interface{}{nil, new([]int)} {
		_, err := paramValue(reflect.ValueOf(val))
		if err == nil {
			t.Errorf("%v (%T): got nil, want error", val, val)
		}
	}
}

func TestParamType(t *testing.T) {
	for _, test := range []struct {
		val  interface{}
		want *bq.QueryParameterType
	}{
		{0, int64ParamType},
		{uint32(32767), int64ParamType},
		{3.14, float64ParamType},
		{float32(3.14), float64ParamType},
		{math.NaN(), float64ParamType},
		{true, boolParamType},
		{"", stringParamType},
		{"string", stringParamType},
		{time.Now(), timestampParamType},
		{[]byte("foo"), bytesParamType},
		{[]int{}, &bq.QueryParameterType{Type: "ARRAY", ArrayType: int64ParamType}},
		{[3]bool{}, &bq.QueryParameterType{Type: "ARRAY", ArrayType: boolParamType}},
		{S1{}, &bq.QueryParameterType{
			Type: "STRUCT",
			StructTypes: []*bq.QueryParameterTypeStructTypes{
				{Name: "A", Type: int64ParamType},
				{Name: "B", Type: &bq.QueryParameterType{
					Type: "STRUCT",
					StructTypes: []*bq.QueryParameterTypeStructTypes{
						{Name: "D", Type: stringParamType},
					},
				}},
				{Name: "C", Type: boolParamType},
			},
		}},
	} {
		got, err := paramType(reflect.TypeOf(test.val))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%v (%T): got %v, want %v", test.val, test.val, got, test.want)
		}
	}
}

func TestParamTypeErrors(t *testing.T) {
	for _, val := range []interface{}{
		nil, uint(0), new([]int), make(chan int),
	} {
		_, err := paramType(reflect.TypeOf(val))
		if err == nil {
			t.Errorf("%v (%T): got nil, want error", val, val)
		}
	}
}

func TestIntegration_ScalarParam(t *testing.T) {
	c := getClient(t)
	for _, test := range scalarTests {
		got, err := paramRoundTrip(c, test.val)
		if err != nil {
			t.Fatal(err)
		}
		if !equal(got, test.val) {
			t.Errorf("\ngot  %#v (%T)\nwant %#v (%T)", got, got, test.val, test.val)
		}
	}
}

func TestIntegration_OtherParam(t *testing.T) {
	c := getClient(t)
	for _, test := range []struct {
		val  interface{}
		want interface{}
	}{
		{[]int(nil), []Value(nil)},
		{[]int{}, []Value(nil)},
		{[]int{1, 2}, []Value{int64(1), int64(2)}},
		{[3]int{1, 2, 3}, []Value{int64(1), int64(2), int64(3)}},
		{S1{}, []Value{int64(0), nil, false}},
		{s1, []Value{int64(1), []Value{"s"}, true}},
	} {
		got, err := paramRoundTrip(c, test.val)
		if err != nil {
			t.Fatal(err)
		}
		if !equal(got, test.want) {
			t.Errorf("\ngot  %#v (%T)\nwant %#v (%T)", got, got, test.want, test.want)
		}
	}
}

func paramRoundTrip(c *Client, x interface{}) (Value, error) {
	q := c.Query("select ?")
	q.Parameters = []QueryParameter{{Value: x}}
	it, err := q.Read(context.Background())
	if err != nil {
		return nil, err
	}
	var val []Value
	err = it.Next(&val)
	if err != nil {
		return nil, err
	}
	if len(val) != 1 {
		return nil, errors.New("wrong number of values")
	}
	return val[0], nil
}

func equal(x1, x2 interface{}) bool {
	if reflect.TypeOf(x1) != reflect.TypeOf(x2) {
		return false
	}
	switch x1 := x1.(type) {
	case float64:
		if math.IsNaN(x1) {
			return math.IsNaN(x2.(float64))
		}
		return x1 == x2
	case time.Time:
		// BigQuery is only accurate to the microsecond.
		return x1.Round(time.Microsecond).Equal(x2.(time.Time).Round(time.Microsecond))
	default:
		return reflect.DeepEqual(x1, x2)
	}
}
