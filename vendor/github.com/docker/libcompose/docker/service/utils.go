package service

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/libcompose/project"
)

// DefaultDependentServices return the dependent services (as an array of ServiceRelationship)
// for the specified project and service. It looks for : links, volumesFrom, net and ipc configuration.
// It uses default project implementation and append some docker specific ones.
func DefaultDependentServices(p *project.Project, s project.Service) []project.ServiceRelationship {
	result := project.DefaultDependentServices(p, s)

	result = appendNs(p, result, s.Config().NetworkMode, project.RelTypeNetNamespace)
	result = appendNs(p, result, s.Config().Ipc, project.RelTypeIpcNamespace)

	return result
}

func appendNs(p *project.Project, rels []project.ServiceRelationship, conf string, relType project.ServiceRelationshipType) []project.ServiceRelationship {
	service := GetContainerFromIpcLikeConfig(p, conf)
	if service != "" {
		rels = append(rels, project.NewServiceRelationship(service, relType))
	}
	return rels
}

// GetContainerFromIpcLikeConfig returns name of the service that shares the IPC
// namespace with the specified service.
func GetContainerFromIpcLikeConfig(p *project.Project, conf string) string {
	ipc := container.IpcMode(conf)
	if !ipc.IsContainer() {
		return ""
	}

	name := ipc.Container()
	if name == "" {
		return ""
	}

	if p.ServiceConfigs.Has(name) {
		return name
	}
	return ""
}
