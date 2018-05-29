package docker

import (
	"time"

	"golang.org/x/net/context"
)

// Stop stops the container with a default timeout.
func (p *Project) Stop(containerID string) error {
	return p.StopWithTimeout(containerID, DefaultStopTimeout)
}

// StopWithTimeout stops the container with the specified timeout.
func (p *Project) StopWithTimeout(containerID string, timeout int) error {
	timeoutDuration := time.Duration(timeout) * time.Second

	return p.Client.ContainerStop(context.Background(), containerID, &timeoutDuration)
}
