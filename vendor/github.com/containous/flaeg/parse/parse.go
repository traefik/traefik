package parse

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Parser is an interface that allows the contents of a flag.Getter to be set.
type Parser interface {
	flag.Getter
	SetValue(interface{})
}

// BoolValue bool Value type
type BoolValue bool

// Set sets bool value from the given string value.
func (b *BoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = BoolValue(v)
	return err
}

// Get returns the bool value.
func (b *BoolValue) Get() interface{} { return bool(*b) }

func (b *BoolValue) String() string { return fmt.Sprintf("%v", *b) }

// IsBoolFlag return true
func (b *BoolValue) IsBoolFlag() bool { return true }

// SetValue sets the duration from the given bool-asserted value.
func (b *BoolValue) SetValue(val interface{}) {
	*b = BoolValue(val.(bool))
}

// BoolFlag optional interface to indicate boolean flags that can be
// supplied without "=value" text
type BoolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

// IntValue int Value
type IntValue int

// Set sets int value from the given string value.
func (i *IntValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = IntValue(v)
	return err
}

// Get returns the int value.
func (i *IntValue) Get() interface{} { return int(*i) }

func (i *IntValue) String() string { return fmt.Sprintf("%v", *i) }

// SetValue sets the IntValue from the given int-asserted value.
func (i *IntValue) SetValue(val interface{}) {
	*i = IntValue(val.(int))
}

// Int64Value int64 Value
type Int64Value int64

// Set sets int64 value from the given string value.
func (i *Int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = Int64Value(v)
	return err
}

// Get returns the int64 value.
func (i *Int64Value) Get() interface{} { return int64(*i) }

func (i *Int64Value) String() string { return fmt.Sprintf("%v", *i) }

// SetValue sets the Int64Value from the given int64-asserted value.
func (i *Int64Value) SetValue(val interface{}) {
	*i = Int64Value(val.(int64))
}

// UintValue uint Value
type UintValue uint

// Set sets uint value from the given string value.
func (i *UintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = UintValue(v)
	return err
}

// Get returns the uint value.
func (i *UintValue) Get() interface{} { return uint(*i) }

func (i *UintValue) String() string { return fmt.Sprintf("%v", *i) }

// SetValue sets the UintValue from the given uint-asserted value.
func (i *UintValue) SetValue(val interface{}) {
	*i = UintValue(val.(uint))
}

// Uint64Value uint64 Value
type Uint64Value uint64

// Set sets uint64 value from the given string value.
func (i *Uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = Uint64Value(v)
	return err
}

// Get returns the uint64 value.
func (i *Uint64Value) Get() interface{} { return uint64(*i) }

func (i *Uint64Value) String() string { return fmt.Sprintf("%v", *i) }

// SetValue sets the Uint64Value from the given uint64-asserted value.
func (i *Uint64Value) SetValue(val interface{}) {
	*i = Uint64Value(val.(uint64))
}

// StringValue string Value
type StringValue string

// Set sets string value from the given string value.
func (s *StringValue) Set(val string) error {
	*s = StringValue(val)
	return nil
}

// Get returns the string value.
func (s *StringValue) Get() interface{} { return string(*s) }

func (s *StringValue) String() string { return string(*s) }

// SetValue sets the StringValue from the given string-asserted value.
func (s *StringValue) SetValue(val interface{}) {
	*s = StringValue(val.(string))
}

// Float64Value float64 Value
type Float64Value float64

// Set sets float64 value from the given string value.
func (f *Float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = Float64Value(v)
	return err
}

// Get returns the float64 value.
func (f *Float64Value) Get() interface{} { return float64(*f) }

func (f *Float64Value) String() string { return fmt.Sprintf("%v", *f) }

// SetValue sets the Float64Value from the given float64-asserted value.
func (f *Float64Value) SetValue(val interface{}) {
	*f = Float64Value(val.(float64))
}

// Duration is a custom type suitable for parsing duration values.
// It supports `time.ParseDuration`-compatible values and suffix-less digits; in
// the latter case, seconds are assumed.
type Duration time.Duration

// Set sets the duration from the given string value.
func (d *Duration) Set(s string) error {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		*d = Duration(time.Duration(v) * time.Second)
		return nil
	}

	v, err := time.ParseDuration(s)
	*d = Duration(v)
	return err
}

// Get returns the duration value.
func (d *Duration) Get() interface{} { return time.Duration(*d) }

// String returns a string representation of the duration value.
func (d *Duration) String() string { return (*time.Duration)(d).String() }

// SetValue sets the duration from the given Duration-asserted value.
func (d *Duration) SetValue(val interface{}) {
	*d = val.(Duration)
}

// MarshalText serialize the given duration value into a text.
func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText deserializes the given text into a duration value.
// It is meant to support TOML decoding of durations.
func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

// MarshalJSON serializes the given duration value.
func (d *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(*d))
}

// UnmarshalJSON deserializes the given text into a duration value.
func (d *Duration) UnmarshalJSON(text []byte) error {
	if v, err := strconv.ParseInt(string(text), 10, 64); err == nil {
		*d = Duration(time.Duration(v))
		return nil
	}

	// We use json unmarshal on value because we have the quoted version
	var value string
	err := json.Unmarshal(text, &value)
	if err != nil {
		return err
	}
	v, err := time.ParseDuration(value)
	*d = Duration(v)
	return err
}

// TimeValue time.Time Value
type TimeValue time.Time

// Set sets time.Time value from the given string value.
func (t *TimeValue) Set(s string) error {
	v, err := time.Parse(time.RFC3339, s)
	*t = TimeValue(v)
	return err
}

// Get returns the time.Time value.
func (t *TimeValue) Get() interface{} { return time.Time(*t) }

func (t *TimeValue) String() string { return (*time.Time)(t).String() }

// SetValue sets the TimeValue from the given time.Time-asserted value.
func (t *TimeValue) SetValue(val interface{}) {
	*t = TimeValue(val.(time.Time))
}

// SliceStrings parse slice of strings
type SliceStrings []string

// Set adds strings elem into the the parser.
// It splits str on , and ;
func (s *SliceStrings) Set(str string) error {
	fargs := func(c rune) bool {
		return c == ',' || c == ';'
	}
	// get function
	slice := strings.FieldsFunc(str, fargs)
	*s = append(*s, slice...)
	return nil
}

// Get []string
func (s *SliceStrings) Get() interface{} { return []string(*s) }

// String return slice in a string
func (s *SliceStrings) String() string { return fmt.Sprintf("%v", *s) }

// SetValue sets []string into the parser
func (s *SliceStrings) SetValue(val interface{}) {
	*s = SliceStrings(val.([]string))
}

// LoadParsers loads default parsers and custom parsers given as parameter.
// Return a map [reflect.Type]parsers
// bool, int, int64, uint, uint64, float64,
func LoadParsers(customParsers map[reflect.Type]Parser) (map[reflect.Type]Parser, error) {
	parsers := map[reflect.Type]Parser{}

	var boolParser BoolValue
	parsers[reflect.TypeOf(true)] = &boolParser

	var intParser IntValue
	parsers[reflect.TypeOf(1)] = &intParser

	var int64Parser Int64Value
	parsers[reflect.TypeOf(int64(1))] = &int64Parser

	var uintParser UintValue
	parsers[reflect.TypeOf(uint(1))] = &uintParser

	var uint64Parser Uint64Value
	parsers[reflect.TypeOf(uint64(1))] = &uint64Parser

	var stringParser StringValue
	parsers[reflect.TypeOf("")] = &stringParser

	var float64Parser Float64Value
	parsers[reflect.TypeOf(float64(1.5))] = &float64Parser

	var durationParser Duration
	parsers[reflect.TypeOf(Duration(time.Second))] = &durationParser

	var timeParser TimeValue
	parsers[reflect.TypeOf(time.Now())] = &timeParser

	for rType, parser := range customParsers {
		parsers[rType] = parser
	}
	return parsers, nil
}
