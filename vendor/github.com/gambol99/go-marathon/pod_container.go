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

// PodContainer describes a container in a pod
type PodContainer struct {
	Name         string             `json:"name,omitempty"`
	Exec         *PodExec           `json:"exec,omitempty"`
	Resources    *Resources         `json:"resources,omitempty"`
	Endpoints    []*PodEndpoint     `json:"endpoints,omitempty"`
	Image        *PodContainerImage `json:"image,omitempty"`
	Env          map[string]string  `json:"-"`
	Secrets      map[string]Secret  `json:"-"`
	User         string             `json:"user,omitempty"`
	HealthCheck  *PodHealthCheck    `json:"healthCheck,omitempty"`
	VolumeMounts []*PodVolumeMount  `json:"volumeMounts,omitempty"`
	Artifacts    []*PodArtifact     `json:"artifacts,omitempty"`
	Labels       map[string]string  `json:"labels,omitempty"`
	Lifecycle    PodLifecycle       `json:"lifecycle,omitempty"`
}

// PodLifecycle describes the lifecycle of a pod
type PodLifecycle struct {
	KillGracePeriodSeconds *float64 `json:"killGracePeriodSeconds,omitempty"`
}

// PodCommand is the command to run as the entrypoint of the container
type PodCommand struct {
	Shell string `json:"shell,omitempty"`
}

// PodExec contains the PodCommand
type PodExec struct {
	Command PodCommand `json:"command,omitempty"`
}

// PodArtifact describes how to obtain a generic artifact for a pod
type PodArtifact struct {
	URI        string `json:"uri,omitempty"`
	Extract    bool   `json:"extract,omitempty"`
	Executable bool   `json:"executable,omitempty"`
	Cache      bool   `json:"cache,omitempty"`
	DestPath   string `json:"destPath,omitempty"`
}

// NewPodContainer creates an empty PodContainer
func NewPodContainer() *PodContainer {
	return &PodContainer{
		Endpoints:    []*PodEndpoint{},
		Env:          map[string]string{},
		VolumeMounts: []*PodVolumeMount{},
		Artifacts:    []*PodArtifact{},
		Labels:       map[string]string{},
		Resources:    NewResources(),
	}
}

// SetName sets the name of a pod container
func (p *PodContainer) SetName(name string) *PodContainer {
	p.Name = name
	return p
}

// SetCommand sets the shell command of a pod container
func (p *PodContainer) SetCommand(name string) *PodContainer {
	p.Exec = &PodExec{
		Command: PodCommand{
			Shell: name,
		},
	}
	return p
}

// CPUs sets the CPUs of a pod container
func (p *PodContainer) CPUs(cpu float64) *PodContainer {
	p.Resources.Cpus = cpu
	return p
}

// Memory sets the memory of a pod container
func (p *PodContainer) Memory(memory float64) *PodContainer {
	p.Resources.Mem = memory
	return p
}

// Storage sets the storage capacity of a pod container
func (p *PodContainer) Storage(disk float64) *PodContainer {
	p.Resources.Disk = disk
	return p
}

// GPUs sets the GPU requirements of a pod container
func (p *PodContainer) GPUs(gpu int32) *PodContainer {
	p.Resources.Gpus = gpu
	return p
}

// AddEndpoint appends an endpoint for a pod container
func (p *PodContainer) AddEndpoint(endpoint *PodEndpoint) *PodContainer {
	p.Endpoints = append(p.Endpoints, endpoint)
	return p
}

// SetImage sets the image of a pod container
func (p *PodContainer) SetImage(image *PodContainerImage) *PodContainer {
	p.Image = image
	return p
}

// EmptyEnvironment initialized env to empty
func (p *PodContainer) EmptyEnvs() *PodContainer {
	p.Env = make(map[string]string)
	return p
}

// AddEnvironment adds an environment variable for a pod container
func (p *PodContainer) AddEnv(name, value string) *PodContainer {
	if p.Env == nil {
		p = p.EmptyEnvs()
	}
	p.Env[name] = value
	return p
}

// ExtendEnvironment extends the environment for a pod container
func (p *PodContainer) ExtendEnv(env map[string]string) *PodContainer {
	if p.Env == nil {
		p = p.EmptyEnvs()
	}
	for k, v := range env {
		p.AddEnv(k, v)
	}
	return p
}

// AddSecret adds a secret to the environment for a pod container
func (p *PodContainer) AddSecret(name, secretName string) *PodContainer {
	if p.Env == nil {
		p = p.EmptyEnvs()
	}
	p.Env[name] = secretName
	return p
}

// SetUser sets the user to run the pod as
func (p *PodContainer) SetUser(user string) *PodContainer {
	p.User = user
	return p
}

// SetHealthCheck sets the health check of a pod container
func (p *PodContainer) SetHealthCheck(healthcheck *PodHealthCheck) *PodContainer {
	p.HealthCheck = healthcheck
	return p
}

// AddVolumeMount appends a volume mount to a pod container
func (p *PodContainer) AddVolumeMount(mount *PodVolumeMount) *PodContainer {
	p.VolumeMounts = append(p.VolumeMounts, mount)
	return p
}

// AddArtifact appends an artifact to a pod container
func (p *PodContainer) AddArtifact(artifact *PodArtifact) *PodContainer {
	p.Artifacts = append(p.Artifacts, artifact)
	return p
}

// AddLabel adds a label to a pod container
func (p *PodContainer) AddLabel(key, value string) *PodContainer {
	p.Labels[key] = value
	return p
}

// SetLifecycle sets the lifecycle of a pod container
func (p *PodContainer) SetLifecycle(lifecycle PodLifecycle) *PodContainer {
	p.Lifecycle = lifecycle
	return p
}
