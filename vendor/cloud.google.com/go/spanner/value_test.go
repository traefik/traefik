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
	"math"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/golang/protobuf/proto"
	proto3 "github.com/golang/protobuf/ptypes/struct"
	sppb "google.golang.org/genproto/googleapis/spanner/v1"
)

var (
	t1 = mustParseTime("2016-11-15T15:04:05.999999999Z")
	// Boundaries
	t2 = mustParseTime("0000-01-01T00:00:00.000000000Z")
	t3 = mustParseTime("9999-12-31T23:59:59.999999999Z")
	// Local timezone
	t4 = time.Now()
	d1 = mustParseDate("2016-11-15")
	d2 = mustParseDate("1678-01-01")
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustParseDate(s string) civil.Date {
	d, err := civil.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Test encoding Values.
func TestEncodeValue(t *testing.T) {
	var (
		tString = stringType()
		tInt    = intType()
		tBool   = boolType()
		tFloat  = floatType()
		tBytes  = bytesType()
		tTime   = timeType()
		tDate   = dateType()
	)
	for i, test := range []struct {
		in       interface{}
		want     *proto3.Value
		wantType *sppb.Type
	}{
		// STRING / STRING ARRAY
		{"abc", stringProto("abc"), tString},
		{NullString{"abc", true}, stringProto("abc"), tString},
		{NullString{"abc", false}, nullProto(), nil},
		{[]string{"abc", "bcd"}, listProto(stringProto("abc"), stringProto("bcd")), listType(tString)},
		{[]NullString{{"abcd", true}, {"xyz", false}}, listProto(stringProto("abcd"), nullProto()), listType(tString)},
		// BYTES / BYTES ARRAY
		{[]byte("foo"), bytesProto([]byte("foo")), tBytes},
		{[]byte(nil), nullProto(), nil},
		{[][]byte{nil, []byte("ab")}, listProto(nullProto(), bytesProto([]byte("ab"))), listType(tBytes)},
		{[][]byte(nil), nullProto(), nil},
		// INT64 / INT64 ARRAY
		{7, intProto(7), tInt},
		{[]int{31, 127}, listProto(intProto(31), intProto(127)), listType(tInt)},
		{int64(81), intProto(81), tInt},
		{[]int64{33, 129}, listProto(intProto(33), intProto(129)), listType(tInt)},
		{NullInt64{11, true}, intProto(11), tInt},
		{NullInt64{11, false}, nullProto(), nil},
		{[]NullInt64{{35, true}, {131, false}}, listProto(intProto(35), nullProto()), listType(tInt)},
		// BOOL / BOOL ARRAY
		{true, boolProto(true), tBool},
		{NullBool{true, true}, boolProto(true), tBool},
		{NullBool{true, false}, nullProto(), nil},
		{[]bool{true, false}, listProto(boolProto(true), boolProto(false)), listType(tBool)},
		{[]NullBool{{true, true}, {true, false}}, listProto(boolProto(true), nullProto()), listType(tBool)},
		// FLOAT64 / FLOAT64 ARRAY
		{3.14, floatProto(3.14), tFloat},
		{NullFloat64{3.1415, true}, floatProto(3.1415), tFloat},
		{NullFloat64{math.Inf(1), true}, floatProto(math.Inf(1)), tFloat},
		{NullFloat64{3.14159, false}, nullProto(), nil},
		{[]float64{3.141, 0.618, math.Inf(-1)}, listProto(floatProto(3.141), floatProto(0.618), floatProto(math.Inf(-1))), listType(tFloat)},
		{[]NullFloat64{{3.141, true}, {0.618, false}}, listProto(floatProto(3.141), nullProto()), listType(tFloat)},
		// TIMESTAMP / TIMESTAMP ARRAY
		{t1, timeProto(t1), tTime},
		{NullTime{t1, true}, timeProto(t1), tTime},
		{NullTime{t1, false}, nullProto(), nil},
		{[]time.Time{t1, t2, t3, t4}, listProto(timeProto(t1), timeProto(t2), timeProto(t3), timeProto(t4)), listType(tTime)},
		{[]NullTime{{t1, true}, {t1, false}}, listProto(timeProto(t1), nullProto()), listType(tTime)},
		// DATE / DATE ARRAY
		{d1, dateProto(d1), tDate},
		{NullDate{d1, true}, dateProto(d1), tDate},
		{NullDate{civil.Date{}, false}, nullProto(), nil},
		{[]civil.Date{d1, d2}, listProto(dateProto(d1), dateProto(d2)), listType(tDate)},
		{[]NullDate{{d1, true}, {civil.Date{}, false}}, listProto(dateProto(d1), nullProto()), listType(tDate)},
		// GenericColumnValue
		{GenericColumnValue{tString, stringProto("abc")}, stringProto("abc"), tString},
		{GenericColumnValue{tString, nullProto()}, nullProto(), tString},
		// not actually valid (stringProto inside int list), but demonstrates pass-through.
		{
			GenericColumnValue{
				Type:  listType(tInt),
				Value: listProto(intProto(5), nullProto(), stringProto("bcd")),
			},
			listProto(intProto(5), nullProto(), stringProto("bcd")),
			listType(tInt),
		},
	} {
		got, gotType, err := encodeValue(test.in)
		if err != nil {
			t.Fatalf("#%d: got error during encoding: %v, want nil", i, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("#%d: got encode result: %v, want %v", i, got, test.want)
		}
		if !reflect.DeepEqual(gotType, test.wantType) {
			t.Errorf("#%d: got encode type: %v, want %v", i, gotType, test.wantType)
		}
	}
}

// Test decoding Values.
func TestDecodeValue(t *testing.T) {
	for i, test := range []struct {
		in   *proto3.Value
		t    *sppb.Type
		want interface{}
		fail bool
	}{
		// STRING
		{stringProto("abc"), stringType(), "abc", false},
		{nullProto(), stringType(), "abc", true},
		{stringProto("abc"), stringType(), NullString{"abc", true}, false},
		{nullProto(), stringType(), NullString{}, false},
		// STRING ARRAY
		{
			listProto(stringProto("abc"), nullProto(), stringProto("bcd")),
			listType(stringType()),
			[]NullString{{"abc", true}, {}, {"bcd", true}},
			false,
		},
		{nullProto(), listType(stringType()), []NullString(nil), false},
		// BYTES
		{bytesProto([]byte("ab")), bytesType(), []byte("ab"), false},
		{nullProto(), bytesType(), []byte(nil), false},
		// BYTES ARRAY
		{listProto(bytesProto([]byte("ab")), nullProto()), listType(bytesType()), [][]byte{[]byte("ab"), nil}, false},
		{nullProto(), listType(bytesType()), [][]byte(nil), false},
		//INT64
		{intProto(15), intType(), int64(15), false},
		{nullProto(), intType(), int64(0), true},
		{intProto(15), intType(), NullInt64{15, true}, false},
		{nullProto(), intType(), NullInt64{}, false},
		// INT64 ARRAY
		{listProto(intProto(91), nullProto(), intProto(87)), listType(intType()), []NullInt64{{91, true}, {}, {87, true}}, false},
		{nullProto(), listType(intType()), []NullInt64(nil), false},
		// BOOL
		{boolProto(true), boolType(), true, false},
		{nullProto(), boolType(), true, true},
		{boolProto(true), boolType(), NullBool{true, true}, false},
		{nullProto(), boolType(), NullBool{}, false},
		// BOOL ARRAY
		{listProto(boolProto(true), boolProto(false), nullProto()), listType(boolType()), []NullBool{{true, true}, {false, true}, {}}, false},
		{nullProto(), listType(boolType()), []NullBool(nil), false},
		// FLOAT64
		{floatProto(3.14), floatType(), 3.14, false},
		{nullProto(), floatType(), 0.00, true},
		{floatProto(3.14), floatType(), NullFloat64{3.14, true}, false},
		{nullProto(), floatType(), NullFloat64{}, false},
		// FLOAT64 ARRAY
		{
			listProto(floatProto(math.Inf(1)), floatProto(math.Inf(-1)), nullProto(), floatProto(3.1)),
			listType(floatType()),
			[]NullFloat64{{math.Inf(1), true}, {math.Inf(-1), true}, {}, {3.1, true}},
			false,
		},
		{nullProto(), listType(floatType()), []NullFloat64(nil), false},
		// TIMESTAMP
		{timeProto(t1), timeType(), t1, false},
		{timeProto(t1), timeType(), NullTime{t1, true}, false},
		{nullProto(), timeType(), NullTime{}, false},
		// TIMESTAMP ARRAY
		{listProto(timeProto(t1), timeProto(t2), timeProto(t3), nullProto()), listType(timeType()), []NullTime{{t1, true}, {t2, true}, {t3, true}, {}}, false},
		{nullProto(), listType(timeType()), []NullTime(nil), false},
		// DATE
		{dateProto(d1), dateType(), d1, false},
		{dateProto(d1), dateType(), NullDate{d1, true}, false},
		{nullProto(), dateType(), NullDate{}, false},
		// DATE ARRAY
		{listProto(dateProto(d1), dateProto(d2), nullProto()), listType(dateType()), []NullDate{{d1, true}, {d2, true}, {}}, false},
		{nullProto(), listType(dateType()), []NullDate(nil), false},
		// STRUCT ARRAY
		// STRUCT schema is equal to the following Go struct:
		// type s struct {
		//     Col1 NullInt64
		//     Col2 []struct {
		//         SubCol1 float64
		//         SubCol2 string
		//     }
		// }
		{
			in: listProto(
				listProto(
					intProto(3),
					listProto(
						listProto(floatProto(3.14), stringProto("this")),
						listProto(floatProto(0.57), stringProto("siht")),
					),
				),
				listProto(
					nullProto(),
					nullProto(),
				),
				nullProto(),
			),
			t: listType(
				structType(
					mkField("Col1", intType()),
					mkField(
						"Col2",
						listType(
							structType(
								mkField("SubCol1", floatType()),
								mkField("SubCol2", stringType()),
							),
						),
					),
				),
			),
			want: []NullRow{
				{
					Row: Row{
						fields: []*sppb.StructType_Field{
							mkField("Col1", intType()),
							mkField(
								"Col2",
								listType(
									structType(
										mkField("SubCol1", floatType()),
										mkField("SubCol2", stringType()),
									),
								),
							),
						},
						vals: []*proto3.Value{
							intProto(3),
							listProto(
								listProto(floatProto(3.14), stringProto("this")),
								listProto(floatProto(0.57), stringProto("siht")),
							),
						},
					},
					Valid: true,
				},
				{
					Row: Row{
						fields: []*sppb.StructType_Field{
							mkField("Col1", intType()),
							mkField(
								"Col2",
								listType(
									structType(
										mkField("SubCol1", floatType()),
										mkField("SubCol2", stringType()),
									),
								),
							),
						},
						vals: []*proto3.Value{
							nullProto(),
							nullProto(),
						},
					},
					Valid: true,
				},
				{},
			},
			fail: false,
		},
		{
			in: listProto(
				listProto(
					intProto(3),
					listProto(
						listProto(floatProto(3.14), stringProto("this")),
						listProto(floatProto(0.57), stringProto("siht")),
					),
				),
				listProto(
					nullProto(),
					nullProto(),
				),
				nullProto(),
			),
			t: listType(
				structType(
					mkField("Col1", intType()),
					mkField(
						"Col2",
						listType(
							structType(
								mkField("SubCol1", floatType()),
								mkField("SubCol2", stringType()),
							),
						),
					),
				),
			),
			want: []*struct {
				Col1      NullInt64
				StructCol []*struct {
					SubCol1 NullFloat64
					SubCol2 string
				} `spanner:"Col2"`
			}{
				{
					Col1: NullInt64{3, true},
					StructCol: []*struct {
						SubCol1 NullFloat64
						SubCol2 string
					}{
						{
							SubCol1: NullFloat64{3.14, true},
							SubCol2: "this",
						},
						{
							SubCol1: NullFloat64{0.57, true},
							SubCol2: "siht",
						},
					},
				},
				{
					Col1: NullInt64{},
					StructCol: []*struct {
						SubCol1 NullFloat64
						SubCol2 string
					}(nil),
				},
				nil,
			},
			fail: false,
		},
		// GenericColumnValue
		{stringProto("abc"), stringType(), GenericColumnValue{stringType(), stringProto("abc")}, false},
		{nullProto(), stringType(), GenericColumnValue{stringType(), nullProto()}, false},
		// not actually valid (stringProto inside int list), but demonstrates pass-through.
		{
			in: listProto(intProto(5), nullProto(), stringProto("bcd")),
			t:  listType(intType()),
			want: GenericColumnValue{
				Type:  listType(intType()),
				Value: listProto(intProto(5), nullProto(), stringProto("bcd")),
			},
			fail: false,
		},
	} {
		gotp := reflect.New(reflect.TypeOf(test.want))
		if err := decodeValue(test.in, test.t, gotp.Interface()); err != nil {
			if !test.fail {
				t.Errorf("%d: cannot decode %v(%v): %v", i, test.in, test.t, err)
			}
			continue
		}
		if test.fail {
			t.Errorf("%d: decoding %v(%v) succeeds unexpectedly, want error", i, test.in, test.t)
			continue
		}
		got := reflect.Indirect(gotp).Interface()
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: unexpected decoding result - got %v, want %v", i, got, test.want)
			continue
		}
	}
}

// Test error cases for decodeValue.
func TestDecodeValueErrors(t *testing.T) {
	for i, test := range []struct {
		in *proto3.Value
		t  *sppb.Type
		v  interface{}
	}{
		{nullProto(), stringType(), nil},
		{nullProto(), stringType(), 1},
	} {
		err := decodeValue(test.in, test.t, test.v)
		if err == nil {
			t.Errorf("#%d: want error, got nil", i)
		}
	}
}

// Test NaN encoding/decoding.
func TestNaN(t *testing.T) {
	// Decode NaN value.
	f := 0.0
	nf := NullFloat64{}
	// To float64
	if err := decodeValue(floatProto(math.NaN()), floatType(), &f); err != nil {
		t.Errorf("decodeValue returns %q for %v, want nil", err, floatProto(math.NaN()))
	}
	if !math.IsNaN(f) {
		t.Errorf("f = %v, want %v", f, math.NaN())
	}
	// To NullFloat64
	if err := decodeValue(floatProto(math.NaN()), floatType(), &nf); err != nil {
		t.Errorf("decodeValue returns %q for %v, want nil", err, floatProto(math.NaN()))
	}
	if !math.IsNaN(nf.Float64) || !nf.Valid {
		t.Errorf("f = %v, want %v", f, NullFloat64{math.NaN(), true})
	}
	// Encode NaN value
	// From float64
	v, _, err := encodeValue(math.NaN())
	if err != nil {
		t.Errorf("encodeValue returns %q for NaN, want nil", err)
	}
	x, ok := v.GetKind().(*proto3.Value_NumberValue)
	if !ok {
		t.Errorf("incorrect type for v.GetKind(): %T, want *proto3.Value_NumberValue", v.GetKind())
	}
	if !math.IsNaN(x.NumberValue) {
		t.Errorf("x.NumberValue = %v, want %v", x.NumberValue, math.NaN())
	}
	// From NullFloat64
	v, _, err = encodeValue(NullFloat64{math.NaN(), true})
	if err != nil {
		t.Errorf("encodeValue returns %q for NaN, want nil", err)
	}
	x, ok = v.GetKind().(*proto3.Value_NumberValue)
	if !ok {
		t.Errorf("incorrect type for v.GetKind(): %T, want *proto3.Value_NumberValue", v.GetKind())
	}
	if !math.IsNaN(x.NumberValue) {
		t.Errorf("x.NumberValue = %v, want %v", x.NumberValue, math.NaN())
	}
}

func TestGenericColumnValue(t *testing.T) {
	for _, test := range []struct {
		in   GenericColumnValue
		want interface{}
		fail bool
	}{
		{GenericColumnValue{stringType(), stringProto("abc")}, "abc", false},
		{GenericColumnValue{stringType(), stringProto("abc")}, 5, true},
		{GenericColumnValue{listType(intType()), listProto(intProto(91), nullProto(), intProto(87))}, []NullInt64{{91, true}, {}, {87, true}}, false},
		{GenericColumnValue{intType(), intProto(42)}, GenericColumnValue{intType(), intProto(42)}, false}, // trippy! :-)
	} {
		// We take a copy and mutate because we're paranoid about immutability.
		inCopy := GenericColumnValue{
			Type:  proto.Clone(test.in.Type).(*sppb.Type),
			Value: proto.Clone(test.in.Value).(*proto3.Value),
		}
		gotp := reflect.New(reflect.TypeOf(test.want))
		if err := inCopy.Decode(gotp.Interface()); err != nil {
			if !test.fail {
				t.Errorf("cannot decode %v to %v: %v", test.in, test.want, err)
			}
			continue
		}
		if test.fail {
			t.Errorf("decoding %v to %v succeeds unexpectedly", test.in, test.want)
		}
		// mutations to inCopy should be invisible to gotp.
		inCopy.Type.Code = sppb.TypeCode_TIMESTAMP
		inCopy.Value.Kind = &proto3.Value_NumberValue{NumberValue: 999}
		got := reflect.Indirect(gotp).Interface()
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("unexpected decode result - got %v, want %v", got, test.want)
		}

		// Test we can go backwards as well.
		v, err := NewGenericColumnValue(test.want)
		if err != nil {
			t.Errorf("NewGenericColumnValue failed: %v", err)
			continue
		}
		if !reflect.DeepEqual(*v, test.in) {
			t.Errorf("unexpected encode result - got %v, want %v", v, test.in)
		}
		// If want is a GenericColumnValue, mutate its underlying value to validate
		// we have taken a deep copy.
		if gcv, ok := test.want.(GenericColumnValue); ok {
			gcv.Type.Code = sppb.TypeCode_TIMESTAMP
			gcv.Value.Kind = &proto3.Value_NumberValue{NumberValue: 999}
			if !reflect.DeepEqual(*v, test.in) {
				t.Errorf("expected deep copy - got %v, want %v", v, test.in)
			}
		}
	}
}

func runBench(b *testing.B, size int, f func(a []int) (*proto3.Value, *sppb.Type, error)) {
	a := make([]int, size)
	for i := 0; i < b.N; i++ {
		f(a)
	}
}

func BenchmarkEncodeIntArrayOrig1(b *testing.B) {
	runBench(b, 1, encodeIntArrayOrig)
}

func BenchmarkEncodeIntArrayOrig10(b *testing.B) {
	runBench(b, 10, encodeIntArrayOrig)
}

func BenchmarkEncodeIntArrayOrig100(b *testing.B) {
	runBench(b, 100, encodeIntArrayOrig)
}

func BenchmarkEncodeIntArrayOrig1000(b *testing.B) {
	runBench(b, 1000, encodeIntArrayOrig)
}

func BenchmarkEncodeIntArrayFunc1(b *testing.B) {
	runBench(b, 1, encodeIntArrayFunc)
}

func BenchmarkEncodeIntArrayFunc10(b *testing.B) {
	runBench(b, 10, encodeIntArrayFunc)
}

func BenchmarkEncodeIntArrayFunc100(b *testing.B) {
	runBench(b, 100, encodeIntArrayFunc)
}

func BenchmarkEncodeIntArrayFunc1000(b *testing.B) {
	runBench(b, 1000, encodeIntArrayFunc)
}

func BenchmarkEncodeIntArrayReflect1(b *testing.B) {
	runBench(b, 1, encodeIntArrayReflect)
}

func BenchmarkEncodeIntArrayReflect10(b *testing.B) {
	runBench(b, 10, encodeIntArrayReflect)
}

func BenchmarkEncodeIntArrayReflect100(b *testing.B) {
	runBench(b, 100, encodeIntArrayReflect)
}

func BenchmarkEncodeIntArrayReflect1000(b *testing.B) {
	runBench(b, 1000, encodeIntArrayReflect)
}

func encodeIntArrayOrig(a []int) (*proto3.Value, *sppb.Type, error) {
	vs := make([]*proto3.Value, len(a))
	var err error
	for i := range a {
		vs[i], _, err = encodeValue(a[i])
		if err != nil {
			return nil, nil, err
		}
	}
	return listProto(vs...), listType(intType()), nil
}

func encodeIntArrayFunc(a []int) (*proto3.Value, *sppb.Type, error) {
	v, err := encodeArray(len(a), func(i int) interface{} { return a[i] })
	if err != nil {
		return nil, nil, err
	}
	return v, listType(intType()), nil
}

func encodeIntArrayReflect(a []int) (*proto3.Value, *sppb.Type, error) {
	v, err := encodeArrayReflect(a)
	if err != nil {
		return nil, nil, err
	}
	return v, listType(intType()), nil
}

func encodeArrayReflect(a interface{}) (*proto3.Value, error) {
	va := reflect.ValueOf(a)
	len := va.Len()
	vs := make([]*proto3.Value, len)
	var err error
	for i := 0; i < len; i++ {
		vs[i], _, err = encodeValue(va.Index(i).Interface())
		if err != nil {
			return nil, err
		}
	}
	return listProto(vs...), nil
}
