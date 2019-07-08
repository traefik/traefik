package types

import (
	"encoding/json"
	"strconv"
	"time"
)

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

// String returns a string representation of the duration value.
func (d Duration) String() string { return (time.Duration)(d).String() }

// MarshalText serialize the given duration value into a text.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText deserializes the given text into a duration value.
// It is meant to support TOML decoding of durations.
func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

// MarshalJSON serializes the given duration value.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d))
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
