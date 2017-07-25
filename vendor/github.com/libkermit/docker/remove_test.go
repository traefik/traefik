package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type removeClient struct {
	client.Client
	removeFunc func(context.Context, string, types.ContainerRemoveOptions) error
}

func (c *removeClient) ContainerRemove(context context.Context, id string, options types.ContainerRemoveOptions) error {
	if c.removeFunc != nil {
		return c.removeFunc(context, id, options)
	}
	return nil
}

func TestProjectRemoteError(t *testing.T) {
	project := NewProject(&removeClient{
		removeFunc: func(context context.Context, id string, options types.ContainerRemoveOptions) error {
			return errors.New("error happens")
		},
	})
	err := project.Remove("my_container")
	if err == nil {
		t.Fatalf("Expected an error, got nothing")
	}
}

func TestProjectRemoteNoError(t *testing.T) {
	project := NewProject(&removeClient{})
	err := project.Remove("my_container")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
