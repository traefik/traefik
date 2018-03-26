package servicefabric

import (
	sf "github.com/jjcollinge/servicefabric"
)

// ServiceItemExtended provides a flattened view
// of the service with details of the application
// it belongs too and the replicas/partitions
type ServiceItemExtended struct {
	sf.ServiceItem
	Application sf.ApplicationItem
	Partitions  []PartitionItemExtended
	Labels      map[string]string
}

// PartitionItemExtended provides a flattened view
// of a services partitions
type PartitionItemExtended struct {
	sf.PartitionItem
	Replicas  []sf.ReplicaItem
	Instances []sf.InstanceItem
}

// sfClient is an interface for Service Fabric client's to implement.
// This is purposely a subset of the total Service Fabric API surface.
type sfClient interface {
	GetApplications() (*sf.ApplicationItemsPage, error)
	GetServices(appName string) (*sf.ServiceItemsPage, error)
	GetPartitions(appName, serviceName string) (*sf.PartitionItemsPage, error)
	GetReplicas(appName, serviceName, partitionName string) (*sf.ReplicaItemsPage, error)
	GetInstances(appName, serviceName, partitionName string) (*sf.InstanceItemsPage, error)
	GetServiceExtensionMap(service *sf.ServiceItem, app *sf.ApplicationItem, extensionKey string) (map[string]string, error)
	GetServiceLabels(service *sf.ServiceItem, app *sf.ApplicationItem, prefix string) (map[string]string, error)
	GetProperties(name string) (bool, map[string]string, error)
}

// replicaInstance interface provides a unified interface
// over replicas and instances
type replicaInstance interface {
	GetReplicaData() (string, *sf.ReplicaItemBase)
}
