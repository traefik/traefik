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

// PodBackoff describes the backoff for re-run attempts of a pod
type PodBackoff struct {
	Backoff        *int     `json:"backoff,omitempty"`
	BackoffFactor  *float64 `json:"backoffFactor,omitempty"`
	MaxLaunchDelay *int     `json:"maxLaunchDelay,omitempty"`
}

// PodUpgrade describes the policy for upgrading a pod in-place
type PodUpgrade struct {
	MinimumHealthCapacity *float64 `json:"minimumHealthCapacity,omitempty"`
	MaximumOverCapacity   *float64 `json:"maximumOverCapacity,omitempty"`
}

// PodPlacement supports constraining which hosts a pod is placed on
type PodPlacement struct {
	Constraints           *[]Constraint `json:"constraints"`
	AcceptedResourceRoles []string    `json:"acceptedResourceRoles,omitempty"`
}

// PodSchedulingPolicy is the overarching pod scheduling policy
type PodSchedulingPolicy struct {
	Backoff   *PodBackoff   `json:"backoff,omitempty"`
	Upgrade   *PodUpgrade   `json:"upgrade,omitempty"`
	Placement *PodPlacement `json:"placement,omitempty"`
}

// Constraint describes the constraint for pod placement
type Constraint struct {
	FieldName  string `json:"fieldName"`
	Operator   string `json:"operator"`
	Value      string `json:"value,omitempty"`
}

// NewPodPlacement creates an empty PodPlacement
func NewPodPlacement() *PodPlacement {
	return &PodPlacement{
		Constraints:           &[]Constraint{},
		AcceptedResourceRoles: []string{},
	}
}

// AddConstraint adds a new constraint
//		constraints:	the constraint definition, one constraint per array element
func (r *PodPlacement) AddConstraint(constraint Constraint) *PodPlacement {
	c := *r.Constraints
	c = append(c, constraint)
	r.Constraints = &c

	return r
}

// NewPodSchedulingPolicy creates an empty PodSchedulingPolicy
func NewPodSchedulingPolicy() *PodSchedulingPolicy {
	return &PodSchedulingPolicy{
		Placement: NewPodPlacement(),
	}
}
