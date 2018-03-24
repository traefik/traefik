package marathon

import (
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/gambol99/go-marathon"
)

func getSegmentNameSuffix(serviceName string) string {
	if len(serviceName) > 0 {
		return "-service-" + provider.Normalize(serviceName)
	}
	return ""
}

func (p *Provider) getSubDomain(name string) string {
	if p.GroupsAsSubDomains {
		splitedName := strings.Split(strings.TrimPrefix(name, "/"), "/")
		provider.ReverseStringSlice(&splitedName)
		reverseName := strings.Join(splitedName, ".")
		return reverseName
	}
	return strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1)
}

func (p *Provider) taskFilter(task marathon.Task, application marathon.Application) bool {
	if task.State != string(taskStateRunning) {
		return false
	}

	if ready := p.readyChecker.Do(task, application); !ready {
		log.Infof("Filtering unready task %s from application %s", task.ID, application.ID)
		return false
	}

	return true
}
