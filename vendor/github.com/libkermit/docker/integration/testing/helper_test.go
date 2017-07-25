package dockerit

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	d "github.com/libkermit/docker"
	docker "github.com/libkermit/docker/testing"
)

func setupTest(t *testing.T) *docker.Project {
	return cleanContainers(t)
}

func cleanContainers(t *testing.T) *docker.Project {
	client, err := dockerclient.NewEnvClient()
	if err != nil {
		t.Fatal(err)
	}
	// FIXME(vdemeester) Fix this
	client.UpdateClientVersion(d.CurrentAPIVersion)

	ctx := context.Background()

	filterArgs := filters.NewArgs()
	if filterArgs, err = filters.ParseFlag(d.KermitLabelFilter, filterArgs); err != nil {
		t.Fatal(err)
	}

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, container := range containers {
		t.Logf("cleaning container %s…", container.ID)
		if err := client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			t.Errorf("Error while removing container %s : %v\n", container.ID, err)
		}
	}

	return docker.NewProject(client)
}
