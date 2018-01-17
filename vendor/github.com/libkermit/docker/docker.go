// Package docker aims to provide simple "helper" methods to ease the use of
// docker in (integration) tests.
//
// It does support a subset of options compared to actual client api, as it
// is more focused on needs for integration tests.
package docker

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	// KermitLabel is the label used to add to containers
	KermitLabel = "com.github.vdemeester.libkermit"
)

var (
	// KermitHeader defines default kermit headers to pass through the API
	KermitHeader = map[string]string{
		"User-Agent": "libkermit-1.0",
	}

	// KermitLabels defines default docker labels kermit will but on
	// containers and such.
	KermitLabels = map[string]string{
		KermitLabel: "true",
	}

	// KermitLabelFilter is the filter to use to find containers managed by kermit
	KermitLabelFilter = fmt.Sprintf("label=%s", KermitLabel)

	// DefaultStopTimeout is the default timeout for the stop command
	DefaultStopTimeout = 10

	// CurrentAPIVersion defines the current "lower" docker API version
	// supported by libkermit.
	CurrentAPIVersion = "v1.21"
)

// Project holds docker related project attributes, like docker client, labels
// to put on the containers, and so on.
type Project struct {
	Client APIClient
	Labels map[string]string
}

// APIClient is a lightweight representation of docker client.APIClient
type APIClient interface {
	client.ContainerAPIClient
	client.ImageAPIClient
}

// NewProjectFromEnv creates a project with a client that is build from environment variables.
func NewProjectFromEnv() (*Project, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	ping := types.Ping{APIVersion: CurrentAPIVersion}
	cli.NegotiateAPIVersionPing(ping)
	return NewProject(cli), nil
}

// NewProject creates a project with the given client and the default attributes.
func NewProject(client APIClient) *Project {
	return &Project{
		Client: client,
		Labels: KermitLabels,
	}
}
