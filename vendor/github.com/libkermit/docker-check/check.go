// Package docker aims to provide simple "helper" methods to ease the use of
// docker in (integration) tests using the go-check testing package.
//
// It does support a subset of options compared to actual client api, as it
// is more focused on needs for integration tests.
package docker

import (
	"github.com/go-check/check"

	"github.com/docker/docker/client"
	"github.com/libkermit/docker"
)

// Project holds docker related project attributes, like docker client, labels
// to put on the containers, and so on.
type Project struct {
	project *docker.Project
}

// NewProjectFromEnv creates a project with a client that is build from environment variables.
func NewProjectFromEnv(c *check.C) *Project {
	client, err := client.NewEnvClient()
	c.Assert(err, check.IsNil, check.Commentf("Error while getting a docker client from env"))
	return NewProject(client)
}

// NewProject creates a project with the given client and the default attributes.
func NewProject(client client.APIClient) *Project {
	return &Project{
		project: &docker.Project{
			Client: client,
			Labels: docker.KermitLabels,
		},
	}
}
