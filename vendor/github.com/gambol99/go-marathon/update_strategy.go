/*
Copyright 2014 Rohith All rights reserved.

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

// UpgradeStrategy is the upgrade strategy applied to an application.
type UpgradeStrategy struct {
	MinimumHealthCapacity *float64 `json:"minimumHealthCapacity,omitempty"`
	MaximumOverCapacity   *float64 `json:"maximumOverCapacity,omitempty"`
}

// SetMinimumHealthCapacity sets the minimum health capacity.
func (us UpgradeStrategy) SetMinimumHealthCapacity(cap float64) UpgradeStrategy {
	us.MinimumHealthCapacity = &cap
	return us
}

// SetMaximumOverCapacity sets the maximum over capacity.
func (us UpgradeStrategy) SetMaximumOverCapacity(cap float64) UpgradeStrategy {
	us.MaximumOverCapacity = &cap
	return us
}
