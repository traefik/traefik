// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2017 Canonical Ltd
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

package snap

import (
	"errors"
)

// RestartCondition encapsulates the different systemd 'restart' options
type RestartCondition string

// These are the supported restart conditions
const (
	RestartNever      RestartCondition = "never"
	RestartOnSuccess  RestartCondition = "on-success"
	RestartOnFailure  RestartCondition = "on-failure"
	RestartOnAbnormal RestartCondition = "on-abnormal"
	RestartOnAbort    RestartCondition = "on-abort"
	RestartOnWatchdog RestartCondition = "on-watchdog"
	RestartAlways     RestartCondition = "always"
)

var RestartMap = map[string]RestartCondition{
	"no":          RestartNever,
	"never":       RestartNever,
	"on-success":  RestartOnSuccess,
	"on-failure":  RestartOnFailure,
	"on-abnormal": RestartOnAbnormal,
	"on-abort":    RestartOnAbort,
	"on-watchdog": RestartOnWatchdog,
	"always":      RestartAlways,
}

// ErrUnknownRestartCondition is returned when trying to unmarshal an unknown restart condition
var ErrUnknownRestartCondition = errors.New("invalid restart condition")

func (rc RestartCondition) String() string {
	if rc == "never" {
		return "no"
	}
	return string(rc)
}

// UnmarshalYAML so RestartCondition implements yaml's Unmarshaler interface
func (rc *RestartCondition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string

	if err := unmarshal(&v); err != nil {
		return err
	}

	nrc, ok := RestartMap[v]
	if !ok {
		return ErrUnknownRestartCondition
	}

	*rc = nrc

	return nil
}
