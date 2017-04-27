package provider

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	// DockerAPIVersion is a constant holding the version of the Docker API traefik will use
	DockerAPIVersion string = "1.21"
)

// get container ID from inside a container
func getContainerID() (string, error) {
	cgroup := "/proc/self/cgroup"

	file, err := os.Open(cgroup)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "docker") {
			i := strings.LastIndex(line, "/")
			containerID := line[i+1:]
			return containerID, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("Failed to get container ID from %s", cgroup)
}
