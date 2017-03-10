package docker

// Clean stops and removes (by default, controllable with the keep) kermit containers
func (p *Project) Clean(keep bool) error {
	containers, err := p.List()
	if err != nil {
		return err
	}
	for _, container := range containers {
		if err := p.StopWithTimeout(container.ID, 1); err != nil {
			return err
		}
		if !keep {
			if err := p.Remove(container.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
