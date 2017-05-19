/*
Copyright 2017 Rohith All rights reserved.

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

// UnreachableStrategy is the unreachable strategy applied to an application.
type UnreachableStrategy struct {
	InactiveAfterSeconds *float64 `json:"inactiveAfterSeconds,omitempty"`
	ExpungeAfterSeconds  *float64 `json:"expungeAfterSeconds,omitempty"`
}

// SetInactiveAfterSeconds sets the period after which instance will be marked as inactive.
func (us UnreachableStrategy) SetInactiveAfterSeconds(cap float64) UnreachableStrategy {
	us.InactiveAfterSeconds = &cap
	return us
}

// SetExpungeAfterSeconds sets the period after which instance will be expunged.
func (us UnreachableStrategy) SetExpungeAfterSeconds(cap float64) UnreachableStrategy {
	us.ExpungeAfterSeconds = &cap
	return us
}
