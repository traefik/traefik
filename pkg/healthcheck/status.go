package healthcheck

type healthStatus struct {
	up       bool
	failures int
	passes   int
}

type healthStatusResult struct {
	update    bool
	count     int
	threshold int
}

func (h *healthStatus) observe(up bool, fails, passes int) healthStatusResult {
	if h.up == up {
		h.failures = 0
		h.passes = 0
		return healthStatusResult{update: true}
	}

	if up {
		h.failures = 0
		h.passes++
		if h.passes < passes {
			return healthStatusResult{count: h.passes, threshold: passes}
		}

		h.up = true
		h.passes = 0
		return healthStatusResult{update: true, count: passes, threshold: passes}
	}

	h.passes = 0
	h.failures++
	if h.failures < fails {
		return healthStatusResult{count: h.failures, threshold: fails}
	}

	h.up = false
	h.failures = 0
	return healthStatusResult{update: true, count: fails, threshold: fails}
}
