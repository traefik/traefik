/*
Copyright 2017 The go-marathon Authors All rights reserved.

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

package marathon

import (
	"encoding/json"
	"fmt"
)

// UnreachableStrategyAbsenceReasonDisabled signifies the reason of disabled unreachable strategy
const UnreachableStrategyAbsenceReasonDisabled = "disabled"

// UnreachableStrategy is the unreachable strategy applied to an application.
type UnreachableStrategy struct {
	EnabledUnreachableStrategy
	AbsenceReason string
}

// EnabledUnreachableStrategy covers parameters pertaining to present unreachable strategies.
type EnabledUnreachableStrategy struct {
	InactiveAfterSeconds *float64 `json:"inactiveAfterSeconds,omitempty"`
	ExpungeAfterSeconds  *float64 `json:"expungeAfterSeconds,omitempty"`
}

type unreachableStrategy UnreachableStrategy

// UnmarshalJSON unmarshals the given JSON into an UnreachableStrategy. It
// populates parameters for present strategies, and otherwise only sets the
// absence reason.
func (us *UnreachableStrategy) UnmarshalJSON(b []byte) error {
	var u unreachableStrategy
	var errEnabledUS, errNonEnabledUS error
	if errEnabledUS = json.Unmarshal(b, &u); errEnabledUS == nil {
		*us = UnreachableStrategy(u)
		return nil
	}

	if errNonEnabledUS = json.Unmarshal(b, &us.AbsenceReason); errNonEnabledUS == nil {
		return nil
	}

	return fmt.Errorf("failed to unmarshal unreachable strategy: unmarshaling into enabled returned error '%s'; unmarshaling into non-enabled returned error '%s'", errEnabledUS, errNonEnabledUS)
}

// MarshalJSON marshals the unreachable strategy.
func (us *UnreachableStrategy) MarshalJSON() ([]byte, error) {
	if us.AbsenceReason == "" {
		return json.Marshal(us.EnabledUnreachableStrategy)
	}

	return json.Marshal(us.AbsenceReason)
}

// SetInactiveAfterSeconds sets the period after which instance will be marked as inactive.
func (us *UnreachableStrategy) SetInactiveAfterSeconds(cap float64) *UnreachableStrategy {
	us.InactiveAfterSeconds = &cap
	return us
}

// SetExpungeAfterSeconds sets the period after which instance will be expunged.
func (us *UnreachableStrategy) SetExpungeAfterSeconds(cap float64) *UnreachableStrategy {
	us.ExpungeAfterSeconds = &cap
	return us
}
