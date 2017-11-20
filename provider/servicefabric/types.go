package servicefabric

import (
	"encoding/xml"

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

// ServiceExtensionLabels provides the structure for
// deserialising the XML document used to store labels in an Extension
type ServiceExtensionLabels struct {
	XMLName xml.Name `xml:"Labels"`
	Label   []struct {
		XMLName xml.Name `xml:"Label"`
		Value   string   `xml:",chardata"`
		Key     string   `xml:"Key,attr"`
	}
}
