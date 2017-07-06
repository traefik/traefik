package docker

// IsRunning checks if the container is running or not
func (p *Project) IsRunning(containerID string) (bool, error) {
	return p.containerStatus(containerID, "running")
}

// IsStopped checks if the container is running or not
func (p *Project) IsStopped(containerID string) (bool, error) {
	return p.containerStatus(containerID, "stopped")
}

// IsPaused checks if the container is running or not
func (p *Project) IsPaused(containerID string) (bool, error) {
	return p.containerStatus(containerID, "paused")
}

func (p *Project) containerStatus(containerID, status string) (bool, error) {
	containerJSON, err := p.Inspect(containerID)
	if err != nil {
		return false, err
	}
	return containerJSON.State.Status == status, nil

}
