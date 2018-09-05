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

// ExecutorResources are the resources supported by an executor (a task running a pod)
type ExecutorResources struct {
	Cpus float64 `json:"cpus,omitempty"`
	Mem  float64 `json:"mem,omitempty"`
	Disk float64 `json:"disk,omitempty"`
}

// Resources are the full set of resources for a task
type Resources struct {
	Cpus float64 `json:"cpus"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk,omitempty"`
	Gpus int32   `json:"gpus,omitempty"`
}

// NewResources creates an empty Resources
func NewResources() *Resources {
	return &Resources{}
}
