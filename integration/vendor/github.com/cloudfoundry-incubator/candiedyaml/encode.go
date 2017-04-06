/*
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

package candiedyaml

import (
	"bytes"
	"encoding/base64"
	"io"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var (
	timeTimeType  = reflect.TypeOf(time.Time{})
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
	numberType    = reflect.TypeOf(Number(""))
	nonPrintable  = regexp.MustCompile("[^\t\n\r\u0020-\u007E\u0085\u00A0-\uD7FF\uE000-\uFFFD]")
	multiline     = regexp.MustCompile("\n|\u0085|\u2028|\u2029")

	shortTags = map[string]string{
		yaml_NULL_TAG:      "!!null",
		yaml_BOOL_TAG:      "!!bool",
		yaml_STR_TAG:       "!!str",
		yaml_INT_TAG:       "!!int",
		yaml_FLOAT_TAG:     "!!float",
		yaml_TIMESTAMP_TAG: "!!timestamp",
		yaml_SEQ_TAG:       "!!seq",
		yaml_MAP_TAG:       "!!map",
		yaml_BINARY_TAG:    "!!binary",
	}
)

type Marshaler interface {
	MarshalYAML() (tag string, value interface{}, err error)
}

// An Encoder writes JSON objects to an output stream.
type Encoder struct {
	w       io.Writer
	emitter yaml_emitter_t
	event   yaml_event_t
	flow    bool
	err     error
}

func Marshal(v interface{}) ([]byte, error) {
	b := bytes.Buffer{}
	e := NewEncoder(&b)
	err := e.Encode(v)
	return b.Bytes(), err
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{w: w}
	yaml_emitter_initialize(&e.emitter)
	yaml_emitter_set_output_writer(&e.emitter, e.w)
	yaml_stream_start_event_initialize(&e.event, yaml_UTF8_ENCODING)
	e.emit()
	yaml_document_start_event_initialize(&e.event, nil, nil, true)
	e.emit()

	return e
}

func (e *Encoder) Encode(v interface{}) (err error) {
	defer recovery(&err)

	if e.err != nil {
		return e.err
	}

	e.marshal("", reflect.ValueOf(v), true)

	yaml_document_end_event_initialize(&e.event, true)
	e.emit()
	e.emitter.open_ended = false
	yaml_stream_end_event_initialize(&e.event)
	e.emit()

	return nil
}

func (e *Encoder) emit() {
	if !yaml_emitter_emit(&e.emitter, &e.event) {
		panic("bad emit")
	}
}

func (e *Encoder) marshal(tag string, v reflect.Value, allowAddr bool) {
	vt := v.Type()

	if vt.Implements(marshalerType) {
		e.emitMarshaler(tag, v)
		return
	}

	if vt.Kind() != reflect.Ptr && allowAddr {
		if reflect.PtrTo(vt).Implements(marshalerType) {
			e.emitAddrMarshaler(tag, v)
			return
		}
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			e.emitNil()
		} else {
			e.marshal(tag, v.Elem(), allowAddr)
		}
	case reflect.Map:
		e.emitMap(tag, v)
	case reflect.Ptr:
		if v.IsNil() {
			e.emitNil()
		} else {
			e.marshal(tag, v.Elem(), true)
		}
	case reflect.Struct:
		e.emitStruct(tag, v)
	case reflect.Slice:
		e.emitSlice(tag, v)
	case reflect.String:
		e.emitString(tag, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.emitInt(tag, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.emitUint(tag, v)
	case reflect.Float32, reflect.Float64:
		e.emitFloat(tag, v)
	case reflect.Bool:
		e.emitBool(tag, v)
	default:
		panic("Can't marshal type yet: " + v.Type().String())
	}
}

func (e *Encoder) emitMap(tag string, v reflect.Value) {
	e.mapping(tag, func() {
		var keys stringValues = v.MapKeys()
		sort.Sort(keys)
		for _, k := range keys {
			e.marshal("", k, true)
			e.marshal("", v.MapIndex(k), true)
		}
	})
}

func (e *Encoder) emitStruct(tag string, v reflect.Value) {
	if v.Type() == timeTimeType {
		e.emitTime(tag, v)
		return
	}

	fields := cachedTypeFields(v.Type())

	e.mapping(tag, func() {
		for _, f := range fields {
			fv := fieldByIndex(v, f.index)
			if !fv.IsValid() || f.omitEmpty && isEmptyValue(fv) {
				continue
			}

			e.marshal("", reflect.ValueOf(f.name), true)
			e.flow = f.flow
			e.marshal("", fv, true)
		}
	})
}

func (e *Encoder) emitTime(tag string, v reflect.Value) {
	t := v.Interface().(time.Time)
	bytes, _ := t.MarshalText()
	e.emitScalar(string(bytes), "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (e *Encoder) mapping(tag string, f func()) {
	implicit := tag == ""
	style := yaml_BLOCK_MAPPING_STYLE
	if e.flow {
		e.flow = false
		style = yaml_FLOW_MAPPING_STYLE
	}
	yaml_mapping_start_event_initialize(&e.event, nil, []byte(tag), implicit, style)
	e.emit()

	f()

	yaml_mapping_end_event_initialize(&e.event)
	e.emit()
}

func (e *Encoder) emitSlice(tag string, v reflect.Value) {
	if v.Type() == byteSliceType {
		e.emitBase64(tag, v)
		return
	}

	implicit := tag == ""
	style := yaml_BLOCK_SEQUENCE_STYLE
	if e.flow {
		e.flow = false
		style = yaml_FLOW_SEQUENCE_STYLE
	}
	yaml_sequence_start_event_initialize(&e.event, nil, []byte(tag), implicit, style)
	e.emit()

	n := v.Len()
	for i := 0; i < n; i++ {
		e.marshal("", v.Index(i), true)
	}

	yaml_sequence_end_event_initialize(&e.event)
	e.emit()
}

func (e *Encoder) emitBase64(tag string, v reflect.Value) {
	if v.IsNil() {
		e.emitNil()
		return
	}

	s := v.Bytes()

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))

	base64.StdEncoding.Encode(dst, s)
	e.emitScalar(string(dst), "", yaml_BINARY_TAG, yaml_DOUBLE_QUOTED_SCALAR_STYLE)
}

func (e *Encoder) emitString(tag string, v reflect.Value) {
	var style yaml_scalar_style_t
	s := v.String()

	if nonPrintable.MatchString(s) {
		e.emitBase64(tag, v)
		return
	}

	if v.Type() == numberType {
		style = yaml_PLAIN_SCALAR_STYLE
	} else {
		event := yaml_event_t{
			implicit: true,
			value:    []byte(s),
		}

		rtag, _ := resolveInterface(event, false)
		if tag == "" && rtag != yaml_STR_TAG {
			style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
		} else if multiline.MatchString(s) {
			style = yaml_LITERAL_SCALAR_STYLE
		} else {
			style = yaml_PLAIN_SCALAR_STYLE
		}
	}

	e.emitScalar(s, "", tag, style)
}

func (e *Encoder) emitBool(tag string, v reflect.Value) {
	s := strconv.FormatBool(v.Bool())
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitInt(tag string, v reflect.Value) {
	s := strconv.FormatInt(v.Int(), 10)
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitUint(tag string, v reflect.Value) {
	s := strconv.FormatUint(v.Uint(), 10)
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitFloat(tag string, v reflect.Value) {
	f := v.Float()

	var s string
	switch {
	case math.IsNaN(f):
		s = ".nan"
	case math.IsInf(f, 1):
		s = "+.inf"
	case math.IsInf(f, -1):
		s = "-.inf"
	default:
		s = strconv.FormatFloat(f, 'g', -1, v.Type().Bits())
	}

	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitNil() {
	e.emitScalar("null", "", "", yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitScalar(value, anchor, tag string, style yaml_scalar_style_t) {
	implicit := tag == ""
	if !implicit {
		style = yaml_PLAIN_SCALAR_STYLE
	}

	stag := shortTags[tag]
	if stag == "" {
		stag = tag
	}

	yaml_scalar_event_initialize(&e.event, []byte(anchor), []byte(stag), []byte(value), implicit, implicit, style)
	e.emit()
}

func (e *Encoder) emitMarshaler(tag string, v reflect.Value) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		e.emitNil()
		return
	}

	m := v.Interface().(Marshaler)
	if m == nil {
		e.emitNil()
		return
	}
	t, val, err := m.MarshalYAML()
	if err != nil {
		panic(err)
	}
	if val == nil {
		e.emitNil()
		return
	}

	e.marshal(t, reflect.ValueOf(val), false)
}

func (e *Encoder) emitAddrMarshaler(tag string, v reflect.Value) {
	if !v.CanAddr() {
		e.marshal(tag, v, false)
		return
	}

	va := v.Addr()
	if va.IsNil() {
		e.emitNil()
		return
	}

	m := v.Interface().(Marshaler)
	t, val, err := m.MarshalYAML()
	if err != nil {
		panic(err)
	}

	if val == nil {
		e.emitNil()
		return
	}

	e.marshal(t, reflect.ValueOf(val), false)
}
