package servicefabric

import (
	sfsdk "github.com/jjcollinge/servicefabric"
)

// ServiceItemExtended provides a flattened view
// of the service with details of the application
// it belongs too and the replicas/partitions
type ServiceItemExtended struct {
	sfsdk.ServiceItem
	HasHTTPEndpoint bool
	IsHealthy       bool
	Application     sfsdk.ApplicationItem
	Partitions      []PartitionItemExtended
	Labels          map[string]string
}

// PartitionItemExtended provides a flattened view
// of a services partitions
type PartitionItemExtended struct {
	sfsdk.PartitionItem
	Replicas  []sfsdk.ReplicaItem
	Instances []sfsdk.InstanceItem
}
