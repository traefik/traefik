package docker

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Pull pulls the given reference (image)
func (p *Project) Pull(ref string) error {
	return p.ensureImageExists(ref, true)
}

func (p *Project) ensureImageExists(ref string, force bool) error {
	if !force {
		// Check if ref is already there
		_, _, err := p.Client.ImageInspectWithRaw(context.Background(), ref)
		if err != nil && !client.IsErrNotFound(err) {
			return err
		}
		if err == nil {
			return nil
		}
	}

	// And pull it
	responseBody, err := p.Client.ImagePull(context.Background(), ref, types.ImagePullOptions{})
	if err != nil {
		fmt.Printf("%v", err)
		return err
	}
	defer responseBody.Close()

	_, err = ioutil.ReadAll(responseBody)
	return err
}
