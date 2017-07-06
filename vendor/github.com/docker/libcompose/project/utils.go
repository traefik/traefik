package project

import (
	"strings"
)

// DefaultDependentServices return the dependent services (as an array of ServiceRelationship)
// for the specified project and service. It looks for : links, volumesFrom, net and ipc configuration.
func DefaultDependentServices(p *Project, s Service) []ServiceRelationship {
	config := s.Config()
	if config == nil {
		return []ServiceRelationship{}
	}

	result := []ServiceRelationship{}
	for _, link := range config.Links {
		result = append(result, NewServiceRelationship(link, RelTypeLink))
	}

	for _, volumesFrom := range config.VolumesFrom {
		result = append(result, NewServiceRelationship(volumesFrom, RelTypeVolumesFrom))
	}

	for _, dependsOn := range config.DependsOn {
		result = append(result, NewServiceRelationship(dependsOn, RelTypeDependsOn))
	}

	if config.NetworkMode != "" {
		if strings.HasPrefix(config.NetworkMode, "service:") {
			serviceName := config.NetworkMode[8:]
			result = append(result, NewServiceRelationship(serviceName, RelTypeNetworkMode))
		}
	}

	return result
}

// NameAlias returns the name and alias based on the specified string.
// If the name contains a colon (like name:alias) it will split it, otherwise
// it will return the specified name as name and alias.
func NameAlias(name string) (string, string) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], parts[0]
}
