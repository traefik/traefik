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
	"fmt"
)

// Pod is the definition for an pod in marathon
type Pod struct {
	ID      string            `json:"id,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	Version string            `json:"version,omitempty"`
	User    string            `json:"user,omitempty"`
	// Non-secret environment variables. Actual secrets are stored in Secrets
	// Magic happens at marshaling/unmarshaling to get them into the correct schema
	Env               map[string]string    `json:"-"`
	Secrets           map[string]Secret    `json:"-"`
	Containers        []*PodContainer      `json:"containers,omitempty"`
	Volumes           []*PodVolume         `json:"volumes,omitempty"`
	Networks          []*PodNetwork        `json:"networks,omitempty"`
	Scaling           *PodScalingPolicy    `json:"scaling,omitempty"`
	Scheduling        *PodSchedulingPolicy `json:"scheduling,omitempty"`
	ExecutorResources *ExecutorResources   `json:"executorResources,omitempty"`
}

// PodScalingPolicy is the scaling policy of the pod
type PodScalingPolicy struct {
	Kind         string `json:"kind"`
	Instances    int    `json:"instances"`
	MaxInstances int    `json:"maxInstances,omitempty"`
}

// NewPod create an empty pod
func NewPod() *Pod {
	return &Pod{
		Labels:     map[string]string{},
		Env:        map[string]string{},
		Containers: []*PodContainer{},
		Secrets:    map[string]Secret{},
		Volumes:    []*PodVolume{},
		Networks:   []*PodNetwork{},
	}
}

// Name sets the name / ID of the pod i.e. the identifier for this pod
func (p *Pod) Name(id string) *Pod {
	p.ID = validateID(id)
	return p
}

// SetUser sets the user to run the pod as
func (p *Pod) SetUser(user string) *Pod {
	p.User = user
	return p
}

// EmptyLabels empties the labels in a pod
func (p *Pod) EmptyLabels() *Pod {
	p.Labels = make(map[string]string)
	return p
}

// AddLabel adds a label to a pod
func (p *Pod) AddLabel(key, value string) *Pod {
	p.Labels[key] = value
	return p
}

// SetLabels sets the labels for a pod
func (p *Pod) SetLabels(labels map[string]string) *Pod {
	p.Labels = labels
	return p
}

// EmptyEnvs empties the environment variables for a pod
func (p *Pod) EmptyEnvs() *Pod {
	p.Env = make(map[string]string)
	return p
}

// AddEnv adds an environment variable to a pod
func (p *Pod) AddEnv(name, value string) *Pod {
	if p.Env == nil {
		p = p.EmptyEnvs()
	}
	p.Env[name] = value
	return p
}

// ExtendEnv extends the environment with the new environment variables
func (p *Pod) ExtendEnv(env map[string]string) *Pod {
	if p.Env == nil {
		p = p.EmptyEnvs()
	}

	for k, v := range env {
		p.AddEnv(k, v)
	}
	return p
}

// AddContainer adds a container to a pod
func (p *Pod) AddContainer(container *PodContainer) *Pod {
	p.Containers = append(p.Containers, container)
	return p
}

// EmptySecrets empties the secret sources in a pod
func (p *Pod) EmptySecrets() *Pod {
	p.Secrets = make(map[string]Secret)
	return p
}

// GetSecretSource gets the source of the named secret
func (p *Pod) GetSecretSource(name string) (string, error) {
	if val, ok := p.Secrets[name]; ok {
		return val.Source, nil
	}
	return "", fmt.Errorf("secret does not exist")
}

// AddSecret adds the secret to the pod
func (p *Pod) AddSecret(envVar, secretName, sourceName string) *Pod {
	if p.Secrets == nil {
		p = p.EmptySecrets()
	}
	p.Secrets[secretName] = Secret{EnvVar: envVar, Source: sourceName}
	return p
}

// AddVolume adds a volume to a pod
func (p *Pod) AddVolume(vol *PodVolume) *Pod {
	p.Volumes = append(p.Volumes, vol)
	return p
}

// AddNetwork adds a PodNetwork to a pod
func (p *Pod) AddNetwork(net *PodNetwork) *Pod {
	p.Networks = append(p.Networks, net)
	return p
}

// Count sets the count of the pod
func (p *Pod) Count(count int) *Pod {
	p.Scaling = &PodScalingPolicy{
		Kind:      "fixed",
		Instances: count,
	}
	return p
}

// SetPodSchedulingPolicy sets the PodSchedulingPolicy of a pod
func (p *Pod) SetPodSchedulingPolicy(policy *PodSchedulingPolicy) *Pod {
	p.Scheduling = policy
	return p
}

// SetExecutorResources sets the resources for the pod executor
func (p *Pod) SetExecutorResources(resources *ExecutorResources) *Pod {
	p.ExecutorResources = resources
	return p
}

// SupportsPods determines if this version of marathon supports pods
// If HEAD returns 200 it does
func (r *marathonClient) SupportsPods() (bool, error) {
	if err := r.apiHead(marathonAPIPods, nil); err != nil {
		// If we get a 404 we can return a strict false, otherwise it could be
		// a valid error
		if apiErr, ok := err.(*APIError); ok && apiErr.ErrCode == ErrCodeNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Pod gets a pod object from marathon by name
func (r *marathonClient) Pod(name string) (*Pod, error) {
	uri := buildPodURI(name)
	result := new(Pod)
	if err := r.apiGet(uri, nil, result); err != nil {
		return nil, err
	}

	return result, nil
}

// Pods gets all pods from marathon
func (r *marathonClient) Pods() ([]Pod, error) {
	var result []Pod
	if err := r.apiGet(marathonAPIPods, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreatePod creates a new pod in Marathon
func (r *marathonClient) CreatePod(pod *Pod) (*Pod, error) {
	result := new(Pod)
	if err := r.apiPost(marathonAPIPods, &pod, result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeletePod deletes a pod from marathon
func (r *marathonClient) DeletePod(name string, force bool) (*DeploymentID, error) {
	uri := fmt.Sprintf("%s?force=%v", buildPodURI(name), force)

	deployID := new(DeploymentID)
	if err := r.apiDelete(uri, nil, deployID); err != nil {
		return nil, err
	}

	return deployID, nil
}

// UpdatePod creates a new pod in Marathon
func (r *marathonClient) UpdatePod(pod *Pod, force bool) (*Pod, error) {
	uri := fmt.Sprintf("%s?force=%v", buildPodURI(pod.ID), force)
	result := new(Pod)

	if err := r.apiPut(uri, pod, result); err != nil {
		return nil, err
	}

	return result, nil
}

// PodVersions gets all the deployed versions of a pod
func (r *marathonClient) PodVersions(name string) ([]string, error) {
	uri := buildPodVersionURI(name)
	var result []string
	if err := r.apiGet(uri, nil, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// PodByVersion gets a pod by a version identifier
func (r *marathonClient) PodByVersion(name, version string) (*Pod, error) {
	uri := fmt.Sprintf("%s/%s", buildPodVersionURI(name), version)
	result := new(Pod)
	if err := r.apiGet(uri, nil, result); err != nil {
		return nil, err
	}

	return result, nil
}

func buildPodVersionURI(name string) string {
	return fmt.Sprintf("%s/%s::versions", marathonAPIPods, trimRootPath(name))
}

func buildPodURI(path string) string {
	return fmt.Sprintf("%s/%s", marathonAPIPods, trimRootPath(path))
}
