package cloudflare

import (
	"encoding/json"
	"time"
)

// Duration implements json.Marshaler and json.Unmarshaler for time.Duration
// using the fmt.Stringer interface of time.Duration and time.ParseDuration.
type Duration struct {
	time.Duration
}

// MarshalJSON encodes a Duration as a JSON string formatted using String.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// UnmarshalJSON decodes a Duration from a JSON string parsed using time.ParseDuration.
func (d *Duration) UnmarshalJSON(buf []byte) error {
	var str string

	err := json.Unmarshal(buf, &str)
	if err != nil {
		return err
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	d.Duration = dur
	return nil
}

var (
	_ = json.Marshaler((*Duration)(nil))
	_ = json.Unmarshaler((*Duration)(nil))
)
