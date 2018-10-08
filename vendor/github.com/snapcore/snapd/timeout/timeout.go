// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package timeout

import (
	"encoding/json"
	"time"
)

// Timeout is a time.Duration that knows how to roundtrip to json and yaml
type Timeout time.Duration

// DefaultTimeout specifies the timeout for services that do not specify StopTimeout
var DefaultTimeout = Timeout(30 * time.Second)

// MarshalJSON is from the json.Marshaler interface
func (t Timeout) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON is from the json.Unmarshaler interface
func (t *Timeout) UnmarshalJSON(buf []byte) error {
	var str string
	if err := json.Unmarshal(buf, &str); err != nil {
		return err
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*t = Timeout(dur)

	return nil
}

func (t *Timeout) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*t = Timeout(dur)
	return nil
}

// String returns a string representing the duration
func (t Timeout) String() string {
	return time.Duration(t).String()
}

// Seconds returns the duration as a floating point number of seconds.
func (t Timeout) Seconds() float64 {
	return time.Duration(t).Seconds()
}
