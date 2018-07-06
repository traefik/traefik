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

// PodVolume describes a volume on the host
type PodVolume struct {
	Name string `json:"name,omitempty"`
	Host string `json:"host,omitempty"`
}

// PodVolumeMount describes how to mount a volume into a task
type PodVolumeMount struct {
	Name      string `json:"name,omitempty"`
	MountPath string `json:"mountPath,omitempty"`
}

// NewPodVolume creates a new PodVolume
func NewPodVolume(name, path string) *PodVolume {
	return &PodVolume{
		Name: name,
		Host: path,
	}
}

// NewPodVolumeMount creates a new PodVolumeMount
func NewPodVolumeMount(name, mount string) *PodVolumeMount {
	return &PodVolumeMount{
		Name:      name,
		MountPath: mount,
	}
}
