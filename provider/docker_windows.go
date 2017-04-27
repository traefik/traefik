package provider

import "errors"

const (
	// DockerAPIVersion is a constant holding the version of the Docker API traefik will use
	DockerAPIVersion string = "1.24"
)

// get container ID from inside a container
func getContainerID() (string, error) {
	return "", errors.New("Not implemented")
}
