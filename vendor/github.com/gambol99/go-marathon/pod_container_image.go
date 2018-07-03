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

// ImageType represents the image format type
type ImageType string

const (
	// ImageTypeDocker is the docker format
	ImageTypeDocker ImageType = "DOCKER"

	// ImageTypeAppC is the appc format
	ImageTypeAppC ImageType = "APPC"
)

// PodContainerImage describes how to retrieve the container image
type PodContainerImage struct {
	Kind      ImageType `json:"kind,omitempty"`
	ID        string    `json:"id,omitempty"`
	ForcePull bool      `json:"forcePull,omitempty"`
}

// NewPodContainerImage creates an empty PodContainerImage
func NewPodContainerImage() *PodContainerImage {
	return &PodContainerImage{}
}

// SetKind sets the Kind of the image
func (i *PodContainerImage) SetKind(typ ImageType) *PodContainerImage {
	i.Kind = typ
	return i
}

// SetID sets the ID of the image
func (i *PodContainerImage) SetID(id string) *PodContainerImage {
	i.ID = id
	return i
}

// NewDockerPodContainerImage creates a docker PodContainerImage
func NewDockerPodContainerImage() *PodContainerImage {
	return NewPodContainerImage().SetKind(ImageTypeDocker)
}
