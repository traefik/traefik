package docker

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
)

// Create lets you create a container with the specified image, and default
// configuration.
func (p *Project) Create(image string) (types.ContainerJSON, error) {
	return p.CreateWithConfig(image, ContainerConfig{})
}

// CreateWithConfig lets you create a container with the specified image, and
// some custom simple configuration.
func (p *Project) CreateWithConfig(image string, containerConfig ContainerConfig) (types.ContainerJSON, error) {
	err := p.ensureImageExists(image, false)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	labels := mergeLabels(p.Labels, containerConfig.Labels)
	config := &container.Config{
		Image:  image,
		Labels: labels,
	}

	if len(containerConfig.Entrypoint) > 0 {
		config.Entrypoint = strslice.StrSlice(containerConfig.Entrypoint)
	}
	if len(containerConfig.Cmd) > 0 {
		config.Cmd = strslice.StrSlice(containerConfig.Cmd)
	}

	var containerName string
	if containerConfig.Name != "" {
		containerName = containerConfig.Name
	} else {
		containerName = fmt.Sprintf("kermit_%s", randSeq(10))
	}

	response, err := p.Client.ContainerCreate(context.Background(), config, &container.HostConfig{}, nil, containerName)
	if err != nil {
		return types.ContainerJSON{}, err
	}

	return p.Inspect(response.ID)
}

func mergeLabels(defaultLabels, additionnalLabels map[string]string) map[string]string {
	labels := make(map[string]string, len(defaultLabels))
	if len(additionnalLabels) > 0 {
		for key, value := range additionnalLabels {
			labels[key] = value
		}
	}
	// default labels overrides additionnals
	for key, value := range defaultLabels {
		labels[key] = value
	}
	return labels
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
