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
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

type Unmarshaler interface {
	UnmarshalYAML(tag string, value interface{}) error
}

// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

type Decoder struct {
	parser        yaml_parser_t
	event         yaml_event_t
	replay_events []yaml_event_t
	useNumber     bool

	anchors          map[string][]yaml_event_t
	tracking_anchors [][]yaml_event_t
}

type ParserError struct {
	ErrorType   YAML_error_type_t
	Context     string
	ContextMark YAML_mark_t
	Problem     string
	ProblemMark YAML_mark_t
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("yaml: [%s] %s at line %d, column %d", e.Context, e.Problem, e.ProblemMark.line+1, e.ProblemMark.column+1)
}

type UnexpectedEventError struct {
	Value     string
	EventType yaml_event_type_t
	At        YAML_mark_t
}

func (e *UnexpectedEventError) Error() string {
	return fmt.Sprintf("yaml: Unexpect event [%d]: '%s' at line %d, column %d", e.EventType, e.Value, e.At.line+1, e.At.column+1)
}

func recovery(err *error) {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		}

		var tmpError error
		switch r := r.(type) {
		case error:
			tmpError = r
		case string:
			tmpError = errors.New(r)
		default:
			tmpError = errors.New("Unknown panic: " + reflect.ValueOf(r).String())
		}

		*err = tmpError
	}
}

func Unmarshal(data []byte, v interface{}) error {
	d := NewDecoder(bytes.NewBuffer(data))
	return d.Decode(v)
}

func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{
		anchors:          make(map[string][]yaml_event_t),
		tracking_anchors: make([][]yaml_event_t, 1),
	}
	yaml_parser_initialize(&d.parser)
	yaml_parser_set_input_reader(&d.parser, r)
	return d
}

func (d *Decoder) Decode(v interface{}) (err error) {
	defer recovery(&err)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("Expected a pointer or nil but was a %s at %s", rv.String(), d.event.start_mark)
	}

	if d.event.event_type == yaml_NO_EVENT {
		d.nextEvent()

		if d.event.event_type != yaml_STREAM_START_EVENT {
			return errors.New("Invalid stream")
		}

		d.nextEvent()
	}

	d.document(rv)
	return nil
}

func (d *Decoder) UseNumber() { d.useNumber = true }

func (d *Decoder) error(err error) {
	panic(err)
}

func (d *Decoder) nextEvent() {
	if d.event.event_type == yaml_STREAM_END_EVENT {
		d.error(errors.New("The stream is closed"))
	}

	if d.replay_events != nil {
		d.event = d.replay_events[0]
		if len(d.replay_events) == 1 {
			d.replay_events = nil
		} else {
			d.replay_events = d.replay_events[1:]
		}
	} else {
		if !yaml_parser_parse(&d.parser, &d.event) {
			yaml_event_delete(&d.event)

			d.error(&ParserError{
				ErrorType:   d.parser.error,
				Context:     d.parser.context,
				ContextMark: d.parser.context_mark,
				Problem:     d.parser.problem,
				ProblemMark: d.parser.problem_mark,
			})
		}
	}

	last := len(d.tracking_anchors)
	// skip aliases when tracking an anchor
	if last > 0 && d.event.event_type != yaml_ALIAS_EVENT {
		d.tracking_anchors[last-1] = append(d.tracking_anchors[last-1], d.event)
	}
}

func (d *Decoder) document(rv reflect.Value) {
	if d.event.event_type != yaml_DOCUMENT_START_EVENT {
		d.error(fmt.Errorf("Expected document start at %s", d.event.start_mark))
	}

	d.nextEvent()
	d.parse(rv)

	if d.event.event_type != yaml_DOCUMENT_END_EVENT {
		d.error(fmt.Errorf("Expected document end at %s", d.event.start_mark))
	}

	d.nextEvent()
}

func (d *Decoder) parse(rv reflect.Value) {
	if !rv.IsValid() {
		// skip ahead since we cannot store
		d.valueInterface()
		return
	}

	anchor := string(d.event.anchor)
	switch d.event.event_type {
	case yaml_SEQUENCE_START_EVENT:
		d.begin_anchor(anchor)
		d.sequence(rv)
		d.end_anchor(anchor)
	case yaml_MAPPING_START_EVENT:
		d.begin_anchor(anchor)
		d.mapping(rv)
		d.end_anchor(anchor)
	case yaml_SCALAR_EVENT:
		d.begin_anchor(anchor)
		d.scalar(rv)
		d.end_anchor(anchor)
	case yaml_ALIAS_EVENT:
		d.alias(rv)
	case yaml_DOCUMENT_END_EVENT:
	default:
		d.error(&UnexpectedEventError{
			Value:     string(d.event.value),
			EventType: d.event.event_type,
			At:        d.event.start_mark,
		})
	}
}

func (d *Decoder) begin_anchor(anchor string) {
	if anchor != "" {
		events := []yaml_event_t{d.event}
		d.tracking_anchors = append(d.tracking_anchors, events)
	}
}

func (d *Decoder) end_anchor(anchor string) {
	if anchor != "" {
		events := d.tracking_anchors[len(d.tracking_anchors)-1]
		d.tracking_anchors = d.tracking_anchors[0 : len(d.tracking_anchors)-1]
		// remove the anchor, replaying events shouldn't have anchors
		events[0].anchor = nil
		// we went one too many, remove the extra event
		events = events[:len(events)-1]
		// if nested, append to all the other anchors
		for i, e := range d.tracking_anchors {
			d.tracking_anchors[i] = append(e, events...)
		}
		d.anchors[anchor] = events
	}
}

func (d *Decoder) indirect(v reflect.Value, decodingNull bool) (Unmarshaler, reflect.Value) {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(Unmarshaler); ok {
				var temp interface{}
				return u, reflect.ValueOf(&temp)
			}
		}

		v = v.Elem()
	}

	return nil, v
}

func (d *Decoder) sequence(v reflect.Value) {
	if d.event.event_type != yaml_SEQUENCE_START_EVENT {
		d.error(fmt.Errorf("Expected sequence start at %s", d.event.start_mark))
	}

	u, pv := d.indirect(v, false)
	if u != nil {
		defer func() {
			if err := u.UnmarshalYAML(yaml_SEQ_TAG, pv.Interface()); err != nil {
				d.error(err)
			}
		}()
		_, pv = d.indirect(pv, false)
	}

	v = pv

	// Check type of target.
	switch v.Kind() {
	case reflect.Interface:
		if v.NumMethod() == 0 {
			// Decoding into nil interface?  Switch to non-reflect code.
			v.Set(reflect.ValueOf(d.sequenceInterface()))
			return
		}
		// Otherwise it's invalid.
		fallthrough
	default:
		d.error(fmt.Errorf("Expected an array, slice or interface{} but was a %s at %s", v, d.event.start_mark))
	case reflect.Array:
	case reflect.Slice:
		break
	}

	d.nextEvent()

	i := 0
done:
	for {
		switch d.event.event_type {
		case yaml_SEQUENCE_END_EVENT, yaml_DOCUMENT_END_EVENT:
			break done
		}

		// Get element of array, growing if necessary.
		if v.Kind() == reflect.Slice {
			// Grow slice if necessary
			if i >= v.Cap() {
				newcap := v.Cap() + v.Cap()/2
				if newcap < 4 {
					newcap = 4
				}
				newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
				reflect.Copy(newv, v)
				v.Set(newv)
			}
			if i >= v.Len() {
				v.SetLen(i + 1)
			}
		}

		if i < v.Len() {
			// Decode into element.
			d.parse(v.Index(i))
		} else {
			// Ran out of fixed array: skip.
			d.parse(reflect.Value{})
		}
		i++
	}

	if i < v.Len() {
		if v.Kind() == reflect.Array {
			// Array.  Zero the rest.
			z := reflect.Zero(v.Type().Elem())
			for ; i < v.Len(); i++ {
				v.Index(i).Set(z)
			}
		} else {
			v.SetLen(i)
		}
	}
	if i == 0 && v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	if d.event.event_type != yaml_DOCUMENT_END_EVENT {
		d.nextEvent()
	}
}

func (d *Decoder) mapping(v reflect.Value) {
	u, pv := d.indirect(v, false)
	if u != nil {
		defer func() {
			if err := u.UnmarshalYAML(yaml_MAP_TAG, pv.Interface()); err != nil {
				d.error(err)
			}
		}()
		_, pv = d.indirect(pv, false)
	}
	v = pv

	// Decoding into nil interface?  Switch to non-reflect code.
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		v.Set(reflect.ValueOf(d.mappingInterface()))
		return
	}

	// Check type of target: struct or map[X]Y
	switch v.Kind() {
	case reflect.Struct:
		d.mappingStruct(v)
		return
	case reflect.Map:
	default:
		d.error(fmt.Errorf("Expected a struct or map but was a %s at %s ", v, d.event.start_mark))
	}

	mapt := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(mapt))
	}

	d.nextEvent()

	keyt := mapt.Key()
	mapElemt := mapt.Elem()

	var mapElem reflect.Value
done:
	for {
		switch d.event.event_type {
		case yaml_MAPPING_END_EVENT:
			break done
		case yaml_DOCUMENT_END_EVENT:
			return
		}

		key := reflect.New(keyt)
		d.parse(key.Elem())

		if !mapElem.IsValid() {
			mapElem = reflect.New(mapElemt).Elem()
		} else {
			mapElem.Set(reflect.Zero(mapElemt))
		}

		d.parse(mapElem)

		v.SetMapIndex(key.Elem(), mapElem)
	}

	d.nextEvent()
}

func (d *Decoder) mappingStruct(v reflect.Value) {

	structt := v.Type()
	fields := cachedTypeFields(structt)

	d.nextEvent()

done:
	for {
		switch d.event.event_type {
		case yaml_MAPPING_END_EVENT:
			break done
		case yaml_DOCUMENT_END_EVENT:
			return
		}

		key := ""
		d.parse(reflect.ValueOf(&key))

		// Figure out field corresponding to key.
		var subv reflect.Value

		var f *field
		for i := range fields {
			ff := &fields[i]
			if ff.name == key {
				f = ff
				break
			}

			if f == nil && strings.EqualFold(ff.name, key) {
				f = ff
			}
		}

		if f != nil {
			subv = v
			for _, i := range f.index {
				if subv.Kind() == reflect.Ptr {
					if subv.IsNil() {
						subv.Set(reflect.New(subv.Type().Elem()))
					}
					subv = subv.Elem()
				}
				subv = subv.Field(i)
			}
		}
		d.parse(subv)
	}

	d.nextEvent()
}

func (d *Decoder) scalar(v reflect.Value) {
	val := string(d.event.value)
	wantptr := null_values[val]

	u, pv := d.indirect(v, wantptr)

	var tag string
	if u != nil {
		defer func() {
			if err := u.UnmarshalYAML(tag, pv.Interface()); err != nil {
				d.error(err)
			}
		}()

		_, pv = d.indirect(pv, wantptr)
	}
	v = pv

	var err error
	tag, err = resolve(d.event, v, d.useNumber)
	if err != nil {
		d.error(err)
	}

	d.nextEvent()
}

func (d *Decoder) alias(rv reflect.Value) {
	val, ok := d.anchors[string(d.event.anchor)]
	if !ok {
		d.error(fmt.Errorf("missing anchor: '%s' at %s", d.event.anchor, d.event.start_mark))
	}

	d.replay_events = val
	d.nextEvent()
	d.parse(rv)
}

func (d *Decoder) valueInterface() interface{} {
	var v interface{}

	anchor := string(d.event.anchor)
	switch d.event.event_type {
	case yaml_SEQUENCE_START_EVENT:
		d.begin_anchor(anchor)
		v = d.sequenceInterface()
	case yaml_MAPPING_START_EVENT:
		d.begin_anchor(anchor)
		v = d.mappingInterface()
	case yaml_SCALAR_EVENT:
		d.begin_anchor(anchor)
		v = d.scalarInterface()
	case yaml_ALIAS_EVENT:
		rv := reflect.ValueOf(&v)
		d.alias(rv)
		return v
	case yaml_DOCUMENT_END_EVENT:
		d.error(&UnexpectedEventError{
			Value:     string(d.event.value),
			EventType: d.event.event_type,
			At:        d.event.start_mark,
		})

	}
	d.end_anchor(anchor)

	return v
}

func (d *Decoder) scalarInterface() interface{} {
	_, v := resolveInterface(d.event, d.useNumber)

	d.nextEvent()
	return v
}

// sequenceInterface is like sequence but returns []interface{}.
func (d *Decoder) sequenceInterface() []interface{} {
	var v = make([]interface{}, 0)

	d.nextEvent()

done:
	for {
		switch d.event.event_type {
		case yaml_SEQUENCE_END_EVENT, yaml_DOCUMENT_END_EVENT:
			break done
		}

		v = append(v, d.valueInterface())
	}

	if d.event.event_type != yaml_DOCUMENT_END_EVENT {
		d.nextEvent()
	}

	return v
}

// mappingInterface is like mapping but returns map[interface{}]interface{}.
func (d *Decoder) mappingInterface() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})

	d.nextEvent()

done:
	for {
		switch d.event.event_type {
		case yaml_MAPPING_END_EVENT, yaml_DOCUMENT_END_EVENT:
			break done
		}

		key := d.valueInterface()

		// Read value.
		m[key] = d.valueInterface()
	}

	if d.event.event_type != yaml_DOCUMENT_END_EVENT {
		d.nextEvent()
	}

	return m
}
