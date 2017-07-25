/*
Copyright 2015 Rohith All rights reserved.

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func createPortMapping(containerPort int, protocol string) *PortMapping {
	return &PortMapping{
		ContainerPort: containerPort,
		HostPort:      0,
		ServicePort:   0,
		Protocol:      protocol,
	}
}

func TestDockerAddParameter(t *testing.T) {
	docker := NewDockerApplication().Container.Docker
	docker.AddParameter("k1", "v1").AddParameter("k2", "v2")

	assert.Equal(t, 2, len(*docker.Parameters))
	assert.Equal(t, (*docker.Parameters)[0].Key, "k1")
	assert.Equal(t, (*docker.Parameters)[0].Value, "v1")
	assert.Equal(t, (*docker.Parameters)[1].Key, "k2")
	assert.Equal(t, (*docker.Parameters)[1].Value, "v2")

	docker.EmptyParameters()
	assert.NotNil(t, docker.Parameters)
	assert.Equal(t, 0, len(*docker.Parameters))
}

func TestDockerExpose(t *testing.T) {
	app := NewDockerApplication()
	app.Container.Docker.Expose(8080).Expose(80, 443)

	portMappings := app.Container.Docker.PortMappings
	assert.Equal(t, 3, len(*portMappings))

	assert.Equal(t, *createPortMapping(8080, "tcp"), (*portMappings)[0])
	assert.Equal(t, *createPortMapping(80, "tcp"), (*portMappings)[1])
	assert.Equal(t, *createPortMapping(443, "tcp"), (*portMappings)[2])
}

func TestDockerExposeUDP(t *testing.T) {
	app := NewDockerApplication()
	app.Container.Docker.ExposeUDP(53).ExposeUDP(5060, 6881)

	portMappings := app.Container.Docker.PortMappings
	assert.Equal(t, 3, len(*portMappings))
	assert.Equal(t, *createPortMapping(53, "udp"), (*portMappings)[0])
	assert.Equal(t, *createPortMapping(5060, "udp"), (*portMappings)[1])
	assert.Equal(t, *createPortMapping(6881, "udp"), (*portMappings)[2])
}

func TestPortMappingLabels(t *testing.T) {
	pm := createPortMapping(80, "tcp")

	pm.AddLabel("hello", "world").AddLabel("foo", "bar")

	assert.Equal(t, 2, len(*pm.Labels))
	assert.Equal(t, "world", (*pm.Labels)["hello"])
	assert.Equal(t, "bar", (*pm.Labels)["foo"])

	pm.EmptyLabels()

	assert.NotNil(t, pm.Labels)
	assert.Equal(t, 0, len(*pm.Labels))
}

func TestVolume(t *testing.T) {
	container := NewDockerApplication().Container

	container.Volume("hp1", "cp1", "RW")
	container.Volume("hp2", "cp2", "R")

	assert.Equal(t, 2, len(*container.Volumes))
	assert.Equal(t, (*container.Volumes)[0].HostPath, "hp1")
	assert.Equal(t, (*container.Volumes)[0].ContainerPath, "cp1")
	assert.Equal(t, (*container.Volumes)[0].Mode, "RW")
	assert.Equal(t, (*container.Volumes)[1].HostPath, "hp2")
	assert.Equal(t, (*container.Volumes)[1].ContainerPath, "cp2")
	assert.Equal(t, (*container.Volumes)[1].Mode, "R")
}

func TestExternalVolume(t *testing.T) {
	container := NewDockerApplication().Container

	container.Volume("", "cp", "RW")
	ev := (*container.Volumes)[0].SetExternalVolume("myVolume", "dvdi")

	ev.AddOption("prop", "pval")
	ev.AddOption("dvdi", "rexray")

	ev1 := (*container.Volumes)[0].External
	assert.Equal(t, ev1.Name, "myVolume")
	assert.Equal(t, ev1.Provider, "dvdi")
	if assert.Equal(t, len(*ev1.Options), 2) {
		assert.Equal(t, (*ev1.Options)["dvdi"], "rexray")
		assert.Equal(t, (*ev1.Options)["prop"], "pval")
	}

	// empty the external volume again
	(*container.Volumes)[0].EmptyExternalVolume()
	ev2 := (*container.Volumes)[0].External
	assert.Equal(t, ev2.Name, "")
	assert.Equal(t, ev2.Provider, "")
}
