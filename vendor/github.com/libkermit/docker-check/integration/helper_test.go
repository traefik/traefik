package dockerit

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/go-check/check"
	d "github.com/libkermit/docker"
	docker "github.com/libkermit/docker-check"
)

func setupTest(c *check.C) *docker.Project {
	return cleanContainers(c)
}

func cleanContainers(c *check.C) *docker.Project {
	client, err := dockerclient.NewEnvClient()
	c.Assert(err, check.IsNil)
	// FIXME(vdemeester) fix this
	client.UpdateClientVersion(d.CurrentAPIVersion)

	filterArgs := filters.NewArgs()
	filterArgs, err = filters.ParseFlag(d.KermitLabelFilter, filterArgs)
	c.Assert(err, check.IsNil)

	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	c.Assert(err, check.IsNil)

	for _, container := range containers {
		c.Logf("cleaning container %s…", container.ID)
		if err := client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			c.Errorf("Error while removing container %s : %v\n", container.ID, err)
		}
	}

	return docker.NewProject(client)
}
